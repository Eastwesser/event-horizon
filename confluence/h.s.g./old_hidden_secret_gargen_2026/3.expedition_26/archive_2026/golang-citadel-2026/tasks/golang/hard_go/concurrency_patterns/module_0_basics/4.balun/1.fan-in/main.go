package main

// FAN IN (сбор) нужен, когда нужно смержить входные данные из сервисов, смержить метрики

import (
	"fmt"
	"sync"
)

// получаем слайс каналов, вариативный channels
func MergeChannels[T any](channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	outputCh := make(chan T)
	wg.Add(len(channels)) //  проще, если знаешь количество заранее.
	// для каждого канала из множества создается горутина
	
	for _, channel := range channels {
		// Джоба в отдельной горутине
		go func(ch <- chan T) {
			// "wg.Add(1)" В каждой итерации добавляем 1 горутину
			// "wg.Add(1)" в цикле — удобнее, если количество может меняться.
			defer wg.Done()
			for value := range ch {
				outputCh <- value // вычитывает и записывает данные в результирующий канал
			}
		}(channel) // здесь можно и не указывать channel (замыкания), потому что версия Го 1.25 (выше 1.22, поэтому нет необходимости писать channel)

	}

	go func() {
		wg.Wait() // после того, как мы убедимся, что все пишущие горутины с 17 строки завершатся, можем закрывать канал.
		close(outputCh)
	}()

	return outputCh // мы не блочим клиента, и сразу ему отдаем, как в фоне отрабортают горутины, и потом отдельной горутиной закроется канал.
}

func main() {
	channel1 := make(chan int)
	channel2 := make(chan int)
	channel3 := make(chan int)

	go func() {
		defer func() {
			close(channel1)
			close(channel2)
			close(channel3)
		}()

		// пишем 100 значений, потом через defer из 44 строчки закрываем эти каналы
		for i := 0; i < 100; i += 3 {
			channel1 <- i
			channel2 <- i + 1
			channel3 <- i + 2
		}
	}()

	// сливаем данные из трёх каналов в один канал, проходимся рэйнджом, как будто бы юзаем один канал
	for value := range MergeChannels(channel1, channel2, channel3) {
		// range будет идти, пока канал не закроется. Мы не знаем, какая горутина запишет последней.
		fmt.Println(value)
	}

}
