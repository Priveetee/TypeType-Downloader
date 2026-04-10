package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.models.JobOptions
import redis.clients.jedis.JedisPooled

class WorkerStatusLoop(
    private val config: AppConfig,
    private val redis: JedisPooled,
    private val progressStore: JobProgressStore,
) {
    fun run(worker: (String, JobOptions) -> Unit, decodeStoredOptions: (String) -> JobOptions) {
        while (true) {
            val item = redis.blpop(0, config.redisQueueKey) ?: continue
            val raw = item.getOrNull(1) ?: continue
            val payload = JobOptionsCodec.decodeQueue(raw)
            val id = payload?.id ?: raw
            val options = payload?.options
                ?.let { JobOptionsNormalizer.normalize(it, audioPassthroughDefault = config.audioPassthroughDefault) }
                ?: decodeStoredOptions(id)
            worker(id, options)
        }
    }

    fun markRunning(id: String) {
        progressStore.clearCancelled(id)
        redis.setex(JobRedisKeys.job(id), config.jobTtlSeconds, "running")
        progressStore.update(id, WorkerProgressComposer.running())
    }

    fun markFinished(id: String, value: String) {
        redis.setex(JobRedisKeys.job(id), config.jobTtlSeconds, value)
    }

    fun markFailed(id: String) {
        redis.setex(JobRedisKeys.job(id), config.jobTtlSeconds, "failed:0")
        progressStore.update(id, WorkerProgressComposer.failed())
    }

    fun shouldCancel(id: String): Boolean = progressStore.isCancelled(id)
}
