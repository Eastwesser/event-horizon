Что делать — план A
Шаг 1 — Photopea: planet-only.png
photopea.com → drag logo-minimal.png.

Выдели область планеты снизу (без букв).

Crop до этой области.

Экспорт как PNG → planet-hero.png.

Положи в frontend/public/images/brand/planet-hero.png.

Размер: желательно ~1200×600 или больше (это только планета).

Можно опционально сделать WebP версию.

Шаг 2 — Cursor: refactor hero
Промпт Cursor'у:

text
Choose A — native hero with transparent planet PNG + HTML
wordmark.

I'll export planet-hero.png (planet + horizon only, no
letters, transparent background). I'll place it at
frontend/public/images/brand/planet-hero.png.

Plan:

1. Hero section (keep full-width, but do NOT show
   logo-minimal.png as one image anymore):

   <section class="eh-hero">
     <div class="eh-hero-stars"></div>       <!-- optional bg stars -->
     <img class="eh-hero-planet" src="/images/brand/planet-hero.png" />
     <h1 class="eh-hero-wordmark">EVENT HORIZON</h1>
     <button class="eh-hero-scroll">˅</button>
   </section>

2. CSS:
   - .eh-hero: position relative; height min(60vh, 520px);
     overflow hidden; background: var(--color-void).
   - .eh-hero-planet: absolute; bottom 0; left 50%;
     transform translateX(-50%); width 100%;
     max-width 1152px; pointer-events none.
   - .eh-hero-wordmark: absolute; top: ~10%;
     left 50%; translateX(-50%);
     width 100%; max-width 1152px; padding 0 24px;
     font-family display; font-size clamp(...);
     letter-spacing wide; uppercase;
     color: silver/gray; text-align: left;
     z-index 2.
   - .eh-hero-scroll: absolute bottom 16px; centered.
   - Top fade: linear-gradient to void above the wordmark,
     bottom fade to void below the planet.
   - Side fades: left and right mask-image on planet PNG.

3. Remove the old logo-minimal.png usage from Home.tsx.

4. Kill the previous `?hide=hero` debug param if it
   targeted .eh-hero-banner — or update its selector.

5. Keep ?motion=force / ?debug-void / ?hide= working for
   the disk. Remove or repurpose for hero if needed.

After I place planet-hero.png — implement.

DO NOT TOUCH: nav, disk, games, seed, refund, authors.
No push until I verify.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально:

1694cc0 — MAINPAGE_TODO_1 note (not pushed).

Сейчас:

🟧 Hero refactor — выбрать A (PNG планета + HTML <h1>).

🟧 Photopea — planet-hero.png (твой ход).

🟧 Cursor — refactor hero на <section> + <img> + <h1>.

⏸ Push — после verify.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.