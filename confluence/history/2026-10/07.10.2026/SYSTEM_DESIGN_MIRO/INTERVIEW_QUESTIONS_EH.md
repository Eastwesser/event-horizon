# Event Horizon — interview cheat sheet (where / ports / how)

Living notes for v1.1.0. Topology diagram: `README.md` + `confluence/architecture/EH_SCHEMAS.md`. Miro board: https://miro.com/app/board/uXjVJLLg9us=/

## Entry points

| Surface | Port / URL | Role |
|---------|------------|------|
| React FE (dev) | `:5173` | Vite client; proxies `/api` → balancer |
| Balancer | `:8079` | Public HTTP + WS; load-balances gateway replicas |
| Gateway ×3 | `:8081`–`:8083` | JWT, HTTP→gRPC, OpenAPI `/docs`, `/openapi.yaml` |
| NATS JetStream | `:4222` (+ cluster peers) | Events bus (`EVENTS` stream via nats-hub) |
| ClickHouse | `:8123` / `:9000` | Analytics |
| Prometheus / Grafana / Jaeger | compose defaults | Metrics / dashboards / traces |

Public API prefix: **`/api/`** (no `/api/v1/`). FE axios `baseURL: '/api'` + relative paths (`/profile`, not `/api/profile`).

## gRPC services (internal)

| Service | gRPC | Data |
|---------|------|------|
| Auth | `:50051` | PG `:5460` + Redis sessions |
| Game | `:50052` | PG `:5461` |
| Billing | `:50053` | PG `:5462` + Redis |
| Leaderboard | `:50054` | PG `:5463` + Redis Sorted Set (`:6382`) |
| Shop | `:50055` | PG `:5465` + Redis |
| Analytics | `:50057` | ClickHouse |
| Payment | `:50058` | subscription / CanPurchaseMerch |
| Inventory | `:50059` | PG + Mongo (catalog source) + Outbox |
| Profile | `:50060` | PG `:5464` |
| Authors | `:50061` | PG `:5468` + Redis |
| History | `:50062` | PG `:5469` |
| Fulfillment / Notification | compose | NATS consumers |
| NATS Hub | compose | stream bootstrap |

## Patterns (point to code)

| Pattern | Where |
|---------|--------|
| Clean Architecture | `services/*/internal/{handler,service,repository,model}` |
| Unary interceptors | Recovery → Logger → Validate on gRPC servers |
| Outbox | Inventory reference (`services/inventory`) |
| Rate limit / CB | Gateway / balancer (ask current week notes) |
| LB real-time | Redis Sorted Set → WS `/ws/leaderboard` |
| Auth sessions | JWT + Redis; roles `user` \| `author` \| `admin` |
| bcrypt | cost 12 |
| Purchase path | Shop spend tickets → NATS `purchase.paid` → fulfillment/notification |
| History ingest | NATS `user.registered` **and** `event.user.registered` (normalized) |

## Health

Each metrics HTTP exposes `/health` (liveness) and `/ready` (dependency ping). Balancer front door: `http://localhost:8079`.

## Load / 10k DAU model

Targets: `confluence/history/2026-10/07.10.2026/LOAD_TESTS/METRICS.md`.  
Runbook: `confluence/history/2026-10/08.10.2026/LOAD_RESULTS/README.md`.  
k6 scripts: `deployments/k6/` (prefer `browse.js` / `purchase.js`).

## Interview angles

1. **Why balancer in front of gateway?** Horizontal gateway replicas; sticky not required for JWT.
2. **Why Outbox on inventory?** Durable publish after DB commit; inventory is the reference service.
3. **Lamps vs tickets?** Lamps = boosts; tickets = shop currency.
4. **Boosted runs?** Not ranked; no shop profit from boosted scores.
5. **Physical merch floor?** ≥ 100_000 tickets; needs active subscription (`CanPurchaseMerch`).
6. **Security baseline?** JWT Redis, validate interceptor, no hardcoded secrets, rate limiter surfaces — full CSRF/XSS pass still open on ticklist §3.
