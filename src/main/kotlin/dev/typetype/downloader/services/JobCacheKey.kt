package dev.typetype.downloader.services

import java.security.MessageDigest

object JobCacheKey {
    fun from(url: String, optionsJson: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val payload = "${url.trim()}|$optionsJson"
        val bytes = digest.digest(payload.toByteArray())
        return bytes.joinToString("") { "%02x".format(it) }
    }
}
