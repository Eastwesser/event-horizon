package main

import (
	"fmt"
	"sync"
	"time"
)

type Semaphor struct {
	buffer chan struct{}
}

func (s *Semaphor) Release() {
	s.buffer <- struct{}{}
}

func (s *Semaphor) Acquire() {
	<-s.buffer
}

func NewSemaphor(bufferSize int) *Semaphor {
	return &Semaphor{
		buffer: make(chan struct{}, bufferSize),
	}
}

func main() {
	fmt.Println("")
	sema := NewSemaphor(concurrencyLimit)

	// 1. Подготовка данных (просто список ID)
	urls := make([]int, totalURLs)

	for i := 0; i < totalURLs; i++ {
		urls[i] = i
	}

	fmt.Println("Start downloading...")
	start := time.Now()

	wg := sync.WaitGroup{}
	n := len(urls)

	for i := 0; i < n; i++ {
		sema.Release()
		wg.Add(1)

		go func(idx int) {
			defer func() {
				sema.Acquire()
				wg.Done()
			}()
			download(urls[idx])
		}(i)
	}

	wg.Wait()

	// Реализовать запуск download(url) параллельно, но с ограничением в concurrency limit 5 одновременных запросов

	fmt.Printf("Took: %v\n", time.Since(start))
}
