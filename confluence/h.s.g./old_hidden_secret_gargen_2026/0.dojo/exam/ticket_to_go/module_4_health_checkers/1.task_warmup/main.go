package main

import (
	"fmt"
	"time"
)

/*
Разминка: select + timeout.

time.After создаёт новый таймер на каждый select — в горячем цикле лучше NewTimer/Reset.
*/
func main() {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(30 * time.Millisecond)
		ch <- "ok"
	}()

	select {
	case result := <-ch:
		fmt.Println("Got result:", result)
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Timeout!")
	}

	slow := make(chan string)
	select {
	case result := <-slow:
		fmt.Println("Got result:", result)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("Timeout!")
	}
}
