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

// приходит входящий канал, мы создаем результирующий
func PipelineSplitChannel[T any](inputCh <-chan T, n int) []<-chan T {
	outputChannels := make([]chan T, n)
	for i := 0; i < n; i++ {
		outputChannels[i] = make(chan T)
	}

	// в фоне начинаем процессить
	go func() {
		// идем Round Robin и пишем в каждый из каналов по индексу, так работает распределение
		// Round Robin: вопрос: почему Round Robin, а не случайное распределение или по нагрузке? Round Robin — простой и равномерный.
		idx := 0
		// Round Robin — это последовательное распределение по кругу, независимо от того, пуст канал или нет.
		/*
			Round Robin — это как очередь в банке:
				Первый клиент → окно 1
				Второй клиент → окно 2
				Третий клиент → окно 1 (снова)
				Четвёртый клиент → окно 2 (снова)
				Не проверяем, свободно ли окно — просто по очереди.
		*/
		for value := range inputCh {
			outputChannels[idx] <- value // запись в канал по индексу блокирующим способом (неблокируюзая через селект с дефолтом). Если один из выходных каналов не читается - вся горутина заблокируется.
			idx = (idx + 1) % n          // Переходим к следующему каналу, "%" создает круг.
		}

		for _, ch := range outputChannels {
			close(ch)
		}
	}()

	// typecast to []<-chan T (one-way channel), либо возвращаем на 14 строке просто []chan T, и не пишем все, что ниже (34-37)
	resultChannels := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		resultChannels[i] = outputChannels[i]
	}
	// Почему нельзя вернуть []chan T напрямую? Однонаправленные каналы <-chan T защищают от записи снаружи.

	return resultChannels
}

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
	var wg sync.WaitGroup
	wg.Add(n)

	outputCh := make(chan string)
	splitChs := PipelineSplitChannel(inputCh, n) // разбиваем поток из каналов на несколько каналов

	// n - фактор параллелизма. Запускаем n горутин для параллельной обработки
	// Это комбинация Pipeline + Fan-Out: одна стадия масштабируется через несколько воркеров
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			for data := range splitChs[idx] { // splitChs[idx] - каждой горутине дал свой канал, чтобы очередь не стопорилась
				outputCh <- fmt.Sprintf("data sent: %s by worker %d", data, idx) // Используем idx, а не i
			}
		}(i)
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
