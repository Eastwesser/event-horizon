//go:build integration

package repository_test

import (
	"context"
	"database/sql"
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
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	"github.com/Eastwesser/event-horizon/services/billing/migrations"
)

func billingPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("BILLING_TEST_DATABASE_URL"); dsn != "" {
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
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "billing_test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("testcontainers postgres unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/billing_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("pool: %v", err)
	}
	// wait ready
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
	cleanup := func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
	return pool, cleanup
}

func TestBilling_AddSpend_OutboxAndOptimisticVersion(t *testing.T) {
	pool, cleanup := billingPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresBillingRepo(pool)
	userID := uuid.NewString()

	bal, err := repo.AddBalance(ctx, userID, repository.Tickets, 100, "test_credit", "ref-add-1")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if bal != 100 {
		t.Fatalf("balance=%d want 100", bal)
	}

	var version int
	if err := pool.QueryRow(ctx,
		`SELECT version FROM user_currencies WHERE user_id=$1 AND currency_type=$2`,
		userID, repository.Tickets,
	).Scan(&version); err != nil {
		t.Fatalf("version: %v", err)
	}
	if version < 1 {
		t.Fatalf("version=%d", version)
	}

	var outbox int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox WHERE processed=false`).Scan(&outbox); err != nil {
		t.Fatalf("outbox: %v", err)
	}
	if outbox < 1 {
		t.Fatal("expected outbox row after AddBalance")
	}

	bal, err = repo.SpendBalance(ctx, userID, repository.Tickets, 40, "test_spend", "ref-spend-1")
	if err != nil {
		t.Fatalf("spend: %v", err)
	}
	if bal != 60 {
		t.Fatalf("balance=%d want 60", bal)
	}

	_, err = repo.SpendBalance(ctx, userID, repository.Tickets, 1000, "over", "ref-over")
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestBilling_ConcurrentSpend_NoNegative(t *testing.T) {
	pool, cleanup := billingPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresBillingRepo(pool)
	userID := uuid.NewString()
	if _, err := repo.AddBalance(ctx, userID, repository.Tickets, 50, "seed", "ref-seed"); err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			_, err := repo.SpendBalance(ctx, userID, repository.Tickets, 40, "race", fmt.Sprintf("ref-race-%d", i))
			errCh <- err
		}()
	}
	var fails int
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			fails++
		}
	}
	if fails != 1 {
		t.Fatalf("exactly one spend should fail under race, fails=%d", fails)
	}
	bal, err := repo.GetBalance(ctx, userID, repository.Tickets)
	if err != nil && err != sql.ErrNoRows {
		// GetBalance returns 0,nil on no rows; pgx may differ — just read
	}
	_ = err
	if bal < 0 {
		t.Fatalf("negative balance %d", bal)
	}
	if bal != 10 {
		// 50-40=10 if one succeeded
		got, _ := repo.GetBalance(ctx, userID, repository.Tickets)
		if got != 10 {
			t.Fatalf("balance=%d want 10", got)
		}
	}
}
