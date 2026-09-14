# Promise: Инсайты от Мусаси

## 🔑 Суть паттерна

**Promise (промис)** — паттерн для работы с асинхронными операциями, который представляет значение, которое будет доступно в будущем (или ошибку).

**Принцип:** Запускаем асинхронную операцию, получаем "обещание" результата, обрабатываем его когда он готов.

**Аналогия:** Ты заказываешь пиццу (Promise). Пока она готовится, можешь делать другие дела. Когда пицца готова, ты получаешь её (success) или узнаёшь, что заказ отменён (error).

## 📊 Когда использовать?

- **Асинхронные операции:** HTTP-запросы, чтение файлов, запросы к БД
- **Обработка результатов:** когда нужно обработать результат после завершения операции
- **Цепочки операций:** можно комбинировать несколько промисов
- **Обработка ошибок:** централизованная обработка ошибок асинхронных операций

**Примеры:**
- HTTP-клиент: запрос → Promise → обработка ответа
- Файловая система: чтение файла → Promise → обработка данных
- База данных: запрос → Promise → обработка результата

## 🏗️ Архитектура

```
[Асинхронная функция] → [Promise] → [Обработчики]
                           ↓
                    [Канал результата]
                           ↓
              [Success Handler] или [Error Handler]
```

**Важно:** Promise запускает операцию сразу при создании, не дожидаясь вызова `Then()`.

## ⚙️ Реализация

### Базовая реализация

```go
type result[T any] struct {
    val T
    err error
}

type Promise[T any] struct {
    resultCh chan result[T]
}

func NewPromise[T any](asyncFn func() (T, error)) *Promise[T] {
    promise := &Promise[T]{
        resultCh: make(chan result[T], 1), // Буферизованный канал
    }

    go func() {
        defer close(promise.resultCh)
        val, err := asyncFn()
        promise.resultCh <- result[T]{val: val, err: err}
    }()

    return promise
}

func (p *Promise[T]) Then(successFn func(T), errorFn func(error)) {
    go func() {
        result := <-p.resultCh
        if result.err == nil {
            successFn(result.val)
        } else {
            errorFn(result.err)
        }
    }()
}

func (p *Promise[T]) Await() (T, error) {
    result := <-p.resultCh
    return result.val, result.err
}
```

### Использование

```go
// Вариант 1: Then (неблокирующий)
promise := NewPromise(func() (string, error) {
    return fetchData()
})

promise.Then(
    func(value string) {
        fmt.Println("Success:", value)
    },
    func(err error) {
        fmt.Println("Error:", err)
    },
)

// Вариант 2: Await (блокирующий)
promise := NewPromise(func() (int, error) {
    return calculate()
})

value, err := promise.Await()
if err != nil {
    // Обработка ошибки
}
```

## 🔄 Promise vs Future

| Аспект | Promise | Future |
|--------|---------|--------|
| **Запуск** | Сразу при создании | Ленивый (при первом обращении) |
| **Когда использовать** | Когда нужно запустить сразу | Когда нужно отложить выполнение |
| **Реализация** | Горутина запускается в конструкторе | Горутина запускается при первом `Await()` |

**В Go:** Обычно используют Promise (запуск сразу), так как это проще и предсказуемее.

## 🔄 Promise vs Callback

| Аспект | Promise | Callback |
|--------|---------|----------|
| **Стиль** | Декларативный | Императивный |
| **Цепочки** | Легко комбинировать | Сложно (callback hell) |
| **Ошибки** | Централизованная обработка | Нужно передавать в каждый callback |

**Пример callback hell:**
```go
// ПЛОХО: Callback hell
fetchData(func(data string) {
    processData(data, func(result int) {
        saveResult(result, func(err error) {
            // Глубокая вложенность
        })
    })
})

// ХОРОШО: Promise
promise := NewPromise(fetchData)
promise.Then(processData, handleError)
```

## ⚠️ Ловушки

### 1. Программа завершается до выполнения Promise

**Проблема:** Если не ждать завершения Promise, программа может завершиться до выполнения.

```go
// ПЛОХО: программа завершится до выполнения
promise := NewPromise(asyncJob)
promise.Then(success, error)
// Программа завершается здесь

// ХОРОШО: ждём завершения
promise := NewPromise(asyncJob)
promise.Then(success, error)
time.Sleep(2 * time.Second) // Или использовать Await()
```

**Решение:** Использовать `Await()` или `sync.WaitGroup` для ожидания.

### 2. Множественное чтение из канала

**Проблема:** Если несколько горутин читают из одного канала, результат будет прочитан только один раз.

```go
// ПЛОХО: только одна горутина получит результат
promise := NewPromise(asyncJob)
go func() { result := <-promise.resultCh }()
go func() { result := <-promise.resultCh }() // Эта горутина не получит результат

// ХОРОШО: используем Await() или один Then()
promise := NewPromise(asyncJob)
value, err := promise.Await() // Блокирующий вызов
```

**Решение:** Использовать `Await()` для блокирующего чтения или один `Then()`.

### 3. Буферизованный vs небуферизованный канал

**Вопрос:** Почему используется буферизованный канал `make(chan result[T], 1)`?

**Ответ:** 
- **Буферизованный:** Результат может быть записан до того, как кто-то начнёт читать
- **Небуферизованный:** Запись блокируется, пока кто-то не начнёт читать

**Пример:**
```go
// Буферизованный: результат записывается сразу
resultCh: make(chan result[T], 1)

// Если результат готов до вызова Then(), он сохраняется в буфере
promise := NewPromise(fastJob) // Выполняется быстро
time.Sleep(100 * time.Millisecond)
promise.Then(success, error) // Результат уже в буфере
```

### 4. Закрытие канала

**Проблема:** Если канал закрыт, чтение из него вернёт zero value.

**Решение:** Всегда закрывать канал после записи (`defer close()`).

```go
go func() {
    defer close(promise.resultCh) // Гарантированно закроем канал
    val, err := asyncFn()
    promise.resultCh <- result[T]{val: val, err: err}
}()
```

## 💡 Вопросы для размышления

1. **Почему Promise запускается сразу при создании?** (Чтобы не терять время на ожидание)
2. **Можно ли отменить Promise?** (Да, используя `context.Context`)
3. **В чём разница между `Then()` и `Await()`?** (`Then()` — неблокирующий, `Await()` — блокирующий)
4. **Можно ли комбинировать несколько Promise?** (Да, можно создать `PromiseAll` или `PromiseRace`)
5. **Почему используется буферизованный канал?** (Чтобы результат мог быть записан до чтения)

## 🔧 Практические примеры

### Пример 1: HTTP-запрос

```go
promise := NewPromise(func() ([]byte, error) {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
})

promise.Then(
    func(data []byte) {
        fmt.Println("Data received:", len(data))
    },
    func(err error) {
        fmt.Println("Error:", err)
    },
)
```

### Пример 2: Чтение файла

```go
promise := NewPromise(func() (string, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return "", err
    }
    return string(data), nil
})

content, err := promise.Await()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Config:", content)
```

### Пример 3: Запрос к БД

```go
promise := NewPromise(func() ([]User, error) {
    var users []User
    err := db.Find(&users).Error
    return users, err
})

promise.Then(
    func(users []User) {
        fmt.Printf("Found %d users\n", len(users))
    },
    func(err error) {
        fmt.Println("DB error:", err)
    },
)
```

### Пример 4: Комбинация Promise (PromiseAll)

```go
func PromiseAll[T any](promises []*Promise[T]) *Promise[[]T] {
    return NewPromise(func() ([]T, error) {
        results := make([]T, len(promises))
        for i, p := range promises {
            val, err := p.Await()
            if err != nil {
                return nil, err
            }
            results[i] = val
        }
        return results, nil
    })
}

// Использование
promises := []*Promise[string]{
    NewPromise(fetchData1),
    NewPromise(fetchData2),
    NewPromise(fetchData3),
}

allPromise := PromiseAll(promises)
results, err := allPromise.Await()
```

### Пример 5: Promise с контекстом (отмена)

```go
func NewPromiseWithContext[T any](ctx context.Context, asyncFn func() (T, error)) *Promise[T] {
    promise := &Promise[T]{
        resultCh: make(chan result[T], 1),
    }

    go func() {
        defer close(promise.resultCh)
        
        // Канал для результата
        done := make(chan result[T], 1)
        go func() {
            val, err := asyncFn()
            done <- result[T]{val: val, err: err}
        }()

        select {
        case <-ctx.Done():
            var zero T
            promise.resultCh <- result[T]{val: zero, err: ctx.Err()}
        case result := <-done:
            promise.resultCh <- result
        }
    }()

    return promise
}
```

## 🎯 Преимущества Promise

1. **Декларативность:** описываем ЧТО делать, а не КАК
2. **Обработка ошибок:** централизованная обработка через error handler
3. **Композиция:** легко комбинировать несколько Promise
4. **Читаемость:** нет callback hell

## 🔄 Promise vs Другие паттерны

| Паттерн | Назначение | Когда использовать |
|---------|-----------|-------------------|
| **Promise** | Асинхронная операция с обработчиками | Когда нужно обработать результат |
| **Channel** | Коммуникация между горутинами | Когда нужно передавать данные |
| **Future** | Ленивая асинхронная операция | Когда нужно отложить выполнение |

## 🔧 Расширенные возможности

### Promise.Then с цепочкой

```go
func (p *Promise[T]) ThenChain(next func(T) (U, error)) *Promise[U] {
    return NewPromise(func() (U, error) {
        val, err := p.Await()
        if err != nil {
            var zero U
            return zero, err
        }
        return next(val)
    })
}

// Использование
promise := NewPromise(fetchData)
    .ThenChain(processData)
    .ThenChain(saveData)

result, err := promise.Await()
```

### Promise.Race (первый завершившийся)

```go
func PromiseRace[T any](promises []*Promise[T]) *Promise[T] {
    return NewPromise(func() (T, error) {
        resultCh := make(chan result[T], len(promises))
        
        for _, p := range promises {
            go func(promise *Promise[T]) {
                val, err := promise.Await()
                resultCh <- result[T]{val: val, err: err}
            }(p)
        }
        
        return (<-resultCh).val, (<-resultCh).err
    })
}
```

---

*"Promise — это не просто обёртка над горутиной. Это способ думать об асинхронности как о потоке данных, который может завершиться успехом или ошибкой. Это мост между синхронным и асинхронным миром."* — Мусаси
