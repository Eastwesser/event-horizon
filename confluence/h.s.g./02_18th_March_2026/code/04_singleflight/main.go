// Singleflight: одинаковые in-flight запросы → один внешний вызов, остальные ждут.
// Паттерн из задачи Dadata (платный API).
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	val string
	err error
	done chan struct{}
}

type Group struct {
	mu sync.Mutex
	m  map[string]*result
}

func (g *Group) Do(key string, fn func() (string, error)) (string, error) {
	g.mu.Lock()
	if r, ok := g.m[key]; ok {
		g.mu.Unlock()
		<-r.done
		return r.val, r.err
	}
	r := &result{done: make(chan struct{})}
	g.m[key] = r
	g.mu.Unlock()

	r.val, r.err = fn()
	close(r.done)

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return r.val, r.err
}

func main() {
	g := &Group{m: make(map[string]*result)}
	var calls atomic.Int32
	fn := func() (string, error) {
		calls.Add(1)
		time.Sleep(100 * time.Millisecond)
		return "addr", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, _ := g.Do("fias-1", fn)
			fmt.Println(v)
		}()
	}
	wg.Wait()
	fmt.Println("external calls:", calls.Load()) // 1
}
