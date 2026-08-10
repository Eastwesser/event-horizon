package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

type PostgresRepo struct {
    db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
    return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateOutboxEvent(ctx context.Context, eventType string, payload []byte) error {
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
    `, eventType, payload)
    return err
}

// CreateItemWithOutbox — сохраняет товар и событие в outbox в одной транзакции
func (r *PostgresRepo) CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if item.ID == "" {
        item.ID = uuid.New().String()
    }
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()

    attrsJSON, err := json.Marshal(item.Attributes)
    if err != nil {
        return fmt.Errorf("marshal attributes: %w", err)
    }

    // 1. Сохраняем товар
    query := `
        INSERT INTO inventory_items (
            id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `
    _, err = tx.ExecContext(ctx, query,
        item.ID, item.AuthorID, item.Type, item.Name, item.Description,
        item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
        item.CreatedAt, item.UpdatedAt,
    )
    if err != nil {
        return err
    }

    // 👇 2. ОБНОВЛЯЕМ eventPayload с правильным item.ID!
    var event map[string]interface{}
    if err := json.Unmarshal(eventPayload, &event); err != nil {
        return err
    }
    event["item_id"] = item.ID  // ← вставляем ID!
    updatedPayload, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // 3. Сохраняем обновлённое событие в outbox
    _, err = tx.ExecContext(ctx, `
        INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
    `, eventType, updatedPayload)
    if err != nil {
        return err
    }

    return tx.Commit()
}

// CreateItem — устаревший метод, используй CreateItemWithOutbox
func (r *PostgresRepo) CreateItem(ctx context.Context, item *model.Item) error {
    // Оставляем для обратной совместимости, но лучше использовать CreateItemWithOutbox
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if item.ID == "" {
        item.ID = uuid.New().String()
    }
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()

    attrsJSON, err := json.Marshal(item.Attributes)
    if err != nil {
        return fmt.Errorf("marshal attributes: %w", err)
    }

    query := `
        INSERT INTO inventory_items (
            id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `
    _, err = tx.ExecContext(ctx, query,
        item.ID, item.AuthorID, item.Type, item.Name, item.Description,
        item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
        item.CreatedAt, item.UpdatedAt,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}

// GetItem — получить товар по ID
func (r *PostgresRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
    query := `
        SELECT id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        FROM inventory_items WHERE id = $1
    `
    var item model.Item
    var attrsJSON []byte
    var images []string

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
        &item.Price, &item.Stock, &attrsJSON, pq.Array(&images),
        &item.CreatedAt, &item.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }

    if len(attrsJSON) > 0 {
        if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
            return nil, fmt.Errorf("unmarshal attributes: %w", err)
        }
    }
    item.Images = images
    return &item, nil
}

func (r *PostgresRepo) UpdateItem(ctx context.Context, item *model.Item) error {
    item.UpdatedAt = time.Now()

    attrsJSON, err := json.Marshal(item.Attributes)
    if err != nil {
        return fmt.Errorf("marshal attributes: %w", err)
    }

    query := `
        UPDATE inventory_items SET
            author_id = $1, type = $2, name = $3, description = $4,
            price = $5, stock = $6, attributes = $7, images = $8, updated_at = $9
        WHERE id = $10
    `
    _, err = r.db.ExecContext(ctx, query,
        item.AuthorID, item.Type, item.Name, item.Description,
        item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
        item.UpdatedAt, item.ID,
    )
    return err
}

func (r *PostgresRepo) DeleteItem(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM inventory_items WHERE id = $1", id)
    return err
}

func (r *PostgresRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
    var conditions []string
    var args []interface{}
    argCount := 1

    if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
        conditions = append(conditions, fmt.Sprintf("author_id = $%d", argCount))
        args = append(args, authorID)
        argCount++
    }
    if itemType, ok := filters["type"].(string); ok && itemType != "" {
        conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
        args = append(args, itemType)
        argCount++
    }
    if priceMin, ok := filters["price_min"].(float64); ok && priceMin > 0 {
        conditions = append(conditions, fmt.Sprintf("price >= $%d", argCount))
        args = append(args, priceMin)
        argCount++
    }
    if priceMax, ok := filters["price_max"].(float64); ok && priceMax > 0 {
        conditions = append(conditions, fmt.Sprintf("price <= $%d", argCount))
        args = append(args, priceMax)
        argCount++
    }

    if attrs, ok := filters["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
        attrsJSON, err := json.Marshal(attrs)
        if err == nil {
            conditions = append(conditions, fmt.Sprintf("attributes @> $%d", argCount))
            args = append(args, string(attrsJSON))
            argCount++
        }
    }

    if query, ok := filters["query"].(string); ok && query != "" {
        conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
        searchTerm := "%" + query + "%"
        args = append(args, searchTerm)
        argCount++
    }

    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }

    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_items %s", whereClause)
    var total int64
    if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    query := fmt.Sprintf(`
        SELECT id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        FROM inventory_items %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereClause, argCount, argCount+1)
    args = append(args, limit, offset)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var items []*model.Item
    for rows.Next() {
        var item model.Item
        var attrsJSON []byte
        var images []string

        err := rows.Scan(
            &item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
            &item.Price, &item.Stock, &attrsJSON, pq.Array(&images),
            &item.CreatedAt, &item.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }

        if len(attrsJSON) > 0 {
            if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
                return nil, 0, err
            }
        }
        item.Images = images
        items = append(items, &item)
    }

    return items, total, nil
}

func (r *PostgresRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}

func (r *PostgresRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}
