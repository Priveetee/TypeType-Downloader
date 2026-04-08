package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.CreateJobResponse
import dev.typetype.downloader.models.JobResponse
import redis.clients.jedis.JedisPooled
import java.net.URI
import java.time.Duration
import java.time.Instant
import java.util.UUID

class JobService(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
) {
    fun enqueue(url: String): CreateJobResponse {
        validateUrl(url)
        val cacheKey = JobCacheKey.fromUrl(url)
        val id = UUID.randomUUID().toString()
        val reusable = jobsRepository.findReusableByCacheKey(cacheKey)
        if (reusable != null) {
            jobsRepository.insertDoneFromCache(id = id, url = url, cached = reusable)
            redis.setex(redisJobKey(id), config.jobTtlSeconds, "done:cached")
            return CreateJobResponse(id = id, cached = true)
        }
        val queueSize = redis.llen(config.redisQueueKey)
        if (queueSize >= config.maxQueueSize) {
            throw QueueSaturatedException("Queue is full")
        }
        jobsRepository.insertQueued(id = id, url = url, cacheKey = cacheKey)
        redis.rpush(config.redisQueueKey, id)
        redis.setex(redisJobKey(id), config.jobTtlSeconds, "queued")
        return CreateJobResponse(id = id, cached = false)
    }

    fun get(id: String): JobResponse? {
        val row = jobsRepository.getById(id) ?: return null
        return row.toResponse(presignUrl(row), row.artifactExpiresAt?.toString())
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
