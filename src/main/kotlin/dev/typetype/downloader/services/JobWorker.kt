package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobStatus
import redis.clients.jedis.JedisPooled
import kotlin.concurrent.thread

class JobWorker(
    private val jobsRepository: JobsRepository,
    private val redis: JedisPooled,
    private val ytDlpService: YtDlpService,
    private val tokenServiceClient: TokenServiceClient,
    private val storageService: GarageStorageService,
    private val config: AppConfig,
    private val progressStore: JobProgressStore,
) {
    private val statusLoop = WorkerStatusLoop(config, redis, progressStore)
    private val uploadExecutor = ArtifactUploadExecutor(config, storageService, jobsRepository, statusLoop, progressStore)

    fun start() {
        repeat(config.maxConcurrentWorkers) { index ->
            thread(name = "job-worker-$index", isDaemon = true) {
                statusLoop.run(::process, ::decodeStoredOptions)
            }
        }
    }

    private fun process(id: String, options: JobOptions) {
        val job = jobsRepository.getById(id) ?: return
        if (!jobsRepository.markRunningIfQueued(id)) return
        statusLoop.markRunning(id)
        try {
            val startedAt = System.nanoTime()
            val token = if (shouldUseTokenFor(options)) tokenServiceClient.fetchForUrl(job.url) else null
            val result = ytDlpService.download(
                url = job.url,
                token = token,
                options = options,
                onProgress = { progressStore.update(id, it) },
            ) {
                statusLoop.shouldCancel(id)
            }
            val status = if (result.error == null) JobStatus.DONE else JobStatus.FAILED
            if (status == JobStatus.DONE) {
                progressStore.update(id, WorkerProgressComposer.finalizing(result))
                val filePath = result.filePath
                if (filePath != null) {
                    uploadExecutor.submitDone(
                        id = id,
                        cacheKey = job.cacheKey,
                        filePath = filePath,
                        startedAtNs = startedAt,
                        result = result,
                    )
                    return
                }
            }
            val durationMs = (System.nanoTime() - startedAt) / 1_000_000
            val updated = jobsRepository.markFinishedIfRunning(
                id = id,
                status = status,
                durationMs = durationMs,
                title = result.title,
                error = result.error,
                artifactKey = null,
                artifactExpiresAt = null,
            )
            if (updated) {
                statusLoop.markFinished(id, "${status.name.lowercase()}:$durationMs")
                progressStore.update(id, WorkerProgressComposer.completed(status, result, null))
            }
            result.filePath?.parent?.let(FileTreeCleaner::deleteDirectory)
        } catch (error: Throwable) {
            jobsRepository.markFinishedIfRunning(
                id = id,
                status = JobStatus.FAILED,
                durationMs = 0,
                title = "",
                error = error.message ?: "worker failed",
                artifactKey = null,
                artifactExpiresAt = null,
            )
            statusLoop.markFailed(id)
        }
    }

    private fun decodeStoredOptions(id: String): JobOptions {
        val row = jobsRepository.getById(id) ?: return JobOptions()
        return runCatching { JobOptionsCodec.decode(row.optionsJson) }.map(JobOptionsNormalizer::normalize).getOrElse { JobOptions() }
    }

    private fun shouldUseTokenFor(options: JobOptions): Boolean {
        val hasExactSelection = options.videoItag.isNotBlank() || options.audioItag.isNotBlank() ||
            options.height != null || options.fps != null || options.videoCodec.isNotBlank() ||
            options.audioCodec.isNotBlank() || options.bitrate != null
        return !hasExactSelection
    }

    fun stop() {
        uploadExecutor.stop()
    }

}
