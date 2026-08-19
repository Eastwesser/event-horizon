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
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"github.com/Eastwesser/event-horizon/services/shop/migrations"
)

func shopDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("SHOP_TEST_DATABASE_URL"); dsn != "" {
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
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "shop_test"},
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
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/shop_test?sslmode=disable", host, port.Port())
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

func TestShop_PurchaseItemWithStock_AndVersionBump(t *testing.T) {
	db, cleanup := shopDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresShopRepo(db)

	itemID := uuid.NewString()
	userID := uuid.NewString()
	_, err := db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, price, category, available, stock, version)
		VALUES ($1, 'merch hat', 't', 50, 'merch', true, 2, 1)`, itemID)
	if err != nil {
		t.Fatalf("seed item: %v", err)
	}

	if err := repo.PurchaseItemWithStock(ctx, userID, itemID, 50, &repository.OutboxRecord{
		EventType: "shop.purchased",
		Payload:   []byte(`{"user_id":"` + userID + `","item_id":"` + itemID + `"}`),
	}); err != nil {
		t.Fatalf("purchase: %v", err)
	}

	var outboxN int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox WHERE event_type='shop.purchased' AND processed=false`).Scan(&outboxN); err != nil {
		t.Fatalf("outbox count: %v", err)
	}
	if outboxN != 1 {
		t.Fatalf("outbox rows=%d want 1", outboxN)
	}

	var stock, version int
	if err := db.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id=$1`, itemID).Scan(&stock, &version); err != nil {
		t.Fatal(err)
	}
	if stock != 1 {
		t.Fatalf("stock=%d want 1", stock)
	}
	if version != 2 {
		t.Fatalf("version=%d want 2", version)
	}

	var owned int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory WHERE user_id=$1 AND item_id=$2`, userID, itemID).Scan(&owned)
	if owned != 1 {
		t.Fatalf("owned=%d", owned)
	}

	// exhaust stock
	u2 := uuid.NewString()
	if err := repo.PurchaseItemWithStock(ctx, u2, itemID, 50, nil); err != nil {
		t.Fatalf("second purchase: %v", err)
	}
	err = repo.PurchaseItemWithStock(ctx, uuid.NewString(), itemID, 50, nil)
	if err == nil {
		t.Fatal("expected out of stock")
	}
}
