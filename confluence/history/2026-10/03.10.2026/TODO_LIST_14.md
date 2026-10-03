ОК. Шип B — Refund / cancel purchase
Что подтверждаю
API:

POST /api/shop/purchase/:id/cancel (или /refund) под JWT.

Restore tickets (refund).

Return inventory stock +1.

Remove owned row.

Idempotent — повторный вызов не делает двойной refund.

FE:

Owned shop detail → кнопка Отменить покупку.

My inventory → тоже кнопка.

Confirm dialog → success → Купить снова, stock обновлён.

Scope:

Только карты / merch через текущий purchase path.

No admin UI, no partial refunds, no games.

Out of scope: A / C / /shop/items thin DTO.

Мой ответ Cursor'у
text
OK — ship B (Refund / cancel purchase).

Two clarifications:

1. Confirm modal
   Include a clear text: «Вернуть X билетиков? Карта уйдёт из
   инвентаря, товар снова станет доступен.» with the actual
   ticket amount pulled from the purchase record.

2. Refund timing / audit
   - Do we have a purchase record (shop purchase table) with
     price paid? Use that price for the refund — not the current
     item price (which may have changed).
   - Store a refund marker / status so we can audit later
     (refunded_at timestamp), even if we don't build admin UI yet.

Everything else as written.

DO NOT TOUCH: Noiz, author registration, /shop/items thin DTO,
seed, games, nav, admin panel.
Что важно проверить после реализации
Тест на Сугед:

#	Проверка	Ожидание
1	Купи Сугед (если уже куплен — сначала cancel)	Сток -1, билетики -10 000
2	На detail Сугед → Отменить покупку	Confirm dialog
3	Подтверди	Билетики +10 000, сток +1, карта исчезает из инвентаря
4	Купить снова	Доступна
5	Повторный cancel (если как-то вызвать)	Idempotent — не списывает/не возвращает дважды
6	Цена при покупке ≠ текущая	Refund по той цене, что была при покупке
7	Inventory my-list	Сугед ушёл после refund
8	Shop grid	Сток обновился
Особое внимание #5 и #6 — это два места, где refund часто ломается:

Двойной refund = инфляция.

Refund по текущей цене ≠ уплаченной = эксплойт (купить дешевле, продать дороже).

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

Element fix ✅ (56 карт)

Bottleneck — smaller list DTO ✅ (-26%)

Текущая волна (B — approve):

🟧 API POST /api/shop/purchase/:id/cancel

🟧 Refund по цене покупки (не текущей)

🟧 Idempotent (refunded_at marker)

🟧 FE: Отменить покупку в detail + inventory

🟧 Confirm dialog с суммой

🟧 Возврат stock +1

🟧 Remove from inventory

Следующие волны (по порядку):

🟧 A. Noiz reviews в UI (данные готовы для 8 сета)

🟧 C. Author registration

Отложено:

🟧 /shop/items thin DTO

Ещё дальше:

🟧 Эмодзи → SVG

🟧 Полиш игр