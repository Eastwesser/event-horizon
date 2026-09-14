package main

import (
	"fmt"
	"sync"
)

func concurrentSquares(numbers []int) {
	var wg sync.WaitGroup
	wg.Add(len(numbers))

	for _, n := range numbers {
		go func() {
			defer wg.Done()
			square := n * n
			fmt.Println(square)
		}()
	}

	wg.Wait()
}

func main() {
	numbers := []int{2, 4, 6, 8, 10}
	concurrentSquares(numbers)
}
