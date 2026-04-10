package dev.typetype.downloader.services

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class JobProgressState(
    val stage: String,
    val progressPercent: Int? = null,
    val downloadedBytes: Long? = null,
    val totalBytes: Long? = null,
    val etaSeconds: Long? = null,
    val speedBytesPerSecond: Long? = null,
    val videoCodec: String? = null,
    val audioCodec: String? = null,
    val fps: Double? = null,
    val container: String? = null,
    val formatId: String? = null,
    val tokenFetchMs: Long? = null,
    val ytdlpMs: Long? = null,
    val uploadMs: Long? = null,
    val totalMs: Long? = null,
)

object JobProgressStateCodec {
    private val json = Json { ignoreUnknownKeys = true }

    fun encode(state: JobProgressState): String = json.encodeToString(JobProgressState.serializer(), state)

    fun decode(raw: String): JobProgressState? =
        runCatching { json.decodeFromString(JobProgressState.serializer(), raw) }.getOrNull()
}

object JobRedisKeys {
    fun job(id: String): String = "downloader:job:$id"
    fun progress(id: String): String = "downloader:progress:$id"
    fun cancel(id: String): String = "downloader:cancel:$id"
}
