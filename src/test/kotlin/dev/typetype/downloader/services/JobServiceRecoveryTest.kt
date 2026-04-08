package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobStatus
import io.mockk.every
import io.mockk.mockk
import io.mockk.slot
import io.mockk.verify
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class JobServiceRecoveryTest {
    private val jobsRepository = mockk<JobsRepository>(relaxed = true)
    private val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
    private val storage = mockk<GarageStorageService>(relaxed = true)
    private val service = JobService(jobsRepository, redis, storage, config())

    @Test
    fun `enqueue stores normalized options json`() {
        val optionsSlot = slot<String>()
        every { jobsRepository.findReusableByCacheKey(any()) } returns null
        every { redis.llen("downloader:queue") } returns 0L
        every { jobsRepository.insertQueued(any(), any(), any(), capture(optionsSlot)) } returns Unit
        val options = JobOptions(quality = "720p", format = "webm", sponsorBlock = true)
        val created = service.enqueue("https://www.youtube.com/watch?v=dQw4w9WgXcQ", options)
        val decoded = JobOptionsCodec.decode(optionsSlot.captured)
        assertFalse(created.cached)
        assertEquals("720p", decoded.quality)
        assertEquals("webm", decoded.format)
        assertTrue(decoded.sponsorBlockCategories.isNotEmpty())
    }

    @Test
    fun `recovery requeues pending jobs with stored options`() {
        val one = row("a", JobStatus.QUEUED, JobOptionsCodec.encode(JobOptions(quality = "1080p")))
        val two = row("b", JobStatus.RUNNING, JobOptionsCodec.encode(JobOptions(mode = dev.typetype.downloader.models.DownloadMode.AUDIO)))
        val payloads = mutableListOf<String>()
        every { jobsRepository.listQueuedOrRunning() } returns listOf(one, two)
        every { redis.rpush("downloader:queue", capture(payloads)) } returns 1L
        service.recoverPendingJobs()
        verify { jobsRepository.resetRunningToQueued() }
        verify { redis.del("downloader:queue") }
        verify(exactly = 2) { redis.setex(any(), 600L, "queued") }
        assertEquals(setOf("a", "b"), payloads.mapNotNull { JobOptionsCodec.decodeQueue(it)?.id }.toSet())
    }

    private fun row(id: String, status: JobStatus, optionsJson: String): JobRow = JobRow(
        id = id,
        url = "https://www.youtube.com/watch?v=$id",
        cacheKey = "cache-$id",
        optionsJson = optionsJson,
        status = status,
        durationMs = 0,
        title = "",
        error = null,
        artifactKey = null,
        artifactExpiresAt = Instant.now().plusSeconds(300),
    )

    private fun config(): AppConfig = AppConfig(
        httpPort = 18093,
        dbUrl = "jdbc:postgresql://localhost:55432/typetype_downloader",
        dbUser = "typetype",
        dbPassword = "typetype",
        redisHost = "localhost",
        redisPort = 56379,
        redisQueueKey = "downloader:queue",
        maxConcurrentWorkers = 2,
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
