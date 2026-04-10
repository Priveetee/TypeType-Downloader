package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobResponse
import dev.typetype.downloader.models.JobStatus
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class JobResponseCacheStoreTest {
    @Test
    fun `stores and retrieves cached response`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        val store = JobResponseCacheStore(redis)
        val response = JobResponse(
            id = "x",
            url = "https://youtube.com/watch?v=x",
            status = JobStatus.QUEUED,
            durationMs = 0,
            title = "",
            stage = "queued",
        )
        val encoded = kotlinx.serialization.json.Json.encodeToString(JobResponse.serializer(), response)
        every { redis.get("downloader:response:x") } returns encoded
        store.put("x", response, 1)
        val loaded = store.get("x")
        assertNotNull(loaded)
        assertEquals("x", loaded.id)
        verify { redis.setex("downloader:response:x", 1, any()) }
    }

    @Test
    fun `returns short ttl for running jobs`() {
        val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
        val store = JobResponseCacheStore(redis)
        assertEquals(1L, store.ttlFor(JobStatus.RUNNING))
        assertEquals(5L, store.ttlFor(JobStatus.DONE))
    }
}
