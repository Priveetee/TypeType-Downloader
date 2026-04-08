package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import java.nio.file.Files
import java.time.Instant
import kotlin.concurrent.thread

class JobWorker(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val ytDlpService: YtDlpService,
    private val tokenServiceClient: TokenServiceClient,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
) {
    fun start() {
        repeat(config.maxConcurrentWorkers) { index ->
            thread(name = "job-worker-$index", isDaemon = true) {
                while (true) {
                    val item = redis.blpop(0, config.redisQueueKey) ?: continue
                    val raw = item.getOrNull(1) ?: continue
                    val payload = JobOptionsCodec.decodeQueue(raw)
                    val id = payload?.id ?: raw
                    val options = payload?.options ?: JobOptions()
                    process(id, options)
                }
            }
        }
    }

    private fun process(id: String, options: JobOptions) {
        val job = jobsRepository.getById(id) ?: return
        jobsRepository.markRunning(id)
        redis.setex(redisJobKey(id), config.jobTtlSeconds, "running")
        try {
            val startedAt = System.nanoTime()
            val token = tokenServiceClient.fetchForUrl(job.url)
            val result = ytDlpService.download(job.url, token, options)
            val durationMs = (System.nanoTime() - startedAt) / 1_000_000
            val status = if (result.error == null) JobStatus.DONE else JobStatus.FAILED
            val artifact = if (status == JobStatus.DONE && result.filePath != null) {
                uploadArtifact(job.cacheKey, result.filePath)
            } else {
                null
            }
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
            result.filePath?.parent?.let { deleteDirectory(it) }
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

    private fun uploadArtifact(cacheKey: String, filePath: java.nio.file.Path): StorageArtifact {
        val expiresAt = Instant.now().plusSeconds(config.s3ArtifactTtlSeconds)
        val extension = filePath.fileName.toString().substringAfterLast('.', "bin")
        val objectKey = "cache/$cacheKey.$extension"
        storageService.putFile(objectKey, filePath, contentType(extension))
        return StorageArtifact(objectKey = objectKey, expiresAt = expiresAt)
    }

    private fun contentType(extension: String): String = when (extension.lowercase()) {
        "mp4" -> "video/mp4"
        "webm" -> "video/webm"
        "mkv" -> "video/x-matroska"
        "m4a" -> "audio/mp4"
        "mp3" -> "audio/mpeg"
        else -> "application/octet-stream"
    }

    private fun deleteDirectory(dir: java.nio.file.Path) {
        if (!Files.exists(dir)) return
        Files.walk(dir).sorted(Comparator.reverseOrder()).forEach { Files.deleteIfExists(it) }
    }

    private fun redisJobKey(id: String): String = "downloader:job:$id"
}
