package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
Yandex LB (из YANDEX_LB_TASK.md):
Task 1 — балансировщик, приоритет наименее нагруженному (least connections), не RR.
Task 2 идея: после 3 фейлов исключать backend (здесь: временный quarantine).
*/
type Request any
type Response string

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
	Addr() string
}

type BackendImpl struct {
	addr    string
	inflight atomic.Int64
	fails    atomic.Int32
	deadUntil atomic.Int64 // unix nano
}

func NewBackend(addr string) *BackendImpl { return &BackendImpl{addr: addr} }
func (b *BackendImpl) Addr() string       { return b.addr }

func (b *BackendImpl) Invoke(ctx context.Context, req Request) (Response, error) {
	b.inflight.Add(1)
	defer b.inflight.Add(-1)
	time.Sleep(5 * time.Millisecond) // демо: иначе least-conn всегда берёт первый (inflight=0)
	// demo: "bad" address fails
	if b.addr == "bad:1" {
		b.fails.Add(1)
		return "", fmt.Errorf("backend %s down", b.addr)
	}
	b.fails.Store(0)
	return Response("ok@" + b.addr), nil
}

func (b *BackendImpl) healthy() bool {
	return time.Now().UnixNano() >= b.deadUntil.Load()
}

type Balancer struct {
	backs []*BackendImpl
}

func NewBalancer(addrs []string) *Balancer {
	b := &Balancer{backs: make([]*BackendImpl, len(addrs))}
	for i, a := range addrs {
		b.backs[i] = NewBackend(a)
	}
	return b
}

func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
	var (
		chosen *BackendImpl
		best   int64 = 1<<63 - 1
	)
	for _, back := range b.backs {
		if !back.healthy() {
			continue
		}
		n := back.inflight.Load()
		if n < best {
			best = n
			chosen = back
		}
	}
	if chosen == nil {
		return "", fmt.Errorf("no healthy backends")
	}
	resp, err := chosen.Invoke(ctx, req)
	if err != nil && chosen.fails.Load() >= 3 {
		chosen.deadUntil.Store(time.Now().Add(200 * time.Millisecond).UnixNano())
		fmt.Println("quarantine", chosen.addr)
	}
	return resp, err
}

func main() {
	lb := NewBalancer([]string{"a:1", "b:1", "bad:1"})
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := lb.Invoke(context.Background(), i)
			fmt.Println(i, resp, err)
		}(i)
	}
	wg.Wait()
}
