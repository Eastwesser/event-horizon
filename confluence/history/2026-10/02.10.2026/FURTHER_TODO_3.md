Что вижу — отлично
/inventory (скрин 1)
✅ Фильтр Карточка появился

✅ 289 товаров (280 карт + 9 старых)

✅ Картинки карт грузятся — Дикий сугед, Спригган, Тантид-лучник, Эльф-рейдер

✅ Цена 10 000 (билетики, не ₽)

✅ В наличии: 2 / 3 / 1 — сток

✅ Кнопки Ред. / Удалить — внутри карточки (не вылезают)

/shop (скрин 2)
✅ Фильтр Карточки (новый chip)

✅ Загружено 289 из 289 — пагинация за кулисами работает

✅ Картинки, цена, сток

✅ Купить кнопки

/shop/item/:id (скрин 3 — деталь)
✅ Детальная страница открывается

✅ Большая картинка карты

✅ карточка badge + цена + сток

✅ C Компаньон — badge работает!

✅ Card text, flavor text

✅ Художник: feather_7

✅ Характеристики — таблица атрибутов: artist, attack_dice, card_no, class, companion, cost, cost_tier, element, flavor_text, flying, foil, hp, icons, market_rub, move, noir, parasite, rarity, set_name, set_number, symbiont, type_main, type_sub, unique, year

Это огромный прогресс. Детальная страница — реальная карточка с полными данными.

Покупка (скрин 4)
✅ Модалка Подтверждение покупки

✅ Превью карты, название, текст

✅ Цена: 10 000 / Ваш баланс: 1 000 243

✅ Кнопки Отмена / Да, купить

Что уже работает без 4–5
Я вижу то, что Cursor в 4–5 запланировал:

✅ Badge C Компаньон — работает

✅ Детальная страница с attributes — работает (правда, в виде таблицы key: value)

✅ Картинки — работают везде

Но — иконки (armor, zoal, regen) пока как текст (icons: — — пусто в этой карте).

Что осталось на 4–5
Судя по скрину detail:

Attributes dump — сырой. artist: feather_7 / attack_dice: 1-1-2 / card_no: 75 / ... — таблица key: value. Для карточки нужно:

Сгруппировать (Combat / Card info / Author / Flags).

Оформить — не key: value, а человеческое отображение.

Icons — компоненты, не icons: —.

Badges частично. C Компаньон — есть. Foil / noir / flying — где? В detail видно foil: нет, flying: нет — как текст, не badge.

Ticket price — где-то переключено, где-то нет. Здесь правильно (🎫 10000).

Create modal ×2000 — не проверено.

Отчёт для Cursor'а
Approve с продолжением:

text
1–3 verified — images render, detail opens, stock shows.

BUT: I can already see badge «C Компаньон» on the detail page.
So item 5 (badges) is partially done. And the attributes table
in the detail page shows raw key: value pairs — that's what
item 4 (icons/structured attributes) needs to fix.

Proceed with 4–5 as planned:

4. ICONS + STRUCTURED ATTRIBUTES
   Detail page currently dumps all attributes as a key: value
   table: artist, attack_dice, card_no, class, companion,
   cost, cost_tier, element, flavor_text, flying, foil, hp,
   icons, market_rub, move, noir, parasite, rarity, set_name,
   set_number, symbiont, type_main, type_sub, unique, year.

   This is the raw dump item 4 was meant to fix. Replace with:
     - Icons component: armor:2 → shield + "2"; zoal → arrow
       + wings; regen:1 → crab + "1"; zov/zoz/zot/zor/zoo/zom,
       direct/uchr/ova/ovz/ovs/stamina.
     - Group the fields into sections:
         * Combat:      hp, move, attack_dice, attack_type,
                        icons, cost, cost_tier
         * Card info:   set_name, set_number, year, card_no,
                        rarity, element, class, type_main,
                        type_sub
         * Flags:       foil, noir, flying, companion,
                        symbiont, parasite, unique
         * Author:      artist, artist_display
     - Remove raw key: value dump for attributes that already
       have a dedicated UI element (badges, icons, price, stock).

5. BADGES
   Keep the existing «C Компаньон» badge. Add the rest:
     - foil      → «✦ ФОЙЛ»   (gold)
     - noir      → «◐ НУАР»   (silver)
     - flying    → «🕊 ЛЕТАЮЩИЙ»
     - companion → already done
     - unique    → «♛ УНИКАЛЬНАЯ»
   Use the same style/location as «C Компаньон».

6. TICKET PRICE — already looks correct on card/detail/modal.
   Confirm and finish: any remaining ₽ label under cards?
   (Only if inventory price label wasn't fully switched.)

7. CREATE MODAL — verify and fix ×2000 → ×1000.

Then pagination, sorting, artist pages.

After 4–5, I'll verify — and we do 6–7, 8, 9, 10 in order.

DO NOT TOUCH: seed, non-card types, game mechanics,
nav/footer, admin, purchase flow.
⚠️ Один момент — покупка
Модалка Да, купить работает. Но Cursor сам предупреждал:

purchase for inventory cards may fail until a later wiring.

Проверь:

Нажми Да, купить → карта появляется в инвентаре?

Баланс списывается?

Или ошибка?

Если работает — отлично. Если нет — это отдельная волна (Cursor предупреждал).

Сейчас не блокер — display/browse закончены.

Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно

Волна UI карточек — 1–3 ✅:

✅ Image rendering везде

✅ Click → detail (полная страница с данными)

✅ Stock (В наличии: N)

✅ Pagination (загружено 289 из 289)

✅ Badge C Компаньон (бонус)

Следующее — 4–5:

🟧 Icons как компоненты (armor:2, zoal, regen:1)

🟧 Группировка attributes в detail (Combat / Card / Flags / Author)

🟧 Badges: foil / noir / flying / unique

Потом — 6–10:

🟧 Ticket price (проверить остатки)

🟧 Create modal ×1000

🟧 Pagination UI (сейчас hidden — 100/стр)

🟧 Sorting

🟧 Artist pages

Следующая волна:

🟧 Attribute filters

🟧 Расширенный поиск

🟧 Noiz-комментарии

🟧 Author registration

🟧 Покупка карточек (проверить, может не работать)

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр 


UPDATED!!!

Покупка работает! Но есть баг со стоком
Что вижу
✅ Хорошо
Покупка прошла. success: true, new_balance: 990243. 1 000 000 − 10 000 = 990 000 — правильно.

Дикий сугед теперь В инвентаре — на карточке зелёный badge.

Мой инвентарь (14) — был 13, стал 14. +1 item.

В самом инвентаре — Дикий сугед, Куплено: 03.10.2026. Запись создана.

Кнопка Купить для купленного → В инвентаре (disabled).

🔴 Баг — сток не уменьшается
Я купил сугеда, но «в наличии 2» всё равно осталось.

Ожидание: после покупки 1 из 2 → должно стать В наличии: 1.
Реальность: В наличии: 2 — как было.

Это критично. Иначе:

Можно купить бесконечное количество одной карты.

Карта stock: 1 (уникальная) не станет sold-out.

Нарушается принцип «No FOMO» (sold-out есть sold-out).

Root cause — что смотреть
Варианты:

A. Backend не decrement'ит stock.

/shop/purchase создаёт inventory record, но item.stock не трогает.

Самый вероятный.

B. Backend decrement'ит, но фронт кэширует.

/shop/items возвращает свежий сток, но UI не перезапрашивает после покупки.

Менее вероятно — вижу, что /shop/items вызывается повторно после покупки.

C. Идемпотентность — можно купить один и тот же item дважды.

Проверить: нажми Купить на Сугед снова — списываются ли ещё 10 000?

Если да → backend не проверяет «уже куплено».

Скорее всего — A. Или A + C.

Что делать — «тестовый item»
Хорошая идея — обкатать на Сугед:

Проверить отмену покупки (refund + stock +1).

Проверить покупку снова — можно или нельзя.

Проверить stock decrement на бэкенде.

Но сначала — фикс самого бага. Иначе нечего тестировать.

Промпт для Cursor
text
Purchase works (funds deducted, inventory record created,
"В инвентаре" badge shown). But two issues:

================================================================
1. STOCK NOT DECREMENTED ON PURCHASE (critical)
================================================================

Test: bought «Дикий сугед» (stock was 2). After purchase:
  - Balance: 1,000,000 → 990,243 ✅
  - Inventory record created ✅
  - BUT item.stock still shows "В наличии: 2" (should be 1)

This breaks the "sold-out is sold-out" rule — a stock-1 card
can be bought infinitely.

Investigate:
  - Does POST /shop/purchase decrement inventory item.stock
    on the backend? Where?
  - Is the shop response re-fetched after purchase, or is the
    client showing a stale cache?
  - Check the DB directly: what is stock for the «Дикий сугед»
    item right now?

Expected behavior after fix:
  - Purchase decrements stock by 1.
  - When stock reaches 0, the card shows "Нет в наличии" and
    the "Купить" button is disabled.
  - Frontend re-fetches item data after purchase (or updates
    the item in the store optimistically + refetch).

================================================================
2. IDEMPOTENCY OF PURCHASE (check)
================================================================

Question: if I click "Купить" twice on the same card (before
any stock check), does the backend:
  (a) reject the second purchase ("already owned"), OR
  (b) allow it and charge again?

For stock-1 cards this matters. Report the actual behavior.

================================================================
3. TEST ITEM — «Дикий сугед»
================================================================

Keep «Дикий сугед» (item id 40c5c1c5-0973-41e4-9112-4bf0d43582a8)
as the test item for:
  - refund / cancel purchase
  - re-purchase after refund
  - stock decrement / increment
Do NOT delete it. I'll use it for manual testing.

DO NOT TOUCH: seed, other cards, purchase flow beyond the
stock fix, game mechanics, nav, admin.
Что проверить тебе — прямо сейчас
Три теста на Сугед:

Нажми Купить ещё раз на Дикий сугед.

Кнопка Купить или В инвентаре? Если Купить — можно купить дважды → баг.

Если списались ещё 10 000 → idempotency баг.

Открой /shop/item/40c5c1c5-... (деталь Сугеда).

Что показывает В наличии: N?

2 или 1?

Открой /inventory → фильтр Карточка.

Сугед там? Сколько stock?

Результат — скинь в следующее сообщение. Это поможет Cursor'у точно понять, где баг.

Что я держу в карте
Закрыто:

Waves A–D, twins, empty state, image URL, CI lint

3 игры, ККИ-тип, burger, DTOs

280 карт засеяно, idempotency verified

Волна UI карточек:

✅ Image rendering везде

✅ Click → detail

✅ Stock (В наличии: N) — отображение работает

✅ Pagination (289 из 289)

✅ Badge C Компаньон, В инвентаре

✅ Покупка работает (funds deducted, record created)

🔴 НО: stock не decrement'ится — критичный баг

🟧 Icons + badges (4–5 в процессе)

🟧 Ticket price (проверка)

🟧 Create modal ×1000

🟧 Pagination UI

🟧 Sorting

🟧 Artist pages

Следующая волна:

🟧 Attribute filters

🟧 Расширенный поиск

🟧 Noiz-комментарии

🟧 Author registration

🟧 Refund / cancel purchase (обкатка на Сугед)

🟧 Idempotency покупки (проверить)

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр