package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus
import java.time.Instant

data class JobRow(
    val id: String,
    val url: String,
    val cacheKey: String,
    val status: JobStatus,
    val durationMs: Long,
    val title: String,
    val error: String?,
    val artifactKey: String?,
    val artifactExpiresAt: Instant?,
)
