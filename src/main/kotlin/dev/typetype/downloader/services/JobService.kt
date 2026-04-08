package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.CreateJobResponse
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobResponse
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import java.net.URI
import java.time.Duration
import java.time.Instant
import java.util.UUID

enum class CancelJobResult { NOT_FOUND, NOT_CANCELLABLE, CANCELLED }

enum class DeleteJobResult { NOT_FOUND, CONFLICT_RUNNING, DELETED }

class JobService(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
) {
    fun enqueue(url: String, requestedOptions: JobOptions): CreateJobResponse {
        val resolvedUrl = SourceUrlResolver.resolve(url)
        validateUrl(resolvedUrl)
        val options = JobOptionsNormalizer.normalize(requestedOptions)
        val optionsJson = JobOptionsCodec.encode(options)
        val cacheKey = JobCacheKey.from(resolvedUrl, optionsJson)
        val id = UUID.randomUUID().toString()
        val reusable = jobsRepository.findReusableByCacheKey(cacheKey)
        if (reusable != null) {
            jobsRepository.insertDoneFromCache(id = id, url = resolvedUrl, optionsJson = optionsJson, cached = reusable)
            redis.setex(redisJobKey(id), config.jobTtlSeconds, "done:cached")
            return CreateJobResponse(id = id, cached = true)
        }
        val queueSize = redis.llen(config.redisQueueKey)
        if (queueSize >= config.maxQueueSize) {
            throw QueueSaturatedException("Queue is full")
        }
        jobsRepository.insertQueued(id = id, url = resolvedUrl, cacheKey = cacheKey, optionsJson = optionsJson)
        val payload = JobOptionsCodec.encodeQueue(JobOptionsCodec.QueuePayload(id = id, options = options))
        redis.rpush(config.redisQueueKey, payload)
        redis.setex(redisJobKey(id), config.jobTtlSeconds, "queued")
        return CreateJobResponse(id = id, cached = false)
    }

    fun get(id: String): JobResponse? {
        val row = jobsRepository.getById(id) ?: return null
        return row.toResponse(presignUrl(row), row.artifactExpiresAt?.toString())
    }

    fun cancel(id: String): CancelJobResult {
        val row = jobsRepository.getById(id) ?: return CancelJobResult.NOT_FOUND
        if (row.status == JobStatus.DONE || row.status == JobStatus.FAILED) {
            return CancelJobResult.NOT_CANCELLABLE
        }
        return if (jobsRepository.markCancelled(id)) {
            redis.setex(redisJobKey(id), config.jobTtlSeconds, "failed:0")
            CancelJobResult.CANCELLED
        } else {
            CancelJobResult.NOT_CANCELLABLE
        }
    }

    fun delete(id: String): DeleteJobResult {
        val existing = jobsRepository.getById(id) ?: return DeleteJobResult.NOT_FOUND
        if (existing.status == JobStatus.RUNNING) return DeleteJobResult.CONFLICT_RUNNING
        val deleted = jobsRepository.deleteIfNotRunning(id) ?: return DeleteJobResult.NOT_FOUND
        deleted.artifactKey?.let { storageService.deleteObject(it) }
        redis.del(redisJobKey(id))
        return DeleteJobResult.DELETED
    }

    fun recoverPendingJobs() {
        jobsRepository.resetRunningToQueued()
        redis.del(config.redisQueueKey)
        jobsRepository.listQueuedOrRunning().forEach { row ->
            val options = runCatching { JobOptionsCodec.decode(row.optionsJson) }.getOrElse { JobOptions() }
            val payload = JobOptionsCodec.encodeQueue(JobOptionsCodec.QueuePayload(id = row.id, options = options))
            redis.rpush(config.redisQueueKey, payload)
            redis.setex(redisJobKey(row.id), config.jobTtlSeconds, "queued")
        }
    }

    private fun validateUrl(url: String) {
        val parsed = runCatching { URI(url) }.getOrElse { throw IllegalArgumentException("Invalid URL") }
        val scheme = parsed.scheme?.lowercase() ?: throw IllegalArgumentException("Invalid URL")
        if (scheme != "http" && scheme != "https") {
            throw IllegalArgumentException("Unsupported URL scheme")
        }
        if (parsed.host.isNullOrBlank()) {
            throw IllegalArgumentException("Missing URL host")
        }
    }

    private fun redisJobKey(id: String): String = "downloader:job:$id"

    private fun presignUrl(row: JobRow): String? {
        val key = row.artifactKey ?: return null
        val expiresAt = row.artifactExpiresAt ?: return null
        val now = Instant.now()
        if (expiresAt <= now) return null
        val seconds = Duration.between(now, expiresAt).seconds.coerceIn(1, 900)
        return storageService.presignGet(key, Duration.ofSeconds(seconds))
    }

    private fun JobRow.toResponse(artifactUrl: String?, artifactExpiresAt: String?): JobResponse = JobResponse(
        id = id,
        url = url,
        status = status,
        durationMs = durationMs,
        title = title,
        error = error,
        artifactUrl = artifactUrl,
        artifactExpiresAt = artifactExpiresAt,
    )
}
