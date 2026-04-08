package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import java.time.Instant
import kotlin.concurrent.thread

class JobWorker(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val ytDlpService: YtDlpService,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
) {
    fun start() {
        repeat(config.maxConcurrentWorkers) { index ->
            thread(name = "job-worker-$index", isDaemon = true) {
                while (true) {
                    val item = redis.blpop(0, config.redisQueueKey) ?: continue
                    val id = item.getOrNull(1) ?: continue
                    process(id)
                }
            }
        }
    }

    private fun process(id: String) {
        val job = jobsRepository.getById(id) ?: return
        jobsRepository.markRunning(id)
        redis.setex(redisJobKey(id), config.jobTtlSeconds, "running")
        try {
            val startedAt = System.nanoTime()
            val result = ytDlpService.extractTitle(job.url)
            val durationMs = (System.nanoTime() - startedAt) / 1_000_000
            val status = if (result.error == null) JobStatus.DONE else JobStatus.FAILED
            val artifact = if (status == JobStatus.DONE) uploadArtifact(job.id, job.url, job.cacheKey, result.title) else null
            jobsRepository.markFinished(
                id = id,
                status = status,
                durationMs = durationMs,
                title = result.title,
                error = result.error,
                artifactKey = artifact?.objectKey,
                artifactExpiresAt = artifact?.expiresAt,
            )
            redis.setex(redisJobKey(id), config.jobTtlSeconds, "${status.name.lowercase()}:$durationMs")
        } catch (error: Throwable) {
            jobsRepository.markFinished(
                id = id,
                status = JobStatus.FAILED,
                durationMs = 0,
                title = "",
                error = error.message ?: "worker failed",
                artifactKey = null,
                artifactExpiresAt = null,
            )
            redis.setex(redisJobKey(id), config.jobTtlSeconds, "failed:0")
        }
    }

    private fun uploadArtifact(id: String, url: String, cacheKey: String, title: String): StorageArtifact {
        val expiresAt = Instant.now().plusSeconds(config.s3ArtifactTtlSeconds)
        val objectKey = "cache/$cacheKey.txt"
        val payload = "id=$id\nurl=$url\ntitle=$title\n"
        storageService.putBytes(objectKey, payload.toByteArray(), "text/plain")
        return StorageArtifact(objectKey = objectKey, expiresAt = expiresAt)
    }

    private fun redisJobKey(id: String): String = "downloader:job:$id"
}
