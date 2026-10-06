# Pre-prod debt checklist — **v1.1.0+** (living)

**How to use:** tick `- [ ]` → `- [x]` as you finish items, then update the **Progress** table at the bottom.

Source plan: [`README.md` → Планы на следующие спринты](../../../README.md#-планы-на-следующие-спринты).  
Tracks map: [`FINAL_TRACKS.md`](./FINAL_TRACKS.md).  
**Games** backlog (parked): [`GAME_INSIGHTS.md`](../../history/2026-10/06.10.2026/GAME_INSIGHTS.md).  
**Platform** backlog (active): [`PLATFORM_INSIGHTS.md`](../../history/2026-10/06.10.2026/PLATFORM_INSIGHTS.md).

**Axes:** complexity XS–XL · risk low/med/high · dependency noted inline.

---

## Wave 1 — Quick wins (XS–S) — ✅ closed

- [x] **PUT 403 investigation** — XS · low · ✅ 03.10.2026
- [x] **Emoji → SVG / PNG** (app chrome + shop) — S · low · ✅ 03.10.2026 · *In-game emoji still open → see GAME_INSIGHTS*
- [x] **Retry + jitter (gateway)** — S · low · ✅ 03.10.2026
- [x] **Alerts → Telegram (Alertmanager)** — S · low · ✅ 03.10.2026
- [x] **Circuit breaker + Bulkhead** — S · low · ✅ 03.10.2026
- [x] **Rate limiter** — S · low · ✅ 03.10.2026

---

## Wave 2 — Tech debt + content (M) — ✅ closed (parity leftovers → Track B / insights)

- [x] **/shop/items thin DTO** — M · low
- [x] **Реальные товары (cleanup placeholders)** — M · low · ✅ Shop cleanup v2 06.10.2026 · Berserk CCG ~281 untouched · 5 skins + 4 examples
- [x] **Полиш игр (Wave 2 #6)** — 5/5 classic · ✅ 04.10.2026 · *Further polish → GAME_INSIGHTS*
- [x] **Лампочки как бусты** — M · med · ✅ Phase 1 allowlist all 8 games (`ee75936`) · *Per-game effects + UX copy → GAME_INSIGHTS §2–11*
- [x] **Уровни (Flappy 1–10 pilot)** — M · med · ✅ Flappy only · *Towers+ rest → Track B phase 2 / GAME_INSIGHTS*
- [x] **Достижения** — M · med · ✅ 04.10.2026
- [x] **3 new games playable** — twenty48 / gears / companion · ✅ (parity ≠ polish)

---

## Wave 3 — Author registration (XL)

### C1. Application (S) — ✅

- [x] Route `/register-author` · form · `POST /api/authors/apply` · `pending` · public register → `role=user`
- [x] `author.application.submitted` → notify admins · ✅ Track A #2

### C2. Admin approval (M) — ✅

- [x] `/admin` заявки · approve/reject · Auth `role=author`
- [x] `author.application.approved` → notify author · ✅ Track A #2
- [x] JWT role refresh without re-login · ✅ Track A #3 (`f784419` lineage)

### C3. Author dashboard (L) — ✅

- [x] `/author/dashboard` · cards · sales · profile · burger link

### C4. Payouts / analytics (L) — 🟧 deferred

- [ ] Earnings (lamps / tickets)
- [ ] Payouts
- [ ] Monetization model locked

**Also closed with Track A #1:** Refund window 7d + fulfilled guard · ✅

---

## Wave 4 — Infrastructure — ✅ closed (data-plane leftover = Track C)

- [x] Helm charts for k3s (app wrap) · Wave 4.5 · 05.10.2026
- [x] Consul evaluated → skip · 05.10.2026
- [x] k6 correctness · 05.10.2026
- [x] Unit tests ≥70% + CI coverage gate · PR #2
- [x] OpenAPI sync v1.1.0
- [x] CI/CD · Ansible · k3s · Compose DNS

### Track C leftover

- [ ] **k3s data plane** — NATS + Postgres StatefulSets in Helm · only if prod demo needs it

---

## Wave 5 — Longer term

- [ ] **Bottleneck — server-side page/filter/sort** — L · med · defer until ~**500+** cards (today ~281)
- [x] Observability alerts
- [x] NATS cluster

---

## Track A — Product closure — ✅ 3/3

| # | Item | Status |
|---|------|--------|
| 1 | Refund window 7d + fulfilled | ✅ |
| 2 | Author notifications | ✅ |
| 3 | JWT role refresh | ✅ |

---

## Track B — Games (parked → GAME_INSIGHTS)

| Phase | Item | Status |
|-------|------|--------|
| Shop cleanup | Keep Berserk + 5 skins + 4 examples; hide junk | ✅ |
| Phase 1 Boost | Allowlist 8 games + FE hook | ✅ (`ee75936` + build-fix) |
| Hotfix → UX → effects → levels | See GAME_INSIGHTS | 🟧 **parked** |

## Track D — Platform (active → PLATFORM_INSIGHTS)

| Phase | Item | Status |
|-------|------|--------|
| P1 Shop examples & art | Replace/remove placeholders; unique arts | ✅ |
| P2 Shop spinner + load | Center spinner; profile slow /shop | 🟧 **next** |
| P3 Chrome polish | Non-game buttons | 🟧 |

---

## Track C — Infra (later)

- [ ] k3s NATS + Postgres StatefulSets

---

## Deferred (do not start unless Emma OK)

- [ ] C4 payouts / monetization lock
- [ ] Wave 2 #5b shop/boost achievements
- [ ] Multi-VU purchase EXPLAIN (only if real slow report)
- [ ] Catalog bottleneck (500+ cards)
- [ ] Cursor pull-in
- [ ] `backfill-noiz-reviews.py` one-shot
- [ ] 3D / OSS game remakes
- [ ] Companion daily-gift economy (full)
- [ ] Mass multi-account LB QA seeding

---

## Progress

| Wave / Track | Done / Total |
|--------------|--------------|
| 1 Quick wins | **6 / 6** |
| 2 Content + debt | **7 / 7** (+ 5/5 polish wave; parity → Track B) |
| 3 Author reg | **3 / 4** phases (C4 deferred) + refund + notif + JWT |
| 4 Infra | **8 / 8** app; data-plane → Track C |
| 5 Long term | **2 / 3** (bottleneck deferred) |
| Track A Product | **3 / 3** |
| Track B Games | parked (GAME_INSIGHTS) |
| Track D Platform | **0 / 3** phases (P1 next) |
| Track C Infra | **0 / 1** |

---

## Notes

- **GAME_INSIGHTS.md** — games only (parked).
- **PLATFORM_INSIGHTS.md** — shop/chrome/infra (active).
- Do not mix C4 into a “quick commit”.
- Berserk CCG cards (~281) = real merch — never mutate without explicit OK.
- Standing: agent does not push; writes commit messages; Emma OKs push.
- Related debt: [`STILL_TECH_DEBT.md`](../../tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md)
