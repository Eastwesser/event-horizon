package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
	"github.com/Eastwesser/event-horizon/services/payment/internal/repository"
)

type PaymentService struct {
	repo              *repository.PostgresRepo
	cache             *repository.RedisRepo
	boostyURL         string
	webhookSecret     string
	subscriptionDays  int
}

func New(
	repo *repository.PostgresRepo,
	cache *repository.RedisRepo,
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
	// Happy path: activate subscription in a single DB transaction.
	if err := s.repo.CompletePaymentAndActivateSubscription(ctx, paymentID, providerRef, sub, "payment.completed", event); err != nil {
		// Idempotency: Boosty may retry the same delivery, and we already handled this `payment_id`.
		// In that case, return the already-active subscription (so merch unlock remains correct).
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
