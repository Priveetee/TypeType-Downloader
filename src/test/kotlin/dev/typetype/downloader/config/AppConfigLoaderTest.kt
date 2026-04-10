package dev.typetype.downloader.config

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class AppConfigLoaderTest {
    @Test
    fun `loads non-empty default config`() {
        val config = AppConfigLoader.load()
        assertTrue(config.httpPort > 0)
        assertTrue(config.dbUrl.isNotBlank())
        assertTrue(config.redisQueueKey.isNotBlank())
        assertTrue(config.ytdlpBin.isNotBlank())
        assertTrue(config.s3Endpoint.isNotBlank())
        assertTrue(config.maxConcurrentWorkers >= 1)
        assertTrue(config.uploadConcurrency >= 1)
        assertTrue(config.maxQueueSize >= 1)
        assertTrue(config.jobTtlSeconds >= 1)
        assertTrue(config.dbPoolSize >= 1)
        assertTrue(config.dbMinIdle >= 0)
    }

    @Test
    fun `default region and bucket stay stable`() {
        val config = AppConfigLoader.load()
        assertEquals("garage", config.s3Region)
        assertEquals("typetype-downloads", config.s3Bucket)
        assertEquals(8, config.dbPoolSize)
        assertEquals(1, config.dbMinIdle)
        assertEquals(2, config.uploadConcurrency)
        assertEquals(1, config.ytdlpConcurrentFragments)
        assertEquals(10, config.ytdlpRetries)
        assertEquals(10, config.ytdlpFragmentRetries)
        assertEquals(30, config.ytdlpSocketTimeoutSeconds)
        assertEquals(false, config.audioPassthroughDefault)
        assertEquals(600, config.tokenCacheTtlSeconds)
    }
}
