package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import java.nio.file.Files
import java.nio.file.Path
import java.util.concurrent.TimeUnit

data class YtDlpResult(
    val title: String,
    val filePath: Path?,
    val error: String?,
)

class YtDlpService(private val config: AppConfig) {
    fun download(url: String, token: TokenPayload?, options: JobOptions): YtDlpResult {
        val workDir = Files.createTempDirectory("typetype-download-")
        val process = ProcessBuilder(buildCommand(url, workDir, token, options))
            .directory(workDir.toFile())
            .redirectErrorStream(true)
            .start()
        val finished = process.waitFor(config.ytdlpTimeoutSeconds, TimeUnit.SECONDS)
        if (!finished) {
            process.destroyForcibly()
            deleteDirectory(workDir)
            return YtDlpResult(title = "", filePath = null, error = "yt-dlp timeout")
        }
        val output = process.inputStream.bufferedReader().readLines()
        if (process.exitValue() != 0) {
            deleteDirectory(workDir)
            val error = output.lastOrNull { it.isNotBlank() } ?: "yt-dlp failed"
            return YtDlpResult(title = "", filePath = null, error = error)
        }
        val filePath = selectOutputFile(workDir, options)
        if (filePath == null) {
            deleteDirectory(workDir)
            return YtDlpResult(title = "", filePath = null, error = "yt-dlp output file missing")
        }
        val title = output.firstOrNull { isTitleLine(it) } ?: filePath.fileName.toString()
        return YtDlpResult(title = title, filePath = filePath, error = null)
    }

    private fun buildCommand(url: String, workDir: Path, token: TokenPayload?, options: JobOptions): List<String> {
        val command = mutableListOf(
            config.ytdlpBin,
            "--no-warnings",
            "--no-playlist",
            "--no-progress",
            "--print",
            "title",
            "-o",
            "${workDir.toAbsolutePath()}/%(id)s.%(ext)s",
        )
        when {
            options.thumbnailOnly -> command.addAll(listOf("--skip-download", "--write-thumbnail"))
            options.mode == DownloadMode.AUDIO -> command.addAll(listOf("-f", "bestaudio/best", "--extract-audio", "--audio-format", "mp3"))
            else -> command.addAll(listOf("-f", "bv*+ba/b", "--merge-output-format", "mp4"))
        }
        if (options.sponsorBlock && !options.thumbnailOnly) {
            command.addAll(listOf("--sponsorblock-remove", "sponsor,selfpromo,interaction,intro,outro,preview,filler,music_offtopic"))
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

    private fun selectOutputFile(workDir: Path, options: JobOptions): Path? {
        val files = Files.list(workDir).use { stream -> stream.filter { Files.isRegularFile(it) }.toList() }
        if (files.isEmpty()) return null
        val preferred = when {
            options.thumbnailOnly -> setOf("jpg", "jpeg", "png", "webp")
            options.mode == DownloadMode.AUDIO -> setOf("mp3", "m4a", "opus", "aac", "flac", "wav", "webm")
            else -> setOf("mp4", "mkv", "webm", "mov")
        }
        return files.firstOrNull { ext(it) in preferred } ?: files.maxByOrNull { Files.size(it) }
    }

    private fun isTitleLine(value: String): Boolean = value.isNotBlank() && !value.startsWith("[")

    private fun ext(path: Path): String = path.fileName.toString().substringAfterLast('.', "").lowercase()

    private fun deleteDirectory(dir: Path) {
        if (!Files.exists(dir)) return
        Files.walk(dir).sorted(Comparator.reverseOrder()).forEach { Files.deleteIfExists(it) }
    }
}
