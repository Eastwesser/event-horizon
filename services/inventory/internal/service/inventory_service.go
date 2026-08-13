package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/nats-io/nats.go"
)

// InventoryService — бизнес-логика сервиса инвентаря.
// Кеширование — в repository.CachedRepository (decorator над InventoryRepository).
type InventoryService struct {
	repo   repository.InventoryRepository
	outbox repository.ItemOutboxWriter // PostgreSQL outbox; nil for MongoDB
	js     nats.JetStreamContext
}

// NewInventoryService создает новый сервис инвентаря.
func NewInventoryService(
	repo repository.InventoryRepository,
	outbox repository.ItemOutboxWriter,
	js nats.JetStreamContext,
) *InventoryService {
	return &InventoryService{
		repo:   repo,
		outbox: outbox,
		js:     js,
	}
}

// GetItem получает товар по ID (cache-aside в CachedRepository).
func (s *InventoryService) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.repo.GetItem(ctx, id)
}

// CreateItem создает товар. Для PostgreSQL использует Outbox, для MongoDB — прямой Insert.
func (s *InventoryService) CreateItem(ctx context.Context, item *model.Item) error {
	if item.AuthorID == "" {
		return fmt.Errorf("author_id is required")
	}
	if item.Type == "" {
		return fmt.Errorf("type is required")
	}
	if item.Name == "" {
		return fmt.Errorf("name is required")
	}
	if item.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if item.Stock < 0 {
		return fmt.Errorf("stock cannot be negative")
	}

	if s.outbox != nil {
		event := map[string]interface{}{
			"event":      "item.created",
			"item_id":    "",
			"author_id":  item.AuthorID,
			"type":       item.Type,
			"name":       item.Name,
			"price":      item.Price,
			"stock":      item.Stock,
			"attributes": item.Attributes,
			"images":     item.Images,
			"timestamp":  time.Now().Unix(),
		}
		eventPayload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		return s.outbox.CreateItemWithOutbox(ctx, item, "inventory.item.created", eventPayload)
	}

	return s.repo.CreateItem(ctx, item)
}

// UpdateItem обновляет товар.
func (s *InventoryService) UpdateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		return fmt.Errorf("id is required")
	}
	if item.Version <= 0 {
		cur, err := s.repo.GetItem(ctx, item.ID)
		if err != nil {
			return err
		}
		item.Version = cur.Version
	}
	return s.repo.UpdateItem(ctx, item)
}

// DeleteItem удаляет товар.
func (s *InventoryService) DeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.DeleteItem(ctx, id)
}

// SearchItems выполняет поиск (кеш — в CachedRepository).
func (s *InventoryService) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, filters, limit, offset)
}

// GetByAuthor возвращает товары автора (с пагинацией).
func (s *InventoryService) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, limit, offset)
}

// GetByType возвращает товары по типу (с пагинацией).
func (s *InventoryService) GetByType(ctx context.Context, itemType string, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, map[string]interface{}{"type": itemType}, limit, offset)
}

// BulkCreateItems - массовое создание товаров
func (s *InventoryService) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if item.AuthorID == "" {
			return fmt.Errorf("author_id is required for all items")
		}
		if item.Type == "" {
			return fmt.Errorf("type is required for all items")
		}
		if item.Name == "" {
			return fmt.Errorf("name is required for all items")
		}
		if item.Price < 0 {
			return fmt.Errorf("price cannot be negative")
		}
		if item.Stock < 0 {
			return fmt.Errorf("stock cannot be negative")
		}
	}

	return s.repo.BulkCreateItems(ctx, items)
}

// ReserveItem - резервирование товара (уменьшение stock)
func (s *InventoryService) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	if id == "" {
		return 0, fmt.Errorf("id is required")
	}
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	return s.repo.ReserveItem(ctx, id, quantity)
}

// SoftDeleteItem - мягкое удаление
func (s *InventoryService) SoftDeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.SoftDeleteItem(ctx, id)
}

// RestoreItem - восстановление после мягкого удаления
func (s *InventoryService) RestoreItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.RestoreItem(ctx, id)
}

// GetStats - статистика по товарам
func (s *InventoryService) GetStats(ctx context.Context) (*model.Stats, error) {
	return s.repo.GetStats(ctx)
}
