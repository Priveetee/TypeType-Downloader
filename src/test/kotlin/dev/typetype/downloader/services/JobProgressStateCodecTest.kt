package dev.typetype.downloader.services

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class JobProgressStateCodecTest {
    @Test
    fun `encodes and decodes progress state`() {
        val state = JobProgressState(
            stage = "downloading",
            progressPercent = 42,
            downloadedBytes = 4200,
            totalBytes = 10000,
            etaSeconds = 7,
            speedBytesPerSecond = 2048,
            videoCodec = "avc1",
            audioCodec = "opus",
            fps = 60.0,
            container = "mp4",
            formatId = "137+251",
            tokenFetchMs = 25,
            ytdlpMs = 1800,
            uploadMs = 450,
            totalMs = 2275,
        )
        val encoded = JobProgressStateCodec.encode(state)
        val decoded = JobProgressStateCodec.decode(encoded)
        assertNotNull(decoded)
        assertEquals(state, decoded)
    }
}
