# Wave 2 — TODO 1 (04.10.2026)

**Source:** [`TODO_FINAL_PRE_PROD_DETAILS.md`](../../../architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md), Wave 2.  
**Rule:** one item per PR; do not mix with Wave 3 (Author registration).  
**Sequential order, not parallel.**

## Order

1. `/shop/items` thin DTO — ~1 day — foundation (inventory list twin)
2. Реальные товары — 2–3 days — content cleanup
3. Лампочки как бусты — 1–2 weeks — game feature
4. Уровни сложности (1–20) — ~2 weeks — game feature
5. Достижения (achievements) — 1–2 weeks — new feature + DB
6. Полиш игр (5 games, one by one) — 1–2 days each
   - 6.1 Flappy — textures
   - 6.2 Towers — animations / GAME OVER
   - 6.3 Hanoi — drag polish
   - 6.4 Memory — flip / skins
   - 6.5 Hexagon — gameplay polish

**Total:** ~6–8 weeks.

## DoD per item

### 1. `/shop/items` thin DTO

- Gateway list DTO for `GET /api/shop/items` omits heavy / purchase-only fields.
- Detail / shop inventory paths unchanged where purchase fields are needed.
- Payload smaller (−20–30% where description was large); FE shop grid still works (primary catalog from inventory).

### 2. Реальные товары

- Delete named placeholders (Ключница Дракон, test брелоки, junk merch names, etc.).
- **Keep:** Berserk `карточка` + `game_skin` / `profile_theme` with a real `game_id`.
- **Drop:** junk merch / test names without meaningful game linkage.
- SQL/script + backup; `/shop` shows no placeholder names.

### 3. Лампочки как бусты

- `POST /api/game/boost` spends lamps, returns boost token.
- At least one game applies the bonus; FE confirm UI in-game.

### 4. Уровни сложности (1–20)

- Level picker UI; mechanics scale with level; score stores level; leaderboard filter if needed.

### 5. Достижения

- DB: `achievements` + `user_achievements`; event triggers; Profile badges + unlock toast.

### 6. Полиш игр

- One game = one PR. Visual/gameplay polish without regressions; before/after note or screenshot.

## Commit style

- `feat(shop): thin list DTO for /shop/items`
- `chore(seed): drop placeholder items, keep Berserk CCG`
- `feat(game): lamps as boosts`
- `feat(game): difficulty levels 1–20`
- `feat(profile): achievements`
- `feat(flappy): textures + polish`
- …

## Rebuild + verify

- FE: `make fe-preview`
- Backend: `bash scripts/rebuild-services.sh <svc>`
- Screenshots / network tab for catalog payload where relevant

## Not in this wave

- Refund window implementation (after Wave 2)
- Author registration (Wave 3)
- In-game emoji (part of game polish)
- `backfill-noiz-reviews.py` (PUT 403) — separate ticket if still unstaged
