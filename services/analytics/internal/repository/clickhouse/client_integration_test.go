//go:build integration

package clickhouse_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/repository"
	ch "github.com/Eastwesser/event-horizon/services/analytics/internal/repository/clickhouse"
)

func clickhouseURL(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	if u := os.Getenv("ANALYTICS_TEST_CLICKHOUSE_URL"); u != "" {
		return u, func() {}
	}

	req := testcontainers.ContainerRequest{
		Image: "clickhouse/clickhouse-server:24.8-alpine",
		Env: map[string]string{
			"CLICKHOUSE_SKIP_USER_SETUP": "1",
		},
		ExposedPorts: []string{"8123/tcp"},
		WaitingFor:   wait.ForHTTP("/ping").WithPort("8123/tcp").WithStartupTimeout(90 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Skipf("testcontainers clickhouse unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "8123")
	return fmt.Sprintf("http://%s:%s", host, port.Port()), func() { _ = c.Terminate(ctx) }
}

func TestAnalytics_InsertAndDAU(t *testing.T) {
	base, cleanup := clickhouseURL(t)
	defer cleanup()

	ctx := context.Background()
	client := ch.New(base, "eventhorizon")
	if err := client.EnsureSchema(ctx); err != nil {
		t.Fatalf("schema: %v", err)
	}
	repo := repository.New(client, "eventhorizon")
	if err := repo.Record(ctx, "user-a", "login", `{}`); err != nil {
		t.Fatalf("record: %v", err)
	}

	// MergeTree inserts are usually visible immediately over HTTP, but wait briefly.
	deadline := time.Now().Add(5 * time.Second)
	var mau int64
	var err error
	for time.Now().Before(deadline) {
		mau, err = repo.MAU(ctx, 7)
		if err == nil && mau >= 1 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("mau: %v", err)
	}
	if mau < 1 {
		t.Fatalf("mau=%d want >=1", mau)
	}

	days, err := repo.DAU(ctx, 7)
	if err != nil {
		t.Fatalf("dau: %v", err)
	}
	if len(days) < 1 {
		t.Fatalf("dau empty: %+v", days)
	}
}
