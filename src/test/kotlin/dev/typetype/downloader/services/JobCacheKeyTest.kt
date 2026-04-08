package dev.typetype.downloader.services

import dev.typetype.downloader.models.DownloadMode
import dev.typetype.downloader.models.JobOptions
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotEquals

class JobCacheKeyTest {
    @Test
    fun `same URL yields same key`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        val options = JobOptionsCodec.encode(JobOptions())
        val first = JobCacheKey.from(url, options)
        val second = JobCacheKey.from(url, options)
        assertEquals(first, second)
    }

    @Test
    fun `trimmed URL keeps same key`() {
        val options = JobOptionsCodec.encode(JobOptions())
        val clean = JobCacheKey.from("https://www.youtube.com/watch?v=dQw4w9WgXcQ", options)
        val padded = JobCacheKey.from("  https://www.youtube.com/watch?v=dQw4w9WgXcQ  ", options)
        assertEquals(clean, padded)
    }

    @Test
    fun `different options yield different keys`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        val first = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptions(mode = DownloadMode.VIDEO)))
        val second = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptions(mode = DownloadMode.AUDIO)))
        assertNotEquals(first, second)
    }
}
