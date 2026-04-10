package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import redis.clients.jedis.JedisPooled

class TokenCacheStore(
    private val redis: JedisPooled,
    private val config: AppConfig,
) {
    fun get(videoId: String): TokenPayload? {
        if (config.tokenCacheTtlSeconds <= 0) return null
        val raw = redis.get(key(videoId)) ?: return null
        val parts = raw.split('|', limit = 2)
        if (parts.size != 2) return null
        val visitorData = parts[0]
        val streamingPot = parts[1]
        if (visitorData.isBlank() || streamingPot.isBlank()) return null
        return TokenPayload(visitorData = visitorData, streamingPot = streamingPot)
    }

    fun put(videoId: String, token: TokenPayload) {
        if (config.tokenCacheTtlSeconds <= 0) return
        if (token.visitorData.isBlank() || token.streamingPot.isBlank()) return
        redis.setex(key(videoId), config.tokenCacheTtlSeconds, "${token.visitorData}|${token.streamingPot}")
    }

    private fun key(videoId: String): String = "downloader:token:$videoId"
}
