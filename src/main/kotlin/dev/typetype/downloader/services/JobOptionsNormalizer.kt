package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.SubtitlesOptions

object JobOptionsNormalizer {
    private val allowedVideoQualities = setOf("best", "1080p", "720p", "480p", "worst")
    private val allowedAudioQualities = setOf("best", "worst")
    private val allowedVideoFormats = setOf("mp4", "webm", "mkv", "mov")
    private val allowedAudioFormats = setOf("mp3", "m4a", "aac", "opus", "flac", "wav")

    private val defaultSponsorBlockCategories = listOf(
        "sponsor",
        "selfpromo",
        "interaction",
        "intro",
        "outro",
        "preview",
        "filler",
        "music_offtopic",
    )

    private val allowedSponsorBlockCategories = defaultSponsorBlockCategories.toSet()

    fun normalize(options: JobOptions): JobOptions {
        val quality = normalizeQuality(options)
        val format = normalizeFormat(options)
        val subtitles = normalizeSubtitles(options.subtitles)
        val sponsorBlockCategories = normalizeSponsorBlockCategories(options)
        return options.copy(
            quality = quality,
            format = format,
            subtitles = subtitles,
            sponsorBlockCategories = sponsorBlockCategories,
        )
    }

    private fun normalizeQuality(options: JobOptions): String {
        if (options.thumbnailOnly) return "best"
        val raw = options.quality.trim().lowercase().ifBlank { "best" }
        val allowed = if (options.mode == DownloadMode.AUDIO) allowedAudioQualities else allowedVideoQualities
        return if (raw in allowed) raw else "best"
    }

    private fun normalizeFormat(options: JobOptions): String {
        if (options.thumbnailOnly) return ""
        val raw = options.format.trim().lowercase()
        if (options.mode == DownloadMode.AUDIO) {
            if (raw.isBlank()) return "mp3"
            return if (raw in allowedAudioFormats) raw else "mp3"
        }
        if (raw.isBlank()) return "mp4"
        return if (raw in allowedVideoFormats) raw else "mp4"
    }

    private fun normalizeSubtitles(input: SubtitlesOptions): SubtitlesOptions {
        if (!input.enabled) return SubtitlesOptions()
        val langs = input.languages.map { it.trim() }.filter { it.isNotBlank() }.ifEmpty { listOf("en") }
        val format = input.format.trim().ifBlank { "srt" }.lowercase()
        return input.copy(languages = langs, format = format)
    }

    private fun normalizeSponsorBlockCategories(options: JobOptions): List<String> {
        if (!options.sponsorBlock) return emptyList()
        val custom = options.sponsorBlockCategories
            .map { it.trim().lowercase() }
            .filter { it in allowedSponsorBlockCategories }
            .distinct()
        return if (custom.isEmpty()) defaultSponsorBlockCategories else custom
    }
}
