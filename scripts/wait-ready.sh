#!/bin/sh
# Waits until every service of the local ecosystem answers on the host, then the gateway reports ready.
# Usage: sh scripts/wait-ready.sh [timeout_seconds]   (default 180)
set -eu

TIMEOUT="${1:-180}"
PING_URLS="
http://localhost:8081/ping
http://localhost:8082/ping
http://localhost:8083/ping
http://localhost:8084/ping
http://localhost:8085/ping
http://localhost:8086/ping
"
READY_URL="${GATEWAY_READY_URL:-http://localhost:8080/ready}"

# wait_for <url>: polls once per second until the URL answers 2xx or the timeout expires.
wait_for() {
  url="$1"
  i=0
  until curl -fsS -o /dev/null "$url" 2>/dev/null; do
    i=$((i + 1))
    if [ "$i" -ge "$TIMEOUT" ]; then
      echo "timeout waiting for $url" >&2
      return 1
    fi
    sleep 1
  done
  echo "ok: $url"
}

for url in $PING_URLS; do
  wait_for "$url"
done
wait_for "$READY_URL"
