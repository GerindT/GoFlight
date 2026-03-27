# GoFlight

GoFlight is a Go-based flight data aggregator with a terminal-style Vue frontend.  
It combines flight status + weather data, adds resilience patterns (timeouts, circuit breakers, cache fallback), and exposes production-friendly health and metrics endpoints.

## Features

### Backend

- `GET /api/v1/dashboard/:flight` aggregation endpoint
- Cache-first reads with Redis + stale-cache fallback
- Fan-out/fan-in concurrency for upstream calls
- Circuit breakers via `sony/gobreaker`
- Structured logs (`slog`) with request IDs
- Metrics endpoint: `GET /metrics`
- Health endpoints: `GET /healthz`, `GET /readyz`

### Frontend (Terminal UI)

- Shell-like typing experience (no classic form input)
- Command history (`ArrowUp` / `ArrowDown`)
- Command aliases (`h`, `cls`, `f`, `nf`, etc.)
- Tab completion suggestions (shown in a bottom status line)
- Multi-command chaining with `&&`
- Built-in commands: `help`, `flight`, `status`, `uptime`, `neofetch`, `boot`, `height`, `clear`

## Tech Stack

- **Backend:** Go, Gin, Prometheus client, Redis, Gobreaker
- **Frontend:** Vue 3 + Vite
- **Dev tooling:** Air (backend hot reload), Docker Compose

## Project Layout

- `cmd/api/main.go` - API startup, middleware, graceful shutdown
- `internal/services/aggregator.go` - core aggregation and resilience
- `internal/external/*.go` - upstream API + Redis clients
- `internal/handlers/flight_handler.go` - HTTP handler layer
- `internal/middleware/*.go` - logger + metrics middleware
- `frontend/` - terminal-style Vue app
- `scripts/dev-local.sh` - start local full stack
- `scripts/stop-local.sh` - stop local processes
- `k8s/*.yaml` - Kubernetes manifests

## Quickstart (Local)

1. Copy env template:
   - `cp .env.example .env`
2. Add your API keys in `.env`:
   - `AVIATIONSTACK_API_KEY`
   - `OPENWEATHER_API_KEY`
3. Start full local stack:
   - `bash scripts/dev-local.sh`
4. Open:
   - Frontend: `http://localhost:5173`
   - API health: `http://localhost:8080/healthz`
5. Stop:
   - `bash scripts/stop-local.sh`

## Terminal Commands (Frontend)

- `help` - list commands
- `flight LH123` - fetch flight dashboard
- `status` - print local UI/API status
- `uptime` - show session uptime
- `neofetch` - print splash
- `boot` - run boot animation
- `height 620` - set terminal height
- `clear` - clear terminal output
- Command chaining:
  - `neofetch && status && flight LH123`

## Hot Reload

- Frontend: Vite dev server
- Backend: Air (auto-installed to `./.bin` by `scripts/dev-local.sh` if missing)

## Verify

- Backend tests:
  - `go test ./...`
- Frontend build:
  - `npm --prefix frontend run build`

## Deployment Notes

- Docker image and compose are included for local/runtime use.
- Kubernetes manifests are in `k8s/` (namespace, configmap, secret, deployment, service, ingress, hpa).
- `k8s/secret.yaml` uses demo placeholders; use proper secret management in production (Sealed Secrets, ESO, Vault, etc.).
