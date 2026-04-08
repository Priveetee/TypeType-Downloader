package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus

data class JobRow(
    val id: String,
    val url: String,
    val status: JobStatus,
    val durationMs: Long,
    val title: String,
    val error: String?,
)
