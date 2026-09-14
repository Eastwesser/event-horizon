package main

import (
	"fmt"
	"sync"
	"time"
)

// EventBus — pub/sub; Publish не блокируется на медленном подписчике (буфер + drop/async).
type EventBus struct {
	mu   sync.RWMutex
	subs map[string][]chan string
}

func NewEventBus() *EventBus {
	return &EventBus{subs: make(map[string][]chan string)}
}

func (b *EventBus) Subscribe(topic string) <-chan string {
	ch := make(chan string, 16)
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Publish(topic, data string) {
	b.mu.RLock()
	subs := append([]chan string{}, b.subs[topic]...)
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- data:
		default:
			// не блокируем остальных
		}
	}
}

func main() {
	bus := NewEventBus()
	a := bus.Subscribe("orders")
	b := bus.Subscribe("orders")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("A:", <-a)
	}()
	go func() {
		defer wg.Done()
		fmt.Println("B:", <-b)
	}()
	time.Sleep(10 * time.Millisecond)
	bus.Publish("orders", "created")
	wg.Wait()
}
