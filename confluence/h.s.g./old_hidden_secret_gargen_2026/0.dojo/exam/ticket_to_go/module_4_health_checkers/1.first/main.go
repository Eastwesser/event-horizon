package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"
)

/*
Health Checker: ticker + per-check WithTimeout + RWMutex.

EH: /health (liveness) vs /ready (dependency ping).
*/
type HealthChecker struct {
	url      string
	interval time.Duration
	timeout  time.Duration
	status   bool
	mu       sync.RWMutex
	client   *http.Client
}

func NewHealthChecker(url string, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		url:      url,
		interval: interval,
		timeout:  timeout,
		client:   &http.Client{},
	}
}

func (hc *HealthChecker) Healthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

func (hc *HealthChecker) check(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hc.url, nil)
	ok := false
	if err == nil {
		resp, err := hc.client.Do(req)
		if err == nil {
			ok = resp.StatusCode == http.StatusOK
			_ = resp.Body.Close()
		}
	}
	hc.mu.Lock()
	hc.status = ok
	hc.mu.Unlock()
}

func (hc *HealthChecker) Start(ctx context.Context, onCheck func()) {
	t := time.NewTicker(hc.interval)
	defer t.Stop()
	hc.check(ctx)
	if onCheck != nil {
		onCheck()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			hc.check(ctx)
			if onCheck != nil {
				onCheck()
			}
		}
	}
}

func main() {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hc := NewHealthChecker(srv.URL, 50*time.Millisecond, 40*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var n int
	done := make(chan struct{})
	go hc.Start(ctx, func() {
		n++
		fmt.Printf("check#%d healthy=%v\n", n, hc.Healthy())
		if n >= 5 {
			cancel()
			close(done)
		}
	})
	<-done
}
