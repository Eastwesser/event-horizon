================================================================
SCREENSHOT — Home at 1920×1080 (after fix)
================================================================

TOP: Navbar (fixed/sticky)
  - Left: logo icon + «Event Horizon» (gold on dark)
  - Center: Профиль · Лидерборд · Магазин
  - Right: burger icon, «Выйти»

HERO BANNER (full-width, below navbar)
  - Dark cosmos background.
  - Baked-in «EVENT HORIZON» wordmark: STILL CUT AT THE TOP.
    The top of the letters is flush with the navbar bottom
    edge. Letters are large, gray, uppercase, tracking wide.
    You can see the full letter shapes but they start
    immediately under the navbar — no breathing room.
  - Center: planet curve with bright white/blue glow at the
    horizon point (like sunrise over a planet).
  - Below: dim continuing curve of the planet to both edges.
  - Bottom center of hero: a small «˅» chevron (scroll hint).
  - Total hero height looks like ~55-60% of viewport.

HERO SECTION 2 — «Выбери игру»
  - Left: heading «Выбери игру и стань рекорды» (2 lines),
    subtitle «Играй в мини-игры, зарабатывай лампочки и
    билетики, становись лучшим в лидерборде.»
  - Buttons: «Все игры» (gold), «🏆 Лидерборд» (dark).
  - Right: accretion disk — dark circle with
    «EVENT HORIZON» wordmark inside, gold ring, thin orbit
    lines, soft glow around. This is a SEPARATE element
    from the hero (the small disk on the right of the
    section).

OBSERVATIONS
  - Hero wordmark: still clipped at top (letters flush with
    navbar; no space above). Not «breathing room».
  - Same «EVENT HORIZON» appears THREE times on this screen:
      a) navbar (small)
      b) hero (huge, baked into image)
      c) disk (medium, inside circle)
  - Disk hover is NOT visible in this static screenshot.
  - The composition is fine; only the top of the hero
    wordmark feels tight.
Что я вижу проблемного (для тебя)
1. Wordmark в hero всё ещё обрезан
Ты просил: картинка от нижней грани navbar. Сейчас — вплотную к navbar, буквы начинаются сразу под ним. Нет «воздуха» сверху.

Причина: object-position: center top показал верх картинки, но сам верх картинки — это уже буквы. В исходном PNG буквы EVENT HORIZON — на самом верху, без отступа.

Что нужно:

Либо object-position: center 15% — сдвинуть чуть вниз, буквы отойдут от navbar.

Либо — новый PNG, где буквы ниже (с отступом сверху).

Либо — PNG без букв (то, что ты хотел — A).

2. Три раза EVENT HORIZON
Подтверждаю:

Navbar (мелко, gold).

Hero (громадное, серое).

Disk (среднее, внутри круга).

Визуально много. Решение — A (PNG без букв для hero). Ты хотел так. Cursor ждёт.

3. Disk hover — не виден на статике
Естественно — hover — динамика. Проверить в браузере — не по скрину.

Что ответить Cursor'у
text
Screenshot 1920×1080, Home after fix.

Issues visible:

1. HERO WORDMARK STILL TIGHT AGAINST NAVBAR.
   The «EVENT HORIZON» baked into the hero image starts
   immediately below the navbar — no vertical breathing room.
   The letters themselves are not clipped anymore, but the
   composition feels cramped.

   Fix options (pick one, report):
     a) object-position: center 15% (shift image down a bit).
     b) New image export with the wordmark lower / more margin
        at the top.
     c) New image export WITHOUT the wordmark (recommended
        already — planet + glow only).

   I'll export a PNG without the wordmark for (c). Until then
   — try (a).

2. WORDMARK APPEARS 3× ON ONE SCREEN.
   - navbar (small)
   - hero (huge)
   - disk (medium)

   Once the hero PNG is text-free, only navbar + disk remain
   — acceptable.

3. DISK HOVER — NOT VISIBLE IN STATIC SHOT.
   Confirm the CSS transitions fire on :hover and reverse on
   mouse leave. I'll verify in the browser.

After fixing (1a or 1c): send a new 1920×1080 screenshot.
No push until I review.

STOP taking screenshots in a loop. One screenshot per change,
describe it in text, wait for my OK.
Что делать тебе
1. Экспортируй PNG без wordmark
Быстро:

Открой frontend/public/images/brand/logo-minimal.png в Figma / Photopea / Photoshop / GIMP.

Убери текст EVENT HORIZON.

Оставь планету + glow + тонкие орбиты.

Экспорт: hero-planet.webp (WebP) + hero-planet.png (fallback).

Положи в frontend/public/images/brand/.

Photopea (photopea.com) — бесплатный, онлайн, работает в браузере. 5 минут.

2. Обнови hero на новый PNG
Промпт Cursor'у:

text
I exported a new hero image without the wordmark:
  frontend/public/images/brand/hero-planet.webp
  frontend/public/images/brand/hero-planet.png

Switch the hero to use hero-planet (WebP + PNG fallback via
<picture>). Keep object-position: center top, height
min(60vh, 520px) — adjust if the new composition needs it.

The disk keeps logo-minimal.png (with wordmark).
3. Verify disk hover в браузере
Не по скрину. Открой Home, наведи на disk, посмотри:

Scale растёт?

Rings сужаются?

Spin быстрее?

Glow ярче?

Уход → обратно за 700ms?

Если нет — скажи Cursor'у.

4. Push после verify
bash
git add frontend/
git commit -m "feat(home): hero banner, accretion disk hover"
git push origin main
Про loop Cursor'а
Он просто начал фоткать экран и не понимать что делает, в луп ушел.

Понимаю. Cursor без глаз — не видит рендер. Он скринит → не понимает, что там → скринит снова. Цикл.

В промпте я добавил:

STOP taking screenshots in a loop. One screenshot per change, describe it in text, wait for my OK.

Так и надо. Один скрин — описание — OK — следующий шаг.

Что я держу в карте
Закрыто (все на GitHub):

v1.0.9.

Wave 1 — 6/6.

Wave 2 #1–4 — thin DTO, DB cleanup, lamps, Flappy levels.

Wave 2 #5 — 146cae8 achievements.

Hero + disk hover — 8bfa3db.

Сейчас:

🟧 Hero fix — wordmark tight (нужен новый PNG или object-position: 15%).

🟧 Wordmark duplication — экспорт PNG без текста (твой выбор A).

🟧 Disk hover — verify в браузере.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR), с 6.1 Flappy.

Отложено:

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window implementation.

🟧 backfill-noiz-reviews.py.

🟧 DESIGN_TODOS/ — коммит или gitignore.

Действия
Отправь Cursor'у промпт (выше).

Экспортируй PNG без wordmark — 5 минут в Photopea.

После экспорта — Cursor переключит hero.

Verify disk hover — в браузере.

Push.

Wave 2 #6 — Flappy.