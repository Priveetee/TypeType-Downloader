package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class JobOptions(
    val mode: DownloadMode = DownloadMode.VIDEO,
    val quality: String = "best",
    val format: String = "",
    val sponsorBlock: Boolean = false,
    val sponsorBlockCategories: List<String> = emptyList(),
    val thumbnailOnly: Boolean = false,
    val subtitles: SubtitlesOptions = SubtitlesOptions(),
)
