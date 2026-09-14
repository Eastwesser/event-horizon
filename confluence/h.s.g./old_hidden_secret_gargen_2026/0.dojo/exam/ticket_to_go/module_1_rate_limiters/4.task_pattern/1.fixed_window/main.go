package main

import (
	"fmt"
	"sync"
	"time"
)

/*
Fixed Window: лимит N запросов на окно длиной window.

Плюс: просто.
Минус (скажи на собесе): burst на границе — в конце окна + в начале следующего
можно пропустить почти 2N за короткое время.
*/
type FixedWindow struct {
	limit  int
	window time.Duration
	count  int
	start  time.Time
	mu     sync.Mutex
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:  limit,
		window: window,
		start:  time.Now(),
	}
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	if now.Sub(fw.start) >= fw.window {
		fw.start = now
		fw.count = 0
	}
	if fw.count >= fw.limit {
		return false
	}
	fw.count++
	return true
}

func main() {
	rl := NewFixedWindow(3, time.Second)
	for i := 0; i < 8; i++ {
		ok := rl.Allow()
		fmt.Printf("req %d: %v\n", i, ok)
		time.Sleep(250 * time.Millisecond)
	}
}
