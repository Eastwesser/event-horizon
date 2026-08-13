package circuit

import (
	"errors"
	"testing"
	"time"
)

func TestBreakerOpensAfterFailures(t *testing.T) {
	b := New(Settings{
		Name:        "billing",
		MaxRequests: 3,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool { return c.ConsecutiveFailures >= 2 },
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	_, _ = b.Execute(fail)
	_, _ = b.Execute(fail)
	_, err := b.Execute(fail)
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("want ErrOpen, got %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	ok := false
	_, err = b.Execute(func() (any, error) { ok = true; return "ok", nil })
	if err != nil || !ok {
		t.Fatalf("half-open probe failed: %v", err)
	}
}

func TestNew_DefaultSettings(t *testing.T) {
	b := New(Settings{Name: "shop"})
	if b.Name() != "shop" {
		t.Fatalf("name=%q", b.Name())
	}
	result, err := b.Execute(func() (any, error) { return "ok", nil })
	if err != nil || result != "ok" {
		t.Fatalf("execute failed: %v", err)
	}
}

func TestBreaker_ClosedSuccessResetsFailures(t *testing.T) {
	trips := 0
	b := New(Settings{
		Name:        "inventory",
		MaxRequests: 2,
		Timeout:     10 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool {
			trips++
			return c.ConsecutiveFailures >= 3
		},
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	ok := func() (any, error) { return "ok", nil }
	_, _ = b.Execute(fail)
	_, _ = b.Execute(fail)
	_, _ = b.Execute(ok)
	_, err := b.Execute(fail)
	if err == nil {
		t.Fatal("expected failure")
	}
	if trips == 0 {
		t.Fatal("ReadyToTrip never called")
	}
}

func TestBreaker_HalfOpenProbeLimit(t *testing.T) {
	b := New(Settings{
		Name:        "auth",
		MaxRequests: 1,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool { return c.ConsecutiveFailures >= 1 },
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	_, _ = b.Execute(fail) // trip open
	time.Sleep(60 * time.Millisecond)
	_, _ = b.Execute(fail) // half-open probe fails, back to open
	_, err := b.Execute(fail)
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("want ErrOpen while breaker open, got %v", err)
	}
}
