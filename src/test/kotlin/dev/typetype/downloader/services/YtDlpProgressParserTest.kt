package dev.typetype.downloader.services

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class YtDlpProgressParserTest {
    @Test
    fun `parses download line with size speed and eta`() {
        val line = "[download]  50.0% of 20.00MiB at 2.00MiB/s ETA 00:05"
        val parsed = YtDlpProgressParser.parse(line)
        assertNotNull(parsed)
        assertEquals("downloading", parsed.stage)
        assertEquals(50, parsed.progressPercent)
        assertEquals(20L * 1024L * 1024L, parsed.totalBytes)
        assertEquals(2L * 1024L * 1024L, parsed.speedBytesPerSecond)
        assertEquals(5L, parsed.etaSeconds)
    }

    @Test
    fun `parses merger line with container`() {
        val line = "[Merger] Merging formats into \"abc123.webm\""
        val parsed = YtDlpProgressParser.parse(line)
        assertNotNull(parsed)
        assertEquals("merging", parsed.stage)
        assertEquals("webm", parsed.container)
    }

    @Test
    fun `parses extract audio destination container`() {
        val line = "[ExtractAudio] Destination: video-id.m4a"
        val parsed = YtDlpProgressParser.parse(line)
        assertNotNull(parsed)
        assertEquals("extracting", parsed.stage)
        assertEquals("m4a", parsed.container)
    }
}
