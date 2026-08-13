# Week 4 status — Event Horizon (13.08.2026)

## Done (priorities from w4 agent + Kozirev HW, adapted)

- [x] `platform/pkg/closer` — LIFO graceful shutdown (slog, no zap/`go get`)
- [x] `platform/pkg/logger` — structured slog wrapper (`LOG_LEVEL` / `LOG_FORMAT`)
- [x] Auth DI: thin `cmd/main.go` + `internal/app/{app,di}.go`
- [x] Auth `internal/config/interfaces.go` (ConfigProvider + typed views)
- [x] Env templates: `deployments/env/*.env.template` + README
- [x] Auth integration + in-process gRPC e2e under `//go:build integration`
- [x] `task w4-check`, `task test-integration-auth`
- [x] `go.work` includes `./platform`

## Deferred / next

- Roll DI/`platform` to billing, inventory, shop, …
- Full Kozirev `internal/api/.../v1` file-per-RPC split (auth still uses `handler/`)
- Real testcontainers + CI job (needs Docker + network for module pull)
- Zap logger swap if you want exact course deps

## Rebuild

```bash
(cd services/auth && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-service ./cmd/main.go)
docker build -f Dockerfile.auth.bin -t eastwesser/auth:latest .
```

Verify:

```bash
task w4-check
(cd platform && go test ./...)
(cd services/auth && GOWORK=off go test ./internal/... -count=1)
# with Postgres:
# AUTH_TEST_DATABASE_URL='postgres://...' task test-integration-auth
```
