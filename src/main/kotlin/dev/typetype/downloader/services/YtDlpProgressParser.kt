package dev.typetype.downloader.services

import java.util.Locale

data class YtDlpProgressUpdate(
    val stage: String,
    val progressPercent: Int? = null,
    val downloadedBytes: Long? = null,
    val totalBytes: Long? = null,
    val etaSeconds: Long? = null,
    val speedBytesPerSecond: Long? = null,
    val videoCodec: String? = null,
    val audioCodec: String? = null,
    val fps: Double? = null,
    val container: String? = null,
    val formatId: String? = null,
)

object YtDlpProgressParser {
    private val progressRegex = Regex("""\[download\]\s+([0-9.]+)%\s+of\s+([^\s]+)(?:\s+at\s+([^\s]+))?(?:\s+ETA\s+([^\s]+))?.*""")
    private val destinationRegex = Regex("""\[download\]\s+Destination:\s+(.+)""")
    private val formatRegex = Regex("""\[info\].*?format\(s\):\s+(.+)""")
    private val mergeRegex = Regex("""\[Merger\]\s+Merging formats into\s+"([^"]+)"""")
    private val ffmpegAudioRegex = Regex("""\[ExtractAudio\].*Destination:\s+(.+)""")

    fun parse(line: String): YtDlpProgressUpdate? {
        val trimmed = line.trim()
        progressRegex.matchEntire(trimmed)?.let { match ->
            val percent = match.groupValues[1].toDoubleOrNull()?.toInt()
            val total = parseSizeToBytes(match.groupValues[2])
            val speed = parseSpeedToBytesPerSecond(match.groupValues.getOrNull(3).orEmpty())
            val eta = parseEtaToSeconds(match.groupValues.getOrNull(4).orEmpty())
            val downloaded = if (percent != null && total != null) (total * (percent.coerceIn(0, 100) / 100.0)).toLong() else null
            return YtDlpProgressUpdate(
                stage = "downloading",
                progressPercent = percent,
                downloadedBytes = downloaded,
                totalBytes = total,
                etaSeconds = eta,
                speedBytesPerSecond = speed,
            )
        }
        ffmpegAudioRegex.matchEntire(trimmed)?.let { match ->
            val container = match.groupValues[1].substringAfterLast('.', "").lowercase(Locale.ROOT)
            return YtDlpProgressUpdate(stage = "extracting", container = container.ifBlank { null })
        }
        if (trimmed.startsWith("[ExtractAudio]")) return YtDlpProgressUpdate(stage = "extracting")
        if (trimmed.startsWith("[Merger]")) {
            val container = mergeRegex.find(trimmed)?.groupValues?.getOrNull(1)?.substringAfterLast('.', "")
            return YtDlpProgressUpdate(stage = "merging", container = container?.takeIf { it.isNotBlank() })
        }
        destinationRegex.matchEntire(trimmed)?.let { match ->
            val container = match.groupValues[1].substringAfterLast('.', "").lowercase(Locale.ROOT)
            return YtDlpProgressUpdate(stage = "starting", container = container.ifBlank { null })
        }
        formatRegex.matchEntire(trimmed)?.let { match ->
            val id = match.groupValues[1].split(',').firstOrNull()?.trim()?.takeIf { it.isNotBlank() }
            return YtDlpProgressUpdate(stage = "resolving", formatId = id)
        }
        return null
    }

    private fun parseSizeToBytes(value: String): Long? {
        val m = Regex("""([0-9.]+)([KMGTP]?i?B)""").matchEntire(value.trim()) ?: return null
        val number = m.groupValues[1].toDoubleOrNull() ?: return null
        val unit = m.groupValues[2]
        val factor = when (unit) {
            "B" -> 1.0
            "KiB" -> 1024.0
            "MiB" -> 1024.0 * 1024.0
            "GiB" -> 1024.0 * 1024.0 * 1024.0
            "TiB" -> 1024.0 * 1024.0 * 1024.0 * 1024.0
            "kB" -> 1000.0
            "MB" -> 1000.0 * 1000.0
            "GB" -> 1000.0 * 1000.0 * 1000.0
            "TB" -> 1000.0 * 1000.0 * 1000.0 * 1000.0
            else -> return null
        }
        return (number * factor).toLong()
    }

    private fun parseSpeedToBytesPerSecond(value: String): Long? {
        if (value.isBlank() || value == "N/A") return null
        val normalized = value.removeSuffix("/s")
        return parseSizeToBytes(normalized)
    }

    private fun parseEtaToSeconds(value: String): Long? {
        if (value.isBlank() || value == "N/A") return null
        if (':' in value) {
            val parts = value.split(':').mapNotNull { it.toLongOrNull() }
            if (parts.isEmpty()) return null
            return when (parts.size) {
                2 -> parts[0] * 60 + parts[1]
                3 -> parts[0] * 3600 + parts[1] * 60 + parts[2]
                else -> null
            }
        }
        return value.toLongOrNull()
    }
}
