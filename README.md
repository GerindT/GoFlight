# GoFlight

GoFlight is a concurrent API gateway that aggregates flight details and weather data with caching, circuit breakers, and Prometheus metrics.

## Features

- Gin API endpoint: `GET /api/v1/dashboard/:flight`
- Cache-first reads with Redis and stale-cache fallback
- Fan-out/fan-in upstream calls guarded by context
- Circuit breakers using `sony/gobreaker`
- Structured request logging with `log/slog`
- Prometheus metrics at `GET /metrics`
- Health endpoints: `GET /healthz`, `GET /readyz`

## Project Layout

- `cmd/api/main.go` - app wiring and graceful shutdown
- `internal/services/aggregator.go` - aggregation logic and resilience patterns
- `internal/external/*.go` - upstream and cache clients
- `internal/handlers/flight_handler.go` - HTTP controller
- `internal/middleware/*.go` - logging and metrics middleware
- `k8s/*.yaml` - Kubernetes manifests

## Run Locally

1. Copy env file and set API keys:
   - `cp .env.example .env`
2. Start backend + frontend in one command:
   - `bash scripts/dev-local.sh`
3. Open app:
   - `http://localhost:5173`
4. Example API requests:
   - `curl http://localhost:8080/healthz`
   - `curl http://localhost:8080/readyz`
   - `curl http://localhost:8080/api/v1/dashboard/LH123`
   - `curl http://localhost:8080/metrics`
5. Stop local services:
   - `bash scripts/stop-local.sh`

### Hot Reload

- Frontend hot reload: enabled by Vite (`npm --prefix frontend run dev`).
- Backend hot reload: enabled via `air` through `scripts/dev-local.sh`.
- `dev-local.sh` auto-installs `air` to `./.bin` if missing.

## Tests

- `go test ./...`
- Frontend build check:
  - `npm --prefix frontend run build`

## Docker

- Build: `docker build -t goflight:local .`
- Run:
  - `docker run --rm -p 8080:8080 --env-file .env goflight:local`

## Kubernetes (Minikube)

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
```

## Notes

- `k8s/secret.yaml` uses base64 placeholders for demo only.
- For production secret management, prefer Sealed Secrets or External Secrets Operator.
