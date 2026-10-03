Wave 2 #3 verified + pushed. Next: Wave 2 #4 — Уровни
сложности (1–20).

Plan first, no code.

Design questions:
  1. Which games get levels? All 8, or subset?
     - Flappy — speed, gap size, pipes density
     - Towers — block width, speed
     - Hanoi — number of rings (already 3–8?)
     - Memory — grid size, time limit
     - Hexagon — board size, tile spawn rate
     - 2048 — board size, spawn rate
     - Orbits — orb spawn, merge rate
     - Companion — probably no levels (soft game)
  2. UI: where does the player pick a level?
     - Pre-game dropdown in GameShell?
     - Separate page /leaderboard per level?
  3. Leaderboard: separate per level, or one combined?
     - Option A: separate (Flappy L5 ≠ Flappy L10)
     - Option B: combined, level as multiplier
     - Option C: single level only for ranked, others casual
  4. Rewards: do higher levels give more lamps/tickets?
  5. Persistence: user's current level saved?
  6. Scope for v1:
     - One game first (Flappy levels 1–10)?
     - All 8 games simultaneously?
  7. Interaction with Wave 2 #3 (boost):
     - Can you boost + pick level? Probably yes.
     - Does boost still skip leaderboard on all levels?

Recommend:
  - Pilot on Flappy: levels 1–10.
  - Separate leaderboard per level.
  - Save current level in user profile.
  - Rewards scale with level (multiplier).

DO NOT TOUCH: seed, refund, author registration.
Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f thin DTO.

Wave 2 #2 — a881587 DB cleanup.

Wave 2 #3 — 8c16926 lamps as boosts.

Сейчас:

⏸ Rebuild game + gateway

⏸ Verify Wave 2 #3 (19 пунктов)

⏸ Push (если не ушло)

⏸ Бэкап на флешку

🟧 Wave 2 #4 — Уровни сложности

Wave 2 (осталось):

🟧 #4 Уровни сложности

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py

