package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class ResolvedOutput(
    val videoItag: String? = null,
    val audioItag: String? = null,
    val height: Int? = null,
    val fps: Int? = null,
    val videoCodec: String? = null,
    val audioCodec: String? = null,
    val container: String? = null,
    val bitrate: Int? = null,
    val fileName: String? = null,
    val effectiveVideoCodec: String? = null,
    val effectiveAudioCodec: String? = null,
    val effectiveFps: Double? = null,
    val effectiveFormatId: String? = null,
)
