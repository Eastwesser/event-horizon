Итог 1→3
1–2 FE:

Dead icons dropped ✅

Collapse Class / Icons ✅

tsc clean

3 Sanity:

277/280 с hp — 3 без? (возможно артефакты без hp)

icons: array — всегда ✅

price/stock: числа ✅

card_no/set_number/artist_id — 280/280 ✅

foil 11 / noir 3 ✅

7 companion ✅

artifact + hp без attack ✅ (Склеп Керсам + другие)

Dead icons отсутствуют ✅

Фильтры работают на реальных данных.

🔴 Наблюдение 1 — element: "леса" на 56 картах
Cursor:

element: "леса" on 56 cards — rest use English (woods missing). Filter «Лес» (woods) skips those 56 until seed/normalize.

Что произошло:

В seed есть карты с element: "леса" (русский) и с element: "woods" (английский).

56 карт — русский, остальные — английский.

Фильтр ищет по woods → 56 карт пропускаются.

Где именно: смотри скрины — Спригган (7 сет woods) имел element: "леса" в detail (скрин от 03.10). Значит 7 сет или часть карт попала с русским element.

Это баг данных, не UI.

Варианты:

A. Fix seed — заменить леса → woods во всех info-файлах, пересобрать items.

B. Нормализовать в gateway — маппить леса → woods при чтении.

C. Fix фильтр — принимать и леса, и woods.

Мой выбор — A. Seed должен быть каноничным. Русский — только display, в БД — английский код.

Но: исправление seed + reseed 56 карт — риск. Можно сделать update-only (не delete): найти items с element=леса, обновить attributes.element="woods".

🟧 Наблюдение 2 — /api/shop/items thin DTO
Cursor:

/api/shop/items — thin DTO (attributes: null, stock: null). FE catalog path via inventory is fine; shop service alone isn’t enough for filters.

Что это значит:

/shop/items возвращает старый формат — без attributes и stock.

Каталог через /inventory/items — полный.

Фильтры работают — потому что FE использует inventory.

/shop/items — избыточный, можно отложить.

Это не блокер. Но техдолг: два эндпоинта, разные DTO — путаница в будущем.

Что делать — 2 варианта
A. Fix element: леса → woods в seed
Плюс: данные чистые, каноничные.
Минус: reseed 56 items — риск, если что-то пойдёт не так.

Промпт:

text
Data fix: normalize element values.

Sanity found 56 cards with attributes.element = "леса"
(Russian) while the rest use English codes (woods, darkness,
mountains, steppes, swamps, neutral).

Fix the seed source first:
  - Grep all *_cards_info.md for "леса" / "Стихия Леса" etc.
  - Replace with "woods" (English, canonical).
Then update the existing 56 items in DB (NOT recreate):
  - Find items where attributes.element = "леса".
  - Update to "woods".
  - Do NOT touch other fields, stock, images.
Verify:
  - curl .../inventory/items?element=woods → count includes the
    previously-missing 56.
  - No items left with element = "леса".
Report counts before/after.
B. Нормализация в gateway
Плюс: не трогать данные, безопаснее.
Минус: русский в БД — остаётся. Если будешь добавлять новые карты через UI — могут появиться новые "леса".

Мой выбор: A — фикс в источнике. B — временное решение.

Что ответить Cursor'у
Приоритет:

Fix element data — 56 карт сейчас недоступны через фильтр. Это функциональный баг.

Bottleneck wave — архитектура, не горит.

Thin DTO /shop/items — отложить.

Промпт:

text
Two things:

1. DATA FIX (priority) — normalize element values.

56 cards have attributes.element = "леса" (Russian), breaking
the "Лес" filter. Fix:

  a) Seed source: grep all *_cards_info.md and replace
     "леса"/"Стихия Леса" with "woods" (English code).
  b) Live DB: update the existing 56 items' attributes.element
     from "леса" to "woods". Do NOT delete, do NOT recreate,
     do NOT touch stock/images/other fields.
  c) Verify:
     - curl .../inventory/items?element=woods → includes the 56.
     - no items left with element = "леса".

Report counts before/after, files changed.

Also check: are there other Russian element variants in the
data? (darkness/тьма, mountains/горы, steppes/степи,
swamps/болота, neutral/нейтрал.) Normalize all to English
codes the same way.

2. THEN bottleneck wave.

Thin DTO /shop/items — defer. FE uses inventory path, so it
doesn't block anything.
Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Stock fix + idempotency + purchase

4 Icons + grouped attrs

5 Badges, 6 Tickets, 7 Create modal ×1000

Пращник fixed

8 Pagination, 9 Sorting, 10 Artist pages

A–H visual + UX polish

1→4 final polish

Attribute filters ✅

1→3 polish ✅ (dead icons, collapse, sanity)

Критичный фикс (сейчас):

🔴 element: "леса" на 56 картах — фильтр Лес их пропускает

🟧 Возможно другие русские значения (тьма, горы, степи, болота, нейтрал)

Следующая волна:

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии в UI

🟧 Author registration

🟧 Refund / cancel purchase

Техдолг (отложено):

🟧 /shop/items — thin DTO, разобраться позже

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр