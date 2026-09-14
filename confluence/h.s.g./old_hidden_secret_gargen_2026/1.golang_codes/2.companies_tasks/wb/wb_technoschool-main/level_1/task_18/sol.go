package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Counter struct{ v int64 }

func (c *Counter) Inc()         { atomic.AddInt64(&c.v, 1) }
func (c *Counter) Value() int64 { return atomic.LoadInt64(&c.v) }

func main() {
	var c Counter
	var wg sync.WaitGroup
	workers, per := 10, 1000
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < per; j++ {
				c.Inc()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println(c.Value()) // ожидаемо 10000
}
