package dev.typetype.downloader.services

import java.nio.file.Files
import java.nio.file.Path
import java.util.Comparator

object FileTreeCleaner {
    fun deleteDirectory(path: Path) {
        if (!Files.exists(path)) return
        Files.walk(path).use { stream ->
            stream.sorted(Comparator.reverseOrder()).forEach { Files.deleteIfExists(it) }
        }
    }
}
