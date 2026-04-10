package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.CreateJobResponse
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobResponse
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import java.net.URI
import java.util.UUID

enum class CancelJobResult { NOT_FOUND, NOT_CANCELLABLE, CANCELLED }

enum class DeleteJobResult { NOT_FOUND, CONFLICT_RUNNING, DELETED }

class JobService(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
    private val progressStore: JobProgressStore,
) {
    private val queuePublisher = JobQueuePublisher(redis, config, progressStore)
    private val responseCacheStore = JobResponseCacheStore(redis)

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
            queuePublisher.markCached(id, reusable.artifactKey)
            return CreateJobResponse(id = id, cached = true)
        }
        val queueSize = redis.llen(config.redisQueueKey)
        if (queueSize >= config.maxQueueSize) {
            throw QueueSaturatedException("Queue is full")
        }
        jobsRepository.insertQueued(id = id, url = resolvedUrl, cacheKey = cacheKey, optionsJson = optionsJson)
        queuePublisher.enqueue(id, options)
        return CreateJobResponse(id = id, cached = false)
    }

    fun get(id: String): JobResponse? {
        responseCacheStore.get(id)?.let { return it }
        val row = jobsRepository.getById(id) ?: return null
        val progress = progressStore.get(id)
        val response = JobViewBuilder.build(row, storageService, progress)
        responseCacheStore.put(id, response, responseCacheStore.ttlFor(row.status))
        return response
    }

    fun cancel(id: String): CancelJobResult {
        val row = jobsRepository.getById(id) ?: return CancelJobResult.NOT_FOUND
        if (row.status == JobStatus.DONE || row.status == JobStatus.FAILED) {
            return CancelJobResult.NOT_CANCELLABLE
        }
        return if (jobsRepository.markCancelled(id)) {
            redis.setex(JobRedisKeys.job(id), config.jobTtlSeconds, "failed:0")
            progressStore.setCancelled(id)
            progressStore.update(id, JobProgressState(stage = "cancelled"))
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
        redis.del(JobRedisKeys.job(id))
        progressStore.clear(id)
        progressStore.clearCancelled(id)
        responseCacheStore.clear(id)
        return DeleteJobResult.DELETED
    }

    fun recoverPendingJobs() {
        jobsRepository.resetRunningToQueued()
        queuePublisher.resetQueue()
        queuePublisher.recoverPending(jobsRepository.listQueuedOrRunning())
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

}
