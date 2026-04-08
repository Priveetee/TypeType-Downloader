package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.SubtitlesOptions

object JobOptionsNormalizer {
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
        val subtitles = normalizeSubtitles(options.subtitles)
        val sponsorBlockCategories = normalizeSponsorBlockCategories(options)
        return options.copy(subtitles = subtitles, sponsorBlockCategories = sponsorBlockCategories)
    }

    private fun normalizeSubtitles(input: SubtitlesOptions): SubtitlesOptions {
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
