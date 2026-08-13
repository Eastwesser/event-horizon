package repository_test

import (
	"context"
	"testing"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	item *model.Item
}

func (s *stubRepo) CreateItem(context.Context, *model.Item) error { return nil }
func (s *stubRepo) GetItem(_ context.Context, id string) (*model.Item, error) {
	if s.item != nil && s.item.ID == id {
		return s.item, nil
	}
	return nil, context.Canceled
}
func (s *stubRepo) UpdateItem(context.Context, *model.Item) error   { return nil }
func (s *stubRepo) DeleteItem(context.Context, string) error        { return nil }
func (s *stubRepo) SearchItems(context.Context, map[string]interface{}, int, int) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) GetByAuthor(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) GetByType(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) BulkCreateItems(context.Context, []*model.Item) error { return nil }
func (s *stubRepo) ReserveItem(context.Context, string, int) (int, error) {
	return 0, nil
}
func (s *stubRepo) SoftDeleteItem(context.Context, string) error { return nil }
func (s *stubRepo) RestoreItem(context.Context, string) error    { return nil }
func (s *stubRepo) GetStats(context.Context) (*model.Stats, error) {
	return &model.Stats{}, nil
}

func TestNewCachedRepository_nilCachePassthrough(t *testing.T) {
	base := &stubRepo{item: &model.Item{ID: "1", Name: "x"}}
	repo := repository.NewCachedRepository(base, nil)
	require.Equal(t, base, repo)
}

func TestCachedRepository_implementsInterface(t *testing.T) {
	base := &stubRepo{item: &model.Item{ID: "1", Name: "x"}}
	var _ repository.InventoryRepository = repository.NewCachedRepository(base, nil)
}
