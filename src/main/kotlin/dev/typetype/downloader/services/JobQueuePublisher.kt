package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.models.JobOptions
import redis.clients.jedis.JedisPooled

class JobQueuePublisher(
    private val redis: JedisPooled,
    private val config: AppConfig,
    private val progressStore: JobProgressStore,
) {
    fun markCached(id: String, artifactKey: String?) {
        redis.setex(JobRedisKeys.job(id), config.jobTtlSeconds, "done:cached")
        progressStore.update(
            id,
            JobProgressState(
                stage = "cached",
                progressPercent = 100,
                container = artifactKey?.substringAfterLast('.', ""),
            ),
        )
    }

    fun enqueue(id: String, options: JobOptions) {
        val payload = JobOptionsCodec.encodeQueue(JobOptionsCodec.QueuePayload(id = id, options = options))
        redis.pipelined().use { pipe ->
            pipe.rpush(config.redisQueueKey, payload)
            pipe.setex(JobRedisKeys.job(id), config.jobTtlSeconds, "queued")
            pipe.sync()
        }
        progressStore.update(id, JobProgressState(stage = "queued", progressPercent = 0))
    }

    fun resetQueue() {
        redis.del(config.redisQueueKey)
    }

    fun recoverPending(rows: List<JobRow>) {
        redis.pipelined().use { pipe ->
            rows.forEach { row ->
                val options = runCatching { JobOptionsCodec.decode(row.optionsJson) }
                    .map { JobOptionsNormalizer.normalize(it, audioPassthroughDefault = config.audioPassthroughDefault) }
                    .getOrElse { JobOptions() }
                val payload = JobOptionsCodec.encodeQueue(JobOptionsCodec.QueuePayload(id = row.id, options = options))
                pipe.rpush(config.redisQueueKey, payload)
                pipe.setex(JobRedisKeys.job(row.id), config.jobTtlSeconds, "queued")
            }
            pipe.sync()
        }
        rows.forEach { row ->
            progressStore.update(row.id, JobProgressState(stage = "queued", progressPercent = 0))
            progressStore.clearCancelled(row.id)
        }
    }
}
