Хорошо. Значит — источник не внутри VOID
Раз квадрат не совпадает ни с одним debug-outline
Значит — источник вне VOID-слоёв. Что-то в родителе или сестринском элементе.

Промпт Cursor'у:

text
Debug outlines done. None of them matches the visible square.

  - red (.eh-disk): not visible
  - green (.eh-disk-glow): pulsing sphere, not the square
  - blue (.eh-disk-pull): not visible
  - magenta (.eh-disk-rings): unknown, not the square
  - cyan (.eh-disk-core): circle where particles pull in
  - orange (svg.particles): "like it, but not it"

So the square is NOT any .eh-disk-* layer. It's a parent or
a sibling element.

Next step: expand the debug outlines to the full ancestor
chain and to siblings of .eh-disk.

Add temporary outlines (behind ?debug-void=1) to:
  - every ancestor of .eh-disk up to <body>  — alternate colors
  - the section that wraps the hero + «Выбери игру» + disk
  - any bg-* wrapper (bg-void, bg-nebula) that contains the disk
  - any element with a background-color or background-image
    within 3 levels of the disk

Report the full ancestry from .eh-disk up to <body>: tag,
class, computed background, size.

Then I can point at the exact element.

No push.
Пока Cursor добавляет outlines предкам — проверь сам в DevTools
Быстрый способ найти источник:

DevTools → Elements.

Выбери .eh-disk.

Иди вверх по дереву:

.eh-disk → parent → parent → до <body>.

На каждом проверяй Computed:

background-color

background-image

box-shadow

filter

backdrop-filter

Если на одном из предков есть background или filter — это он.

Мой прогноз — backdrop-filter или bg-* на предке
Cursor писал:

parents transparent until .bg-void #0b0e1a
No backdrop-filter / mix-blend-mode / filter on VOID layers.

«On VOID layers» — но не сказал про родителей! Может быть у родителя есть:

backdrop-filter: blur() — размывает фон за ним — даёт прямоугольник размытого фона.

background: rgba(...) — если фон чуть отличается от page-void.

bg-nebula — на wrapper'е с диском — визуально чуть светлее void.

Часто — backdrop-filter на hero-секции или на .eh-glow wrapper'е даёт прямоугольник.

Что делать прямо сейчас
DevTools → Elements → выбери .eh-disk → иди вверх по дереву.

На каждом родителе Computed filter:

backdrop-filter

filter

background

Скинь — список предков с их computed.

Или — отправь Cursor'у промпт выше — он сделает outlines для предков — визуально сразу увидим.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🔴 Square — не в VOID-слоях, в предке.

Промпт Cursor'у — expand outlines.

Или — ты сам DevTools вверх по дереву.

🟧 Rings thicker — после square.

🟧 PNG crop — Photopea.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.