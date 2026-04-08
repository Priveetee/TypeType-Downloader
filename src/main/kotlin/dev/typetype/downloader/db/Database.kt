package dev.typetype.downloader.db

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import dev.typetype.downloader.config.AppConfig
import java.sql.Connection

object Database {
    private lateinit var dataSource: HikariDataSource

    fun init(config: AppConfig) {
        if (::dataSource.isInitialized) return
        val hikari = HikariConfig().apply {
            jdbcUrl = config.dbUrl
            username = config.dbUser
            password = config.dbPassword
            maximumPoolSize = 8
            minimumIdle = 1
        }
        dataSource = HikariDataSource(hikari)
        withConnection { connection ->
            connection.createStatement().use { statement ->
                statement.execute(
                    """
                    CREATE TABLE IF NOT EXISTS jobs (
                      id TEXT PRIMARY KEY,
                      source_url TEXT NOT NULL,
                      cache_key TEXT NOT NULL,
                      options_json TEXT NOT NULL DEFAULT '{}',
                      status TEXT NOT NULL,
                      duration_ms BIGINT NOT NULL,
                      title TEXT NOT NULL,
                      error TEXT,
                      artifact_key TEXT,
                      artifact_expires_at TIMESTAMPTZ,
                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                      started_at TIMESTAMPTZ,
                      finished_at TIMESTAMPTZ
                    )
                    """.trimIndent()
                )
                statement.execute("ALTER TABLE jobs ADD COLUMN IF NOT EXISTS cache_key TEXT")
                statement.execute("ALTER TABLE jobs ADD COLUMN IF NOT EXISTS options_json TEXT")
                statement.execute("ALTER TABLE jobs ADD COLUMN IF NOT EXISTS artifact_key TEXT")
                statement.execute("ALTER TABLE jobs ADD COLUMN IF NOT EXISTS artifact_expires_at TIMESTAMPTZ")
                statement.execute("UPDATE jobs SET cache_key = id WHERE cache_key IS NULL")
                statement.execute("UPDATE jobs SET options_json = '{}' WHERE options_json IS NULL")
                statement.execute("ALTER TABLE jobs ALTER COLUMN cache_key SET NOT NULL")
                statement.execute("ALTER TABLE jobs ALTER COLUMN options_json SET NOT NULL")
                statement.execute("CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)")
                statement.execute("CREATE INDEX IF NOT EXISTS idx_jobs_cache_key ON jobs(cache_key)")
                statement.execute("CREATE INDEX IF NOT EXISTS idx_jobs_artifact_expiry ON jobs(artifact_expires_at)")
            }
        }
    }

    fun <T> withConnection(block: (Connection) -> T): T = dataSource.connection.use(block)

    fun close() {
        if (::dataSource.isInitialized) {
            dataSource.close()
        }
    }
}
