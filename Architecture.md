# TypeType-Downloader Architecture

TypeType-Downloader is an asynchronous media download service for TypeType.

## Responsibilities

- Accept download jobs from TypeType-Server
- Execute `yt-dlp`/`ffmpeg` work asynchronously
- Persist durable job state in PostgreSQL
- Use Dragonfly for queueing and short-lived worker state
- Store generated artifacts in S3-compatible object storage (Garage)

## Runtime Components

- API: Kotlin/Ktor service exposing `/health`, `POST /jobs`, `GET /jobs/{id}`
- Worker: background consumer polling queue key in Dragonfly
- Database: PostgreSQL for jobs, options, artifacts, and cache mappings
- Queue/cache: Dragonfly (Redis protocol) for enqueue/dequeue and transient status
- Object storage: Garage S3 for artifact caching and signed downloads
- Token integration: TypeType-Token for YouTube poToken/visitorData requests

## Job Lifecycle

1. API receives a job request and validates options.
2. API computes a cache key from source URL and normalized options.
3. If a non-expired artifact exists in Garage, API returns cache hit metadata.
4. Otherwise, API creates a DB row with `queued` status and enqueues the job.
5. Worker pops the job, sets status to `running`, and executes tools.
6. Worker uploads artifact to Garage and updates DB to `done` or `failed`.
7. API status endpoint returns one of `queued|running|done|failed`.

## Storage Strategy

- Cache artifacts in Garage for 2 hours
- Prefer direct signed URL delivery to clients
- Keep local temporary files ephemeral and delete after upload

## Operational Notes

- No production credentials in repository files
- All limits and toggles are controlled via environment variables
- Deployments follow TypeType pattern: build in CI, pull image, restart stack
