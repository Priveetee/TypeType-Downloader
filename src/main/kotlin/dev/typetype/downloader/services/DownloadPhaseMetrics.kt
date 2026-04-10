package dev.typetype.downloader.services

import kotlin.math.max

data class DownloadPhaseMetrics(
    val tokenFetchMs: Long = 0,
    val ytdlpMs: Long = 0,
    val uploadMs: Long = 0,
) {
    val totalMs: Long get() = max(0, tokenFetchMs) + max(0, ytdlpMs) + max(0, uploadMs)
}
