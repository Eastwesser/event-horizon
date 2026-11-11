Что именно чиним
Смотрю postgres_repo.go ещё раз, чтобы дать точный патч, а не «примерно».

Purchase (строки 156–169) — текущий порядок
go
tx, err := r.db.BeginTx(ctx, nil)                                    // 156
_, err = tx.ExecContext(ctx, `INSERT INTO inventory ...`)             // 163  ← ЛОК 1: inventory
_, err = tx.ExecContext(ctx, `INSERT INTO purchases ...`)             // 169  ← ЛОК 2: purchases
Cancel (строки 278–344) — текущий порядок
go
tx, err := r.db.BeginTx(ctx, nil)                                    // 278
err = tx.QueryRowContext(ctx, `... FOR UPDATE OF p ...`)              // 289-297 ← ЛОК 1: purchases (p)
qerr := tx.QueryRowContext(ctx, `...`)                                // 302       ← ЛОК 2: ?
res, err := tx.ExecContext(ctx, `...`)                                // 326       ← ЛОК 3: inventory
Проблема: purchase = inventory → purchases. cancel = purchases → inventory. Противоположный порядок → deadlock.

Патч
Есть три способа. Я предлагаю комбинированный (самый надёжный):

В cancel — выровнять порядок: сначала inventory (или items), потом purchases. Но там FOR UPDATE OF p в JOIN — надо аккуратно.

В purchase — добавить SELECT ... FROM items WHERE id = $1 FOR UPDATE до INSERT INTO inventory. Тогда оба метода начинают с items — общий якорь, deadlock исчезает.

В обоих — SET LOCAL lock_timeout = '2s' — если deadlock всё же случится, fail через 2s, а не 30s.

Патч для purchase (строки 156–169)
go
// PurchaseItemWithStock — обёртка, чтобы не менять сигнатуру
func (r *PostgresRepo) PurchaseItemWithStock(
    ctx context.Context,
    userID, itemID, purchaseID string,
    price int,
    records []OutboxRecord,
) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // P0 deadlock fix: общий якорь с CancelPurchase — сначала лок на items.
    // Без этого purchase идёт inventory → purchases, cancel — purchases → inventory,
    // и при одновременной покупке/отмене одной пары (user_id, item_id) получаем 40P01.
    if _, err := tx.ExecContext(ctx, `SELECT 1 FROM items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
        return fmt.Errorf("lock item: %w", err)
    }

    // 2s lock_timeout: если deadlock всё же случится — fail быстро, а не 30s.
    if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '2000ms'`); err != nil {
        return fmt.Errorf("set lock_timeout: %w", err)
    }

    // Дальше — как было, порядок уже безопасен (items → inventory → purchases).
    _, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
    if err != nil {
        return err
    }
    _, err = tx.ExecContext(ctx, `INSERT INTO purchases (user_id, item_id, price) VALUES ($1, $2, $3)`, userID, itemID, price)
    if err != nil {
        return err
    }
    // ... outbox, stock decrement, commit — как было
    return tx.Commit()
}
Важно: items — таблица в shop БД (eventhorizon_shop), не в inventory. Смотрю миграции: items есть в 20260711000000_init_shop_schema.sql. Значит, SELECT 1 FROM items WHERE id = $1 FOR UPDATE — корректно для shop-БД. inventory и purchases тоже там. Хорошо.

Патч для cancel (строки 278–344)
Тут сложнее, потому что cancel использует FOR UPDATE OF p в JOIN. Мне нужно видеть сам запрос (строки 289–302), чтобы дать точный патч. Но по grep вижу:

text
289: err = tx.QueryRowContext(ctx, `...`      ← начало запроса
297:         FOR UPDATE OF p                  ← конец запроса
302: qerr := tx.QueryRowContext(ctx, `...`    ← второй запрос
326: res, err := tx.ExecContext(ctx, `...`)   ← UPDATE inventory
Патч для cancel (структурно, без точного текста запроса):

go
func (r *PostgresRepo) RefundPurchase(
    ctx context.Context,
    userID, itemID string,
    ignoreFulfilled bool,
) (*RefundResult, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // P0 deadlock fix: тот же якорь, что в PurchaseItemWithStock.
    // Сначала items → потом purchases. Порядок совпадает с purchase.
    if _, err := tx.ExecContext(ctx, `SELECT 1 FROM items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
        return nil, fmt.Errorf("lock item: %w", err)
    }
    if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '2000ms'`); err != nil {
        return nil, fmt.Errorf("set lock_timeout: %w", err)
    }

    // Дальше — существующий запрос с FOR UPDATE OF p.
    // Порядок блокировок теперь: items (якорь) → purchases (p) → inventory (UPDATE).
    // Совпадает с purchase: items → inventory → purchases? 
    // НЕТ. Purchase: items → inventory → purchases. Cancel: items → purchases → inventory.
    // Всё ещё расходится на inventory/purchases.
    // ...
}
Стоп. Даже с якорем items порядок не совпадает:

purchase: items → inventory → purchases

cancel: items → purchases → inventory

Якорь items сериализует доступ к одному item, поэтому два concurrent-запроса на один item не могут выполняться одновременно — второй ждёт на items FOR UPDATE. Это уже убивает deadlock для одного item.

Но deadlock может быть меж-item: purchase(itemA) держит items[A], ждёт inventory[A]; cancel(itemB) держит items[B], ждёт inventory[B]; и если inventory — одна таблица, а purchases — другая, то inventory и purchases — разные ресурсы, и deadlock между ними возможен только если одна транзакция держит inventory, ждёт purchases, а другая — наоборот. С якорем items это невозможно для одного item. Для разных items — возможно, если inventory и purchases — общие таблицы (они общие).

Но inventory и purchases — разные таблицы. purchase вставляет в обе. cancel обновляет purchases (status) и inventory (delete). Deadlock между ними — только если одна и та же строка в inventory/purchases затрагивается. А inventory — PK (user_id, item_id), purchases — PK id. Разные строки.

Вывод: якорь items FOR UPDATE достаточен для устранения deadlock на одном item. Меж-item deadlock — маловероятен (разные строки в разных таблицах), но lock_timeout = 2s — страховка.

Однако есть тонкость: cancel в строке 302 делает второй QueryRowContext — если это SELECT ... FROM inventory WHERE ... FOR UPDATE, то порядок: items → purchases → inventory. А purchase: items → inventory → purchases. Для одного item — сериализовано якорем. Для разных — оба могут держать items[X] и items[Y], и пытаться взять inventory/purchases в разном порядке → deadlock возможен.

Самый надёжный фикс: в cancel переставить inventory (DELETE/UPDATE) до purchases (UPDATE status). Тогда оба метода: items → inventory → purchases. Единый порядок. Это то, что я предлагаю.