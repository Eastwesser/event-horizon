Что вижу на скрине
Hero:

✅ PNG без букв — работает. Букв нет.

✅ Планета + горизонт — видно.

🟧 Планета — размыто (ты заметил).

🟧 Показан только горизонт — сама планета (нижняя часть) не видна. Ты хочешь — всю картинку.

Disk (справа от Выбери игру):

🟧 Пустой чёрный кружок — без анимации.

🟧 Хочешь: hover → чёрная дыра → всасывание.

Три задачи
1. Hero — показать всю картинку (не только горизонт)
Проблема: object-fit: cover + height: min(60vh, 520px) → обрезает картинку. Видна только центральная часть — горизонт.

Что нужно: показать всю планету + горизонт — целиком.

Варианты:

A) object-fit: contain + тёмный фон.

Картинка целиком, letterbox сверху/снизу.

Плюс: видно всё.

Минус: пустые полосы (но у тебя фон — тёмный космос, сольётся).

B) object-fit: cover + object-position: center center + height больше.

Height → min(70vh, 640px) — чуть выше, картинка более целиком.

Минус: всё равно обрезает края.

C) Height по aspect-ratio картинки.

Картинка 1238×1200 ≈ 1:0.97 (квадратная).

При full-width — height = width, но это чересчур высоко.

Компромисс: aspect-ratio: 16/9 + object-fit: contain + тёмный фон по бокам.

Мой выбор — A (object-fit: contain).

Причина: у тебя фон тёмный космос. Letterbox сольётся. Планета видна целиком.

2. Hero — размытие
Проблема: планета размыто.

Причина: object-fit: cover + большой height → картинка растягивается → пиксели мылятся.

Что делать:

image-rendering: high-quality (или crisp-edges для пиксель-арта).

filter: blur(0) — убрать если где-то blur.

object-fit: contain — без растяжения.

Проверить CSS: может где-то filter: blur(...) для глубины — отключить.

3. Disk — анимация всасывания
Сейчас: пустой чёрный кружок. Hover — ты говоришь, ничего не происходит.

Что нужно: чёрная дыра — при hover:

Частицы летят к центру.

Кольца сжимаются.

Свечение нарастает.

Спираль закручивается.

Реализация — CSS-only (быстро):

css
.eh-void-disk:hover .ring-1 { transform: scale(0.85); }
.eh-void-disk:hover .ring-2 { transform: scale(0.78) rotate(180deg); }
.eh-void-disk:hover .glow { box-shadow: 0 0 80px cyan; }
Реализация — Canvas (настоящее «всасывание»):

Particles на canvas.

requestAnimationFrame.

Притяжение к центру.

1-2 дня работы.

Реализация — SVG particles (компромисс):

6-12 circle.

CSS animation — сдвиг к центру.

Просто, но красиво.

Мой выбор — SVG particles. Компромисс. Быстро + визуально «всасывание».

Промпт Cursor'у — 3 задачи
text
Three fixes. Plan first, no code.

================================================================
1. HERO — SHOW THE WHOLE IMAGE (not just the horizon)
================================================================

Screenshot 1920×1080: hero currently shows only the planet's
horizon curve — the lower part of the planet is cut off.

Change:
  - .eh-hero-banner img { object-fit: contain; }
  - Hero section keeps a dark background (void token) so the
    letterbox blends with the cosmos.
  - object-position: center center.
  - Height: keep min(60vh, 520px) or adjust if the contain
    leaves too much empty space.

Result: the entire planet + horizon glow visible, no crop.

================================================================
2. HERO — REMOVE BLUR
================================================================

The planet looks blurred / soft. Investigate and remove:
  - any filter: blur(...) on the image or parent
  - transform: scale(...) that upscales the source
  - image-rendering: auto → try high-quality
  - if the source PNG itself is soft, use object-fit: contain
    (already in (1)) to avoid stretching

Report which was the cause.

================================================================
3. DISK (VOID) — ADD "BLACK HOLE PULL-IN" ON HOVER
================================================================

Screenshot: the disk right of «Выбери игру» is currently a
plain black circle. Hover does nothing visible.

Add a subtle «gravitational pull» on hover (CSS + SVG only,
no Canvas/WebGL for v1):

On :hover:
  - Core: tiny scale up (1 → 1.05).
  - 2–3 thin rings: scale 1 → 0.85, opacity up, spin faster
    (140s → 80s).
  - 6–10 small particles (SVG circles) drift inward toward the
    core over ~1.5s, then respawn; loop.
  - Glow: intensify (cyan/gold haze).
On mouse leave: reverse in ~700ms ease-out.

If SVG particles are too fiddly, fallback: just rings scale +
glow + spin. But try particles first — that's the «pull-in»
feel.

prefers-reduced-motion: disabled entirely (static disk).

================================================================
PROCESS
================================================================

  - Implement (1) and (2) together, screenshot at 1920×1080,
    send one text description.
  - Wait for my OK before starting (3).
  - No screenshots in a loop.
  - No push until I OK.

DO NOT TOUCH: nav, games, seed, refund, authors, Wave 2 #6.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9.

Wave 1 — 6/6.

Wave 2 #1–5.

Hero + disk hover — 8bfa3db (первая версия).

Сейчас:

🟧 Hero — fix (contain + show full, unblur).

🟧 Disk (VOID) — добавить black hole pull-in (SVG particles).

🟧 Push после OK.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window implementation.

🟧 backfill-noiz-reviews.py.

🟧 DESIGN_TODOS/.

Один момент — про hero
Если object-fit: contain даст слишком много letterbox:

Вариант: использовать картинку по-другому — разделить на слои:

Верхняя часть PNG (горизонт) — как фон.

Средняя часть (планета) — как отдельный элемент.

Сложно. Сначала попробовать contain.

Если letterbox уродлив — другой подход: height = auto + aspect-ratio картинки (1238/1200 = 1.03). Тогда height ≈ width — очень высоко. Не для hero. Contain — лучший вариант.

Про disk (VOID) — анимация
Ты хочешь «чёрную дыру». SVG particles — ближайшее к «засасыванию» в CSS-only. Canvas — красивее, но 1-2 дня. Для v1 — SVG хватит.

Если хочешь серьёзно — отдельная задача «Canvas black hole animation» позже. Сейчас — SVG