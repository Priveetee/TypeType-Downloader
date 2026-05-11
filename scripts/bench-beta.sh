#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:18082/api/downloader}"
SMALL_URL="${SMALL_URL:-https://www.youtube.com/watch?v=dQw4w9WgXcQ}"
LARGE_URL="${LARGE_URL:-https://www.youtube.com/watch?v=AKeUssuu3Is}"
OUT_DIR="${OUT_DIR:-$(pwd)/bench-out}"
RUN_LARGE="${RUN_LARGE:-0}"
FETCH_ARTIFACT="${FETCH_ARTIFACT:-1}"

mkdir -p "$OUT_DIR"

now_ms() { date +%s%3N; }

payload() {
  local url="$1" mode="$2" height="${3:-}"
  URL="$url" MODE="$mode" HEIGHT="$height" python3 - <<'PY'
import json, os
options = {"mode": os.environ["MODE"], "format": "mp4", "container": "mp4", "allowQualityFallback": True}
if os.environ["HEIGHT"]:
    options["height"] = int(os.environ["HEIGHT"])
print(json.dumps({"url": os.environ["URL"], "options": options}))
PY
}

field() {
  BODY="$1" NAME="$2" python3 - <<'PY'
import json, os
value = json.loads(os.environ["BODY"]).get(os.environ["NAME"], "")
print(value if value is not None else "")
PY
}

wait_job() {
  local id="$1" body status
  for _ in $(seq 1 7200); do
    body="$(curl -fsS "$BASE_URL/jobs/$id")"
    status="$(field "$body" status)"
    if [ "$status" = "done" ]; then
      printf '%s\n' "$body"
      return 0
    fi
    if [ "$status" = "failed" ]; then
      printf '%s\n' "$body" >&2
      return 1
    fi
    sleep 1
  done
  printf 'timed out waiting for job %s\n' "$id" >&2
  return 1
}

print_job_metrics() {
  BODY="$1" python3 - <<'PY'
import json, os
d = json.loads(os.environ["BODY"])
resolved = d.get("resolved") or {}
keys = ["id", "status", "totalMs", "downloadMs", "muxMs", "runTimeMs", "downloadedBytes", "totalBytes"]
parts = [f"{key}={d.get(key, '')}" for key in keys]
parts.extend(f"{key}={resolved.get(key, '')}" for key in ["videoItag", "audioItag", "height", "container"])
print("job " + " ".join(parts))
PY
}

bench_go() {
  local name="$1" url="$2" mode="$3" height="${4:-}"
  local start created id cached body end artifact_start artifact_end code
  start="$(now_ms)"
  created="$(curl -fsS -X POST "$BASE_URL/jobs" -H 'Content-Type: application/json' -d "$(payload "$url" "$mode" "$height")")"
  id="$(field "$created" id)"
  cached="$(field "$created" cached)"
  body="$(wait_job "$id")"
  end="$(now_ms)"
  printf 'go_%s enqueue_to_done_ms=%s cached=%s\n' "$name" "$((end - start))" "$cached"
  print_job_metrics "$body"
  if [ "$FETCH_ARTIFACT" = "1" ]; then
    artifact_start="$(now_ms)"
    code="$(curl -fsS -L -o /dev/null -w '%{http_code}' "$BASE_URL/jobs/$id/artifact")"
    artifact_end="$(now_ms)"
    printf 'go_%s artifact_http=%s artifact_ms=%s\n' "$name" "$code" "$((artifact_end - artifact_start))"
  fi
}

bench_go small_video "$SMALL_URL" video 1080
bench_go small_audio "$SMALL_URL" audio

if [ "$RUN_LARGE" = "1" ]; then
  bench_go large_audio "$LARGE_URL" audio
  bench_go large_video "$LARGE_URL" video 1080
fi
