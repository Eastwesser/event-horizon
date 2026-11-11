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

PurchaseItemWithStock не трогаем — он уже в правильном порядке.

Полная замена RefundPurchase
Копируй целиком, с комментами. Остальной файл не трогай.

go
// RefundPurchase marks the latest completed purchase as REFUNDED, removes the
// inventory row, and restores shop.items stock. Idempotent via refunded_at.
// Enforces 7-day refundable_until and blocks physical items after fulfilled_at.
//
// P0 deadlock fix (10.11.2026): lock order must match PurchaseItemWithStock —
// items → inventory → purchases. Before this fix RefundPurchase locked
// purchases first (FOR UPDATE OF p) and items last, while PurchaseItemWithStock
// locked items first — concurrent buy+cancel on the same item deadlocked (40P01).
// The category join is now done with a plain SELECT (no FOR UPDATE on items
// through the join), so the explicit items lock below is the single anchor.
func (r *PostgresShopRepo) RefundPurchase(ctx context.Context, userID, itemID string, ignoreFulfilled bool) (*RefundResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 2s lock_timeout: if a deadlock somehow still happens, fail fast (40P01)
	// instead of hanging for 30s behind a gateway timeout.
	if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '2000ms'`); err != nil {
		return nil, fmt.Errorf("set lock_timeout: %w", err)
	}

	// 1) Anchor on items FIRST — same first lock as PurchaseItemWithStock.
	//    FOR UPDATE serialises buy+cancel on the same item before touching
	//    inventory/purchases, which removes the cross-order deadlock.
	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
		return nil, fmt.Errorf("lock item: %w", err)
	}

	// 2) Read the candidate purchase. No JOIN here — category is fetched
	//    separately below so this SELECT does not implicitly lock items
	//    through the join (which made the old lock order plan-dependent).
	var purchaseID string
	var price int
	var refundableUntil time.Time
	var fulfilledAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT id, price, refundable_until, fulfilled_at
		FROM purchases
		WHERE user_id = $1 AND item_id = $2
		  AND status = 'COMPLETED' AND refunded_at IS NULL
		ORDER BY COALESCE(completed_at, purchased_at) DESC
		LIMIT 1
		FOR UPDATE
	`, userID, itemID).Scan(&purchaseID, &price, &refundableUntil, &fulfilledAt)
	if err == sql.ErrNoRows {
		var alreadyID string
		var alreadyPrice int
		qerr := tx.QueryRowContext(ctx, `
			SELECT id, price FROM purchases
			WHERE user_id = $1 AND item_id = $2
			  AND (status = 'REFUNDED' OR refunded_at IS NOT NULL)
			ORDER BY refunded_at DESC NULLS LAST
			LIMIT 1
		`, userID, itemID).Scan(&alreadyID, &alreadyPrice)
		if qerr == nil {
			return &RefundResult{PurchaseID: alreadyID, Price: alreadyPrice, AlreadyRefunded: true}, nil
		}
		return nil, fmt.Errorf("purchase not found")
	}
	if err != nil {
		return nil, err
	}

	// 3) Category for the physical-item gate. Plain SELECT — no lock needed,
	//    items is already anchored above.
	var category string
	if err := tx.QueryRowContext(ctx, `SELECT category FROM items WHERE id = $1`, itemID).Scan(&category); err != nil {
		return nil, fmt.Errorf("load item category: %w", err)
	}

	now := time.Now().UTC()
	if now.After(refundableUntil) {
		return nil, model.ErrRefundWindowExpired
	}
	if fulfilledAt.Valid && model.IsPhysicalItem(category) && !ignoreFulfilled {
		return nil, model.ErrAlreadyFulfilled
	}

	// 4) Close the purchase row (we already hold its lock from step 2).
	res, err := tx.ExecContext(ctx, `
		UPDATE purchases
		SET status = 'REFUNDED', refunded_at = NOW()
		WHERE id = $1 AND status = 'COMPLETED' AND refunded_at IS NULL
	`, purchaseID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &RefundResult{PurchaseID: purchaseID, Price: price, AlreadyRefunded: true}, nil
	}

	// 5) Drop the inventory row (second lock, after items + purchases).
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inventory WHERE user_id = $1 AND item_id = $2
	`, userID, itemID); err != nil {
		return nil, err
	}

	// 6) Restore stock (items was already locked in step 1).
	if _, err := tx.ExecContext(ctx, `
		UPDATE items SET stock = stock + 1, version = version + 1 WHERE id = $1
	`, itemID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &RefundResult{PurchaseID: purchaseID, Price: price}, nil
}
Что изменилось (чтобы ты видел, а не верил на слово)
SET LOCAL lock_timeout = '2000ms' — в начале. Если deadlock всё же будет (меж-item), fail через 2s, а не 30s.

SELECT 1 FROM items WHERE id = $1 FOR UPDATE — якорь, первый лок. Теперь purchase и cancel начинают с одного места.

Убран JOIN items из запроса purchases — category читается отдельным SELECT (без FOR UPDATE, items уже залочен якорем). Старый JOIN давал неявный shared lock на items в порядке, зависящем от плана → непредсказуемо.

FOR UPDATE OF p → FOR UPDATE — p больше нет (нет JOIN), лок на purchases явный.

DELETE inventory — после UPDATE purchases. Раньше было UPDATE purchases → DELETE inventory → UPDATE items. Теперь: items (якорь) → purchases (SELECT FOR UPDATE + UPDATE) → inventory (DELETE) → items (UPDATE stock). Порядок совпадает с purchase: items → inventory → purchases? Стоп.

Проверю ещё раз порядок:

PurchaseItemWithStock:

items FOR UPDATE (строка ~453)

items UPDATE stock (строка ~461)

inventory INSERT (строка ~471)

purchases INSERT (строка ~476)

outbox INSERT

Мой RefundPurchase:

items FOR UPDATE (якорь)

purchases SELECT FOR UPDATE

purchases UPDATE

inventory DELETE

items UPDATE stock

Расхождение: purchase = items → inventory → purchases. cancel = items → purchases → inventory. Всё ещё разный порядок между inventory и purchases.

Но — якорь items FOR UPDATE сериализует оба метода на одном item. Пока purchase(itemX) держит items[X], cancel(itemX) ждёт на items[X]. Они никогда не выполняются одновременно на одном item. Значит, inventory и purchases для одного item не пересекаются. Deadlock невозможен для одного item.

Меж-item deadlock: purchase(itemA) держит items[A], ждёт inventory[A]; cancel(itemB) держит items[B], ждёт inventory[B]. Разные строки inventory (PK user_id, item_id), разные строки items (PK id). Нет пересечения. Deadlock невозможен.

Единственный теоретический риск — purchases с разными id, но одинаковым item_id (история покупок одного item разными юзерами). purchase(itemX, user1) вставляет purchases (новый id), cancel(itemX, user2) обновляет purchases (старый id). Разные строки purchases. Нет пересечения.

Вывод: якорь items FOR UPDATE достаточен. Порядок inventory/purchases после якоря не важен, потому что якорь сериализует всё на одном item. lock_timeout — страховка.

Но чтобы совсем было чисто (и на будущее, если кто-то добавит лок на items в purchase без якоря), лучше выровнять порядок до конца. Меняю cancel на: items (якорь) → inventory DELETE → purchases UPDATE → items UPDATE stock. Тогда полностью совпадает с purchase: items → inventory → purchases.

Но проблема: cancel должен прочитать purchases, чтобы получить purchaseID и price. Если сначала DELETE inventory, а purchases ещё не прочитан — мы не знаем, что удалять. Значит, purchases SELECT должен быть до inventory DELETE.

Компромисс: items (якорь) → purchases SELECT FOR UPDATE → inventory DELETE → purchases UPDATE → items UPDATE. Это мой текущий патч. inventory и purchases переставлены относительно purchase, но якорь это компенсирует.

Альтернатива (полный порядок): items (якорь) → purchases SELECT (без FOR UPDATE, просто read) → inventory DELETE → purchases UPDATE ... WHERE id=... (берёт лок на purchases здесь) → items UPDATE. Тогда порядок: items → inventory → purchases → items. Совпадает с purchase до purchases. Но purchases SELECT без FOR UPDATE — read не берёт лок, и между read и UPDATE может вклиниться другая транзакция. Для cancel это ок, потому что WHERE id = $1 AND status = 'COMPLETED' AND refunded_at IS NULL + RowsAffected() == 0 → idempotent (AlreadyRefunded). То есть read-then-update с проверкой RowsAffected — безопасно.

Я выберу этот вариант — он полностью выравнивает порядок. Патч обновлённый:

go
// RefundPurchase marks the latest completed purchase as REFUNDED, removes the
// inventory row, and restores shop.items stock. Idempotent via refunded_at.
// Enforces 7-day refundable_until and blocks physical items after fulfilled_at.
//
// P0 deadlock fix (10.11.2026): lock order now matches PurchaseItemWithStock —
// items → inventory → purchases. Before this fix RefundPurchase locked
// purchases first (FOR UPDATE OF p) and items last, while PurchaseItemWithStock
// locked items first — concurrent buy+cancel on the same item deadlocked (40P01).
//
// The purchases SELECT is a plain read (no FOR UPDATE): the actual row lock is
// taken by the UPDATE below, and RowsAffected()==0 still means "already refunded",
// so idempotency is preserved. The items anchor (FOR UPDATE) serialises buy and
// cancel on the same item before they touch inventory/purchases at all.
func (r *PostgresShopRepo) RefundPurchase(ctx context.Context, userID, itemID string, ignoreFulfilled bool) (*RefundResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 2s lock_timeout: if a deadlock somehow still happens, fail fast (40P01)
	// instead of hanging for 30s behind a gateway timeout.
	if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '2000ms'`); err != nil {
		return nil, fmt.Errorf("set lock_timeout: %w", err)
	}

	// 1) Anchor on items FIRST — same first lock as PurchaseItemWithStock.
	//    This serialises buy+cancel on the same item before inventory/purchases
	//    are touched, which removes the cross-order deadlock.
	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM items WHERE id = $1 FOR UPDATE`, itemID); err != nil {
		return nil, fmt.Errorf("lock item: %w", err)
	}

	// 2) Read the candidate purchase (plain read — the row lock happens in the
	//    UPDATE below, keeping the lock order items → inventory → purchases).
	var purchaseID string
	var price int
	var refundableUntil time.Time
	var fulfilledAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT id, price, refundable_until, fulfilled_at
		FROM purchases
		WHERE user_id = $1 AND item_id = $2
		  AND status = 'COMPLETED' AND refunded_at IS NULL
		ORDER BY COALESCE(completed_at, purchased_at) DESC
		LIMIT 1
	`, userID, itemID).Scan(&purchaseID, &price, &refundableUntil, &fulfilledAt)
	if err == sql.ErrNoRows {
		var alreadyID string
		var alreadyPrice int
		qerr := tx.QueryRowContext(ctx, `
			SELECT id, price FROM purchases
			WHERE user_id = $1 AND item_id = $2
			  AND (status = 'REFUNDED' OR refunded_at IS NOT NULL)
			ORDER BY refunded_at DESC NULLS LAST
			LIMIT 1
		`, userID, itemID).Scan(&alreadyID, &alreadyPrice)
		if qerr == nil {
			return &RefundResult{PurchaseID: alreadyID, Price: alreadyPrice, AlreadyRefunded: true}, nil
		}
		return nil, fmt.Errorf("purchase not found")
	}
	if err != nil {
		return nil, err
	}

	// 3) Category for the physical-item gate. Plain SELECT — items is already
	//    anchored above, no extra lock needed.
	var category string
	if err := tx.QueryRowContext(ctx, `SELECT category FROM items WHERE id = $1`, itemID).Scan(&category); err != nil {
		return nil, fmt.Errorf("load item category: %w", err)
	}

	now := time.Now().UTC()
	if now.After(refundableUntil) {
		return nil, model.ErrRefundWindowExpired
	}
	if fulfilledAt.Valid && model.IsPhysicalItem(category) && !ignoreFulfilled {
		return nil, model.ErrAlreadyFulfilled
	}

	// 4) Drop the inventory row (second lock, after items — matches purchase).
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inventory WHERE user_id = $1 AND item_id = $2
	`, userID, itemID); err != nil {
		return nil, err
	}

	// 5) Close the purchase row. Row lock is taken here (third lock, matches
	//    purchase: items → inventory → purchases). RowsAffected()==0 means
	//    another tx already refunded it — keep idempotency.
	res, err := tx.ExecContext(ctx, `
		UPDATE purchases
		SET status = 'REFUNDED', refunded_at = NOW()
		WHERE id = $1 AND status = 'COMPLETED' AND refunded_at IS NULL
	`, purchaseID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &RefundResult{PurchaseID: purchaseID, Price: price, AlreadyRefunded: true}, nil
	}

	// 6) Restore stock (items was already locked in step 1).
	if _, err := tx.ExecContext(ctx, `
		UPDATE items SET stock = stock + 1, version = version + 1 WHERE id = $1
	`, itemID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &RefundResult{PurchaseID: purchaseID, Price: price}, nil
}
Итог
Файл: /home/denismatveev/event_horizon/services/shop/internal/repository/postgres_repo.go

Функция: RefundPurchase — заменить целиком (строки 278–344).

Не трогать: PurchaseItemWithStock (уже правильный), PurchaseItem (легаси, не используется в саге), всё остальное.

Порядок блокировок после фикса:

Метод	Порядок
PurchaseItemWithStock	items FOR UPDATE → items UPDATE stock → inventory INSERT → purchases INSERT → outbox INSERT
RefundPurchase	items FOR UPDATE (якорь) → inventory DELETE → purchases UPDATE → items UPDATE stock
Совпадает: items → inventory → purchases. Якорь items FOR UPDATE сериализует buy/cancel на одном item. Deadlock невозможен.

Страховка: lock_timeout = 2s — если что-то меж-item всё же сцепится, fail через 2s с понятной ошибкой, а не 30s зависание.