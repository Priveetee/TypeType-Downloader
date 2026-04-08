package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions

object YtDlpOptionResolver {
    private val audioFormats = setOf("mp3", "m4a", "aac", "opus", "flac", "wav")
    private val videoFormats = setOf("mp4", "webm", "mkv", "mov")

    fun audioSelector(quality: String): String = if (quality.lowercase() == "worst") "worstaudio/worst" else "bestaudio/best"

    fun videoSelector(quality: String): String = when (quality.lowercase()) {
        "1080p" -> "bv*[height<=1080]+ba/b[height<=1080]"
        "720p" -> "bv*[height<=720]+ba/b[height<=720]"
        "480p" -> "bv*[height<=480]+ba/b[height<=480]"
        "worst" -> "worst"
        else -> "bv*+ba/b"
    }

    fun audioFormat(raw: String): String {
        val value = raw.trim().lowercase()
        return if (value in audioFormats) value else "mp3"
    }

    fun videoFormat(raw: String): String {
        val value = raw.trim().lowercase()
        return if (value in videoFormats) value else "mp4"
    }

    fun preferredExtensions(options: JobOptions): List<String> = when {
        options.thumbnailOnly -> listOf("jpg", "jpeg", "png", "webp")
        options.mode == DownloadMode.AUDIO -> {
            val requested = audioFormat(options.format)
            listOf(requested, "mp3", "m4a", "opus", "aac", "flac", "wav", "webm").distinct()
        }
        else -> {
            val requested = videoFormat(options.format)
            listOf(requested, "mp4", "mkv", "webm", "mov").distinct()
        }
    }
}
