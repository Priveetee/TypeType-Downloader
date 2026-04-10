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
    val errorCode: String? = null,
    val artifactUrl: String? = null,
    val artifactExpiresAt: String? = null,
    val resolved: ResolvedOutput? = null,
    val progressPercent: Int? = null,
    val downloadedBytes: Long? = null,
    val totalBytes: Long? = null,
    val etaSeconds: Long? = null,
    val speedBytesPerSecond: Long? = null,
    val stage: String? = null,
    val queuedAt: String? = null,
    val startedAt: String? = null,
    val finishedAt: String? = null,
    val queueWaitMs: Long? = null,
    val runTimeMs: Long? = null,
    val tokenFetchMs: Long? = null,
    val ytdlpMs: Long? = null,
    val uploadMs: Long? = null,
    val totalMs: Long? = null,
)
