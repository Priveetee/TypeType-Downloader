FROM golang:1.26-trixie AS build

ARG BUILD_VERSION=1.2.4-dev
ARG BUILD_REVISION=development
ARG BUILD_TIME=unknown
WORKDIR /src
RUN apt-get update \
    && apt-get install -y --no-install-recommends pkg-config libavformat-dev libavcodec-dev libavutil-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w -X typetype-downloader-go/internal/buildinfo.Version=$BUILD_VERSION -X typetype-downloader-go/internal/buildinfo.Revision=$BUILD_REVISION -X typetype-downloader-go/internal/buildinfo.BuildTime=$BUILD_TIME" -o /out/typetype-downloader-go ./cmd/server

FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg libavformat61 libavcodec61 libavutil59 \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --system --create-home --home-dir /app typetype
WORKDIR /app
COPY --from=build /out/typetype-downloader-go /usr/local/bin/typetype-downloader-go
RUN mkdir -p /app/data && chown -R typetype:typetype /app

USER typetype
EXPOSE 18093
ENTRYPOINT ["/usr/local/bin/typetype-downloader-go"]
