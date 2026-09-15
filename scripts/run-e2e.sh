#!/usr/bin/env sh
set -eu

api_pid=""
cleanup() {
  if [ -n "$api_pid" ]; then
    kill "$api_pid" 2>/dev/null || true
    wait "$api_pid" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

docker compose up -d
make migrate-up
DATABASE_URL="${DATABASE_URL:-postgres://feature_flags:feature_flags@localhost:5432/feature_flags?sslmode=disable}" \
REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}" \
HTTP_ADDR=:8081 go run ./cmd/api > /tmp/feature-flag-manager-e2e.log 2>&1 &
api_pid=$!

attempt=0
until curl -fsS http://localhost:8081/healthz >/dev/null; do
  attempt=$((attempt + 1))
  if [ "$attempt" -eq 30 ]; then
    cat /tmp/feature-flag-manager-e2e.log
    exit 1
  fi
  sleep 1
done

python3 -m pytest e2e --base-url http://localhost:8081 -v
