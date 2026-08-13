package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyPaid   = errors.New("payment already completed")
	ErrUnauthorized  = errors.New("webhook unauthorized")
	ErrInvalidStatus = errors.New("invalid payment status")
)

const (
	PlanPresent = "present" // 200 RUB — theme "Present"
	PlanFuture  = "future"  // 300 RUB — theme "Future"

	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusActive    = "active"
	StatusExpired   = "expired"
)

func PlanAmountRub(plan string) (int, error) {
	switch plan {
	case PlanPresent:
		return 200, nil
	case PlanFuture:
		return 300, nil
	default:
		return 0, ErrInvalidInput
	}
}

type Payment struct {
	ID          string
	UserID      string
	Plan        string
	AmountRub   int
	Status      string
	Provider    string
	ProviderRef string
	CheckoutURL string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Subscription struct {
	ID        string
	UserID    string
	Plan      string
	Status    string
	AmountRub int
	PaymentID string
	StartsAt  time.Time
	ExpiresAt time.Time
}

func (s *Subscription) IsActive(now time.Time) bool {
	if s == nil {
		return false
	}
	return s.Status == StatusActive && s.ExpiresAt.After(now)
}
