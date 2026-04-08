package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

object JobOptionsCodec {
    private val json = Json { ignoreUnknownKeys = true; encodeDefaults = true }

    fun encode(options: JobOptions): String = json.encodeToString(JobOptions.serializer(), options)

    fun decode(raw: String): JobOptions = json.decodeFromString(JobOptions.serializer(), raw)

    fun encodeQueue(payload: QueuePayload): String = json.encodeToString(QueuePayload.serializer(), payload)

    fun decodeQueue(raw: String): QueuePayload? = runCatching {
        json.decodeFromString(QueuePayload.serializer(), raw)
    }.getOrNull()

    @Serializable
    data class QueuePayload(
        val id: String,
        val options: JobOptions,
    )
}
