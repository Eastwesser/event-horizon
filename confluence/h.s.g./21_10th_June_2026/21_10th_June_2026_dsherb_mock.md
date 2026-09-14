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
	
	ch1 := make(chan int, 1000)
	
	for i := 0; i < len(ch1); i++ {
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

===

Мьютексы



	var totalSum int
	
	var wg sync.WaitGroup
	var mu sync.mutex

	ch1 := make(chan int, 1000)
	
	for i := 0; i < len(ch1); i++ {
	    wg.Add(1)
	    
	    go func() {
	        defer wg.Done
		mu.Lock()

	        v := NetworkRequest()
	        ch1 <- v

		mu.Unlock()
	    }()
	    
	    
	}
	
	wg.Wait()

	
    fmt.Print(totalSum)




атомики 

package main

import (
	"fmt"
	"math/rand"
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
	
	
	var totalSum int64
	
	var wg sync.WaitGroup
	
	for i := 0; i < len(ch1); i++ {
	    wg.Add(1)
	    
	    go func() {
	        defer wg.Done
	        v := NetworkRequest()
	        atomic.AddInt64(&totalSum, int64(v))
	    }()
	    
	    
	}
	
	wg.Wait()

	
    fmt.Print(totalSum)
}


И еще есть вариант с семафорами, там два семафора и 1 канал. struct{}{} передаем.