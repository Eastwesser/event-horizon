Wave 2 #3 — Лампочки как бусты
Промпт Cursor'у:

text
Wave 2 #2 verified + pushed. Next: Wave 2 #3 — Лампочки как
бусты в играх.

Plan first, no code.

Design questions:
  1. What's a "boost" per game?
     - Flappy   → slow-motion (5 sec)?
     - Towers   → one block auto-corrects?
     - Hanoi    → hint next move?
     - Memory   → reveal one pair?
     - Hexagon  → highlight best move?
     - 2048     → undo one move?
     - Orbits    → ?
     - Companion → ?
  2. Cost: how many lamps per boost? Fixed (10?) or scales?
  3. Where: in-game button? Pre-game toggle? Per-session?
  4. Backend:
     - POST /api/game/boost { game_id, boost_type } → deduct
       lamps, return boost token
     - Game service or billing service? Which owns lamps?
  5. Frontend:
     - Button in GameShell
     - Confirm dialog with cost
     - Visual indicator when boost active
  6. Abuse guard:
     - Rate limit per minute?
     - Max boosts per session?
     - Do boosts affect leaderboard? Separate?
  7. Scope for v1:
     - All 8 games, or one pilot first (Flappy)?

Recommend: pilot on ONE game (Flappy) first, then replicate.

DO NOT TOUCH: seed, refund, author registration.
Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f thin DTO.

Wave 2 #2 — a881587 DB cleanup.

Сейчас:

⏸ Verify DB cleanup (13 пунктов)

⏸ Push 8 коммитов

⏸ Бэкап на флешку

⏸ Wave 2 #3 — Лампочки как бусты

Wave 2 (осталось):

🟧 #3 Лампочки как бусты

🟧 #4 Уровни сложности

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py