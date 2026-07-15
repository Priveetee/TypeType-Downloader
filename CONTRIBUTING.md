# Contributing to TypeType Downloader

Thank you for helping improve TypeType downloads.

## Scope

This repository owns download jobs, media transfer, stream selection, SABR segment consumption, muxing, progress reporting, cancellation, cleanup, and artifact storage.

Open bug reports and feature requests in the [central TypeType issue tracker](https://github.com/TypeType-Video/TypeType/issues). Mention that the problem affects `TypeType-Downloader` and link the issue from your pull request.

Extraction and provider metadata belong in [TypeType-Server](https://github.com/TypeType-Video/TypeType-Server). The frontend consumes downloads through the Server gateway and must not call this service directly.

## Set up the project

Use Go 1.26 with a C toolchain and the libavformat, libavcodec, and libavutil development libraries installed.

```sh
git switch dev
go mod download
go test ./...
```

The provided Compose stack starts PostgreSQL, Dragonfly, Garage, and the downloader for integration work.

## Programming preferences

- Use Go for production implementation and follow standard Go naming and package conventions.
- Prefer the standard library, adding a dependency only when it materially improves the result.
- Prefer clear names and structure over explanatory comments, but comments are welcome whenever a contributor finds them useful.
- Keep files under 160 lines where practical and split them before they become difficult to reason about.
- Keep extraction in TypeType-Server.
- Keep HTTP handlers thin and put behavior in the owning package.
- Pass `context.Context` through jobs, workers, downloads, storage, and muxing.
- Preserve cancellation and cleanup on every failure path.
- Use typed errors when API behavior depends on the failure class.
- Avoid global mutable state.
- Do not reintroduce yt-dlp.
- Keep stream-copy muxing as the default; do not introduce transcoding without prior discussion.
- Add tests for job lifecycle, retries, cancellation, storage, progress, and failure classification.
- Keep generated media, local databases, credentials, and test artifacts out of Git.
- Keep benchmarks and proof-of-concept code out of production paths unless they are behind explicit flags.
- Verify GPL-3.0-or-later compatibility before adding a dependency.
- Coordinate any API contract change with TypeType-Server.

## Required checks

```sh
test -z "$(gofmt -l cmd internal)"
go test ./...
go build ./...
bash -n scripts/e2e-beta.sh
bash -n scripts/bench-beta.sh
DB_PASSWORD=change-me S3_ACCESS_KEY=change-me S3_SECRET_KEY=change-me docker compose config --quiet
```

Run `gofmt -w` on changed Go files before these checks.

## Commits and pull requests

Create your branch from `dev` and open the pull request against `dev`.

Use commit messages in this form:

```text
type: short description
```

Common types are `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, and `style`. Use the imperative mood and keep the first line under 72 characters.

Describe the affected job stage, storage mode, tests, failure behavior, and any required Server update in the pull request.

Contributions to this repository are distributed under [GPL-3.0-or-later](LICENSE).
