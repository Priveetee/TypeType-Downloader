package dev.typetype.downloader.routes

import dev.typetype.downloader.models.CreateJobRequest
import dev.typetype.downloader.services.CancelJobResult
import dev.typetype.downloader.services.DeleteJobResult
import dev.typetype.downloader.services.JobService
import io.ktor.http.HttpStatusCode
import io.ktor.server.request.receive
import io.ktor.server.response.respondText
import io.ktor.server.response.respond
import io.ktor.server.response.respondRedirect
import io.ktor.server.response.respondTextWriter
import io.ktor.server.routing.delete
import io.ktor.server.routing.Route
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import kotlinx.coroutines.delay
import kotlinx.serialization.json.Json

fun Route.jobRoutes(jobService: JobService) {
    val json = Json
    post("/jobs") {
        val body = call.receive<CreateJobRequest>()
        if (body.url.isBlank()) {
            return@post call.respond(HttpStatusCode.BadRequest, mapOf("error" to "url is required"))
        }
        val created = jobService.enqueue(body.url, body.options)
        call.respond(HttpStatusCode.Created, created)
    }

    get("/jobs/{id}") {
        val id = call.parameters["id"] ?: return@get call.respond(
            HttpStatusCode.BadRequest,
            mapOf("error" to "id is required"),
        )
        val job = jobService.get(id) ?: return@get call.respond(
            HttpStatusCode.NotFound,
            mapOf("error" to "job not found"),
        )
        call.respond(HttpStatusCode.OK, job)
    }

    get("/jobs/{id}/artifact") {
        val id = call.parameters["id"] ?: return@get call.respond(
            HttpStatusCode.BadRequest,
            mapOf("error" to "id is required"),
        )
        val job = jobService.get(id) ?: return@get call.respond(
            HttpStatusCode.NotFound,
            mapOf("error" to "job not found"),
        )
        val artifactUrl = job.artifactUrl ?: return@get call.respond(
            HttpStatusCode.Conflict,
            mapOf("error" to "artifact not ready"),
        )
        call.respondRedirect(artifactUrl, permanent = false)
    }

    get("/jobs/{id}/events") {
        val id = call.parameters["id"] ?: return@get call.respond(
            HttpStatusCode.BadRequest,
            mapOf("error" to "id is required"),
        )
        call.respondTextWriter(contentType = io.ktor.http.ContentType.Text.EventStream) {
            while (true) {
                val job = jobService.get(id)
                if (job == null) {
                    write("event: error\n")
                    write("data: not_found\n\n")
                    flush()
                    break
                }
                write("event: progress\n")
                write("data: ${json.encodeToString(dev.typetype.downloader.models.JobResponse.serializer(), job)}\n\n")
                flush()
                if (job.status == dev.typetype.downloader.models.JobStatus.DONE || job.status == dev.typetype.downloader.models.JobStatus.FAILED) {
                    break
                }
                delay(1000)
            }
        }
    }

    post("/jobs/{id}/cancel") {
        val id = call.parameters["id"] ?: return@post call.respond(HttpStatusCode.BadRequest, mapOf("error" to "id is required"))
        when (jobService.cancel(id)) {
            CancelJobResult.CANCELLED -> call.respond(HttpStatusCode.Accepted, mapOf("status" to "cancelled"))
            CancelJobResult.NOT_CANCELLABLE -> call.respond(HttpStatusCode.Conflict, mapOf("error" to "job is not cancellable"))
            CancelJobResult.NOT_FOUND -> call.respond(HttpStatusCode.NotFound, mapOf("error" to "job not found"))
        }
    }

    delete("/jobs/{id}") {
        val id = call.parameters["id"] ?: return@delete call.respond(HttpStatusCode.BadRequest, mapOf("error" to "id is required"))
        when (jobService.delete(id)) {
            DeleteJobResult.DELETED -> call.respondText("", status = HttpStatusCode.NoContent)
            DeleteJobResult.CONFLICT_RUNNING -> call.respond(HttpStatusCode.Conflict, mapOf("error" to "job is running"))
            DeleteJobResult.NOT_FOUND -> call.respond(HttpStatusCode.NotFound, mapOf("error" to "job not found"))
        }
    }
}
