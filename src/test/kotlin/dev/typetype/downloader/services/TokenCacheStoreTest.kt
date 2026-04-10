package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class TokenCacheStoreTest {
    @Test
    fun `stores and reads cached token`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        val config = config(tokenCacheTtlSeconds = 300)
        val store = TokenCacheStore(redis, config)
        val token = TokenPayload(visitorData = "visitor", streamingPot = "pot")

        every { redis.get("downloader:token:video-1") } returns "visitor|pot"

        store.put("video-1", token)
        val loaded = store.get("video-1")

        assertNotNull(loaded)
        assertEquals("visitor", loaded.visitorData)
        assertEquals("pot", loaded.streamingPot)
        verify { redis.setex("downloader:token:video-1", 300L, "visitor|pot") }
    }

    @Test
    fun `disabled cache ttl skips write and read`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        val store = TokenCacheStore(redis, config(tokenCacheTtlSeconds = 0))
        store.put("video-2", TokenPayload(visitorData = "v", streamingPot = "p"))
        assertNull(store.get("video-2"))
        verify(exactly = 0) { redis.setex(any<String>(), any<Long>(), any<String>()) }
    }

    private fun config(tokenCacheTtlSeconds: Long): AppConfig = AppConfig(
        httpPort = 18093,
        dbUrl = "jdbc:postgresql://localhost:55432/typetype_downloader",
        dbUser = "typetype",
        dbPassword = "typetype",
        dbPoolSize = 8,
        dbMinIdle = 1,
        redisHost = "localhost",
        redisPort = 56379,
        redisQueueKey = "downloader:queue",
        maxConcurrentWorkers = 1,
        uploadConcurrency = 1,
        maxQueueSize = 100,
        jobTtlSeconds = 600,
        ytdlpBin = "yt-dlp",
        ytdlpTimeoutSeconds = 120,
        ytdlpConcurrentFragments = 1,
        ytdlpRetries = 10,
        ytdlpFragmentRetries = 10,
        ytdlpSocketTimeoutSeconds = 30,
        ytdlpHttpChunkSize = "",
        ytdlpExternalDownloader = "",
        ytdlpExternalDownloaderArgs = "",
        audioPassthroughDefault = false,
        tokenCacheTtlSeconds = tokenCacheTtlSeconds,
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
