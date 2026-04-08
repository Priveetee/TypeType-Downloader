package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import java.net.URI
import java.net.URLEncoder
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.nio.charset.StandardCharsets
import java.time.Duration

class TokenServiceClient(private val config: AppConfig) {
    private val client = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(3)).build()
    private val json = Json { ignoreUnknownKeys = true }

    fun fetchForUrl(url: String): TokenPayload? {
        val videoId = extractYoutubeVideoId(SourceUrlResolver.resolve(url)) ?: return null
        val encoded = URLEncoder.encode(videoId, StandardCharsets.UTF_8)
        val uri = URI("${config.tokenServiceUrl}/potoken?videoId=$encoded")
        val request = HttpRequest.newBuilder(uri).timeout(Duration.ofSeconds(10)).GET().build()
        val response = runCatching { client.send(request, HttpResponse.BodyHandlers.ofString()) }.getOrNull() ?: return null
        if (response.statusCode() != 200) return null
        val body = runCatching { json.decodeFromString(TokenResponse.serializer(), response.body()) }.getOrNull() ?: return null
        if (body.visitorData.isBlank() || body.streamingPot.isBlank()) return null
        return TokenPayload(visitorData = body.visitorData, streamingPot = body.streamingPot)
    }

    private fun extractYoutubeVideoId(url: String): String? {
        val uri = runCatching { URI(url) }.getOrNull() ?: return null
        val host = uri.host?.lowercase() ?: return null
        val query = uri.query.orEmpty()
        if (host.endsWith("youtube.com")) {
            return query.split("&").mapNotNull {
                val parts = it.split("=", limit = 2)
                if (parts.size == 2 && parts[0] == "v") parts[1] else null
            }.firstOrNull { it.isNotBlank() }
        }
        if (host == "youtu.be") {
            return uri.path.trim('/').ifBlank { null }
        }
        return null
    }

    @Serializable
    private data class TokenResponse(val visitorData: String = "", val streamingPot: String = "")
}
