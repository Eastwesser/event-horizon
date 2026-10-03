Что подтвердилось
✅ Stock fix работает
Купил Пращника → сток исчез → недоступен другим.

Это правильное поведение — «sold-out is sold-out».

Один игрок купил — карта недоступна для остальных.

Цель достигнута.

✅ Detail page — отлично
Скриншот Сприггана:

Badge ЛЕТАЮЩИЙ — работает ✅

Секции: Бой / Карта / Автор — вместо сырого dump ✅

Поля с человеческими названиями:

Здоровье: 7

Ход: —

Удар: Простой удар 1-2-2

Стоимость: 3 · Элитная

Иконки: —

Выпуск: Сердце Роя · #8 · 2026

Номер: 96

Редкость: Обычная

Стихия: леса

Тип: Существо — Летающее

Класс: Страж леса

Художник: Ж.Насыранбеков

ID: zh_lasyranbekov

Card text, flavor text — на месте

Цена 🎫 10000, В наличии: 3 ✅

4–5 закрыты отлично. Даже лучше, чем я ожидал.

🔴 Новый баг — не та картинка у Пращника
Ты сказал:

Я купил пращника (хотя это нихуя не пращник, а дракон)

Что произошло:

В seed-файле *_cards_info.md (или в файловой структуре) 019_prashchnik_4_common.jpg содержит не ту картинку — там дракон, не пращник.

Данные seed'а правильные — имя, цена, атрибуты настоящего Пращника.

Картинка — чужая.

Почему: скорее всего, файл перепутан в исходном каталоге — либо при копировании, либо в исходнике карты был неверный JPG.

Хорошо: ты нашёл правильный:

/home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/019_prashchnik_4_common.jpg

Описание — в /home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/CRIT_PRASHNIK.md

Промпт для Cursor
text
Fix: wrong image on «Пращник» (item id — the one we just bought).

Problem:
  Seed used 019_prashchnik_4_common.jpg, but that file
  contains the wrong artwork (a dragon, not the Пращник card).
  All other data (name, stats, cost, artist) is correct.

Correct source files:
  /home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/019_prashchnik_4_common.jpg
  /home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/CRIT_PRASHNIK.md

Task:
  1. Read CRIT_PRASHNIK.md — it has the correct card data
     (check if any fields differ from what we seeded).
  2. Upload the correct JPG via POST /api/uploads.
  3. Update the existing «Пращник» item (do NOT create a new
     one):
     - images: [new upload URL]
     - update any attributes that differ per CRIT_PRASHNIK.md
  4. Do NOT touch stock, purchase records, or inventory.

Verify: open /shop/item/<prashnik-id> → correct artwork.

Do NOT re-seed. Do NOT touch other cards.
Также стоит проверить:

Другие карты могут иметь ту же проблему. Если один JPG перепутан — могут быть ещё. Спроси Cursor'а:

text
Also: audit for other cards whose image doesn't match
the card name. We found one (Пращник). Are there others?
Quick check: compare filename slug to card name for all 280.
Но это может быть долго. Опционально.

Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix verified — покупка decrement'ит, sold-out работает

4 Icons + grouped attrs ✅ (Бой / Карта / Автор)

5 Badges ✅ (ЛЕТАЮЩИЙ, foil, noir, companion, unique)

6 Tickets ✅ (🎫 на всех карточках)

7 Create modal ×1000 ✅

Idempotency ✅ (второй раз — В инвентаре disabled)

Текущее:

🔴 Пращник — неверная картинка, нужен фикс

🟧 Проверить, нет ли других карт с той же проблемой

Дальше по плану:

🟧 8 Pagination UI (100/страница)

🟧 9 Sorting (name / rarity / price / artist / set)

🟧 10 Artist pages

Следующая волна:

🟧 Attribute filters

🟧 Расширенный поиск (референс proberserk)

🟧 Noiz-комментарии в UI

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

