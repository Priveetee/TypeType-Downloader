package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobStatus
import java.nio.file.Path
import java.time.Instant
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit

class ArtifactUploadExecutor(
    private val config: AppConfig,
    private val storageService: GarageStorageService,
    private val jobsRepository: JobsRepository,
    private val statusLoop: WorkerStatusLoop,
    private val progressStore: JobProgressStore,
) {
    private val uploadPool = Executors.newFixedThreadPool(config.uploadConcurrency) { runnable ->
        Thread(runnable, "upload-worker").apply { isDaemon = true }
    }

    fun submitDone(
        id: String,
        cacheKey: String,
        filePath: Path,
        startedAtNs: Long,
        result: YtDlpResult,
        metrics: DownloadPhaseMetrics,
    ) {
        uploadPool.submit {
            val uploadStartedAt = System.nanoTime()
            val durationMs = elapsedMs(startedAtNs)
            runCatching {
                val artifact = uploadArtifact(cacheKey, filePath)
                val uploadMs = PhaseTiming.elapsedMs(uploadStartedAt)
                val measured = result.withMetrics(
                    tokenFetchMs = metrics.tokenFetchMs,
                    ytdlpMs = metrics.ytdlpMs,
                    uploadMs = uploadMs,
                )
                val updated = jobsRepository.markFinishedIfRunning(
                    id = id,
                    status = JobStatus.DONE,
                    durationMs = durationMs,
                    title = measured.title,
                    error = null,
                    artifactKey = artifact.objectKey,
                    artifactExpiresAt = artifact.expiresAt,
                )
                if (!updated) {
                    storageService.deleteObject(artifact.objectKey)
                } else {
                    statusLoop.markFinished(id, "done:$durationMs")
                    progressStore.update(id, WorkerProgressComposer.completed(JobStatus.DONE, measured, artifact))
                }
            }.onFailure { error ->
                jobsRepository.markFinishedIfRunning(
                    id = id,
                    status = JobStatus.FAILED,
                    durationMs = durationMs,
                    title = result.title,
                    error = error.message ?: "artifact upload failed",
                    artifactKey = null,
                    artifactExpiresAt = null,
                )
                statusLoop.markFailed(id)
            }
            FileTreeCleaner.deleteDirectory(filePath.parent)
        }
    }

    fun stop() {
        uploadPool.shutdown()
        uploadPool.awaitTermination(5, TimeUnit.SECONDS)
    }

    private fun elapsedMs(startedAtNs: Long): Long = (System.nanoTime() - startedAtNs) / 1_000_000

    private fun uploadArtifact(cacheKey: String, filePath: Path): StorageArtifact {
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
}
