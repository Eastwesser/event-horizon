package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
)

type mockStore struct {
	upsertFn               func(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error
	getByUserIDFn          func(ctx context.Context, userID string) (*model.Author, error)
	listFn                 func(ctx context.Context, limit, offset int) ([]*model.Author, int64, error)
	getPendingFn           func(ctx context.Context, userID string) (*model.AuthorApplication, error)
	getLatestFn            func(ctx context.Context, userID string) (*model.AuthorApplication, error)
	insertAppFn            func(ctx context.Context, a *model.AuthorApplication) error
	listAppsFn             func(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error)
	approveFn              func(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error)
	rejectFn               func(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error)
	revertFn               func(ctx context.Context, applicationID string) (*model.AuthorApplication, error)
	insertCalls            int
	approveCalls           int
	rejectCalls            int
	revertCalls            int
	upsertCalls            int
}

func (m *mockStore) Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error {
	m.upsertCalls++
	if m.upsertFn != nil {
		return m.upsertFn(ctx, a, eventType, eventPayload)
	}
	return nil
}
func (m *mockStore) GetByUserID(ctx context.Context, userID string) (*model.Author, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, userID)
	}
	return nil, model.ErrNotFound
}
func (m *mockStore) List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockStore) GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	if m.getPendingFn != nil {
		return m.getPendingFn(ctx, userID)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	if m.getLatestFn != nil {
		return m.getLatestFn(ctx, userID)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) InsertApplication(ctx context.Context, a *model.AuthorApplication) error {
	m.insertCalls++
	if m.insertAppFn != nil {
		return m.insertAppFn(ctx, a)
	}
	return nil
}
func (m *mockStore) ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error) {
	if m.listAppsFn != nil {
		return m.listAppsFn(ctx, status, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockStore) ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
	m.approveCalls++
	if m.approveFn != nil {
		return m.approveFn(ctx, applicationID, reviewerID)
	}
	return nil, nil, model.ErrApplicationMissing
}
func (m *mockStore) RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
	m.rejectCalls++
	if m.rejectFn != nil {
		return m.rejectFn(ctx, applicationID, reviewerID, note)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
	m.revertCalls++
	if m.revertFn != nil {
		return m.revertFn(ctx, applicationID)
	}
	return nil, model.ErrApplicationMissing
}

type mockCache struct {
	store  map[string]*model.Author
	gets   int
	sets   int
	dels   int
	getErr error
}

func (c *mockCache) Get(_ context.Context, userID string) (*model.Author, error) {
	c.gets++
	if c.getErr != nil {
		return nil, c.getErr
	}
	if c.store == nil {
		return nil, errors.New("miss")
	}
	a, ok := c.store[userID]
	if !ok {
		return nil, errors.New("miss")
	}
	cp := *a
	return &cp, nil
}
func (c *mockCache) Set(_ context.Context, a *model.Author) error {
	c.sets++
	if c.store == nil {
		c.store = map[string]*model.Author{}
	}
	cp := *a
	c.store[a.UserID] = &cp
	return nil
}
func (c *mockCache) Delete(_ context.Context, userID string) error {
	c.dels++
	if c.store != nil {
		delete(c.store, userID)
	}
	return nil
}

func TestUpsertProfile_Validation(t *testing.T) {
	svc := New(&mockStore{}, nil)
	if _, err := svc.UpsertProfile(context.Background(), "", "Name", "", "", ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty user: %v", err)
	}
	if _, err := svc.UpsertProfile(context.Background(), "u1", "", "", "", ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty name: %v", err)
	}
}

func TestUpsertProfile_SuccessWithPortfolioAndCache(t *testing.T) {
	store := &mockStore{
		upsertFn: func(_ context.Context, a *model.Author, eventType string, _ map[string]any) error {
			if a.UserID != "u1" || a.DisplayName != "Den" || a.Portfolio != "https://gh.com/x" || !a.Active {
				t.Fatalf("bad author: %+v", a)
			}
			if eventType != "author.upserted" {
				t.Fatalf("event=%s", eventType)
			}
			a.ID = "aid-1"
			return nil
		},
	}
	cache := &mockCache{}
	svc := New(store, cache)
	got, err := svc.UpsertProfile(context.Background(), "u1", "Den", "bio", "av", "https://gh.com/x")
	if err != nil {
		t.Fatal(err)
	}
	if got.Portfolio != "https://gh.com/x" || store.upsertCalls != 1 || cache.sets != 1 {
		t.Fatalf("got=%+v upsert=%d sets=%d", got, store.upsertCalls, cache.sets)
	}
}

func TestGetAuthor_CacheHit(t *testing.T) {
	cache := &mockCache{store: map[string]*model.Author{
		"u1": {UserID: "u1", DisplayName: "cached"},
	}}
	store := &mockStore{
		getByUserIDFn: func(context.Context, string) (*model.Author, error) {
			t.Fatal("repo should not be called on cache hit")
			return nil, nil
		},
	}
	got, err := New(store, cache).GetAuthor(context.Background(), "u1")
	if err != nil || got.DisplayName != "cached" || cache.gets != 1 {
		t.Fatalf("got=%+v err=%v gets=%d", got, err, cache.gets)
	}
}

func TestGetAuthor_CacheMissLoadsRepo(t *testing.T) {
	cache := &mockCache{getErr: errors.New("miss")}
	store := &mockStore{
		getByUserIDFn: func(_ context.Context, userID string) (*model.Author, error) {
			return &model.Author{UserID: userID, DisplayName: "from-db"}, nil
		},
	}
	got, err := New(store, cache).GetAuthor(context.Background(), "u1")
	if err != nil || got.DisplayName != "from-db" || cache.sets != 1 {
		t.Fatalf("got=%+v err=%v sets=%d", got, err, cache.sets)
	}
}

func TestListAuthors_Delegates(t *testing.T) {
	store := &mockStore{
		listFn: func(_ context.Context, limit, offset int) ([]*model.Author, int64, error) {
			if limit != 10 || offset != 5 {
				t.Fatalf("limit=%d offset=%d", limit, offset)
			}
			return []*model.Author{{UserID: "u1"}}, 1, nil
		},
	}
	list, total, err := New(store, nil).ListAuthors(context.Background(), 10, 5)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestSubmitApplication_ValidationAndPrivileged(t *testing.T) {
	svc := New(&mockStore{}, nil)
	_, err := svc.SubmitApplication(context.Background(), "", "user", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty user: %v", err)
	}
	_, err = svc.SubmitApplication(context.Background(), "u1", "author", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrAlreadyPrivileged) {
		t.Fatalf("author: %v", err)
	}
	_, err = svc.SubmitApplication(context.Background(), "u1", "ADMIN", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrAlreadyPrivileged) {
		t.Fatalf("admin: %v", err)
	}
}

func TestSubmitApplication_ReturnsExistingPending(t *testing.T) {
	existing := &model.AuthorApplication{ID: "app-1", UserID: "u1", Status: model.ApplicationPending}
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			return existing, nil
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "u1", "user", "Name", "p", "motivation", "a@b.c")
	if err != nil || got.ID != "app-1" || store.insertCalls != 0 {
		t.Fatalf("got=%+v err=%v inserts=%d", got, err, store.insertCalls)
	}
}

func TestSubmitApplication_CreatesNew(t *testing.T) {
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			return nil, model.ErrApplicationMissing
		},
		insertAppFn: func(_ context.Context, a *model.AuthorApplication) error {
			if a.UserID != "u1" || a.Status != model.ApplicationPending {
				t.Fatalf("bad app: %+v", a)
			}
			if a.Payload.DisplayName != "Name" || a.Payload.Motivation != "why me" || a.Payload.ContactEmail != "a@b.c" {
				t.Fatalf("payload: %+v", a.Payload)
			}
			if a.ID == "" {
				t.Fatal("id empty")
			}
			return nil
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "  u1  ", "user", "  Name  ", " pf ", "  why me  ", "  a@b.c  ")
	if err != nil || got.Status != model.ApplicationPending || store.insertCalls != 1 {
		t.Fatalf("got=%+v err=%v inserts=%d", got, err, store.insertCalls)
	}
}

func TestSubmitApplication_InsertRaceReturnsPending(t *testing.T) {
	pending := &model.AuthorApplication{ID: "race-1", UserID: "u1", Status: model.ApplicationPending}
	calls := 0
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			calls++
			if calls == 1 {
				return nil, model.ErrApplicationMissing
			}
			return pending, nil
		},
		insertAppFn: func(context.Context, *model.AuthorApplication) error {
			return errors.New("unique violation")
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "u1", "user", "Name", "", "motivation", "a@b.c")
	if err != nil || got.ID != "race-1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGetMyApplication_Validation(t *testing.T) {
	_, err := New(&mockStore{}, nil).GetMyApplication(context.Background(), "  ")
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
}

func TestGetMyApplication_Delegates(t *testing.T) {
	want := &model.AuthorApplication{ID: "a1", UserID: "u1"}
	store := &mockStore{
		getLatestFn: func(_ context.Context, userID string) (*model.AuthorApplication, error) {
			if userID != "u1" {
				t.Fatalf("user=%s", userID)
			}
			return want, nil
		},
	}
	got, err := New(store, nil).GetMyApplication(context.Background(), " u1 ")
	if err != nil || got.ID != "a1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestListApplications_StatusFilter(t *testing.T) {
	svc := New(&mockStore{
		listAppsFn: func(_ context.Context, status string, _, _ int) ([]*model.AuthorApplication, int64, error) {
			return []*model.AuthorApplication{{ID: "1", Status: model.ApplicationStatus(status)}}, 1, nil
		},
	}, nil)
	if _, _, err := svc.ListApplications(context.Background(), "weird", 10, 0); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("want invalid, got %v", err)
	}
	list, total, err := svc.ListApplications(context.Background(), "PENDING", 10, 0)
	if err != nil || total != 1 || list[0].Status != "pending" {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestApproveApplication_ValidationAndSuccess(t *testing.T) {
	svc := New(&mockStore{}, nil)
	if _, _, err := svc.ApproveApplication(context.Background(), "", "r1"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	now := time.Now().UTC()
	author := &model.Author{UserID: "u1", DisplayName: "Den", Active: true}
	app := &model.AuthorApplication{ID: "app-1", UserID: "u1", Status: model.ApplicationApproved, ReviewedAt: &now}
	store := &mockStore{
		approveFn: func(_ context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
			if applicationID != "app-1" || reviewerID != "admin-1" {
				t.Fatalf("%s %s", applicationID, reviewerID)
			}
			return app, author, nil
		},
	}
	cache := &mockCache{}
	gotApp, gotAuthor, err := New(store, cache).ApproveApplication(context.Background(), " app-1 ", " admin-1 ")
	if err != nil || gotApp.Status != model.ApplicationApproved || gotAuthor.UserID != "u1" || cache.sets != 1 {
		t.Fatalf("app=%+v author=%+v err=%v sets=%d", gotApp, gotAuthor, err, cache.sets)
	}
}

func TestApproveApplication_AlreadyReviewed(t *testing.T) {
	store := &mockStore{
		approveFn: func(context.Context, string, string) (*model.AuthorApplication, *model.Author, error) {
			return nil, nil, model.ErrAlreadyReviewed
		},
	}
	_, _, err := New(store, nil).ApproveApplication(context.Background(), "app-1", "admin")
	if !errors.Is(err, model.ErrAlreadyReviewed) {
		t.Fatalf("%v", err)
	}
}

func TestRejectApplication_ValidationAndSuccess(t *testing.T) {
	if _, err := New(&mockStore{}, nil).RejectApplication(context.Background(), "a", "", "note"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatal("expected invalid")
	}
	store := &mockStore{
		rejectFn: func(_ context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
			if applicationID != "app-1" || reviewerID != "admin" || note != "nope" {
				t.Fatalf("%s %s %s", applicationID, reviewerID, note)
			}
			return &model.AuthorApplication{ID: applicationID, Status: model.ApplicationRejected, ReviewerNote: note}, nil
		},
	}
	got, err := New(store, nil).RejectApplication(context.Background(), " app-1 ", " admin ", "  nope  ")
	if err != nil || got.Status != model.ApplicationRejected || store.rejectCalls != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestRevertApplication_ClearsCache(t *testing.T) {
	if _, err := New(&mockStore{}, nil).RevertApplication(context.Background(), ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatal("expected invalid")
	}
	store := &mockStore{
		revertFn: func(_ context.Context, id string) (*model.AuthorApplication, error) {
			return &model.AuthorApplication{ID: id, UserID: "u1", Status: model.ApplicationPending}, nil
		},
	}
	cache := &mockCache{store: map[string]*model.Author{"u1": {UserID: "u1"}}}
	got, err := New(store, cache).RevertApplication(context.Background(), "app-1")
	if err != nil || got.Status != model.ApplicationPending || cache.dels != 1 {
		t.Fatalf("got=%+v err=%v dels=%d", got, err, cache.dels)
	}
}
