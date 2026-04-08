package dev.typetype.downloader.routes

import io.ktor.http.HttpStatusCode
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get

fun Route.healthRoutes() {
    get("/health") {
        call.respond(HttpStatusCode.OK, mapOf("status" to "ok", "service" to "typetype-downloader"))
    }
}
