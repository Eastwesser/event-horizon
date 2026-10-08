# Interview — patterns & where they live (v1.1.0)

## Methods / patterns

| Pattern | Where | One-liner |
|---------|--------|-----------|
| Clean Architecture | `services/*/internal/{handler,service,repository,model}` | Handler → Service → Repository only |
| gRPC interceptors | each service app bootstrap | Recovery → Logger → Validate |
| Outbox | `services/inventory` | DB commit then durable NATS publish |
| JWT + Redis sessions | Auth + Gateway | Roles `user` \| `author` \| `admin`; bcrypt cost 12 |
| Leaderboard | Leaderboard svc + Redis Sorted Set | Hot path for tops; WS fan-out |
| WebSockets | Balancer / gateway `WS /ws/leaderboard` | Redis SS updates → client; **not** Telegram |
| Notifications | Notification svc + FE `NotificationBell` | In-app bell; soft-fail when logged out |
| Rate limit | Gateway / balancer | Login burst capped (~100 rps) — explains k6 500 VU auth fails |
| Circuit / resilience | week notes + gateway clients | Prefer fail-fast over cascading |
| Purchase path | Shop → NATS `purchase.paid` | Fulfillment + Notification consumers |
| History ingest | History worker | Subscribes `user.registered` **and** `event.user.registered` |
| Dual currency | Billing | Lamps = boosts; tickets = shop |
| Boost unranked | Game submit + FE BoostCheckbox | Boosted runs not in LB / no shop profit |
| Observability | Prometheus / Grafana / Jaeger | `/health` + `/ready` per metrics HTTP |

## WebSockets — truth

- **Leaderboard live updates:** Redis Sorted Set → service → **`ws://localhost:8079/ws/leaderboard`**.
- **Notifications:** HTTP poll / API + in-app bell (not a second WS product surface for players).
- Players should **not** get Telegram spam for routine events (identity/Telegram = future / optional).

## FE architecture note

Backend-first delivery, then FE. After v1.1.0 cleanup: **one folder per game** under `frontend/src/components/Games/*` (store + UI co-located); shop/inventory stores under `Shop/` / `Inventory/`. Layout is OK for the course; chrome via `Layout/` + shared `ui/`.

## SQL practice (admin / interview)

```sql
-- Authors with goods count
SELECT a.id, a.display_name, COUNT(i.id) AS goods
FROM authors a
LEFT JOIN inventory_items i ON i.author_id = a.id AND i.deleted_at IS NULL
GROUP BY a.id, a.display_name
ORDER BY goods DESC;

-- Top sold items (shop purchases)
SELECT p.item_id, COUNT(*) AS sales, SUM(p.price) AS tickets
FROM purchases p
GROUP BY p.item_id
ORDER BY sales DESC
LIMIT 20;

-- Users with active-looking payment/sub (adjust to your payment schema)
-- JOIN auth users ↔ payment subscriptions for admin "subscription column"
```

## What could be better / risks

- Auth write path does not hold 500 VU login (by design: bcrypt 12 + rate limit) — scale with cache/async or accept limit.
- Miro/Mermaid drift until Denis exports board.
- Avatar upload, 108-author seed, C4 payouts still product-open.
- Full CSRF/XSS/SQL pass still a checklist (baseline note only).

**Anti-patterns avoided:** hardcoded secrets in repo, reverse Clean Architecture deps, double `/api/api` FE prefix (fixed).
