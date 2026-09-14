package main

import (
	"fmt"
	"sync"
	"time"
)

// merge сливает N каналов в один. Закрывает out, когда все входы исчерпаны.
// На собесе: fan-in + WaitGroup + close(out) только писателем после Wait.
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		c := c // capture (до Go 1.22 обязательно; после — всё равно явно полезно)
		go func() {
			defer wg.Done()
			for num := range c { // range по каналу, НЕ range <-c
				out <- num
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func asChan(nums ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, n := range nums {
			c <- n
			time.Sleep(time.Duration(n) * 10 * time.Millisecond)
		}
		close(c)
	}()
	return c
}

func main() {
	a := asChan(10, 30, 50)
	b := asChan(20, 40, 60)
	c := asChan(70, 80)

	for n := range merge(a, b, c) {
		fmt.Println(n)
	}
	fmt.Println("Done!")
}
