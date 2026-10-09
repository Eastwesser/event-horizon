# Event Horizon v1.1.0 — final status pack

**As of:** 09.10.2026 · **Ticklist:** `TICKLIST_LAST_TODO_1.md`  
Miro board: https://miro.com/app/board/uXjVJLLg9us=/ · screenshots: `miro/miro_pics/` · review: `miro/MIRO_REVIEW.md`  
Mermaid / text schemas: `confluence/architecture/EH_SCHEMAS.md`, `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md`  
Ports cheat sheet: `SYSTEM_DESIGN_MIRO/INTERVIEW_QUESTIONS_EH.md`  
Security (RU audit): `SECURITY_AUDIT_RU.md` · baseline: `SECURITY_BASELINE.md`

---

## A. What was done in v1.1.0 (shipped in code / docs)

### Platform / repo
- Scripts moved under `scripts/` (ops / metrics / loadtest); Dockerfiles stay at root for `Makefile -f Dockerfile.*.bin .`
- FE: one folder per game; shop/inventory stores co-located
- Emma → Denis rename on active polish docs + shop SQL comments
- **HTTP API prefix:** canonical **`/api/v1/*`** (gateway + OpenAPI + FE `baseURL` + k6); legacy `/api/*` rewritten once on gateway

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

### Docs / representation (08–09.10)
- `WHAT_HAVE_WE_DONE.md`, `SECURITY_BASELINE.md`, **`SECURITY_AUDIT_RU.md`** (цифры + CSRF/XSS/SQL)
- `LOAD_RESULTS/`, `INTERVIEW_PATTERNS.md`, `QUESTIONS_FROM_ELEN.md` answers
- `SERVICES_RU_ONE_LINERS.md`, `MERMAID_INDEX.md`
- Miro PNGs + `miro/MIRO_REVIEW.md` (METRICS OK; sticker drifts listed)
- Boosty: **published** — `BOOSTY_DONE/BOOSTY_DONE.md`

---

## B. Left to do (honest buckets)

### B1 — Denis runtime — **DONE 08.10 night**
```bash
make seed-v110       # docker exec split A/B + themes + history + migrate-profile
make test-k6         # browse CORE green  (paths now /api/v1 — rebuild gateway after pull)
make test-k6-purchase
```
Still optional IRL: play ranked game for profile/LB zeros; git tag `v1.1.0` after Miro sticker polish; paste Boosty post URL into `BOOSTY_DONE`.

### B2 — Representation pack
| Item | Status | Action |
|------|--------|--------|
| **Miro scheme** | PNGs in `miro_pics/` | Fix drifts in `MIRO_REVIEW.md` (MCP grey, `/api/v1` stickers, Authors≠Mongo, ports) |
| **METRICS vs Miro** | **OK** | Planning targets match; not local k6 numbers |
| **Mermaid / EH_SCHEMAS** | Exists | Soft sync after sticker fixes |
| **Interview questions** | Ports + patterns + Elen Q | Expand when you ask |
| **Happy-path pitch** | Video (Denis), not Miro boxes | §F below still valid as storyboard |
| **Boosty** | **Done** | Optional post URL in `BOOSTY_DONE` |

### B3 — Product leftovers (not blockers for tag)
- Avatar upload; LB nicknamed seed ~10; nick→LB by user id
- 108 card artists as authors; author active-sub seed
- Shop filters by game/theme niche; art regen badge/fenechka
- Tamagotchi **tickets** gift (points gift already)
- Record-beaten notification deep-link
- AuthZ full matrix after author seed

### B4 — Parked / deferred (do **not** start)
Avatar polish wave, **108 authors**, **MCP**, **C4** payouts, flower skin / 3D / Dodo / Sims, optional 1.1.1…1.1.8 per-game Boosty posts.

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

API prefix: **`/api/v1/`** (legacy `/api/` rewritten). FE: `baseURL: '/api/v1'`.

---

## D. Commands (demo path)

```bash
cp .env.example .env          # JWT_SECRET, DB/Grafana passwords
make deploy                   # thin stack (NATS path; no Kafka)
make status                   # health/ready sweep
# FE:
cd frontend && npm i && npm run dev   # :5173 → proxy /api → :8079 (covers /api/v1)

# Optional:
make deploy-heavy             # Kafka
make test-k6                  # browse.js @ :8079 (/api/v1)
make seed-admin               # if needed
make seed-v110
```

---

## E. Boosty — done

Site: https://boosty.to/eastwesser  
Status file: `BOOSTY_DONE/BOOSTY_DONE.md` (tier ×2 + v1.1.0 post marked done 09.10).  
Draft text remains in `07.10.2026/BOOSTY_FIX.md` if you need to edit later.

---

## F. Short representation / happy path (pitch + Miro story)

**One sentence:** Event Horizon — Go microservices game platform: play → rank → earn lamps/tickets → shop/authors merch → subscribe → history & analytics, with real-time leaderboard over NATS/Redis/WS.

**Happy path (Denis video, not Miro boxes):**

1. **Register / Login** → Balancer `:8079` → Gateway `/api/v1/auth/*` → Auth (`50051`) → JWT in Redis; publish `user.registered` (+ `event.user.registered`) → History + Profile.  
2. **Play** (e.g. Flappy) → Game (`50052`) submit score → NATS `score.updated` → Leaderboard Redis SS + Profile bests + achievements unlock.  
3. **Optional Boost** (−lamps via Billing) → run **not ranked**, no shop profit.  
4. **Leaderboard live** → Redis → WS `/ws/leaderboard` → FE.  
5. **Shop** → catalog → spend **tickets** → purchase events → Fulfillment + Notification; physical merch needs Payment subscription.  
6. **Authors** → publish goods after author role + sub rules.  
7. **Admin / Analytics** → users, inventory stats, DAU/MAU/retention (ClickHouse).

**Artifacts:** Miro PNGs + `MIRO_REVIEW.md` + this happy path + Mermaid (bump title to v1.1.0 when stickers cleaned).

---

## G. Close-out order (remaining)

1. ~~Seeds + CORE k6~~  
2. ~~Boosty publish~~  
3. Miro sticker polish from `MIRO_REVIEW.md` (Denis IRL)  
4. Rebuild gateway after `/api/v1` pull; optional re-run `make test-k6`  
5. `git tag v1.1.0` when you are happy with Miro  
6. Parked: avatar / 108 / MCP / C4 — later  
