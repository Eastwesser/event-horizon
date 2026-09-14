package main

import (
	"fmt"
)

// Filter (фильтр для каналов) — пропускает только те элементы, которые проходят условие
// ch -> filter -> ch (на выходе только отфильтрованные элементы)
// Предикат (predicate) — функция, которая возвращает true/false для каждого элемента

func Filter[T any](inputCh <-chan T, predicate func(T) bool) <-chan T {
	outputCh := make(chan T)

	go func() {
		defer close(outputCh)
		for value := range inputCh {
			// Пропускаем только те элементы, для которых предикат возвращает true
			if predicate(value) {
				outputCh <- value
			}
			// Если predicate(value) == false, элемент просто пропускается (не записывается)
		}
	}()

	return outputCh
}

func main() {
	// 1. создаю канал
	channel := make(chan int)

	go func() {
		defer close(channel) // закрываю канал
		// в отдельной горутине пишу множество значений
		for i := 0; i < 10; i++ {
			channel <- i
		}
	}()

	// Предикат: функция, которая проверяет условие (возвращает true/false)
	isOdd := func(value int) bool {
		return value%2 != 0 // проверяем, что число нечётное
	}

	for value := range Filter(channel, isOdd) {
		fmt.Println(value) // фильтруем только те, которые являются НЕЧЁТНЫМИ и выводим их
	}
}
