package main

import (
	"fmt"
	"sync"
	"time"
)

/*
Leaky Bucket: очередь «капель» утекает с постоянной rate.
Входящий burst сглаживается: если очередь полна — drop.

Отличие от Token Bucket (скажи на собесе):
- Token Bucket разрешает burst до capacity, потом ровный rate.
- Leaky Bucket выдаёт наружу строго равномерно (выходной rate фиксирован).
*/
type LeakyBucket struct {
	capacity int
	leakEvery time.Duration
	level    int
	mu       sync.Mutex
	stop     chan struct{}
}

func NewLeakyBucket(capacity int, leakEvery time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		capacity:  capacity,
		leakEvery: leakEvery,
		stop:      make(chan struct{}),
	}
	go lb.leak()
	return lb
}

func (lb *LeakyBucket) leak() {
	t := time.NewTicker(lb.leakEvery)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			lb.mu.Lock()
			if lb.level > 0 {
				lb.level--
			}
			lb.mu.Unlock()
		case <-lb.stop:
			return
		}
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if lb.level >= lb.capacity {
		return false
	}
	lb.level++
	return true
}

func (lb *LeakyBucket) Close() { close(lb.stop) }

func main() {
	lb := NewLeakyBucket(3, 300*time.Millisecond)
	defer lb.Close()

	for i := 0; i < 6; i++ {
		fmt.Printf("burst %d: %v\n", i, lb.Allow())
	}
	time.Sleep(1 * time.Second)
	fmt.Println("after leak:")
	for i := 0; i < 3; i++ {
		fmt.Printf("later %d: %v\n", i, lb.Allow())
	}
}
