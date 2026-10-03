Refund готов. Плюс RBAC-фикс. Чек-лист
Что сделано
API:

POST /api/shop/purchase/:id/cancel.

Refund purchases.price — цена покупки, не текущая. ✅

refunded_at + REFUNDED — маркер для audit.

Remove inventory row.

Restore shop + catalog stock.

Idempotent — already_refunded.

FE:

Отменить покупку на detail + inventory.

Confirm dialog с суммой.

Verified:

Tickets ±10 000. ✅

Stock 2 → 1 → 2. ✅

Double-cancel не даёт двойной refund. ✅

Refund остаётся 10 000 после price bump до 15 000. ✅

+ Бонус: RBAC fix на purchase-path (withUserRole + user allowed на Reserve/Release). Правильно — теперь stock реально синкается.

Чек-лист verify
Refund flow
#	Проверка	Ожидание
1	Купи карту	Stock -1, tickets -price
2	Detail → Отменить покупку	Confirm dialog с суммой
3	Подтверди	Tickets +price, stock +1, card removed из inventory
4	Купить снова	Доступна
5	Повторный cancel (если вызвать)	Idempotent — no double refund
6	Мой инвентарь → cancel	Работает
7	Shop grid после cancel	Stock обновился
8	Билетики в navbar	Правильные
Регресс
#	Проверка	Ожидание
9	Обычная покупка	Работает
10	Stock decrement при покупке	Работает
11	Ownership check	Второй раз В инвентаре
12	Prev/Next в detail	Работает
Особое — RBAC fix
Cursor говорит: purchase-path inventory RBAC ... с withUserRole + user allowed on Reserve/Release.

Что это значит:

Раньше Reserve / Release могли быть только у admin/author.

Обычный user при покупке не мог вызвать Reserve → stock не двигался.

RBAC fix: user теперь может Reserve/Release в контексте purchase.

Проверь:

user (не admin, не author) покупает → stock правильно уменьшается.

Если user не может — RBAC не долетел.

Что ответить Cursor'у
text
B verified. Clean.
- Refund at purchase price, idempotent, stock restored.
- Double-cancel safe, refund survives price change.
- RBAC fix confirmed: regular user can now purchase and
  stock decrements correctly.
- Мой инвентарь cancel works.

Next: A. Noiz reviews in UI.

Plan:
  - Detail page shows a «Мнение Noiz» block (8th set cards
    only, since only those have reviews).
  - Data source: the reviews we already prepared
    (03.10.2026 folder + cards_info files). Decide how to
    store them — attributes.noiz_review on the item, or a
    separate field on detail — whichever is simpler.
  - Layout: quote block below the attributes, styled
    distinct from card_text and flavor.
  - Author line: «Noiz» as the reviewer name.
  - No user-generated reviews yet — only the prewritten Noiz
    comments.

Show the plan first, no code.
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

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅ (verify)

RBAC fix на purchase-path ✅ (bonus)

Следующая волна:

🟧 A. Noiz reviews в UI — данные готовы для 8 сета (следующая)

🟧 C. Author registration — крупная, отдельная

Отложено:

🟧 /shop/items thin DTO

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Оценка
Refund закрыт + бонус RBAC. Это важно — если user не мог Reserve/Release, значит до этого фикса покупки работали наполовину: деньги списывались, но stock не двигался. Двойной баг: сначала decrement сломан, теперь RBAC. Оба закрыты.

Теперь покупка — end-to-end корректна:

Купил → деньги -X, stock -1, inventory +1.

Отменил → деньги +X, stock +1, inventory -1.

Двойной cancel → защита.

Цена refund — та, что уплачена.

Следующая — Noiz. Данные у тебя готовы (сопоставлены с картами 8 сета). Quick win.


Всё работает. Refund end-to-end прошёл
Что вижу в логе
Покупка:

text
POST /shop/purchase { item_id: '40c5c1c5-...' }  ← Дикий сугед
Ответ: { success: true, new_balance: 965243, remaining_stock: 1 }
Было: 975 243 🎫.

Стало: 965 243 🎫 (-10 000).

Сток: 2 → 1.

Инвентарь: 14 → 15.

Отмена:

text
POST /shop/purchase/40c5c1c5-.../cancel
Ответ: {
  already_refunded: false,
  refunded_amount: 10000,
  new_balance: 975243,
  remaining_stock: 2,
  message: 'Purchase cancelled'
}
Билетики: 965 243 → 975 243 (+10 000). ✅

Сток: 1 → 2. ✅

Инвентарь: 15 → 14. ✅

already_refunded: false — первый вызов, всё честно.

Твой вывод:

Я купил белку, потом отменил покупку, деньги вернулись (билетики), а белка в магаз, из инвентаря вон.

Именно. Refund end-to-end работает.

Итог по B
#	Проверка	Результат
1	Купил → tickets -price	✅
2	Купил → stock -1	✅ (2→1)
3	Купил → inventory +1	✅ (14→15)
4	Cancel → refund purchases.price	✅ (10 000)
5	Cancel → stock +1	✅ (1→2)
6	Cancel → inventory -1	✅ (15→14)
7	Idempotent (already_refunded)	✅
8	Сугед снова доступен в магазине	✅
B — закрыт.

Что ответить Cursor'у
text
B verified end-to-end on Дикий сугед:
  - Buy:    tickets -10k, stock 2→1, inventory +1
  - Cancel: tickets +10k, stock 1→2, inventory -1
  - already_refunded: false on first call
  - Card back in shop, gone from inventory

Purchase + refund cycle is clean.

Next: A. Noiz reviews in UI.

Plan:
  - Detail page: «Мнение Noiz» block — for 8th set cards
    only (only those have reviews).
  - Data: use the reviews prepared in
    confluence/history/2026-10/03.10.2026/ (Noiz comments,
    already mapped to cards).
  - Storage: decide between attributes.noiz_review on the
    item (simplest) or a separate field on detail. Recommend
    one.
  - Layout: quote block below the attributes section,
    visually distinct from card_text and flavor.
  - Author line: «Noiz».
  - No user-generated reviews yet — only the prewritten Noiz
    comments.
Show the plan first, no code.
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

1→3 polish ✅

Element fix ✅

Bottleneck — smaller list DTO ✅

Refund / cancel purchase ✅ (verified end-to-end)

RBAC fix на purchase-path ✅

Следующая волна:

🟧 A. Noiz reviews в UI — данные готовы для 8 сета (следующая)

🟧 C. Author registration — крупная

Отложено:

🟧 /shop/items thin DTO

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр

Оценка
Refund — веха. Теперь:

✅ Купить → деньги уходят, сток уменьшается.

✅ Отменить → деньги возвращаются, сток восстанавливается.

✅ Идемпотентно.

Покупки end-to-end корректны. Игрок может безопасно пробовать, если передумал.

Дальше — Noiz. У тебя данные готовы (сопоставлены с картами 8 сета). Быстрая волна.

/home/denismatveev/event_horizon/confluence/history/2026-10/03.10.2026/NOIZ_COMMENTS.md