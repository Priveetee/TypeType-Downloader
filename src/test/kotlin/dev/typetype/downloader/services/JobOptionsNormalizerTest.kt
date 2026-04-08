package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
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

    @Test
    fun `fills default sponsorblock categories when enabled and empty`() {
        val normalized = JobOptionsNormalizer.normalize(JobOptions(sponsorBlock = true))
        assertEquals(
            listOf("sponsor", "selfpromo", "interaction", "intro", "outro", "preview", "filler", "music_offtopic"),
            normalized.sponsorBlockCategories,
        )
    }

    @Test
    fun `filters and normalizes custom sponsorblock categories`() {
        val input = JobOptions(
            sponsorBlock = true,
            sponsorBlockCategories = listOf(" Sponsor ", "intro", "invalid", "INTRO"),
        )
        val normalized = JobOptionsNormalizer.normalize(input)
        assertEquals(listOf("sponsor", "intro"), normalized.sponsorBlockCategories)
    }

    @Test
    fun `normalizes video quality and format to allowed values`() {
        val normalized = JobOptionsNormalizer.normalize(JobOptions(quality = " 2160P ", format = "avi"))
        assertEquals("best", normalized.quality)
        assertEquals("mp4", normalized.format)
    }

    @Test
    fun `normalizes audio defaults and allowed audio format`() {
        val normalized = JobOptionsNormalizer.normalize(
            JobOptions(mode = DownloadMode.AUDIO, quality = "", format = " M4A "),
        )
        assertEquals("best", normalized.quality)
        assertEquals("m4a", normalized.format)
    }

    @Test
    fun `thumbnail only ignores quality and format`() {
        val normalized = JobOptionsNormalizer.normalize(
            JobOptions(thumbnailOnly = true, quality = "1080p", format = "webm"),
        )
        assertEquals("best", normalized.quality)
        assertEquals("", normalized.format)
    }

    @Test
    fun `disabled subtitles reset to stable defaults`() {
        val normalized = JobOptionsNormalizer.normalize(
            JobOptions(
                subtitles = SubtitlesOptions(enabled = false, auto = true, embed = true, languages = listOf("fr"), format = "vtt"),
            )
        )
        assertEquals(SubtitlesOptions(), normalized.subtitles)
    }
}
