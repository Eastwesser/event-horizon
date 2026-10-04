Push. Дальше Hanoi
Что принято
Flappy — fb0376b (local) — verify ok.
Towers — c748ef8 (local) — verify ok.

towerDraw.ts — bevelled blocks, glow, chevrons.

Game Over → Modal.

Shake + drop pulse.

Rainbow skin → SVG.

tsc + build — clean.

«Если будет желание — доработаю» — правильно. Wave 2 #6 — базовый polish, не идеал.

Промпт Cursor'у — push + Hanoi
text
Towers and Flappy accepted — commit as-is, do not keep
iterating. Push both, then start Hanoi.

Push:
  - fb0376b (Flappy polish)
  - c748ef8 (Towers polish)

Report hashes after push + new HEAD.

Next: Hanoi polish (one PR, same rules).

Hanoi current state (from HanoiTower.tsx + related CSS):
  - 3 pegs, disks sized by level.
  - Drag: previously had the stretch regression; verify
    current behavior first.
  - GAME OVER / Victory: check if there's already a modal
    (there was a Victу modal earlier — verify it still
    matches the new ui/Modal primitive).
  - In-game emoji to replace? (check the file)

Plan first, no code.

Scope for v1:
  1. Drag polish — disk must not stretch during drag;
     snap to peg; highlight target peg on hover.
  2. Victory modal — align to ui/Modal (reuse Flappy/Towers
     pattern if already there).
  3. Shake / pulse on invalid move.
  4. In-game emoji → SVG if any.
  5. Do NOT change game logic — move validation, minimal
     move count, level data — leave the store alone.

Rules:
  - One game per PR. Hanoi only.
  - Keep skins / colors for disks intact.
  - After Hanoi: Memory, then Hexagon.

DO NOT TOUCH: hero, nav, disk, seed, refund, authors,
Flappy, Towers, other games.
Что я держу в карте
Закрыто (в origin/main):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A.

VOID v3.

Smooth scroll.

Локально (пуш сразу):

fb0376b — Flappy.

c748ef8 — Towers.

Wave 2 осталось:

🟧 Hanoi (следующее).

🟧 Memory.

🟧 Hexagon.

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.