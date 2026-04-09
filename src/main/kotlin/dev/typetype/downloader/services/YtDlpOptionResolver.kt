package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions

object YtDlpOptionResolver {
    private val audioFormats = setOf("mp3", "m4a", "aac", "opus", "flac", "wav")
    private val videoFormats = setOf("mp4", "webm", "mkv", "mov")

    fun audioSelector(options: JobOptions): String {
        val audioItag = options.audioItag
        if (audioItag.isNotBlank()) {
            val strict = "ba[format_id='$audioItag']/b[format_id='$audioItag']"
            return if (options.allowQualityFallback) "$strict/bestaudio/best" else strict
        }
        return if (options.quality.lowercase() == "worst") "worstaudio/worst" else "bestaudio/best"
    }

    fun videoSelector(options: JobOptions): String {
        if (options.videoItag.isNotBlank() || options.audioItag.isNotBlank() || hasTupleSelection(options)) {
            val exact = exactVideoSelector(options)
            if (!options.allowQualityFallback) return exact
            return "$exact/${qualityVideoSelector(options.quality, allowFallback = true)}"
        }
        return qualityVideoSelector(options.quality, options.allowQualityFallback)
    }

    private fun hasTupleSelection(options: JobOptions): Boolean {
        return options.height != null || options.fps != null || options.videoCodec.isNotBlank() ||
            options.audioCodec.isNotBlank() || options.bitrate != null
    }

    private fun exactVideoSelector(options: JobOptions): String {
        val videoFilters = mutableListOf<String>()
        val audioFilters = mutableListOf<String>()
        options.videoItag.takeIf { it.isNotBlank() }?.let { videoFilters += "[format_id='$it']" }
        options.audioItag.takeIf { it.isNotBlank() }?.let { audioFilters += "[format_id='$it']" }
        options.height?.let { videoFilters += "[height=$it]" }
        options.fps?.let { videoFilters += "[fps=$it]" }
        options.videoCodec.takeIf { it.isNotBlank() }?.let { videoFilters += "[vcodec^=$it]" }
        options.audioCodec.takeIf { it.isNotBlank() }?.let { audioFilters += "[acodec^=$it]" }
        options.bitrate?.let { videoFilters += "[tbr>=$it][tbr<=${it + 300}]" }
        val video = "bv*${videoFilters.joinToString("")}"
        val audio = "ba${audioFilters.joinToString("")}"
        return if (options.videoItag.isNotBlank() && options.audioItag.isBlank()) {
            "$video+$audio/b[format_id='${options.videoItag}']"
        } else {
            "$video+$audio"
        }
    }

    private fun qualityVideoSelector(quality: String, allowFallback: Boolean): String = when (quality.lowercase()) {
        "1080p" -> if (allowFallback) "bv*[height<=1080]+ba/b[height<=1080]" else "bv*[height=1080]+ba/b[height=1080]"
        "720p" -> if (allowFallback) "bv*[height<=720]+ba/b[height<=720]" else "bv*[height=720]+ba/b[height=720]"
        "480p" -> if (allowFallback) "bv*[height<=480]+ba/b[height<=480]" else "bv*[height=480]+ba/b[height=480]"
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
