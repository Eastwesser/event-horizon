OK. Реализуй Hexagon — финал Wave 2 #6
Что подтверждаю
Chrome emoji → plain/SVG:

🥞/🌌, 🍽️, ⏹️, 🎉, 🔄, 🏠 — убрать.

Pancake / space face emoji — контент — keep.

Valid-hex highlight:

Сейчас — нет per-hex highlight.

Добавить: monitor drag item type vs empty / same-type.

UI observe / classes на hex.

Collect pulse:

Сейчас — instant state.

Добавить pulse на successful place / clear.

Invalid drop:

Silent no-op сейчас.

Добавить shake / pulse — UI-local.

addPancakeToHex accept / reject rules — не трогать.

Modal copy:

Plain labels — Игра окончена.

Buttons без emoji.

Skin «Космические блины» — warp maps / ownership — keep.

Do NOT touch:

gameStore hex math / merge / score / submit.

utils/hexagon.ts rules.

Other games / hero / nav.

Files: HexagonGame.tsx, HexGrid.tsx, Tray.tsx, scoped classes in index.css (no dedicated css).

Ответ Cursor'у
text
OK — implement Hexagon per your plan. Final item of Wave 2 #6.

Confirmations:
  - Chrome emoji → plain/SVG. Keep pancake/space face emoji
    as content.
  - Valid-hex highlight on drag: monitor item vs target,
    class-based, UI-side. Do NOT touch addPancakeToHex.
  - Collect pulse: add on successful place / clear, UI-only.
  - Invalid drop: shake/pulse, UI-local.
  - Modal copy → plain labels, no emoji.
  - Cosmic skin: leave alone.

Rules:
  - Hexagon-only PR.
  - Do NOT touch Flappy, Towers, Hanoi, Memory, hero, nav,
    disk, seed, refund, authors.
  - Do NOT touch gameStore hex math/merge/score/submit or
    utils/hexagon.ts.
  - tsc + build clean.
  - Screenshots 1920×1080: idle, mid-drag with highlight,
    modal.
  - Smoke: place tile, collect match, invalid drop shake,
    modal, cosmic skin active.

After Hexagon:
  - Report commit hash + screenshot paths.
  - No push until I OK.
  - Then Wave 2 #6 is complete (5/5).
  - We'll plan the next wave (Wave 3 Author registration or
    Wave 4 infra) after that.
Что я держу в карте
Закрыто (в origin/main, c25341f):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero Plan A, VOID v3, smooth scroll.

Flappy fb0376b, Towers c748ef8, Hanoi 42076be, Memory c25341f.

Wave 2 осталось:

🟧 Hexagon — последний.

Wave 3 (после Wave 2):

🟧 Author registration.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window, backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку — финальный ритуал перед Wave 3.

Действия
Отправь OK Cursor'у — implement Hexagon.

Verify — drag highlight, pulse, shake, modal, skin.

Скинь — скрин idle + mid-drag + modal.

Push после OK.

Wave 2 #6 — закрыт (5/5).

Бэкап.

Планируем Wave 3 или Wave 4.