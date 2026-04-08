package dev.typetype.downloader.services

import kotlin.test.Test
import kotlin.test.assertEquals

class SourceUrlResolverTest {
    @Test
    fun `keeps plain youtube url unchanged`() {
        val url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
        assertEquals(url, SourceUrlResolver.resolve(url))
    }

    @Test
    fun `unwraps encoded watch wrapper`() {
        val wrapped = "https://watch.invalid/watch?v=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3Dg3AI_ZkMv1I"
        val expected = "https://www.youtube.com/watch?v=g3AI_ZkMv1I"
        assertEquals(expected, SourceUrlResolver.resolve(wrapped))
    }

    @Test
    fun `unwraps encoded player wrapper`() {
        val wrapped = "https://player.invalid/watch?v=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3D7fqWBwErkxA"
        val expected = "https://www.youtube.com/watch?v=7fqWBwErkxA"
        assertEquals(expected, SourceUrlResolver.resolve(wrapped))
    }
}
