package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

/*
Health Checker + exponential backoff при ошибках.

Идея: пока unhealthy — не долбим endpoint каждые 50ms;
interval *= 2 до max. После успеха — сброс к base interval.

На собесе: backoff экономит зависимости и свой CPU; для liveness
внутреннего процесса backoff обычно не нужен (локальный /health).
*/
type BackoffChecker struct {
	url      string
	base     time.Duration
	max      time.Duration
	timeout  time.Duration
	status   bool
	mu       sync.RWMutex
	client   *http.Client
}

func (bc *BackoffChecker) Healthy() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.status
}

func (bc *BackoffChecker) set(ok bool) {
	bc.mu.Lock()
	bc.status = ok
	bc.mu.Unlock()
}

func (bc *BackoffChecker) once(parent context.Context) bool {
	ctx, cancel := context.WithTimeout(parent, bc.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bc.url, nil)
	if err != nil {
		bc.set(false)
		return false
	}
	resp, err := bc.client.Do(req)
	if err != nil {
		bc.set(false)
		return false
	}
	defer resp.Body.Close()
	ok := resp.StatusCode == 200
	bc.set(ok)
	return ok
}

func (bc *BackoffChecker) Start(ctx context.Context) {
	wait := bc.base
	for {
		ok := bc.once(ctx)
		if ok {
			wait = bc.base
		} else {
			wait *= 2
			if wait > bc.max {
				wait = bc.max
			}
		}
		fmt.Printf("healthy=%v next_wait=%v\n", ok, wait)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func main() {
	var n int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n++
		if n <= 3 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	bc := &BackoffChecker{
		url:     srv.URL,
		base:    30 * time.Millisecond,
		max:     200 * time.Millisecond,
		timeout: 40 * time.Millisecond,
		client:  &http.Client{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 450*time.Millisecond)
	defer cancel()
	bc.Start(ctx)
	fmt.Println("final healthy:", bc.Healthy())
}
