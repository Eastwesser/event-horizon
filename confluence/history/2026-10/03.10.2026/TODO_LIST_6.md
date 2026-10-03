Что вижу
✅ Хорошо
Prev/Next + счётчик — работает, переключение быстрое.

Клик по Григорию Давыдову → его карты (/authors/:id) ✅

Фойл-фильтр — работает, все карты с ✦ ФОЙЛ badge.

Сортировка — пока нравится.

Detail page — секции Бой / Карта / Автор, Сет · Год · Художник — компактно.

Модалка покупки — карта крупная, цена/баланс/кнопки — ок.

Показать текст карты — collapsed, разворачивается.

🔴 Проблемы — 4 фикса
1. Белые углы у hero-картинки в detail.
Скрин 5 (Карбагаш): карта в detail прямоугольная, углы белые — потому что fit="contain" не обрезает углы. Контейнер имеет rounded-md, но картинка не следует ему.

Что нужно: контейнер с border-radius + overflow: hidden + фон --color-nebula. Тогда белые углы исчезнут — станут тёмными (letterbox по бокам карты — тёмный).

2. Описание карты в карточке магазина — не нужно.
Скрин 6 (Карбагаш в сетке): На глубине:zov. На поверхности:zor. При... — обрезанный текст карты.

Что нужно: убрать это из ShopItemCard. Оставить:

Картинка.

Название + badges.

В наличии: N.

Цена + Купить.

Убрать: description (обрезанное описание).

3. Описание в модалке — не нужно.
Скрин 7: На глубине:zov. На поверхности:zor. При всплытии — метание "языка" на 1, только по летающему существу стоимостью 5 или менее; оно приземляется с move 1, на свободную клетку рядом с Карбагашем до конца вашего хода.

Что нужно: заменить на преимущества/атрибуты карты:

Стихия: Болота

Редкость: Редкая

Фойл: Да

HP: 18

Удар: 4-5-7

Класс: Акванит

Короткая сводка, не полный текст.

4. Кнопка покупки в detail — добавить.
Ты сказал: «Можно кстати и туда добавить кнопку покупки, чтобы пользователь много раз не тыкался».

Что нужно: в ShopItemDetail — кнопка Купить рядом с ценой. Чтобы не возвращаться в сетку. После покупки → В инвентаре.

Дополни к промпту
Обновлённый промпт для Cursor:

text
Wave: final polish + purchase UX. Four items.

================================================================
1. DETAIL HERO — white corners (BUG)
================================================================

Screen: /shop/item/... (Карбагаш). Hero uses fit="contain",
but the image's own corners are still rectangular — white
shows at the four corners of the image inside the container.

Cause: with contain, the image doesn't reach the container's
rounded corners. The container has rounded-md, but the image
sits inside without clipping.

Fix:
  - The hero container: rounded-md + overflow: hidden +
    background: var(--color-nebula).
  - Image inside: object-fit: contain.
  - Result: letterbox sides are dark (nebula), and no white
    shows at the corners of the image.
Apply to ShopItemDetail, InventoryItemDetail heroes.

================================================================
2. SHOP CARD — remove card text
================================================================

Screen: /shop — ShopItemCard shows a truncated card text
("На глубине:zov. На поверхности:zor. При...").

Remove: the truncated description.
Keep:
  - Image
  - Name + badges (Карточка, С Компаньон, ✦ ФОЙЛ, ...)
  - "В наличии: N" / "Нет в наличии"
  - Price + Купить button

Result: cards are shorter, cleaner, align better.

================================================================
3. PURCHASE MODAL — replace description with attributes
================================================================

Screen: PurchaseModal shows the full card text
("На глубине:zov. На поверхности:zor. При всплытии —...")

Replace the description with a compact attributes summary:
  - Редкость:    Редкая / Обычная / Необычная / Ультраредкая
  - Стихия:      Болота / Леса / Горы / Степи / Тьма / Нейтрал
  - Фойл:        Да / — (только если true)
  - Нуар:        Да / — (только если true)
  - Летающий:    Да / — (только если true)
  - HP:          N
  - Удар:        X-X-X
  - Стоимость:   N · Рядовая/Элитная
  - Класс:       ... (если есть)

Show only the fields that exist. Skip empty.
Keep: image, name, price, balance, buttons.

================================================================
4. DETAIL PAGE — buy button inline
================================================================

Screen: /shop/item/...

Add a "Купить" button on the detail page itself, near the
price + stock line.

Behavior:
  - Same purchase flow as the shop modal.
  - If already owned: show "В инвентаре" (disabled).
  - If stock = 0: show "Нет в наличии" (disabled).
  - After successful purchase: update state, show "В инвентаре",
    decrement stock locally.
  - Do NOT navigate away.

================================================================
PROCESS
================================================================

Order: 1 → 2 → 3 → 4.

Show plan first, no code. Wait for OK.

DO NOT TOUCH: seed, purchase backend, game mechanics, nav,
admin, attribute-filters wave (that's next after this).
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

A–H visual + UX polish ✅ (verify: в целом ок, 4 пункта полиша)

Текущий полиш:

🟧 Detail hero — white corners fix

🟧 Shop card — remove card text

🟧 Purchase modal — attributes вместо текста

🟧 Detail — inline buy button

Следующая волна:

🟧 Attribute filters (element / rarity / artist / foil / stats / icons) — proberserk-style

🟧 Bottleneck fix (server-side page / smaller DTO)

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

