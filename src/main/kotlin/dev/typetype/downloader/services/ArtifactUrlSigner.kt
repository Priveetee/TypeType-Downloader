package dev.typetype.downloader.services

import java.time.Duration
import java.time.Instant

object ArtifactUrlSigner {
    private val cache = ArtifactUrlCache()

    fun sign(storageService: GarageStorageService, objectKey: String, expiresAt: Instant, fileName: String?): String? {
        val now = Instant.now()
        if (expiresAt <= now) return null
        val cacheKey = "$objectKey|${fileName.orEmpty()}|${expiresAt.toEpochMilli()}"
        cache.get(cacheKey)?.let { return it }
        val seconds = Duration.between(now, expiresAt).seconds.coerceIn(1, 900)
        val url = storageService.presignGet(objectKey, Duration.ofSeconds(seconds), fileName)
        val ttlMs = (seconds.coerceAtMost(20) * 1000).coerceAtLeast(1000)
        cache.put(cacheKey, url, ttlMs)
        return url
    }
}
