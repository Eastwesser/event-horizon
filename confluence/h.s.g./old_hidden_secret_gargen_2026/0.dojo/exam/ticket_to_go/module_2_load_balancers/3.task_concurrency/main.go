package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/*
Worker Pool + graceful shutdown.

На собесе:
1) jobs chan → N воркеров читают range jobs
2) close(jobs) после отправки → воркеры выходят из range
3) wg.Wait() ждёт доработку
4) context — для отмены «посередине» (не путать с close(jobs))
*/
func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d: canceled\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("worker %d: jobs closed\n", id)
				return
			}
			time.Sleep(30 * time.Millisecond)
			fmt.Printf("worker %d: job %d\n", id, job)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int, 8)
	var wg sync.WaitGroup
	const workers = 3

	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, &wg)
	}

	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // graceful: больше задач нет

	wg.Wait()
	fmt.Println("pool drained")
}
