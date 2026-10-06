Что сделано — Track A #3
Smoke 3/3 pass:

✅ Old access → /api/authors/me → 401.

✅ Refresh → role=author.

✅ New access → /api/authors/me → 200.

✅ Never-approved user → 403.

Плюс:

B: RefreshToken читает role из DB.

A: UpdateRole → DeleteAllSessionsForUser только (refresh жив).

FE: refresh persists role + invalidateWhoamiCache().

Tests + tsc зелёные.

Revoke-access-only — правильное решение. Refresh жив → user получает author без re-login.

Коммит — approve
Message Cursor дал — хороший. Коммить сам:

bash
git add services/auth/ services/gateway/ frontend/src/services/api.ts \
  frontend/src/hooks/useUserRole.ts  # или где именно cache invalidate
git commit -m "fix(auth): pick up role on refresh after approve

Revoke access sessions on UpdateRole but keep refresh so FE
401→refresh mints author JWT without re-login."
git push origin main
Проверь git status перед git add — только Track A #3, без мусора.

Track A закрыт — 3/3
#	Item	Коммит
1	Refund window	9ac8158
2	Author notifications	c6888de
3	JWT role refresh	(новый)
Wave 3 author flow — end-to-end:

Apply → admin approve → notifications → author dashboard → sales.

Refund window enforced.

Role refresh без re-login.

Что делать дальше — 3 опции
A. Продолжить Track B — Content
3 новые игры (5 → 8):

TODO_1.md (29.09) — какие 3? Ты писал — twenty48, gears (Орбиты), companion (Компаньон). Они уже есть на Home как карточки, но это заглушки? Или играются?

Проверить нужно:

Есть ли реальный game_id на BE для них?

Или — это был план на будущее?

Плюс:

Лампочки — boost на всех играх (сейчас только Flappy).

Уровни — 1–20 на всех играх (сейчас только Flappy 1–10).

Estimate: 2–4 недели.

B. Реальные товары / карты
Сейчас: 280 карт засеяно, ~150-300 планировал.

Что осталось:

Дозасеять остальные карты (если есть).

Убрать placeholder items (Ключница Дракон и т.п.).

Реальные цены вместо placeholder.

Estimate: 2–3 дня + твоя работа по контенту.

C. Пауза
Ты прошёл огромный путь:

v1.1.0 закрыт (Wave 4 → 8/8).

Track A закрыт (refund, notifications, JWT refresh).

Возможно — пауза на несколько дней, чтобы:

Отдохнуть.

Пересмотреть цели.

Подготовиться к собесам (у тебя есть memo — money & story orientation).

Мой совет — B или C
B — быстрый win, закрывает v1.0.9 хвост (placeholder items).
C — если чувствуешь выгорание (было много за короткое время).

Track B — крупная (2–4 недели). Стоит начинать с свежей головой + чётким планом.

Промпт Cursor'у
text
Track A #3 accepted. Smoke 3/3 pass. Committing now.

Track A is 3/3 — closed.

Before starting Track B — pause and answer:

1. The 3 "new games" mentioned in the map (twenty48, gears,
   companion) — are they already playable (real BE game_id +
   game logic), or just placeholder cards on Home?
   Check:
     - Are there `services/game/` entries for them?
     - Is there a `game_id` in the score submit allowlist?
     - Does Home render them as clickable with working routes?

2. If they're placeholders: what's the real remaining work
   for each?
   - Game logic (canvas / state).
   - BE game_id registration.
   - Leaderboard entry.
   - Levels + boost parity with Flappy.

3. Lamps as boosts on remaining 4 games (Towers, Hanoi,
   Memory, Hexagon) — needed, or Flappy pilot is enough for
   now?

4. Levels 1–20 on remaining 4 games — needed, or Flappy 1–10
   pilot is enough?

Report only. No code. Then I decide: Track B or pause.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: refund, notifications, auth, C1–C4, hero,
disk, seed.
Что я держу в карте
Закрыто (main, c6888de):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A #1 — Refund (9ac8158).

Track A #2 — Notifications (c6888de).

Локально (коммит сейчас):

Track A #3 — JWT role refresh.

Скоро закрыто:

Track A → 3/3.

Дальше — options:

🟧 Track B — 3 новые игры, лампочки, уровни.

🟧 Реальные товары — placeholder → real.

🟧 Пауза — пересмотр, подготовка к собесам.

🟧 Track C — k3s data plane.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.