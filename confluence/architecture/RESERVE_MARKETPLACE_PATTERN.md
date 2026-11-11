// // PurchaseItemWithStock — покупка с проверкой стока. purchaseID is the row id
// // and the PurchaseUUID of every outbox event written in this transaction.
// func (r *PostgresShopRepo) PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []OutboxRecord) error {
// 	tx, err := r.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	// Pessimistic row lock + optimistic version bump on stock change.
// 	var stock, version int
// 	err = tx.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&stock, &version)
// 	if err != nil {
// 		return err
// 	}
// 	if stock <= 0 {
// 		return fmt.Errorf("item out of stock")
// 	}

// 	res, err := tx.ExecContext(ctx, `
//         UPDATE items SET stock = stock - 1, version = version + 1 WHERE id = $1 AND version = $2`, itemID, version)
// 	if err != nil {
// 		return err
// 	}
// 	if n, _ := res.RowsAffected(); n == 0 {
// 		return fmt.Errorf("optimistic lock conflict on item %s", itemID)
// 	}

// 	// Добавляем в инвентарь
// 	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = tx.ExecContext(ctx, `
//         INSERT INTO purchases (id, user_id, item_id, price, status, refundable_until)
//         VALUES ($1, $2, $3, $4, 'COMPLETED', NOW() + INTERVAL '7 days')`, purchaseID, userID, itemID, price)
// 	if err != nil {
// 		return err
// 	}

// 	for _, rec := range records {
// 		if rec.EventType == "" || len(rec.Payload) == 0 {
// 			continue
// 		}
// 		if _, err := tx.ExecContext(ctx, `
//             INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
// 			rec.EventType, rec.Payload); err != nil {
// 			return err
// 		}
// 	}

// 	return tx.Commit()
// }

// PurchaseItemWithStock — покупка с проверкой стока.
// P0 (10.11.2026): резерв мгновенный (UPDATE ... WHERE stock >= 1), без
// SELECT FOR UPDATE. Лок на items держится микросекунды, не 2с (Billing
// gRPC внутри лока — это была причина contention на 10 items).
// Order: items (UPDATE) → inventory → purchases → outbox.
// Совпадает с RefundPurchase (items → inventory → purchases).
func (r *PostgresShopRepo) PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []OutboxRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1) Мгновенный резерв: одна строка, атомарно, без SELECT FOR UPDATE.
	var stock, version int
	err = tx.QueryRowContext(ctx, `
		UPDATE items SET stock = stock - 1, version = version + 1
		WHERE id = $1 AND stock >= 1
		RETURNING stock, version
	`, itemID).Scan(&stock, &version)
	if err == sql.ErrNoRows {
		return fmt.Errorf("item out of stock")
	}
	if err != nil {
		return err
	}

	// 2) Inventory row.
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		return err
	}

	// 3) Purchase row.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO purchases (id, user_id, item_id, price, status, refundable_until)
		VALUES ($1, $2, $3, $4, 'COMPLETED', NOW() + INTERVAL '7 days')`, purchaseID, userID, itemID, price)
	if err != nil {
		return err
	}

	// 4) Outbox.
	for _, rec := range records {
		if rec.EventType == "" || len(rec.Payload) == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
			rec.EventType, rec.Payload); err != nil {
			return err
		}
	}

	return tx.Commit()
}



Что изменилось:

Убран SELECT stock, version FROM items WHERE id = $1 FOR UPDATE.

Заменён на UPDATE items SET stock = stock - 1 ... WHERE stock >= 1 RETURNING stock, version.

sql.ErrNoRows → item out of stock (тот же контракт).

version больше не используется для optimistic lock (UPDATE атомарный), но RETURNING оставил — может пригодиться для метрик.

Порядок локов — items → inventory → purchases → outbox (совпадает с RefundPurchase).

defer tx.Rollback() — если ошибка после резерва, откат вернёт stock (автоматически). Компенсация не нужна в этой транзакции — откат делает всё.