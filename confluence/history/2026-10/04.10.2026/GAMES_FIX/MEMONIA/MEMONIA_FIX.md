Push. Дальше Memory
Что принято — 42076be
✅ Valid/invalid peg hover — при selected.

✅ Board shake + pulse — при illegal.

✅ Victory Modal — plain labels, как Flappy/Towers.

✅ Emoji stripped.

✅ Map: Hanoi checked off.

Ответ Cursor'у
text
Hanoi accepted. Push 42076be, then start Memory.

Push:
  - 42076be (Hanoi polish)
Report new HEAD.

Next: Memory polish (one PR, same rules).

Memory current state (from MemoryGame.tsx / memory.css /
MemoryBoard.tsx / MemoryCard.tsx):
  - Flip animation: was earlier a regression; verify
    current state (flip smooth? not snapping?).
  - Skins (Карточки со зверями) — verify still working.
  - Timer / moves / pairs counters — do they use emoji?
  - GAME OVER / Victory — check if it has a modal yet.
  - Emoji in controls/help/notifications.

Plan first, no code.

Scope for v1:
  1. Flip animation — smooth, keep card art visible during
     flip; no snap. If broken, restore.
  2. Victory modal — ui/Modal, plain labels, Flappy/Towers
     pattern. (If none yet, add.)
  3. Shake/pulse on mismatch (two cards revealed but not
     a pair) — subtle.
  4. Emoji → SVG / plain across counters, controls, help,
     notifications, modal.
  5. Skins — keep «Карточки со зверями» working.

Do NOT change:
  - reveal / match logic
  - board size, pair count, timer, scoring
  - submit payload
  - skins behavior

Rules:
  - Memory-only PR.
  - Do NOT touch Flappy, Towers, Hanoi, Hexagon, hero,
    nav, disk, seed, refund, authors.
  - Screenshots: idle, mid-flip, victory modal.
  - tsc + build clean.
  - Smoke: flip, match a pair, mismatch shake, complete
    board, victory modal, submit score.

Then: Hexagon (last of Wave 2 #6).
No push until I OK.
Что я держу в карте
Закрыто (в origin/main, c748ef8):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A, VOID v3, smooth scroll.

Flappy fb0376b, Towers c748ef8.

Локально (пуш сейчас):

42076be — Hanoi.

Wave 2 осталось:

🟧 Memory (следующее).

🟧 Hexagon (последнее).

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь Cursor'у — push + Memory plan.

Verify Memory — flip, skins, victory, shake.

Скинь — скрин.

Push после OK.

Hexagon — последний.

Wave 2 — 4 из 5 игр закрыто. Hexagon остался.