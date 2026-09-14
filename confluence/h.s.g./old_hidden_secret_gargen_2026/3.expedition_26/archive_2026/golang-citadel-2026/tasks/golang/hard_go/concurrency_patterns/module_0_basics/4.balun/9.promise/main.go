package main

import (
	"fmt"
	"time"
)

// Если успешно, то выполнится. если не успешно, вернется ошибка

// Promise (промис) — паттерн для работы с асинхронными операциями
// Представляет значение, которое будет доступно в будущем (или ошибку)
//
// Аналогия: Ты заказываешь пиццу (Promise). Пока она готовится, можешь делать другие дела.
// Когда пицца готова, ты получаешь её (success) или узнаёшь, что заказ отменён (error).

// result — структура для хранения результата или ошибки
type result[T any] struct {
	val T
	err error
}

// Promise — обёртка над асинхронной операцией
// Хранит канал, в который будет записан результат выполнения
type Promise[T any] struct {
	resultCh chan result[T]
}

// в промис приходит асинхронная функция (мы её запускаем в горутине)
func NewPromise[T any](asyncFn func() (T, error)) *Promise[T] {
	promise := &Promise[T]{
		resultCh: make(chan result[T]),
	}

	go func() {
		defer close(promise.resultCh) // закрываем канал резалтов

		val, err := asyncFn()                             // когда асинхронная функция выполнится
		promise.resultCh <- result[T]{val: val, err: err} // она запишет в канал некоторые значения
		// can be in a single goroutine (но тогда пришлось бы передавать саксес и эррор функции)
	}()

	return promise
}

// Then регистрирует обработчики успеха и ошибки
// successFn — вызывается при успешном выполнении
// errorFn — вызывается при ошибке
func (p *Promise[T]) Then(successFn func(T), errorFn func(error)) {
	go func() {
		result := <-p.resultCh // Ждём результата (блокируемся, если результат ещё не готов)
		if result.err == nil {
			successFn(result.val) // выполнится, если будет успех
		} else {
			errorFn(result.err) // выполнится, если будет ошибка
		}
	}()
}

func main() {
	// Пример 1: Успешное выполнение
	asyncJob := func() (string, error) {
		time.Sleep(time.Second)
		return "ok", nil
	}

	promise := NewPromise(asyncJob)
	promise.Then(
		func(value string) {
			fmt.Println("success", value)
		},
		func(err error) {
			fmt.Println("error", err.Error())
		},
	)

	// Ждём завершения promise (asyncJob выполняется 1 секунду)
	time.Sleep(2 * time.Second)
}
