#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p .logs

if [[ ! -f .env ]]; then
  cp .env.example .env
fi

# Export .env so backend/frontend processes inherit configuration.
set -a
source .env
set +a

if ss -ltn "( sport = :6379 )" | rg -q ":6379"; then
  echo "Port 6379 is already in use."
  echo "Stop the conflicting service (e.g. system Redis) and retry:"
  echo "  sudo systemctl stop redis"
  exit 1
fi

echo "Starting Redis via docker compose..."
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  docker compose up -d redis
else
  echo "Docker is not running; continuing without Redis. /readyz will likely report not ready."
fi

echo "Installing frontend dependencies..."
npm --prefix frontend install

echo "Ensuring backend hot-reload tool (air) is available..."
if ! command -v air >/dev/null 2>&1; then
  GOBIN="$ROOT_DIR/.bin" go install github.com/air-verse/air@latest
fi

echo "Starting backend with hot reload on :8080..."
PATH="$ROOT_DIR/.bin:$PATH" air -c .air.toml > .logs/backend.log 2>&1 &
BACKEND_PID=$!
echo "$BACKEND_PID" > .backend.pid

echo "Starting frontend on :5173..."
npm --prefix frontend run dev > .logs/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "$FRONTEND_PID" > .frontend.pid

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"
echo "Frontend URL: http://localhost:5173"
echo "Backend URL:  http://localhost:8080"
echo "Logs: .logs/backend.log and .logs/frontend.log"
