package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

/*
Round Robin — основной LB на middle-собесе.

Next() потокобезопасен: Mutex (или atomic для индекса).
Пустой список серверов — отдельный edge case (здесь паникуем явно / можно вернуть error).
*/
type RoundRobin struct {
	servers []string
	idx     uint64
}

func NewRoundRobin(servers ...string) *RoundRobin {
	if len(servers) == 0 {
		panic("round robin: no servers")
	}
	return &RoundRobin{servers: append([]string(nil), servers...)}
}

func (rr *RoundRobin) Next() string {
	i := atomic.AddUint64(&rr.idx, 1) - 1
	return rr.servers[i%uint64(len(rr.servers))]
}

// RoundRobinMu — классика «с мьютексом», если интервьюер просит без atomics.
type RoundRobinMu struct {
	servers []string
	cur     int
	mu      sync.Mutex
}

func (rr *RoundRobinMu) Next() string {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	s := rr.servers[rr.cur]
	rr.cur = (rr.cur + 1) % len(rr.servers)
	return s
}

func main() {
	rr := NewRoundRobin("a:8080", "b:8080", "c:8080")
	fmt.Println("--- atomic RR ---")
	for i := 0; i < 7; i++ {
		fmt.Println(rr.Next())
	}

	mu := &RoundRobinMu{servers: []string{"x", "y"}}
	fmt.Println("--- mutex RR ---")
	for i := 0; i < 4; i++ {
		fmt.Println(mu.Next())
	}
}
