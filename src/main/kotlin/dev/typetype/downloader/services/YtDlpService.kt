package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import java.nio.file.Files
import java.nio.file.Path
import java.util.concurrent.TimeUnit

data class YtDlpResult(
    val title: String,
    val filePath: Path?,
    val error: String?,
)

class YtDlpService(private val config: AppConfig) {
    fun download(url: String, token: TokenPayload?): YtDlpResult {
        val workDir = Files.createTempDirectory("typetype-download-")
        val process = ProcessBuilder(buildCommand(url, workDir, token))
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
        return if (process.exitValue() == 0) {
            val filePath = Files.list(workDir).use { stream ->
                stream.filter { Files.isRegularFile(it) }.findFirst().orElse(null)
            }
            if (filePath == null) {
                deleteDirectory(workDir)
                YtDlpResult(title = "", filePath = null, error = "yt-dlp output file missing")
            } else {
                val title = output.firstOrNull { it.isNotBlank() } ?: filePath.fileName.toString()
                YtDlpResult(title = title, filePath = filePath, error = null)
            }
        } else {
            deleteDirectory(workDir)
            val error = output.lastOrNull { it.isNotBlank() } ?: "yt-dlp failed"
            YtDlpResult(title = "", filePath = null, error = error)
        }
    }

    private fun buildCommand(url: String, workDir: Path, token: TokenPayload?): List<String> {
        val command = mutableListOf(
            config.ytdlpBin,
            "--no-warnings",
            "--no-playlist",
            "--print",
            "title",
            "-f",
            "bv*+ba/b",
            "--merge-output-format",
            "mp4",
            "-o",
            "${workDir.toAbsolutePath()}/%(id)s.%(ext)s",
        )
        if (token != null) {
            command.add("--extractor-args")
            command.add(
                "youtube:player_client=web;po_token=web.gvs+${token.streamingPot};visitor_data=${token.visitorData}"
            )
        }
        command.add(url)
        return command
    }

    private fun deleteDirectory(dir: Path) {
        if (!Files.exists(dir)) return
        Files.walk(dir).sorted(Comparator.reverseOrder()).forEach { Files.deleteIfExists(it) }
    }
}
