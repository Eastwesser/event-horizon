package main

import (
	"fmt"
	"sync"
)

// time.Sleep использовать нельзя. это будет не валидным ответом на собеседовании
// Расскажи подробно что происходит
// Как сделать так, чтобы работало оставив канал небуферезированным?
func main() {
	ch := make(chan int)
	wg := &sync.WaitGroup{}
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func(v int) {
			defer wg.Done()
			ch <- v * v
		}(i)
	}
	wg.Wait()
	var sum int
	for v := range ch {
		sum += v
	}
	fmt.Printf("result: %d\n", sum)
}
