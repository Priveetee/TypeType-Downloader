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
    private val tokenCacheStore: TokenCacheStore,
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
            val token = if (shouldUseTokenFor(options)) resolveToken(job.url) else null
            val tokenMs = PhaseTiming.elapsedMs(startedAt)
            progressStore.update(id, WorkerProgressComposer.running(DownloadPhaseMetrics(tokenFetchMs = tokenMs)))
            val ytdlpStartedAt = System.nanoTime()
            val result = ytDlpService.download(
                url = job.url,
                token = token,
                options = options,
                onProgress = { progressStore.update(id, it) },
            ) {
                statusLoop.shouldCancel(id)
            }
            val ytdlpMs = PhaseTiming.elapsedMs(ytdlpStartedAt)
            val withPhases = result.withMetrics(tokenFetchMs = tokenMs, ytdlpMs = ytdlpMs)
            val status = if (result.error == null) JobStatus.DONE else JobStatus.FAILED
            if (status == JobStatus.DONE) {
                progressStore.update(id, WorkerProgressComposer.finalizing(withPhases))
                val filePath = withPhases.filePath
                if (filePath != null) {
                    uploadExecutor.submitDone(
                        id = id,
                        cacheKey = job.cacheKey,
                        filePath = filePath,
                        startedAtNs = startedAt,
                        result = withPhases,
                        metrics = DownloadPhaseMetrics(tokenFetchMs = tokenMs, ytdlpMs = ytdlpMs),
                    )
                    return
                }
            }
            val durationMs = PhaseTiming.elapsedMs(startedAt)
            val updated = jobsRepository.markFinishedIfRunning(
                id = id,
                status = status,
                durationMs = durationMs,
                title = withPhases.title,
                error = withPhases.error,
                artifactKey = null,
                artifactExpiresAt = null,
            )
            if (updated) {
                statusLoop.markFinished(id, "${status.name.lowercase()}:$durationMs")
                progressStore.update(id, WorkerProgressComposer.completed(status, withPhases, null))
            }
            withPhases.filePath?.parent?.let(FileTreeCleaner::deleteDirectory)
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
        return runCatching { JobOptionsCodec.decode(row.optionsJson) }
            .map { JobOptionsNormalizer.normalize(it, audioPassthroughDefault = config.audioPassthroughDefault) }
            .getOrElse { JobOptions() }
    }

    private fun resolveToken(url: String): TokenPayload? {
        val videoId = tokenServiceClient.resolveVideoId(url) ?: return null
        tokenCacheStore.get(videoId)?.let { return it }
        val fetched = tokenServiceClient.fetchForVideoId(videoId) ?: return null
        tokenCacheStore.put(videoId, fetched)
        return fetched
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
