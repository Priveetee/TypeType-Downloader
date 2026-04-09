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
            JobOptions(thumbnailOnly = true, quality = "1080p", format = "webm", videoItag = "137", height = 1080),
        )
        assertEquals("best", normalized.quality)
        assertEquals("", normalized.format)
        assertEquals("", normalized.videoItag)
        assertEquals(null, normalized.height)
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

    @Test
    fun `normalizes exact selection fields`() {
        val normalized = JobOptionsNormalizer.normalize(
            JobOptions(
                videoItag = " 137 ",
                audioItag = "abc",
                height = 1080,
                fps = -5,
                videoCodec = " avc1.640028 ",
                audioCodec = " mp4a.40.2 ",
                bitrate = 0,
            ),
        )
        assertEquals("137", normalized.videoItag)
        assertEquals("", normalized.audioItag)
        assertEquals(1080, normalized.height)
        assertEquals(null, normalized.fps)
        assertEquals("avc1.640028", normalized.videoCodec)
        assertEquals("mp4a.40.2", normalized.audioCodec)
        assertEquals(null, normalized.bitrate)
    }
}
