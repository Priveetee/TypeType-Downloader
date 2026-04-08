package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import java.util.concurrent.TimeUnit

data class YtDlpResult(val title: String, val error: String?)

class YtDlpService(private val config: AppConfig) {
    fun extractTitle(url: String): YtDlpResult {
        val process = ProcessBuilder(
            config.ytdlpBin,
            "--skip-download",
            "--print",
            "title",
            "--no-warnings",
            url,
        ).start()
        val finished = process.waitFor(config.ytdlpTimeoutSeconds, TimeUnit.SECONDS)
        if (!finished) {
            process.destroyForcibly()
            return YtDlpResult(title = "", error = "yt-dlp timeout")
        }
        val stdout = process.inputStream.bufferedReader().readText().trim()
        val stderr = process.errorStream.bufferedReader().readText().trim()
        return if (process.exitValue() == 0) {
            YtDlpResult(title = stdout, error = null)
        } else {
            val error = if (stderr.isNotBlank()) stderr else "yt-dlp failed"
            YtDlpResult(title = "", error = error)
        }
    }
}
