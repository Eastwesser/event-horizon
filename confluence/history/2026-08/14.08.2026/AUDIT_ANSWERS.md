# Audit answers — 14.08.2026

Cross-check of leftovers, FE/BE sync, and the 12 questions.  
Completed trackers from 13.08 archived to [`confluence/done/2026-08-13/`](../../done/2026-08-13/).

---

## Leftovers moved / cleaned

| Item | Action |
|------|--------|
| W1–W8 status, PAYMENT, AUTHORS/HISTORY/ANALYTICS, MCP_RAG, HARDENING, FRONTEND_*, FINAL_PRIORITY | → `confluence/done/2026-08-13/` |
| Pointer left in 13.08 | `MOVED_TO_DONE.md` |

Still open (not moved):

| Area | Notes |
|------|--------|
| `confluence/tech_debt/*` | Mix of ancient + real debt (load test, Envoy idea, Ansible). Do **not** treat as current sprint list. |
| Whole-service coverage ≥70% | Still deferred for handlers/DI; critical packages improved 14.08 (below) |
| OpenAPI gaps | auth refresh / logout / whoami / update-role missing from `docs/openapi.yaml` |
| Integration tests | Not every service has one (Q2) |
| Squirrel | Not used anywhere (Q4) |
| History FE page | No dedicated UI yet (API exists) |

---

## Coverage push (14.08)

| Package | Coverage | ≥70% |
|---------|----------|------|
| billing/converter | 100% | ✅ |
| inventory/converter | 71.4% | ✅ |
| inventory/service | **86.9%** | ✅ |
| payment/model | 87.5% | ✅ |
| gateway/circuit | 77.3% | ✅ |
| auth/jwt | ~93% | ✅ |
| auth/service | 64.9% | ❌ |
| billing/service | 30.6% | ❌ |
| shop/service | 12.6% | ❌ |
| analytics/clickhouse | 23.2% | ❌ |

**Verdict:** core/critical packages are in decent shape. Whole `internal/` ≥70% still needs repo-interface refactors for billing/shop/auth Redis+PG paths — not blocking product.

---

## Q1 — Unit tests per backend service?

| Service | Unit tests? | Notes |
|---------|-------------|--------|
| auth | ✅ | jwt, service, converter, config |
| billing | ✅ | converter 100%, validation, partial service |
| shop | ✅ | converter, merch_gate, helpers |
| inventory | ✅ | converter, service (~87%), model, cached_repo |
| game | ✅ thin | converter only |
| leaderboard | ✅ thin | converter |
| profile | ✅ thin | converter |
| gateway | ✅ | circuit, middleware roles |
| payment | ✅ | model |
| analytics | ✅ | clickhouse client/params |
| mcp | ✅ | pgxtool, rag |
| authors | ✅ thin | validate tests added 14.08 |
| history | ✅ thin | validate tests added 14.08 |
| balancer / fulfillment / notification | ❌ | no `*_test.go` |

**Answer:** Not each service, but **core money/auth/inventory/gateway logic is covered**. Thin converters on older services. OK to leave as-is for product; fill authors/history/service + balancer later if needed.

---

## Q2 — Integration tests (≥1 per service)?

| Service | Integration? |
|---------|----------------|
| auth | ✅ `tests/integration` (`//go:build integration`) |
| billing | ✅ testcontainers PG |
| shop | ✅ testcontainers PG |
| inventory | ✅ testcontainers PG (+ mongo build tag) |
| analytics / authors / history / payment / game / leaderboard / profile / gateway / … | ❌ none tagged |

**Answer:** **No — not one per service.** Only auth + billing + shop + inventory meet the bar today. Gap for: payment, authors, history, analytics, game, leaderboard, profile, gateway, fulfillment, notification.

---

## Q3 — Full routes list — FE ↔ BE sync?

### Gateway HTTP (source of truth)

Auth, profile, billing, shop, payment, authors, history, analytics, inventory CRUD, leaderboard, game submit, WS, `/openapi.yaml`, `/docs`, `/health`, `/ready`.

### OpenAPI (`docs/openapi.yaml`)

Covers most public API paths. **Missing vs gateway:**

- `POST /api/auth/refresh`
- `GET /api/auth/whoami`
- `POST /api/auth/logout`
- `POST /api/auth/update-role`

### Frontend routes (`App.tsx`)

| FE route | Backend used |
|----------|----------------|
| `/login` `/register` | auth |
| `/` + games | game submit |
| `/game/hanoi` | local only (OK) |
| `/leaderboard` | leaderboard + WS |
| `/profile` | profile / auth user |
| `/shop` `/infiniteshop` | shop + payment merch gate |
| `/inventory` | inventory |
| `/subscription` | payment |
| `/authors` | authors |
| `/analytics` | analytics (admin) |
| — | **no FE page for `/api/history`** |

**Answer:** Mostly in sync after F2–F7. Gaps: OpenAPI auth extras; **History UI missing**; inventory bulk/stats are API-only (admin tools OK).

---

## Q4 — Squirrel in each service?

**Answer: No.** Zero `Masterminds/squirrel` usage. SQL is raw string + `database/sql` / `pgx` / Mongo driver.  
Not a flaw unless you want query-builder consistency — current style matches course/inventory reference.

---

## Q5 — Redis used everywhere as needed?

| Service | Redis? | Role |
|---------|--------|------|
| auth | ✅ | sessions |
| billing | ✅ | balance cache |
| shop | ✅ | cache |
| inventory | ✅ | cache decorator |
| leaderboard | ✅ | Sorted Set top |
| gateway | ✅ | rate limit |
| payment | ✅ | subscription cache |
| authors | ✅ | profile cache |
| profile | ✅ | profile cache |
| mcp | ✅ | inspect tools |
| game / history / analytics / balancer / fulfillment / notification | ❌ | **not needed** (CH for analytics; PG append for history; game was cleaned of dead Redis) |

**Answer:** Yes for cache/session/leaderboard paths. Absence on game/history/analytics is intentional.

---

## Q6 — FE connected to backend? v1.0.6 vs newer stack?

**Answer:** FE is **connected** to the post-1.0.6 stack now:

- Payment / Authors / Analytics clients + pages
- Merch gate → Payment
- Inventory / Shop / Auth / Game / Leaderboard as before

README badge still says **v1.0.6**; runtime has Payment, Authors, History, Analytics, ClickHouse, MCP — treat as **1.0.6 + Stage-1 hardening** (docs/Miro still labeled 1.0.6). Bump version label when you cut a release.

Remaining FE lag: **History timeline page** only.

---

## Q7 — MCP on? Build / fine-tune?

**Answer:**

- Binary exists: `services/mcp/mcp-server`
- Cursor config: `.cursor/mcp.json` → that binary + NATS/Redis/PG/Prydwen
- Tools + offline TF-IDF RAG done; tests for validate + rag smoke

**You should:** ensure Cursor MCP panel shows `event-horizon` green; rebuild only if you change MCP code:

```bash
(cd services/mcp && CGO_ENABLED=0 go build -ldflags="-s -w" -o mcp-server ./cmd/main.go)
```

Fine-tune optional: richer Prydwen docs, Qdrant embeddings (deferred). Not blocking.

---

## Q8 — OpenAPI full?

**Answer: Almost.** Payment/authors/history/analytics/inventory present.  
Gaps: auth refresh/whoami/logout/update-role (see Q3). Gateway serves Swagger UI at `GET /docs` from hand-written YAML — not codegen.

---

## Q9 — Circuit breaker + microservice patterns?

| Pattern | Status |
|---------|--------|
| Circuit breaker (gateway → all gRPC) | ✅ |
| Rate limit (gateway Redis) | ✅ |
| Outbox → NATS | ✅ billing, inventory, authors, payment |
| Optimistic locking `version` | ✅ billing, shop, inventory |
| gRPC interceptors Recovery/Logger/Validate | ✅ live gRPC services |
| `/health` + `/ready` | ✅ metrics/HTTP |
| Clean Architecture layers | ✅ |
| Shop TX for purchase | ✅ |
| CQRS-ish leaderboard (NATS write / Redis read) | ✅ |

Shop does **not** use classic outbox table (publishes in purchase TX path / events differently) — billing/inventory/authors/payment are the outbox reference set.

---

## Q10 — swaggo / swagger?

**Answer:** **No swaggo.** Intentional (W1): single Gin gateway + `docs/openapi.yaml` + embedded Swagger UI at `/docs`.  
Not a flaw unless you want annotation-driven codegen. Do not add a second HTTP gateway.

---

## Q11 — zap or slog? Logging inspected?

**Answer: `log/slog` is the standard** (app DI + gRPC logger interceptors).  
Many workers/handlers still use `log.Printf` (gateway WS, outbox workers, leaderboard handler, mongo).  

Week notes deferred Zap. **Current state = slog + stdlib log leftovers.** Fine for now; unify workers to slog later.

---

## Q12 — DB transactions + outbox?

### Transactions (`BeginTx` / commit)

| Service | TX? |
|---------|-----|
| shop | ✅ purchase / inventory ops |
| billing | ✅ balance mutations |
| inventory (PG) | ✅ create/update/reserve + outbox |
| payment | ✅ checkout/confirm paths |
| authors | ✅ upsert + outbox insert |
| game / history / auth | mostly single statements (auth sessions via Redis) |

### Outbox workers

| Service | Outbox |
|---------|--------|
| billing | ✅ |
| inventory | ✅ |
| authors | ✅ |
| payment | ✅ |
| shop | ❌ no outbox table (TX yes) |
| others | ❌ |

**Answer:** Money/content write paths that need dual-write safety use TX+outbox on billing/inventory/authors/payment. Shop uses TX without outbox table.

---

## Recommended next (only if you care)

1. OpenAPI: add the 4 auth routes  
2. Integration smoke: payment + authors + history (1 test each)  
3. FE: History page  
4. Version bump README → 1.0.7 when you push images  
5. Optional: slog-ify outbox workers  

Coverage whole-service ≥70%: leave deferred unless you want repo-interface refactor week.
