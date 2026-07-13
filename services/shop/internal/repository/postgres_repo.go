package repository

import (
    "context"
    "database/sql"
    "fmt"
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
        SELECT i.id, i.name, i.description, i.price, i.category, i.game_id, i.image_url, i.available
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
        err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price,
            &item.Category, &item.GameID, &item.ImageURL, &item.Available)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }
    return items, nil
}
