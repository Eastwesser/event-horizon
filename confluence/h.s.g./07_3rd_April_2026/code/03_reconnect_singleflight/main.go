// Thundering herd при reconnect: singleflight на loadHistory + простой rate limit.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type flight struct {
	mu sync.Mutex
	m  map[string]*call
}

type call struct {
	done chan struct{}
	val  string
}

func (f *flight) Do(key string, fn func() string) string {
	f.mu.Lock()
	if c, ok := f.m[key]; ok {
		f.mu.Unlock()
		<-c.done
		return c.val
	}
	c := &call{done: make(chan struct{})}
	f.m[key] = c
	f.mu.Unlock()
	c.val = fn()
	close(c.done)
	f.mu.Lock()
	delete(f.m, key)
	f.mu.Unlock()
	return c.val
}

func main() {
	var dbHits atomic.Int32
	f := &flight{m: make(map[string]*call)}
	limit := make(chan struct{}, 5) // max 5 concurrent reconnects

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()
			hist := f.Do("chat-7", func() string {
				dbHits.Add(1)
				time.Sleep(20 * time.Millisecond)
				return "history"
			})
			_ = hist
		}()
	}
	wg.Wait()
	fmt.Println("reconnects=50, dbHits=", dbHits.Load()) // ≪ 50
}
