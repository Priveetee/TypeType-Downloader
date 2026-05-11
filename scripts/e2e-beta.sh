#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:18082/api/downloader}"
VIDEO_URL="${VIDEO_URL:-https://www.youtube.com/watch?v=dQw4w9WgXcQ}"
export BASE_URL VIDEO_URL

health="$(curl -fsS "$BASE_URL/health")"
printf 'health: %s\n' "$health"

create_body="$(python3 - <<'PY'
import json, os
print(json.dumps({
    "url": os.environ["VIDEO_URL"],
    "options": {"mode": "VIDEO", "quality": "1080p", "format": "mp4", "height": 1080, "allowQualityFallback": True},
}))
PY
)"

created="$(curl -fsS -X POST "$BASE_URL/jobs" -H 'Content-Type: application/json' -d "$create_body")"
printf 'created: %s\n' "$created"
job_id="$(CREATED="$created" python3 - <<'PY'
import json, os
print(json.loads(os.environ["CREATED"])["id"])
PY
)"

for _ in $(seq 1 180); do
  status_body="$(curl -fsS "$BASE_URL/jobs/$job_id")"
  status="$(STATUS_BODY="$status_body" python3 - <<'PY'
import json, os
print(json.loads(os.environ["STATUS_BODY"])["status"])
PY
)"
  printf 'status: %s\n' "$status"
  if [ "$status" = "done" ]; then
    curl -fL -o /tmp/typetype-downloader-e2e.bin "$BASE_URL/jobs/$job_id/artifact"
    bytes="$(wc -c < /tmp/typetype-downloader-e2e.bin)"
    printf 'artifact bytes: %s\n' "$bytes"
    test "$bytes" -gt 100000
    exit 0
  fi
  if [ "$status" = "failed" ]; then
    printf '%s\n' "$status_body" >&2
    exit 1
  fi
  sleep 1
done

printf 'timed out waiting for job %s\n' "$job_id" >&2
exit 1
