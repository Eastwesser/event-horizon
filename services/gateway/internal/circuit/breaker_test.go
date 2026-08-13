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
