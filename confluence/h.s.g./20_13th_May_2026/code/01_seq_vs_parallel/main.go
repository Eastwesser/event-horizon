package main

import (
	"fmt"
	"sync"
	"time"
)

var value int
var mu sync.Mutex

func process() {
	time.Sleep(5 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	value++
}

func main() {
	// Последовательно (медленно) — для демо ускорили sleep.
	start := time.Now()
	value = 0
	for range 10 {
		process()
	}
	fmt.Println("seq value:", value, "took", time.Since(start))

	// Параллельно с WaitGroup + Mutex
	start = time.Now()
	value = 0
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			process()
		}()
	}
	wg.Wait()
	fmt.Println("par value:", value, "took", time.Since(start))
}
