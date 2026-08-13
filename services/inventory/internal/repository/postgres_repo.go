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

// PostgresRepo — адаптер для PostgreSQL, реализующий интерфейс InventoryRepository.
// Использует JSONB для гибких атрибутов и поддерживает транзакции с Outbox.
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo создает новый репозиторий PostgreSQL.
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// ---------------------- Вспомогательные функции ----------------------

// toPostgresFilter преобразует map фильтров в SQL условия и аргументы.
func toPostgresFilter(filters map[string]interface{}) ([]string, []interface{}, int) {
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

	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		placeholders := make([]string, len(types))
		for i := range types {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, types[i])
			argCount++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ", ")))
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

	conditions = append(conditions, "deleted_at IS NULL")

	return conditions, args, argCount
}

// ---------------------- CRUD-методы (реализация интерфейса) ----------------------

// CreateItem создает новый товар в PostgreSQL.
func (r *PostgresRepo) CreateItem(ctx context.Context, item *model.Item) error {
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
	_, err = r.db.ExecContext(ctx, query,
		item.ID, item.AuthorID, item.Type, item.Name, item.Description,
		item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

// GetItem возвращает товар по ID (игнорирует удаленные).
func (r *PostgresRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
	query := `
		SELECT id, author_id, type, name, description, price, stock, COALESCE(version,1), attributes, images, created_at, updated_at
		FROM inventory_items WHERE id = $1 AND deleted_at IS NULL
	`
	var item model.Item
	var attrsJSON []byte
	var images []string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
		&item.Price, &item.Stock, &item.Version, &attrsJSON, pq.Array(&images),
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrItemNotFound
		}
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

// UpdateItem обновляет все поля товара (стратегия PUT) with optimistic locking.
func (r *PostgresRepo) UpdateItem(ctx context.Context, item *model.Item) error {
	item.UpdatedAt = time.Now()

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	expected := item.Version
	if expected <= 0 {
		expected = 1
	}
	query := `
		UPDATE inventory_items SET
			author_id = $1, type = $2, name = $3, description = $4,
			price = $5, stock = $6, attributes = $7, images = $8, updated_at = $9,
			version = version + 1
		WHERE id = $10 AND deleted_at IS NULL AND version = $11
	`
	res, err := r.db.ExecContext(ctx, query,
		item.AuthorID, item.Type, item.Name, item.Description,
		item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
		item.UpdatedAt, item.ID, expected,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrVersionConflict
	}
	item.Version = expected + 1
	return nil
}

// DeleteItem выполняет жесткое удаление товара.
func (r *PostgresRepo) DeleteItem(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM inventory_items WHERE id = $1", id)
	return err
}

// SoftDeleteItem выполняет мягкое удаление.
func (r *PostgresRepo) SoftDeleteItem(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE inventory_items SET deleted_at = NOW() WHERE id = $1
	`, id)
	return err
}

// SearchItems выполняет поиск с фильтрами.
func (r *PostgresRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	conditions, args, argCount := toPostgresFilter(filters)

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
		SELECT id, author_id, type, name, description, price, stock, COALESCE(version,1), attributes, images, created_at, updated_at
		FROM inventory_items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	paginationArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, paginationArgs...)
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
			&item.Price, &item.Stock, &item.Version, &attrsJSON, pq.Array(&images),
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

// BulkCreateItems — массовая вставка товаров для начальной загрузки или импорта.
// Использует транзакцию для атомарности.
func (r *PostgresRepo) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO inventory_items (
			id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		attrsJSON, err := json.Marshal(item.Attributes)
		if err != nil {
			return fmt.Errorf("marshal attributes: %w", err)
		}

		_, err = tx.ExecContext(ctx, query,
			item.ID, item.AuthorID, item.Type, item.Name, item.Description,
			item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
			item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ReserveItem — atomic stock decrement with optimistic version bump (single conditional UPDATE).
func (r *PostgresRepo) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	var remaining int
	err := r.db.QueryRowContext(ctx, `
		UPDATE inventory_items
		SET stock = stock - $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND stock >= $1
		RETURNING stock
	`, quantity, id).Scan(&remaining)
	if err != nil {
		if err == sql.ErrNoRows {
			var exists bool
			_ = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE id = $1 AND deleted_at IS NULL)`, id).Scan(&exists)
			if !exists {
				return 0, model.ErrItemNotFound
			}
			return 0, model.ErrNotEnoughStock
		}
		return 0, err
	}
	return remaining, nil
}

// RestoreItem — восстанавливает мягко удаленный товар.
func (r *PostgresRepo) RestoreItem(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE inventory_items SET deleted_at = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrItemNotFound
	}

	return nil
}

// GetStats — возвращает статистику по товарам.
func (r *PostgresRepo) GetStats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{
		ByType:   make(map[string]int64),
		ByAuthor: make(map[string]int64),
	}

	// Общее количество
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM inventory_items WHERE deleted_at IS NULL
	`).Scan(&stats.TotalItems)
	if err != nil {
		return nil, err
	}

	// Группировка по типу
	rows, err := r.db.QueryContext(ctx, `
		SELECT type, COUNT(*) FROM inventory_items
		WHERE deleted_at IS NULL
		GROUP BY type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var typ string
		var count int64
		if err := rows.Scan(&typ, &count); err != nil {
			return nil, err
		}
		stats.ByType[typ] = count
	}

	// Группировка по автору (топ 10)
	rows, err = r.db.QueryContext(ctx, `
		SELECT author_id, COUNT(*) FROM inventory_items
		WHERE deleted_at IS NULL
		GROUP BY author_id
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var authorID string
		var count int64
		if err := rows.Scan(&authorID, &count); err != nil {
			return nil, err
		}
		stats.ByAuthor[authorID] = count
	}

	return stats, nil
}

// GetByAuthor возвращает все товары автора (с пагинацией по умолчанию 1000).
func (r *PostgresRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}

// GetByType возвращает все товары указанного типа (с пагинацией по умолчанию 1000).
func (r *PostgresRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}

// ---------------------- Outbox методы (только для PostgreSQL) ----------------------

// CreateOutboxEvent создает событие в таблице outbox.
func (r *PostgresRepo) CreateOutboxEvent(ctx context.Context, eventType string, payload []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
	`, eventType, payload)
	return err
}

// CreateItemWithOutbox — сохраняет товар и событие в outbox в одной транзакции.
// Этот метод НЕ входит в интерфейс InventoryRepository, он специфичен для PostgreSQL.
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

	var event map[string]interface{}
	if err := json.Unmarshal(eventPayload, &event); err != nil {
		return err
	}
	event["item_id"] = item.ID
	updatedPayload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
	`, eventType, updatedPayload)
	if err != nil {
		return err
	}

	return tx.Commit()
}