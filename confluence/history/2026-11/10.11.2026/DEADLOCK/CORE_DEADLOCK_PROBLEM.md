Корневая причина (подтверждена)
Метод	Порядок блокировок
PurchaseItemWithStock (строка ~445)	items (FOR UPDATE) → inventory (INSERT) → purchases (INSERT) → outbox (INSERT)
RefundPurchase (строка ~278)	purchases (FOR UPDATE OF p) → inventory (DELETE) → items (UPDATE)
Вот оно. purchase = items → inventory → purchases. cancel = purchases → inventory → items. Противоположный порядок по всем трём таблицам. При одновременном purchase(itemX) и cancel(itemX) — deadlock на inventory и purchases/items.

Плюс RefundPurchase использует JOIN items i ON i.id = p.item_id с FOR UPDATE OF p — Postgres может взять shared lock на items (из JOIN) до FOR UPDATE OF p, и порядок становится зависимым от плана — ещё хуже.

Фикс: единый порядок items → inventory → purchases
PurchaseItemWithStock — уже в этом порядке (items FOR UPDATE → inventory → purchases). Не трогаем.

RefundPurchase — надо переставить:

Сначала SELECT ... FROM items WHERE id = $1 FOR UPDATE (якорь).

Потом SELECT ... FROM purchases ... FOR UPDATE OF p (без JOIN items — category берём отдельно, до лока purchases, или через LEFT JOIN без FOR UPDATE на items).

Потом UPDATE purchases.

Потом DELETE inventory.

Потом UPDATE items (stock + 1).

Плюс SET LOCAL lock_timeout = '2000ms' в обоих — страховка.

Что заменить
Один файл: /home/denismatveev/event_horizon/services/shop/internal/repository/postgres_repo.go

Одна функция: RefundPurchase (строки 278–344, целиком).