FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app

COPY gradlew ./
COPY gradle/ ./gradle/
RUN ./gradlew --version --no-daemon -q

COPY build.gradle.kts settings.gradle.kts gradle.properties ./
RUN ./gradlew dependencies --no-daemon -q || true

COPY src/ ./src/
RUN ./gradlew installDist --no-daemon -q

FROM eclipse-temurin:21-jre-alpine AS runner
RUN apk add --no-cache yt-dlp ffmpeg
RUN addgroup -S typetype && adduser -S typetype -G typetype
WORKDIR /app

COPY --from=builder /app/build/install/TypeType-Downloader/ /app/

USER typetype
EXPOSE 18093
ENTRYPOINT ["/app/bin/TypeType-Downloader"]
