package main

// Pipeline (конвейер) — цепочка независимых стадий обработки данных
// Каждая стадия работает в своей горутине и передаёт данные следующей
// Пример: логи → проверка секретов → маскировка → аналитика → БД
// Преимущество: параллельная обработка (пока одна стадия обрабатывает элемент,
// следующая может обрабатывать предыдущий)

import (
	"fmt"
	"sync"
)

// Приходит входной канал,  все как обычно, в 21 строке имитируем парсинг
func parse(inputCh <-chan string) <-chan string {
	outputCh := make(chan string)

	go func() {
		defer close(outputCh)
		for datae := range inputCh {
			outputCh <- fmt.Sprintf("parsed - %s", datae)
		}
	}()

	return outputCh
}

// для функции send входным каналом является канал, который возвращает функция parse
func send(inputCh <-chan string, n int) <-chan string {
	outputCh := make(chan string)
	var wg sync.WaitGroup
	wg.Add(n)

	// n - фактор параллелизма. Запускаем n горутин для параллельной обработки
	// Это комбинация Pipeline + Fan-Out: одна стадия масштабируется через несколько воркеров
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for data := range inputCh {
				outputCh <- fmt.Sprintf("data sent: %s by worker %d", data, i) // n горутин будут заниматься отправкой
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outputCh)
	}()

	return outputCh
}

func main() {
	channel := make(chan string)

	go func() {
		defer close(channel)

		for i := 0; i < 5; i++ {
			channel <- "value"
		}
	}()

	// тут у меня парсилка, и я тут передаю некоторый параметр (2 - некоторый параллель фактор)
	for value := range send(parse(channel), 2) {
		fmt.Println(value)
	}
}
