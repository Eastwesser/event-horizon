package main

import (
	"fmt"
	"sync"
)

func worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("Worker %d обработал: %s\n", id, job)
	}

	fmt.Printf("Worker %d завершил работу\n", id)
}

func dataProducer(jobs chan<- string) {
	defer close(jobs)

	for i := 1; i <= 20; i++ {
		data := fmt.Sprintf("данные-%d", i)
		jobs <- data
		fmt.Printf("Отправлено: %s\n", data)
	}

	fmt.Println("Производитель данных завершил работу")
}

// runWorkerPool запускает пул воркеров с указанным количеством
func runWorkerPool(numWorkers int) {
	if numWorkers <= 0 {
		fmt.Println("Ошибка: количество воркеров должно быть положительным числом")
		return
	}

	fmt.Printf("Запуск с %d воркерами\n", numWorkers)

	jobs := make(chan string, 5) // допустим буфер в 5 элементов
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	// производитель данных в главной горутине
	dataProducer(jobs)

	wg.Wait()
	fmt.Println("Все воркеры завершили работу. Программа завершена.")
}

func main() {
	fmt.Println("=== Тест с 1 воркером ===")
	runWorkerPool(1)

	fmt.Println("\n=== Тест с 3 воркерами ===")
	runWorkerPool(3)

	fmt.Println("\n=== Тест с 5 воркерами ===")
	runWorkerPool(5)
}
