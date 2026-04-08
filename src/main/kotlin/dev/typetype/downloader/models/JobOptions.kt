package dev.typetype.downloader.models

import kotlinx.serialization.Serializable

@Serializable
data class JobOptions(
    val mode: DownloadMode = DownloadMode.VIDEO,
    val sponsorBlock: Boolean = false,
    val thumbnailOnly: Boolean = false,
    val subtitles: SubtitlesOptions = SubtitlesOptions(),
)
