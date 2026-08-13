# Week 2 status — Event Horizon (13.08.2026)

## Done (high priorities from w2 agent)

- [x] `internal/model/errors.go` on all gRPC services
- [x] `internal/converter/` proto↔model on all gRPC services
- [x] Handlers wired to converters (auth GetUser, inventory create/update/bulk, billing currency)
- [x] Auth service returns `*model.User` and uses domain errors
- [x] `//go:generate mockery` on `UserRepository` + hand-written mock (offline-safe)
- [x] Auth service unit tests (register/login/role/get)
- [x] Converter unit tests on auth/billing/game/inventory/leaderboard/profile/shop
- [x] Inventory mongo test tagged `//go:build integration`
- [x] `task test-coverage`, `task w2-check`
- [x] README testing section updated

## Deferred / optional

- Full testify/suite + mockery regen (proxy 502 / modcache perms — run locally when network is fine)
- 40–70% coverage across every package (W2 homework bar); EH now has real unit tests, deepen in later weeks
- Full Kozirev folder split (`internal/api/.../v1`, `repository/part/`) — not forced on living EH layout
- Testcontainers / CI test job → W4 / CI backlog

## Your rebuild (code changed vs last Docker push)

```bash
for svc in auth gateway billing inventory profile shop leaderboard game; do
  (cd services/$svc && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${svc}-service ./cmd/main.go)
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
```

Verify:

```bash
task w2-check
GOWORK=off go test ./internal/converter/ ./internal/service/ -count=1   # from services/auth
```
