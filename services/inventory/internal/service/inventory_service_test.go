package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

type memRepo struct {
	items          map[string]*model.Item
	lastFilters    map[string]interface{}
	lastLimit      int
	lastOffset     int
	reserveCalls   int
	releaseCalls   int
	softDeleteIDs  []string
	restoreIDs     []string
	bulkCalls      int
	searchErr      error
	reserveErr     error
	releaseErr     error
	softErr        error
	restoreErr     error
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
func (m *memRepo) SearchItems(_ context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	m.lastFilters = filters
	m.lastLimit = limit
	m.lastOffset = offset
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	includeDeleted, _ := filters["include_deleted"].(bool)
	var out []*model.Item
	for _, it := range m.items {
		if author, ok := filters["author_id"].(string); ok && author != "" && it.AuthorID != author {
			continue
		}
		if typ, ok := filters["type"].(string); ok && typ != "" && it.Type != typ {
			continue
		}
		if id, ok := filters["id"].(string); ok && id != "" && it.ID != id {
			continue
		}
		if !includeDeleted && it.DeletedAt != nil {
			continue
		}
		cp := *it
		out = append(out, &cp)
	}
	return out, int64(len(out)), nil
}
func (m *memRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
	return m.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}
func (m *memRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
	return m.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}
func (m *memRepo) BulkCreateItems(_ context.Context, items []*model.Item) error {
	m.bulkCalls++
	for _, it := range items {
		_ = m.CreateItem(context.Background(), it)
	}
	return nil
}
func (m *memRepo) ReserveItem(_ context.Context, id string, quantity int) (int, error) {
	m.reserveCalls++
	if m.reserveErr != nil {
		return 0, m.reserveErr
	}
	it, ok := m.items[id]
	if !ok {
		return 0, errors.New("not found")
	}
	it.Stock -= quantity
	return it.Stock, nil
}
func (m *memRepo) ReleaseItem(_ context.Context, id string, quantity int) (int, error) {
	m.releaseCalls++
	if m.releaseErr != nil {
		return 0, m.releaseErr
	}
	it, ok := m.items[id]
	if !ok {
		return 0, errors.New("not found")
	}
	it.Stock += quantity
	return it.Stock, nil
}
func (m *memRepo) SoftDeleteItem(_ context.Context, id string) error {
	m.softDeleteIDs = append(m.softDeleteIDs, id)
	if m.softErr != nil {
		return m.softErr
	}
	it, ok := m.items[id]
	if !ok {
		return errors.New("not found")
	}
	now := time.Now().UTC()
	it.DeletedAt = &now
	return nil
}
func (m *memRepo) RestoreItem(_ context.Context, id string) error {
	m.restoreIDs = append(m.restoreIDs, id)
	if m.restoreErr != nil {
		return m.restoreErr
	}
	it, ok := m.items[id]
	if !ok {
		return errors.New("not found")
	}
	it.DeletedAt = nil
	return nil
}
func (m *memRepo) GetStats(context.Context) (*model.Stats, error) { return &model.Stats{}, nil }

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

func TestSearchItems_DelegatesFilters(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "auth-1", Type: "card", Name: "live"},
		"b": {ID: "b", AuthorID: "auth-1", Type: "card", Name: "gone", DeletedAt: &now},
		"c": {ID: "c", AuthorID: "other", Type: "merch", Name: "x"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	got, total, err := svc.SearchItems(context.Background(), map[string]interface{}{
		"author_id": "auth-1", "include_deleted": true,
	}, 20, 0)
	if err != nil || total != 2 || len(got) != 2 {
		t.Fatalf("got=%d total=%d err=%v", len(got), total, err)
	}
	if repo.lastLimit != 20 || repo.lastOffset != 0 {
		t.Fatalf("limit=%d offset=%d", repo.lastLimit, repo.lastOffset)
	}
}

func TestGetByAuthor_UsesSearchFilters(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "auth-1", Type: "card"},
		"b": {ID: "b", AuthorID: "auth-2", Type: "card"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	got, total, err := svc.GetByAuthor(context.Background(), "auth-1", 10, 5)
	if err != nil || total != 1 || len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("got=%+v total=%d err=%v", got, total, err)
	}
	if repo.lastFilters["author_id"] != "auth-1" || repo.lastLimit != 10 || repo.lastOffset != 5 {
		t.Fatalf("filters=%v limit=%d offset=%d", repo.lastFilters, repo.lastLimit, repo.lastOffset)
	}
}

func TestGetByType_UsesSearchFilters(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "x", Type: "брелок"},
		"b": {ID: "b", AuthorID: "x", Type: "card"},
	}}
	got, total, err := NewInventoryService(repo, nil, nil).GetByType(context.Background(), "брелок", 50, 0)
	if err != nil || total != 1 || got[0].Type != "брелок" {
		t.Fatalf("got=%+v total=%d err=%v", got, total, err)
	}
	if repo.lastFilters["type"] != "брелок" {
		t.Fatalf("filters=%v", repo.lastFilters)
	}
}

func TestReserveAndRelease_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Stock: 10},
	}}
	svc := NewInventoryService(repo, nil, nil)
	left, err := svc.ReserveItem(context.Background(), "i1", 3)
	if err != nil || left != 7 || repo.reserveCalls != 1 {
		t.Fatalf("left=%d err=%v calls=%d", left, err, repo.reserveCalls)
	}
	left, err = svc.ReleaseItem(context.Background(), "i1", 2)
	if err != nil || left != 9 || repo.releaseCalls != 1 {
		t.Fatalf("left=%d err=%v calls=%d", left, err, repo.releaseCalls)
	}
}

func TestReleaseItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if _, err := svc.ReleaseItem(context.Background(), "", 1); err == nil {
		t.Fatal("empty id")
	}
	if _, err := svc.ReleaseItem(context.Background(), "i1", 0); err == nil {
		t.Fatal("zero qty")
	}
}

func TestSoftDeleteAndRestore_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", AuthorID: "a", Name: "n"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	if err := svc.SoftDeleteItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
	if repo.items["i1"].DeletedAt == nil || len(repo.softDeleteIDs) != 1 {
		t.Fatal("expected soft delete")
	}
	// default search hides soft-deleted
	got, total, err := svc.GetByAuthor(context.Background(), "a", 10, 0)
	if err != nil || total != 0 || len(got) != 0 {
		t.Fatalf("should hide deleted: got=%d total=%d", len(got), total)
	}
	// include_deleted shows it
	got, total, err = svc.SearchItems(context.Background(), map[string]interface{}{
		"author_id": "a", "include_deleted": true,
	}, 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("include_deleted: total=%d err=%v", total, err)
	}
	if err := svc.RestoreItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
	if repo.items["i1"].DeletedAt != nil || len(repo.restoreIDs) != 1 {
		t.Fatal("expected restore")
	}
}

func TestDeleteItem_Success(t *testing.T) {
	if err := NewInventoryService(&memRepo{}, nil, nil).DeleteItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
}

func TestReserveRelease_RepoErrorsPropagate(t *testing.T) {
	repo := &memRepo{
		items:      map[string]*model.Item{"i1": {ID: "i1", Stock: 1}},
		reserveErr: errors.New("reserve fail"),
		releaseErr: errors.New("release fail"),
	}
	svc := NewInventoryService(repo, nil, nil)
	if _, err := svc.ReserveItem(context.Background(), "i1", 1); err == nil {
		t.Fatal("expected reserve error")
	}
	repo.reserveErr = nil
	if _, err := svc.ReleaseItem(context.Background(), "i1", 1); err == nil {
		t.Fatal("expected release error")
	}
}
