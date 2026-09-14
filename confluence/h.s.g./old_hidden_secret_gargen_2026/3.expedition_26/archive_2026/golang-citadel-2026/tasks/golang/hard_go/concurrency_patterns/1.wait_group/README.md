# Паттерн: WaitGroup — Базовая синхронизация горутин

## 🎯 Задача

Запустить N горутин параллельно и **дождаться их завершения**, прежде чем продолжить.

## 🗡️ Код из задачи

```go
func main() {
	var wg sync.WaitGroup
	ch1 := make(chan int, 10)
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, i := range nums {
		wg.Add(1)

		go func() {
			defer wg.Done()
			ch1 <- i
		}()
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	for v := range ch1 {
		fmt.Println(v)
	}
}
```

## 🔍 Разбор по шагам

### Шаг 1: Инициализация WaitGroup
```go
var wg sync.WaitGroup  // Счетчик горутин = 0
```

### Шаг 2: Запуск горутин
```go
for _, i := range nums {
    wg.Add(1)  // Счетчик++
    
    go func() {
        defer wg.Done()  // Счетчик-- при выходе
        ch1 <- i
    }()
}
```

**Важно**: `wg.Add(1)` вызывается **до** `go func()`, чтобы избежать race condition.

### Шаг 3: Закрытие канала после завершения
```go
go func() {
    wg.Wait()    // Ждем, пока счетчик не станет 0
    close(ch1)   // Закрываем канал
}()
```

Без этого `close(ch1)` main застрял бы на `for v := range ch1` навечно.

### Шаг 4: Чтение из канала
```go
for v := range ch1 {
    fmt.Println(v)  // Печатаем значения по мере поступления
}
```

## ⚠️ BUG в коде! (Классическая ловушка)

**Проблема**: Все горутины получают **одно и то же значение** `i` (последнее).

```go
for _, i := range nums {
    go func() {
        ch1 <- i  // ❌ i - это переменная цикла!
    }()
}
```

**Почему?** Переменная `i` **переиспользуется** на каждой итерации. К моменту запуска горутины `i` уже может быть изменено.

**Вывод**:
```
10
10
10
10
...
```

## ✅ Исправление

### Способ 1: Передать через параметр
```go
for _, i := range nums {
    wg.Add(1)
    go func(val int) {  // val - локальная копия
        defer wg.Done()
        ch1 <- val
    }(i)
}
```

### Способ 2: Создать локальную переменную
```go
for _, i := range nums {
    i := i  // Тень переменной (shadowing)
    wg.Add(1)
    go func() {
        defer wg.Done()
        ch1 <- i
    }()
}
```

### Способ 3 (Go 1.22+): Использовать новую семантику range
```go
// В Go 1.22+ переменная i теперь уникальна для каждой итерации
for _, i := range nums {
    wg.Add(1)
    go func() {
        defer wg.Done()
        ch1 <- i  // Теперь безопасно!
    }()
}
```

## 💡 Правило Додзё

> **"WaitGroup — это счётчик, а не барьер. Каждый `Add(1)` — это обещание завершить. Каждый `Done()` — это выполненное обещание. `Wait()` — это ожидание всех обещаний."**

## 🧪 Эксперименты

### Тест 1: Без WaitGroup
```go
func main() {
    for i := 0; i < 5; i++ {
        go fmt.Println(i)
    }
    // main завершится до печати!
}
```

**Вывод**: Ничего (или 1-2 числа, если повезет).

### Тест 2: С WaitGroup
```go
func main() {
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            fmt.Println(n)
        }(i)
    }
    wg.Wait()
    fmt.Println("Done!")
}
```

**Вывод**: Все 5 чисел (в произвольном порядке), затем "Done!".

### Тест 3: Ошибка — Add после Wait
```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    time.Sleep(1 * time.Second)
}()

wg.Wait()  // Ждем завершения

wg.Add(1)  // ❌ PANIC: sync: WaitGroup is reused before previous Wait has returned
```

**Правило**: После `Wait()` нельзя повторно использовать WaitGroup без re-initialization.

### Тест 4: Счетчик уходит в минус
```go
var wg sync.WaitGroup
wg.Done()  // ❌ PANIC: sync: negative WaitGroup counter
```

**Правило**: `Done()` без `Add()` — это паника.

## 🛠️ Типичные паттерны

### 1. Fan-out (параллельная обработка)
```go
func processAll(items []Item) {
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(it Item) {
            defer wg.Done()
            process(it)
        }(item)
    }
    wg.Wait()
}
```

### 2. Fan-out + сбор результатов
```go
func sumAll(nums []int) int {
    results := make(chan int, len(nums))
    var wg sync.WaitGroup
    
    for _, n := range nums {
        wg.Add(1)
        go func(val int) {
            defer wg.Done()
            results <- compute(val)
        }(n)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    sum := 0
    for r := range results {
        sum += r
    }
    return sum
}
```

### 3. Таймаут с WaitGroup
```go
func processWithTimeout(items []Item, timeout time.Duration) error {
    var wg sync.WaitGroup
    done := make(chan struct{})
    
    for _, item := range items {
        wg.Add(1)
        go func(it Item) {
            defer wg.Done()
            process(it)
        }(item)
    }
    
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return errors.New("timeout")
    }
}
```

## 🎓 Вопросы для медитации

1. Что произойдет, если вызвать `wg.Add(5)` один раз вместо 5 раз `wg.Add(1)`?
2. Можно ли использовать один WaitGroup для нескольких циклов горутин?
3. Чем отличается `wg.Add(1); go f()` от `go func() { wg.Add(1); f() }()`?
4. Как реализовать `WaitGroup` с таймаутом без `select`?

## 🔗 Альтернативы WaitGroup

| Паттерн | Когда использовать |
|---------|-------------------|
| **Channel + close** | Когда нужен сигнал завершения одной горутины |
| **errgroup.Group** | Когда нужно прервать все при первой ошибке |
| **context.Context** | Когда нужна отмена операций |
| **sync.WaitGroup** | Когда просто нужно дождаться N горутин |

---

**Задание**: Реализуй функцию `func downloadAll(urls []string) []Result`, которая:
1. Скачивает все URL параллельно
2. Возвращает результаты **в том же порядке**, что и URLs
3. Не использует `time.Sleep` или фиксированные таймауты

Объясни, почему нельзя использовать `wg.Wait()` внутри `for` цикла с `range` по каналу.
