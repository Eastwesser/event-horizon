OK. Boost-first, потом levels game-by-game
Решение
Порядок:

Track B phase 1 — Boost на 7 игр (один PR).

Track B phase 2 — Levels game-by-game (7 PRs, по одному).

Не все скопом — по значимой части. Boost — одна общая фича (allowlist + UI). Levels — per game (баланс разный).

Промпт Cursor'у — Phase 1 Boost
text
Track B phase 1 — Boost на всех 8 играх. Plan first, no code.

Context:
  - Flappy: /game/boost/start работает, 10 lamps,
    boosted → unranked, 5s slow-mo.
  - Остальные 7 игр: BE allowlist только flappy; FE UI нет.
  - Deferred: levels (phase 2, game-by-game).

Scope phase 1:

BE (game service):
  1. StartBoost allowlist → все 8 game_id:
     hexagon, flappy, memory, towers, hanoi, twenty48,
     gears, companion.
  2. Cost 10 lamps для всех.
  3. Reuse active boost для той же сессии (idempotent per
     user + game + session).
  4. Boosted run → not ranked (submit level: 1, ranked:
     false) — same pattern as Flappy.

FE (shared):
  5. Общий hook useBoost (или эквивалент):
       - pre-game checkbox «Использовать boost (10
         лампочек)»
       - вызов /api/game/boost/start при старте
       - boost_id в submit
  6. Прокатить на 7 играх:
       - Towers, Hanoi, Memory, Hexagon (chrome already
         shared) — 4 игры
       - twenty48, gears, companion — 3 игры
  7. Балансовый эффект per game — минимальный:
       - Towers: block fall speed ×0.8
       - Hanoi: auto-suggest next move? (или временно —
         без эффекта, только unranked)
       - Memory: подсветка одной пары
       - Hexagon: valid-hex highlight ярче
       - twenty48: undo 1 шаг
       - gears: merge rate ×0.8
       - companion: без эффекта, только unranked + флаг
     Согласуй эффект с игрой; главное — не сломать physics.

Rules:
  - Один PR на phase 1.
  - Не трогать physics / collision / score (только
    то, что разрешено выше).
  - Flappy остаётся как есть.
  - tsc + go test clean.
  - Smoke: для каждой из 7 игр —
      · pre-game checkbox виден
      · StartBoost списывает 10 лампочек
      · boosted run → not ranked (проверить через
        leaderboard)
      · не-boosted run → ranked
  - Screenshots: 1 игра с checkbox + 1 boosted run +
    1 unranked toast.

No push until I OK.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: refund, notifications, auth, shop/Berserk,
C4, hero, disk, seed.
Что я держу в карте
Закрыто (main):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3 (refund, notifications, JWT).

Track B shop cleanup — 281 cards + 5 skins + 4 examples.

8 игр playable.

Track B parity — в работе:

🟧 Phase 1 — Boost на 7 игр (approve → plan).

🟧 Phase 2 — Levels game-by-game (7 PRs).

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.