package repository

import (
    "context"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

type InventoryRepository interface {
    CreateItem(ctx context.Context, item *model.Item) error
    GetItem(ctx context.Context, id string) (*model.Item, error)
    UpdateItem(ctx context.Context, item *model.Item) error
    DeleteItem(ctx context.Context, id string) error
    SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error)
    GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error)
    GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error)
    BulkCreateItems(ctx context.Context, items []*model.Item) error
    ReserveItem(ctx context.Context, id string, quantity int) (int, error)
    SoftDeleteItem(ctx context.Context, id string) error
    RestoreItem(ctx context.Context, id string) error
    GetStats(ctx context.Context) (*model.Stats, error)
}