package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	ch1 := make(chan int, 10)
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, i := range nums {
		wg.Add(1)

		go func() {
			defer wg.Done()
			ch1 <- i
		}()

	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	for v := range ch1 {
		fmt.Println(v)
	}
}
