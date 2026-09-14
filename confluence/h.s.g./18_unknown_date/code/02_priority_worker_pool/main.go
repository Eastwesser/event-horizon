package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type taskItem struct {
	fn       func() error
	priority int
	seq      int // FIFO внутри одного приоритета
	index    int
}

type taskHeap []*taskItem

func (h taskHeap) Len() int { return len(h) }
func (h taskHeap) Less(i, j int) bool {
	if h[i].priority != h[j].priority {
		return h[i].priority > h[j].priority // 5 выше 1
	}
	return h[i].seq < h[j].seq
}
func (h taskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *taskHeap) Push(x any) {
	item := x.(*taskItem)
	item.index = len(*h)
	*h = append(*h, item)
}
func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

// WorkerPool — приоритетная очередь + N воркеров + graceful Stop.
type WorkerPool struct {
	mu      sync.Mutex
	cond    *sync.Cond
	h       taskHeap
	seq     int
	stopped bool
	wg      sync.WaitGroup
	pending int
}

func NewWorkerPool(n int) *WorkerPool {
	p := &WorkerPool{}
	p.cond = sync.NewCond(&p.mu)
	for i := 0; i < n; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

func (p *WorkerPool) Add(task func() error, priority int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.seq++
	heap.Push(&p.h, &taskItem{fn: task, priority: priority, seq: p.seq})
	p.pending++
	p.cond.Signal()
}

func (p *WorkerPool) PendingCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pending
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		p.mu.Lock()
		for p.h.Len() == 0 && !p.stopped {
			p.cond.Wait()
		}
		if p.h.Len() == 0 && p.stopped {
			p.mu.Unlock()
			return
		}
		item := heap.Pop(&p.h).(*taskItem)
		p.pending--
		p.mu.Unlock()

		if err := item.fn(); err != nil {
			fmt.Println("task error:", err)
		}
	}
}

func (p *WorkerPool) Stop() {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}
	p.stopped = true
	p.cond.Broadcast()
	p.mu.Unlock()
	p.wg.Wait()
}

func main() {
	pool := NewWorkerPool(3)

	pool.Add(func() error {
		time.Sleep(80 * time.Millisecond)
		fmt.Println("Задача 1 (приор 1)")
		return nil
	}, 1)
	pool.Add(func() error { fmt.Println("Задача 2 (приор 5)"); return nil }, 5)
	pool.Add(func() error { fmt.Println("Задача 3 (приор 5)"); return nil }, 5)
	pool.Add(func() error { fmt.Println("Задача 4 (приор 3)"); return nil }, 3)
	time.Sleep(20 * time.Millisecond)
	pool.Add(func() error { fmt.Println("Задача 5 (приор 5)"); return nil }, 5)

	pool.Stop()
	fmt.Println("pending after stop:", pool.PendingCount())
}
