package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

type memRepo struct {
	items map[string]*model.Item
}

func (m *memRepo) CreateItem(_ context.Context, item *model.Item) error {
	if m.items == nil {
		m.items = map[string]*model.Item{}
	}
	cp := *item
	m.items[item.ID] = &cp
	return nil
}
func (m *memRepo) GetItem(_ context.Context, id string) (*model.Item, error) {
	it, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *it
	return &cp, nil
}
func (m *memRepo) UpdateItem(_ context.Context, item *model.Item) error {
	cp := *item
	m.items[item.ID] = &cp
	return nil
}
func (m *memRepo) DeleteItem(context.Context, string) error { return nil }
func (m *memRepo) SearchItems(context.Context, map[string]interface{}, int, int) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (m *memRepo) GetByAuthor(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (m *memRepo) GetByType(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (m *memRepo) BulkCreateItems(context.Context, []*model.Item) error { return nil }
func (m *memRepo) ReserveItem(context.Context, string, int) (int, error)  { return 0, nil }
func (m *memRepo) SoftDeleteItem(context.Context, string) error            { return nil }
func (m *memRepo) RestoreItem(context.Context, string) error               { return nil }
func (m *memRepo) GetStats(context.Context) (*model.Stats, error)          { return &model.Stats{}, nil }

func TestUpdateItem_LoadsVersionWhenOmitted(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Name: "old", Version: 3, CreatedAt: now, UpdatedAt: now},
	}}
	svc := NewInventoryService(repo, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "i1", Name: "new", Version: 0})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Version != 3 || got.Name != "new" {
		t.Fatalf("%+v", got)
	}
}

func TestUpdateItem_KeepsExplicitVersion(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Name: "old", Version: 3, CreatedAt: now, UpdatedAt: now},
	}}
	svc := NewInventoryService(repo, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "i1", Name: "new", Version: 9})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Version != 9 {
		t.Fatalf("version=%d", got.Version)
	}
}

func TestUpdateItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.UpdateItem(context.Background(), &model.Item{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateItem_LoadVersionRepoError(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "missing", Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	_, err := svc.GetItem(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetItem_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{"i1": {ID: "i1", Name: "n"}}}
	svc := NewInventoryService(repo, nil, nil)
	got, err := svc.GetItem(context.Background(), "i1")
	if err != nil || got.Name != "n" {
		t.Fatalf("got %+v err=%v", got, err)
	}
}

func TestCreateItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	base := &model.Item{AuthorID: "a1", Type: "t", Name: "n", Price: 1, Stock: 1}
	cases := []struct {
		name string
		item *model.Item
	}{
		{"missing_author", func() *model.Item { i := *base; i.AuthorID = ""; return &i }()},
		{"missing_type", func() *model.Item { i := *base; i.Type = ""; return &i }()},
		{"missing_name", func() *model.Item { i := *base; i.Name = ""; return &i }()},
		{"negative_price", func() *model.Item { i := *base; i.Price = -1; return &i }()},
		{"negative_stock", func() *model.Item { i := *base; i.Stock = -1; return &i }()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := svc.CreateItem(context.Background(), tc.item); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreateItem_ViaRepo(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{}}
	svc := NewInventoryService(repo, nil, nil)
	item := &model.Item{ID: "i1", AuthorID: "a1", Type: "t", Name: "n", Price: 10, Stock: 5}
	if err := svc.CreateItem(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Name != "n" {
		t.Fatalf("%+v", got)
	}
}

type stubOutbox struct {
	called bool
}

func (s *stubOutbox) CreateItemWithOutbox(_ context.Context, _ *model.Item, eventType string, _ []byte) error {
	s.called = true
	if eventType != "inventory.item.created" {
		return errors.New("bad event type")
	}
	return nil
}

func TestCreateItem_ViaOutbox(t *testing.T) {
	ob := &stubOutbox{}
	svc := NewInventoryService(&memRepo{}, ob, nil)
	item := &model.Item{AuthorID: "a1", Type: "t", Name: "n", Price: 1, Stock: 1}
	if err := svc.CreateItem(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	if !ob.called {
		t.Fatal("outbox not called")
	}
}

func TestDeleteItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.DeleteItem(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestReserveItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if _, err := svc.ReserveItem(context.Background(), "", 1); err == nil {
		t.Fatal("empty id")
	}
	if _, err := svc.ReserveItem(context.Background(), "i1", 0); err == nil {
		t.Fatal("zero qty")
	}
	if _, err := svc.ReserveItem(context.Background(), "i1", -1); err == nil {
		t.Fatal("negative qty")
	}
}

func TestSoftDeleteAndRestore_RequireID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.SoftDeleteItem(context.Background(), ""); err == nil {
		t.Fatal("soft delete")
	}
	if err := svc.RestoreItem(context.Background(), ""); err == nil {
		t.Fatal("restore")
	}
}

func TestBulkCreateItems_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	ok := &model.Item{AuthorID: "a", Type: "t", Name: "n", Price: 0, Stock: 0}
	cases := []struct {
		name  string
		items []*model.Item
	}{
		{"empty_slice", nil},
		{"missing_author", []*model.Item{{Type: "t", Name: "n"}}},
		{"missing_type", []*model.Item{{AuthorID: "a", Name: "n"}}},
		{"missing_name", []*model.Item{{AuthorID: "a", Type: "t"}}},
		{"negative_price", []*model.Item{{AuthorID: "a", Type: "t", Name: "n", Price: -1}}},
		{"negative_stock", []*model.Item{{AuthorID: "a", Type: "t", Name: "n", Stock: -1}}},
		{"valid", []*model.Item{ok}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.BulkCreateItems(context.Background(), tc.items)
			if tc.name == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if tc.name == "empty_slice" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGetStats_DelegatesToRepo(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	stats, err := svc.GetStats(context.Background())
	if err != nil || stats == nil {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
}
