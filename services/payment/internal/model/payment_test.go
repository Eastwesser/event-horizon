package model_test

import (
	"errors"
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

func TestSubscriptionIsActive_Nil(t *testing.T) {
	var sub *model.Subscription
	if sub.IsActive(time.Now()) {
		t.Fatal("nil subscription must not be active")
	}
}

func TestSubscriptionIsActive_WrongStatus(t *testing.T) {
	now := time.Now().UTC()
	sub := &model.Subscription{Status: model.StatusExpired, ExpiresAt: now.Add(time.Hour)}
	if sub.IsActive(now) {
		t.Fatal("expired status must not be active")
	}
}

func TestPlanAmountRub_Table(t *testing.T) {
	cases := []struct {
		plan    string
		want    int
		wantErr bool
	}{
		{model.PlanPresent, 200, false},
		{model.PlanFuture, 300, false},
		{"free", 0, true},
		{"", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.plan, func(t *testing.T) {
			got, err := model.PlanAmountRub(tc.plan)
			if tc.wantErr {
				if !errors.Is(err, model.ErrInvalidInput) {
					t.Fatalf("plan=%q err=%v", tc.plan, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("plan=%q got=%d err=%v", tc.plan, got, err)
			}
		})
	}
}
