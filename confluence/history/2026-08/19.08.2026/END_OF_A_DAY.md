# End of day — 19.08.2026

Summary of what we achieved today. **Nothing committed yet** — Emma does manual commits (see [Recommended commits](#recommended-commits) below).

---

## Executive summary

| Area | Outcome |
|------|---------|
| **Launch / deploy** | Fixed nats-hub build path, go.work MCP version, Kafka `:9092` clash → `:19092` |
| **P0–P2** | Closed (integration tests, shop outbox, unit smokes, slog migration) |
| **Payment** | YooKassa **local stub** + idempotent `ConfirmPayment`; end-to-end smoke **passed** |
| **Boosty** | Researched — no creator webhook URL in UI; stays **manual / secondary** fallback |
| **Gateway errors** | `handleRPCError` / `writeGRPCError` on RPC routes; inventory 409; merch **403** + `subscription_required` |
| **API routes** | Confirmed prefix **`/api/`** (no v1); fixed frontend double-`/api` bug; documented in `API_ROUTES.md` |
| **Docs** | README v1.0.7, system design Mermaid, tech_debt reorg, STATUS_CODES §7 done |

---

## 1. Infrastructure & launch fixes

- **nats-hub Makefile**: build `./cmd/main.go` (not `./main.go`).
- **go.work / mcp**: aligned on Go `1.25.7`.
- **Kafka vs Game**: host port `19092:9092` in compose so both can run.
- **Docker images** rebuilt/pushed where needed: gateway, inventory, nats-hub, auth, payment, shop (local).
- **`.env` incident**: file restored after accidental truncation; **do not edit `.env` unless explicitly asked**.

See: [`CONTINUE.md`](CONTINUE.md), [`18.08.2026/LAUNCHING_PROBLEMS.md`](../18.08.2026/LAUNCHING_PROBLEMS.md).

---

## 2. P2 backend hardening (done)

- **Shop outbox**: `shop.purchased` in same PG tx as purchase; worker → NATS (`20260819120000_add_outbox.sql`).
- **Unit smokes**: balancer least-conn, fulfillment `HandlePurchasePaid`, notification edge cases.
- **slog**: outbox workers (payment, authors, billing, inventory, shop) + gateway WebSocket.
- **otelgrpc**: DI fixes in billing, game, leaderboard, shop.
- **Integration tests**: payment, authors, history, analytics (`make test-integration`).

See: [`P2_RESULTS.md`](P2_RESULTS.md), [`14.08.2026/TODO_FINAL_LIST.md`](../14.08.2026/TODO_FINAL_LIST.md).

---

## 3. Payment — YooKassa stub (verified)

**Problem:** Real YooKassa registration blocked without legal entity (самозанятый / ИП / ООО).

**Solution:** Local stub webhook:

- `POST /api/payment/yookassa/webhook` — maps simplified YooKassa payload → `ConfirmPayment`.
- `ConfirmPayment` **idempotent** — duplicate webhook → HTTP 200 (same subscription).
- Boosty webhook: flexible secret from body / headers / query; duplicate → 200.

**Smoke test:** ✅ passed (checkout → webhook → subscription active → merch allowed → repeat webhook idempotent).

Script: [`yookassa/smoke_yookassa_stub.sh`](yookassa/smoke_yookassa_stub.sh)  
Results: [`PAYMENT_SUCCESSFUL_TESTS.md`](PAYMENT_SUCCESSFUL_TESTS.md)

---

## 4. Boosty — decision (not blocked, deferred)

- Creator UI has **no** webhook/callback URL settings (Telegram/Discord bots only).
- EH redirect stub: `BOOSTY_CHECKOUT_URL?payment_id=…&plan=…&amount=…`
- **Decision:** Boosty = secondary / manual path; admin can grant merch access; real autosubscription deferred until provider API or polling strategy exists.

Docs: [`boosty/boosty_start.md`](boosty/boosty_start.md), [`boosty/boosty_plan.md`](boosty/boosty_plan.md), [`final_todo/final_todo.md`](final_todo/final_todo.md).

---

## 5. Gateway error mapping & STATUS_CODES §7 (done)

All four follow-ups from [`STATUS_CODES.md`](../../architecture/STATUS_CODES.md) §7:

| Item | Done |
|------|------|
| Normalize gRPC → HTTP via `handleRPCError` / `writeGRPCError` | ✅ All RPC-backed gateway routes |
| Inventory version conflict → **409** | ✅ `Aborted` mapping + inventory paths |
| Merch gate → **403** + `code: subscription_required` | ✅ Shop gRPC + gateway |
| OpenAPI in sync | ✅ YooKassa webhook, shop/payment status codes |

Shop changes:

- `ErrSubscriptionRequired` → gRPC `PermissionDenied` → HTTP 403.
- `PurchaseItem` returns gRPC errors (not `{success:false}` in 200 body).

Payment webhooks use `writePaymentConfirmError` (duplicate → 200).

---

## 6. API route convention & frontend fix (done)

**Canonical public prefix:** `/api/…` — **not** `/api/v1/…` in v1.0.x.

**Frontend rule:** axios `baseURL: '/api'` + paths **without** `/api` prefix  
(`api.post('/shop/purchase')` → `/api/shop/purchase`).

**Fixed double-prefix bug** in:

- `authorsApi.ts`, `paymentApi.ts`, `historyApi.ts`, `analyticsApi.ts`, `inventoryApi.ts`

**Also fixed:** game score stores (`flappy`, `memory`, `tower`) now use `api.post('/game/submit')` so JWT is sent (gateway requires auth).

**New doc:** [`confluence/architecture/API_ROUTES.md`](../../architecture/API_ROUTES.md) — full route table + role matrix + frontend module map.

README updated with API prefix section.

---

## 7. Frontend (today)

- **History page** + `historyApi.ts` + route in `App.tsx`.
- **Service API alignment** — paths normalized (see §6).
- **shopStore** — reads `error.response.data.error` for 403 merch messages.
- **Leaderboard** — uses `getLeaderboard()` from shared `api.ts`.

---

## 8. Docs & tech debt organization

- README refreshed for **v1.0.7** (Fulfillment, Notification, Payment storage, providers).
- System design: [`event-horizon-v1.0.7-system-design.md`](../../architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md).
- Tech debt split: `confluence/tech_debt/DONE/` vs `CURRENT_DEBT/`.
- Aggregated backlog: [`final_todo/final_todo.md`](final_todo/final_todo.md).

---

## Still open (not today)

From [`final_todo/final_todo.md`](final_todo/final_todo.md):

- ≥70% coverage + CI gate
- Live Boosty **signed** webhook verification (if spec exists)
- Real YooKassa credentials + public HTTPS webhook
- Frontend redesign / MCP smoke / optional binary commits
- Push refreshed **shop** image to registry if cluster still runs old tag

---

## Recommended commits

Execute in order. **Do not commit `.env`.**

### 1) Docs: architecture + README + API routes

```bash
git add README.md \
  confluence/architecture/EH_SCHEMAS.md \
  confluence/architecture/STATUS_CODES.md \
  confluence/architecture/API_ROUTES.md \
  confluence/architecture/SYSTEM_DESIGN/HOW_DOES_EH_WORK.md \
  confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md \
  confluence/history/2026-08/14.08.2026/TODO_FINAL_LIST.md \
  docs/openapi.yaml \
  services/gateway/api/openapi.yaml

git commit -m "$(cat <<'EOF'
docs: refresh v1.0.7 architecture, API routes, and status codes

Document /api/ prefix convention, complete STATUS_CODES follow-ups,
and add API_ROUTES.md with route table and RBAC matrix.
EOF
)"
```

### 2) Tech debt folder reorg + history notes

```bash
git add confluence/tech_debt/ \
  confluence/history/2026-08/18.08.2026/ \
  confluence/history/2026-08/19.08.2026/

git commit -m "$(cat <<'EOF'
docs: reorganize tech debt and add 19.08 session history

Split DONE vs CURRENT_DEBT; add payment smoke results, Boosty/YooKassa
notes, P2 results, and end-of-day summary.
EOF
)"
```

### 3) Gateway: payment stub, error mapping, merch gate

```bash
git add services/gateway/internal/app/gateway.go \
  services/payment/internal/service/payment_service.go \
  services/shop/internal/handler/grpc_handler.go \
  services/shop/internal/model/errors.go \
  services/shop/internal/service/shop_service.go

git commit -m "$(cat <<'EOF'
feat(gateway): yookassa webhook stub, gRPC error mapping, merch 403

Normalize RPC errors via handleRPCError; idempotent payment confirm;
shop purchase returns PermissionDenied for subscription_required.
EOF
)"
```

### 4) P2 backend hardening (outbox + slog + otelgrpc)

```bash
git add services/authors/internal/worker/outbox_worker.go \
  services/billing/internal/worker/outbox_worker.go \
  services/billing/internal/app/di.go \
  services/inventory/internal/worker/outbox_worker.go \
  services/payment/internal/worker/outbox_worker.go \
  services/gateway/internal/websocket/hub.go \
  services/game/internal/app/di.go \
  services/leaderboard/internal/app/di.go \
  services/shop/internal/app/di.go \
  services/shop/internal/repository/postgres_repo.go \
  services/shop/internal/repository/postgres_integration_test.go \
  services/shop/internal/worker/ \
  services/shop/migrations/20260819120000_add_outbox.sql

git commit -m "$(cat <<'EOF'
feat(backend): complete p2 hardening with shop outbox and slog migration

Transactional shop.purchased outbox, slog in workers and gateway WS,
otelgrpc wiring fixes across services.
EOF
)"
```

### 5) Tests + go modules/workspace updates

```bash
git add go.work go.work.sum \
  services/analytics/go.mod services/analytics/go.sum \
  services/authors/go.mod services/authors/go.sum \
  services/history/go.mod services/history/go.sum \
  services/payment/go.mod services/payment/go.sum \
  services/mcp/go.mod \
  services/balancer/go.mod \
  services/analytics/internal/repository/clickhouse/client_integration_test.go \
  services/authors/internal/repository/postgres_integration_test.go \
  services/balancer/internal/balancer/least_conn_test.go \
  services/fulfillment/internal/config/config_test.go \
  services/fulfillment/internal/service/fulfillment_test.go \
  services/history/internal/repository/postgres_integration_test.go \
  services/notification/internal/config/config_test.go \
  services/notification/internal/service/notify_test.go \
  services/payment/internal/repository/postgres_integration_test.go

git commit -m "$(cat <<'EOF'
test: add integration and smoke coverage across services

P1.5 integration tests and P2 unit smokes; go.work and module tidy
(including removal of nested balancer go.mod).
EOF
)"
```

### 6) Frontend: history, API path normalization, game JWT

```bash
git add frontend/src/App.tsx \
  frontend/src/components/Home/Home.tsx \
  frontend/src/components/History/ \
  frontend/src/components/Leaderboard/Leaderboard.tsx \
  frontend/src/components/Leaderboard/LeaderboardFull.tsx \
  frontend/src/services/api.ts \
  frontend/src/services/historyApi.ts \
  frontend/src/services/analyticsApi.ts \
  frontend/src/services/authorsApi.ts \
  frontend/src/services/inventoryApi.ts \
  frontend/src/services/paymentApi.ts \
  frontend/src/store/flappyStore.ts \
  frontend/src/store/memoryStore.ts \
  frontend/src/store/towerStore.ts \
  frontend/src/store/shopStore.ts

git commit -m "$(cat <<'EOF'
feat(frontend): history page and normalize API paths under /api

Fix double /api prefix in service clients; game submits use axios
with JWT; add History UI and shared getLeaderboard helper.
EOF
)"
```

### 7) Infra/build changes

```bash
git add Makefile deployments/docker-compose.cluster.yml

git commit -m "$(cat <<'EOF'
chore(devops): fix nats-hub build, kafka port, compose defaults

Build nats-hub from cmd/main.go; map Kafka to host 19092 to avoid
clash with game metrics on 9092.
EOF
)"
```

### 8) Optional — compiled binaries (only if you version them)

```bash
git add services/auth/auth-service \
  services/gateway/gateway-service \
  services/inventory/inventory-service \
  services/nats-hub/nats-hub \
  services/payment/payment-service \
  services/shop/shop-service

git commit -m "$(cat <<'EOF'
chore(build): refresh local service binaries
EOF
)"
```

---

## Quick verify after commits

```bash
make build-nats-hub   # or full build
make test-integration # if infra up
# Payment smoke (gateway + payment + auth running):
bash confluence/history/2026-08/19.08.2026/yookassa/smoke_yookassa_stub.sh
```

---

## Files touched today (reference)

**Gateway:** `services/gateway/internal/app/gateway.go`  
**Payment:** `services/payment/internal/service/payment_service.go`  
**Shop:** handler, model, service, outbox worker, migration  
**Frontend:** `api.ts`, all `*Api.ts`, game stores, leaderboard, history, shopStore  
**Docs:** README, STATUS_CODES, API_ROUTES, OpenAPI (×2), tech_debt, history/19.08.2026/*

---

*Session closed 19.08.2026 — ready for manual commit sequence above.*
