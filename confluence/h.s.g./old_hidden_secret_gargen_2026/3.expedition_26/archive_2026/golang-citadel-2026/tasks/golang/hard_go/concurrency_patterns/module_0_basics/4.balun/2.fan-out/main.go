package main

// FAN OUT (распределение) нужен, когда нужно распределить входные данные из одного канала в несколько сервисов 
// сервис пишет кластер базы данных, состоящий из нескольких шардов, и нам нужно выбирать, в какой из шардов писать
// каждый канал будет отвечать за конкретный узел базы данных
// Когда хотим разделить поток данных по множеству каналов

import (
	"fmt"
	"sync"
)

// приходит входящий канал, мы создаем результирующий
func SplitChannel[T any](inputCh <-chan T, n int) []<-chan T {
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
			idx = (idx + 1) % n // Переходим к следующему каналу, "%" создает круг.
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

func main() {
	channel := make(chan int) // был один канал, а стало два

	go func() {
		defer close(channel)
		for i := 0; i < 10; i++ {
			channel <- i
		}
	}()

	channels := SplitChannel(channel, 2)
	var wg sync.WaitGroup
	wg.Add(2)

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
