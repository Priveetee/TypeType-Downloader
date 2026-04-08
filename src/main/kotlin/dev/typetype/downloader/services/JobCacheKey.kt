package dev.typetype.downloader.services

import java.security.MessageDigest

object JobCacheKey {
    fun fromUrl(url: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val bytes = digest.digest(url.trim().toByteArray())
        return bytes.joinToString("") { "%02x".format(it) }
    }
}
