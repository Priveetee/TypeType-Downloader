package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus

class JobsRepository {
    fun insertQueued(id: String, url: String) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                "INSERT INTO jobs (id, source_url, status, duration_ms, title, error) VALUES (?, ?, ?, ?, ?, ?)"
            ).use { statement ->
                statement.setString(1, id)
                statement.setString(2, url)
                statement.setString(3, JobStatus.QUEUED.name)
                statement.setLong(4, 0L)
                statement.setString(5, "")
                statement.setString(6, null)
                statement.executeUpdate()
            }
        }
    }

    fun getById(id: String): JobRow? = Database.withConnection { connection ->
        connection.prepareStatement(
            "SELECT id, source_url, status, duration_ms, title, error FROM jobs WHERE id = ?"
        ).use { statement ->
            statement.setString(1, id)
            statement.executeQuery().use { rs ->
                if (!rs.next()) return@use null
                JobRow(
                    id = rs.getString("id"),
                    url = rs.getString("source_url"),
                    status = JobStatus.valueOf(rs.getString("status")),
                    durationMs = rs.getLong("duration_ms"),
                    title = rs.getString("title"),
                    error = rs.getString("error"),
                )
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

    fun markFinished(id: String, status: JobStatus, durationMs: Long, title: String, error: String?) {
        Database.withConnection { connection ->
            connection.prepareStatement(
                "UPDATE jobs SET status = ?, duration_ms = ?, title = ?, error = ?, finished_at = NOW() WHERE id = ?"
            ).use { statement ->
                statement.setString(1, status.name)
                statement.setLong(2, durationMs)
                statement.setString(3, title)
                statement.setString(4, error)
                statement.setString(5, id)
                statement.executeUpdate()
            }
        }
    }
}
