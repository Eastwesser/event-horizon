package balancer

import (
	"log/slog"
	"sync/atomic"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestGetLeastConnBackend_PicksLowest(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-busy:8080",
		"http://gw-idle:8080",
		"http://gw-mid:8080",
	}, testLogger())
	if got := len(lb.backends); got != 3 {
		t.Fatalf("backends=%d want 3", got)
	}

	atomic.StoreInt32(&lb.backends[0].ActiveConns, 10)
	atomic.StoreInt32(&lb.backends[1].ActiveConns, 1)
	atomic.StoreInt32(&lb.backends[2].ActiveConns, 4)

	got := lb.getLeastConnBackend()
	if got == nil {
		t.Fatal("expected a backend")
	}
	if got.URL.Host != "gw-idle:8080" {
		t.Fatalf("host=%s want gw-idle:8080", got.URL.Host)
	}
}

func TestGetLeastConnBackend_Empty(t *testing.T) {
	lb := NewLeastConnBalancer(nil, testLogger())
	if b := lb.getLeastConnBackend(); b != nil {
		t.Fatalf("expected nil, got %v", b.URL)
	}
}

// TestGetLeastConnBackend_SkipsUnhealthy — все равны, но два помечены unhealthy.
func TestGetLeastConnBackend_SkipsUnhealthy(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-1:8080",
		"http://gw-2:8080",
		"http://gw-3:8080",
	}, testLogger())

	lb.backends[0].Healthy.Store(false)
	lb.backends[1].Healthy.Store(false)
	// только gw-3 healthy

	got := lb.getLeastConnBackend()
	if got == nil {
		t.Fatal("expected a backend")
	}
	if got.URL.Host != "gw-3:8080" {
		t.Fatalf("host=%s want gw-3:8080", got.URL.Host)
	}
}

// TestGetLeastConnBackend_AllUnhealthy — все unhealthy → nil.
func TestGetLeastConnBackend_AllUnhealthy(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-1:8080",
		"http://gw-2:8080",
	}, testLogger())

	lb.backends[0].Healthy.Store(false)
	lb.backends[1].Healthy.Store(false)

	if b := lb.getLeastConnBackend(); b != nil {
		t.Fatalf("expected nil, got %v", b.URL)
	}
}

// TestGetLeastConnBackend_Jitter — при равных ActiveConns выбирается не всегда первый.
func TestGetLeastConnBackend_Jitter(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-1:8080",
		"http://gw-2:8080",
		"http://gw-3:8080",
	}, testLogger())

	// все ActiveConns = 0, все Healthy
	seen := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		b := lb.getLeastConnBackend()
		if b == nil {
			t.Fatal("expected a backend")
		}
		seen[b.URL.Host] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("jitter not working: only %d distinct backend(s) in 100 picks", len(seen))
	}
}

func TestHealthyBackends(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-1:8080",
		"http://gw-2:8080",
		"http://gw-3:8080",
	}, testLogger())

	if got := lb.HealthyBackends(); got != 3 {
		t.Fatalf("healthy=%d want 3", got)
	}
	lb.backends[1].Healthy.Store(false)
	if got := lb.HealthyBackends(); got != 2 {
		t.Fatalf("healthy=%d want 2", got)
	}
}