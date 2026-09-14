package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 1. Выход по условию - естественное завершение
func exitByCondition() {
	fmt.Println("\n=== 1. Выход по условию ===")

	go func() {
		for i := 1; i <= 3; i++ {
			fmt.Printf("Горутина 1: шаг %d\n", i)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("Горутина 1: завершилась по условию")
	}()

	time.Sleep(500 * time.Millisecond)
}

// 2. Остановка через канал уведомления
func exitByChannel() {
	fmt.Println("\n=== 2. Остановка через канал уведомления ===")

	stopChan := make(chan bool)

	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("Горутина 2: получила сигнал остановки")
				return
			default:
				fmt.Println("Горутина 2: работает...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	time.Sleep(300 * time.Millisecond)
	stopChan <- true
	time.Sleep(100 * time.Millisecond)
}

// 3. Остановка через контекст
func exitByContext() {
	fmt.Println("\n=== 3. Остановка через контекст ===")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Горутина 3: контекст отменен")
				return
			default:
				fmt.Println("Горутина 3: работает...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
}

// 4. Остановка через runtime.Goexit()
func exitByGoexit() {
	fmt.Println("\n=== 4. Остановка через runtime.Goexit() ===")

	go func() {
		defer func() {
			fmt.Println("Горутина 4: defer выполнен")
		}()

		fmt.Println("Горутина 4: начинаю работу")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Горутина 4: вызываю Goexit()")
		runtime.Goexit()
		fmt.Println("Горутина 4: этот код не выполнится")
	}()

	time.Sleep(500 * time.Millisecond)
}

// 5. Остановка через panic (с восстановлением)
func exitByPanic() {
	fmt.Println("\n=== 5. Остановка через panic (с восстановлением) ===")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Горутина 5: восстановилась после panic: %v\n", r)
			}
		}()

		fmt.Println("Горутина 5: начинаю работу")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Горутина 5: вызываю panic")
		panic("искусственная ошибка")
	}()

	time.Sleep(500 * time.Millisecond)
}

// 6. Остановка через WaitGroup
func exitByWaitGroup() {
	fmt.Println("\n=== 6. Остановка через WaitGroup ===")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fmt.Println("Горутина 6: начинаю работу")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Горутина 6: завершаюсь")
	}()

	fmt.Println("Ожидаю завершения горутины 6...")
	wg.Wait()
	fmt.Println("Горутина 6: завершена")
}

// 7. Остановка через закрытие канала
func exitByChannelClose() {
	fmt.Println("\n=== 7. Остановка через закрытие канала ===")

	dataChan := make(chan int)

	go func() {
		for value := range dataChan {
			fmt.Printf("Горутина 7: получила %d\n", value)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("Горутина 7: канал закрыт, завершаюсь")
	}()

	dataChan <- 1
	dataChan <- 2
	close(dataChan)
	time.Sleep(200 * time.Millisecond)
}

// 8. Остановка через select с несколькими каналами
func exitByMultipleChannels() {
	fmt.Println("\n=== 8. Остановка через select с несколькими каналами ===")

	stopChan := make(chan bool)
	timeoutChan := make(chan bool)

	go func() {
		for {
			select {
			case <-stopChan:
				fmt.Println("Горутина 8: получила сигнал остановки")
				return
			case <-timeoutChan:
				fmt.Println("Горутина 8: таймаут истек")
				return
			default:
				fmt.Println("Горутина 8: работает...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Запускаем таймер в отдельной горутине
	go func() {
		time.Sleep(400 * time.Millisecond)
		timeoutChan <- true
	}()

	time.Sleep(200 * time.Millisecond)
	stopChan <- true
	time.Sleep(100 * time.Millisecond)
}

func main() {
	fmt.Println("Демонстрация всех способов остановки горутин")

	exitByCondition()
	exitByChannel()
	exitByContext()
	exitByGoexit()
	exitByPanic()
	exitByWaitGroup()
	exitByChannelClose()
	exitByMultipleChannels()

	fmt.Println("\n=== Все способы продемонстрированы ===")
}
