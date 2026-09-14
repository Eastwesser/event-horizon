// Fan-out: значения слайса в канал, читаем из main.
package main

import (
	"fmt"
	"sync"
)

func main() {
	s1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ch := make(chan int, len(s1))

	var wg sync.WaitGroup
	workers := 3
	jobs := make(chan int, len(s1))
	for _, v := range s1 {
		jobs <- v
	}
	close(jobs)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range jobs {
				ch <- v // «обработка» = проброс
			}
		}()
	}
	go func() { wg.Wait(); close(ch) }()

	for v := range ch {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
