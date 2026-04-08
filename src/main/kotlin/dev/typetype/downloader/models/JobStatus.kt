package dev.typetype.downloader.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class JobStatus {
    @SerialName("queued")
    QUEUED,

    @SerialName("running")
    RUNNING,

    @SerialName("done")
    DONE,

    @SerialName("failed")
    FAILED,
}
