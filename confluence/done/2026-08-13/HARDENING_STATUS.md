# Hardening status — FINAL_PRIORITY_TASKS (13.08.2026)

## Priority 1 — Security / integrity

| # | Item | Status |
|---|------|--------|
| 1.1 | ClickHouse parameterized queries + injection test | ✅ |
| 1.2 | RBAC e2e (gateway roles + inventory gRPC `x-user-role` + middleware test) | ✅ |
| 1.3 | Optimistic locking `version` (billing / inventory / shop migrations + SQL) | ✅ |
| 1.4 | Circuit breaker on gateway gRPC (billing / shop / payment) | ✅ |

## Priority 2 — FE / integrations

| # | Item | Status |
|---|------|--------|
| 2.1 | Shop merch → Payment `CanPurchaseMerch` | ✅ |
| 2.2 | FE `paymentApi` / `authorsApi` / `analyticsApi` | ✅ |
| 2.3 | Dead Redis removed from Game + History config | ✅ |

## Priority 3 — Tests

| # | Item | Status |
|---|------|--------|
| 3.1 | `make test-unit` / `test-smoke` / `test-k6` | ✅ pipeline present |
| 3.2 | testcontainers Billing/Shop/Inventory | ✅ `make test-integration` (verified with Docker) |
| 3.3 | Coverage critical packages | ✅ jwt **92.9%**, billing converter **100%**, inventory converter **71%**, inventory UpdateItem service tests, merch gate |

## Soft debt — backend

| Item | Status |
|------|--------|
| Gateway gRPC→HTTP map + inventory version → 409 | ✅ `writeGRPCError` + handler `Aborted` |
| OpenAPI sync payment/authors/history/analytics | ✅ `docs/openapi.yaml` |
| Circuit breaker on **all** gRPC clients | ✅ auth/game/profile/billing/shop/payment/authors/history/analytics/inventory/leaderboard |
| Proto `version` on inventory Item + UpdateItemRequest | ✅ field 12 / 10; gateway PUT accepts `version` |
| Whole-service coverage ≥70% | 🟡 deferred — critical packages above bar; full `internal/` still below (handlers/DI/app) |

## Soft debt — frontend

Full list: [`FRONTEND_REMAINING.md`](./FRONTEND_REMAINING.md) · Hanoi: [`FRONTEND_KHANOY_TOWERS.md`](./FRONTEND_KHANOY_TOWERS.md)

| # | Item | Status |
|---|------|--------|
| F1 | API clients payment / authors / analytics | ✅ |
| F2 | UI wired to those APIs | ✅ |
| F3 | Ханойская башня (`/game/hanoi`) | ✅ |
| F4 | Subscription / Payment UX (`/subscription`) | ✅ |
| F5 | Authors UI (`/authors`) | ✅ |
| F6 | Analytics dashboard (`/analytics`, admin) | ✅ |
| F7 | Shop `canPurchaseMerch` UX | ✅ |

## Priority 4 — Docs

| # | Item | Status |
|---|------|--------|
| 4.1 | ORCHESTRATOR.md refreshed | ✅ |
| 4.2 | Prydwen key files (EH / Outbox / NATS / PG perf / ClickHouse) | ✅ |

## Priority 5 — MCP / RAG

✅ Done — `services/mcp` + `.cursor/mcp.json` + `MCP_RAG_STATUS.md`

## Verified this pass

- Gateway `GOWORK=off go build` with full CB wrap
- Inventory proto regen + converter/service tests
- Frontend `npm run build` (Hanoi, Subscription, Authors, Analytics, merch gate)
- jwt coverage **92.9%**

## Rebuild & push (when Emma asks)

```bash
(cd services/gateway && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go)
(cd services/inventory && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go)

docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker build -f Dockerfile.inventory.bin -t eastwesser/inventory:latest .
docker push eastwesser/gateway:latest
docker push eastwesser/inventory:latest
```
