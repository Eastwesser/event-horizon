package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"google.golang.org/grpc"
)

type mockShopStore struct {
	item              *repository.Item
	getItemErr        error
	owned             bool
	ownedErr          error
	ownedIDs          []string
	listOwnedErr      error
	purchaseErr       error
	purchaseCalls     int
	refund            *repository.RefundResult
	refundErr         error
	refundCalls       int
	listRows          []repository.PurchaseRecord
	listTotal         int64
	listSales         int64
	listTickets       int64
	listErr           error
	inventory         []repository.Item
	items             []repository.Item
	getItemsErr       error
}

func (m *mockShopStore) GetItems(context.Context, string, string) ([]repository.Item, error) {
	if m.getItemsErr != nil {
		return nil, m.getItemsErr
	}
	return m.items, nil
}
func (m *mockShopStore) GetItemByID(context.Context, string) (*repository.Item, error) {
	if m.getItemErr != nil {
		return nil, m.getItemErr
	}
	if m.item == nil {
		return nil, errors.New("not found")
	}
	cp := *m.item
	return &cp, nil
}
func (m *mockShopStore) IsItemOwned(context.Context, string, string) (bool, error) {
	return m.owned, m.ownedErr
}
func (m *mockShopStore) ListOwnedItemIDs(context.Context, string) ([]string, error) {
	if m.listOwnedErr != nil {
		return nil, m.listOwnedErr
	}
	if m.ownedIDs != nil {
		return m.ownedIDs, nil
	}
	return nil, nil
}
func (m *mockShopStore) PurchaseItemWithStock(_ context.Context, _, _ string, _ int, _ *repository.OutboxRecord) error {
	m.purchaseCalls++
	return m.purchaseErr
}
func (m *mockShopStore) GetUserInventory(context.Context, string) ([]repository.Item, error) {
	return m.inventory, nil
}
func (m *mockShopStore) RefundPurchase(context.Context, string, string) (*repository.RefundResult, error) {
	m.refundCalls++
	if m.refundErr != nil {
		return nil, m.refundErr
	}
	if m.refund == nil {
		return nil, errors.New("purchase not found")
	}
	cp := *m.refund
	return &cp, nil
}
func (m *mockShopStore) MarkPurchaseFulfilled(context.Context, string) error {
	return nil
}
func (m *mockShopStore) ListPurchasesByItemIDs(context.Context, []string, int, int) ([]repository.PurchaseRecord, int64, int64, int64, error) {
	return m.listRows, m.listTotal, m.listSales, m.listTickets, m.listErr
}
func (m *mockShopStore) CreateItemFromInventory(context.Context, string, string, string, float64, int) error {
	return nil
}

type mockShopCache struct {
	deleted []string
	items   []repository.Item
	getErr  error
}

func (m *mockShopCache) GetItems(context.Context, string) ([]repository.Item, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.items, nil
}
func (m *mockShopCache) SetItems(context.Context, string, []repository.Item, time.Duration) error {
	return nil
}
func (m *mockShopCache) Delete(_ context.Context, key string) error {
	m.deleted = append(m.deleted, key)
	return nil
}

type stubBilling struct {
	spendBal     int32
	spendOK      bool
	spendMsg     string
	spendErr     error
	spendCalls   int
	addBal       int32
	addErr       error
	addCalls     int
	getBal       int32
	getBalErr    error
}

func (b *stubBilling) GetBalance(context.Context, *billingPb.GetBalanceRequest, ...grpc.CallOption) (*billingPb.GetBalanceResponse, error) {
	if b.getBalErr != nil {
		return nil, b.getBalErr
	}
	return &billingPb.GetBalanceResponse{Balance: b.getBal}, nil
}
func (b *stubBilling) GetAllBalances(context.Context, *billingPb.GetAllBalancesRequest, ...grpc.CallOption) (*billingPb.GetAllBalancesResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) AddCurrency(context.Context, *billingPb.AddCurrencyRequest, ...grpc.CallOption) (*billingPb.AddCurrencyResponse, error) {
	b.addCalls++
	if b.addErr != nil {
		return nil, b.addErr
	}
	return &billingPb.AddCurrencyResponse{NewBalance: b.addBal, Success: true}, nil
}
func (b *stubBilling) SpendCurrency(context.Context, *billingPb.SpendCurrencyRequest, ...grpc.CallOption) (*billingPb.SpendCurrencyResponse, error) {
	b.spendCalls++
	if b.spendErr != nil {
		return nil, b.spendErr
	}
	return &billingPb.SpendCurrencyResponse{
		Success:    b.spendOK,
		Message:    b.spendMsg,
		NewBalance: b.spendBal,
	}, nil
}
func (b *stubBilling) GetTransactionHistory(context.Context, *billingPb.GetTransactionHistoryRequest, ...grpc.CallOption) (*billingPb.GetTransactionHistoryResponse, error) {
	return nil, errors.New("unused")
}

func newTestShop(store *mockShopStore, cache *mockShopCache, billing *stubBilling) *shopService {
	if cache == nil {
		cache = &mockShopCache{getErr: errors.New("miss")}
	}
	return &shopService{
		pgRepo:    store,
		redisRepo: cache,
		billing:   billing,
	}
}

func withPayment(s *shopService, p stubPayment) *shopService {
	s.payment = p
	return s
}

func TestNew_ConstructsService(t *testing.T) {
	svc := New(&mockShopStore{}, &mockShopCache{}, nil, &stubBilling{}, nil)
	if svc == nil {
		t.Fatal("nil service")
	}
}

func TestPurchaseItem_NotFoundUnavailableOwned(t *testing.T) {
	billing := &stubBilling{spendOK: true, spendBal: 90}
	store := &mockShopStore{getItemErr: errors.New("missing")}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrItemNotFound) {
		t.Fatalf("not found: %v", err)
	}

	store = &mockShopStore{item: &repository.Item{ID: "i1", Available: false, Price: 10, Category: "card"}}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrItemUnavailable) {
		t.Fatalf("unavailable: %v", err)
	}

	store = &mockShopStore{item: &repository.Item{ID: "i1", Available: true, Price: 10, Category: "card"}, owned: true}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrAlreadyOwned) {
		t.Fatalf("owned: %v", err)
	}
	if billing.spendCalls != 0 {
		t.Fatalf("should not spend when owned")
	}
}

func TestPurchaseItem_MerchBlocked(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "m1", Available: true, Price: 50, Category: "merch"}}
	billing := &stubBilling{spendOK: true}
	svc := withPayment(newTestShop(store, nil, billing), stubPayment{allowed: false})
	_, err := svc.PurchaseItem(context.Background(), "u1", "m1")
	if !errors.Is(err, model.ErrSubscriptionRequired) {
		t.Fatalf("%v", err)
	}
	if billing.spendCalls != 0 || store.purchaseCalls != 0 {
		t.Fatalf("merch gate should short-circuit")
	}
}

func TestPurchaseItem_InsufficientFunds(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "i1", Available: true, Price: 100, Category: "card", Name: "Card"}}
	billing := &stubBilling{spendErr: errors.New("insufficient balance")}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrInsufficientFunds) {
		t.Fatalf("%v", err)
	}

	billing = &stubBilling{spendOK: false, spendMsg: "insufficient balance"}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrInsufficientFunds) {
		t.Fatalf("msg path: %v", err)
	}
}

func TestPurchaseItem_Success(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "item-uuid-1", Available: true, Price: 25, Category: "card", Name: "Keychain"}}
	cache := &mockShopCache{getErr: errors.New("miss")}
	billing := &stubBilling{spendOK: true, spendBal: 75}
	bal, err := newTestShop(store, cache, billing).PurchaseItem(context.Background(), "u1", "item-uuid-1")
	if err != nil || bal != 75 {
		t.Fatalf("bal=%d err=%v", bal, err)
	}
	if billing.spendCalls != 1 || store.purchaseCalls != 1 {
		t.Fatalf("spend=%d purchase=%d", billing.spendCalls, store.purchaseCalls)
	}
	if len(cache.deleted) == 0 {
		t.Fatal("expected cache invalidation")
	}
}

func TestPurchaseItem_RecordFailureRefunds(t *testing.T) {
	store := &mockShopStore{
		item:        &repository.Item{ID: "i1", Available: true, Price: 10, Category: "card", Name: "x"},
		purchaseErr: errors.New("db down"),
	}
	billing := &stubBilling{spendOK: true, spendBal: 0, addBal: 10}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if err == nil || !strings.Contains(err.Error(), "failed to record purchase") {
		t.Fatalf("%v", err)
	}
	if billing.addCalls != 1 {
		t.Fatalf("expected refund AddCurrency, got %d", billing.addCalls)
	}
}

func TestCancelPurchase_NotFound(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "card"},
		refundErr: errors.New("purchase not found"),
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrPurchaseNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_AlreadyRefundedIdempotent(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{AlreadyRefunded: true},
	}
	billing := &stubBilling{getBal: 42}
	got, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err != nil || !got.AlreadyRefunded || got.NewBalance != 42 || got.RefundedAmount != 0 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if billing.addCalls != 0 {
		t.Fatal("should not AddCurrency when already refunded")
	}
}

func TestCancelPurchase_Success(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{PurchaseID: "p1", Price: 25},
	}
	cache := &mockShopCache{}
	billing := &stubBilling{addBal: 100}
	got, err := newTestShop(store, cache, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err != nil || got.AlreadyRefunded || got.RefundedAmount != 25 || got.NewBalance != 100 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if billing.addCalls != 1 || store.refundCalls != 1 || len(cache.deleted) == 0 {
		t.Fatalf("add=%d refund=%d deleted=%v", billing.addCalls, store.refundCalls, cache.deleted)
	}
}

func TestCancelPurchase_MissingCatalogItemStillRefunds(t *testing.T) {
	store := &mockShopStore{
		getItemErr: errors.New("gone"),
		refund:     &repository.RefundResult{PurchaseID: "p1", Price: 7},
	}
	billing := &stubBilling{addBal: 7}
	got, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "orphan")
	if err != nil || got.RefundedAmount != 7 {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestListPurchasesByItemIDs_Delegates(t *testing.T) {
	store := &mockShopStore{
		listRows:    []repository.PurchaseRecord{{ID: "p1", ItemID: "i1", Price: 10}},
		listTotal:   1,
		listSales:   1,
		listTickets: 10,
	}
	rows, total, sales, tickets, err := newTestShop(store, nil, &stubBilling{}).
		ListPurchasesByItemIDs(context.Background(), []string{"i1"}, 20, 0)
	if err != nil || total != 1 || sales != 1 || tickets != 10 || len(rows) != 1 {
		t.Fatalf("rows=%v total=%d sales=%d tickets=%d err=%v", rows, total, sales, tickets, err)
	}
}

func TestGetInventory_Delegates(t *testing.T) {
	store := &mockShopStore{inventory: []repository.Item{{ID: "i1", Name: "n"}}}
	got, err := newTestShop(store, nil, &stubBilling{}).GetInventory(context.Background(), "u1")
	if err != nil || len(got) != 1 || got[0].ID != "i1" {
		t.Fatalf("%v err=%v", got, err)
	}
}

func TestGetItems_CacheHitSetsOwned(t *testing.T) {
	store := &mockShopStore{ownedIDs: []string{"i1"}}
	cache := &mockShopCache{items: []repository.Item{{ID: "i1", Name: "cached"}}}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err != nil || len(got) != 1 || !got[0].Owned {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestGetItems_CacheMissLoadsPostgres(t *testing.T) {
	store := &mockShopStore{
		items:    []repository.Item{{ID: "i2", Name: "db"}},
		ownedIDs: []string{},
	}
	cache := &mockShopCache{getErr: errors.New("miss")}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "card", "hexagon", "u1")
	if err != nil || len(got) != 1 || got[0].Name != "db" || got[0].Owned {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestGetItems_BatchOwnedMixed(t *testing.T) {
	store := &mockShopStore{
		ownedIDs: []string{"i2"},
	}
	cache := &mockShopCache{items: []repository.Item{
		{ID: "i1", Name: "a"},
		{ID: "i2", Name: "b"},
		{ID: "i3", Name: "c"},
	}}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err != nil || len(got) != 3 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if got[0].Owned || !got[1].Owned || got[2].Owned {
		t.Fatalf("owned flags: %+v %+v %+v", got[0], got[1], got[2])
	}
}

func TestGetItems_PostgresError(t *testing.T) {
	store := &mockShopStore{getItemsErr: errors.New("db")}
	cache := &mockShopCache{getErr: errors.New("miss")}
	_, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelPurchase_AddCurrencyFails(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{PurchaseID: "p1", Price: 25},
	}
	billing := &stubBilling{addErr: errors.New("billing down")}
	_, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err == nil || !strings.Contains(err.Error(), "failed to refund tickets") {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_RefundWindowExpired(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "game_skin"},
		refundErr: model.ErrRefundWindowExpired,
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrRefundWindowExpired) {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_AlreadyFulfilledPhysical(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "merch"},
		refundErr: model.ErrAlreadyFulfilled,
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrAlreadyFulfilled) {
		t.Fatalf("%v", err)
	}
}

func TestGetInventory_ComputesCanCancel(t *testing.T) {
	until := time.Now().UTC().Add(48 * time.Hour)
	store := &mockShopStore{
		inventory: []repository.Item{{
			ID:              "i1",
			Category:        "game_skin",
			PurchaseID:      "p1",
			RefundableUntil: &until,
		}},
	}
	got, err := newTestShop(store, nil, &stubBilling{}).GetInventory(context.Background(), "u1")
	if err != nil || len(got) != 1 || !got[0].CanCancel {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestSetKafkaProducer(t *testing.T) {
	s := &shopService{}
	s.SetKafkaProducer(nil)
	if s.kafkaProd != nil {
		t.Fatal("expected nil producer")
	}
}
