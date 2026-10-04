Что вижу
Квадрат — едва различимый, но есть. Правый край виден — граница от ~80% ширины. Верх — на уровне ~5%. Левый край — менее заметен (сливается), но есть.

Цвет границы: чуть темнее / другого оттенка, чем фон вне квадрата. Право и верх — заметнее.

Причина
Это точно тот же .eh-disk-glow — после фикса радиального градиента, но:

Градиент не полностью прозрачен на краю (или box-shadow остался).

Или — blur всё ещё обрезается по контейнеру.

Или — .eh-disk-pull имеет rotateX(52deg) → плоскость наклонена → её граница видна как прямоугольник в 3D.

Или — там осталась backdrop-filter / маска с жёсткой границей.

Правая грань выражена сильнее — вероятно это граница .eh-disk контейнера или .eh-disk-pull после rotateX.

Что делать
Промпт Cursor'у:

text
Glow is much better but a faint rectangular boundary is
still visible around the disk — right and top edges most
noticeable. Screenshot attached conceptually:

  - The border color is slightly darker/different from the
    surrounding page background.
  - Right edge visible around ~80% of the disk width.
  - Top edge visible near the top of the glow.

Possible causes to check:

1. .eh-disk-glow radial-gradient does not reach fully
   transparent at the edge — the last stop is not 0 alpha,
   so the container edge shows.
   Fix: make sure the last stop is rgba(..., 0) at 100%.

2. filter: blur is still applied — even with a mask, blur
   may extend beyond mask bounds if the mask is on the
   wrong element.
   Fix: ensure mask-image covers the same box as the blur.
   Or remove blur entirely and rely on radial-gradient.

3. .eh-disk-pull has rotateX(52deg) + transform-style:
   preserve-3d — the 3D plane's silhouette may be visible
   as a parallelogram at the edges when it should be
   transparent.
   Fix: check if a background-color or box-shadow on
   .eh-disk-pull is bleeding at the edges. Remove any
   background on that element.

4. .eh-disk container itself may have a background-color
   or border that shows through.
   Fix: force background: transparent on .eh-disk,
   .eh-disk-pull, .eh-disk-rings.

Verify in browser: hover → no visible rectangular boundary.
Мой прогноз
Скорее всего — одно из:

Radial-gradient не доходит до rgba(0,0,0,0). Последний stop — полупрозрачный, не прозрачный. Граница = этот остаток.

.eh-disk-pull имеет background или box-shadow, который проецируется на плоскость после rotateX. Видно как прямоугольник в 3D.

.eh-disk имеет родительский background — например, bg-nebula — который при hover становится заметным на фоне page-void.

Что проверить в DevTools
Быстро — 2 минуты:

Elements → наведи на диск → highlight каждый:

.eh-disk

.eh-disk-glow

.eh-disk-pull

.eh-disk-rings

Смотри — чей bounding box совпадает с видимым квадратом.

Computed → background-color у каждого.

Computed → filter у каждого.

Скинь — что нашёл.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

✅ VOID v3 — анимация работает.

🟧 Glow square — остался, едва заметный.

🟧 Wordmark flush — Photopea crop.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

PLUS:

Что делать
1. Квадрат — окончательный fix
Причина — скорее всего radial-gradient не доходит до 0 alpha или .eh-disk-pull проецирует границу после rotateX.

2. Кольца — толще на 1–2px
Сейчас: ~1px line.
Хочешь: 2–3px.

Что менять:

stroke-width (если SVG).

border-width (если CSS border).

height / thickness (если radial-gradient).

Промпт Cursor'у:

text
Two fixes. Plan first, no code.

================================================================
1. GLOW SQUARE — final pass
================================================================

Right and top edges of the glow are still visible as a
faint rectangular boundary. Diagnose in DevTools:

  - Check .eh-disk-glow computed background: does the
    radial-gradient's last stop reach rgba(0,0,0,0) at 100%?
  - Check .eh-disk-pull: any background, box-shadow, or
    filter that could bleed at the edges after rotateX?
  - Check .eh-disk / .eh-disk-rings: any background-color
    leaking through?

Fix the actual cause. Do NOT add more masks on top of
masks — find the leak.

================================================================
2. RINGS — slightly thicker
================================================================

The rings (accretion disk) that flatten to the equator are
currently ~1px — too thin to read as «forming a disc».

Increase thickness by 1–2px:
  - If SVG <ellipse>: stroke-width: 1 → 2 or 2.5.
  - If CSS border: 1px → 2px.
  - If radial-gradient: increase the inner color stop band.
Keep the same colors (gold, cyan, violet) and glow balance.

Verify in browser: rings read clearly as they flatten to the
equator; no square boundary around the glow.

================================================================
PROCESS
================================================================
  - Diagnose (1) with DevTools, report the cause.
  - Then apply both fixes in one commit.
  - One screenshot at 1920×1080 with hover active.
  - No push until I OK.

DO NOT TOUCH: hero, nav, games, seed, refund, authors.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

✅ VOID v3 — анимация работает.

🟧 Glow square — окончательный fix.

🟧 Rings thicker (+1–2px).

🟧 Wordmark flush — Photopea crop.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

ADDITIONALLY:
Понял. Сначала коммит — потом анимация, потом PNG
План действий
Порядок:

Коммит — зафиксировать текущий VOID v3 (работает, с квадратом и тонкими кольцами — но это лучше, чем пустота).

Fix glow square + rings thicker — Cursor доделает анимацию.

PNG crop в Photopea — финальный штрих.

Verify + push.

Wave 2 #6 — Flappy.

Commit — что фиксировать
Cursor писал:

6722de1 — VOID v2 as-is (local only, not pushed).

Плюс поверх — v3 (reducedMotion fix, stars 4-point, equator pull). Всё это — local, не запушено.

Промпт Cursor'у:

text
Before more VOID work, commit the current state.

Please commit everything VOID-related that's currently
uncommitted on top of 6722de1:

  - Stars 4-point (already in 6722de1 or after?)
  - Equator pull (rings flatten, spin pause)
  - reducedMotion debug + system respect
  - Any theme.css updates since 6722de1

Subject: feat(home): void v3 — equator pull, 4-point stars,
reduced-motion respect

Body:
  - Rings flatten to equator on hover (--ry 0.4 → 0.06).
  - Spin pauses, glow up, particles spiral.
  - Stars: 4-point for colored particles; tiny gold = dots.
  - prefers-reduced-motion: fully static (verified).
  - Debug logs behind import.meta.env.DEV.

Do NOT push. Just commit locally so we can revert if needed.
После коммита — git log покажет hash — точка отката есть.

Дальше — 2 fix
После commit:

Glow square — окончательный fix (radial-gradient / pull).

Rings thicker — +1–2px.

Промпт — тот же, что я дал выше.

Потом — PNG
Ты наконец созрел. 5 минут в Photopea:

Открой logo-minimal.png.

Crop слева/справа по границам букв EVENT HORIZON.

Экспорт → replace.

Скинь Cursor'у:

text
Replaced logo-minimal.png — wordmark is now flush with
image edges. Verify hero wordmark aligns with «Выбери игру»
text block edges.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас — план:

⏸ Commit VOID v3 (local).

🟧 Fix glow square + rings thicker.

🟧 PNG crop — 5 минут.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт на commit.

После commit hash — отправь fix glow + rings.

Параллельно — crop PNG.

Push когда ок.