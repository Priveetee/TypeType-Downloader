package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.CreateJobResponse
import dev.typetype.downloader.models.JobResponse
import redis.clients.jedis.JedisPooled
import java.net.URI
import java.util.UUID

class JobService(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val config: AppConfig,
) {
    fun enqueue(url: String): CreateJobResponse {
        validateUrl(url)
        val id = UUID.randomUUID().toString()
        jobsRepository.insertQueued(id = id, url = url)
        redis.rpush(config.redisQueueKey, id)
        redis.setex(redisJobKey(id), 600, "queued")
        return CreateJobResponse(id = id)
    }

    fun get(id: String): JobResponse? {
        val row = jobsRepository.getById(id) ?: return null
        return row.toResponse()
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

    private fun JobRow.toResponse(): JobResponse = JobResponse(
        id = id,
        url = url,
        status = status,
        durationMs = durationMs,
        title = title,
        error = error,
    )
}
