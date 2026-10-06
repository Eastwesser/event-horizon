Build упал на 2 ошибках TypeScript. Нужно починить до verify
Что вижу
text
src/components/Games/Hexagon/HexGrid.tsx:158:15
  TS6133: 'item' is declared but its value is never read.
  hover: (item: DragPancake, monitor) => {

src/components/Games/Memonia/MemoryGame.tsx:46:5
  TS6133: 'resetGame' is declared but its value is never read.
Обе — TS6133 (unused variable). Не error логики, просто не используются:

HexGrid.tsx:158 — в hover: (item, monitor) => — item не используется.

MemoryGame.tsx:46 — resetGame импортирован/деструктурирован, но не вызывается.

Build упал → make fe-preview не работает → нельзя сделать smoke boost.

Причина: cursor'овский Phase 1 внёс эти переменные (или оставил старые), но не использует их.

Что делать — 2 варианта
Вариант A — попросить Cursor'а починить (правильно)
Он внёс это — пусть уберёт. Fix trivial:

HexGrid.tsx: либо убрать item из аргументов, либо использовать (если должен был использовать).

MemoryGame.tsx: либо убрать resetGame из деструктуризации, либо вызвать его где-то.

⚠️ Возможен скрытый баг: если Cursor должен был использовать item в hover или resetGame при boost effect — это не просто unused, это недоделанная фича. Спроси явно.

Вариант B — быстрый workaround (сам, если просто unused)
Убрать переменные:

tsx
// HexGrid.tsx:158
hover: (_item: DragPancake, monitor) => {

// MemoryGame.tsx:46
// удалить resetGame из деструктуризации, если он реально не нужен
Минус: если фича должна использовать — потеряем её.

Мой выбор — A, но с уточнением
Спроси Cursor'а:

HexGrid hover — item должен использоваться? (Memory flash? boost effect?).

MemoryGame resetGame — должен вызываться при boost? Или restart?

Если должны — это недоделка, доработать.
Если просто unused — убрать и пересобрать.

Промпт Cursor'у
text
Build broke on `make fe-preview`:

  HexGrid.tsx:158 TS6133: 'item' declared but never read
    hover: (item: DragPancake, monitor) => {

  MemoryGame.tsx:46 TS6133: 'resetGame' declared but never read

Before verify — two questions, then fix:

1. HexGrid hover(item, monitor):
   - Was `item` supposed to be used for boost (brighten
     hover, valid-hex highlight)?
   - Or is it genuinely unused after Phase 1?

2. MemoryGame resetGame:
   - Should it be called on boost start (reset unmatched
     pair / restart)?
   - Or is it leftover from a previous refactor?

If both were supposed to be USED — implement the usage:
  - HexGrid hover: use `item` for the boost-aware
    highlight (or remove if not needed).
  - MemoryGame: call resetGame() where boost is armed
    (or remove if unused).

If both genuinely unused — just remove:
  - HexGrid.tsx: rename `item` → `_item`.
  - MemoryGame.tsx: drop `resetGame` from destructure.

Then:
  - tsc --noEmit clean.
  - npm run build clean.
  - Report what you fixed (unused vs missing usage).

No commit yet. No push.

Standing rule: you don't push; you write commit messages.
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

Локально (не запушено):

ee75936 — Phase 1 boost — build broken.

Track B parity:

🟧 Phase 1 — boost — build fix перед verify.

🟧 Phase 2 — levels game-by-game.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.