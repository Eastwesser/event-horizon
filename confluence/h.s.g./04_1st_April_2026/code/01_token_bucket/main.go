// Token bucket rate limiter per IP (условие из 04_full_task).
// Lazy refill: tokens += elapsed * rate, capped by maxTokens.
package main

import (
	"fmt"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	rate      float64
	maxTokens float64
	buckets   map[string]*bucket
	ttl       time.Duration
}

func NewRateLimiter(rate, maxTokens float64) *RateLimiter {
	return &RateLimiter{
		rate:      rate,
		maxTokens: maxTokens,
		buckets:   make(map[string]*bucket),
		ttl:       2 * time.Minute,
	}
}

func (rl *RateLimiter) CanAccept(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &bucket{tokens: rl.maxTokens - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(b.tokens+elapsed*rl.rate, rl.maxTokens)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Cleanup удаляет бакеты, к которым давно не обращались.
func (rl *RateLimiter) Cleanup(now time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for ip, b := range rl.buckets {
		if now.Sub(b.last) > rl.ttl {
			delete(rl.buckets, ip)
		}
	}
}

func main() {
	rl := NewRateLimiter(2, 3) // 2 token/s, burst 3
	ip := "1.2.3.4"
	for i := 0; i < 5; i++ {
		fmt.Printf("req %d: %v\n", i+1, rl.CanAccept(ip))
	}
	time.Sleep(600 * time.Millisecond) // ~1.2 tokens
	fmt.Println("after sleep:", rl.CanAccept(ip))
	rl.Cleanup(time.Now())
}
