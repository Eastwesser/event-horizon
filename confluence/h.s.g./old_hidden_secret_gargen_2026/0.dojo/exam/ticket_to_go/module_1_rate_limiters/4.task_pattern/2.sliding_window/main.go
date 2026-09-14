package main

import (
	"fmt"
	"sync"
	"time"
)

/*
Sliding Window Log: храним timestamp каждого запроса, режем всё старше window.

Плюс: нет пограничного burst Fixed Window.
Минус: память O(лимит) на ключ (user/IP); под нагрузкой дороже.
Альтернатива на проде: Sliding Window Counter (приближённо, меньше памяти).
*/
type SlidingWindow struct {
	limit  int
	window time.Duration
	hits   []time.Time
	mu     sync.Mutex
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:  limit,
		window: window,
		hits:   make([]time.Time, 0, limit),
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// drop expired
	i := 0
	for i < len(sw.hits) && sw.hits[i].Before(cutoff) {
		i++
	}
	sw.hits = sw.hits[i:]

	if len(sw.hits) >= sw.limit {
		return false
	}
	sw.hits = append(sw.hits, now)
	return true
}

func main() {
	rl := NewSlidingWindow(3, time.Second)
	for i := 0; i < 8; i++ {
		ok := rl.Allow()
		fmt.Printf("req %d: %v\n", i, ok)
		time.Sleep(250 * time.Millisecond)
	}
}
