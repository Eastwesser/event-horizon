// Демо Outbox: заказ и событие в «одной транзакции» (in-memory),
// poller шлёт в брокер (канал), аналитика читает асинхронно.
// На собесе: Junior=go func, Middle=broker, Senior=outbox+SKIP LOCKED+идемпотентность.
package main

import (
	"fmt"
	"sync"
	"time"
)

type OutboxEvent struct {
	ID      int
	Payload string
	Sent    bool
}

var (
	mu     sync.Mutex
	orders []string
	outbox []OutboxEvent
	nextID = 1
)

func createOrder(item string) {
	mu.Lock()
	defer mu.Unlock()
	orders = append(orders, item)
	outbox = append(outbox, OutboxEvent{ID: nextID, Payload: "track:" + item})
	nextID++
	fmt.Println("order OK (user answered immediately):", item)
}

func poller(broker chan<- string, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}
		mu.Lock()
		for i := range outbox {
			if !outbox[i].Sent {
				outbox[i].Sent = true
				payload := outbox[i].Payload
				mu.Unlock()
				broker <- payload // at-least-once: retry if fail
				mu.Lock()
			}
		}
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	broker := make(chan string, 8)
	done := make(chan struct{})
	go poller(broker, done)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			fmt.Println("analytics got:", <-broker)
		}
	}()

	createOrder("sku-1")
	createOrder("sku-2")
	createOrder("sku-3")
	wg.Wait()
	close(done)
}
