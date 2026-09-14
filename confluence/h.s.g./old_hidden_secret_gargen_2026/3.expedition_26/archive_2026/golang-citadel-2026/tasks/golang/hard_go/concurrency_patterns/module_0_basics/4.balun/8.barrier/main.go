package main

import (
	"log"
	"sync"
)

// Barrier (барьер синхронизации) — примитив синхронизации, который заставляет
// N горутин ждать друг друга, пока все не достигнут точки синхронизации
//
// НЕ ПУТАТЬ С БАРЬЕРАМИ ПАМЯТИ (memory barriers)!!!
// Барьер синхронизации — это точка, где все горутины должны встретиться
//
// Аналогия: Группа туристов идёт по тропе. Все должны дойти до контрольной точки,
// прежде чем продолжить путь дальше. Самый медленный определяет скорость группы.

type Barrier struct {
	mutex sync.Mutex
	count int // Количество горутин, которые достигли барьера
	size  int // Общее количество горутин, которые должны встретиться

	beforeCh chan struct{}
	afterCh  chan struct{}
}

func NewBarrier(size int) *Barrier {
	return &Barrier{
		size:     size,
		beforeCh: make(chan struct{}, size),
		afterCh:  make(chan struct{}, size),
	}
}

// Before ждёт, пока все N горутин не достигнут барьера
// Когда последняя горутина приходит, все продолжают одновременно
func (b *Barrier) Before () {
	b.mutex.Lock()

	b.count++              // первая горутина увеличит каунт на 1
	if b.count == b.size { // но сайз равен 3
		for i := 0; i < b.size; i++ {
			b.beforeCh <- struct{}{} // последняя горутина приходит и отпускает
		}
	}

	b.mutex.Unlock()
	<-b.beforeCh // если сайз не равен 3, значит мы блокируемся
}

func (b *Barrier) After() {
	b.mutex.Lock()

	b.count--
	if b.count == 0 {
		for i := 0; i < b.size; i++ {
			b.afterCh <- struct{}{}
		}
	}

	b.mutex.Unlock()
	<-b.afterCh
}

func main() {
	var wg sync.WaitGroup
	wg.Add(3)

	bootstrap := func() {
		log.Println("bootstrap")
	}

	work := func() {
		log.Println("work")
	}

	count := 3 // number of workers, 3 coordinated operations
	barrier := NewBarrier(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < count; j++ {
				// wait for all the workers to finish previous loop
				barrier.Before() // barrier before GC
				bootstrap()      // GC (garbage collection with StW)
				// wait for the other workers to bootstrap
				barrier.After() // restart all the threads
				work()
			}
		}()
	}

	wg.Wait()
}
