package dev.typetype.downloader.services

import java.time.Instant

data class StorageArtifact(
    val objectKey: String,
    val expiresAt: Instant,
)
