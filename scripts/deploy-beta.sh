#!/usr/bin/env bash
set -euo pipefail

image=ghcr.io/priveetee/typetype-downloader-beta:latest
name=typetype-beta-downloader
volume=typetype-beta-downloader-data
server=$(docker ps -q \
  --filter label=com.docker.compose.project=typetype-beta-stack \
  --filter label=com.docker.compose.service=typetype-server)
test -n "$server"
network=$(docker inspect "$server" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{end}}')
test -n "$network"

docker pull "$image"
docker volume create "$volume" >/dev/null
docker run --rm --user root --entrypoint /bin/sh \
  -v "$volume:/app/data" "$image" \
  -c 'chown -R typetype:typetype /app/data'
docker rm -f "$name" >/dev/null 2>&1 || true
docker run -d \
  --name "$name" \
  --network "$network" \
  --network-alias typetype-downloader \
  --restart unless-stopped \
  -e HTTP_PORT=18093 \
  -e PUBLIC_BASE_URL=/api/downloader \
  -e TYPETYPE_API_BASE=http://typetype-server:8080 \
  -e DATA_DIR=/app/data \
  -e STORAGE_BACKEND=local \
  -e MAX_CONCURRENT_WORKERS=2 \
  -e DOWNLOAD_WORKERS=8 \
  -v "$volume:/app/data" \
  "$image" >/dev/null

for attempt in $(seq 1 30); do
  if docker exec "$server" wget -q -O- http://typetype-downloader:18093/health; then
    exit 0
  fi
  sleep 1
done
docker logs "$name"
exit 1
