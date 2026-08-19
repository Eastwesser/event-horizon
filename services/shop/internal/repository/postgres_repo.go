package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Item struct {
	ID          string
	Name        string
	Description string
	Price       int
	Category    string
	GameID      *string
	ImageURL    string
	Available   bool
	Owned       bool
	PurchasedAt *time.Time
}

type PostgresShopRepo struct {
	db *sql.DB
}

func NewPostgresShopRepo(db *sql.DB) *PostgresShopRepo {
	return &PostgresShopRepo{db: db}
}

func (r *PostgresShopRepo) GetItems(ctx context.Context, category, gameID string) ([]Item, error) {
	query := `SELECT id, name, description, price, category, game_id, image_url, available FROM items WHERE available = true`
	args := []interface{}{}
	argIdx := 1

	if category != "" && category != "all" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	if gameID != "" {
		query += fmt.Sprintf(" AND (game_id = $%d OR game_id IS NULL)", argIdx)
		args = append(args, gameID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var gameID sql.NullString
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Category,
			&gameID,
			&item.ImageURL,
			&item.Available,
		)
		if err != nil {
			return nil, err
		}
		if gameID.Valid {
			item.GameID = &gameID.String
		}
		items = append(items, item)
	}
	return items, nil
}

// CreateItemFromInventory создаёт товар из события инвентаря
func (r *PostgresShopRepo) CreateItemFromInventory(ctx context.Context, itemID, name, description string, price float64) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO items (id, name, description, price, category, game_id, image_url, available)
        VALUES ($1, $2, $3, $4, 'merch', '', '', true)
        ON CONFLICT (id) DO NOTHING
    `, itemID, name, description, int(price))
	return err
}

func (r *PostgresShopRepo) GetItemByID(ctx context.Context, itemID string) (*Item, error) {
	query := `SELECT id, name, description, price, category, game_id, image_url, available FROM items WHERE id = $1`
	var item Item
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.Name, &item.Description, &item.Price,
		&item.Category, &item.GameID, &item.ImageURL, &item.Available,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PostgresShopRepo) IsItemOwned(ctx context.Context, userID, itemID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM inventory WHERE user_id = $1 AND item_id = $2)`
	err := r.db.QueryRowContext(ctx, query, userID, itemID).Scan(&exists)
	return exists, err
}

func (r *PostgresShopRepo) PurchaseItem(ctx context.Context, userID, itemID string, price int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем в инвентарь
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		return err
	}

	// Добавляем в историю покупок
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases (user_id, item_id, price) VALUES ($1, $2, $3)`, userID, itemID, price)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresShopRepo) GetUserInventory(ctx context.Context, userID string) ([]Item, error) {
	query := `
        SELECT i.id, i.name, i.description, i.price, i.category, i.game_id, i.image_url, i.available, inv.purchased_at
        FROM inventory inv
        JOIN items i ON inv.item_id = i.id
        WHERE inv.user_id = $1
        ORDER BY inv.purchased_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var gameID sql.NullString
		var purchasedAt time.Time
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Category,
			&gameID,
			&item.ImageURL,
			&item.Available,
			&purchasedAt,
		)
		if err != nil {
			return nil, err
		}
		if gameID.Valid {
			item.GameID = &gameID.String
		}
		item.PurchasedAt = &purchasedAt
		items = append(items, item)
	}
	return items, nil
}

func (r *PostgresShopRepo) CreatePendingPurchase(ctx context.Context, userID, itemID string, price int) (string, error) {
	var purchaseID string
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO purchases (user_id, item_id, price, status) 
        VALUES ($1, $2, $3, 'PENDING') 
        RETURNING id
    `, userID, itemID, price).Scan(&purchaseID)
	return purchaseID, err
}

func (r *PostgresShopRepo) CompletePurchase(ctx context.Context, purchaseID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE purchases SET status = 'COMPLETED', completed_at = NOW() 
        WHERE id = $1 AND status = 'PENDING'
    `, purchaseID)
	return err
}

func (r *PostgresShopRepo) CancelPurchase(ctx context.Context, purchaseID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE purchases SET status = 'CANCELLED' 
        WHERE id = $1 AND status = 'PENDING'
    `, purchaseID)
	return err
}

type OutboxRecord struct {
	EventType string
	Payload   []byte
}

// PurchaseItemWithStock — покупка с проверкой стока. If outbox is set, the event
// is written in the same TX (worker publishes to NATS).
func (r *PostgresShopRepo) PurchaseItemWithStock(ctx context.Context, userID, itemID string, price int, outbox *OutboxRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Pessimistic row lock + optimistic version bump on stock change.
	var stock, version int
	err = tx.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&stock, &version)
	if err != nil {
		return err
	}
	if stock <= 0 {
		return fmt.Errorf("item out of stock")
	}

	res, err := tx.ExecContext(ctx, `
        UPDATE items SET stock = stock - 1, version = version + 1 WHERE id = $1 AND version = $2`, itemID, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("optimistic lock conflict on item %s", itemID)
	}

	// Добавляем в инвентарь
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		return err
	}

	// Добавляем в историю покупок
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases (user_id, item_id, price, status) VALUES ($1, $2, $3, 'COMPLETED')`, userID, itemID, price)
	if err != nil {
		return err
	}

	if outbox != nil && outbox.EventType != "" && len(outbox.Payload) > 0 {
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
			outbox.EventType, outbox.Payload); err != nil {
			return err
		}
	}

	return tx.Commit()
}
