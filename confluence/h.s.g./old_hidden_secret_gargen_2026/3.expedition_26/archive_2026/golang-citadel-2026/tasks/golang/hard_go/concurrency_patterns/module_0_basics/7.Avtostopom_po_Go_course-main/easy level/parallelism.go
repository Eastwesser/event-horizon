package main

import (
	"fmt"
	"time"
)

// time.Sleep использовать нельзя. это будет не валидным ответом на собеседовании
// 1
// Что выведет код и почему?
// Как исправить?
func main() {
	var counter int
	for i := 0; i < 1000; i++ {
		go func() {
			counter++
		}()
	}
	fmt.Println(counter)
}

// 2
// Что выведет код и почему?
// Как исправить оставив канал небуферезированным?
func main() {
	ch := make(chan int)
	ch <- 1
	go func() {
		fmt.Println(<-ch)
	}()
}

// 3
// Расскажи подробно что происходит?(спойлер: почему будет panic?)
// Как сделать так, чтобы работало?
func main() {
	x := make(map[int]int, 1)
	go func() { x[1] = 2 }()
	go func() { x[3] = 7 }()
	go func() { x[123] = 10 }()
	go func() { x[1] = 2 }()
	go func() { x[34] = 7 }()
	go func() { x[1432] = 10 }()
	go func() { x[1] = 2 }()
	go func() { x[100] = 7 }()
	go func() { x[34] = 10 }()
	go func() { x[1] = 2 }()
	time.Sleep(100 * time.Millisecond) //блокируемся на 100 миллисекунд
	fmt.Println("x[1] =", x[1])
}
