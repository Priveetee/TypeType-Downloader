plugins {
    kotlin("jvm") version "2.2.20"
    kotlin("plugin.serialization") version "2.2.20"
    application
    id("jacoco")
}

group = "dev.typetype"
version = "0.1.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation("io.ktor:ktor-server-core-jvm:3.4.1")
    implementation("io.ktor:ktor-server-netty-jvm:3.4.1")
    implementation("io.ktor:ktor-server-call-logging-jvm:3.4.1")
    implementation("io.ktor:ktor-server-content-negotiation-jvm:3.4.1")
    implementation("io.ktor:ktor-serialization-kotlinx-json-jvm:3.4.1")
    implementation("io.ktor:ktor-server-status-pages-jvm:3.4.1")
    implementation("org.postgresql:postgresql:42.7.8")
    implementation("com.zaxxer:HikariCP:6.3.0")
    implementation("redis.clients:jedis:6.1.0")
    implementation("software.amazon.awssdk:s3:2.31.69")
    implementation("ch.qos.logback:logback-classic:1.5.20")
    testImplementation(kotlin("test"))
    testImplementation("io.mockk:mockk:1.13.12")
}

application {
    mainClass = "dev.typetype.downloader.MainKt"
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)
}

jacoco {
    toolVersion = "0.8.12"
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required.set(true)
        html.required.set(true)
        csv.required.set(false)
    }
}

tasks.jacocoTestCoverageVerification {
    dependsOn(tasks.test)
    violationRules {
        rule {
            limit {
                counter = "LINE"
                value = "COVEREDRATIO"
                minimum = "0.05".toBigDecimal()
            }
        }
    }
}

tasks.check {
    dependsOn(tasks.jacocoTestCoverageVerification)
}
