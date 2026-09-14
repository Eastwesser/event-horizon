package main

import (
	"fmt"
	"sync"
	"time"
)

func merge(cs ...<-chan int) <-chan int {
	buffer := len(cs)
	out := make(chan int, buffer)
	ch := make(chan int)
	defer close(ch)

	wg := sync.WaitGroup{}
	wg.Add(len(cs))

	for _, ch := range cs {
		go func() {
			defer wg.Done()
			for num := range <-ch {
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
			time.Sleep(time.Duration(n) * 10 * time.Millisecond) // имитация разной скорости
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
