package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func worker(id int, jobs <-chan string, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d: канал закрыт, завершаю работу\n", id)
				return
			}
			fmt.Printf("Worker %d обработал: %s\n", id, job)
			time.Sleep(100 * time.Millisecond)
		case <-ctx.Done():
			fmt.Printf("Worker %d: получен сигнал завершения\n", id)
			return
		}
	}
}

func dataProducer(jobs chan<- string, ctx context.Context) {
	defer close(jobs)

	for i := 1; i <= 100; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Производитель: получен сигнал завершения, останавливаюсь")
			return
		default:
			data := fmt.Sprintf("данные-%d", i)

			select {
			// select ожидает либо завершение контекста, либо отправку данных в канал
			case jobs <- data:
				fmt.Printf("Отправлено: %s\n", data)
				time.Sleep(200 * time.Millisecond)
			case <-ctx.Done():
				fmt.Println("Производитель: получен сигнал завершения во время отправки")
				return
			}
		}
	}

	fmt.Println("Производитель: все данные отправлены")
}

func runWorkerPool(numWorkers int) {
	if numWorkers <= 0 {
		fmt.Println("Ошибка: количество воркеров должно быть положительным числом")
		return
	}

	fmt.Printf("Запуск с %d воркерами\n", numWorkers)
	fmt.Println("Нажмите Ctrl+C для корректного завершения")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan // получаем сигнал завершения, который приходит из ОС(он не завершит программу, а только завершит горутины)
		fmt.Printf("\nПолучен сигнал %v, начинаю graceful shutdown...\n", sig)
		cancel()
	}()

	jobs := make(chan string, 5)
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, ctx, &wg)
	}

	dataProducer(jobs, ctx)
	wg.Wait()

	fmt.Println("Все воркеры завершили работу. Программа завершена.")
}

func main() {
	runWorkerPool(3)
}
