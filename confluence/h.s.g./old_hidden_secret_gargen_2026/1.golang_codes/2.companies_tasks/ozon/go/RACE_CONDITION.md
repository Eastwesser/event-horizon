# Race Condition

## Грейд
18-19

## Вопрос
Что такое race condition? Как её обнаружить и исправить?

## Ответ

**Race condition** — состояние гонки, когда несколько горутин одновременно обращаются к общим данным, и хотя бы одна из них выполняет запись, без синхронизации.

### Пример с ошибкой

```go
func main() {
    counter := 0
    for i := 0; i < 1000; i++ {
        go func() {
            counter++ // race condition!
        }()
    }
    time.Sleep(time.Second)
    fmt.Println(counter) // не 1000 (может быть 987, 1000, 1023)
}
```

## Дополнительные вопросы

### Как обнаружить race condition?

```bash
go run -race main.go
go build -race
go test -race
```

Детектор расставляет точки наблюдения за памятью и сообщает, где произошёл конфликт.

### Как исправить race condition?

```text
Способ	Код
Mutex	mu.Lock(); counter++; mu.Unlock()
Atomic	atomic.AddInt64(&counter, 1)
Канал	ch <- 1 и counter += <-ch
```

### Какие знаешь средства для предотвращения race condition?

- Mutex / RWMutex — эксклюзивный доступ
- Atomic операции — для счётчиков, флагов
- Каналы — передача владения данными
- sync.Once — однократная инициализация
- sync.WaitGroup — ожидание завершения

## Что важно сказать на собеседовании

### Почему counter++ — это проблема?

counter++ не атомарный. Он состоит из:

- Чтение counter
- Инкремент
- Запись counter

Две горутины могут прочитать одно и то же значение, увеличить его и записать — один инкремент потеряется.

### Что будет в коде с map?

```go
m := make(map[int]int)
go func() { m[1] = 2 }()
go func() { m[1] = 7 }()
// panic: concurrent map writes
```
Map в Go небезопасен для конкурентной записи (упадёт с паникой).

### Пример исправления
```go
var mu sync.Mutex
var counter int

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// или с atomic
var counter atomic.Int64

func increment() {
    counter.Add(1)
}
```
