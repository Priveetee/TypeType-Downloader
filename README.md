<div align="center">
  <img src="https://raw.githubusercontent.com/Priveetee/TypeType/main/assets/banner.svg" alt="TypeType" width="100%">
  <h1>TypeType-Downloader</h1>
  <p>Native Go download backend for TypeType.</p>
</div>

<div align="center">

[![Downloader](https://img.shields.io/badge/downloader-Go-00add8)](https://go.dev)
[![Muxer](https://img.shields.io/badge/muxer-libavformat-5cb85c)](https://ffmpeg.org)
[![License](https://img.shields.io/badge/license-GPL--3.0--or--later-blue)](LICENSE)

</div>

TypeType-Downloader receives jobs from TypeType-Server, downloads direct media streams, muxes audio and video, and exposes finished artifacts through local storage or S3-compatible storage. Extraction stays in TypeType-Server.

## What this is

A production Go service for video and audio downloads in the TypeType stack.

It provides async jobs, progress metadata, SSE events, parallel HTTP Range downloads, stream-copy muxing, Postgres persistence, Dragonfly cache state, and Garage/S3 artifact storage.

## What this is not

- Not an extractor.
- Not a frontend service.
- Not a yt-dlp wrapper.
- Not a public media proxy for arbitrary URLs.

## Stack

| Role | Tool |
|---|---|
| Language | Go |
| HTTP API | Go standard library |
| Downloads | Parallel HTTP Range requests |
| Muxing | libavformat through cgo |
| Storage | Local filesystem or S3/Garage |
| Persistence | PostgreSQL |
| Cache | Dragonfly |
| Runtime image | Wolfi-based Docker image |

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Basic health check |
| `GET` | `/health/deep` | Service, Postgres, Dragonfly, and S3 health |
| `POST` | `/jobs` | Create a download job |
| `GET` | `/jobs/{id}` | Read job status |
| `GET` | `/jobs/{id}/events` | Stream job updates as SSE |
| `GET` | `/jobs/{id}/artifact` | Download or redirect to artifact |
| `POST` | `/jobs/{id}/cancel` | Cancel queued or running job |
| `DELETE` | `/jobs/{id}` | Delete a non-running job and artifact |

## Job model

| Field | Meaning |
|---|---|
| `status` | Stable lifecycle: `queued`, `running`, `done`, `failed` |
| `stage` | Operational phase: `extract`, `download`, `mux`, `done`, or last known phase |
| `artifactUrl` | Relative gateway path when exposed through TypeType-Server |
| `resolved` | Selected stream metadata and output filename |

The terminal source of truth is `status`. Cancellation is represented as `status="failed"` with `errorCode="cancelled"`.

## Configuration

| Variable | Purpose |
|---|---|
| `HTTP_PORT` | HTTP port, default `18093` |
| `PUBLIC_BASE_URL` | Public base URL used in job responses |
| `TYPETYPE_API_BASE` | TypeType-Server base URL for `/streams` |
| `DATABASE_URL` or `DB_URL` | Postgres connection |
| `REDIS_HOST`, `REDIS_PORT` | Dragonfly connection |
| `STORAGE_BACKEND` | `local` or `s3` |
| `S3_ENDPOINT` | Internal S3/Garage endpoint |
| `S3_PUBLIC_ENDPOINT` | Public endpoint used for presigned URLs |
| `S3_BUCKET` | Artifact bucket |
| `S3_ACCESS_KEY`, `S3_SECRET_KEY` | S3 credentials |
| `MAX_CONCURRENT_WORKERS` | Concurrent jobs |
| `DOWNLOAD_WORKERS` | Range workers per stream |
| `DOWNLOAD_CHUNK_SIZE` | Range chunk size |
| `MUXER` | `avformat` by default |

## Storage

| Mode | Behavior |
|---|---|
| `local` | Keeps artifacts under `DATA_DIR/artifacts/<job-id>/` |
| `s3` | Uploads artifacts to S3/Garage and returns presigned redirects |

`S3_ENDPOINT` is used inside the container. `S3_PUBLIC_ENDPOINT` is used for URLs returned to clients.

## Development

```bash
gofmt -w cmd internal
go test ./...
go build ./...
```

Run the local Compose stack:

```bash
DB_PASSWORD=change-me \
S3_ACCESS_KEY=change-me \
S3_SECRET_KEY=change-me \
docker compose up -d --build
```

## Related projects

- [TypeType](https://github.com/Priveetee/TypeType) is the deployment stack.
- [TypeType-Server](https://github.com/Priveetee/TypeType-Server) extracts stream metadata and gates the downloader API.
- [TypeType web](https://github.com/Priveetee/TypeType) consumes download jobs from the frontend.

## License

GPL-3.0-or-later. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for runtime notices.
