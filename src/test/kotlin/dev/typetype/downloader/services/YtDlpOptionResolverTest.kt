package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import kotlin.test.Test
import kotlin.test.assertEquals

class YtDlpOptionResolverTest {
    @Test
    fun `resolves audio selector and format defaults`() {
        assertEquals("worstaudio/worst", YtDlpOptionResolver.audioSelector("worst"))
        assertEquals("bestaudio/best", YtDlpOptionResolver.audioSelector("1080p"))
        assertEquals("mp3", YtDlpOptionResolver.audioFormat("avi"))
        assertEquals("m4a", YtDlpOptionResolver.audioFormat(" M4A "))
    }

    @Test
    fun `resolves video selector and format defaults`() {
        assertEquals("bv*[height<=720]+ba/b[height<=720]", YtDlpOptionResolver.videoSelector("720p"))
        assertEquals("bv*+ba/b", YtDlpOptionResolver.videoSelector("best"))
        assertEquals("mp4", YtDlpOptionResolver.videoFormat("unknown"))
        assertEquals("webm", YtDlpOptionResolver.videoFormat(" WEBM "))
    }

    @Test
    fun `preferred extensions prioritize requested normalized format`() {
        val audio = YtDlpOptionResolver.preferredExtensions(
            JobOptions(mode = DownloadMode.AUDIO, quality = "best", format = "opus"),
        )
        val video = YtDlpOptionResolver.preferredExtensions(
            JobOptions(mode = DownloadMode.VIDEO, quality = "best", format = "mkv"),
        )
        val thumbnail = YtDlpOptionResolver.preferredExtensions(JobOptions(thumbnailOnly = true))

        assertEquals("opus", audio.first())
        assertEquals("mkv", video.first())
        assertEquals(listOf("jpg", "jpeg", "png", "webp"), thumbnail)
    }
}
