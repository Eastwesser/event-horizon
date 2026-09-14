package main

import "fmt"

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Первый канал для исходных чисел
	ch1 := make(chan int)
	// Второй канал для результатов
	ch2 := make(chan int)

	// Горутина 1: генерация чисел
	go func() {
		defer close(ch1)
		for _, x := range numbers {
			ch1 <- x
		}
	}()

	// Горутина 2: обработка (умножение на 2)
	go func() {
		defer close(ch2)
		for x := range ch1 {
			ch2 <- x * 2
		}
	}()

	// Чтение из второго канала и вывод в stdout
	for result := range ch2 {
		fmt.Println(result)
	}
}
