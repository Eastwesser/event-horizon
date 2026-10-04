Да, растянуть. По границам shell — как Выбери игру
Что ты хочешь
EVENT HORIZON — большая, растянута по ширине shell:

Левая граница E в EVENT = левая граница Выбери.

Правая граница N в HORIZON = правая граница карточек игр (или кнопки Лидерборд).

Не центр, а растянуто по ширине. Как в блокбастерах на постерах.

Как это сделать — CSS
letter-spacing + font-size + transform: scaleX():

Вариант A — letter-spacing (правильный)
css
.eh-hero-wordmark {
  font-size: clamp(4rem, 10vw, 8rem);
  letter-spacing: 0.15em;   /* трекинг */
  width: 100%;
  max-width: <same as shell>;
  text-align: justify;       /* растянет буквы по всей ширине */
}
Проблема: text-align: justify не работает на одной строке без пробелов между словами. Нужен hack:

или — display: flex + justify-content: space-between на буквах.

или — transform: scaleX(...).

Вариант B — scaleX (визуально, но искажает)
css
.eh-hero-wordmark {
  transform: scaleX(1.05);   /* растянуть на 5% */
}
Минус: буквы искажаются — шире, чем надо.

Вариант C — flex + space-between (наиболее правильно)
Разбить EVENT HORIZON на буквы:

html
<h1 class="eh-hero-wordmark" aria-label="EVENT HORIZON">
  <span>E</span><span>V</span><span>E</span><span>N</span><span>T</span>
  <span> </span>
  <span>H</span>...<span>N</span>
</h1>
css
.eh-hero-wordmark {
  display: flex;
  justify-content: space-between;
  width: 100%;
  max-width: <shell>;
}
Буквы разъедутся по всей ширине. Постер-стиль.

Минус: доступность — screen-reader прочитает по буквам. Fix: aria-label="EVENT HORIZON" + каждую <span> aria-hidden.

Мой выбор — C (flex + space-between)
Причины:

Точный контроль границ — E в EVENT прилипает к левому краю, N в HORIZON — к правому.

Не искажает буквы (в отличие от scaleX).

Соответствует стилю — блокбастер / постер.

Адаптивно — flex работает везде.

Плюс: внутри — пробел между EVENT и HORIZON. Этот пробел — <span> с flex-grow или отдельный spacer.

Промпт Cursor'у
text
Hero wordmark — stretch, don't center.

Goal:
  - Left edge of «E» in «EVENT» aligns with left edge of
    «Выбери игру и» (shell content box left).
  - Right edge of «N» in «HORIZON» aligns with right edge
    of the games grid cards / «Лидерборд» button (shell
    content box right).
  - Full width, no centering.

Best approach — flex with space-between:

  <h1 class="eh-hero-wordmark" aria-label="EVENT HORIZON">
    <span aria-hidden="true">E</span>
    <span aria-hidden="true">V</span>
    ...
    <span aria-hidden="true" class="eh-hero-wordmark-gap"></span>
    <span aria-hidden="true">H</span>
    ...
    <span aria-hidden="true">N</span>
  </h1>

  .eh-hero-wordmark {
    display: flex;
    justify-content: space-between;
    width: 100%;
    max-width: <same as shell>;
    margin: 0 auto;
    padding: 0 <same horizontal padding as shell>;
  }
  .eh-hero-wordmark-gap { flex: 1; } /* the space between words */

Do NOT use text-align: justify (won't work for one line).
Do NOT use transform: scaleX (distorts letters).

Font size: should scale with container width so the
letters fill the height proportionally. Try:
  font-size: clamp(3rem, 8vw, 7rem);

Verify in browser at 1920×1080:
  - «E» flush with «Выбери» left edge.
  - «N» flush with the games grid right edge.

No push.
Что проверить после
#	Проверка	Ожидание
1	E в EVENT	flush лево с Выбери
2	N в HORIZON	flush право с карточками
3	Пробел между	равномерный
4	Размер букв	большой, заполняет ширину
5	1920×1080	ок
6	375×812	адаптив — сжимается
7	Screen reader	читает EVENT HORIZON не по буквам
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально (не запушено):

1694cc0 — MAINPAGE_TODO_1.

Hero Plan A refactor.

+ wordmark stretch — новое.

Сейчас:

🟧 Wordmark — растянуть по shell (flex space-between).

🟧 PNG прямоугольник — всё ещё виден (можно заодно усилить mask).

🟧 Verify.

⏸ Commit + push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — flex space-between.

Проверь — E flush с Выбери, N — с карточками.

Заодно — усиль mask для PNG (см. предыдущий промпт).

Push.

Wave 2 #6.


UPDATED INFO:

Дополняю. Финальная версия промпта Cursor'у
Что нужно отправить
Один промпт — три задачи:

Wordmark stretch — flex space-between.

PNG mask — усилить радиально.

Verify + commit — после.

Итоговый промпт
text
Three fixes for the hero, then verify + commit.

================================================================
1. WORDMARK — STRETCH, DON'T CENTER
================================================================

Goal:
  - Left edge of «E» in «EVENT» = left edge of «Выбери игру и»
    (shell content box left).
  - Right edge of «N» in «HORIZON» = right edge of the games
    grid cards / «Лидерборд» button (shell content box right).
  - Full width of the shell, no centering.

Approach — flex with space-between:

  <h1 class="eh-hero-wordmark" aria-label="EVENT HORIZON">
    <span aria-hidden="true">E</span>
    <span aria-hidden="true">V</span>
    <span aria-hidden="true">E</span>
    <span aria-hidden="true">N</span>
    <span aria-hidden="true">T</span>
    <span aria-hidden="true" class="eh-hero-wordmark-gap"></span>
    <span aria-hidden="true">H</span>
    <span aria-hidden="true">O</span>
    <span aria-hidden="true">R</span>
    <span aria-hidden="true">I</span>
    <span aria-hidden="true">Z</span>
    <span aria-hidden="true">O</span>
    <span aria-hidden="true">N</span>
  </h1>

  .eh-hero-wordmark {
    display: flex;
    justify-content: space-between;
    width: 100%;
    max-width: <same as shell>;
    margin: 0 auto;
    padding: 0 <same horizontal padding as shell>;
    font-size: clamp(3rem, 8vw, 7rem);
    letter-spacing: 0;
  }
  .eh-hero-wordmark-gap { flex: 1; }

Do NOT use text-align: justify — won't work for a single line.
Do NOT use transform: scaleX — distorts letters.

================================================================
2. PLANET PNG — RECTANGLE STILL VISIBLE
================================================================

Screenshot: the planet PNG is still visibly a rectangle on
the page void.
  - Vertical edges at both sides (~1470px on the right).
  - Horizontal top edge just below the wordmark (~200px).
  - Horizontal bottom edge around ~440px.

Cause: the PNG's dark background color does not match
page-void (#0b0e1a) exactly, so its rectangular boundary
shows.

Fix — stronger mask-image on .eh-hero-planet:

  mask-image: radial-gradient(
    ellipse 75% 85% at center,
    black 50%,
    transparent 100%
  );
  -webkit-mask-image: same (for Safari).

Also check: report the PNG's corner pixel color (upper-left,
upper-right, lower-left, lower-right) and the page-void
color. If the delta is >5% per channel, we may need to
re-export the PNG with a matched background.

Extend the existing side fade from ~8% to ~25% if needed.

================================================================
3. VERIFY + COMMIT
================================================================

After both fixes, verify in browser at 1920×1080:
  - «E» flush with «Выбери» left edge.
  - «N» flush with games grid right edge.
  - No visible rectangular boundary around the planet.
  - Whole planet + horizon visible.
  - Top/bottom fades smooth.
  - Wordmark stays crisp on zoom (it's HTML, not raster).

If everything is clean → commit:

  feat(home): native hero — stretched wordmark + smooth
  planet mask

  Body:
    - Wordmark: flex space-between, E and N flush with
      shell edges; aria-label retained; spans aria-hidden.
    - Planet PNG: radial mask-image to remove rectangular
      boundary; extended side fade.
    - Removed old logo-minimal.png usage (already done in
      previous commit if separate).

Do NOT push until I confirm.

DO NOT TOUCH: nav, disk, games, seed, refund, authors.

Report: PNG corner colors vs page-void color, and screenshot
description after fixes.
Про 5174
Cursor написал:

The extra Vite process (port 5174) was aborted and cleaned up. Your app is still up on http://localhost:5173/.

Ок. Только 5173 — используй его.

Что проверить тебе после
#	Проверка	Ожидание
1	E в EVENT	flush лево с Выбери
2	N в HORIZON	flush право с карточками
3	Пробел между словами	равномерный
4	Размер букв	большой, заполняет shell
5	PNG прямоугольник	исчез (радиально mask)
6	Планета целиком	видна
7	Fades сверху/снизу	плавные
8	1920×1080	ок
9	375×812	адаптив
10	Screen reader	читает EVENT HORIZON
11	Zoom Ctrl+	буквы резкие
Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

VOID core fix — 5a18d99.

VOID v3 — 2abb45d / ebf6a92.

Hero baseline — 8bfa3db.

Локально (не запушено):

1694cc0 — MAINPAGE_TODO_1.

Hero Plan A refactor (planet PNG + HTML wordmark).

Сейчас:

🟧 Wordmark — stretch flex space-between.

🟧 PNG mask — радиально.

🟧 Verify + commit.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

Действия
Отправь Cursor'у промпт — 3 задачи.

Проверь 11 пунктов.

Скинь — скрин + PNG corner colors (если Cursor отрепортит).

Push после OK.

Wave 2 #6.