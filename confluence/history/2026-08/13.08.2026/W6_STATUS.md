# Week 6 status — Event Horizon (13.08.2026)

## Done (auth already existed — upgraded to course W6 shape)

- [x] `services/auth/internal/jwt` — access (15m) + refresh (7d) generator/validator
- [x] Refresh tokens in Redis (`auth:refresh:{jti}` + per-user set); access sessions kept
- [x] Proto: `refresh_token` on Login, `RefreshToken`, `Whoami` RPCs (regenerated locally)
- [x] Gateway: login returns refresh; `POST /api/auth/refresh`; `GET /api/auth/whoami`
- [x] `pkg/redisclient` shared Cache facade
- [x] Optional gRPC `interceptor.Auth` (not wired on all public RPCs by default — opt-in)
- [x] Env: `JWT_ACCESS_MINUTES`, `JWT_REFRESH_DAYS`

## Deferred

- Wire Auth interceptor on protected RPCs across services
- Full OpenAPI update for refresh/whoami

## Also done (cache decorator — 13.08.2026)

- [x] `repository.CachedRepository` — decorator over `InventoryRepository` + `RedisCacheRepo` adapter
- [x] Service layer cache removed; outbox via `ItemOutboxWriter` on decorated repo
- Adapter pattern: `RedisCacheRepo` stays the low-level Redis client; `CachedRepository` wraps the DB repo so service code never talks to cache directly (no double-cache)

## Rebuild auth + gateway

```bash
(cd services/auth && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-service ./cmd/main.go)
docker build -f Dockerfile.auth.bin -t eastwesser/auth:latest .
(cd services/gateway && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go)
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
```

If `go-redis` missing for `pkg/redisclient` (VPN off):

```bash
(cd pkg/redisclient && go mod download)
```

Verify:

```bash
(cd services/auth && go test ./internal/jwt ./internal/service -count=1)
```
