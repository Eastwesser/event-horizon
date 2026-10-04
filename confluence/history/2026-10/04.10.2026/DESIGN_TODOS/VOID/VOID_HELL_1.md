Что вижу на последнем скрине (после rotateX fix)
До этого я смотрел на предыдущий скрин — там квадрат был виден. На этом — тоже вижу. Реальный квадрат, не мнимый. Правый верхний угол, правый нижний, левый верхний — границы заметны.

Причём заметнее на правой стороне и по верху. Тёмно-синеватая граница, чуть отличающаяся от фона.

Cursor сказал, что rotateX убран, но квадрат остался. Значит — это НЕ rotateX. Что-то другое.

Причина — не rotateX
Раз rotateX убран, а квадрат остался — это:

.eh-disk-glow — radial-gradient не доходит до 0 alpha? Или blur обрезается?

.eh-disk или .eh-disk-pull — background-color проецируется?

Родительский контейнер — bg-nebula или bg-void — накладывается как прямоугольник за диском?

SVG viewBox — имеет background?

План — слой за слоем
Твой подход — правильный. Идём по слоям.

Промпт Cursor'у:

text
Square is STILL visible after removing rotateX. So the
source is something else.

Do a layer-by-layer elimination. In the DEV build (only
import.meta.env.DEV), add a temporary URL param to outline
each layer:

  ?debug-void=1  → every .eh-disk-* layer gets a 1px solid
                   border of a distinct color:
                     .eh-disk       → red
                     .eh-disk-glow  → green
                     .eh-disk-pull  → blue
                     .eh-disk-rings → magenta
                     .eh-disk-core  → cyan
                     svg.particles  → orange

Then I can see in the browser which bounding box matches
the visible square.

Report before implementing: which layers exist and what
their current computed styles are for:
  - background
  - background-color
  - filter
  - mask-image
  - box-shadow
  - mix-blend-mode

Include any inline styles or parent backgrounds that could
bleed through.

DO NOT change anything yet. Just add the debug outlines
behind a flag, and paste the computed styles for each
layer.

No push.
Пока Cursor делает — что можно проверить тебе
DevTools → Elements:

Ctrl+Shift+C → наведи на видимую границу квадрата.

DevTools выделит элемент — имя класса.

Скинь — что выделилось.

Если DevTools выделяет:

.eh-disk-glow → glow виноват.

.eh-disk-pull → pull виноват.

.eh-disk → контейнер.

Что-то другое → это не VOID элемент, а родитель.

Плюс — Computed на выделенном:

background-color — ?

box-shadow — ?

filter — ?

mask-image — ?

Скинь — сразу увидим источник.

Ещё одна идея — «зеркало»
Квадрат имеет странную симметрию — правый верхний + правый нижний углы сильнее. Это напоминает:

SVG viewBox с background fill.

Или — backdrop-filter на родителе.

Или — filter: drop-shadow(...) с непрозрачным фоном.

Или — outline на focused элементе (после tab).

Тоже стоит проверить.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🔴 Square — не убран, не rotateX.

🟧 Layer debug — план.

🟧 Rings thicker — после square.

🟧 PNG crop — Photopea.

⏸ Push.

⏸ Wave 2 #6.