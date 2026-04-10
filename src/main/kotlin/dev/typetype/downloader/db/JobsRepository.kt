package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus
import java.sql.Timestamp
import java.sql.Types
import java.time.Instant

class JobsRepository {
    private val columns = "id, source_url, cache_key, options_json, status, duration_ms, title, error, artifact_key, artifact_expires_at, created_at, started_at, finished_at"

    fun insertQueued(id: String, url: String, cacheKey: String, optionsJson: String) {
        val sql = "INSERT INTO jobs (id, source_url, cache_key, options_json, status, duration_ms, title, error) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
        Database.withConnection { connection ->
            connection.prepareStatement(sql).use {
                it.setString(1, id); it.setString(2, url); it.setString(3, cacheKey); it.setString(4, optionsJson)
                it.setString(5, JobStatus.QUEUED.name); it.setLong(6, 0L); it.setString(7, ""); it.setString(8, null)
                it.executeUpdate()
            }
        }
    }

    fun insertDoneFromCache(id: String, url: String, optionsJson: String, cached: JobRow) {
        val sql = "INSERT INTO jobs (id, source_url, cache_key, options_json, status, duration_ms, title, error, artifact_key, artifact_expires_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        Database.withConnection { connection ->
            connection.prepareStatement(sql).use {
                it.setString(1, id); it.setString(2, url); it.setString(3, cached.cacheKey); it.setString(4, optionsJson)
                it.setString(5, JobStatus.DONE.name); it.setLong(6, 0L); it.setString(7, cached.title); it.setString(8, null)
                it.setString(9, cached.artifactKey); setInstant(it, 10, cached.artifactExpiresAt); it.executeUpdate()
            }
        }
    }

    fun findReusableByCacheKey(cacheKey: String): JobRow? = selectOne(
        "SELECT $columns FROM jobs WHERE cache_key = ? AND status = ? AND artifact_key IS NOT NULL AND artifact_expires_at > NOW() ORDER BY finished_at DESC NULLS LAST LIMIT 1",
    ) { it.setString(1, cacheKey); it.setString(2, JobStatus.DONE.name) }

    fun getById(id: String): JobRow? = selectOne("SELECT $columns FROM jobs WHERE id = ?") { it.setString(1, id) }

    fun listQueuedOrRunning(): List<JobRow> = selectMany(
        "SELECT $columns FROM jobs WHERE status = ? OR status = ? ORDER BY created_at ASC",
    ) { it.setString(1, JobStatus.QUEUED.name); it.setString(2, JobStatus.RUNNING.name) }

    fun markRunningIfQueued(id: String): Boolean = update(
        "UPDATE jobs SET status = ?, started_at = NOW(), error = NULL WHERE id = ? AND status = ?",
    ) { it.setString(1, JobStatus.RUNNING.name); it.setString(2, id); it.setString(3, JobStatus.QUEUED.name) } > 0

    fun markCancelled(id: String): Boolean = update(
        "UPDATE jobs SET status = ?, error = ?, finished_at = NOW() WHERE id = ? AND status IN (?, ?)",
    ) { it.setString(1, JobStatus.FAILED.name); it.setString(2, "job cancelled"); it.setString(3, id); it.setString(4, JobStatus.QUEUED.name); it.setString(5, JobStatus.RUNNING.name) } > 0

    fun deleteIfNotRunning(id: String): JobRow? = selectOne(
        "DELETE FROM jobs WHERE id = ? AND status <> ? RETURNING $columns",
    ) { it.setString(1, id); it.setString(2, JobStatus.RUNNING.name) }

    fun markFinishedIfRunning(id: String, status: JobStatus, durationMs: Long, title: String, error: String?, artifactKey: String?, artifactExpiresAt: Instant?): Boolean = update(
        "UPDATE jobs SET status = ?, duration_ms = ?, title = ?, error = ?, artifact_key = ?, artifact_expires_at = ?, finished_at = NOW() WHERE id = ? AND status = ?",
    ) {
        it.setString(1, status.name); it.setLong(2, durationMs); it.setString(3, title); it.setString(4, error)
        it.setString(5, artifactKey); setInstant(it, 6, artifactExpiresAt); it.setString(7, id); it.setString(8, JobStatus.RUNNING.name)
    } > 0

    fun resetRunningToQueued(): Int = update(
        "UPDATE jobs SET status = ?, error = NULL WHERE status = ?",
    ) { it.setString(1, JobStatus.QUEUED.name); it.setString(2, JobStatus.RUNNING.name) }

    private fun update(sql: String, bind: (java.sql.PreparedStatement) -> Unit): Int = Database.withConnection { connection ->
        connection.prepareStatement(sql).use { bind(it); it.executeUpdate() }
    }

    private fun selectOne(sql: String, bind: (java.sql.PreparedStatement) -> Unit): JobRow? = Database.withConnection { connection ->
        connection.prepareStatement(sql).use { bind(it); it.executeQuery().use { rs -> if (rs.next()) rowFrom(rs) else null } }
    }

    private fun selectMany(sql: String, bind: (java.sql.PreparedStatement) -> Unit): List<JobRow> = Database.withConnection { connection ->
        connection.prepareStatement(sql).use {
            bind(it)
            it.executeQuery().use { rs ->
                val rows = mutableListOf<JobRow>()
                while (rs.next()) rows += rowFrom(rs)
                rows
            }
        }
    }

    private fun setInstant(statement: java.sql.PreparedStatement, index: Int, value: Instant?) {
        if (value == null) {
            statement.setNull(index, Types.TIMESTAMP_WITH_TIMEZONE)
            return
        }
        statement.setTimestamp(index, Timestamp.from(value))
    }
}
