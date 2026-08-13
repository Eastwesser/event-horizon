package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreatePayment(ctx context.Context, p *model.Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments (id, user_id, plan, amount_rub, status, provider, checkout_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.UserID, p.Plan, p.AmountRub, p.Status, p.Provider, p.CheckoutURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepo) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, plan, amount_rub, status, provider, COALESCE(provider_ref,''), COALESCE(checkout_url,''), created_at, updated_at
		FROM payments WHERE id = $1`, id)
	var p model.Payment
	if err := row.Scan(&p.ID, &p.UserID, &p.Plan, &p.AmountRub, &p.Status, &p.Provider, &p.ProviderRef, &p.CheckoutURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// CompletePaymentAndActivateSubscription writes payment completion + subscription + outbox in one TX.
func (r *PostgresRepo) CompletePaymentAndActivateSubscription(
	ctx context.Context,
	paymentID, providerRef string,
	sub *model.Subscription,
	eventType string,
	eventPayload map[string]any,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM payments WHERE id = $1 FOR UPDATE`, paymentID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return err
	}
	if status == model.StatusCompleted {
		return model.ErrAlreadyPaid
	}
	if status != model.StatusPending {
		return model.ErrInvalidStatus
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE payments SET status = $2, provider_ref = $3, updated_at = $4 WHERE id = $1`,
		paymentID, model.StatusCompleted, providerRef, now,
	)
	if err != nil {
		return err
	}

	// Expire previous active subscriptions for this user.
	_, err = tx.Exec(ctx, `
		UPDATE subscriptions SET status = $2, updated_at = $3
		WHERE user_id = $1 AND status = $4`,
		sub.UserID, model.StatusExpired, now, model.StatusActive,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, plan, status, amount_rub, payment_id, starts_at, expires_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		sub.ID, sub.UserID, sub.Plan, sub.Status, sub.AmountRub, sub.PaymentID, sub.StartsAt, sub.ExpiresAt, now, now,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("marshal outbox: %w", err)
	}
	outboxID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		outboxID, eventType, body,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) GetActiveSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, plan, status, amount_rub, COALESCE(payment_id::text,''), starts_at, expires_at
		FROM subscriptions
		WHERE user_id = $1 AND status = $2 AND expires_at > NOW()
		ORDER BY expires_at DESC
		LIMIT 1`, userID, model.StatusActive)
	var s model.Subscription
	if err := row.Scan(&s.ID, &s.UserID, &s.Plan, &s.Status, &s.AmountRub, &s.PaymentID, &s.StartsAt, &s.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}
