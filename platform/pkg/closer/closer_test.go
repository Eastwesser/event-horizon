package closer

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestCloseAllOrderAndOnce(t *testing.T) {
	c := New(slog.Default())
	var order []string
	c.AddNamed("a", func(context.Context) error {
		order = append(order, "a")
		return nil
	})
	c.AddNamed("b", func(context.Context) error {
		order = append(order, "b")
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.CloseAll(ctx); err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "b" || order[1] != "a" {
		t.Fatalf("expected LIFO order b,a got %v", order)
	}
	// second call is no-op
	if err := c.CloseAll(ctx); err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 {
		t.Fatalf("CloseAll should run once, got order %v", order)
	}
}

func TestCloseAllJoinsErrors(t *testing.T) {
	c := New(slog.Default())
	c.AddNamed("x", func(context.Context) error { return errors.New("boom") })
	c.AddNamed("y", func(context.Context) error { return nil })
	err := c.CloseAll(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
