package main

import (
	"fmt"
	"time"
)

type Future[T any] struct {
	resultCh <-chan T
}

func NewFuture[T any](resultCh <-chan T) Future[T] {
	return Future[T]{
		resultCh: resultCh,
	}
}

func (f *Future[T]) Get() T {
	return <-f.resultCh // метод Get() - блокирующий, только мы инкапсулировали канал, чтобы не зависеть от типа данных
}

type Promise[T any] struct {
	resultCh chan T
}

func NewPromise[T any]() Promise[T] {
	return Promise[T]{
		resultCh: make(chan T),
	}
}

func (p *Promise[T]) Set(value T) {
	p.resultCh <- value
	close(p.resultCh)
}

func (p *Promise[T]) GetFuture() Future[T] {
	return NewFuture(p.resultCh)
}

func main() {
	promise := NewPromise[string]()

	go func() {
		time.Sleep(time.Second)
		promise.Set("agreement") // как расписка, что мы вернем человеку деньги в будущем, если берем в долг
	}()

	future := promise.GetFuture() // GetFuture() - это метод, который возвращает Future[T]
	// хочу дождаться, когда она выполнится
	result := future.Get() // Get() - это блокирующий вызов, в будущем будет получен результат, а будущее доступно через какое-то время
	fmt.Println(result)
}
