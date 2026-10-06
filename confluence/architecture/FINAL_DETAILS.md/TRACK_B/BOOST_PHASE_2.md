OK. План чистый. Реализуй Phase 1
Что подтверждаю
BE:

✅ Allowlist → все 8 game_id.

✅ Cost 10 lamps всем.

✅ Idempotent reuse active (без повторного списания).

✅ Тесты: Flappy + 1 non-Flappy + unknown rejected. Towers больше не fail.

✅ Flappy physics/slow-mo не трогать.

FE — shared:

✅ useGameBoost(gameId) — hook.

✅ useBoost / boostBusy / boostId / boosted.

✅ armBoost() → POST.

✅ Checkbox + RU copy как у Flappy.

✅ Toast: «Забег с boost — не попал в лидерборд».

✅ Wire boost_id в 6 stores/компонентов.

✅ Flappy — оставить как есть или тонкий refactor.

Effects (min, physics-safe):

Game	Effect
Towers	fall speed ×0.8
Hanoi	unranked only (suggest = rules risk) — правильно
Memory	flash one unmatched pair once at start
Hexagon	brighter valid-hex highlight
twenty48	one undo
gears	spawn/drop ×0.8 (не collision)
companion	unranked only
Flappy	unchanged (5s slow-mo)
Out of scope: Levels, refund/auth/shop/Berserk/C4, score formulas, collision.

Один момент — Hanoi suggest
Согласен. Suggest next move — rules risk. Unranked only — правильно для phase 1.

Если позже захочешь suggest — сначала проверить, не нарушает ли rules минимального числа ходов.

Один момент — Companion
Companion — soft tamagotchi. Boost эффект неясен.
Unranked only — ок для phase 1.

Позже — если будет эффект (ускорение feed/play/rest или скидка на care actions) — отдельно.

Ответ Cursor'у
text
OK — implement phase 1. One PR.

Confirmations:
  - BE allowlist → all 8 game_id.
  - Cost 10 lamps, idempotent reuse active.
  - Tests: flappy + 1 non-flappy + unknown rejected;
    towers no longer fails.
  - FE hook useGameBoost; wire boost_id into
    towerStore, gameStore, memoryStore, HanoiTower,
    Twenty48, Gears, Companion.
  - Flappy: leave as-is (thin refactor OK if trivial,
    no UX change).
  - Effects per your table (min, physics-safe).
  - Hanoi: unranked only (no suggest for v1).
  - Companion: unranked only.

Rules:
  - One PR phase 1.
  - Do NOT touch physics / collision / score formulas.
  - tsc + go test clean.
  - Smoke: для каждой из 7 игр — checkbox visible,
    StartBoost −10 lamps, boosted run → ranked: false,
    non-boosted run → ranked: true.
  - Screenshots: 1 checkbox + 1 boosted GO + 1 unranked
    toast.

No push until I OK.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: refund, notifications, auth, shop/Berserk,
C4, hero, disk, seed, levels (phase 2).
Что я держу в карте
Закрыто (main):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

Track B shop cleanup — 281 cards + 5 skins + 4 examples.

8 игр playable.

Track B parity — в работе:

🟧 Phase 1 — Boost на 7 игр (approve → implement).

🟧 Phase 2 — Levels game-by-game (7 PRs).

Track C (later):

🟧 k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.