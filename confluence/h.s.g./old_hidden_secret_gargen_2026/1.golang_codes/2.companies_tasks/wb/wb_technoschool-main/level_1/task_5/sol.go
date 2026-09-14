package main

import (
	"fmt"
	"time"
)

func main() {
	N := 5 // время работы в секундах
	fmt.Printf("Программа будет работать %d секунд\n", N)

	// Канал для передачи данных
	dataChan := make(chan int)

	// Горутина для отправки данных
	go func() {
		timer := time.NewTimer(time.Duration(N) * time.Second)
		defer timer.Stop()

		for i := 1; ; i++ {
			select {
			case dataChan <- i:
				fmt.Printf("Отправлено: %d\n", i)
				time.Sleep(500 * time.Millisecond)
			case <-timer.C:
				fmt.Println("Время истекло, закрываю канал")
				close(dataChan)
				return
			}
		}
	}()

	// Чтение данных из канала
	for value := range dataChan {
		fmt.Printf("Получено: %d\n", value)
	}

	fmt.Println("Программа завершена")
}
