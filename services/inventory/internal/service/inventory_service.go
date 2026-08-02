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

type InventoryService struct {
    repo  repository.InventoryRepository
    cache *repository.RedisCacheRepo
    js    nats.JetStreamContext
}

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

func (s *InventoryService) GetItem(ctx context.Context, id string) (*model.Item, error) {
    if id == "" {
        return nil, fmt.Errorf("id is required")
    }

    if s.cache != nil {
        if item, err := s.cache.GetItem(ctx, id); err == nil {
            return item, nil
        }
    }

    item, err := s.repo.GetItem(ctx, id)
    if err != nil {
        return nil, err
    }

    if s.cache != nil && item != nil {
        _ = s.cache.SetItem(ctx, item)
    }

    return item, nil
}

// CreateItem — создаёт товар и сохраняет событие в outbox
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

    // Формируем событие
    event := map[string]interface{}{
        "event":      "item.created",
        "item_id":    item.ID,
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

    // Приводим repo к PostgresRepo для доступа к CreateItemWithOutbox
    pgRepo, ok := s.repo.(*repository.PostgresRepo)
    if !ok {
        return fmt.Errorf("repository is not PostgresRepo, cannot use outbox")
    }

    if err := pgRepo.CreateItemWithOutbox(ctx, item, "inventory.item.created", eventPayload); err != nil {
        return err
    }

    // Сохраняем в кеш
    if s.cache != nil {
        _ = s.cache.SetItem(ctx, item)
    }

    return nil
}

func (s *InventoryService) UpdateItem(ctx context.Context, item *model.Item) error {
    if item.ID == "" {
        return fmt.Errorf("id is required")
    }

    if err := s.repo.UpdateItem(ctx, item); err != nil {
        return err
    }

    if s.cache != nil {
        _ = s.cache.SetItem(ctx, item)
    }

    return nil
}

func (s *InventoryService) DeleteItem(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("id is required")
    }

    if err := s.repo.DeleteItem(ctx, id); err != nil {
        return err
    }

    if s.cache != nil {
        _ = s.cache.DeleteItem(ctx, id)
    }

    return nil
}

func (s *InventoryService) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
    return s.repo.SearchItems(ctx, filters, limit, offset)
}

func (s *InventoryService) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*model.Item, int64, error) {
    return s.repo.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, limit, offset)
}

func (s *InventoryService) GetByType(ctx context.Context, itemType string, limit, offset int) ([]*model.Item, int64, error) {
    return s.repo.SearchItems(ctx, map[string]interface{}{"type": itemType}, limit, offset)
}
