package dev.typetype.downloader.services

import java.util.concurrent.ConcurrentHashMap

class ArtifactUrlCache {
    private val values = ConcurrentHashMap<String, Entry>()

    fun get(key: String): String? {
        val entry = values[key] ?: return null
        return if (entry.expiresAtMs > System.currentTimeMillis()) entry.url else {
            values.remove(key)
            null
        }
    }

    fun put(key: String, url: String, ttlMs: Long) {
        val expiresAt = System.currentTimeMillis() + ttlMs.coerceAtLeast(1)
        values[key] = Entry(url = url, expiresAtMs = expiresAt)
    }

    private data class Entry(
        val url: String,
        val expiresAtMs: Long,
    )
}
