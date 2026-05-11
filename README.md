# TypeType Downloader Go

Native Go downloader service for TypeType.

The service receives download jobs over HTTP, asks TypeType-Server for direct stream metadata, downloads selected audio/video streams with parallel HTTP Range requests, and muxes the result with libavformat.

License: GPL-3.0-or-later. The current Docker runtime links against Wolfi FFmpeg packages whose metadata includes GPL components; see `THIRD_PARTY_NOTICES.md`.

## Run

```sh
go run ./cmd/server
```

Defaults:

- `HTTP_ADDR=:18093`
- `HTTP_PORT=18093` is also accepted for compatibility with the current TypeType stack
- `PUBLIC_BASE_URL=http://localhost:18093`
- `TYPETYPE_API_BASE=http://typetype-server:8080`
- `DATA_DIR=data`
- `DATABASE_URL=` disables Postgres persistence when empty
- `DB_URL`, `DB_USER`, and `DB_PASSWORD` are also accepted for compatibility with the current TypeType stack
- `REDIS_HOST=`, `REDIS_PORT=6379`, and `JOB_TTL_SECONDS=600` enable Dragonfly status/cache publishing
- `STORAGE_BACKEND=local`, set `s3` for Garage/S3 artifact storage
- `S3_ENDPOINT=` S3-compatible endpoint, with or without scheme
- `S3_PUBLIC_ENDPOINT=` optional endpoint used only for presigned artifact URLs
- `S3_REGION=garage`
- `S3_BUCKET=` artifact bucket
- `S3_ACCESS_KEY=` access key
- `S3_SECRET_KEY=` secret key
- `S3_USE_SSL=true`
- `S3_PATH_STYLE=true`
- `S3_ARTIFACT_TTL_SECONDS=7200`, or `ARTIFACT_URL_TTL_SECONDS` as an override
- `MAX_CONCURRENT_WORKERS=2`
- `DOWNLOAD_WORKERS=8`
- `DOWNLOAD_CHUNK_SIZE=10485760`
- `DOWNLOAD_RANGE_MODE=query`
- `MUXER=avformat`
- `DOWNLOAD_HTTP2=true`
- `MAX_QUEUE_SIZE=100`

## TypeType Beta Stack

The service accepts the environment contract used by `../TypeType/docker-compose.dev.yml` and `../TypeType/docker-compose.dev.beta-downloader.yml`:

- `HTTP_PORT`
- `DB_URL=jdbc:postgresql://postgres:5432/typetype_downloader`
- `DB_USER`, `DB_PASSWORD`
- `REDIS_HOST=dragonfly`, `REDIS_PORT=6379`, `REDIS_QUEUE_KEY`
- `MAX_CONCURRENT_WORKERS`, `MAX_QUEUE_SIZE`, `JOB_TTL_SECONDS`
- `S3_ENDPOINT=http://garage:3900`
- `S3_PUBLIC_ENDPOINT=http://localhost:3900` for the local Go Beta override
- `S3_REGION=garage`, `S3_BUCKET=typetype-downloads`
- `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_ARTIFACT_TTL_SECONDS`

When `TYPETYPE_API_BASE` is not set, the downloader calls `http://typetype-server:8080/streams`, matching the Beta Compose network.

Run the Beta stack with this Go downloader built locally:

```sh
cd ../TypeType
docker compose -p typetype-go-beta -f docker-compose.dev.yml -f docker-compose.dev.go-downloader.yml up -d --build typetype-downloader typetype-server typetype
```

Then run the end-to-end check through the frontend gateway:

```sh
cd ../TypeType-Downloader
./scripts/e2e-beta.sh
```

Use `BASE_URL=http://localhost:18080/downloader ./scripts/e2e-beta.sh` to bypass the frontend nginx and hit TypeType-Server directly.

## Storage

Local storage keeps completed files under `DATA_DIR/artifacts/<job-id>/` and serves them directly from `GET /jobs/{id}/artifact`.

S3/Garage mode uploads the final muxed file and deletes the local output copy. The artifact endpoint returns a short-lived presigned redirect.

```sh
S3_ENDPOINT=http://garage:3900 \
S3_PUBLIC_ENDPOINT=http://localhost:3900 \
S3_BUCKET=typetype-downloads \
S3_ACCESS_KEY=dev-access \
S3_SECRET_KEY=dev-secret \
go run ./cmd/server
```

`S3_ENDPOINT` is used for uploads and health checks from the downloader process. `S3_PUBLIC_ENDPOINT` is used for presigned download redirects. In local Beta, this keeps multi-GB artifacts out of the Kotlin gateway proxy and lets clients download directly from Garage.

## Benchmarks

Run the reproducible Beta benchmark script from this repository:

```sh
./scripts/bench-beta.sh
```

By default it runs the small video/audio Go jobs. Large 10 hour checks are opt-in:

```sh
RUN_LARGE=1 OUT_DIR=/home/ark/Project/TypeType-Downloader/bench-out ./scripts/bench-beta.sh
```

Recent isolated Beta Go results on Wolfi:

- small 1080p video `dQw4w9WgXcQ`: `782 ms`
- small audio `dQw4w9WgXcQ`: `262 ms`
- 10h audio `AKeUssuu3Is`: `6033 ms`
- 10h 1080p video `AKeUssuu3Is`: `141652 ms`

## Database

Set `DATABASE_URL` or the stack-compatible `DB_URL`/`DB_USER`/`DB_PASSWORD` to persist job metadata to Postgres. The service creates the `downloader_jobs` table automatically and restores completed jobs into the cache on startup.

```sh
DATABASE_URL='postgres://user:pass@localhost:5432/typetype_downloader?sslmode=disable' \
go run ./cmd/server
```

The SQL schema is also available in `migrations/001_jobs.sql`.

## Dragonfly

Set `REDIS_HOST` and `REDIS_PORT` to publish job status snapshots into Dragonfly with `JOB_TTL_SECONDS`. This keeps the downloader aligned with the existing TypeType stack while the authoritative job state remains in-memory plus optional Postgres persistence.

## Cache And Performance

Job requests are deduplicated by normalized URL plus options. If an identical completed job exists, `POST /jobs` returns the existing `id` with `cached: true`; if the job is already queued or running, the existing `id` is returned without enqueueing duplicate work.

The worker pool is bounded by `MAX_CONCURRENT_WORKERS`. Each active job downloads audio and video concurrently, and each stream uses `DOWNLOAD_WORKERS` parallel Range workers with a shared HTTP client per runner.

## API

Create a job:

```sh
curl -sS -X POST http://localhost:18093/jobs \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","options":{"container":"mp4","height":1080}}'
```

Check status:

```sh
curl -sS http://localhost:18093/jobs/<id>
```

Stream events:

```sh
curl -N http://localhost:18093/jobs/<id>/events
```

Download artifact when the job is done:

```sh
curl -L -o output.mp4 http://localhost:18093/jobs/<id>/artifact
```

Cancel a queued or running job:

```sh
curl -sS -X POST http://localhost:18093/jobs/<id>/cancel
```

Delete a non-running job:

```sh
curl -sS -X DELETE http://localhost:18093/jobs/<id>
```

## Development Checks

```sh
gofmt -w cmd internal
go test ./...
go build ./...
```
