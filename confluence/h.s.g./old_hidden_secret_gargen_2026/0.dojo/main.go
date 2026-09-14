package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Имитация сетевого запроса. Эту функцию изменять нельзя
func NetworkRequest() int {
	time.Sleep(time.Millisecond * 1)
	return rand.Intn(100) // Возвращает от 0 до 99
}

func main() {
	// Задача:
	// 1. Запустить 1000 NetworkRequest параллельно.
	// 2. Собрать результаты всех вызовов.
	// 3. Вывести ОБЩУЮ сумму чисел.

	var totalSum int

	var wg sync.WaitGroup

	ch1 := make(chan int)

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			ch1 <- NetworkRequest()
		}()
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	for i := range ch1 {
		totalSum += i
		fmt.Print(totalSum)
	}

}
