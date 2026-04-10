package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import java.nio.file.Path

object YtDlpCommandFactory {
    fun build(config: AppConfig, url: String, workDir: Path, token: TokenPayload?, options: JobOptions): List<String> {
        val command = mutableListOf(
            config.ytdlpBin,
            "--no-simulate",
            "--no-warnings",
            "--no-playlist",
            "--newline",
            "--print",
            "TT_TITLE:%(title)s",
            "--print",
            "TT_META:%(format_id)s|%(vcodec)s|%(acodec)s|%(fps)s|%(ext)s",
            "--concurrent-fragments",
            config.ytdlpConcurrentFragments.coerceAtLeast(1).toString(),
            "--retries",
            config.ytdlpRetries.coerceAtLeast(0).toString(),
            "--fragment-retries",
            config.ytdlpFragmentRetries.coerceAtLeast(0).toString(),
            "--socket-timeout",
            config.ytdlpSocketTimeoutSeconds.coerceAtLeast(1).toString(),
            "-o",
            "${workDir.toAbsolutePath()}/%(id)s.%(ext)s",
        )
        config.ytdlpHttpChunkSize.trim().takeIf { it.isNotBlank() }?.let {
            command.addAll(listOf("--http-chunk-size", it))
        }
        config.ytdlpExternalDownloader.trim().takeIf { it.isNotBlank() }?.let {
            command.addAll(listOf("--downloader", it))
        }
        config.ytdlpExternalDownloaderArgs.trim().takeIf { it.isNotBlank() }?.let {
            command.addAll(listOf("--downloader-args", it))
        }
        when {
            options.thumbnailOnly -> command.addAll(listOf("--skip-download", "--write-thumbnail"))
            options.mode == DownloadMode.AUDIO -> {
                val selector = YtDlpOptionResolver.audioSelector(options)
                command.addAll(listOf("-f", selector))
                if (!options.audioPassthrough) {
                    val audioFormat = YtDlpOptionResolver.audioFormat(options.format)
                    command.addAll(listOf("--extract-audio", "--audio-format", audioFormat))
                }
            }
            else -> {
                val selector = YtDlpOptionResolver.videoSelector(options)
                val videoFormat = YtDlpOptionResolver.videoFormat(options.format)
                command.addAll(listOf("-f", selector, "--merge-output-format", videoFormat))
            }
        }
        if (options.sponsorBlock && !options.thumbnailOnly) {
            command.addAll(listOf("--sponsorblock-remove", options.sponsorBlockCategories.joinToString(",")))
        }
        if (options.subtitles.enabled) {
            command.add("--write-subs")
            if (options.subtitles.auto) command.add("--write-auto-subs")
            command.addAll(listOf("--sub-langs", options.subtitles.languages.joinToString(",")))
            command.addAll(listOf("--sub-format", options.subtitles.format))
            if (options.subtitles.embed && options.mode == DownloadMode.VIDEO && !options.thumbnailOnly) command.add("--embed-subs")
        }
        if (token != null) {
            command.add("--extractor-args")
            command.add("youtube:player_client=web;po_token=web.gvs+${token.streamingPot};visitor_data=${token.visitorData}")
        }
        command.add(url)
        return command
    }
}
