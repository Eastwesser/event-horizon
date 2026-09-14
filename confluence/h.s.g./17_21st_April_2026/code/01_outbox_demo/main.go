// Outbox demo: бизнес-событие + запись в outbox, poller → «Kafka» (канал).
package main

import (
	"fmt"
	"sync"
	"time"
)

type outboxRow struct {
	ID      int
	Payload string
	Sent    bool
}

type store struct {
	mu     sync.Mutex
	orders []string
	outbox []outboxRow
	nextID int
}

func (s *store) createOrder(payload string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders = append(s.orders, payload)
	s.nextID++
	s.outbox = append(s.outbox, outboxRow{ID: s.nextID, Payload: payload})
}

func (s *store) pollOutbox(limit int) []outboxRow {
	s.mu.Lock()
	defer s.mu.Unlock()
	var batch []outboxRow
	for i := range s.outbox {
		if s.outbox[i].Sent {
			continue
		}
		batch = append(batch, s.outbox[i])
		if len(batch) >= limit {
			break
		}
	}
	return batch
}

func (s *store) ack(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.outbox {
		if s.outbox[i].ID == id {
			s.outbox[i].Sent = true
			return
		}
	}
}

func main() {
	s := &store{}
	broker := make(chan string, 8)

	go func() { // poller
		for {
			batch := s.pollOutbox(2)
			if len(batch) == 0 {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			for _, row := range batch {
				broker <- row.Payload
				s.ack(row.ID)
			}
		}
	}()

	s.createOrder("order-1")
	s.createOrder("order-2")

	for i := 0; i < 2; i++ {
		fmt.Println("kafka:", <-broker)
	}
}
