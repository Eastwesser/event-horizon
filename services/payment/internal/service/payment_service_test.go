package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type mockPayStore struct {
	payment     *model.Payment
	sub         *model.Subscription
	createErr   error
	completeErr error
	getSubErr   error
	created     *model.Payment
	setComplete bool
}

func (m *mockPayStore) CreatePayment(_ context.Context, p *model.Payment) error {
	m.created = p
	return m.createErr
}
func (m *mockPayStore) GetPayment(context.Context, string) (*model.Payment, error) {
	if m.payment == nil {
		return nil, errors.New("not found")
	}
	cp := *m.payment
	return &cp, nil
}
func (m *mockPayStore) CompletePaymentAndActivateSubscription(context.Context, string, string, *model.Subscription, string, map[string]any) error {
	m.setComplete = true
	return m.completeErr
}
func (m *mockPayStore) GetActiveSubscription(context.Context, string) (*model.Subscription, error) {
	if m.getSubErr != nil {
		return nil, m.getSubErr
	}
	if m.sub == nil {
		return nil, errors.New("no sub")
	}
	cp := *m.sub
	return &cp, nil
}

type mockSubCache struct {
	sub    *model.Subscription
	gets   int
	sets   int
	getErr error
}

func (m *mockSubCache) GetSubscription(context.Context, string) (*model.Subscription, error) {
	m.gets++
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.sub == nil {
		return nil, errors.New("miss")
	}
	cp := *m.sub
	return &cp, nil
}
func (m *mockSubCache) SetSubscription(_ context.Context, sub *model.Subscription) error {
	m.sets++
	m.sub = sub
	return nil
}

func TestCreateCheckout(t *testing.T) {
	store := &mockPayStore{}
	svc := New(store, nil, "https://boosty.to/eh", "secret", 30)
	if _, err := svc.CreateCheckout(context.Background(), "u1", "free"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	p, err := svc.CreateCheckout(context.Background(), "u1", model.PlanPresent)
	if err != nil || p.CheckoutURL == "" || !strings.Contains(p.CheckoutURL, "payment_id=") {
		t.Fatalf("%+v err=%v", p, err)
	}
	if store.created == nil || store.created.UserID != "u1" {
		t.Fatal("not persisted")
	}
}

func TestCanPurchaseMerch(t *testing.T) {
	now := time.Now().UTC()
	active := &model.Subscription{UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	store := &mockPayStore{sub: active}
	ok, reason := New(store, nil, "", "", 30).CanPurchaseMerch(context.Background(), "u1")
	if !ok || reason != "" {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
	store.getSubErr = errors.New("missing")
	store.sub = nil
	ok, reason = New(store, nil, "", "", 30).CanPurchaseMerch(context.Background(), "u1")
	if ok || !strings.Contains(reason, "subscription") {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
}

func TestGetSubscription_CacheHit(t *testing.T) {
	now := time.Now().UTC()
	cached := &model.Subscription{UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	cache := &mockSubCache{sub: cached}
	store := &mockPayStore{getSubErr: errors.New("should not hit")}
	got, err := New(store, cache, "", "", 30).GetSubscription(context.Background(), "u1")
	if err != nil || got.UserID != "u1" || cache.gets != 1 {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestConfirmPayment_UnauthorizedAndHappy(t *testing.T) {
	store := &mockPayStore{payment: &model.Payment{ID: "p1", UserID: "u1", Plan: model.PlanPresent, AmountRub: 200}}
	svc := New(store, &mockSubCache{}, "https://x", "sekrit", 30)
	_, err := svc.ConfirmPayment(context.Background(), "p1", "ref", "wrong")
	if !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("%v", err)
	}
	sub, err := svc.ConfirmPayment(context.Background(), "p1", "ref", "sekrit")
	if err != nil || sub.UserID != "u1" || !store.setComplete {
		t.Fatalf("%+v err=%v", sub, err)
	}
}

func TestConfirmPayment_AlreadyPaidIdempotent(t *testing.T) {
	now := time.Now().UTC()
	existing := &model.Subscription{ID: "s1", UserID: "u1", Status: model.StatusActive, ExpiresAt: now.Add(time.Hour)}
	store := &mockPayStore{
		payment:     &model.Payment{ID: "p1", UserID: "u1", Plan: model.PlanPresent, AmountRub: 200},
		completeErr: model.ErrAlreadyPaid,
		sub:         existing,
	}
	cache := &mockSubCache{}
	got, err := New(store, cache, "", "sekrit", 30).ConfirmPayment(context.Background(), "p1", "ref", "sekrit")
	if err != nil || got.ID != "s1" || cache.sets != 1 {
		t.Fatalf("%+v err=%v sets=%d", got, err, cache.sets)
	}
}

func TestNew_DefaultSubscriptionDays(t *testing.T) {
	s := New(&mockPayStore{}, nil, "", "", 0)
	if s.subscriptionDays != 30 {
		t.Fatalf("%d", s.subscriptionDays)
	}
}
