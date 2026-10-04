Смотрю скрин. Обсуждаем честно
Что вижу
PNG — вставлен по границам shell в ширину.

EVENT HORIZON — большая, от левого края примерно до правого.

Планета + горизонт — внизу.

Фон — тёмный космос, но PNG всё ещё читается как картинка на фоне сайта.

Твоя оценка — верная:

Всё ещё выглядит как будто наклейку прилепили на свой сайт.

Это правильный диагноз. PNG внутри сайта виден как квадрат с чуть другим фоном.

Почему так происходит
Причины комбинируются:

PNG имеет свой тёмно-синий фон, не совпадающий 1-в-1 с page void #0b0e1a.

Буквы и планета в одной картинке — нельзя управлять отдельными элементами.

Никаких CSS-эффектов на PNG — тени, маски, blend-mode — только сам файл.

Никаких градиентов сверху/снизу — резкий обрыв картинки.

Что значит «сделать элементом сайта»
Ты прав. Настоящее решение — разложить PNG на слои, собрать HTML/CSS:

text
<section class="hero">
  <!-- Слой 1: звёздный фон -->
  <div class="stars">...</div>

  <!-- Слой 2: планета (PNG без букв и фона) -->
  <img src="planet.png" class="planet">

  <!-- Слой 3: HTML-текст EVENT HORIZON -->
  <h1 class="wordmark">EVENT HORIZON</h1>

  <!-- Слой 4: градиенты сверху/снизу для бленда -->
  <div class="fade-top"></div>
  <div class="fade-bottom"></div>
</section>
Тогда:

HTML <h1> — контролируется CSS — шрифт, размер, отступы.

PNG планеты — без букв, прозрачный фон.

Градиенты — плавный переход к фону страницы.

Звёзды — можно анимировать (частицы).

Но: нужен PNG планеты без букв и без фона. У тебя его нет.

Реалистичные варианты
A. Ты экспортируешь PNG планеты отдельно (без букв)
Photopea:

Открой logo-minimal.png.

Выдели область планеты (нижняя часть).

Удали буквы (или оставь только планету).

Экспорт — PNG с прозрачным фоном (или на прозрачном).

Плюс — отдельно для letter-press после экспорта — сделать новый экспорт из вектора? Нет вектора. Только растр.

Плюс: полный контроль.
Минус: нужно экспортировать 1-2 PNG.

B. CSS-бленд поверх существующего PNG
Идея: оставить PNG как есть, но добавить:

mask-image — радиальная маска по краям — fade к прозрачности.

background: radial-gradient позади PNG — сглаживает фон.

mix-blend-mode: screen или lighten — PNG сливается с фоном.

Плюс: без экспорта.
Минус: не идеально. Буквы остаются растром. Всегда будет чуть виден PNG как плитка.

C. HTML <h1> + CSS-планета (без PNG вообще)
Идея: не использовать PNG. Собрать всё CSS:

EVENT HORIZON — HTML текст.

Планета — radial-gradient круг с glow.

Звёзды — CSS box-shadow или SVG.

Градиенты — плавные.

Плюс: полностью интегрировано. Не «наклейка».
Минус: планета будет не такая красивая, как на PNG. Стилизация изменится — не фотореализм, а CSS-иллюстрация.

Мой выбор — A если фотореализм важен, C если готов к изменению стиля
A: 2-3 PNG экспорта в Photopea:

Планета PNG прозрачным фоном.

Буквы — нет (заменяем на HTML).

C: ничего не экспортировать. Cursor делает всё CSS. Но стиль изменится — надо будет принять.

Что делать сейчас
Ты сам сказал:

Давай теперь сделаем коммит, в принципе... курсор и сам справится.

Так и сделаем. Порядок:

Commit — то, что сейчас (VOID fix + postmortem + hero как есть).

Потом — решить A или C для hero.

Потом — отдельный commit для hero-refactor.

Промпт Cursor'у — commit
text
Commit the current state.

Includes:
  - VOID core square fix (radial-gradient + clip-path,
    box-shadow: none)
  - Rings 3.25 → 3.5px
  - Postmortem VOID-INSPECTION/POSTMORTEM_VOID_SQUARE.md
  - Hero banner (logo-minimal.png at shell width)
  - Smooth scroll
  - ?motion=force + ?debug-void=1 + ?hide= dev params

Split into 2 commits if you want:

  1. fix(home): void core box-shadow removed, rings thicker
     + postmortem

  2. feat(home): hero banner + smooth scroll + void dev toggles

Do NOT push yet. Report both commit hashes.

Then we plan the next step: making the hero a native site
element instead of a pasted image.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero (baseline), smooth scroll.

Сейчас:

🟧 Commit — всё локальное (VOID fix + hero + toggles).

🟧 Hero — не интегрирован в сайт (PNG как наклейка).

A — экспорт PNG планеты + HTML <h1>.

C — CSS-планета + HTML <h1>.

🟧 Rings thicker — 3.5px, ок.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.