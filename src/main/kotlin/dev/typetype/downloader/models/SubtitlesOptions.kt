package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class SubtitlesOptions(
    val enabled: Boolean = false,
    val auto: Boolean = false,
    val embed: Boolean = false,
    val languages: List<String> = listOf("en"),
    val format: String = "srt",
)
