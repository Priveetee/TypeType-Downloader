package dev.typetype.downloader

import dev.typetype.downloader.config.AppConfigLoader
import dev.typetype.downloader.db.Database
import dev.typetype.downloader.db.JobsRepository
import dev.typetype.downloader.routes.healthRoutes
import dev.typetype.downloader.routes.jobRoutes
import dev.typetype.downloader.services.JobService
import dev.typetype.downloader.services.JobWorker
import dev.typetype.downloader.services.YtDlpService
import io.ktor.http.HttpStatusCode
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.Application
import io.ktor.server.application.ApplicationStopping
import io.ktor.server.application.install
import io.ktor.server.plugins.calllogging.CallLogging
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.plugins.statuspages.StatusPages
import io.ktor.server.response.respond
import io.ktor.server.routing.routing
import redis.clients.jedis.JedisPooled

fun Application.module() {
    val config = AppConfigLoader.load()
    Database.init(config)
    val redis = JedisPooled(config.redisHost, config.redisPort)
    val jobsRepository = JobsRepository()
    val ytDlpService = YtDlpService(config)
    val jobService = JobService(jobsRepository, redis, config)
    val worker = JobWorker(jobsRepository, redis, ytDlpService, config)
    worker.start()

    install(CallLogging)
    install(ContentNegotiation) { json() }
    install(StatusPages) {
        exception<IllegalArgumentException> { call, cause ->
            call.respond(HttpStatusCode.BadRequest, mapOf("error" to (cause.message ?: "bad request")))
        }
    }

    routing {
        healthRoutes()
        jobRoutes(jobService)
    }

    monitor.subscribe(ApplicationStopping) {
        redis.close()
        Database.close()
    }
}
