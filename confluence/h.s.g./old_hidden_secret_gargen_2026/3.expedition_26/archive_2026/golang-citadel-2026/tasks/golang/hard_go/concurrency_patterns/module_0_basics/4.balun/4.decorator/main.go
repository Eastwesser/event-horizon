package main

import (
	"fmt"
)

// Decorator (декоратор для каналов) — оборачивает канал и преобразует данные на лету
// Проблема: API принимает канал, но мы не можем изменить данные внутри канала напрямую
// Решение: декоратор создаёт новый канал и преобразует данные при чтении
// Это паттерн Decorator для потоков данных

func Decorate[T any](inputCh <-chan T, action func(T) T) <-chan T {
	outputCh := make(chan T)

	// асинхронная джоба, которая в фоне начинает процессить
	go func() {
		defer close(outputCh)
		for number := range inputCh {
			outputCh <- action(number)
		}
	}()

	return outputCh
}

// Adapt — более универсальный и принимает в себя любые типы
// Преобразует канал типа T в канал типа U через функцию преобразования
func Adapt[T, U any](inputCh <-chan T, action func(T) U) <-chan U {
	outputCh := make(chan U)

	go func() {
		defer close(outputCh)
		for number := range inputCh {
			outputCh <- action(number)
		}
	}()

	return outputCh
}


func main() {
	// создаем отдельный канал
	channel := make(chan int)

	go func() {
		defer close(channel) // Важно: закрываем канал после записи
		// в отдельной горутине пишу множества значений
		for i := 0; i < 5; i++ {
			channel <- i
		}
	}()

	mul := func(value int) int {
		return value * value
	}

	// идем рейнджом по функции Декорэйт(канал, преобразующая функция)
	// Каждому значению из канала выполнять эту фунцию мул в виде преобразователя
	for number := range Decorate(channel, mul) {
		fmt.Println(number)
	}
}
