package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import redis.clients.jedis.JedisPooled
import java.util.concurrent.ConcurrentHashMap

class JobProgressStore(
    private val redis: JedisPooled,
    private val config: AppConfig,
) {
    private val snapshots = ConcurrentHashMap<String, Snapshot>()

    fun update(id: String, state: JobProgressState) {
        val now = System.currentTimeMillis()
        val previous = snapshots[id]
        if (previous != null && shouldSkip(previous.state, state, now - previous.writtenAtMs)) {
            snapshots[id] = previous.copy(state = state)
            return
        }
        snapshots[id] = Snapshot(state = state, writtenAtMs = now)
        redis.setex(JobRedisKeys.progress(id), config.jobTtlSeconds, JobProgressStateCodec.encode(state))
    }

    fun get(id: String): JobProgressState? {
        val raw = redis.get(JobRedisKeys.progress(id)) ?: return null
        return JobProgressStateCodec.decode(raw)
    }

    fun clear(id: String) {
        snapshots.remove(id)
        redis.del(JobRedisKeys.progress(id))
    }

    fun setCancelled(id: String) {
        redis.setex(JobRedisKeys.cancel(id), config.jobTtlSeconds, "1")
    }

    fun isCancelled(id: String): Boolean = redis.get(JobRedisKeys.cancel(id)) == "1"

    fun clearCancelled(id: String) {
        redis.del(JobRedisKeys.cancel(id))
    }

    private fun shouldSkip(previous: JobProgressState, next: JobProgressState, elapsedMs: Long): Boolean {
        if (previous.stage != next.stage) return false
        if (elapsedMs >= MIN_WRITE_INTERVAL_MS) return false
        if (previous.progressPercent != next.progressPercent) return false
        if (next.downloadedBytes != null && previous.downloadedBytes != null) {
            val delta = (next.downloadedBytes - previous.downloadedBytes).coerceAtLeast(0)
            if (delta >= MIN_BYTES_DELTA) return false
        } else if (previous.downloadedBytes != next.downloadedBytes) {
            return false
        }
        return previous.totalBytes == next.totalBytes &&
            previous.etaSeconds == next.etaSeconds &&
            previous.speedBytesPerSecond == next.speedBytesPerSecond &&
            previous.videoCodec == next.videoCodec &&
            previous.audioCodec == next.audioCodec &&
            previous.fps == next.fps &&
            previous.container == next.container &&
            previous.formatId == next.formatId
    }

    private data class Snapshot(
        val state: JobProgressState,
        val writtenAtMs: Long,
    )

    companion object {
        private const val MIN_WRITE_INTERVAL_MS = 400L
        private const val MIN_BYTES_DELTA = 256L * 1024L
    }
}
