// Есть 100 url'ов. Нужно скачать их все, но одновременно может выполняться не более 5 запросов.
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	totalURLs        = 100
	concurrencyLimit = 5
)

// Semaphore — counting semaphore на chan struct{}.
type Semaphore struct {
	slots chan struct{}
}

func NewSemaphore(limit int) *Semaphore {
	s := &Semaphore{slots: make(chan struct{}, limit)}
	for i := 0; i < limit; i++ {
		s.slots <- struct{}{}
	}
	return s
}

func (s *Semaphore) Acquire() { <-s.slots }
func (s *Semaphore) Release() { s.slots <- struct{}{} }

func download(url int) {
	time.Sleep(time.Duration(20+url%5*10) * time.Millisecond)
	fmt.Printf("Downloaded: %d\n", url)
}

func main() {
	urls := make([]int, totalURLs)
	for i := 0; i < totalURLs; i++ {
		urls[i] = i
	}

	fmt.Println("Start downloading...")
	start := time.Now()

	sema := NewSemaphore(concurrencyLimit)
	var wg sync.WaitGroup

	for i := 0; i < len(urls); i++ {
		wg.Add(1)
		sema.Acquire()
		go func(idx int) {
			defer wg.Done()
			defer sema.Release()
			download(urls[idx])
		}(i)
	}

	wg.Wait()
	fmt.Printf("Took: %v\n", time.Since(start))
}
