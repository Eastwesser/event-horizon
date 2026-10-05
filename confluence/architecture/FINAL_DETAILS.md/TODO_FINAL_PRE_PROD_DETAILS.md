# Pre-prod debt checklist — **v1.1.0** (post v1.0.9)

**How to use:** tick `- [ ]` → `- [x]` as you finish items, then update the **Progress** table at the bottom.

Source plan: [`README.md` → Планы на следующие спринты](../../../README.md#-планы-на-следующие-спринты).  
k6 archive: [`K6_WAVES/`](../../history/2026-10/05.10.2026/K6_WAVES/).

**Axes:** complexity XS–XL · risk low/med/high · dependency noted inline.

**v1.1.0 focus:** Wave 4.5 — **done 05.10.2026** (Helm chart · Consul skip · OpenAPI sync). C4 payouts stay deferred. Multi-VU purchase latency EXPLAIN — only if a real slow-purchase report appears.

---

## Wave 1 — Quick wins (XS–S, ~5–7 days)

Start here. Low risk, each item independent.

- [x] **PUT 403 investigation** — XS · low · ✅ **Resolved 03.10.2026.** Not a gateway bug. Backfill `login()` preferred `SEED_EMAIL` (author) over `SEED_ADMIN_*` after loading both dotenv files → author JWT; 128/129 set-8 cards are author-owned (OK), Пращник is admin-owned → `you can only edit your own items`. Fix: prefer `SEED_ADMIN_*` in `scripts/backfill-noiz-reviews.py`. Repro: author PUT → 403, admin PUT → 200.
- [x] **Emoji → SVG / PNG** — S · low · ✅ **Done 03.10.2026** (+ chrome sweep same day). `Icon` / `IconLabel` + `gameIcons`; nav, Home tiles, Profile, Leaderboard, Shop/inventory, headers, Berserk chips. **In-game** UI still emoji → game polish wave. Refund 7d design: [`REFUND_WINDOW_DESIGN.md`](../../history/2026-10/03.10.2026/REFUND_WINDOW_DESIGN.md).
- [x] **Retry + jitter (gateway)** — S · low · ✅ **Done 03.10.2026.** Unary interceptor + `client.Dial` on all gateway→service gRPC clients: 3 attempts, exp backoff + full jitter; retries only `Unavailable` / `ResourceExhausted`.
- [x] **Alerts → Telegram (Alertmanager)** — S · low · ✅ **Done 03.10.2026.** `alertmanager` service in compose; Telegram via `TELEGRAM_*` + entrypoint sed; noop receiver if unset; `alerts.yml` mounted (5× Down + HighOrderRate + InventoryDown).
- [x] **Circuit breaker + Bulkhead** — S · low · ✅ **Done 03.10.2026.** `MaxConcurrent: 32` bulkhead on each service breaker; `ErrBulkheadFull` → HTTP 503.
- [x] **Rate limiter** — S · low · ✅ **Done 03.10.2026.** Global ~100 req/s per user/token/IP (`AllowGlobal`); keep tighter submit/login/ws limits; `/health` `/ready` `/metrics` skipped.

**Warm-up order (recommended):** PUT 403 → Emoji → SVG → Retry + jitter.

---

## Wave 2 — Tech debt + content (M, ~2–3 weeks)

- [x] **/shop/items thin DTO** — M · low · — Logical wrap-up of v1.0.9 catalog work. *(Inventory list DTO already shipped; this is the shop-side twin.)*
- [x] **Реальные товары** — M · low · 2–3 days. Drop placeholders (Ключница Дракон…); keep Berserk CCG cards. Needs manual keep/delete list + SQL/script.
- [x] **Полиш игр** — M–L per game · med · ~1 week total. **One game per PR, sequentially** (not 5 games in parallel). ✅ **Wave 2 #6 closed 04.10.2026** (5/5).
  - [x] Flappy — textures / polish · ✅ **Done 04.10.2026** (draw helpers, parallax, Modal GO, shake/flap; physics untouched)
  - [x] Towers — animations / GAME OVER · ✅ **Done 04.10.2026** (draw helpers, Modal GO, shake/drop pulse; physics untouched)
  - [x] Hanoi — drag polish · ✅ **Done 04.10.2026** (target hover valid/invalid, invalid shake, Modal emoji strip; rules untouched)
  - [x] Memory — flip / skins · ✅ **Done 04.10.2026**
  - [x] Hexagon — gameplay polish · ✅ **Done 04.10.2026** (valid-hex drag highlight, place/clear pulse, invalid shake, Modal plain labels; rules untouched)
- [x] **Лампочки как бусты в играх** — M · med · Game service + UI
- [x] **Уровни сложности (1–20)** — M · med · Game service *(Flappy 1–10 pilot)*
- [x] **Достижения (achievements)** — M · med · ✅ **Done 04.10.2026.** Profile `achievements` + `user_achievements`; unlock on `score.updated` (+ GetProfile backfill); RU seed; SVG icon names; FE Profile API badges; toast new only (silent first seed).

---

## Wave 3 — Author registration (XL, ~3–4 weeks)

**Before C4:** decide monetization — what does an author get? (% of sales?)  
C1–C3 can proceed without payouts locked.

### C1. Application (S)

- [x] Route `/register-author` (auth user)
- [x] Form: name, portfolio, reason
- [x] `POST /api/authors/apply`
- [x] DB row: `pending`
- [x] Public register forced to `role=user` (no self-serve author)
- Follow-up (later PR): `author.application.submitted` → notify admins

### C2. Admin approval (M)

- [x] `/admin` → tab «Заявки»
- [x] List pending · approve / reject (double review → 400)
- [x] Approve → upsert authors profile + Auth `role=author` (gateway; revert on Auth fail)
- Follow-up (later PR): `author.application.approved` → notify author
- Follow-up (UX gap): JWT role refresh — re-login needed after approve until token refresh reloads role from Auth DB

### C3. Author dashboard (L)

- [x] `/author/dashboard` (author|admin)
- [x] My cards (list + filter + soft-delete visible + restore)
- [x] Create/edit card (inventory, own items; author_id = me)
- [x] Sales read-only (`GET /api/authors/me/sales`)
- [x] Author profile edit (`PUT /api/authors/me` + portfolio)
- [x] Burger «Автор» link

### C4. Payouts / analytics (L)

- [ ] Earnings (lamps / tickets)
- [ ] Payouts
- [ ] Monetization model locked

---

## Wave 4 — Infrastructure (parallel, M–L)

BEFORE START - CHECK WHAT ALREADY EXISTS 

- [x] **Helm charts for k3s** — M · med · DevOps · ✅ **Wave 4.5 done 05.10.2026.** Chart [`deployments/helm/event-horizon`](../../../deployments/helm/event-horizon/) wraps multi-container app manifests; `make deploy-k3s` prefers Helm. NATS/Postgres StatefulSets still follow-ups in k3s README.
- [x] **Service Discovery (Consul)** — M · med · Infra · ✅ **Evaluated → skip 05.10.2026.** Compose DNS + k3s CoreDNS enough. See [`CONSUL_EVALUATE.md`](../CONSUL_EVALUATE.md).
- [x] **k6 load + DB indexes** — M–L · med · ✅ **Wave 4 (2) correctness closed 05.10.2026.** `browse.js` / `purchase.js` (per-VU unowned); browse 5 VU p95 ~355ms; purchase 1 VU 100% checks. Shop N+1 owned → batch `ListOwnedItemIDs`; gateway no 503 after shop commit. Indexes: **not added** (EXPLAIN showed list query fine; multi-VU purchase p95 deferred — not a functional issue). Archive: [`K6_WAVE_6`](../../history/2026-10/05.10.2026/K6_WAVES/WAVES/K6_WAVE_6.md).
- [x] **Unit tests ≥70%** — M · low · ✅ **Wave 4 (1) done 05.10.2026.** All gated services ≥70% on `internal/service`. `scripts/coverage-gate.sh` + CI `coverage` job on `main` via [PR #2](https://github.com/Eastwesser/event-horizon/pull/2) merged (`f9ff55a`).
- [x] **OpenAPI docs** — M · low · ✅ **Wave 4.5 sync 05.10.2026.** Gateway-only SoT [`services/gateway/api/openapi.yaml`](../../../services/gateway/api/openapi.yaml) v1.1.0 mirrored to [`docs/openapi.yaml`](../../../docs/openapi.yaml). Added cancel, `/ready`, `/api/admin/users`. Swagger at `/docs`.

### Deploy status (partially done)

- [x] CI/CD GitHub Actions *(incl. coverage gate)*
- [x] Ansible
- [x] k3s
- [x] Helm charts *(app wrap; data-plane StatefulSets still open)*
- [x] Service Discovery *(evaluated: k3s DNS / Compose DNS — no Consul)*

---


## Wave 5 — Longer term

- [ ] **Bottleneck fix — server-side page/filter/sort** — L · med · **Defer until both:**
  1. Wave 2 #1 (`/shop/items` thin DTO) is done, **and**
  2. Real catalog pressure (~**500+** cards).  
  Until then: client catalog + v1.0.9 inventory list DTO is enough. Target: `/inventory/items` (and shop twin) page/filter/sort server-side so the client never loads the full catalog.
- [x] **Observability alerts** — ✅ wired with Wave 1 Alerts→Telegram (Alertmanager)
- [x] NATS cluster — works

---

## Progress

Update this table when you tick boxes above.

| Wave | Done / Total |
|------|--------------|
| 1 Quick wins | 6 / 6 |
| 2 Content + debt | 6 / 6 (+ 5 / 5 games) |
| 3 Author reg | 3 / 4 phases (C4 payouts deferred) |
| 4 Infra | **8 / 8** (CI/Ansible/k3s + unit tests + k6 + Helm + Consul skip + OpenAPI sync) |
| 5 Long term | 2 / 3 (NATS + alerts done; catalog page/filter still deferred) |

---

## Notes

- Do not mix Author registration (Wave 3) into a “quick commit” — too large.
- Wave 1–5 here is **future** work; v1.0.9 closed items are not listed (see [`REVIEW_RESULT.md`](../../history/2026-10/03.10.2026/REVIEW_RESULT.md)).
- Related debt log: [`confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`](../../tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md)
