package main

import "fmt"

func main() {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	close(ch) // закрывает ПИСАТЕЛЬ

	for v := range ch {
		fmt.Println("range:", v)
	}
	v, ok := <-ch
	fmt.Println("after close:", v, ok) // 0 false
}
