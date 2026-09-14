// Метафора брокера: sync ждёт аналитику; async кладёт в канал и отвечает сразу.
package main

import (
	"fmt"
	"time"
)

func syncOrder() {
	start := time.Now()
	time.Sleep(80 * time.Millisecond) // медленная аналитика на критическом пути
	fmt.Println("sync: user waited", time.Since(start))
}

func asyncOrder(broker chan<- string) {
	start := time.Now()
	broker <- "track:order-42"
	fmt.Println("async: user waited", time.Since(start), "(accepted)")
}

func main() {
	syncOrder()

	broker := make(chan string, 4)
	go func() {
		for msg := range broker {
			time.Sleep(80 * time.Millisecond)
			fmt.Println("analytics processed", msg)
		}
	}()
	asyncOrder(broker)
	time.Sleep(120 * time.Millisecond)
	close(broker)
}
