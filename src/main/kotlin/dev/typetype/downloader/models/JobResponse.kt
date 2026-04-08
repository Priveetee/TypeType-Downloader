package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class JobResponse(
    val id: String,
    val url: String,
    val status: JobStatus,
    val durationMs: Long,
    val title: String,
    val error: String? = null,
    val artifactUrl: String? = null,
    val artifactExpiresAt: String? = null,
)
