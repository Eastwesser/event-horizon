package repository

import (
	"context"
	"fmt"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

// CachedRepository decorates InventoryRepository with Redis cache-aside (Kozirev W6 pattern).
// RedisCacheRepo remains the low-level storage adapter; this type is the repository decorator.
type CachedRepository struct {
	next  InventoryRepository
	cache *RedisCacheRepo
}

// NewCachedRepository wraps repo with Redis caching. Pass nil cache to disable caching.
func NewCachedRepository(next InventoryRepository, cache *RedisCacheRepo) InventoryRepository {
	if cache == nil || next == nil {
		return next
	}
	return &CachedRepository{next: next, cache: cache}
}

func (c *CachedRepository) CreateItem(ctx context.Context, item *model.Item) error {
	if err := c.next.CreateItem(ctx, item); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

// CreateItemWithOutbox delegates to the wrapped repo when it supports outbox, then warms cache.
func (c *CachedRepository) CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error {
	w, ok := c.next.(ItemOutboxWriter)
	if !ok {
		return fmt.Errorf("underlying repository does not support outbox")
	}
	if err := w.CreateItemWithOutbox(ctx, item, eventType, eventPayload); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if item, err := c.cache.GetItem(ctx, id); err == nil {
		return item, nil
	}
	item, err := c.next.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if item != nil {
		_ = c.cache.SetItem(ctx, item)
	}
	return item, nil
}

func (c *CachedRepository) UpdateItem(ctx context.Context, item *model.Item) error {
	if err := c.next.UpdateItem(ctx, item); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) DeleteItem(ctx context.Context, id string) error {
	if err := c.next.DeleteItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	queryKey, _ := filters["query"].(string)
	if queryKey != "" && len(filters) == 1 {
		if cached, err := c.cache.GetSearchResult(ctx, queryKey, limit, offset); err == nil {
			return cached, int64(len(cached)), nil
		}
	}

	items, total, err := c.next.SearchItems(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if queryKey != "" && len(filters) == 1 {
		_ = c.cache.SetSearchResult(ctx, queryKey, limit, offset, items)
	}
	return items, total, nil
}

func (c *CachedRepository) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
	return c.next.GetByAuthor(ctx, authorID)
}

func (c *CachedRepository) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
	return c.next.GetByType(ctx, itemType)
}

func (c *CachedRepository) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if err := c.next.BulkCreateItems(ctx, items); err != nil {
		return err
	}
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	remaining, err := c.next.ReserveItem(ctx, id, quantity)
	if err != nil {
		return 0, err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return remaining, nil
}

func (c *CachedRepository) SoftDeleteItem(ctx context.Context, id string) error {
	if err := c.next.SoftDeleteItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) RestoreItem(ctx context.Context, id string) error {
	if err := c.next.RestoreItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) GetStats(ctx context.Context) (*model.Stats, error) {
	return c.next.GetStats(ctx)
}

func (c *CachedRepository) warmItem(ctx context.Context, item *model.Item) {
	if item != nil {
		_ = c.cache.SetItem(ctx, item)
	}
}

func (c *CachedRepository) invalidateSearch(ctx context.Context) {
	_ = c.cache.InvalidateSearchCache(ctx)
}
