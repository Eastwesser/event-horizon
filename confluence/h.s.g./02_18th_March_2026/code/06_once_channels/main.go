// once на каналах (без sync.Once): buffered token, кто забрал — выполняет f.
package main

import (
	"fmt"
	"sync"
)

const goroutinesNumber = 10

type once struct {
	ch chan struct{}
}

func newOnce() *once {
	o := &once{ch: make(chan struct{}, 1)}
	o.ch <- struct{}{} // один токен
	return o
}

func (o *once) do(f func()) {
	select {
	case <-o.ch:
		f()
	default:
	}
}

func main() {
	var wg sync.WaitGroup
	o := newOnce()
	wg.Add(goroutinesNumber)
	for i := 0; i < goroutinesNumber; i++ {
		go func() {
			defer wg.Done()
			o.do(func() { fmt.Println("call") })
		}()
	}
	wg.Wait()
}
