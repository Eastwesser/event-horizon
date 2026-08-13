//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/migrations"
)

func dsn(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL (or DATABASE_URL) to run auth integration tests")
	}
	return dsn
}

func TestUserRepository_CreateGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewPostgresUserRepo(pool)
	email := "w4-integration-" + time.Now().Format("150405.000") + "@example.com"
	id, err := repo.Create(ctx, email, "hash", "user")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.ID != id || got.Email != email {
		t.Fatalf("unexpected user: %+v id=%s", got, id)
	}
}
