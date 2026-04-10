package dev.typetype.downloader.services

import java.nio.file.Files
import kotlin.test.Test
import kotlin.test.assertFalse

class FileTreeCleanerTest {
    @Test
    fun `deleteDirectory removes nested tree`() {
        val root = Files.createTempDirectory("typetype-cleaner-test")
        val nested = Files.createDirectories(root.resolve("a/b/c"))
        Files.writeString(nested.resolve("file.txt"), "x")
        FileTreeCleaner.deleteDirectory(root)
        assertFalse(Files.exists(root))
    }
}
