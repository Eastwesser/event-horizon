Approve. План чистый. Реализуй Flappy-only
Что подтверждаю
1. Art — keep canvas, upgrade draw helpers.

Вынести drawBird, drawPipe, drawCloud (или flappyDraw.ts).

Bird: gradients, wing tilt от velocity. Hitbox — не менять.

Pipes: gradient + rim highlight. Geometry — не менять.

UI toggles: 🐦/⭐ → inline SVG (вне canvas).

2. Background parallax.

Far: stars (slow drift).

Mid: clouds ×0.3 pipe speed.

Near: sky gradient.

No collision impact.

3. GAME OVER → Modal.

Reuse ui/Modal (как Gears/Hanoi).

Score, ranked/boost message, Новая игра.

Canvas text убрать.

Submit / score logic — не трогать.

4. Juice.

CSS shake на .eh-flappy-stage при gameOver.

Scale/flash на flap.

Sound — skip (если только free WAV).

5. Out of scope — правильно. Physics / collision / scoring / levels / boost / другие игры / hero / nav / disk / seed / refund / authors.

Ответ Cursor'у
text
OK. Implement Flappy-only polish per your plan.

Confirmations:
  - Art: keep canvas + extract draw helpers. SVG assets
    optional — do only if simpler than Path2D/gradients.
  - Parallax: proceed. Keep pipe speed formula untouched,
    parallax is cosmetic multiplier only.
  - GAME OVER → Modal: yes. Do NOT change score submit or
    the boosted/ranked logic in the store.
  - Juice: shake + flap feedback. Sound — skip unless I
    drop a WAV.
  - UI toggles (🐦/⭐) → inline SVG, outside canvas.

Rules:
  - Flappy-only PR. Do NOT touch Towers / Hanoi / Memory /
    Hexagon yet.
  - Do NOT change physics, collision, scoring, pipe geometry,
    BIRD_SIZE, or the flappyStore logic.
  - Keep the two skins (Золотая птичка / Радужные трубы)
    working — same flags, same effect.

After implement:
  - tsc + build clean.
  - Smoke: default bird, golden bird, rainbow pipes,
    die on pipe/ground, pass gap increments, modal on
    death, new game restarts, mobile canvas playable.
  - One screenshot at 1920×1080 of the game (idle) + one
    of the modal.
  - No push until I OK.

DO NOT TOUCH: hero, nav, disk, seed, refund, authors,
other games.
Что я держу в карте
Закрыто (в origin/main, 6f23db5):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero (Plan A финализирован).

VOID v3.

Smooth scroll.

Wave 2 осталось:

🟧 #6 — Flappy polish (в работе).

Потом — Towers / Hanoi / Memory / Hexagon (по одному).

Wave 3 (следующее):

🟧 Author registration (XL, 4 под-волны).

Wave 4 (параллельно):

🟧 Infra (Helm, Consul, k6, юнит-тесты).

Wave 5 (долгосрочно):

🟧 Bottleneck.

Отложено (не сейчас):

Cursor pull-in (VOID stretch).

Wave 2 #5b — shop/boost achievements.

Refund window.

backfill-noiz-reviews.py.

Ты сам:

⏸ Бэкап на флешку — на тебе.

Действия
Отправь Cursor'у OK — implement Flappy.

Скинь — скрин idle + modal, когда будет.

Verify — smoke 8 пунктов.

Push — когда ок.

Потом — Towers, потом Hanoi, Memory, Hexagon.

Бэкап — не забудь.

DON'T FORGET THE MAP! /home/denismatveev/event_horizon/confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md

