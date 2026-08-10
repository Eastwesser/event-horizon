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
// Использует Redis для кеширования: отдельных товаров и результатов поиска.
type InventoryService struct {
	repo  repository.InventoryRepository
	cache *repository.RedisCacheRepo
	js    nats.JetStreamContext
}

// NewInventoryService создает новый сервис инвентаря.
func NewInventoryService(
	repo repository.InventoryRepository,
	cache *repository.RedisCacheRepo,
	js nats.JetStreamContext,
) *InventoryService {
	return &InventoryService{
		repo:  repo,
		cache: cache,
		js:    js,
	}
}

// GetItem получает товар по ID. Сначала проверяет кеш, затем БД.
func (s *InventoryService) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	// 1. Проверяем кеш
	if s.cache != nil {
		if item, err := s.cache.GetItem(ctx, id); err == nil {
			return item, nil
		}
	}

	// 2. Запрашиваем из БД
	item, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в кеш
	if s.cache != nil && item != nil {
		_ = s.cache.SetItem(ctx, item)
	}

	return item, nil
}

// CreateItem создает товар. Для PostgreSQL использует Outbox, для MongoDB — прямой Insert.
func (s *InventoryService) CreateItem(ctx context.Context, item *model.Item) error {
	// Валидация
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

	// Проверяем, является ли репозиторий PostgresRepo (для Outbox)
	pgRepo, isPostgres := s.repo.(*repository.PostgresRepo)

	if isPostgres {
		// Для PostgreSQL используем Outbox
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

		if err := pgRepo.CreateItemWithOutbox(ctx, item, "inventory.item.created", eventPayload); err != nil {
			return err
		}
	} else {
		// Для MongoDB используем обычный Create
		if err := s.repo.CreateItem(ctx, item); err != nil {
			return err
		}
	}

	// Инвалидируем кеш поиска (так как данные изменились)
	if s.cache != nil {
		_ = s.cache.SetItem(ctx, item)
		_ = s.cache.InvalidateSearchCache(ctx)
	}

	return nil
}

// UpdateItem обновляет товар и инвалидирует кеш.
func (s *InventoryService) UpdateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		return fmt.Errorf("id is required")
	}

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.SetItem(ctx, item)
		_ = s.cache.InvalidateSearchCache(ctx)
	}

	return nil
}

// DeleteItem удаляет товар и инвалидирует кеш.
func (s *InventoryService) DeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	if err := s.repo.DeleteItem(ctx, id); err != nil {
		return err
	}

	if s.cache != nil {
		_ = s.cache.DeleteItem(ctx, id)
		_ = s.cache.InvalidateSearchCache(ctx)
	}

	return nil
}

// SearchItems выполняет поиск с кешированием результатов.
func (s *InventoryService) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	// Генерируем ключ кеша на основе фильтров
	queryKey, ok := filters["query"].(string)
	if !ok {
		queryKey = ""
	}

	// Пытаемся получить из кеша (только если есть query и нет сложных фильтров)
	if s.cache != nil && queryKey != "" && len(filters) == 1 {
		if cachedItems, err := s.cache.GetSearchResult(ctx, queryKey, limit, offset); err == nil {
			// Для простоты возвращаем только список, total считаем из длины
			return cachedItems, int64(len(cachedItems)), nil
		}
	}

	// Запрос в БД
	items, total, err := s.repo.SearchItems(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Сохраняем в кеш (только простые запросы)
	if s.cache != nil && queryKey != "" && len(filters) == 1 {
		_ = s.cache.SetSearchResult(ctx, queryKey, limit, offset, items)
	}

	return items, total, nil
}

// GetByAuthor возвращает товары автора (с пагинацией).
func (s *InventoryService) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*model.Item, int64, error) {
    // Используем SearchItems, так как интерфейс не поддерживает limit/offset в GetByAuthor
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
    
    // Валидация
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

    if err := s.repo.BulkCreateItems(ctx, items); err != nil {
        return err
    }

    // Инвалидируем кеш
    if s.cache != nil {
        _ = s.cache.InvalidateSearchCache(ctx)
    }

    return nil
}

// ReserveItem - резервирование товара (уменьшение stock)
func (s *InventoryService) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
    if id == "" {
        return 0, fmt.Errorf("id is required")
    }
    if quantity <= 0 {
        return 0, fmt.Errorf("quantity must be positive")
    }

    remaining, err := s.repo.ReserveItem(ctx, id, quantity)
    if err != nil {
        return 0, err
    }

    // Инвалидируем кеш
    if s.cache != nil {
        _ = s.cache.DeleteItem(ctx, id)
        _ = s.cache.InvalidateSearchCache(ctx)
    }

    return remaining, nil
}

// SoftDeleteItem - мягкое удаление
func (s *InventoryService) SoftDeleteItem(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("id is required")
    }

    if err := s.repo.SoftDeleteItem(ctx, id); err != nil {
        return err
    }

    if s.cache != nil {
        _ = s.cache.DeleteItem(ctx, id)
        _ = s.cache.InvalidateSearchCache(ctx)
    }

    return nil
}

// RestoreItem - восстановление после мягкого удаления
func (s *InventoryService) RestoreItem(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("id is required")
    }

    if err := s.repo.RestoreItem(ctx, id); err != nil {
        return err
    }

    if s.cache != nil {
        _ = s.cache.DeleteItem(ctx, id)
        _ = s.cache.InvalidateSearchCache(ctx)
    }

    return nil
}

// GetStats - статистика по товарам
func (s *InventoryService) GetStats(ctx context.Context) (*model.Stats, error) {
    return s.repo.GetStats(ctx)
}