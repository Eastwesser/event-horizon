package main

import (
	"fmt"
	"sync"
	"time"
)

/*
Конкурентность ≠ Rate Limiter.

Это Worker Pool / counting semaphore: максимум N *одновременно* выполняющихся работ.
Rate Limiter ограничивает *частоту* (запросов в секунду) — см. 4.task_pattern.
*/
const (
	totalJobs = 50
	limit     = 5
)

func work(id int) {
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("done %d\n", id)
}

func main() {
	sem := make(chan struct{}, limit) // слоты
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < totalJobs; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }() // Release
			work(id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Took: %v (ожидай ~%dms при 50ms/job и limit=%d)\n",
		time.Since(start), totalJobs/limit*50, limit)
}
