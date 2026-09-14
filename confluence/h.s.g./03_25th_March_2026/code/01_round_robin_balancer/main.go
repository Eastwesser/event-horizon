// Client-side Round Robin balancer + retry на другой инстанс + per-try timeout.
// Чинит типичные баги черновика: shadow b, curr++, defer cancel в loop.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Request any
type Response any

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
}

type BackendImpl struct {
	addr    string
	failOdd atomic.Bool
	calls   atomic.Int32
}

func NewBackend(addr string, failFirst bool) *BackendImpl {
	b := &BackendImpl{addr: addr}
	b.failOdd.Store(failFirst)
	return b
}

func (b *BackendImpl) Invoke(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := b.calls.Add(1)
	if b.failOdd.Load() && n == 1 {
		return nil, errors.New("unavailable: " + b.addr)
	}
	return fmt.Sprintf("ok from %s req=%v", b.addr, req), nil
}

type Balancer struct {
	backends []Backend
	next     atomic.Uint64
}

func NewBalancer(backends []Backend) *Balancer {
	return &Balancer{backends: backends}
}

func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
	n := len(b.backends)
	if n == 0 {
		return nil, errors.New("no backends")
	}
	start := int(b.next.Add(1)-1) % n
	var last error
	for i := 0; i < n; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		idx := (start + i) % n
		tryCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		resp, err := b.backends[idx].Invoke(tryCtx, req)
		cancel() // сразу, не defer в цикле
		if err == nil {
			return resp, nil
		}
		last = err
	}
	return nil, fmt.Errorf("all backends failed: %w", last)
}

func main() {
	backends := []Backend{
		NewBackend("a:1", true),
		NewBackend("b:2", false),
		NewBackend("c:3", false),
	}
	lb := NewBalancer(backends)
	for i := 0; i < 5; i++ {
		resp, err := lb.Invoke(context.Background(), i)
		fmt.Println(resp, err)
	}
}
