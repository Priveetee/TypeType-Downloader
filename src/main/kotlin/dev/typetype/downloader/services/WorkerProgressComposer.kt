package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobStatus

object WorkerProgressComposer {
    fun running(): JobProgressState = JobProgressState(stage = "running", progressPercent = 1)

    fun finalizing(result: YtDlpResult): JobProgressState = JobProgressState(
        stage = "finalizing",
        progressPercent = 100,
        downloadedBytes = result.progress?.downloadedBytes,
        totalBytes = result.progress?.totalBytes,
        speedBytesPerSecond = result.progress?.speedBytesPerSecond,
        videoCodec = result.progress?.videoCodec,
        audioCodec = result.progress?.audioCodec,
        fps = result.progress?.fps,
        container = result.progress?.container,
        formatId = result.progress?.formatId,
    )

    fun completed(status: JobStatus, result: YtDlpResult, artifact: StorageArtifact?): JobProgressState {
        if (status == JobStatus.DONE) {
            return JobProgressState(
                stage = "done",
                progressPercent = 100,
                downloadedBytes = result.progress?.downloadedBytes,
                totalBytes = result.progress?.totalBytes,
                speedBytesPerSecond = result.progress?.speedBytesPerSecond,
                videoCodec = result.progress?.videoCodec,
                audioCodec = result.progress?.audioCodec,
                fps = result.progress?.fps,
                container = artifact?.objectKey?.substringAfterLast('.', result.progress?.container.orEmpty())
                    ?.takeIf { it.isNotBlank() },
                formatId = result.progress?.formatId,
            )
        }
        return JobProgressState(stage = "failed", progressPercent = result.progress?.progressPercent)
    }

    fun failed(): JobProgressState = JobProgressState(stage = "failed", progressPercent = null)
}
