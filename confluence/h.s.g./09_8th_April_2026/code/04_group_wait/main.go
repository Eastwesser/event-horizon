// Group: мини-WaitGroup на буферизованном канале (size известен заранее).
package main

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type Group struct {
	c    chan struct{}
	size int
}

func New(size int) *Group {
	return &Group{c: make(chan struct{}, size), size: size}
}

func (g *Group) Done() { g.c <- struct{}{} }

func (g *Group) Wait() {
	for i := 0; i < g.size; i++ {
		<-g.c
	}
}

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	n := len(numbers)
	var (
		res []int
		mu  sync.Mutex
	)
	g := New(n)
	for _, num := range numbers {
		go func(num int) {
			defer g.Done()
			mu.Lock()
			res = append(res, num)
			mu.Unlock()
		}(num)
	}
	g.Wait()
	sort.Ints(res)
	if !reflect.DeepEqual(res, numbers) {
		panic("wrong code")
	}
	fmt.Println("ok:", res)
}
