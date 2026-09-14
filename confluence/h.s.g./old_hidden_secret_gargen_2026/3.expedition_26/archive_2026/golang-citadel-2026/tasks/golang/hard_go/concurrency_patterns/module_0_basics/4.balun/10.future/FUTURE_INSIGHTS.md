# Future: Инсайты от Мусаси

## 🔑 Суть паттерна

**Future (фьючер)** — паттерн для работы с асинхронными операциями, который представляет значение, которое будет доступно в будущем.

**Принцип:** Создаём Future, запускаем операцию, получаем результат через блокирующий вызов `Get()`.

**Аналогия:** Ты заказываешь пиццу. Пока она готовится, можешь делать другие дела. Когда забираешь пиццу (`Get()`), получаешь результат.

## 📊 Когда использовать?

- **Асинхронные операции:** когда нужно выполнить операцию и получить результат позже
- **Блокирующее ожидание:** когда нужно дождаться результата перед продолжением
- **Упрощение работы с каналами:** инкапсуляция канала в удобный API

**Примеры:**
- HTTP-запросы: запрос → Future → Get() → результат
- Файловая система: чтение файла → Future → Get() → данные
- База данных: запрос → Future → Get() → результат

## 🏗️ Архитектура

```
[Асинхронная функция] → [Future] → [Get()] → [Результат]
                           ↓
                    [Канал результата]
                           ↓
                    [Блокирующее чтение]
```

**Важно:** Future запускает операцию сразу при создании (в отличие от "ленивого" Future, который запускается при первом `Get()`).

## ⚙️ Реализация

### Базовая реализация

```go
type Future[T any] struct {
    resultCh chan T
}

func NewFuture[T any](action func() T) *Future[T] {
    future := &Future[T]{
        resultCh: make(chan T, 1), // Буферизованный канал
    }

    go func() {
        defer close(future.resultCh)
        future.resultCh <- action() // Запускаем операцию сразу
    }()

    return future
}

func (f *Future[T]) Get() T {
    return <-f.resultCh // Блокирующий вызов
}
```

### Использование

```go
asyncJob := func() string {
    time.Sleep(time.Second)
    return "success"
}

future := NewFuture(asyncJob)
result := future.Get() // Блокируемся, пока результат не будет готов
fmt.Println(result)
```

## 🔄 Future vs Promise

| Аспект | Future | Promise |
|--------|--------|---------|
| **Запуск** | Сразу при создании | Сразу при создании |
| **Получение результата** | Блокирующий `Get()` | Неблокирующий `Then()` или блокирующий `Await()` |
| **Обработка ошибок** | Через возвращаемое значение | Через error handler |
| **API** | Проще (один метод) | Сложнее (Then, Await) |

**Пример:**
```go
// Future: простой блокирующий вызов
future := NewFuture(fetchData)
result := future.Get() // Блокируемся

// Promise: обработчики или блокирующий вызов
promise := NewPromise(fetchData)
promise.Then(success, error) // Неблокирующий
// или
result, err := promise.Await() // Блокирующий
```

## 🔄 Future vs Promise (детальное сравнение)

### Future (упрощённый подход)

```go
future := NewFuture(func() string {
    return fetchData()
})

result := future.Get() // Простой блокирующий вызов
```

**Плюсы:**
- Простой API (один метод `Get()`)
- Легко понять и использовать
- Подходит для простых случаев

**Минусы:**
- Нет обработки ошибок (нужно возвращать `(T, error)`)
- Нет неблокирующего варианта
- Нет цепочек операций

### Promise (более гибкий подход)

```go
promise := NewPromise(func() (string, error) {
    return fetchData()
})

promise.Then(
    func(value string) { /* success */ },
    func(err error) { /* error */ },
)
```

**Плюсы:**
- Обработка ошибок встроена
- Неблокирующий вариант (`Then`)
- Можно комбинировать в цепочки

**Минусы:**
- Более сложный API
- Больше кода для простых случаев

## ⚠️ Ловушки

### 1. Буферизованный канал

**Вопрос:** Почему используется буферизованный канал `make(chan T, 1)`?

**Ответ:** Чтобы результат мог быть записан до вызова `Get()`. Если результат готов до `Get()`, он сохраняется в буфере.

```go
// Буферизованный: результат может быть записан до Get()
resultCh: make(chan T, 1)

// Если операция выполнилась быстро
future := NewFuture(fastJob) // Результат уже в буфере
time.Sleep(100 * time.Millisecond)
result := future.Get() // Читаем из буфера
```

### 2. Множественный вызов Get()

**Проблема:** Если вызвать `Get()` несколько раз, только первый вызов получит результат.

**Решение:** Использовать буферизованный канал и кэшировать результат.

```go
type Future[T any] struct {
    resultCh chan T
    result   *T      // Кэшированный результат
    once     sync.Once
}

func (f *Future[T]) Get() T {
    f.once.Do(func() {
        f.result = new(T)
        *f.result = <-f.resultCh
    })
    return *f.result
}
```

### 3. Обработка ошибок

**Проблема:** Базовая реализация Future не поддерживает ошибки.

**Решение:** Использовать `Future[Result[T]]` или версию с ошибкой.

```go
type Result[T any] struct {
    Value T
    Error error
}

type Future[T any] struct {
    resultCh chan Result[T]
}

func (f *Future[T]) Get() (T, error) {
    result := <-f.resultCh
    return result.Value, result.Error
}
```

### 4. Программа завершается до выполнения

**Проблема:** Если не вызвать `Get()`, программа может завершиться до выполнения операции.

**Решение:** Всегда вызывать `Get()` или использовать `sync.WaitGroup`.

```go
// ПЛОХО: программа завершится до выполнения
future := NewFuture(asyncJob)
// Забыли вызвать Get()

// ХОРОШО: ждём результат
future := NewFuture(asyncJob)
result := future.Get() // Блокируемся до получения результата
```

## 💡 Вопросы для размышления

1. **В чём разница между Future и Promise?** (Future — простой блокирующий вызов. Promise — обработчики или блокирующий вызов)
2. **Почему используется буферизованный канал?** (Чтобы результат мог быть записан до вызова `Get()`)
3. **Можно ли сделать Future с обработкой ошибок?** (Да, использовать `Future[Result[T]]` или версию с `(T, error)`)
4. **В чём разница между "ленивым" и "eager" Future?** (Ленивый запускается при первом `Get()`, eager — сразу при создании)
5. **Можно ли отменить Future?** (Да, используя `context.Context`)

## 🔧 Практические примеры

### Пример 1: HTTP-запрос

```go
future := NewFuture(func() []byte {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return nil
    }
    defer resp.Body.Close()
    data, _ := io.ReadAll(resp.Body)
    return data
})

data := future.Get()
fmt.Println("Data received:", len(data))
```

### Пример 2: Чтение файла

```go
future := NewFuture(func() string {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return ""
    }
    return string(data)
})

content := future.Get()
fmt.Println("Config:", content)
```

### Пример 3: Запрос к БД

```go
future := NewFuture(func() []User {
    var users []User
    db.Find(&users)
    return users
})

users := future.Get()
fmt.Printf("Found %d users\n", len(users))
```

### Пример 4: Future с ошибками

```go
type FutureWithError[T any] struct {
    resultCh chan Result[T]
}

type Result[T any] struct {
    Value T
    Error error
}

func NewFutureWithError[T any](action func() (T, error)) *FutureWithError[T] {
    future := &FutureWithError[T]{
        resultCh: make(chan Result[T], 1),
    }

    go func() {
        defer close(future.resultCh)
        val, err := action()
        future.resultCh <- Result[T]{Value: val, Error: err}
    }()

    return future
}

func (f *FutureWithError[T]) Get() (T, error) {
    result := <-f.resultCh
    return result.Value, result.Error
}
```

### Пример 5: Future с Promise (комбинация)

```go
// Promise создаёт Future
type Promise[T any] struct {
    resultCh chan T
}

func NewPromise[T any]() *Promise[T] {
    return &Promise[T]{
        resultCh: make(chan T, 1),
    }
}

func (p *Promise[T]) Set(value T) {
    p.resultCh <- value
    close(p.resultCh)
}

func (p *Promise[T]) GetFuture() *Future[T] {
    return &Future[T]{
        resultCh: p.resultCh,
    }
}

// Использование
promise := NewPromise[string]()
go func() {
    time.Sleep(time.Second)
    promise.Set("agreement") // Устанавливаем значение
}()

future := promise.GetFuture() // Получаем Future
result := future.Get()        // Ждём результат
fmt.Println(result)
```

## 🎯 Преимущества Future

1. **Простота:** простой API с одним методом `Get()`
2. **Инкапсуляция:** скрывает детали работы с каналами
3. **Блокирующее ожидание:** удобно, когда нужно дождаться результата
4. **Типобезопасность:** использование дженериков для типобезопасности

## 🔄 Future vs Другие паттерны

| Паттерн | Назначение | Когда использовать |
|---------|-----------|-------------------|
| **Future** | Блокирующее ожидание результата | Когда нужно дождаться результата |
| **Promise** | Асинхронная операция с обработчиками | Когда нужна обработка ошибок и цепочки |
| **Channel** | Коммуникация между горутинами | Когда нужно передавать данные |

## 🔧 Ленивый Future (запуск при первом Get())

```go
type LazyFuture[T any] struct {
    resultCh chan T
    action   func() T
    once     sync.Once
}

func NewLazyFuture[T any](action func() T) *LazyFuture[T] {
    return &LazyFuture[T]{
        resultCh: make(chan T, 1),
        action:   action,
    }
}

func (f *LazyFuture[T]) Get() T {
    f.once.Do(func() {
        // Запускаем операцию только при первом вызове Get()
        go func() {
            defer close(f.resultCh)
            f.resultCh <- f.action()
        }()
    })
    
    return <-f.resultCh
}
```

**Разница:**
- **Eager Future:** Операция запускается сразу при создании
- **Lazy Future:** Операция запускается при первом `Get()`

**Когда использовать:**
- **Eager:** Когда операцию нужно запустить сразу
- **Lazy:** Когда операцию нужно запустить только при необходимости

## 🔧 Future с контекстом (отмена)

```go
func NewFutureWithContext[T any](ctx context.Context, action func() T) *Future[T] {
    future := &Future[T]{
        resultCh: make(chan T, 1),
    }

    go func() {
        defer close(future.resultCh)
        
        // Канал для результата
        done := make(chan T, 1)
        go func() {
            done <- action()
        }()

        select {
        case <-ctx.Done():
            var zero T
            future.resultCh <- zero // Возвращаем zero value при отмене
        case result := <-done:
            future.resultCh <- result
        }
    }()

    return future
}
```

---

*"Future — это не просто обёртка над каналом. Это способ думать об асинхронности как о значении, которое будет доступно в будущем. Это мост между синхронным и асинхронным миром, который делает асинхронность простой."* — Мусаси
