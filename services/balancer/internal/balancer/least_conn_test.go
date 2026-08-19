package balancer

import (
	"sync/atomic"
	"testing"
)

func TestGetLeastConnBackend_PicksLowest(t *testing.T) {
	lb := NewLeastConnBalancer([]string{
		"http://gw-busy:8080",
		"http://gw-idle:8080",
		"http://gw-mid:8080",
	})
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
	lb := NewLeastConnBalancer(nil)
	if b := lb.getLeastConnBackend(); b != nil {
		t.Fatalf("expected nil, got %v", b.URL)
	}
}
