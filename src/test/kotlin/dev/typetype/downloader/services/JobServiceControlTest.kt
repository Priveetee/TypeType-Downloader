package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.models.JobStatus
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals

class JobServiceControlTest {
    private val jobsRepository = mockk<JobsRepository>()
    private val redis = mockk<redis.clients.jedis.JedisPooled>(relaxed = true)
    private val storage = mockk<GarageStorageService>(relaxed = true)
    private val progressStore = JobProgressStore(redis, config())
    private val service = JobService(jobsRepository, redis, storage, config(), progressStore)

    @Test
    fun `cancel returns not found when job does not exist`() {
        every { jobsRepository.getById("missing") } returns null
        assertEquals(CancelJobResult.NOT_FOUND, service.cancel("missing"))
    }

    @Test
    fun `cancel returns conflict for done job`() {
        every { jobsRepository.getById("done") } returns row("done", JobStatus.DONE)
        assertEquals(CancelJobResult.NOT_CANCELLABLE, service.cancel("done"))
        verify(exactly = 0) { jobsRepository.markCancelled(any()) }
    }

    @Test
    fun `cancel marks queued job as cancelled`() {
        every { jobsRepository.getById("queued") } returns row("queued", JobStatus.QUEUED)
        every { jobsRepository.markCancelled("queued") } returns true
        assertEquals(CancelJobResult.CANCELLED, service.cancel("queued"))
        verify { redis.setex("downloader:job:queued", 600L, "failed:0") }
    }

    @Test
    fun `delete rejects running job`() {
        every { jobsRepository.getById("run") } returns row("run", JobStatus.RUNNING)
        assertEquals(DeleteJobResult.CONFLICT_RUNNING, service.delete("run"))
        verify(exactly = 0) { jobsRepository.deleteIfNotRunning(any()) }
    }

    @Test
    fun `delete removes artifact and redis state`() {
        val existing = row("x", JobStatus.FAILED, artifactKey = "cache/x.mp4")
        every { jobsRepository.getById("x") } returns existing
        every { jobsRepository.deleteIfNotRunning("x") } returns existing
        assertEquals(DeleteJobResult.DELETED, service.delete("x"))
        verify { storage.deleteObject("cache/x.mp4") }
        verify { redis.del("downloader:job:x") }
    }

    private fun row(id: String, status: JobStatus, artifactKey: String? = null): JobRow = JobRow(
        id = id,
        url = "https://www.youtube.com/watch?v=test",
        cacheKey = "cache-$id",
        optionsJson = "{}",
        status = status,
        durationMs = 0,
        title = "",
        error = null,
        artifactKey = artifactKey,
        artifactExpiresAt = Instant.now().plusSeconds(300),
        createdAt = Instant.now().minusSeconds(60),
        startedAt = Instant.now().minusSeconds(30),
        finishedAt = null,
    )

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
