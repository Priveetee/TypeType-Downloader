package dev.typetype.downloader.routes

import dev.typetype.downloader.models.CreateJobRequest
import dev.typetype.downloader.services.JobService
import io.ktor.http.HttpStatusCode
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get
import io.ktor.server.routing.post

fun Route.jobRoutes(jobService: JobService) {
    post("/jobs") {
        val body = call.receive<CreateJobRequest>()
        if (body.url.isBlank()) {
            return@post call.respond(HttpStatusCode.BadRequest, mapOf("error" to "url is required"))
        }
        val created = jobService.enqueue(body.url)
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
}
