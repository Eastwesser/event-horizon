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
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/Eastwesser/event-horizon/services/inventory/migrations"
)

func inventoryDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("INVENTORY_TEST_DATABASE_URL"); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Fatal(err)
		}
		if err := migrator.Up(db, migrations.FS); err != nil {
			_ = db.Close()
			t.Fatalf("migrate: %v", err)
		}
		return db, func() { _ = db.Close() }
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "inventory_test"},
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
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/inventory_test?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatal(err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := migrator.Up(db, migrations.FS); err != nil {
		_ = db.Close()
		_ = c.Terminate(ctx)
		t.Fatalf("migrate: %v", err)
	}
	return db, func() {
		_ = db.Close()
		_ = c.Terminate(ctx)
	}
}

func TestInventory_ReserveAndVersionConflict(t *testing.T) {
	db, cleanup := inventoryDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(db)
	id := uuid.NewString()
	author := uuid.NewString()
	now := time.Now().UTC()
	item := &model.Item{
		ID: id, AuthorID: author, Type: "брелок", Name: "test",
		Description: "d", Price: 10, Stock: 5, Attributes: map[string]any{},
		Images: []string{}, CreatedAt: now, UpdatedAt: now, Version: 1,
	}
	if err := repo.CreateItem(ctx, item); err != nil {
		t.Fatalf("create: %v", err)
	}

	left, err := repo.ReserveItem(ctx, id, 2)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if left != 3 {
		t.Fatalf("stock left=%d want 3", left)
	}

	got, err := repo.GetItem(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version < 2 {
		t.Fatalf("version after reserve=%d want >=2", got.Version)
	}

	// stale version update must conflict
	stale := *got
	stale.Version = 1
	stale.Name = "stale"
	if err := repo.UpdateItem(ctx, &stale); err != model.ErrVersionConflict {
		t.Fatalf("want ErrVersionConflict, got %v", err)
	}

	_, err = repo.ReserveItem(ctx, id, 100)
	if err != model.ErrNotEnoughStock {
		t.Fatalf("want ErrNotEnoughStock, got %v", err)
	}
}
