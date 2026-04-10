package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlin.test.Test

class JobProgressStoreTest {
    @Test
    fun `throttles high frequency writes with unchanged stage and tiny deltas`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        every { redis.setex(any<String>(), any<Long>(), any<String>()) } returns "OK"
        val store = JobProgressStore(redis, config())
        store.update("job-1", JobProgressState(stage = "downloading", progressPercent = 10, downloadedBytes = 1000, totalBytes = 10_000))
        store.update("job-1", JobProgressState(stage = "downloading", progressPercent = 10, downloadedBytes = 1100, totalBytes = 10_000))
        verify(exactly = 1) { redis.setex("downloader:progress:job-1", 600L, any()) }
    }

    @Test
    fun `writes immediately when stage changes`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        every { redis.setex(any<String>(), any<Long>(), any<String>()) } returns "OK"
        val store = JobProgressStore(redis, config())
        store.update("job-2", JobProgressState(stage = "downloading", progressPercent = 70))
        store.update("job-2", JobProgressState(stage = "merging", progressPercent = 90))
        verify(exactly = 2) { redis.setex("downloader:progress:job-2", 600L, any()) }
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
        uploadConcurrency = 2,
        maxQueueSize = 100,
        jobTtlSeconds = 600,
        ytdlpBin = "yt-dlp",
        ytdlpTimeoutSeconds = 60,
        ytdlpConcurrentFragments = 1,
        ytdlpRetries = 10,
        ytdlpFragmentRetries = 10,
        ytdlpSocketTimeoutSeconds = 30,
        ytdlpHttpChunkSize = "",
        ytdlpExternalDownloader = "",
        ytdlpExternalDownloaderArgs = "",
        audioPassthroughDefault = false,
        tokenCacheTtlSeconds = 600,
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
