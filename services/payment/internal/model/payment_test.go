package model_test

import (
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

func TestPlanAmountRub(t *testing.T) {
	a, err := model.PlanAmountRub(model.PlanPresent)
	if err != nil || a != 200 {
		t.Fatalf("present: got %d %v", a, err)
	}
	a, err = model.PlanAmountRub(model.PlanFuture)
	if err != nil || a != 300 {
		t.Fatalf("future: got %d %v", a, err)
	}
	if _, err := model.PlanAmountRub("free"); err != model.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSubscriptionIsActive(t *testing.T) {
	now := time.Now().UTC()
	sub := &model.Subscription{Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	if !sub.IsActive(now) {
		t.Fatal("expected active")
	}
	if sub.IsActive(now.Add(2 * time.Hour)) {
		t.Fatal("expected expired")
	}
}
