Принял. Разбор hero + disk + что дальше
Что вижу на скрине
Хорошо:

Full-bleed hero — работает.

Планета + горизонт — красиво.

Navbar поверх — читается.

EVENT HORIZON — большая надпись на фоне (уходит за верхний край).

Проблема — ты сам её назвал:

Надпись вверху слишком уж обрезана... Хотелось бы, чтобы картинка была вставлена по верхнему краю картинки — от конца хэдера (нижней грани хэдера).

То есть:

Сейчас — hero уходит вверх под navbar, object-position: center 60% сдвинуло картинку → верхняя часть (EVENT HORIZON) обрезана.

Хочешь — картинка начинается от нижней грани navbar, и верх картинки (надпись) виден.

Фикс:

object-position: center top — картинка сверху, не сдвинута.

Или — object-position: center 0% → показывает верх картинки.

Или — уменьшить height hero (с min(70vh, 640px) до min(60vh, 520px)) — картинка меньше обрезается.

Мой выбор: object-position: center top + height: min(65vh, 580px). Попробовать.

Второй момент — disk hover
Ты:

Эта же картинка находится внутри круга, там где ВЫБЕРИ ИГРУ. А мы говорили, что там должна быть интересная механика засасывания курсора визуально.

Что вижу на скрине: disk справа от Выбери игру — есть, с надписью EVENT HORIZON.

Hover-эффект — ты не показал в скрине (это второй скрин home-disk-hover-1920x1080.png). Не могу оценить без него.

Cursor заявил: Disk hover — реализован. Проверить:

Наведи на disk → scale растёт, rings сужаются, rotation ×1.8, glow.

Уход → обратно за 700ms.

prefers-reduced-motion → не работает.

Если не работает — фикс.

Один момент — дубль EVENT HORIZON
Сейчас:

Hero — большая EVENT HORIZON (обрезана сверху).

Accretion disk — та же надпись внутри круга.

Navbar — Event Horizon лого.

Три раза одна надпись. Визуально много. Варианты:

Оставить — это стиль (бренд).

Убрать из hero — hero без надписи, надпись только в navbar + disk.

Убрать из disk — оставить чистый диск, EVENT HORIZON — только в hero.

Мой выбор: убрать из hero (оставить чистую планету) + object-position: center top для композиции. Тогда:

Navbar — Event Horizon (компактно).

Hero — планета + горизонт, без надписи.

Disk — EVENT HORIZON внутри.

Если хочешь оставить hero с надписью — фикс object-position, но EVENT HORIZON должна быть видна целиком.

Что делать — 3 варианта
A. Быстрый фикс композиции
text
Fix hero object-position + height so the top (EVENT HORIZON)
is not cut. Change object-position to center top, reduce
height to min(60vh, 520px).
Плюс: быстро.
Минус: дубль надписи остаётся.

B. Убрать надпись из hero
text
Regenerate the hero image WITHOUT the EVENT HORIZON text
(just the planet + horizon glow). Keep object-position:
center top.
Плюс: чисто.
Минус: нужна новая картинка.

C. Комбинация
text
Fix object-position AND regenerate hero without text.
Мой выбор — B.

Причины:

Три раза одна надпись — визуальный шум.

Hero без текста — элегантнее.

EVENT HORIZON остаётся в navbar + disk — достаточно.

Kids-safe: ребёнок не сбит с толку тремя одинаковыми надписями.

Промпт Cursor'у
text
Review hero at 1920×1080.

Two issues:

1. OBJECT-POSITION — EVENT HORIZON text cut at the top
   Current: hero shows the planet + horizon, but the
   «EVENT HORIZON» wordmark baked into logo-minimal.png is
   clipped at the top edge.
   Fix: change object-position to «center top» so the top of
   the image (the wordmark) is visible from the navbar's
   bottom edge downward.
   Also consider: reduce height to min(60vh, 520px) if
   needed for a better crop.

2. WORDMARK DUPLICATION
   Currently three places show «Event Horizon»:
     a) navbar logo
     b) hero (baked into logo-minimal.png)
     c) accretion disk (inside the circle)
   Proposal (recommended): remove the wordmark from the hero.
   Use a version of the image WITHOUT text — just planet +
   horizon glow. Keep the disk with the wordmark.
   If removing text requires a new image export, we can
   instead keep the current image but fix (1) so the full
   wordmark is visible.

3. DISK HOVER — verify
   Send me the second screenshot (home-disk-hover-1920x1080.png)
   in text form: what exactly changes on hover (scale, rings,
   rotation, glow). I can't see images; describe or re-run
   with a console log.

Reply with your plan. No code yet.
Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1–5:

#1 0ee567f

#2 a881587

#3 8c16926

#4 2bbc62f

#5 146cae8 — verify не закрыт, push не сделан

Сейчас (в работе):

🟧 Hero + disk hover — 8bfa3db, review → fix → push

🔴 object-position: center top

🟧 Wordmark duplication — решить

🟧 Disk hover — verify (второй скрин)

⏸ Rebuild profile + gateway (Wave 2 #5)

⏸ Verify 23 пункта

⏸ Push Wave 2 #5

⏸ Бэкап

Wave 2 (осталось):

🟧 #6 Полиш игр (5, one per PR) — завтра, с 6.1 Flappy

Отложено:

🟧 Wave 2 #5b — shop/boost achievements

🟧 Refund window implementation

🟧 backfill-noiz-reviews.py
