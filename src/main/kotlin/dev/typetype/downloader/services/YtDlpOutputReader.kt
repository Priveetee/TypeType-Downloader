package dev.typetype.downloader.services

import java.io.BufferedReader
import java.io.InputStream
import java.io.InputStreamReader
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.atomic.AtomicReference

data class YtDlpExecutionState(
    val title: String?,
    val progress: JobProgressState?,
    val lines: List<String>,
)

class YtDlpOutputReader(
    inputStream: InputStream,
    private val onProgress: (JobProgressState) -> Unit,
) {
    private val lines = CopyOnWriteArrayList<String>()
    private val title = AtomicReference<String?>(null)
    private val progress = AtomicReference<JobProgressState?>(null)
    private val thread = Thread {
        BufferedReader(InputStreamReader(inputStream)).use { reader ->
            while (true) {
                val line = reader.readLine() ?: break
                if (line.isNotBlank()) lines += line
                parseTitle(line)?.let { parsed ->
                    if (parsed.isNotBlank()) title.compareAndSet(null, parsed)
                }
                parseMeta(line)?.let { meta ->
                    val state = merge(progress.get(), meta)
                    progress.set(state)
                    onProgress(state)
                }
                YtDlpProgressParser.parse(line)?.let { parsed ->
                    val state = merge(progress.get(), parsed)
                    progress.set(state)
                    onProgress(state)
                }
            }
        }
    }

    fun start() {
        thread.name = "ytdlp-output-reader"
        thread.isDaemon = true
        thread.start()
    }

    fun await() {
        thread.join()
    }

    fun snapshot(): YtDlpExecutionState = YtDlpExecutionState(
        title = title.get(),
        progress = progress.get(),
        lines = lines.toList(),
    )

    private fun parseTitle(line: String): String? =
        line.removePrefix("TT_TITLE:").takeIf { line.startsWith("TT_TITLE:") }

    private fun parseMeta(line: String): YtDlpProgressUpdate? {
        if (!line.startsWith("TT_META:")) return null
        val payload = line.removePrefix("TT_META:")
        val parts = payload.split('|')
        val formatId = parts.getOrNull(0)?.takeIf { it.isNotBlank() }
        val videoCodec = parts.getOrNull(1)?.takeIf { it.isNotBlank() && it != "none" }
        val audioCodec = parts.getOrNull(2)?.takeIf { it.isNotBlank() && it != "none" }
        val fps = parts.getOrNull(3)?.toDoubleOrNull()
        val container = parts.getOrNull(4)?.takeIf { it.isNotBlank() }
        return YtDlpProgressUpdate(
            stage = "resolving",
            formatId = formatId,
            videoCodec = videoCodec,
            audioCodec = audioCodec,
            fps = fps,
            container = container,
        )
    }

    private fun merge(previous: JobProgressState?, update: YtDlpProgressUpdate): JobProgressState {
        return JobProgressState(
            stage = update.stage,
            progressPercent = update.progressPercent ?: previous?.progressPercent,
            downloadedBytes = update.downloadedBytes ?: previous?.downloadedBytes,
            totalBytes = update.totalBytes ?: previous?.totalBytes,
            etaSeconds = update.etaSeconds ?: previous?.etaSeconds,
            speedBytesPerSecond = update.speedBytesPerSecond ?: previous?.speedBytesPerSecond,
            videoCodec = update.videoCodec ?: previous?.videoCodec,
            audioCodec = update.audioCodec ?: previous?.audioCodec,
            fps = update.fps ?: previous?.fps,
            container = update.container ?: previous?.container,
            formatId = update.formatId ?: previous?.formatId,
        )
    }
}
