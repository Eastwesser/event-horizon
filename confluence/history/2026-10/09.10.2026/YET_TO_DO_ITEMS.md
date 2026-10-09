# v1.1.0 — what shipped (07–09.10.2026) + what remains

High days: **7–9 Oct 2026** · jump **v1.0.9 → v1.1.0**.  
Companion docs: `08.10.2026/V1_1_0_FINAL_STATUS.md`, `TICKLIST_LAST_TODO_1.md`, `PARKED_WISHES.md`, `API_V1.md`.

---

## A. What we did in v1.1.0 (summary)

### Platform / API
- Public REST moved to **`/api/v1/*`** (Gateway + OpenAPI + FE `baseURL` + k6); legacy `/api/*` rewritten once
- Repo hygiene: scripts under `scripts/`, FE game folders, Emma→Denis docs
- CORE load: `make test-k6` / purchase green; legacy blast quarantined

### Product FE
- About, home blurbs, nick Modal, Gears/Tamagotchi naming
- Shop: merch chip, inventory ×qty, card-only filters, skins-by-game / themes niche, physical price floor
- Mobile: safe-area, GameShell `min-h-10` controls, 2048 swipe, shell insets (see § Mobile below)
- Avatar upload (Profile + Authors); shared `getNickname()` for LB submits
- Admin subscription labels + analytics MAU/DAU/retention copy

### Games
- Shared Boost UX; cosmic Flappy/Builder; Memonia map fix; 2048 drag
- Achievements amateur/pro/hero × 8 games
- Companion: daily **+1000 tickets** + care points + room chrome
- Gears: flower palettes + merge-on-contact; Hanoi ring gloss

### Backend events
- History dual NATS register subjects
- Nick → LB Redis refresh (`event.user.nickname.updated`)
- Record-beaten notification → LB deep link
- Seeds: themes/history/LB demo/card-artists authors

### Interview / Miro / Boosty / security
- Patterns, Elen Q, SERVICES_RU, SECURITY_AUDIT_RU, LOAD_POSTURE
- Miro PNGs + sticker paste sheet; Boosty marked done
- MCP exists (stdio Cursor tooling)

---

## B. Mobile adaptive — status

**Done for v1.1.0** (ticklist §13 = 3/3):

- `PageShell` / `index.css`: `env(safe-area-inset-*)`
- `GameShell`: `h-dvh`, control buttons `min-h-10 min-w-10`
- 2048 swipe + mouse drag; ScoreChip compact on small screens
- Navbar brand/logout collapse patterns on narrow widths

**Not a full design-system audit on every device.** If something feels broken on a specific phone, open a bug with screenshot — that is residual QA, not “mobile never started.”

---

## C. Leftovers / blockers / raw (next Cursor subscription)

These are honest remainders — **not** “forgot the wave.”

### Blockers for a clean release tag (you IRL)
| Item | Why |
|------|-----|
| Miro sticker polish | Ports `/api/v1`, Outbox only where real, MCP = Cursor→MCP not player path — `miro/MIRO_STICKERS_PASTE.md` |
| Rebuild gateway + LB + notification | Pick up nick/gift/record-beaten code |
| `make seed-lb-demo` + `make seed-card-artists` | Demo data on live stack |
| Optional `git tag v1.1.0` | Marker on commit, not a branch |

### Product locked (do not implement until new OK)
| Item | Note |
|------|------|
| **C4 payouts** (₽ → authors) | Lock D — `TRACK_C4_MONETIZATION_LOCK.md` |
| 108 **login** authors | Seed = synthetic rows only |
| Full 3D Sims / Dodo photoreal / OSS engines | Visual CSS close shipped instead |
| Avatar crop editor / platform chrome button list | Needs Denis list / design |

### Soft QA / polish (no blocker)
- Flappy LB eye-check after ranked save on deployed stack
- Profile zeros until you play ranked (not a display bug alone)
- Live Grafana vs METRICS targets on a real VM (`LOAD_POSTURE.md` has CORE vs model)
- AuthZ full matrix after large author seed
- Optional: Boosty post URL into `BOOSTY_DONE/`, per-game posts 1.1.1…

### Docs still slightly drift-prone
- Mermaid / `EH_SCHEMAS.md` title bump to v1.1.0 after Miro stickers finalized
- `confluence/architecture/API_ROUTES.md` may still say `/api/` in places — prefer OpenAPI + `API_V1.md` as source of truth

---

## D. Suggested next session order

1. IRL Miro stickers + rebuild + seeds  
2. Smoke: register → play → shop → companion gift → bell deep-link  
3. Tag `v1.1.0` when happy  
4. Only then reopen C4 / art epics if product asks  
