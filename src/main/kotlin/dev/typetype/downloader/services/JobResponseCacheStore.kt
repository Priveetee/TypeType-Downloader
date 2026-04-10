package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobResponse
import dev.typetype.downloader.models.JobStatus
import kotlinx.serialization.json.Json
import redis.clients.jedis.JedisPooled

class JobResponseCacheStore(private val redis: JedisPooled) {
    private val json = Json { ignoreUnknownKeys = true }

    fun get(id: String): JobResponse? {
        val raw = redis.get(cacheKey(id)) ?: return null
        return runCatching { json.decodeFromString(JobResponse.serializer(), raw) }.getOrNull()
    }

    fun put(id: String, response: JobResponse, ttlSeconds: Long) {
        redis.setex(cacheKey(id), ttlSeconds, json.encodeToString(JobResponse.serializer(), response))
    }

    fun ttlFor(status: JobStatus): Long = when (status) {
        JobStatus.QUEUED -> 1
        JobStatus.RUNNING -> 1
        JobStatus.DONE -> 5
        JobStatus.FAILED -> 5
    }

    fun clear(id: String) {
        redis.del(cacheKey(id))
    }

    private fun cacheKey(id: String): String = "downloader:response:$id"
}
