package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

// PaymentStore is the PostgreSQL port.
type PaymentStore interface {
	CreatePayment(ctx context.Context, p *model.Payment) error
	GetPayment(ctx context.Context, id string) (*model.Payment, error)
	CompletePaymentAndActivateSubscription(ctx context.Context, paymentID, providerRef string, sub *model.Subscription, eventType string, eventPayload map[string]any) error
	GetActiveSubscription(ctx context.Context, userID string) (*model.Subscription, error)
	CreateSubscription(ctx context.Context, sub *model.Subscription) error
	RevokeSubscription(ctx context.Context, userID string) error
}

// SubscriptionCache is the Redis port.
type SubscriptionCache interface {
	GetSubscription(ctx context.Context, userID string) (*model.Subscription, error)
	SetSubscription(ctx context.Context, sub *model.Subscription) error
	DeleteSubscription(ctx context.Context, userID string) error
}

type PaymentService struct {
	repo             PaymentStore
	cache            SubscriptionCache
	boostyURL        string
	webhookSecret    string
	subscriptionDays int
}

func New(
	repo PaymentStore,
	cache SubscriptionCache,
	boostyURL, webhookSecret string,
	subscriptionDays int,
) *PaymentService {
	if subscriptionDays <= 0 {
		subscriptionDays = 30
	}
	return &PaymentService{
		repo:             repo,
		cache:            cache,
		boostyURL:        boostyURL,
		webhookSecret:    webhookSecret,
		subscriptionDays: subscriptionDays,
	}
}

func (s *PaymentService) CreateCheckout(ctx context.Context, userID, plan string) (*model.Payment, error) {
	amount, err := model.PlanAmountRub(plan)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	checkout := fmt.Sprintf("%s?payment_id=%s&plan=%s&amount=%d", s.boostyURL, id, plan, amount)
	p := &model.Payment{
		ID:          id,
		UserID:      userID,
		Plan:        plan,
		AmountRub:   amount,
		Status:      model.StatusPending,
		Provider:    "boosty",
		CheckoutURL: checkout,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID, providerRef, webhookSecret string) (*model.Subscription, error) {
	if s.webhookSecret != "" && webhookSecret != "" && webhookSecret != s.webhookSecret {
		return nil, model.ErrUnauthorized
	}
	p, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sub := &model.Subscription{
		ID:        uuid.NewString(),
		UserID:    p.UserID,
		Plan:      p.Plan,
		Status:    model.StatusActive,
		AmountRub: p.AmountRub,
		PaymentID: p.ID,
		StartsAt:  now,
		ExpiresAt: now.Add(time.Duration(s.subscriptionDays) * 24 * time.Hour),
	}
	event := map[string]any{
		"event":           "payment.completed",
		"payment_id":      p.ID,
		"user_id":         p.UserID,
		"plan":            p.Plan,
		"amount_rub":      p.AmountRub,
		"subscription_id": sub.ID,
		"expires_at":      sub.ExpiresAt.Unix(),
		"provider_ref":    providerRef,
		"timestamp":       now.Unix(),
	}
	if err := s.repo.CompletePaymentAndActivateSubscription(ctx, paymentID, providerRef, sub, "payment.completed", event); err != nil {
		if errors.Is(err, model.ErrAlreadyPaid) {
			existing, getErr := s.repo.GetActiveSubscription(ctx, p.UserID)
			if getErr == nil {
				if s.cache != nil {
					_ = s.cache.SetSubscription(ctx, existing)
				}
				return existing, nil
			}
		}
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub)
	}
	return sub, nil
}

func (s *PaymentService) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	if s.cache != nil {
		if cached, err := s.cache.GetSubscription(ctx, userID); err == nil && cached.IsActive(time.Now().UTC()) {
			return cached, nil
		}
	}
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub)
	}
	return sub, nil
}

func (s *PaymentService) CanPurchaseMerch(ctx context.Context, userID string) (bool, string) {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil || !sub.IsActive(time.Now().UTC()) {
		return false, "active subscription required to redeem merch"
	}
	return true, ""
}

func (s *PaymentService) GrantSubscription(ctx context.Context, userID, plan string, days int) (*model.Subscription, error) {
	if days <= 0 {
		days = s.subscriptionDays
	}
	amount, err := model.PlanAmountRub(plan)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sub := &model.Subscription{
		ID:        uuid.NewString(),
		UserID:    userID,
		Plan:      plan,
		Status:    model.StatusActive,
		AmountRub: amount,
		StartsAt:  now,
		ExpiresAt: now.Add(time.Duration(days) * 24 * time.Hour),
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub)
	}
	return sub, nil
}

func (s *PaymentService) RevokeSubscription(ctx context.Context, userID string) error {
	if err := s.repo.RevokeSubscription(ctx, userID); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.DeleteSubscription(ctx, userID) // нужно добавить этот метод в кэш
	}
	return nil
}
