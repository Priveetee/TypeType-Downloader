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
        assertTrue(config.maxQueueSize >= 1)
        assertTrue(config.jobTtlSeconds >= 1)
    }

    @Test
    fun `default region and bucket stay stable`() {
        val config = AppConfigLoader.load()
        assertEquals("garage", config.s3Region)
        assertEquals("typetype-downloads", config.s3Bucket)
    }
}
