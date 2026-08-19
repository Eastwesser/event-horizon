//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
	"github.com/Eastwesser/event-horizon/services/payment/internal/repository"
	"github.com/Eastwesser/event-horizon/services/payment/migrations"
)

func paymentPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("PAYMENT_TEST_DATABASE_URL"); dsn != "" {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("connect env dsn: %v", err)
		}
		if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
			pool.Close()
			t.Fatalf("migrate: %v", err)
		}
		return pool, func() { pool.Close() }
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "payment_test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Skipf("testcontainers postgres unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/payment_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("pool: %v", err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
		pool.Close()
		_ = c.Terminate(ctx)
		t.Fatalf("migrate: %v", err)
	}
	return pool, func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
}

func TestPayment_CompleteActivatesSubscription(t *testing.T) {
	pool, cleanup := paymentPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(pool)
	now := time.Now().UTC()
	userID := uuid.NewString()
	pay := &model.Payment{
		ID: uuid.NewString(), UserID: userID, Plan: model.PlanPresent, AmountRub: 200,
		Status: model.StatusPending, Provider: "boosty", CheckoutURL: "https://example.test/pay",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.CreatePayment(ctx, pay); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	sub := &model.Subscription{
		ID: uuid.NewString(), UserID: userID, Plan: model.PlanPresent, Status: model.StatusActive,
		AmountRub: 200, PaymentID: pay.ID, StartsAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour),
	}
	err := repo.CompletePaymentAndActivateSubscription(ctx, pay.ID, "boosty-ref-1", sub, "payment.completed", map[string]any{
		"user_id": userID, "plan": model.PlanPresent,
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	got, err := repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		t.Fatalf("get sub: %v", err)
	}
	if got.Plan != model.PlanPresent || !got.IsActive(time.Now().UTC()) {
		t.Fatalf("unexpected sub: %+v", got)
	}
}
