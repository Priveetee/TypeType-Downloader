package dev.typetype.downloader.services

import java.io.ByteArrayInputStream
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class YtDlpOutputReaderTest {
    @Test
    fun `reads title and progress metadata`() {
        val lines = listOf(
            "TT_TITLE:Test Video",
            "TT_META:137+251|avc1.640028|opus|60|webm",
            "[download]  20.0% of 10.00MiB at 2.00MiB/s ETA 00:04",
        ).joinToString("\n") + "\n"
        val updates = mutableListOf<JobProgressState>()
        val reader = YtDlpOutputReader(ByteArrayInputStream(lines.toByteArray())) { updates += it }
        reader.start()
        reader.await()
        val state = reader.snapshot()
        assertEquals("Test Video", state.title)
        assertNotNull(state.progress)
        assertEquals("downloading", state.progress.stage)
        assertEquals("avc1.640028", state.progress.videoCodec)
        assertEquals("opus", state.progress.audioCodec)
        assertEquals("137+251", state.progress.formatId)
        assertEquals("webm", state.progress.container)
        assertEquals(20, state.progress.progressPercent)
        assertEquals(2, updates.size)
    }
}
