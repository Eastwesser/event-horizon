# 19.08.2026 — continue from sleep + launch problems

## Build (LAUNCHING_PROBLEMS)

Root cause was **not** a missing `../mcp` folder (go.work already listed `./services/mcp`). Real blockers:

1. `make build-nats-hub` compiled `./main.go` — file lives at `services/nats-hub/cmd/main.go`
2. `services/mcp/go.mod` required `go 1.26.5` while the rest of the workspace is `1.25.7`

Fixes:

- Makefile → `go build … ./cmd/main.go -o nats-hub`
- `go.work` + `services/mcp/go.mod` → `go 1.25.7`

## Kafka vs Game `:9092`

`make deploy` died because **Game metrics already bind host 9092**. Kafka wanted the same port.

Fix: Kafka host mapping **`19092:9092`**. In-network clients still use `kafka:9092`.

## Docker images pushed (19.08)

| Image | Why |
|-------|-----|
| `eastwesser/gateway:latest` | all gRPC CBs + embedded OpenAPI 1.0.7 |
| `eastwesser/inventory:latest` | proto `version` |
| `eastwesser/nats-hub:latest` | `cmd/main.go` binary |

MCP is local stdio — no image.

If `make deploy` failed mid-way, run: `make migrate-all` once Postgres is up.

## P0 / P1

- OpenAPI auth routes (also copied to `services/gateway/api/openapi.yaml` for embed)
- Version **v1.0.7**
- FE `/history` + `/api` prefix fix
- Integration tests: payment, authors, history, analytics (`make test-integration`)

## P2 (19.08)

- Unit smokes: balancer least-conn pick, fulfillment `HandlePurchasePaid` (1ms + stub producer), notification Handle (ignored topic / no telegram / bad JSON)
- Shop transactional outbox for `shop.purchased` (migration + worker; `shop.purchase.failed` still direct NATS)
- `slog` in payment/authors/billing/inventory/shop outbox workers and gateway WebSocket connect/disconnect

Shop image was **not** rebuilt/pushed this pass. After deploy: `make migrate-shop` (or let shop `migrator.Up` on start).

## System design doc (19.08)

Full Mermaid diagrams: [`confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md`](../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md) — services, ports, outboxes, NATS/Kafka, purchase flow, observability.

Local shop image rebuilt: `eastwesser/shop:latest` (outbox worker). Push + `docker compose up -d shop` when ready.
