package dev.typetype.downloader.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class DownloadMode {
    @SerialName("video")
    VIDEO,

    @SerialName("audio")
    AUDIO,
}
