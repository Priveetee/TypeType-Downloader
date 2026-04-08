# TypeType-Downloader

The download backend for TypeType.

A Kotlin/Ktor service with asynchronous workers that wraps `yt-dlp` and `ffmpeg`, queues jobs with Dragonfly, stores durable state in PostgreSQL, and caches output artifacts in S3-compatible storage (Garage).

See [Architecture.md](./Architecture.md) for the runtime flow and component boundaries.

## Stack

| Role | Tool |
|---|---|
| Language | Kotlin |
| Server | Ktor (Netty engine) |
| Downloader | yt-dlp |
| Processing | ffmpeg |
| Build | Gradle (Kotlin DSL) |
| Job state | PostgreSQL via HikariCP |
| Queue/cache | Dragonfly (Redis-compatible) |
| Object storage | Garage (S3-compatible) |

## Development

### Prerequisites

- JDK 21+
- Docker and Docker Compose
- `yt-dlp` and `ffmpeg` (for non-Docker local runs)

### Start dependencies

```bash
cp .env.example .env
docker compose up -d postgres dragonfly garage
```

### Build

```bash
./gradlew build
```

### Run

```bash
./gradlew run
```

The server starts on port `18093` by default.

### Docker image tags (GHCR)

Container tags are published to GHCR with:

- stable image `${{ github.repository }}` on `main` and Git tags `v*`
- beta image `${{ github.repository }}-beta` on `dev`
- `sha-<short-sha>` on every build
- branch tags (`main` on stable image, `dev` on beta image)
- `latest` on default branch (stable image) and on `dev` (beta image)
- `beta` on `dev` (beta image)
- release tags when pushing Git tags like `v1.2.3` (`1.2.3` and `1.2`) on stable image

### Configuration

All configuration is via environment variables.

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `18093` | API listen port |
| `DB_URL` | `jdbc:postgresql://localhost:55432/typetype_downloader` | PostgreSQL JDBC URL |
| `DB_USER` | `typetype` | PostgreSQL user |
| `DB_PASSWORD` | `typetype` | PostgreSQL password |
| `REDIS_HOST` | `localhost` | Dragonfly hostname |
| `REDIS_PORT` | `56379` | Dragonfly port |
| `REDIS_QUEUE_KEY` | `downloader:queue` | Queue key for enqueued jobs |
| `MAX_CONCURRENT_WORKERS` | `2` | Worker count |
| `MAX_QUEUE_SIZE` | `100` | Queue saturation threshold |
| `JOB_TTL_SECONDS` | `600` | TTL for transient job cache entries |
| `YTDLP_BIN` | `yt-dlp` | yt-dlp executable path |
| `YTDLP_TIMEOUT_SECONDS` | `120` | Per-job yt-dlp timeout |
| `ENABLE_TRANSCODE` | `false` | Toggle ffmpeg transcode flows |
| `S3_ENDPOINT` | `http://localhost:3900` | Garage S3 endpoint |
| `S3_REGION` | `garage` | S3 region |
| `S3_BUCKET` | `typetype-downloads` | Bucket for output artifacts |
| `S3_ACCESS_KEY` | `change-me` | S3 access key |
| `S3_SECRET_KEY` | `change-me` | S3 secret key |
| `S3_ARTIFACT_TTL_SECONDS` | `7200` | Artifact cache TTL in seconds |
| `TOKEN_SERVICE_URL` | `http://localhost:8081` | TypeType-Token base URL |

## API

- `GET /health`
- `POST /jobs` accepts:
  - `url` (required)
  - `options.mode` (`video` or `audio`)
  - `options.sponsorBlock` (`true`/`false`)
  - `options.sponsorBlockCategories` (`sponsor,selfpromo,interaction,intro,outro,preview,filler,music_offtopic`)
  - `options.thumbnailOnly` (`true`/`false`)
  - `options.subtitles` (`enabled`, `auto`, `embed`, `languages`, `format`)
  and returns `{ "id": "...", "cached": false|true }`
- `GET /jobs/{id}` returns one of `queued|running|done|failed` and includes a signed `artifactUrl` when available
- `GET /jobs/{id}/artifact` redirects to signed Garage artifact URL when ready

Wrapper URLs are resolved automatically. For example, frontend watch wrappers such as
`https://watch.example/watch?v=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3D...` are normalized to the underlying source URL before processing.

## License

GPL v3 — same licensing direction as TypeType-Server.
