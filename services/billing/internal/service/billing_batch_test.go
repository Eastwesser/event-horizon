package service

import (
	"context"
	"sync"
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

func TestTransactionBatch_AddAndLen(t *testing.T) {
	b := NewTransactionBatch()
	if b.Len() != 0 {
		t.Fatalf("new batch len=%d", b.Len())
	}
	b.Add(repository.Transaction{UserID: "u1", Amount: 10})
	b.Add(repository.Transaction{UserID: "u2", Amount: 20})
	if b.Len() != 2 {
		t.Fatalf("len=%d", b.Len())
	}
}

func TestTransactionBatch_EmptyFlushNoOp(t *testing.T) {
	b := NewTransactionBatch()
	b.Flush(context.Background(), nil) // must not panic on empty batch
}

func TestTransactionBatch_ConcurrentAdd(t *testing.T) {
	b := NewTransactionBatch()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			b.Add(repository.Transaction{UserID: "u", Amount: n})
		}(i)
	}
	wg.Wait()
	if b.Len() != 50 {
		t.Fatalf("len=%d", b.Len())
	}
}

func TestAddCurrency_TableNonPositive(t *testing.T) {
	s := &billingService{}
	cases := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative_one", -1},
		{"large_negative", -999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.AddCurrency(context.Background(), "u1", repository.Tickets, tc.amount, "reason", "ref")
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSpendCurrency_TableNonPositive(t *testing.T) {
	s := &billingService{}
	cases := []struct {
		name   string
		amount int
	}{
		{"zero", 0},
		{"negative", -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.SpendCurrency(context.Background(), "u1", repository.Tickets, tc.amount, "reason", "ref", false)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
