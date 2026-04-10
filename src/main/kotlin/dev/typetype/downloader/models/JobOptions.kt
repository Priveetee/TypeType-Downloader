package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class JobOptions(
    val mode: DownloadMode = DownloadMode.VIDEO,
    val quality: String = "best",
    val format: String = "",
    val audioPassthrough: Boolean = false,
    val videoItag: String = "",
    val audioItag: String = "",
    val height: Int? = null,
    val fps: Int? = null,
    val videoCodec: String = "",
    val audioCodec: String = "",
    val bitrate: Int? = null,
    val allowQualityFallback: Boolean = false,
    val sponsorBlock: Boolean = false,
    val sponsorBlockCategories: List<String> = emptyList(),
    val thumbnailOnly: Boolean = false,
    val subtitles: SubtitlesOptions = SubtitlesOptions(),
)
