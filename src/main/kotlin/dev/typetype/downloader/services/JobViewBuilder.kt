package dev.typetype.downloader.services

import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobResponse
import dev.typetype.downloader.models.JobStatus
import dev.typetype.downloader.models.ResolvedOutput
import java.time.Instant

object JobViewBuilder {
    fun build(row: JobRow, storageService: GarageStorageService, progress: JobProgressState?): JobResponse {
        val options = runCatching { JobOptionsCodec.decode(row.optionsJson) }.map(JobOptionsNormalizer::normalize).getOrElse { JobOptions() }
        val fileName = stableFileName(row, options)
        val artifactUrl = presignUrl(storageService, row, fileName)
        val now = Instant.now()
        val queueWaitMs = if (row.createdAt != null && row.startedAt != null) {
            (row.startedAt.toEpochMilli() - row.createdAt.toEpochMilli()).coerceAtLeast(0)
        } else {
            null
        }
        val runTimeMs = if (row.startedAt != null) {
            val end = row.finishedAt ?: now
            (end.toEpochMilli() - row.startedAt.toEpochMilli()).coerceAtLeast(0)
        } else {
            null
        }
        return JobResponse(
            id = row.id,
            url = row.url,
            status = row.status,
            durationMs = row.durationMs,
            title = row.title,
            error = row.error,
            errorCode = errorCode(row.error),
            artifactUrl = artifactUrl,
            artifactExpiresAt = row.artifactExpiresAt?.toString(),
            resolved = resolved(row, options, fileName, progress),
            progressPercent = progress?.progressPercent ?: progressPercent(row.status),
            downloadedBytes = progress?.downloadedBytes,
            totalBytes = progress?.totalBytes,
            etaSeconds = progress?.etaSeconds,
            speedBytesPerSecond = progress?.speedBytesPerSecond,
            stage = progress?.stage ?: stage(row.status),
            queuedAt = row.createdAt?.toString(),
            startedAt = row.startedAt?.toString(),
            finishedAt = row.finishedAt?.toString(),
            queueWaitMs = queueWaitMs,
            runTimeMs = runTimeMs,
        )
    }

    private fun presignUrl(storageService: GarageStorageService, row: JobRow, fileName: String?): String? {
        val key = row.artifactKey ?: return null
        val expiresAt = row.artifactExpiresAt ?: return null
        return ArtifactUrlSigner.sign(storageService, key, expiresAt, fileName)
    }

    private fun resolved(row: JobRow, options: JobOptions, fileName: String?, progress: JobProgressState?): ResolvedOutput = ResolvedOutput(
        videoItag = options.videoItag.ifBlank { null },
        audioItag = options.audioItag.ifBlank { null },
        height = options.height ?: if (options.allowQualityFallback) null else qualityHeight(options.quality),
        fps = options.fps,
        videoCodec = options.videoCodec.ifBlank { null },
        audioCodec = options.audioCodec.ifBlank { null },
        container = row.artifactKey?.substringAfterLast('.', "").orEmpty().ifBlank { null },
        bitrate = options.bitrate,
        fileName = fileName,
        effectiveVideoCodec = progress?.videoCodec,
        effectiveAudioCodec = progress?.audioCodec,
        effectiveFps = progress?.fps,
        effectiveFormatId = progress?.formatId,
    )

    private fun stableFileName(row: JobRow, options: JobOptions): String? {
        val extension = row.artifactKey?.substringAfterLast('.', "")?.ifBlank { null } ?: return null
        val base = row.title.takeIf { it.isNotBlank() } ?: row.id
        val clean = base.map { if (it.isLetterOrDigit() || it == '-' || it == '_') it else '_' }
            .joinToString("")
            .replace("__", "_")
            .trim('_')
            .ifBlank { row.id }
        val name = if (options.thumbnailOnly) "${clean}_thumbnail" else clean
        return "$name.$extension"
    }

    private fun qualityHeight(quality: String): Int? = when (quality.lowercase()) {
        "1080p" -> 1080
        "720p" -> 720
        "480p" -> 480
        else -> null
    }

    private fun stage(status: JobStatus): String = when (status) {
        JobStatus.QUEUED -> "queued"
        JobStatus.RUNNING -> "downloading"
        JobStatus.DONE -> "finalizing"
        JobStatus.FAILED -> "failed"
    }

    private fun progressPercent(status: JobStatus): Int? = when (status) {
        JobStatus.QUEUED -> 0
        JobStatus.RUNNING -> 50
        JobStatus.DONE -> 100
        JobStatus.FAILED -> null
    }

    private fun errorCode(error: String?): String? {
        val value = error?.lowercase() ?: return null
        return if ("exact selection unavailable" in value) "exact_selection_unavailable" else null
    }
}
