package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import java.nio.file.Paths
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class YtDlpCommandFactoryTest {
    @Test
    fun `adds perf flags and external downloader settings`() {
        val command = YtDlpCommandFactory.build(
            config = config(),
            url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            workDir = Paths.get("/tmp"),
            token = null,
            options = JobOptions(mode = DownloadMode.VIDEO),
        )

        assertTrue(command.contains("--concurrent-fragments"))
        assertTrue(command.contains("8"))
        assertTrue(command.contains("--fragment-retries"))
        assertTrue(command.contains("--socket-timeout"))
        assertTrue(command.contains("--http-chunk-size"))
        assertTrue(command.contains("10M"))
        assertTrue(command.contains("--downloader"))
        assertTrue(command.contains("aria2c"))
        assertTrue(command.contains("--downloader-args"))
    }

    @Test
    fun `audio passthrough avoids extract audio flags`() {
        val command = YtDlpCommandFactory.build(
            config = config(),
            url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            workDir = Paths.get("/tmp"),
            token = null,
            options = JobOptions(mode = DownloadMode.AUDIO, audioPassthrough = true),
        )

        assertFalse(command.contains("--extract-audio"))
        assertFalse(command.contains("--audio-format"))
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
        ytdlpTimeoutSeconds = 120,
        ytdlpConcurrentFragments = 8,
        ytdlpRetries = 4,
        ytdlpFragmentRetries = 4,
        ytdlpSocketTimeoutSeconds = 20,
        ytdlpHttpChunkSize = "10M",
        ytdlpExternalDownloader = "aria2c",
        ytdlpExternalDownloaderArgs = "aria2c:-x16 -s16 -k1M",
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
