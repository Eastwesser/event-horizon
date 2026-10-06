Phase 1 boost accepted (ee75936 + build-fix).

Track B phase 2 — Levels, game-by-game.

Plan for the FIRST game only. No code yet.

Recommended first: Towers (linear physics, easiest to
balance).

Plan:
  1. Levels 1–10 (10 difficulty tiers).
  2. What changes per level:
     - block fall speed,
     - block width variance,
     - target height,
     - reward multiplier.
  3. Pre-game picker UI: reuse Flappy's pattern, or
     per-game design — recommend.
  4. BE: SubmitScore already has `level` param; ensure
     towers sends real level (not 1).
  5. Leaderboard: separate per level (Redis key already
     supports leaderboard:<game>:<level>).
  6. Profile: show highest level per game.
  7. Home: no change.

Questions:
  Q1. Do levels also apply to Memory (pair count) / Hanoi
      (disk count) / Hexagon (board size) / 2048 (target
      tile) / Gears (merge goal) / Companion (care
      cadence)?
      Or restrict levels to physics-driven games
      (Towers, Flappy already done)?
  Q2. Boost + level interaction — both allowed, boosted
      still unranked. Confirm.

Report plan + Q1/Q2 answers. No code until I OK.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: refund, notifications, auth, shop/Berserk,
C4, hero, disk, seed, boost logic.
Что я держу в карте
Закрыто (main):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

Track B shop cleanup.

8 игр playable.

Локально (коммит + push сейчас):

ee75936 — Phase 1 boost.

build-fix (unused vars).

Track B parity — в работе:

🟧 Phase 1 — Boost на 7 играх (verify ✅, push).

🟧 Phase 2 — Levels game-by-game (plan → approve).

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.