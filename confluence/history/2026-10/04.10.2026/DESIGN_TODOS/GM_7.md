Что вижу
Скрин 1 (текущий)
Hero band — ~520px (min(60vh, 520px)).

PNG — квадратная, 1238×1200 ≈ 1:0.97.

При width = shell (~1152px) и object-fit: contain → картинка хочет быть 1152×1118.

Но height = 520px → overflow: hidden → обрезает по вертикали.

Результат: видна только надпись (верхняя часть PNG). Планета + горизонт — обрезаны.

Скрин 2 (старый)
Full-bleed + object-fit: cover → картинка растянута по ширине и обрезана по бокам.

Видна планета + горизонт + надпись.

Почти то, что нужно — но не full-width по границам текста.

Что ты хочешь
Вся картинка — надпись + планета + горизонт — от navbar до Выбери игру. По границам текста слева/справа.

Почему не получается
Физика:

PNG — квадратная (1:0.97).

Shell width — ~1152px.

Чтобы показать всю PNG — height ≈1118px.

Ты хочешь height ≈520px (hero) → в 2 раза меньше → обрезается.

Невозможно иметь и то, и то:

Либо — вся картинка → height ≈ 1150px (не hero, очень высоко).

Либо — height 520px → обрезка (по бокам или по вертикали).

Решения
A. Принять height ≈ 800–1000px
Hero становится высоким — надпись + планета + горизонт влезают целиком.

object-fit: contain + height: auto (или min(80vh, 900px)).

Плюс: вся картинка.

Минус: много места занимает, скролл появляется раньше.

B. Обрезать PNG до нужного aspect-ratio
PNG → обрезать сверху/снизу до aspect 16:9. Тогда при shell width 1152 height ≈ 648 — близко к 520.

Плюс: влезает.

Минус: PNG обрезана — нужен экспорт.

C. Вернуть full-bleed + cover (как скрин 2)
Как старый вариант — full-bleed, cover, height 520.

Плюс: работало.

Минус: не по границам текста (ты хотел иначе).

D. Разделить PNG на два слоя
Верхняя часть (надпись) — HTML <h1>.

Нижняя часть (планета) — PNG-фон.

Тогда — полный контроль.

Минус: нужен новый PNG без надписи.

Мой выбор — A
Принять height ≥ 800px. Hero будет высоким, но вся картинка — как ты хочешь.

Промпт Cursor'у:

text
Stop trying to fit the full square PNG into 520px height.
It's mathematically impossible — the PNG is square (1:0.97),
so at shell width 1152 the natural height is ~1118px.

Change:

  - Hero height: min(80vh, 900px) instead of min(60vh, 520px).
  - object-fit: contain.
  - width: 100% of shell.
  - overflow: visible (or keep hidden if you prefer).
  - Remove any vertical crop.

Result: the whole PNG visible — wordmark + planet + horizon.

If the hero now feels too tall:
  - Keep side mask for edge blend.
  - Consider reducing PNG display size slightly (e.g. max-width 900px) if 1152 is too wide.

Report exact dimensions after change.

Then we still need to solve wordmark alignment (a or b from
my earlier list). I'll decide after seeing the full-image
result.

No push. No disk yet.
Или — если хочешь быстрее — B
Обрежь PNG в Photopea:

Убери пустое сверху и снизу.

Оставь буквы + горизонт.

Aspect ≈ 16:9 или 2:1.

Экспорт → replace logo-minimal.png.

Тогда contain + height 520 — работает, вся картинка видна.

5 минут в Photopea.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5

Hero — в работе (8bfa3db)

Сейчас:

🔴 Hero — height 520 не влезает квадратную PNG.

A: height ≥ 800 (min(80vh, 900px)).

B: обрезать PNG в Photopea до 16:9.

C: вернуть full-bleed cover.

D: HTML <h1> + PNG без букв.

⏸ Disk pull-in — после hero.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

🟧 Wave 2 #5b, Refund window, backfill-noiz-reviews.py, DESIGN_TODOS/.