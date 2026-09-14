package main

import (
	"fmt"
	"time"
)

// Future (фьючер) — паттерн для работы с асинхронными операциями
// Представляет значение, которое будет доступно в будущем
//
// Аналогия: Ты заказываешь пиццу (Future). Пока она готовится, можешь делать другие дела.
// Когда забираешь пиццу (Get()), получаешь результат.

// Future — обёртка над асинхронной операцией
// Операция запускается сразу при создании Future
type Future[T any] struct {
	resultCh chan T
}

func NewFuture[T any](action func() T) Future[T] {
	future := Future[T]{
		resultCh: make(chan T),
	}

	go func() {
		defer close(future.resultCh)
		future.resultCh <- action() // Запускаем операцию сразу
	}()

	return future
}

// Get — блокирующий вызов, ждёт результата
// Инкапсулирует канал, чтобы не зависеть от типа данных
func (f *Future[T]) Get() T {
	return <-f.resultCh // метод Get() - блокирующий, только мы инкапсулировали канал, чтобы не зависеть от типа данных
}

func main() {
	// хочу выполнить асинхронную джобу
	asyncJob := func() interface{} {
		time.Sleep(time.Second)
		return "success"
	}

	future := NewFuture(asyncJob)
	// хочу дождаться, когда она выполнится
	result := future.Get() // Get() - это блокирующий вызов, в будущем будет получен результат, а будущее доступно через какое-то время
	fmt.Println(result)
}
