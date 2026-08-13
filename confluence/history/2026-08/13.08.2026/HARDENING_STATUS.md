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
| 3.2 | testcontainers | 🟡 placeholder `-tags=integration` (wire when Docker CI ready) |
| 3.3 | Coverage >70% everywhere | 🟡 critical paths tested; not full ≥70% yet |

## Priority 4 — Docs

| # | Item | Status |
|---|------|--------|
| 4.1 | ORCHESTRATOR.md refreshed | ✅ |
| 4.2 | Prydwen key files (EH / Outbox / NATS / PG perf / ClickHouse) | ✅ |

## Priority 5 — MCP / RAG

✅ Done — `services/mcp` + `.cursor/mcp.json` + `MCP_RAG_STATUS.md`

## Verified locally

- Hardening unit tests + linux builds (shop/gateway/analytics/inventory/billing) — OK
- MCP: `go test ./internal/...`, binary `services/mcp/mcp-server` — OK
- RAG smoke: Outbox/Shop query returns Prydwen hits — OK

## Rebuild & push

```bash
# binaries already built under services/*/ if you just ran the agent; else:
(cd services/shop && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go)
(cd services/gateway && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go)
(cd services/analytics && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o analytics-service ./cmd/main.go)
(cd services/inventory && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go)
(cd services/billing && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o billing-service ./cmd/main.go)

docker build -f Dockerfile.shop.bin -t eastwesser/shop:latest .
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker build -f Dockerfile.inventory.bin -t eastwesser/inventory:latest .
docker build -f Dockerfile.billing.bin -t eastwesser/billing:latest .

docker push eastwesser/shop:latest
docker push eastwesser/gateway:latest
docker push eastwesser/analytics:latest
docker push eastwesser/inventory:latest
docker push eastwesser/billing:latest

docker compose -f deployments/docker-compose.cluster.yml up -d \
  shop gateway gateway-2 gateway-3 analytics inventory billing
```
