package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import dev.typetype.downloader.models.SubtitlesOptions

object JobOptionsNormalizer {
    fun normalize(options: JobOptions): JobOptions {
        val subtitles = normalizeSubtitles(options.subtitles)
        return options.copy(subtitles = subtitles)
    }

    private fun normalizeSubtitles(input: SubtitlesOptions): SubtitlesOptions {
        val langs = input.languages.map { it.trim() }.filter { it.isNotBlank() }.ifEmpty { listOf("en") }
        val format = input.format.trim().ifBlank { "srt" }.lowercase()
        return input.copy(languages = langs, format = format)
    }
}
