// Redis PubSub метафора: каждое событие доставляется ВСЕМ нодам;
// нода шлёт клиенту только если user локальный.
package main

import (
	"fmt"
	"sync"
)

type Event struct {
	UserID string
	Text   string
}

type Node struct {
	name  string
	local map[string]struct{}
}

func (n *Node) OnEvent(e Event) {
	if _, ok := n.local[e.UserID]; ok {
		fmt.Printf("%s delivers to %s: %s\n", n.name, e.UserID, e.Text)
	}
}

func main() {
	nodes := []*Node{
		{name: "ws-1", local: map[string]struct{}{"u1": {}}},
		{name: "ws-2", local: map[string]struct{}{"u2": {}, "u3": {}}},
	}

	events := []Event{
		{"u2", "hi from chat-service"},
		{"u1", "welcome"},
	}

	var wg sync.WaitGroup
	for _, e := range events {
		for _, n := range nodes {
			wg.Add(1)
			n, e := n, e
			go func() {
				defer wg.Done()
				n.OnEvent(e) // broadcast: все ноды видят событие
			}()
		}
	}
	wg.Wait()
}
