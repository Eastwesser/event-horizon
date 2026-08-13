# Week 3 status — Event Horizon (13.08.2026)

## Done (high priorities from w3 agent + Kozirev HW, adapted)

- [x] `.golangci.yml` at repo root (gci prefix `github.com/Eastwesser/event-horizon/`)
- [x] `task lint` + `task w3-check`
- [x] Shared `pkg/migrator` — Goose-style `-- +goose Up` runner, tracks `goose_db_version` (compatible with `make migrate-*`)
- [x] Shared `pkg/sqb` — lightweight SQL builder (Squirrel-like; zero extra deps while proxy/`go get` is flaky)
- [x] Auth repository uses `pkg/sqb` for queries
- [x] Auto-migrations on startup for auth, billing, game, inventory (postgres), leaderboard, profile, shop (`//go:embed` + `migrator.Up`)
- [x] Modular compose under `deployments/compose/` (core + per-service deps); cluster compose remains canonical
- [x] Cluster app healthchecks hit `/ready` (metrics ports / gateway `:8080`)
- [x] Bin images switched to Alpine + `wget` so healthchecks work (was scratch)
- [x] `services/profile/Dockerfile` added
- [x] `go.work` includes `pkg/migrator`, `pkg/sqb`

## Notes / deferred

- Real `github.com/Masterminds/squirrel` + `pressly/goose` — swap when network/`go get` is reliable; APIs are intentionally close
- Full golangci-lint clean pass across all services (install linter locally, then `task lint`)
- gRPC Gateway (w3 agent “в итоге”) — still later; HTTP stays on Gin gateway
- Multi-stage service Dockerfiles that only `COPY services/<svc>` need the same `/pkg/...` layout as auth/profile if you build without `*.bin`

## Rebuild (code + base image change)

```bash
for svc in auth billing game inventory leaderboard profile shop gateway; do
  (cd services/$svc && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${svc}-service ./cmd/main.go)
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
```

Verify:

```bash
task w3-check
(cd pkg/migrator && go test ./...)
(cd pkg/sqb && go test ./...)
(cd services/auth && GOWORK=off go test ./internal/repository/ ./internal/service/ -count=1)
```
