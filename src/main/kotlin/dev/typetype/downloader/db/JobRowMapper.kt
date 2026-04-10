package dev.typetype.downloader.db

import dev.typetype.downloader.models.JobStatus
import java.sql.ResultSet

fun rowFrom(rs: ResultSet): JobRow = JobRow(
    id = rs.getString("id"),
    url = rs.getString("source_url"),
    cacheKey = rs.getString("cache_key"),
    optionsJson = rs.getString("options_json"),
    status = JobStatus.valueOf(rs.getString("status")),
    durationMs = rs.getLong("duration_ms"),
    title = rs.getString("title"),
    error = rs.getString("error"),
    artifactKey = rs.getString("artifact_key"),
    artifactExpiresAt = rs.getTimestamp("artifact_expires_at")?.toInstant(),
    createdAt = rs.getTimestamp("created_at")?.toInstant(),
    startedAt = rs.getTimestamp("started_at")?.toInstant(),
    finishedAt = rs.getTimestamp("finished_at")?.toInstant(),
)
