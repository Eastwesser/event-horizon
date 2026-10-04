OK. Реализуй Memory
Что подтверждаю
Flip:

CSS 3D rotateY 0.5s — smooth.

Match snap risk: matched cards unmount flip DOM → art может pop.

Fix: fade matched card с __inner still present, или keep face visible через fade.

Mid-flip faces остаются видимыми.

Skins:

«Карточки со зверями» — wired (useSkins + memory_animal_cards). Keep.

Victory:

Сейчас — .memory-game-over overlay.

Заменить на ui/Modal (Flappy / Towers / Hanoi pattern).

Mismatch:

Сейчас — flip-back, без shake.

Добавить board shake / pulse — UI-local, checkMatch не трогать.

Timer:

Нет в UI. Правильно — pairs / moves / combo / score.

Emoji:

Chrome (stats, skin btn, modal, help) → SVG.

Face emoji — контент (фрукты / животные) — keep.

Do NOT change:

flipCard / checkMatch — rules.

15 pairs / board size.

Score formula.

Submit payload.

Skin ownership.

Files: MemoryGame.tsx, MemoryCard.tsx, MemoryBoard.tsx (if needed), memory.css.

Prefer no memoryStore edits — UI observe preferred.

Ответ Cursor'у
text
OK — implement Memory per your plan.

Confirmations:
  - Flip: keep 3D rotateY. Fix match path so face art
    doesn't pop. Prefer fade with __inner still present,
    or another smooth path. Mid-flip faces stay visible.
  - Victory: replace .memory-game-over with ui/Modal,
    Flappy/Towers/Hanoi pattern, plain labels.
  - Mismatch: board shake/pulse. UI-local. Do NOT touch
    checkMatch or score.
  - Chrome emoji → SVG/plain in stats, controls, help,
    toasts, modal. Keep fruit/animal face emoji as
    content.
  - Skins: leave toggle + animal remap intact.

Rules:
  - Memory-only PR.
  - Do NOT touch Flappy, Towers, Hanoi, Hexagon, hero,
    nav, disk, seed, refund, authors.
  - memoryStore: only touch if a mismatch signal is
    unavoidable. Prefer UI-side observation.
  - tsc + build clean.
  - Screenshots 1920×1080: idle, mid-flip, victory modal.
  - Smoke: flip, match, mismatch shake, clear board,
    modal, submit, animal skin.

No push until I OK. Next after Memory: Hexagon.
Что я держу в карте
Закрыто (в origin/main, 42076be):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A, VOID v3, smooth scroll.

Flappy fb0376b, Towers c748ef8, Hanoi 42076be.

Wave 2 осталось:

🟧 Memory (в работе).

🟧 Hexagon (последнее).

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь OK Cursor'у — implement Memory.

Verify — flip, match, mismatch shake, modal, skin, submit.

Скинь — скрин idle + mid-flip + victory.

Push после OK.

Hexagon — последний.

Wave 2 — скоро закрыт. Одна игра осталась.