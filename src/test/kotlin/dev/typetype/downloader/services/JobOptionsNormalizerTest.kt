package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.SubtitlesOptions
import kotlin.test.Test
import kotlin.test.assertEquals

class JobOptionsNormalizerTest {
    @Test
    fun `normalizes subtitle languages and format`() {
        val input = JobOptions(
            subtitles = SubtitlesOptions(
                enabled = true,
                languages = listOf(" en ", "", "fr"),
                format = "  VTT ",
            )
        )
        val normalized = JobOptionsNormalizer.normalize(input)
        assertEquals(listOf("en", "fr"), normalized.subtitles.languages)
        assertEquals("vtt", normalized.subtitles.format)
    }
}
