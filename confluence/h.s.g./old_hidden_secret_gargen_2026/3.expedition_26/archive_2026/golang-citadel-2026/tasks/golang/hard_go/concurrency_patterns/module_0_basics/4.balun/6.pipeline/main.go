package main

// Pipeline (конвейер) — цепочка независимых стадий обработки данных
// Каждая стадия работает в своей горутине и передаёт данные следующей
// Пример: логи → проверка секретов → маскировка → аналитика → БД
// Преимущество: параллельная обработка (пока одна стадия обрабатывает элемент,
// следующая может обрабатывать предыдущий)

// Логи идут, нужно проверить секреты, замаскировать их, посчитать аналитику и отгрузить в базу данных
// Когда у меня есть некоторые промежуточные слои, буфера (очереди, через которые проходят данные),
// у нас есть независимые части этого пайплайна

import (
	"fmt"
)

// generate — генерирует значения из слайса в канал (источник данных)
// Идея в том, что мы слайс преобразуем в канал, 
func generate[T any](values ...T) <-chan T {
	outputCh := make(chan T)

	go func() {
		defer close(outputCh)
		for _, value := range values {
			outputCh <- value
		}
	}()

	return outputCh
}

// process — обрабатывает значения из канала (стадия преобразования)
// Это по сути Decorator, но в контексте Pipeline
func process[T any](inputCh <-chan T, action func(T) T) <-chan T {
	outputCh := make(chan T)

	go func() {
		defer close(outputCh)
		for value := range inputCh {
			outputCh <- action(value)
		}
	}()

	return outputCh
}

func main() {
	// Исходные данные
	values := []int{1, 2, 3, 4, 5}

	// Функция преобразования (возводит в квадрат)
	mul := func(value int) int {
		return value * value
	}

	// Pipeline: generate → process → вывод
	// Декларативный стиль: описываем ЧТО делаем, а не КАК
	// generate(values...) создаёт канал с данными
	// process(...) обрабатывает каждое значение
	// range читает результаты
	for value := range process(generate(values...), mul) {
		fmt.Println(value) // 1, 4, 9, 16, 25
	}
}
