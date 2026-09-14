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
Конкурентность: несколько сервисов — у каждого свой probe-loop.
Общий ctx для stop; статус под RWMutex.
*/
type probe struct {
	name string
	url  string
	ok   bool
	mu   sync.RWMutex
}

func (p *probe) set(ok bool) {
	p.mu.Lock()
	p.ok = ok
	p.mu.Unlock()
}

func (p *probe) get() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ok
}

func (p *probe) checkOnce(parent context.Context, timeout time.Duration, client *http.Client) {
	cctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, p.url, nil)
	if err != nil {
		p.set(false)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		p.set(false)
		return
	}
	_ = resp.Body.Close()
	p.set(resp.StatusCode == http.StatusOK)
}

func (p *probe) loop(ctx context.Context, interval, timeout time.Duration, client *http.Client, ready *sync.WaitGroup) {
	p.checkOnce(ctx, timeout, client)
	ready.Done()

	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.checkOnce(ctx, timeout, client)
		}
	}
}

func main() {
	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okSrv.Close()
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badSrv.Close()

	probes := []*probe{
		{name: "billing", url: okSrv.URL},
		{name: "payments", url: badSrv.URL},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &http.Client{}
	var ready sync.WaitGroup
	ready.Add(len(probes))
	for _, p := range probes {
		p := p
		go p.loop(ctx, 40*time.Millisecond, 100*time.Millisecond, client, &ready)
	}
	ready.Wait()

	for _, p := range probes {
		fmt.Printf("%s healthy=%v\n", p.name, p.get())
	}
	cancel()
}
