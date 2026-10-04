OK. Реализуй
Что подтверждаю
Files:

Home.tsx — 5 rings + SVG ~30 particles + near-hover/escalate classes.

theme.css — idle / hover / escalate / near-pull keyframes + reduced-motion.

SVG structure — чистая:

text
.eh-disk
  .eh-disk-glow
  .eh-disk-pull
    .eh-disk-rings → 5× ring (ellipse)
  .eh-disk-core (без лого)
  svg.eh-disk-particles → 30× circle
Behavior:

5 rings — 160s → 40s, gold / cyan / violet, hover scale 1 → 0.75.

30 particles — 1/2/3px, spiral, fade at center, respawn.

Glow — idle soft, hover stronger.

Near (~80px) — eh-disk--near — slight brighten + drift.

Escalate (>2s) — eh-disk--hot — ×1.5 speed, glow up.

Perf:

transform/opacity only.

prefers-reduced-motion — static.

Всё правильно.

Мой ответ Cursor'у
text
OK — implement VOID v2 (Level 1, SVG only).

Three notes:

1. NEAR + ESCALATE — keep them.
   Both are the "alive" feel. Implement as class toggles:
     - eh-disk--near  (on pointer within ~80px)
     - eh-disk--hot   (after 2s hover, cap at 3s)
   Use wrapper onPointerMove / onPointerEnter / onPointerLeave.

2. REDUCED-MOTION — hard off.
   Static disk, no spin, no particles, no pull.
   Verify in DevTools: Rendering → Emulate prefers-reduced-motion.

3. PERFORMANCE WATCH.
   If the virtual machine lags, check:
     - transforms vs layout thrash (should be transform/opacity only)
     - will-change: transform on .eh-disk-pull during hover
     - devicePixelRatio — maybe cap at 2
   Report if you see fps drops below 55.

After implement:
  - One 1920×1080 screenshot with hover ACTIVE (not idle).
  - Short text: what changes on hover, what on near, what on hot.
  - No loop. No push until I OK.

DO NOT TOUCH: hero, nav, games, seed, refund, authors,
Wave 2 #6.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Сейчас:

✅ Hero — работает.

✅ Smooth scroll — работает.

🟧 Disk v2 (SVG 30 particles) — plan approved → implement.

⏸ Wordmark flush — ждёт crop PNG (Photopea).

⏸ Push после OK.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

Wave 2 #5b, Refund window, backfill-noiz-reviews.py, DESIGN_TODOS/.

Параллельно — Photopea crop PNG
Пока Cursor кодит — 5 минут:

photopea.com

Drag logo-minimal.png.

Image → Crop (или C).

Обрежь слева/справа по границам букв EVENT HORIZON. Сверху/снизу — оставь padding (планета + glow).

File → Export As → PNG → replace frontend/public/images/brand/logo-minimal.png.

Сохрани в git.

Проверь: буквы flush к краям PNG (без padding по бокам).

Потом Cursor'у:

text
Replaced logo-minimal.png with a cropped version — wordmark
is now flush with image edges. Verify hero wordmark aligns
with «Выбери игру» text block edges.
Про производительность — заранее
Если виртуалка лагает:

Проверь в DevTools → Rendering:

FPS meter — включить.

Paint flashing — включить.

Проверь в chrome://gpu:

Hardware acceleration — Enabled?

Если лагает даже idle — проблема в виртуалке (GPU passthrough). Не в коде. На реальной машине — должно быть 60fps.

Скажи Cursor'у:

text
If FPS drops: cap devicePixelRatio at 2, use
will-change: transform on .eh-disk-pull during hover only,
and confirm no box-shadow changes trigger layout.