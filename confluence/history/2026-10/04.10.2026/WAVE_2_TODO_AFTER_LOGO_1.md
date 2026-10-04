3. Wave 2 #6 — Flappy polish
Промпт Cursor'у (последний пункт Wave 2):

text
Hero accepted — commit as-is. Push done.

Wave 2 #6 — final item: game polish (one game per PR,
sequential). Start with Flappy.

Plan first, no code.

Flappy polish scope (v1):
  - In-game emoji → SVG (bird, pipes, clouds).
  - Bird sprite: check if the current bird is emoji or a
    canvas drawing; keep the shape, replace with cleaner
    art or SVG.
  - Pipes: subtle gradient / texture instead of flat color.
  - Background: parallax layers (clouds, stars)?
  - GAME OVER: currently canvas text — move to a proper
    modal (reuse Modal primitive from inventory/shop).
  - Shake on miss + click feedback.
  - Sound: optional, only if a free asset is available.

Constraints:
  - One game at a time. Flappy first, then wait for my OK
    before Towers, Hanoi, Memory, Hexagon.
  - Do NOT change game physics, collision, or scoring.
  - Keep the existing Skins (Золотая птичка, Радужные трубы)
    working.

DO NOT TOUCH: hero, nav, seed, refund, authors, disk.
Show the plan first. No code yet.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально (не запушено):

1694cc0 — MAINPAGE_TODO_1.

dc6869e — Hero Plan A.

6a1e574 — Hero (mask + wordmark).

7674c51 — Hero (центр).

a821ac7 — Hero (планета под надписью).

Сейчас:

⏸ Push — 5 локальных коммитов.

⏸ Бэкап.

🟧 Wave 2 #6 — Flappy polish.

Отложено (не сейчас):

Cursor pull-in (VOID stretch).

Wave 2 #5b — shop/boost achievements.

Refund window.

backfill-noiz-reviews.py.

Дальше (по плану):

Wave 3 — Author registration.

Wave 4 — Infra (Helm, Consul, k6).

Wave 5 — Bottleneck.

Действия
git push origin main — 5 коммитов.

Проверь — GitHub обновлён.

Бэкап на флешку.

Отправь Cursor'у промпт — Flappy polish.

Потом — отдых. Hero закрыт.