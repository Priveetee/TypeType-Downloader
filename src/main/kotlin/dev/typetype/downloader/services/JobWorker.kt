package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import kotlin.concurrent.thread

class JobWorker(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val ytDlpService: YtDlpService,
    private val config: AppConfig,
) {
    fun start() {
        thread(name = "job-worker", isDaemon = true) {
            while (true) {
                val item = redis.blpop(0, config.redisQueueKey) ?: continue
                val id = item.getOrNull(1) ?: continue
                process(id)
            }
        }
    }

    private fun process(id: String) {
        val job = jobsRepository.getById(id) ?: return
        jobsRepository.markRunning(id)
        redis.setex(redisJobKey(id), 600, "running")
        val startedAt = System.nanoTime()
        val result = ytDlpService.extractTitle(job.url)
        val durationMs = (System.nanoTime() - startedAt) / 1_000_000
        val status = if (result.error == null) JobStatus.DONE else JobStatus.FAILED
        jobsRepository.markFinished(
            id = id,
            status = status,
            durationMs = durationMs,
            title = result.title,
            error = result.error,
        )
        redis.setex(redisJobKey(id), 600, "${status.name.lowercase()}:$durationMs")
    }

    private fun redisJobKey(id: String): String = "downloader:job:$id"
}
