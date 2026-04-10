package dev.typetype.downloader.services

import dev.typetype.downloader.db.JobRow
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.JobStatus
import io.mockk.every
import io.mockk.mockk
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

class JobViewBuilderTest {
    @Test
    fun `build exposes resolved fields and progress`() {
        val storage = mockk<GarageStorageService>()
        every { storage.presignGet(any(), any(), any()) } returns "http://garage:3900/signed"
        val options = JobOptions(quality = "1080p", format = "mp4", videoItag = "137", audioItag = "140")
        val row = JobRow(
            id = "job-1",
            url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            cacheKey = "cache",
            optionsJson = JobOptionsCodec.encode(options),
            status = JobStatus.QUEUED,
            durationMs = 0,
            title = "Never Gonna Give You Up",
            error = null,
            artifactKey = "cache/file.mp4",
            artifactExpiresAt = java.time.Instant.now().plusSeconds(600),
            createdAt = java.time.Instant.now().minusSeconds(10),
            startedAt = null,
            finishedAt = null,
        )

        val response = JobViewBuilder.build(row, storage, null)
        assertEquals("queued", response.stage)
        assertEquals(0, response.progressPercent)
        assertNotNull(response.resolved)
        assertEquals("137", response.resolved.videoItag)
        assertEquals("140", response.resolved.audioItag)
        assertEquals(1080, response.resolved.height)
        assertEquals("mp4", response.resolved.container)
        assertTrue(response.resolved.fileName!!.endsWith(".mp4"))
    }

    @Test
    fun `build forwards phase timings from progress`() {
        val storage = mockk<GarageStorageService>()
        every { storage.presignGet(any(), any(), any()) } returns "http://garage:3900/signed"
        val row = JobRow(
            id = "job-2",
            url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            cacheKey = "cache",
            optionsJson = JobOptionsCodec.encode(JobOptions()),
            status = JobStatus.RUNNING,
            durationMs = 0,
            title = "x",
            error = null,
            artifactKey = null,
            artifactExpiresAt = null,
            createdAt = java.time.Instant.now().minusSeconds(30),
            startedAt = java.time.Instant.now().minusSeconds(20),
            finishedAt = null,
        )
        val progress = JobProgressState(
            stage = "downloading",
            tokenFetchMs = 22,
            ytdlpMs = 1500,
            uploadMs = 0,
            totalMs = 1522,
        )

        val response = JobViewBuilder.build(row, storage, progress)
        assertEquals(22, response.tokenFetchMs)
        assertEquals(1500, response.ytdlpMs)
        assertEquals(0, response.uploadMs)
        assertEquals(1522, response.totalMs)
    }
}
