package dev.typetype.downloader

import dev.typetype.downloader.config.AppConfigLoader
import io.ktor.server.application.Application
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty

fun main() {
    val config = AppConfigLoader.load()
    embeddedServer(Netty, port = config.httpPort, module = Application::module).start(wait = true)
}
