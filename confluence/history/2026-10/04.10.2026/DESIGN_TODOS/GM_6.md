Что вижу
Скрин 1 — реальный (1920×1080)
Hero — квадратная картинка (contain) в центре.

Квадрат виден — тёмный фон вокруг, граница между PNG-фоном и page-void — заметна.

EVENT HORIZON — внутри квадрата, ширина меньше, чем Выбери игру под ним.

Disk (VOID) — пустой чёрный кружок + rings + частицы. Статичен, анимации нет.

Page void — совпадает с PNG-void? Вроде да, но край PNG заметен — лёгкая граница.

Скрин 2 — твой Paint-монтаж
Что ты хочешь:

EVENT HORIZON (в hero) — визуально совпадает с Выбери игру и ставь рекорды по вертикальным границам (левый и правый край).

Картинка — full-width по границам текста, не квадрат по центру.

Бленд по краям — PNG-фон сливается с page-void, квадрата не видно.

Три задачи Cursor'у
1. Hero — full-width по границам текста
Сейчас: contain → квадрат в центре.

Что нужно: растянуть картинку по горизонтали так, чтобы её левый и правый край = границы текста Выбери игру (как в Paint-скрине).

Как:

Container — тот же max-width, что у shellInner секции Выбери игру.

Картинка — width: 100%, height: auto, object-fit: contain.

Hero — не full-bleed, а по ширине контейнера.

Высота — по aspect-ratio картинки.

Проверить: левая грань картинки ≈ левая грань текста Выбери. Правая — правая грань Лидерборд кнопки.

Оговорка: PNG квадратная (1238×1200), при width=100% height ≈ width → очень высоко для hero. Вариант:

Обрезать сверху/снизу — overflow: hidden + height = min(60vh, 520px).

Или — использовать только верхнюю часть PNG (буквы + горизонт).

Но — тогда буквы уходят не по границе текста точно.

Хм. Тут дилемма:

Full-width → высоко → не hero.

Fixed height → crop → буквы не совпадут с границей текста.

Решение: PNG подогнать — padding по бокам или отдельный hero-PNG широкий (16:9). Или — HTML <h1> + PNG-фон.

Мой выбор — HTML <h1>:

PNG-фон (без букв).

HTML текст EVENT HORIZON — полностью под контролем:

тот же max-width, что Выбери игру.

тот же контейнер → границы совпадут.

отступ от navbar — задаёшь.

Но — PNG без букв? Ты сказал — нет. Тогда…

Тогда — container-width для PNG (не full-bleed):

Container — max-w-6xl mx-auto (как shell).

Картинка — width: 100%.

Hero — высокий, обрезать снизу (не тронуть буквы + горизонт).

Тогда буквы точно по границам текста.

Промпт:

text
Hero image:
  - Change from full-bleed centered square to container-width.
  - Same container as the shellInner section with «Выбери игру»
    (max-w-6xl mx-auto px-4 sm:px-6, or whatever that shell uses).
  - Image inside: width 100%, height auto, object-fit contain.
  - Hero section: overflow hidden + fixed height
    min(60vh, 520px) so it doesn't grow too tall.
  - Align image top so «EVENT HORIZON» wordmark starts at the
    same left/right edges as «Выбери игру и ставь рекорды».
Result: буквы совпадают по границам с текстом снизу.

2. Бленд по краям PNG
Сейчас: виден квадрат — граница между PNG и page-void.

Что делать:

CSS mask по краям — fade к прозрачности.

Или — поверх PNG наложить gradient — слева и справа (к page-void).

Или — filter: blur по краям? Нет, blur плохо.

Или — PNG имеет свой фон, не совпадающий с page-void по пикселю. Найти точный void color и подогнать PNG-фон.

Лучший подход: CSS mask — mask-image: linear-gradient(...).

css
.eh-hero-banner img {
  mask-image: linear-gradient(
    to right,
    transparent 0%,
    black 8%,
    black 92%,
    transparent 100%
  );
}
Плюс — сверху/снизу то же.

Промпт:

text
Hero image: blend edges with page background.
  - Use CSS mask-image (linear-gradient) on the left and
    right edges — fade to transparent over ~8%.
  - Optional: top and bottom too if edges are visible.
  - Goal: no visible square border between PNG and page void.
3. Disk VOID — анимация
Cursor пишет:

Disk not touched. Waiting for OK before (3).

Т.е. — он ждёт, не добавлял.

Промпт:

text
Now (3) — disk pull-in on hover.

SVG particles + rings shrink/spin/glow on hover.
No logo inside.
prefers-reduced-motion: disabled.

On :hover:
  - Core: scale 1 → 1.05.
  - Rings: scale 1 → 0.85, opacity up, spin 140s → 80s.
  - 6-10 SVG particle circles drift inward over 1.5s, loop.
  - Glow intensifies (cyan/gold).
On leave: reverse 700ms ease-out.
Что ответить Cursor'у
text
Screenshot reviewed. Three changes:

1. HERO — CONTAINER WIDTH, NOT FULL-BLEED
   Currently the hero image is a centered square. I want it
   to span the same width as the «Выбери игру и ставь
   рекорды» text block below.
   - Use the same container (max-w-6xl mx-auto px-4 sm:px-6
     or whatever the shellInner uses).
   - Image inside: width 100%, height auto.
   - Hero: overflow hidden + height min(60vh, 520px).
   - Result: «EVENT HORIZON» wordmark starts at the same left
     and right edges as «Выбери игру и ставь рекорды».

2. HERO — BLEND EDGES
   The PNG's dark background creates a visible square against
   the page void.
   - Add CSS mask-image with linear-gradient on left and right
     (fade to transparent ~8%).
   - Optional top/bottom mask if needed.
   - Result: no square border.

3. DISK (VOID) — PULL-IN ON HOVER
   Go ahead with (3):
   - SVG particles + rings shrink/spin/glow on hover.
   - No logo inside.
   - prefers-reduced-motion disabled.

After 1+2: one screenshot, text description, wait for my OK.
Then 3.
No loop. No push.

DO NOT TOUCH: nav, games, seed, refund, authors, Wave 2 #6.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1 (6/6), Wave 2 #1–5

Hero — 8bfa3db (первая версия, в работе)

Сейчас:

🟧 Hero — container width (не full-bleed квадрат)

🟧 Hero — blend edges (mask)

🟧 Disk VOID — pull-in (SVG particles)

⏸ Push после OK

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

🟧 Wave 2 #5b, Refund window, backfill-noiz-reviews.py, DESIGN_TODOS/.

Один момент — про EVENT HORIZON в PNG
Ты хочешь: буквы совпадают по границам с Выбери игру.

Но — PNG-буквы внутри картинки, не управляются CSS.

Если PNG-картинка шире, чем буквы внутри → при width=100% буквы будут уже, чем контейнер.

Решение:

Или — обрезка PNG до границ букв → тогда PNG-буквы = границы контейнера.

Или — HTML <h1> + PNG без букв.

Или — подгонка padding в CSS.

Скажи Cursor'у:

text
If the PNG's baked-in «EVENT HORIZON» has padding inside the
image, the wordmark won't align to the container edges even
at width: 100%. Options:
  a) Crop the PNG so the wordmark is flush with the image
     edges.
  b) Use an HTML <h1> for the wordmark + text-free PNG
     background.
Which works — report before implementing.