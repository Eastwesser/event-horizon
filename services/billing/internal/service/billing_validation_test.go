package service

import (
	"context"
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
)

func TestAddCurrency_RejectsNonPositive(t *testing.T) {
	s := &billingService{} // pg/redis unused until amount validated
	for _, amount := range []int{0, -1, -100} {
		_, err := s.AddCurrency(context.Background(), "u1", repository.Tickets, amount, "x", "ref")
		if err == nil {
			t.Fatalf("amount=%d: want error", amount)
		}
	}
}

func TestSpendCurrency_RejectsNonPositive(t *testing.T) {
	s := &billingService{}
	for _, amount := range []int{0, -5} {
		_, err := s.SpendCurrency(context.Background(), "u1", repository.Tickets, amount, "x", "ref", false)
		if err == nil {
			t.Fatalf("amount=%d: want error", amount)
		}
	}
}
