package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import dev.typetype.downloader.models.JobOptions
import java.nio.file.Files
import java.nio.file.Path
import java.util.concurrent.TimeUnit

data class YtDlpResult(
    val title: String,
    val filePath: Path?,
    val error: String?,
    val progress: JobProgressState?,
) {
    fun withMetrics(tokenFetchMs: Long, ytdlpMs: Long, uploadMs: Long = 0): YtDlpResult {
        val current = progress ?: JobProgressState(stage = "running")
        val updated = current.copy(
            tokenFetchMs = tokenFetchMs,
            ytdlpMs = ytdlpMs,
            uploadMs = uploadMs,
            totalMs = tokenFetchMs + ytdlpMs + uploadMs,
        )
        return copy(progress = updated)
    }
}

class YtDlpService(private val config: AppConfig) {
    fun download(
        url: String,
        token: TokenPayload?,
        options: JobOptions,
        onProgress: (JobProgressState) -> Unit,
        shouldCancel: () -> Boolean = { false },
    ): YtDlpResult {
        val workDir = Files.createTempDirectory("typetype-download-")
        val process = ProcessBuilder(YtDlpCommandFactory.build(config, url, workDir, token, options))
            .directory(workDir.toFile())
            .redirectErrorStream(true)
            .start()
        val reader = YtDlpOutputReader(process.inputStream, onProgress)
        reader.start()
        val finished = waitFor(process, config.ytdlpTimeoutSeconds, shouldCancel)
        if (!finished) {
            process.destroyForcibly()
            reader.await()
            FileTreeCleaner.deleteDirectory(workDir)
            return YtDlpResult(title = "", filePath = null, error = "yt-dlp timeout", progress = reader.snapshot().progress)
        }
        if (shouldCancel()) {
            process.destroyForcibly()
            reader.await()
            FileTreeCleaner.deleteDirectory(workDir)
            return YtDlpResult(title = "", filePath = null, error = "job cancelled", progress = reader.snapshot().progress)
        }
        reader.await()
        val execution = reader.snapshot()
        val output = execution.lines
        if (process.exitValue() != 0) {
            FileTreeCleaner.deleteDirectory(workDir)
            val error = if (output.any { it.contains("Requested format is not available", ignoreCase = true) }) {
                "exact selection unavailable: requested format is not available"
            } else {
                output.lastOrNull { it.isNotBlank() } ?: "yt-dlp failed"
            }
            return YtDlpResult(title = "", filePath = null, error = error, progress = execution.progress)
        }
        val filePath = selectOutputFile(workDir, options)
        if (filePath == null) {
            FileTreeCleaner.deleteDirectory(workDir)
            return YtDlpResult(title = "", filePath = null, error = "yt-dlp output file missing", progress = execution.progress)
        }
        val title = execution.title ?: output.firstOrNull { isTitleLine(it) } ?: filePath.fileName.toString()
        return YtDlpResult(title = title, filePath = filePath, error = null, progress = execution.progress)
    }

    private fun selectOutputFile(workDir: Path, options: JobOptions): Path? {
        val files = Files.list(workDir).use { stream -> stream.filter { Files.isRegularFile(it) }.toList() }
        if (files.isEmpty()) return null
        val preferred = YtDlpOptionResolver.preferredExtensions(options)
        preferred.forEach { wanted ->
            files.firstOrNull { ext(it) == wanted }?.let { return it }
        }
        return files.maxByOrNull { Files.size(it) }
    }

    private fun isTitleLine(value: String): Boolean = value.isNotBlank() && !value.startsWith("[")
    private fun waitFor(process: Process, timeoutSeconds: Long, shouldCancel: () -> Boolean): Boolean {
        val deadline = System.nanoTime() + TimeUnit.SECONDS.toNanos(timeoutSeconds)
        while (System.nanoTime() < deadline) {
            if (shouldCancel()) return true
            if (process.waitFor(1, TimeUnit.SECONDS)) return true
        }
        return false
    }

    private fun ext(path: Path): String = path.fileName.toString().substringAfterLast('.', "").lowercase()
}
