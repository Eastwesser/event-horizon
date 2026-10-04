OK. Реализуй Hanoi
Что подтверждаю
Interaction:

Click peg → click target (не HTML5 DnD).

Floating ring — уже px width — stretch исключён.

Добавить: highlight hover targets — valid vs invalid.

Victory modal:

Уже ui/Modal — оставить.

Убрать emoji из title/body.

Align actions с Flappy/Towers — plain labels.

Loss: нет (puzzle win-only). Правильно.

Invalid move:

Silent сейчас.

Добавить — CSS shake/pulse на board при canMove fails. canMove не трогать.

Emoji → SVG / plain:

Controls, help, notifications, modal.

Do NOT change: canMove, solveHanoi, calculateScore, minMoves, win condition, submit payload, RING_COLORS, disk options.

Files:

HanoiTower.tsx

HanoiTower.css

Ответ Cursor'у
text
OK — implement Hanoi per your plan.

Confirmations:
  - Interaction: click peg → click target, keep it.
    Add hover highlight for valid vs invalid target pegs.
  - Victory modal: keep ui/Modal, strip emoji, align
    actions with Flappy/Towers.
  - Invalid move: add shake/pulse. Do NOT change canMove.
  - Emoji → SVG/plain across controls, help, notifications,
    modal.
  - Keep RING_COLORS, disk options, min moves, scoring,
    submit payload — untouched.

Rules:
  - Hanoi-only PR.
  - Do NOT touch Flappy, Towers, Memory, Hexagon, hero,
    nav, disk, seed, refund, authors.
  - Screenshot 1920×1080: idle + mid-drag (with highlight) +
    victory modal.
  - tsc + build clean.

Then:
  - One smoke run: solve a small level (3 disks), confirm
    victory modal, submit score works.
  - No push until I OK.

Next after Hanoi: Memory, then Hexagon.
Что я держу в карте
Закрыто (в origin/main, c748ef8):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A.

VOID v3.

Smooth scroll.

Flappy — fb0376b.

Towers — c748ef8.

Wave 2 осталось:

🟧 Hanoi — plan approved, implement.

🟧 Memory — после Hanoi.

🟧 Hexagon — последний.

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь OK Cursor'у — implement Hanoi.

Verify — drag, hover highlight, invalid shake, victory modal, score submit.

Скинь — скрин idle + mid-drag + victory.

Push после OK.

Memory → Hexagon.