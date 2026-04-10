package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobStatus
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.nio.file.Files
import kotlin.test.Test

class ArtifactUploadExecutorTest {
    @Test
    fun `submitDone marks job done and updates progress`() {
        val config = config()
        val storage = mockk<GarageStorageService>(relaxed = true)
        val repo = mockk<JobsRepository>(relaxed = true)
        val statusLoop = mockk<WorkerStatusLoop>(relaxed = true)
        val progress = mockk<JobProgressStore>(relaxed = true)
        val file = Files.createTempFile("upload-exec", ".mp4")
        every { repo.markFinishedIfRunning(any(), any(), any(), any(), any(), any(), any()) } returns true
        val executor = ArtifactUploadExecutor(config, storage, repo, statusLoop, progress)
        val result = YtDlpResult(title = "t", filePath = file, error = null, progress = JobProgressState(stage = "finalizing"))
        executor.submitDone("id-1", "cache-k", file, System.nanoTime(), result)
        executor.stop()
        verify { repo.markFinishedIfRunning("id-1", JobStatus.DONE, any(), "t", null, any(), any()) }
        verify { statusLoop.markFinished("id-1", any()) }
    }

    private fun config(): AppConfig = AppConfig(
        httpPort = 18093,
        dbUrl = "jdbc:postgresql://localhost:55432/typetype_downloader",
        dbUser = "typetype",
        dbPassword = "typetype",
        dbPoolSize = 8,
        dbMinIdle = 1,
        redisHost = "localhost",
        redisPort = 56379,
        redisQueueKey = "downloader:queue",
        maxConcurrentWorkers = 2,
        uploadConcurrency = 1,
        maxQueueSize = 100,
        jobTtlSeconds = 600,
        ytdlpBin = "yt-dlp",
        ytdlpTimeoutSeconds = 60,
        enableTranscode = false,
        s3Endpoint = "http://localhost:3900",
        s3Region = "garage",
        s3Bucket = "typetype-downloads",
        s3AccessKey = "k",
        s3SecretKey = "s",
        s3ArtifactTtlSeconds = 7200,
        tokenServiceUrl = "http://localhost:8081",
    )
}
