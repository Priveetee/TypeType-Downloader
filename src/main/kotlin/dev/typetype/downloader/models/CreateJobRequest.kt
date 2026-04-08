package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class CreateJobRequest(val url: String)
