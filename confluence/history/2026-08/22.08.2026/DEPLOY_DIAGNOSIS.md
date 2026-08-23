# Deploy diagnosis — 22/23.08.2026

After `make deploy`, Auth / Game / Notification / Analytics looked like “net errors”.

## Root cause

Not application bugs in those four services first.

1. **Docker stop/kill** around `2026-08-22T20:54Z` left most containers `Exited (255)` (daemon/host restart while containers ran).
2. **Infra stayed down**: Postgres, Redis, NATS, Kafka, ClickHouse → app containers restart-looped with:
   - `lookup event-horizon-postgres…: no such host`
   - `lookup kafka: no such host`
   - `lookup event-horizon-clickhouse: no such host`
3. Bringing infra back (`docker compose up -d`) restored Auth, Game, etc.

## Per-service notes

| Service | Symptom | Real issue | Fix |
|---------|---------|------------|-----|
| **Auth** | Exited 255 / DNS to Postgres+Redis | Infra down | Start postgres + redis; auth came up healthy |
| **Game** | Exited / `postgres starting up` / DNS | Infra down + race | Start `postgres-game`; game healthy on `:50052` |
| **Notification** | Kafka DNS → **noop consumer** | Started while Kafka down; stays idle until restart | `docker restart deployments-notification-1` after Kafka healthy |
| **Analytics** | ClickHouse auth / DNS / missing DB | (1) CH down (2) `default` user localhost-only (3) DB `eventhorizon` missing before Ping | Compose: `CLICKHOUSE_SKIP_USER_SETUP` + network users.d; Ping uses `default` DB; create schema on start |

## Extra bugs found while recovering

| Issue | Impact | Fix applied |
|-------|--------|-------------|
| Gateway healthcheck `wget --spider` → **HEAD** `/ready` | Gin returns **404** → gateways marked unhealthy | Healthcheck uses GET (`wget -qO-`) |
| Compose warns `JWT_SECRET` empty | Auth falls back to insecure default | Export `JWT_SECRET` in shell/env before `make deploy` (do not commit `.env`) |

## Recovery commands (if this happens again)

```bash
cd deployments
# 1) Infra first
docker compose -f docker-compose.cluster.yml up -d \
  postgres postgres-game postgres-billing postgres-shop postgres-inventory \
  postgres-payment postgres-authors postgres-history postgres-leaderboard postgres-profile \
  redis redis-game redis-billing redis-shop redis-inventory redis-payment redis-authors \
  redis-leaderboard redis-profile \
  nats-1 nats-2 nats-3 kafka clickhouse

# 2) Then apps
docker compose -f docker-compose.cluster.yml up -d

# 3) Restart consumers that fell back to noop
docker restart deployments-notification-1
```

Or simply: `make deploy` again **after** Docker daemon is stable.

## Related architecture docs (same day)

- `confluence/architecture/LOAD_RESILIENCE.md` — peak-load lessons
- `confluence/architecture/GAME_OUTBOX.md` — Game needs Outbox
- `confluence/architecture/PATRONI.md` — Patroni later, not for this failure mode
- `confluence/architecture/BRANDING.md` — logos
