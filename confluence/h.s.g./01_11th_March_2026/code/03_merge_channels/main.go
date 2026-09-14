// merge: N входных каналов → один выходной; закрытие после WaitGroup.
package main

import (
	"fmt"
	"sync"
)

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, ch := range cs {
		ch := ch // capture (безопасно и до Go 1.22)
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func gen(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func main() {
	for v := range merge(gen(1, 2), gen(3), gen(4, 5, 6)) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
