package dev.typetype.downloader.services

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotEquals

class JobCacheKeyTest {
    @Test
    fun `same URL yields same key`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        val first = JobCacheKey.fromUrl(url)
        val second = JobCacheKey.fromUrl(url)
        assertEquals(first, second)
    }

    @Test
    fun `trimmed URL keeps same key`() {
        val clean = JobCacheKey.fromUrl("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
        val padded = JobCacheKey.fromUrl("  https://www.youtube.com/watch?v=dQw4w9WgXcQ  ")
        assertEquals(clean, padded)
    }

    @Test
    fun `different URLs yield different keys`() {
        val first = JobCacheKey.fromUrl("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
        val second = JobCacheKey.fromUrl("https://www.youtube.com/watch?v=oHg5SJYRHA0")
        assertNotEquals(first, second)
    }
}
