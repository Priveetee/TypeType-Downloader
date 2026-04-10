package dev.typetype.downloader.services

import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import java.time.Duration
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals

class ArtifactUrlSignerTest {
    @Test
    fun `reuses short lived cached signed URL`() {
        val storage = mockk<GarageStorageService>()
        every { storage.presignGet(any(), any(), any()) } returns "http://signed/one"
        val expiresAt = Instant.now().plusSeconds(120)
        val first = ArtifactUrlSigner.sign(storage, "cache/object.mp4", expiresAt, "file.mp4")
        val second = ArtifactUrlSigner.sign(storage, "cache/object.mp4", expiresAt, "file.mp4")
        assertEquals(first, second)
        verify(exactly = 1) { storage.presignGet("cache/object.mp4", any<Duration>(), "file.mp4") }
    }
}
