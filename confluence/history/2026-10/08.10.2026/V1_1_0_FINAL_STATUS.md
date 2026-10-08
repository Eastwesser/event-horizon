# Event Horizon v1.1.0 — final status pack

**HEAD:** `e552125` · **Ticklist:** `TICKLIST_LAST_TODO_1.md` (~123 done / ~112 open checkboxes; many open = product/wish/QA)

Miro board (refresh): https://miro.com/app/board/uXjVJLLg9us=/  
Mermaid / text schemas: `confluence/architecture/EH_SCHEMAS.md`, `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md`  
Ports cheat sheet: `SYSTEM_DESIGN_MIRO/INTERVIEW_QUESTIONS_EH.md`

---

## A. What was done in v1.1.0 (shipped in code / docs)

### Platform / repo
- Scripts moved under `scripts/` (ops / metrics / loadtest); Dockerfiles stay at root for `Makefile -f Dockerfile.*.bin .`
- FE: one folder per game; shop/inventory stores co-located
- Emma → Denis rename on active polish docs + shop SQL comments

### Site / chrome
- About page (navbar brand → `/about`); professional copy
- Home blurbs: Builder / Gears / Tamagotchi; hero spacing polish
- Nick change via in-app Modal (no browser `confirm`)
- Profile/LB labels: Gears, Tamagotchi
- Subscription UI: **Базовый** / **Расширенный** + Boosty renew link
- Mobile: safe-area, shell insets, brand/logout collapse, game touch targets

### Shop / inventory / admin
- Card filters only on Cards; **Мерч** type chip
- Inventory FE group-by `item_id` with ×qty
- Cancel copy: «предмет будет удалён…»; create price = tickets; merch option
- Themes vs skins icons (star / palette); seed SQL for themes + cosmic skin rename
- Physical floor 100_000 tickets (SQL); painting example hidden
- Admin Top-N: tickets icon not ₽; retention D0…D7 copy

### History / bugs
- Gateway dual-publish `event.user.registered` + `user.registered`
- History worker normalizes `event.user.registered` → `user.registered`
- Empty history UX + `scripts/seed-history-demo.sql`
- F12: `/api/api/profile` fixed; LB `undefined` spam fixed; DEV-gated API logs

### Games
- Shared boost help + collapsed BoostCheckbox; unranked boosted runs
- Equal-width GO; «На главную» → `/#games`; SVG check on save
- Cosmic Flappy pipes / Builder blocks (no rainbow palette)
- Memonia fruit↔animal map fixed; 2048 swipe + mouse drag
- Balance on all game shells; Pancaker plural; Hanoi/Builder Track-B polish
- Achievements: amateur / pro / hero × 8 games (+ first_play = новичок)

### Docs written this wave
- `WHAT_HAVE_WE_DONE.md`, `SECURITY_BASELINE.md`, `LOAD_RESULTS/README.md`
- `FEEDBACK_VoiceM_2_1.md`, `FEEDBACK_VoiceM_3_1.md`
- `INTERVIEW_QUESTIONS_EH.md` (ports / patterns)

---

## B. Left to do (honest buckets)

### B1 — Denis runtime — **DONE 08.10 night**
```bash
make seed-v110       # docker exec split A/B + themes + history + migrate-profile
make test-k6         # browse CORE green
make test-k6-purchase
```
Still optional IRL: play ranked game for profile/LB zeros; git tag `v1.1.0` after Miro+Boosty.

### B2 — Representation pack (you called these out — still open)
| Item | Status | Action |
|------|--------|--------|
| **Miro scheme** | Old (v1.0.6/7 era note) | Update board → export PNG → `FINAL_SYSTEM_DESIGN_MIRO_SCHEME.md` + `08.10.2026/` |
| **Mermaid / EH_SCHEMAS** | Exists, may drift | Cross-check vs Miro after export |
| **Interview questions** | Ports filled; your full Q list later | Keep `INTERVIEW_QUESTIONS_EH.md`; expand when you ask |
| **Ports list** | Done in interview doc + §C below | Copy into Miro node labels |
| **Commands** | `make deploy` (+ §D) | Enough for demo; document seeds |
| **Boosty copy** | Draft base exists; ×2 + v1.1.0 post **not published** | §E below — paste into Boosty UI |
| **Happy-path pitch** | Not a single one-pager yet | §F below — use with Miro |

### B3 — Product / FE leftovers (code or content)
- Avatar upload; LB nicknamed seed ~10; nick→LB by user id
- 108 card artists as authors; author active-sub seed
- Shop filters by game/theme niche; art regen badge/fenechka
- Tamagotchi **tickets** gift (points gift already)
- Record-beaten notification deep-link
- Full CSRF/XSS/SQL + AuthZ pass (beyond baseline note)
- Per-service RU explainers, patterns write-up, SQL interview list
- MCP / Tetiva — **last**

### B4 — Parked / deferred (do not block 1.1.0 tag)
Flower skin, 3D engines, Dodo boxes, C4 payouts lock, optional 1.1.1…1.1.8 per-game Boosty posts.

---

## C. Ports (all services)

### Public / edge
| Component | Port |
|-----------|------|
| React (dev) | `5173` |
| Balancer (public HTTP + WS) | `8079` |
| Gateway replicas | `8081`–`8083` |
| OpenAPI / docs | via `:8079/docs`, `/openapi.yaml` |
| WS leaderboard | `ws://localhost:8079/ws/leaderboard` |

### gRPC mesh
| Service | gRPC | Primary data |
|---------|------|----------------|
| Auth | `50051` | PG `5460` + Redis sessions |
| Game | `50052` | PG `5461` |
| Billing | `50053` | PG `5462` + Redis |
| Leaderboard | `50054` | PG `5463` + Redis SS `6382` |
| Shop | `50055` | PG `5465` + Redis |
| Analytics | `50057` | ClickHouse `8123`/`9000` |
| Payment | `50058` | sub / CanPurchaseMerch |
| Inventory | `50059` | PG + Mongo + Outbox |
| Profile | `50060` | PG `5464` |
| Authors | `50061` | PG `5468` + Redis |
| History | `50062` | PG `5469` |
| Fulfillment / Notification / NATS Hub | compose | NATS consumers / stream bootstrap |

### Infra
| Component | Port |
|-----------|------|
| NATS JetStream | `4222` (+ cluster peers) |
| ClickHouse | `8123` HTTP / `9000` native |
| Prometheus / Grafana / Jaeger | compose defaults (see `make status`) |

API prefix: **`/api/`** only. FE: `baseURL: '/api'` + relative paths.

---

## D. Commands (demo path)

```bash
cp .env.example .env          # JWT_SECRET, DB/Grafana passwords
make deploy                   # thin stack (NATS path; no Kafka)
make status                   # health/ready sweep
# FE:
cd frontend && npm i && npm run dev   # :5173 → proxy /api → :8079

# Optional:
make deploy-heavy             # Kafka
make test-k6                  # browse.js @ :8079
make seed-admin               # if needed
```

Seeds after first deploy: see §B1.

---

## E. What to write on Boosty (paste-ready)

Site: https://boosty.to/eastwesser

### Profile / about (refresh)
Keep the existing Russian intro in `BOOSTY_FIX.md`, but swap «pet-project» framing for:  
**«игровая микросервисная платформа на Go (Clean Architecture, gRPC, NATS) + React клиент»**.

### Tier 1 — Базовый (refresh benefits)
- Поддержка инфраструктуры и доменов  
- Новости релизов раньше публичных каналов  
- Упоминание в списке поддержавших (по желанию)  
- Доступ к мерчу в магазине Event Horizon (по правилам подписки in-app)

### Tier 2 — Расширенный (×2) — **update this tier**
- Всё из Базового  
- Ранний доступ к заметкам по архитектуре / схемам (Miro / Mermaid)  
- Голос по приоритету мини-игр (1.1.1…1.1.8)  
- Расширенные бонусы сообщества (как договоримся in-app)

### Post — «Event Horizon v1.1.0»
Suggested body (RU):

> Вышел **Event Horizon v1.1.0** — волна полировки продукта и платформы.  
> **Платформа:** чище структура репо и FE по играм; About; мобильные отступы/touch; история событий по NATS; ачивки amateur/pro/hero на все 8 игр.  
> **Магазин:** мерч-тип, инвентарь без дублей (×N), билетики вместо ₽ в админке, темы/скины (космические вместо «радужных»), порог физ. мерча 100 000.  
> **Игры:** единый Boost (не в LB), космические скины, фикс Memonia, 2048 drag, Balance везде, Gears/Tamagotchi.  
> Дальше: обновление Miro-схемы, прогон highload (`make test-k6`), и серии постов 1.1.1–1.1.8 по каждой игре.  
> Поддержать: https://boosty.to/eastwesser

---

## F. Short representation / happy path (pitch + Miro story)

**One sentence:** Event Horizon — Go microservices game platform: play → rank → earn lamps/tickets → shop/authors merch → subscribe → history & analytics, with real-time leaderboard over NATS/Redis/WS.

**Happy path (draw this on Miro left→right):**

1. **Register / Login** → Balancer `:8079` → Gateway → Auth (`50051`) → JWT in Redis; publish `user.registered` (+ `event.user.registered`) → History + Profile.  
2. **Play** (e.g. Flappy) → Game (`50052`) submit score → NATS `score.updated` → Leaderboard Redis SS + Profile bests + achievements unlock.  
3. **Optional Boost** (−lamps via Billing) → run **not ranked**, no shop profit.  
4. **Leaderboard live** → Redis → WS `/ws/leaderboard` → FE.  
5. **Shop** → list Inventory/Shop catalog → spend **tickets** → `shop.purchased` / `purchase.paid` → Fulfillment + Notification; physical merch needs Payment subscription.  
6. **Authors** → publish goods (карточка/мерч/…) after author role + sub rules.  
7. **Admin / Analytics** → users, inventory stats, DAU/MAU/retention (ClickHouse).

**Artifacts to keep in sync:** Miro PNG + this happy path + Mermaid in `event-horizon-v1.0.7-system-design.md` (bump title to v1.1.0 when Miro refreshed).

---

## G. Suggested close-out order (next sessions)

1. `make deploy` + seeds + smoke (B1)  
2. Miro refresh + export image (B2)  
3. Publish Boosty tier×2 + v1.1.0 post (E)  
4. `make test-k6` → `LOAD_RESULTS/`  
5. Interview pack leftovers when you bring your Q list  
6. MCP/Tetiva last  
