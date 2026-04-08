package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class CreateJobResponse(
    val id: String,
    val cached: Boolean,
)
