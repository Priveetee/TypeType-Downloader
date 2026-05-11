# TypeType Downloader

Native Go downloader backend for TypeType.

This service replaces the old Kotlin downloader. It receives download jobs,
downloads direct media streams exposed by TypeType-Server, muxes audio/video
with libavformat, and serves the final artifact through local storage or S3.

## Why This Exists

| Before | Now |
|---|---|
| Kotlin service around external tooling | Native Go backend |
| Downloader also carried extraction-era assumptions | Strict download-only boundary |
| Heavy runtime | Small Wolfi image |
| Less control over progress and artifacts | Native job lifecycle, SSE, S3 redirects |

TypeType-Server extracts stream metadata. This service downloads and muxes.
That boundary is intentional.

## Features

| Feature | Status |
|---|---|
| Async download jobs | Yes |
| Progress API | Yes |
| SSE job events | Yes |
| Parallel HTTP Range downloads | Yes |
| Audio-only downloads | Yes |
| MP4/WebM stream selection | Yes |
| libavformat stream-copy muxing | Yes |
| Local artifact storage | Yes |
| S3/Garage artifact storage | Yes |
| Postgres job persistence | Yes |
| Dragonfly status cache | Yes |
| yt-dlp runtime dependency | No |

## Docker

The Docker image is the normal production runtime for this repository.

```sh
docker build -f Dockerfile.wolfi -t typetype-downloader-go:wolfi .
```

Run with local Compose:

```sh
DB_PASSWORD=change-me \
S3_ACCESS_KEY=change-me \
S3_SECRET_KEY=change-me \
docker compose up -d --build
```

The service listens on port `18093`.

## Configuration

| Variable | Purpose |
|---|---|
| `HTTP_PORT` | HTTP port, default `18093` |
| `PUBLIC_BASE_URL` | Public base URL used in job responses |
| `TYPETYPE_API_BASE` | TypeType-Server base URL for `/streams` |
| `DB_URL` or `DATABASE_URL` | Postgres connection |
| `DB_USER`, `DB_PASSWORD` | Postgres credentials when using `DB_URL` |
| `REDIS_HOST`, `REDIS_PORT` | Dragonfly/Redis-compatible cache |
| `STORAGE_BACKEND` | `local` or `s3` |
| `S3_ENDPOINT` | Internal S3/Garage endpoint used by the service |
| `S3_PUBLIC_ENDPOINT` | Public endpoint used for presigned artifact URLs |
| `S3_BUCKET` | Artifact bucket |
| `S3_ACCESS_KEY`, `S3_SECRET_KEY` | S3 credentials |
| `MAX_CONCURRENT_WORKERS` | Number of concurrent jobs |
| `DOWNLOAD_WORKERS` | Range workers per stream |
| `DOWNLOAD_CHUNK_SIZE` | Range chunk size in bytes |
| `MUXER` | `avformat` by default |

Use `.env.example` as a starting point for local development.

## Storage

| Mode | Behavior |
|---|---|
| `local` | Keeps artifacts under `DATA_DIR/artifacts/<job-id>/` |
| `s3` | Uploads artifacts to S3/Garage and returns presigned redirects |

`S3_ENDPOINT` and `S3_PUBLIC_ENDPOINT` are separate on purpose.

| Endpoint | Used for |
|---|---|
| `S3_ENDPOINT` | Uploads, deletes, health checks from inside the service |
| `S3_PUBLIC_ENDPOINT` | URLs returned to clients for artifact downloads |

This avoids pushing multi-GB artifacts through an HTTP gateway when Garage can
serve them directly.

## API

| Method | Path | Description |
|---|---|---|
| `POST` | `/jobs` | Create a download job |
| `GET` | `/jobs/{id}` | Read job status |
| `GET` | `/jobs/{id}/events` | Stream job updates as SSE |
| `GET` | `/jobs/{id}/artifact` | Download or redirect to artifact |
| `POST` | `/jobs/{id}/cancel` | Cancel queued/running job |
| `DELETE` | `/jobs/{id}` | Delete non-running job and artifact |
| `GET` | `/health` | Basic health check |
| `GET` | `/health/deep` | Service, Postgres, Dragonfly and S3 health |

Create a job:

```sh
curl -sS -X POST http://localhost:18093/jobs \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","options":{"container":"mp4","height":1080}}'
```

Download the finished artifact:

```sh
curl -L -o output.mp4 http://localhost:18093/jobs/<id>/artifact
```

## Benchmarks

Recent Go results from the isolated Beta stack on Wolfi:

| Media | Job time |
|---|---:|
| Small 1080p video `dQw4w9WgXcQ` | `782 ms` |
| Small audio `dQw4w9WgXcQ` | `262 ms` |
| 10h audio `AKeUssuu3Is` | `6033 ms` |
| 10h 1080p video `AKeUssuu3Is` | `141652 ms` |

Run the Go benchmark script:

```sh
./scripts/bench-beta.sh
```

Large checks are opt-in:

```sh
RUN_LARGE=1 ./scripts/bench-beta.sh
```

## Development

```sh
gofmt -w cmd internal
go test ./...
go build ./...
```

Optional end-to-end check:

```sh
./scripts/e2e-beta.sh
```

## License

TypeType Downloader is licensed under `GPL-3.0-or-later`.

The current Docker runtime links against Wolfi FFmpeg packages whose metadata
includes GPL components. Third-party details are documented here:

https://github.com/Priveetee/TypeType-Downloader/blob/main/THIRD_PARTY_NOTICES.md
