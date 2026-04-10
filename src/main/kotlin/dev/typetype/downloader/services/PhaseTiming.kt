package dev.typetype.downloader.services

object PhaseTiming {
    fun elapsedMs(startedAtNs: Long): Long = (System.nanoTime() - startedAtNs) / 1_000_000
}
