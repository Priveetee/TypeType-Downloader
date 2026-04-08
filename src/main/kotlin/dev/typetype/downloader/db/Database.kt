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
                      status TEXT NOT NULL,
                      duration_ms BIGINT NOT NULL,
                      title TEXT NOT NULL,
                      error TEXT,
                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                      started_at TIMESTAMPTZ,
                      finished_at TIMESTAMPTZ
                    )
                    """.trimIndent()
                )
                statement.execute("CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)")
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
