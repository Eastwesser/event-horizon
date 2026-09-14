package main

import (
	"fmt"
	"sync"
	"time"
)

// monitor — периодический воркер с graceful Stop (done + WaitGroup).
// Важно: break в select выходит только из select, нужен return / labeled break.
type monitor struct {
	wg     sync.WaitGroup
	ticker *time.Ticker
	done   chan struct{}
	n      int
}

func (m *monitor) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			m.n++
			fmt.Println("tick", m.n)
			select {
			case <-m.ticker.C:
			case <-m.done:
				return
			}
		}
	}()
}

func (m *monitor) Stop() {
	close(m.done)
	m.wg.Wait()
	m.ticker.Stop()
}

func main() {
	m := &monitor{
		ticker: time.NewTicker(30 * time.Millisecond),
		done:   make(chan struct{}),
	}
	m.Start()
	time.Sleep(100 * time.Millisecond)
	m.Stop()
	fmt.Println("stopped after", m.n, "ticks")
}
