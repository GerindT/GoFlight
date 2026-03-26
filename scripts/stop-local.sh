#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f .backend.pid ]]; then
  kill "$(cat .backend.pid)" 2>/dev/null || true
  rm -f .backend.pid
fi

if [[ -f .frontend.pid ]]; then
  kill "$(cat .frontend.pid)" 2>/dev/null || true
  rm -f .frontend.pid
fi

docker compose stop redis >/dev/null 2>&1 || true
echo "Local services stopped."
