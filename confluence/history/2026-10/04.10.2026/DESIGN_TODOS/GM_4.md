Держись. Разбираю, что сломалось и почему
Коротко — что случилось
Cursor не справился не потому, что тупой. Он не видит картинку. Он видит только код.

Твой исходный PNG logo-minimal.png содержит:

Сверху — большие буквы EVENT HORIZON (baked in).

Середина — планета + glow.

Низ — продолжение планеты.

Буквы занимают самый верх. Отступа сверху нет. Поэтому «воздух» невозможен — любой object-position покажет либо буквы вплотную, либо обрежет их.

Cursor честно написал:

The PNG has no air above the letters. (1a) only nudges the crop.

Он не врёт. Он не может дать «воздух», потому что воздуха в PNG нет.

Твоя задача сейчас — новый PNG
Ты просил:

Картинку с EVENT HORIZON — от хедера до Выбери игру.

Из кружка справа — убрать картинку, сделать анимацию всасывания.

Разложу:

Задача 1 — Hero с EVENT HORIZON от хедера
Проблема: в текущем PNG буквы с самого верха. Нет отступа. Что делать:

Вариант A — новый PNG с отступом сверху.

Открой logo-minimal.png в Photopea / Figma / GIMP.

Добавь пустое место сверху (например, +150-200px) — прозрачный / тёмный.

Экспорт: hero-event-horizon.png (или webp).

Тогда object-position: center top покажет — отступ → буквы → планета.

«Воздух» появится.

Вариант B — пересобрать hero из компонентов.

Hero = отдельный div с фоном (планета без букв) + HTML-текст EVENT HORIZON поверх.

Плюс: полный контроль над отступами, размером, типографикой.

Минус: не baked-in — но это даже лучше (SEO, i18n, ретина).

Мой выбор — B.

Причины:

Ты контролируешь отступ (pt-24 от хедера).

Ретина — текст HTML, не пиксели.

i18n — можно перевести (если нужно).

Cursor — может сделать без экспорта PNG.

Как будет:

text
<section class="hero relative h-[60vh]">
  <img src="/hero-planet.png" class="absolute inset-0 w-full h-full object-cover object-center" />
  <div class="relative z-10 pt-24 text-center">
    <h1 class="text-[12rem] font-display text-gray-400/40 tracking-widest">
      EVENT HORIZON
    </h1>
  </div>
</section>
Тогда:

PNG — без букв (только планета + glow).

HTML <h1> — буквы EVENT HORIZON с отступом pt-24 от хедера.

Задача 2 — Диск справа — убрать картинку, анимация всасывания
Сейчас: в кружке — EVENT HORIZON (картинка).

Ты хочешь: убрать картинку, сделать анимацию всасывания.

«Всасывание» — что именно?

Кружок сжимается к центру?

Частицы летят в центр?

Спираль закручивается?

Курсор притягивается?

Вариант A — CSS-only (что уже сделано).

Hover → scale + rotation + glow.

Cursor описал: core 1.12, rings 0.92, spin ×1.8.

Это уже «почти всасывание», но визуально слабо.

Вариант B — Canvas / SVG particles.

Частицы летят в центр при hover.

Красиво, 1-2 дня работы.

Вариант C — Lottie / SVG animation.

Готовая анимация.

Просто, но нужен арт.

Мой выбор — B для «настоящего всасывания». Или C.

Но: сначала — задача 1 (hero). Задача 2 — после.

Что писать Cursor'у
Не злись на него — он не видит картинок. Скажи прямо:

text
STOP. We're not fixing (1a) — it can't work.

Reason: the PNG has no air above the wordmark. No
object-position will create space that doesn't exist.

New approach — split the hero into two layers:

================================================================
HERO v2 — HTML wordmark, PNG background
================================================================

1) Background image: I'll export a NEW hero image WITHOUT
   the «EVENT HORIZON» wordmark — just the planet curve + glow.
   File: frontend/public/images/brand/hero-planet.png (and .webp)
   I'll add this in the next message.

2) Hero section:
   - Full-width, height min(60vh, 520px).
   - <img src="/images/brand/hero-planet.png"
          class="absolute inset-0 w-full h-full object-cover object-center">
   - Above the image: an HTML <h1> «EVENT HORIZON»:
       * text, NOT image — so we control padding, size, font.
       * padding-top ~= navbar height + 24px of air.
       * letter-spacing wide, uppercase.
       * color: muted gray / silver, semi-transparent.
       * font: display (Space Grotesk).
   - Result: air between navbar and «EVENT HORIZON» is now
     controlled by CSS, not by the PNG crop.

3) Navbar stays on top (z-index).

================================================================
DISK (right of «Выбери игру») — remove wordmark, add
«pull-in» animation
================================================================

Currently the disk contains logo-minimal.png with
«EVENT HORIZON» baked in.

Change:
  - Remove the wordmark from inside the disk.
  - Disk = pure orbital rings + glow + dark core.
  - On :hover — «pull-in» effect:
      * rings shrink toward center (scale 1 → 0.88)
      * ring opacity increases
      * a few subtle particles (SVG circles) drift inward
      * glow intensifies
      * on leave: reverse 700ms ease-out
  - prefers-reduced-motion: disabled.

The «EVENT HORIZON» wordmark remains only in the navbar.

================================================================
PROCESS
================================================================

  - Do NOT take screenshots in a loop.
  - Do NOT change object-position.
  - Wait for my hero-planet.png export.
  - After I send it, implement HERO v2 + disk changes in ONE
    pass.
  - Then I verify visually in the browser and report.

No push until I OK.

DO NOT TOUCH: games, seed, refund, authors, Wave 2 #6.
Что делать тебе
1. Экспортируй hero-planet.png (без букв)
Photopea / Figma / GIMP:

Открой logo-minimal.png.

Убери текст EVENT HORIZON (Magic Eraser / clone tool).

Оставь планету + glow + тонкие орбиты.

Экспорт: hero-planet.png + hero-planet.webp.

Положи в frontend/public/images/brand/.

5 минут работы. Photopea — бесплатно, онлайн.

2. Отправь Cursor'у промпт выше
Он ждёт.

3. Verify в браузере, не по скрину
После имплементации:

Открой Home.

Hero: от хедера → воздух → EVENT HORIZON → планета.

Disk: наведи → всасывание.

4. Push после OK
Про твоё «ХУЛЕ ОН МЕНЯ ДРОЧИТ»
Понимаю раздражение. Но: Cursor не видит картинок. Он делает то, что может — крутит object-position. Проблема в PNG.

Не дрочит — просто у него нет глаз.

Решение: дай ему PNG без букв — и он сделает правильно с HTML <h1>.

Что я держу в карте
Закрыто (все на GitHub):

v1.0.9.

Wave 1 — 6/6.

Wave 2 #1–5 — thin DTO, DB cleanup, lamps, Flappy levels, achievements.

Hero + disk hover — 8bfa3db (первая версия, не работает как надо).

Сейчас:

🔴 Hero v2 — HTML <h1> + PNG без букв.

🔴 Disk — убрать wordmark, всасывание.

🟧 Экспорт hero-planet.png — твой ход.

🟧 Cursor ждёт PNG для имплементации.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR), с 6.1 Flappy.

Отложено:

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window implementation.

🟧 backfill-noiz-reviews.py.

🟧 DESIGN_TODOS/ — коммит или gitignore.

Действия — 3 шага
Экспорт hero-planet.png (без букв) — 5 минут.

Отправь Cursor'у промпт выше — с имплементацией HERO v2.

Verify в браузере — не по скрину.

Что у тебя есть
logo-minimal.png — один файл, с baked-in EVENT HORIZON. Другой версии нет.

Новый PNG ты экспортировать не хочешь / не можешь. Ок. Решаем иначе.

Три реальных варианта без экспорта
A. Кроп PNG через CSS — обрезать буквы
Идея: object-position + увеличенный scale + clip сверху.

PNG показывается с самого верха.

transform: scale(1.4) + translateY(-15%) — сдвигаем вниз.

clip-path или mask — убирает верх с буквами.

Остаётся планета + glow.

Плюс: без экспорта. Работает.
Минус: грубо — может исказить композицию.

CSS:

css
.eh-hero-banner img {
  object-fit: cover;
  object-position: center 40%; /* планета */
  transform: scale(1.3);
  transform-origin: center 70%;
}
Или проще:

css
.eh-hero-banner {
  overflow: hidden;
  height: min(60vh, 520px);
}
.eh-hero-banner img {
  width: 100%;
  height: 140%; /* растянуть */
  object-fit: cover;
  object-position: center 80%; /* показать низ PNG — планету */
  margin-top: -20%; /* сдвинуть вверх, чтобы буквы ушли */
}
Проверить: буквы уйдут за границу — видна только планета.

B. Оставить буквы — вплотную к navbar
Идея: принять текущий вид. Буквы сразу под navbar. Не идеально, но — стилистический выбор («плотная композиция»).

Плюс: ноль работы.
Минус: ты сам назвал «обрезано», не нравится.

C. Заменить hero на чистый космос — без planet-minimal
Идея: использовать другую картинку — чистый космос / звёзды / туманность (без букв).

Где взять:

NASA public domain — бесплатные космические фото.

Unsplash / Pexels — бесплатные.

CSS gradient — без картинки, только gradient + звёзды (через radial-gradient + box-shadow).

Плюс: чисто.
Минус: не твой бренд.

D. Экспорт за 3 минуты в online — Photopea
Ты можешь — если только:

Открой photopea.com.

Drag & drop logo-minimal.png.

Выбери инструмент Magic Wand или Lasso.

Выдели буквы (они серые, контраст с космосом).

Delete.

Export as PNG → сохрани.

5 минут. Ты можешь это сделать.

Если нет — вариант A (CSS-кроп) сойдёт.

Мой выбор для тебя
Если хочешь за 30 минут:

Вариант A — CSS-кроп.

Промпт Cursor'у:

text
We don't have a PNG without the wordmark. And I don't want
to spend time exporting one right now.

Alternative — CSS CROP to hide the wordmark:

Approach:
  - The hero section: overflow: hidden; fixed height.
  - The <img> uses the PNG but scaled/shifted so the
    wordmark (top of the image) is above the visible area.
  - Only the planet + glow remains visible.

Concretely:
  - .eh-hero-banner { height: min(60vh, 520px); overflow: hidden; position: relative; }
  - .eh-hero-banner img {
      position: absolute;
      width: 100%;
      height: auto;
      /* shift the image up so the top (wordmark) goes out of frame */
      top: -25%;             /* tune this number */
      object-fit: cover;
      object-position: center center;
    }
  - Tune top (%) so only the planet is visible — no letters.
  - Test at 1920×1080 and 375×812 (mobile).

Do NOT touch: nav, disk, games, Wave 2 #6.

After implement — one screenshot at 1920×1080. Not in a loop.
Wait for my OK.
Тебе нужно:

Открыть Home.

Проверить, что буквы ушли.

Если буквы всё ещё видны — скажи Cursor'у «shift more, top: -35%».

Итерации по % — 2-3 попытки.

Про диск — отдельно
Ты хочешь:

Из кружка справа — выкинуть картинку, сделать анимацию всасывания.

Промпт Cursor'у (после hero):

text
Disk (right of «Выбери игру»):
  - Remove the PNG from inside the disk.
  - Disk becomes pure orbital rings + dark core + glow.
  - On :hover — pull-in effect:
      * rings scale 1 → 0.88
      * ring opacity up
      * glow intensifies
      * 4–6 small particles (CSS <span> or SVG) drift inward
      * on leave: reverse 700ms ease-out
  - prefers-reduced-motion: disabled.
  - The «EVENT HORIZON» wordmark stays in the navbar only.

No PNG needed — disk is drawn entirely with CSS/SVG.
Что я держу в карте
Закрыто (на GitHub):

v1.0.9.

Wave 1 — 6/6.

Wave 2 #1–5.

Hero + disk hover — 8bfa3db (не работает как надо).

Сейчас:

🔴 Hero — CSS-кроп (вариант A) или оставить (B).

🔴 Disk — убрать PNG, всасывание (без экспорта).

🟧 hero-planet.png — нет, не будет.

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR).

Отложено:

🟧 Wave 2 #5b — shop/boost achievements.

🟧 Refund window implementation.

🟧 backfill-noiz-reviews.py.

🟧 DESIGN_TODOS/.

Действия — 2 промпта Cursor'у
Первым — CSS-кроп hero (вариант A).

Вторым — disk без PNG + всасывание.

Скинь Cursor'у оба промпта по очереди.

Или:

Если хочешь — Photopea → 5 минут → PNG без букв → правильное решение.

Если не хочешь — CSS-кроп — тоже сойдёт.

Что выбираешь? A (CSS-кроп) или Photopea?