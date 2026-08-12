# 🔍 INSPECTOR REPORT — Event Horizon (13.08.2026)

**Scope:** Full audit of all 9 live services (`auth`, `game`, `billing`, `leaderboard`, `shop`, `inventory`, `profile`, `gateway`, `balancer`, `nats-hub`) against the checklist in `confluence/agents/cursor_agents/inspector/INSPECTOR.md`, plus a `docs/openapi.yaml` ↔ Gateway ↔ Frontend gap-check, plus cross-cutting DevOps/infra (migrations, connection pools, k3s, CI/CD, monitoring, `.env`).

**Status:** Report only — no code changed yet. This is the required prerequisite before the `w1`–`w8` week agents run.

**Legend:** 🔴 critical / 🟡 needs work / 🟢 OK

---

## 0. Executive summary

The project is functionally wired end-to-end (register → login → play → score → leaderboard → shop → inventory all connect through NATS/gRPC), but almost nothing has the "production hardening" layer the Inspector checklist demands:

- **Zero services have `*_test.go` coverage** except one dead/broken Mongo test in `inventory`. Target was ≥70%.
- **No service configures a real DB connection pool** except `inventory` (and even that has idle=5, not 10).
- **No service exposes real `/health` + `/ready`** — most only expose `/metrics`; only `gateway`, `inventory`, `nats-hub` have `/health`, and none has `/ready`.
- **No Circuit Breaker or retry-with-jitter anywhere** in the codebase.
- **Gateway's JWT check doesn't verify the signature** — it just base64-decodes the payload. This is a real security hole, not a style nit.
- **Every service uses `log.Printf`/`log`, not structured JSON logging (slog/zap)** — despite w8's task being exactly "replace with Zap."
- **DB topology is inconsistent across 5 different sources of truth** (Makefile, README, service `config.go` defaults, docker-compose, k3s manifests) — worst offender is `inventory` (port/db/user all disagree in 4 places) and `leaderboard` (points at the `game` DB, not its own).
- **Outbox pattern is incomplete or missing** in `billing` (no outbox at all) and `inventory` (worker code exists, but the `CREATE TABLE outbox` migration is missing — fresh DB would break).
- **`profile` genuinely has no Redis**, confirming the project's own flagged gap.
- **`shop`'s purchase flow is not atomic across services** (spend billing, then write locally — no saga/compensation), and duplicates a whole unwired "V2" implementation.
- **`game`'s anti-cheat validators exist for hexagon but are never called** — client-submitted scores are trusted for 3 of 4 games.
- **OpenAPI is actually in decent shape**: nothing implemented is undocumented. The only doc drift is 4 inventory routes (bulk/reserve/soft/restore) documented but not wired into the Gateway, plus 2 frontend bugs (a dead `/api/billing/balance` call and a `/api/api/inventory/...` double-prefix bug in `inventoryApi.ts`).

---

## 1. Per-service scorecards

| Service | Architecture | Security/Data-integrity | Resilience | Observability | Tests |
|---|---|---|---|---|---|
| **auth** | 🟡 | 🔴 no `role` in JWT/schema, bcrypt cost=10 not 12 | 🔴 no pool, no retry/CB | 🟡 no health/ready | 🔴 0% |
| **gateway** | 🟡 | 🔴 JWT signature never verified, no RBAC, rate-limit disabled/buggy | 🔴 no CB | 🔴 plain-text logs | 🔴 0% |
| **balancer** | 🟢 | 🟡 hardcoded backends, no ENV | 🔴 no backend health checks, unsafe `Shutdown(nil)` | 🟡 | 🔴 0% |
| **billing** | 🟡 | 🔴 no outbox, no optimistic locking (lost-update risk) | 🔴 no pool, no CB; Redis-error fallback bug returns stale `0` | 🟡 | 🔴 0% |
| **shop** | 🟡 | 🔴 purchase not atomic across services, no outbox, no `version`, unused "V2" path, no indexes at all | 🔴 no pool, no CB | 🟡 | 🔴 0% |
| **inventory** | 🟡 | 🔴 outbox table never created (migration only adds an index), Mongo dual-stack contradicts "PG-only" directive | 🟢 only service with pool config (close, 5≠10) | 🟡 | 🔴 ~0% (Mongo test doesn't compile) |
| **profile** | 🟡 | 🔴 **confirmed: no Redis at all**; `user.registered` subscription commented out | 🟢 N/A for CB (no sync fan-out to Auth/Game/Billing — CQRS via NATS instead) | 🔴 no health, no logging | 🔴 0% |
| **leaderboard** | 🟡 | 🔴 Postgres restore path reads wrong table (`scores` vs `highscores`), score semantics bug (sums instead of takes max) | 🔴 `os.Exit(0)` skips cleanup, no pool | 🔴 no health | 🔴 0% |
| **game** | 🟡 | 🔴 hexagon anti-cheat validator exists but is **never called**; flappy/towers have no validator at all — client score trusted | 🔴 no pool | 🟡 | 🔴 0% |
| **nats-hub** | 🟢 (thin) | 🔴 no DLQ, no idempotency/dedupe, fake metrics (`nats_hub_connected` hardcoded to `1`) | 🟡 partial graceful shutdown | 🟡 | 🔴 0% |

---

## 2. 🔴 Cross-cutting critical fixes (do these first, they block several week-agents)

1. **Gateway JWT is not actually verified** (`services/gateway/cmd/main.go:135-161`, `getUserIDFromToken` only base64-decodes the payload — no signature/exp check, no role). This is a real auth bypass. Also `POST /api/game/submit` has no JWT check at all and trusts the body's `user_id`.
2. **Rate limiting is effectively off**: the real middleware import + `r.Use()` are commented out in `main.go:42-43,317-318`; the standalone limiter in `internal/ratelimit/limiter.go` doesn't implement "100 req/sec/user" (it's per-endpoint: 10/s submit, 5/s/IP login, 100/min/IP WS), and the alternate `internal/middleware/ratelimit.go` defaults to `allowed=false` for most paths (would 429 everything if re-enabled as-is).
3. **`inventory`'s outbox table is never created** — only an index migration exists (`services/inventory/migrations/20260803130003_add_outbox_fix.sql`), but `postgres_repo.go` inserts into `outbox` in the create-item transaction. A fresh DB deploy will fail. Fresh untracked file `services/inventory/internal/service/producer/item_producer.go` is a stub that also won't compile (imports a `platform/pkg/nats` package that doesn't exist in this repo) and isn't wired into `main.go` — likely leftover/in-progress work, needs to be finished or removed.
4. **`billing` has no outbox for `balance.updated` at all**, and no optimistic locking — concurrent balance updates can lose writes (no `version` column, no `SELECT ... FOR UPDATE`).
5. **`shop`'s purchase flow spends the user's currency via Billing gRPC, then does local DB work in a separate, unrelated transaction** — if the local write fails, the user loses currency with nothing to show for it. There's also a fully-built, unused "V2" implementation (`PurchaseItemV2`/`PurchaseItemWithStock`) that fixes some of this but the handler still calls V1.
6. **`game` anti-cheat is dead code**: `hexagons/validator.go` exists but `game_service.go` never calls it; `flappy`/`towers` have no validator at all. Any client can submit an arbitrary score for 3 of 4 games.
7. **DB topology disagreement for `inventory`** across 4 sources: Makefile says port 5466/db `eventhorizon_inventory`; `config.go` defaults to port 5466/db `inventory`/user `postgres`; README says port 5465/db `eventhorizon_shop`; k3s manifest points it at `postgres-shop`/`eventhorizon_shop` with metrics port 9099 (should be 9096). Pick one and align all four.
8. **`leaderboard` doesn't actually persist to its own Postgres** — compose wires it to the `game` Postgres instance (no `DB_NAME` set, falls back to `eventhorizon_game`), while its own migration creates an orphaned `leaderboard_backup` table that's never read/written by Go code. Its Postgres "restore on startup" queries a `scores` table that `game` never writes to (`game` only writes `highscores`). Net effect: **Redis is the only real source of truth**, and a Redis flush/restart loses the leaderboard. There's also a score-semantics bug — it sums incoming scores instead of taking the max (`ZScore` + add, not "keep highest").
9. **`profile`**: confirmed no Redis anywhere (biggest flagged gap in the whole project), and the `user.registered` NATS subscription that would populate email/nickname from Auth is commented out in `main.go:145-174` — so aggregated profiles are missing Auth data by design right now, independent of the Redis gap.
10. **No service configures a DB connection pool to spec** (target: MaxOpenConns=25/MaxIdleConns=10/ConnMaxLifetime=5m) except `inventory`, which sets idle=5 instead of 10. `auth`, `billing`, `profile`, `leaderboard` use bare `pgxpool.New()`; `shop`, `game` use bare `sql.Open()` with no `Ping`/limits at all.

---

## 3. OpenAPI / Frontend gap-check (docs/openapi.yaml)

**Good news:** every route the Gateway actually implements (19 routes) is documented in `docs/openapi.yaml`. Nothing implemented is missing from the docs.

**Documented but not implemented** (exist as gRPC methods on `inventory`, but the Gateway never exposes them over HTTP):
- `POST /api/inventory/items/bulk`
- `POST /api/inventory/items/{id}/reserve`
- `DELETE /api/inventory/items/{id}/soft`
- `POST /api/inventory/items/{id}/restore`

**Frontend bugs found while cross-checking:**
- `frontend/src/services/api.ts` exports a `getBalance()` calling `GET /api/billing/balance` — this endpoint doesn't exist anywhere (only `/api/billing/balance/all` does). Dead/broken helper.
- `frontend/src/services/inventoryApi.ts` sets its own `BASE_URL = '/api/inventory/items'` on top of an axios client whose `baseURL` is already `/api'` — requests actually go to `/api/api/inventory/items...`, which matches nothing. This is a real frontend bug (flagged now, fix scheduled for after backend work per your instructions).
- Two minor response-shape doc mismatches: OpenAPI documents inventory item endpoints as returning a bare `InventoryItem`, but the Gateway actually wraps it as `{ item: {...} }` (frontend already codes around this correctly, so it's a docs-only fix).

---

## 4. DevOps / infra findings

- **CI never runs tests.** `.github/workflows/main.yml` runs `actionlint` + `make docker-build-all` + `make docker-push-all` — `make test-all` exists but is never invoked.
- **CI Docker images are built from prebuilt binaries**, not multi-stage compiles: `Dockerfile.*.bin` files are `FROM scratch` + `COPY` a binary that must already exist locally. The proper multi-stage Dockerfiles live under `services/*/Dockerfile` but aren't what CI uses.
- **Monitoring is split into two stacks**, and the one that isn't live is stale/broken: `monitoring/docker-compose.monitoring.yml` is an **empty file**, `monitoring/prometheus.yml` only scrapes 5 of 10 services, and `monitoring/grafana/` is an empty directory. The actually-used stack is `deployments/prometheus/prometheus.yml` (wired into the main compose file) — this one is more complete but still has a **port collision**: shop (9095) collides with gateway-1's metrics port, and inventory (9096) collides with gateway-2's.
- **No Alertmanager/Telegram alerting** exists yet (matches the README's own open TODO).
- **k3s has no Helm charts** (matches README's own open TODO) and only covers 8 of 10 live services as one giant multi-container Pod — `balancer` and `nats-hub` aren't in it, and NATS/Postgres/Redis aren't K8s-managed resources.
- **Root `.env`** only has 6 vars (Jaeger, Docker Hub creds, Ansible creds) — none of the ~40+ service-level vars (`DB_*`, `REDIS_*`, `JWT_*`, etc.) that `config.go` files reference. Those currently rely entirely on docker-compose's inline `environment:` blocks or code defaults, not the root `.env`. Also: this file contains a live Docker Hub token and an Ansible become-password in plaintext — worth rotating if this `.env` has ever been shared.

---

## 5. Suggested next step

Per your ordering rule (Inspector → week agents), I'd propose fixing items in **Section 2** first since they're small, self-contained, and unblock the week agents cleanly:
- #1/#2 (Gateway JWT + rate limiting) relates directly to `w1` (gRPC Gateway) and security.
- #3 (inventory outbox table) and #4 (billing outbox) are both about the Outbox/NATS pattern — natural to batch together.
- #7/#8 (DB topology for inventory/leaderboard) should be fixed before any new migrations are added on top of the current mess.
- #9 (Profile Redis) is explicitly the project's own flagged #1 priority.

Everything in this report is a finding, not yet a fix — no files were modified during this audit.
