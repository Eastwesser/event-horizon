package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

type mockBillingStore struct {
	balance     int
	addBal      int
	spendBal    int
	txs         []repository.Transaction
	txTotal     int
	lastLimit   int
	addCalls    int
	spendCalls  int
	deleted     bool
	err         error
	spendErr    error
}

func (m *mockBillingStore) GetBalance(context.Context, string, repository.CurrencyType) (int, error) {
	return m.balance, m.err
}
func (m *mockBillingStore) AddBalance(context.Context, string, repository.CurrencyType, int, string, string) (int, error) {
	m.addCalls++
	if m.err != nil {
		return 0, m.err
	}
	return m.addBal, nil
}
func (m *mockBillingStore) SpendBalance(context.Context, string, repository.CurrencyType, int, string, string) (int, error) {
	m.spendCalls++
	if m.spendErr != nil {
		return 0, m.spendErr
	}
	return m.spendBal, nil
}
func (m *mockBillingStore) GetTransactionHistory(_ context.Context, _ string, _ repository.CurrencyType, limit, _ int) ([]repository.Transaction, int, error) {
	m.lastLimit = limit
	return m.txs, m.txTotal, m.err
}
func (m *mockBillingStore) BatchInsertTransactions(context.Context, []repository.Transaction) error {
	return nil
}

type mockBillingCache struct {
	cached   int
	cacheErr error
	setCalls int
	delCalls int
}

func (m *mockBillingCache) GetBalance(context.Context, string, repository.CurrencyType) (int, error) {
	return m.cached, m.cacheErr
}
func (m *mockBillingCache) SetBalance(context.Context, string, repository.CurrencyType, int, time.Duration) error {
	m.setCalls++
	return nil
}
func (m *mockBillingCache) DeleteBalance(context.Context, string, repository.CurrencyType) error {
	m.delCalls++
	return nil
}

func TestGetBalance_CacheHit(t *testing.T) {
	pg := &mockBillingStore{balance: 999}
	cache := &mockBillingCache{cached: 42}
	bal, err := NewBillingService(pg, cache).GetBalance(context.Background(), "u1", repository.Tickets)
	if err != nil || bal != 42 || cache.setCalls != 0 {
		t.Fatalf("bal=%d err=%v sets=%d", bal, err, cache.setCalls)
	}
}

func TestGetBalance_CacheMissLoadsPG(t *testing.T) {
	pg := &mockBillingStore{balance: 77}
	cache := &mockBillingCache{cached: -1}
	bal, err := NewBillingService(pg, cache).GetBalance(context.Background(), "u1", repository.Lamps)
	if err != nil || bal != 77 || cache.setCalls != 1 {
		t.Fatalf("bal=%d err=%v sets=%d", bal, err, cache.setCalls)
	}
}

func TestAddCurrency_TicketsAndLamps(t *testing.T) {
	pg := &mockBillingStore{addBal: 110}
	cache := &mockBillingCache{cached: -1}
	svc := NewBillingService(pg, cache)
	bal, err := svc.AddCurrency(context.Background(), "u1", repository.Tickets, 10, "reward", "ref-1")
	if err != nil || bal != 110 || pg.addCalls != 1 || cache.delCalls != 1 {
		t.Fatalf("bal=%d err=%v", bal, err)
	}
	_, err = svc.AddCurrency(context.Background(), "u1", repository.Lamps, 0, "x", "r")
	if err == nil || !strings.Contains(err.Error(), "positive") {
		t.Fatalf("%v", err)
	}
}

func TestSpendCurrency_CheckOnlyAndReal(t *testing.T) {
	pg := &mockBillingStore{balance: 5, spendBal: 0}
	cache := &mockBillingCache{cached: -1}
	svc := NewBillingService(pg, cache)
	_, err := svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", true)
	if err == nil || !strings.Contains(err.Error(), "insufficient") {
		t.Fatalf("%v", err)
	}
	pg.balance = 50
	bal, err := svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", true)
	if err != nil || bal != 50 || pg.spendCalls != 0 {
		t.Fatalf("checkOnly bal=%d err=%v spends=%d", bal, err, pg.spendCalls)
	}
	bal, err = svc.SpendCurrency(context.Background(), "u1", repository.Tickets, 10, "buy", "r", false)
	if err != nil || bal != 0 || pg.spendCalls != 1 || cache.delCalls != 1 {
		t.Fatalf("spend bal=%d err=%v", bal, err)
	}
}

func TestSpendCurrency_RepoError(t *testing.T) {
	pg := &mockBillingStore{spendErr: errors.New("db")}
	_, err := NewBillingService(pg, &mockBillingCache{cached: -1}).
		SpendCurrency(context.Background(), "u1", repository.Tickets, 1, "x", "r", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTransactionHistory_ClampsLimit(t *testing.T) {
	pg := &mockBillingStore{txs: []repository.Transaction{{Amount: 1}}, txTotal: 1}
	svc := NewBillingService(pg, &mockBillingCache{cached: -1})
	_, _, err := svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 0, 0)
	if err != nil || pg.lastLimit != 20 {
		t.Fatalf("limit=%d err=%v", pg.lastLimit, err)
	}
	_, _, _ = svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 200, 0)
	if pg.lastLimit != 20 {
		t.Fatalf("limit=%d", pg.lastLimit)
	}
	_, _, _ = svc.GetTransactionHistory(context.Background(), "u1", repository.Tickets, 50, 0)
	if pg.lastLimit != 50 {
		t.Fatalf("limit=%d", pg.lastLimit)
	}
}
