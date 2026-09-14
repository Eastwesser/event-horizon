package main

import (
	"fmt"
	"sync"
)

// TEE разветвитель, в котором данные нужно передать в оба канала
// если пишем в кластер баз данных, кластер состоит из нескольких реплик, и нам нужно написать в два, три, четыре места.

// поток разделяем для реплик, а вдруг одна реплика будет перегружена, вдруг очередь будет копиться
// вместо того, чтобы копить их в одном месте, мы берем и сегментируем поток в каждую реплику

func Tee[T any](inputCh <-chan T, n int) []<-chan T {
	outputChs := make([]chan T, n)
	for i := 0; i < n; i++ {
		outputChs[i] = make(chan T)
	}

	go func() {
		// Берем цикл и прогоняем по множеству каналов, записывая в каждый
		for value := range inputCh {
			for i := 0; i < n; i++ {
				// outputChs[i] <- value // can be non-blocking with select - default
				/*
					Если один канал не читается (полон или нет читателя), 
					вся горутина блокируется, и остальные каналы не получат данные.
				*/

				// Неблокирующая запись: если канал полон, пропускаем для этого канала
				// Это защищает от блокировки всей горутины, если один канал не читается
				select {
				case outputChs[i] <- value:
					// Записали успешно
				default:
					// Канал полон или не читается — пропускаем для этого канала
					// Альтернатива: можно логировать или использовать буферизованные каналы
				}
			}
		}
		for _, channel := range outputChs {
			close(channel)
		}
	}()

	// typecast to []<-chan T (one-way channel), либо возвращаем на 14 строке просто []chan T, и не пишем все, что ниже (34-37)
	resultChs := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		resultChs[i] = outputChs[i]
	}
	// Почему нельзя вернуть []chan T напрямую? Однонаправленные каналы <-chan T защищают от записи снаружи.

	return resultChs
}

func main() {
	channel := make(chan int)

	go func() {
		defer close(channel)
		for i := 0; i < 5; i++ {
			channel <- i
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	channels := Tee(channel, 2)

	go func() {
		defer wg.Done()
		for value := range channels[0] {
			fmt.Println("ch1: ", value)
		}
	}()

	go func() {
		defer wg.Done()
		for value := range channels[1] {
			fmt.Println("ch2: ", value)
		}
	}()

	wg.Wait()
}
