# Week 7 status — Event Horizon (13.08.2026)

## Done

- [x] `platform/pkg/tracing` — OTLP init + `GetTraceID` / `GetSpanID`
- [x] `platform/pkg/metrics` — gRPC unary interceptor (`grpc_requests_total`, `grpc_request_duration_seconds`, `grpc_request_errors_total`, `service_health`)
- [x] Auth pilot: platform tracing + metrics interceptors (replaces inline OTLP setup)
- [x] Inventory: tracing + metrics interceptors
- [x] Shop: tracing + metrics interceptors
- [x] `platform/pkg/logger.WithContext` — adds `trace_id` / `span_id` from OTEL ctx

## Deferred (roll out incrementally)

- Wire tracing/metrics interceptors on billing, game, profile, leaderboard, gateway gRPC clients
- OTEL Collector sidecar in compose (Jaeger direct OTLP is enough for dev)

## Verify locally

```bash
task w7-check
(cd services/auth && GOWORK=off go test ./... -count=1)
(cd services/inventory && GOWORK=off go test ./... -count=1)
(cd services/shop && GOWORK=off go test ./... -count=1)
```

## Rebuild (only changed services)

```bash
for svc in auth inventory shop; do
  (cd services/$svc && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${svc}-service ./cmd/main.go)
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
```

Jaeger UI: http://localhost:16686 — look for `auth`, `inventory`, `shop`.

Prometheus: http://localhost:9090 — query `grpc_requests_total`.
