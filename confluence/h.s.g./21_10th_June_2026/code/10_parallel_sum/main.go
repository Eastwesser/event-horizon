package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func NetworkRequest() int {
	time.Sleep(time.Millisecond)
	return rand.Intn(100)
}

func main() {
	const n = 1000
	var total int64
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			atomic.AddInt64(&total, int64(NetworkRequest()))
		}()
	}
	wg.Wait()
	fmt.Println("sum:", total)
}
