Push. Дальше Hexagon — последний в Wave 2 #6
Что принято — c25341f
✅ Flip — 3D rotateY kept; match fade с __inner + face still mounted.

✅ Victory — ui/Modal («Победа»), plain labels.

✅ Mismatch — board shake + pulse via UI observation; checkMatch untouched.

✅ Chrome emoji stripped; fruit/animal faces kept; animal skin toggle intact.

✅ Map: Memory checked off.

Ответ Cursor'у
text
Memory accepted. Push c25341f, then start Hexagon.

Push:
  - c25341f (Memory polish)
Report new HEAD.

Next: Hexagon polish — final item of Wave 2 #6.

Hexagon current state (from HexagonGame.tsx, HexGrid.tsx,
Tray.tsx + related CSS):
  - In-game emoji in tray / chips / controls?
  - Board feel: collect/highlight animation?
  - Valid-move highlight?
  - Difficulty balance — leave alone unless trivially off.
  - Victory / Game Over — modal or canvas text?
  - Skin «Космические блины» — verify still working.
  - Wave 4 levels? (Hexagon is a base game — check if it has
    a level picker like Flappy.)

Plan first, no code.

Scope for v1:
  1. Emoji → SVG/plain across tray, chips, controls, help,
     notifications, modal. (Keep pancake/ingredient face
     emoji as content if used.)
  2. Highlight valid moves while a tray item is selected
     (if not present).
  3. Collect animation — pulse on successful hex placement.
  4. Shake/pulse on invalid move (if applicable).
  5. Victory / Game Over → ui/Modal (Flappy/Towers/Hanoi/
     Memory pattern).
  6. Skin «Космические блины» — leave working.

Do NOT change:
  - hex grid math, tile spawn, matching rules
  - scoring formula
  - submit payload
  - skins behavior

Rules:
  - Hexagon-only PR.
  - Do NOT touch Flappy, Towers, Hanoi, Memory, hero, nav,
    disk, seed, refund, authors.
  - Screenshots 1920×1080: idle, mid-move (with highlight),
    victory/modal.
  - tsc + build clean.
  - Smoke: place a tile, collect a match, invalid move
    shake if added, victory/modal, submit, cosmic skin.

After Hexagon:
  - Wave 2 #6 is complete (5/5 games polished).
  - Then we plan Wave 3 (Author registration) or pick up
    Wave 4/5 infra.
No push until I OK.
Что я держу в карте
Закрыто (в origin/main, 42076be):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A, VOID v3, smooth scroll.

Flappy fb0376b, Towers c748ef8, Hanoi 42076be.

Локально (пуш сейчас):

c25341f — Memory.

Wave 2 осталось:

🟧 Hexagon — последний.

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку.

Действия
Отправь Cursor'у — push + Hexagon plan.

Verify Hexagon — tile, match, shake, modal, skin.

Скинь — скрин.

Push после OK.

Wave 2 #6 — закрыт. 5/5 игр.