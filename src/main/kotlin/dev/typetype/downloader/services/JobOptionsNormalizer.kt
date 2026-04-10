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

    fun normalize(options: JobOptions, audioPassthroughDefault: Boolean = false): JobOptions {
        val format = normalizeFormat(options)
        val quality = normalizeQuality(options)
        val audioPassthrough = options.mode == DownloadMode.AUDIO &&
            (options.audioPassthrough || (audioPassthroughDefault && options.format.isBlank()))
        val normalizedFormat = if (audioPassthrough && options.mode == DownloadMode.AUDIO) "" else format
        val videoItag = normalizeItag(options.videoItag)
        val audioItag = normalizeItag(options.audioItag)
        val height = normalizePositive(options.height, max = 4320)
        val fps = normalizePositive(options.fps, max = 240)
        val bitrate = normalizePositive(options.bitrate, max = 500_000)
        val videoCodec = normalizeCodec(options.videoCodec)
        val audioCodec = normalizeCodec(options.audioCodec)
        val subtitles = normalizeSubtitles(options.subtitles)
        val sponsorBlockCategories = normalizeSponsorBlockCategories(options)
        if (options.thumbnailOnly) {
            return options.copy(
                quality = "best",
                format = "",
                audioPassthrough = false,
                videoItag = "",
                audioItag = "",
                height = null,
                fps = null,
                videoCodec = "",
                audioCodec = "",
                bitrate = null,
                subtitles = SubtitlesOptions(),
                sponsorBlockCategories = emptyList(),
            )
        }
        return options.copy(
            quality = quality,
            format = normalizedFormat,
            audioPassthrough = audioPassthrough,
            videoItag = videoItag,
            audioItag = audioItag,
            height = height,
            fps = fps,
            videoCodec = videoCodec,
            audioCodec = audioCodec,
            bitrate = bitrate,
            subtitles = subtitles,
            sponsorBlockCategories = sponsorBlockCategories,
        )
    }

    private fun normalizeQuality(options: JobOptions): String {
        val raw = options.quality.trim().lowercase().ifBlank { "best" }
        val allowed = if (options.mode == DownloadMode.AUDIO) allowedAudioQualities else allowedVideoQualities
        return if (raw in allowed) raw else "best"
    }

    private fun normalizeFormat(options: JobOptions): String {
        val raw = options.format.trim().lowercase()
        if (options.mode == DownloadMode.AUDIO) {
            if (raw.isBlank()) return "mp3"
            return if (raw in allowedAudioFormats) raw else "mp3"
        }
        if (raw.isBlank()) return "mp4"
        return if (raw in allowedVideoFormats) raw else "mp4"
    }

    private fun normalizeItag(raw: String): String {
        val value = raw.trim()
        return if (value.all { it.isDigit() }) value else ""
    }

    private fun normalizeCodec(raw: String): String {
        val value = raw.trim()
        if (value.isBlank()) return ""
        return if (value.all { it.isLetterOrDigit() || it == '.' || it == '_' || it == '-' }) value else ""
    }

    private fun normalizePositive(value: Int?, max: Int): Int? {
        val current = value ?: return null
        if (current <= 0) return null
        return if (current > max) max else current
    }

    private fun normalizeSubtitles(input: SubtitlesOptions): SubtitlesOptions {
        if (!input.enabled) return SubtitlesOptions()
        val langs = input.languages.map { it.trim() }.filter { it.isNotBlank() }.ifEmpty { listOf("en") }
        val format = input.format.trim().ifBlank { "srt" }.lowercase()
        return input.copy(languages = langs, format = format)
    }

    private fun normalizeSponsorBlockCategories(options: JobOptions): List<String> {
        if (!options.sponsorBlock || options.thumbnailOnly) return emptyList()
        val custom = options.sponsorBlockCategories
            .map { it.trim().lowercase() }
            .filter { it in allowedSponsorBlockCategories }
            .distinct()
        return if (custom.isEmpty()) defaultSponsorBlockCategories else custom
    }
}
