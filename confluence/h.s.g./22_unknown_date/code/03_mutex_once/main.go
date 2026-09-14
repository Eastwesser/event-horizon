package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		mu    sync.Mutex
		rw    sync.RWMutex
		once  sync.Once
		value int
	)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(func() {
				fmt.Println("init once")
				value = 42
			})
			mu.Lock()
			value++
			mu.Unlock()
			rw.RLock()
			_ = value
			rw.RUnlock()
		}()
	}
	wg.Wait()
	fmt.Println("value:", value)
}
