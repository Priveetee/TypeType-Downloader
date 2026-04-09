package dev.typetype.downloader.services

import dev.typetype.downloader.models.JobOptions
import kotlin.test.Test
import kotlin.test.assertNotEquals

class JobCacheKeyTest {
    @Test
    fun `cache key isolates quality selections`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        val key720 = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptionsNormalizer.normalize(JobOptions(quality = "720p"))))
        val key1080 = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptionsNormalizer.normalize(JobOptions(quality = "1080p"))))
        assertNotEquals(key720, key1080)
    }

    @Test
    fun `cache key isolates exact itag selections`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        val key137 = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptionsNormalizer.normalize(JobOptions(videoItag = "137"))))
        val key136 = JobCacheKey.from(url, JobOptionsCodec.encode(JobOptionsNormalizer.normalize(JobOptions(videoItag = "136"))))
        assertNotEquals(key137, key136)
    }
}
