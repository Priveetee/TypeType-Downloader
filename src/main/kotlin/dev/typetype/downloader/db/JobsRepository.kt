package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus
import java.time.Instant

class JobsRepository {
    fun insertQueued(id: String, url: String, cacheKey: String) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                """
                INSERT INTO jobs (id, source_url, cache_key, status, duration_ms, title, error)
                VALUES (?, ?, ?, ?, ?, ?, ?)
                """.trimIndent()
            ).use { statement ->
                statement.setString(1, id)
                statement.setString(2, url)
                statement.setString(3, cacheKey)
                statement.setString(4, JobStatus.QUEUED.name)
                statement.setLong(5, 0L)
                statement.setString(6, "")
                statement.setString(7, null)
                statement.executeUpdate()
            }
        }
    }
    fun insertDoneFromCache(id: String, url: String, cached: JobRow) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                """
                INSERT INTO jobs (id, source_url, cache_key, status, duration_ms, title, error, artifact_key, artifact_expires_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """.trimIndent()
            ).use { statement ->
                statement.setString(1, id)
                statement.setString(2, url)
                statement.setString(3, cached.cacheKey)
                statement.setString(4, JobStatus.DONE.name)
                statement.setLong(5, 0L)
                statement.setString(6, cached.title)
                statement.setString(7, null)
                statement.setString(8, cached.artifactKey)
                statement.setObject(9, cached.artifactExpiresAt)
                statement.executeUpdate()
            }
        }
    }
    fun findReusableByCacheKey(cacheKey: String): JobRow? = Database.withConnection { connection ->
        connection.prepareStatement(
            """
            SELECT id, source_url, cache_key, status, duration_ms, title, error, artifact_key, artifact_expires_at
            FROM jobs
            WHERE cache_key = ? AND status = ? AND artifact_key IS NOT NULL AND artifact_expires_at > NOW()
            ORDER BY finished_at DESC NULLS LAST
            LIMIT 1
            """.trimIndent()
        ).use { statement ->
            statement.setString(1, cacheKey)
            statement.setString(2, JobStatus.DONE.name)
            statement.executeQuery().use { rs ->
                if (!rs.next()) return@use null
                rowFrom(rs)
            }
        }
    }
    fun getById(id: String): JobRow? = Database.withConnection { connection ->
        connection.prepareStatement(
            """
            SELECT id, source_url, cache_key, status, duration_ms, title, error, artifact_key, artifact_expires_at
            FROM jobs
            WHERE id = ?
            """.trimIndent()
        ).use { statement ->
            statement.setString(1, id)
            statement.executeQuery().use { rs ->
                if (!rs.next()) return@use null
                rowFrom(rs)
            }
        }
    }
    fun markRunning(id: String) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                "UPDATE jobs SET status = ?, started_at = NOW() WHERE id = ?"
            ).use { statement ->
                statement.setString(1, JobStatus.RUNNING.name)
                statement.setString(2, id)
                statement.executeUpdate()
            }
        }
    }
    fun markFinished(
        id: String,
        status: JobStatus,
        durationMs: Long,
        title: String,
        error: String?,
        artifactKey: String?,
        artifactExpiresAt: Instant?,
    ) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                """
                UPDATE jobs
                SET status = ?, duration_ms = ?, title = ?, error = ?, artifact_key = ?, artifact_expires_at = ?, finished_at = NOW()
                WHERE id = ?
                """.trimIndent()
            ).use { statement ->
                statement.setString(1, status.name)
                statement.setLong(2, durationMs)
                statement.setString(3, title)
                statement.setString(4, error)
                statement.setString(5, artifactKey)
                statement.setObject(6, artifactExpiresAt)
                statement.setString(7, id)
                statement.executeUpdate()
            }
        }
    }
}
