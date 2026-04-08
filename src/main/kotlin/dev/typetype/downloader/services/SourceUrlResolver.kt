package dev.typetype.downloader.services

import java.net.URI
import java.net.URLDecoder
import java.nio.charset.StandardCharsets

object SourceUrlResolver {
    private val PRIORITY_KEYS = listOf("url", "v", "u", "target", "video", "href", "q")
    private val YOUTUBE_ID_REGEX = Regex("^[A-Za-z0-9_-]{11}$")

    fun resolve(rawUrl: String): String {
        var current = rawUrl.trim()
        repeat(5) {
            val next = unwrapOnce(current) ?: return current
            if (next == current) return current
            current = next
        }
        return current
    }

    private fun unwrapOnce(url: String): String? {
        val uri = runCatching { URI(url) }.getOrNull() ?: return null
        val query = uri.rawQuery ?: return null
        val params = parseQuery(query)
        val ordered = PRIORITY_KEYS.flatMap { key -> params.filter { it.key.equals(key, ignoreCase = true) } } +
            params.filter { p -> PRIORITY_KEYS.none { it.equals(p.key, ignoreCase = true) } }
        for (param in ordered) {
            val candidate = candidateFrom(param.key, decodeRepeated(param.value), uri.host.orEmpty()) ?: continue
            if (candidate != url) return candidate
        }
        return null
    }

    private fun candidateFrom(key: String, value: String, host: String): String? {
        val normalized = value.trim()
        if (normalized.isBlank()) return null
        if (normalized.startsWith("http://") || normalized.startsWith("https://")) return normalized
        if (normalized.startsWith("www.")) return "https://$normalized"
        val isWrappedYoutubeId = key.equals("v", ignoreCase = true) && YOUTUBE_ID_REGEX.matches(normalized)
        val fromYoutubeHost = host.contains("youtube.com") || host == "youtu.be"
        if (isWrappedYoutubeId && !fromYoutubeHost) return "https://www.youtube.com/watch?v=$normalized"
        return null
    }

    private fun decodeRepeated(raw: String): String {
        var current = raw
        repeat(3) {
            val decoded = runCatching { URLDecoder.decode(current, StandardCharsets.UTF_8) }.getOrElse { return current }
            if (decoded == current) return current
            current = decoded
        }
        return current
    }

    private fun parseQuery(query: String): List<QueryParam> = query.split('&').filter { it.isNotBlank() }.map {
        val parts = it.split('=', limit = 2)
        val key = parts[0]
        val value = if (parts.size == 2) parts[1] else ""
        QueryParam(key = key, value = value)
    }

    private data class QueryParam(val key: String, val value: String)
}
