package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import kotlin.reflect.full.declaredFunctions
import kotlin.reflect.jvm.isAccessible
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class JobWorkerTokenSelectionTest {
    @Test
    fun `uses token for non exact jobs`() {
        assertTrue(shouldUseToken(JobOptions()))
    }

    @Test
    fun `skips token for exact selector inputs`() {
        assertFalse(shouldUseToken(JobOptions(videoItag = "137")))
        assertFalse(shouldUseToken(JobOptions(audioItag = "251")))
        assertFalse(shouldUseToken(JobOptions(height = 720)))
        assertFalse(shouldUseToken(JobOptions(fps = 25)))
        assertFalse(shouldUseToken(JobOptions(videoCodec = "avc1.640028")))
        assertFalse(shouldUseToken(JobOptions(audioCodec = "opus")))
        assertFalse(shouldUseToken(JobOptions(bitrate = 160)))
    }

    private fun shouldUseToken(options: JobOptions): Boolean {
        val method = JobWorker::class.declaredFunctions.first { it.name == "shouldUseTokenFor" }
        method.isAccessible = true
        return method.call(worker(), options) as Boolean
    }

    private fun worker(): JobWorker = JobWorker(
        jobsRepository = io.mockk.mockk(relaxed = true),
        redis = io.mockk.mockk(relaxed = true),
        ytDlpService = io.mockk.mockk(relaxed = true),
        tokenServiceClient = io.mockk.mockk(relaxed = true),
        storageService = io.mockk.mockk(relaxed = true),
        config = dev.typetype.downloader.config.AppConfig(
            httpPort = 18093,
            dbUrl = "jdbc:postgresql://localhost:55432/typetype_downloader",
            dbUser = "typetype",
            dbPassword = "typetype",
            redisHost = "localhost",
            redisPort = 56379,
            redisQueueKey = "downloader:queue",
            maxConcurrentWorkers = 1,
            maxQueueSize = 10,
            jobTtlSeconds = 600,
            ytdlpBin = "yt-dlp",
            ytdlpTimeoutSeconds = 120,
            enableTranscode = false,
            s3Endpoint = "http://localhost:3900",
            s3Region = "garage",
            s3Bucket = "typetype-downloads",
            s3AccessKey = "demo",
            s3SecretKey = "demo",
            s3ArtifactTtlSeconds = 7200,
            tokenServiceUrl = "http://localhost:8081",
        ),
    )
}
