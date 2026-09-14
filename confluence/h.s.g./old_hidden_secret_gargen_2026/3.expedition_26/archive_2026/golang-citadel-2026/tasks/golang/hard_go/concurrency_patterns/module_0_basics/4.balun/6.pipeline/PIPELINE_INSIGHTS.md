# Pipeline: Инсайты от Мусаси

## 🔑 Суть паттерна

**Pipeline (конвейер)** — цепочка независимых стадий обработки данных, где выход одной стадии становится входом следующей.

**Принцип:** `ch → stage1 → ch → stage2 → ch → stage3 → ...`

**Аналогия:** Конвейер на заводе — каждая станция выполняет свою работу, передавая результат дальше.

## 📊 Когда использовать?

- **Многостадийная обработка:** данные проходят через несколько этапов
- **Параллельная обработка:** пока одна стадия обрабатывает элемент, следующая может обрабатывать предыдущий
- **Разделение ответственности:** каждая стадия делает одну вещь
- **Масштабируемость:** можно добавить/убрать стадии без изменения остальных

**Примеры:**
- Обработка логов: парсинг → валидация → маскировка → аналитика → БД
- Обработка изображений: загрузка → ресайз → фильтр → сжатие → сохранение
- ETL (Extract, Transform, Load): извлечение → преобразование → загрузка

## 🏗️ Архитектура

```
[Источник] → [Стадия 1] → [Стадия 2] → [Стадия 3] → [Приёмник]
   ch1          ch2          ch3          ch4
   
Каждая стадия:
- Работает в своей горутине
- Читает из входного канала
- Обрабатывает данные
- Пишет в выходной канал
- Закрывает выходной канал при завершении
```

**Важно:** Стадии работают **параллельно**! Пока стадия 1 обрабатывает элемент N, стадия 2 может обрабатывать элемент N-1.

## ⚙️ Реализация

### Базовый Pipeline

```go
// Стадия 1: Генерация данных
func generate[T any](values ...T) <-chan T {
    outputCh := make(chan T)
    go func() {
        defer close(outputCh)
        for _, value := range values {
            outputCh <- value
        }
    }()
    return outputCh
}

// Стадия 2: Обработка данных
func process[T any](inputCh <-chan T, action func(T) T) <-chan T {
    outputCh := make(chan T)
    go func() {
        defer close(outputCh)
        for value := range inputCh {
            outputCh <- action(value)
        }
    }()
    return outputCh
}

// Использование
for value := range process(generate(1, 2, 3), square) {
    fmt.Println(value)
}
```

### Композиция стадий

```go
// Можно комбинировать любые паттерны
pipeline := process(
    filter(
        generate(1, 2, 3, 4, 5),
        isEven,
    ),
    square,
)
```

## 🔄 Pipeline vs Последовательная обработка

| Аспект | Pipeline | Последовательная обработка |
|--------|----------|---------------------------|
| **Параллелизм** | Да (стадии работают параллельно) | Нет (всё последовательно) |
| **Буферизация** | Каналы между стадиями | Нет |
| **Производительность** | Выше (параллельная обработка) | Ниже |
| **Сложность** | Выше (нужно управлять каналами) | Ниже |

**Пример:**
```go
// Последовательно (медленнее)
for _, v := range values {
    result := process1(process2(process3(v)))
    fmt.Println(result)
}

// Pipeline (быстрее, параллельно)
for result := range process1(process2(process3(generate(values...)))) {
    fmt.Println(result)
}
```

## 🎯 Компоненты Pipeline

### 1. Источник (Source/Generator)
Генерирует данные и отправляет в канал.

```go
func generate[T any](values ...T) <-chan T {
    outputCh := make(chan T)
    go func() {
        defer close(outputCh)
        for _, value := range values {
            outputCh <- value
        }
    }()
    return outputCh
}
```

### 2. Стадия обработки (Stage)
Читает из входного канала, обрабатывает, пишет в выходной.

```go
func stage[T any](inputCh <-chan T, transform func(T) T) <-chan T {
    outputCh := make(chan T)
    go func() {
        defer close(outputCh)
        for value := range inputCh {
            outputCh <- transform(value)
        }
    }()
    return outputCh
}
```

### 3. Приёмник (Sink)
Читает результаты из последней стадии.

```go
for value := range pipeline {
    // Обработка результата
    fmt.Println(value)
}
```

## 🔗 Комбинация с другими паттернами (LEGO-конструктор)

**Ключевая идея:** Паттерны можно комбинировать как детали LEGO. Pipeline — это конструктор, который собирает другие паттерны в единую систему.

### Pipeline + Filter
```go
// Фильтруем чётные, потом возводим в квадрат
pipeline := process(
    filter(generate(1, 2, 3, 4, 5), isEven),
    square,
)
```

### Pipeline + Decorator
```go
// Преобразуем данные на каждой стадии
pipeline := process(
    process(generate(1, 2, 3), multiplyBy2),
    addOne,
)
```

### Pipeline + Fan-Out + Fan-In
```go
// Распределяем по нескольким обработчикам, потом собираем
pipeline := fanIn(
    process(fanOut(generate(1, 2, 3), 3), heavyProcessing),
)
```

### Pipeline + Масштабирование стадий (Fan-Out внутри стадии)

**Важный паттерн:** Одна стадия Pipeline может сама использовать Fan-Out для параллельной обработки.

**Два подхода к масштабированию:**

#### Подход 1: Все воркеры читают из одного канала (конкуренция)

```go
// Стадия parse — обычная обработка
func parse(inputCh <-chan string) <-chan string {
    outputCh := make(chan string)
    go func() {
        defer close(outputCh)
        for data := range inputCh {
            outputCh <- fmt.Sprintf("parsed - %s", data)
        }
    }()
    return outputCh
}

// Стадия send — масштабируется через n воркеров
// ВСЕ воркеры читают из одного канала inputCh (конкуренция за данные)
func send(inputCh <-chan string, n int) <-chan string {
    outputCh := make(chan string)
    var wg sync.WaitGroup
    wg.Add(n)
    
    // n горутин конкурируют за чтение из одного канала
    for i := 0; i < n; i++ {
        go func(workerID int) {
            defer wg.Done()
            for data := range inputCh {
                outputCh <- fmt.Sprintf("data sent: %s by worker %d", data, workerID)
            }
        }(i) // Важно: передаём i как параметр, чтобы избежать замыкания
    }
    
    go func() {
        wg.Wait()
        close(outputCh)
    }()
    
    return outputCh
}
```

**Плюсы:** Простая реализация, автоматическая балансировка нагрузки
**Минусы:** Конкуренция за данные, возможна блокировка, если один воркер медленный

#### Подход 2: Каждый воркер читает из своего канала (изоляция через Fan-Out)

```go
// PipelineSplitChannel — распределяет данные по n каналам (Round Robin)
func PipelineSplitChannel[T any](inputCh <-chan T, n int) []<-chan T {
    outputChannels := make([]chan T, n)
    for i := 0; i < n; i++ {
        outputChannels[i] = make(chan T)
    }
    
    go func() {
        idx := 0
        for value := range inputCh {
            outputChannels[idx] <- value // Round Robin распределение
            idx = (idx + 1) % n
        }
        
        for _, ch := range outputChannels {
            close(ch)
        }
    }()
    
    resultChannels := make([]<-chan T, n)
    for i := 0; i < n; i++ {
        resultChannels[i] = outputChannels[i]
    }
    return resultChannels
}

// Стадия send — каждый воркер читает из своего канала (изоляция)
func send(inputCh <-chan string, n int) <-chan string {
    var wg sync.WaitGroup
    wg.Add(n)
    outputCh := make(chan string)
    splitChs := PipelineSplitChannel(inputCh, n) // Разбиваем поток на n каналов
    
    // Каждая горутина читает из своего канала (нет конкуренции)
    for i := 0; i < n; i++ {
        go func(idx int) {
            defer wg.Done()
            for data := range splitChs[idx] {
                outputCh <- fmt.Sprintf("data sent: %s by worker %d", data, idx)
            }
        }(i)
    }
    
    go func() {
        wg.Wait()
        close(outputCh)
    }()
    
    return outputCh
}
```

**Плюсы:** Нет конкуренции, изоляция данных, предсказуемое распределение
**Минусы:** Более сложная реализация, Round Robin может быть неравномерным

**Когда использовать какой подход:**
- **Подход 1 (конкуренция):** Когда обработка быстрая, данные однородные, нужна простая реализация
- **Подход 2 (изоляция):** Когда обработка медленная, нужна изоляция, важно равномерное распределение

**Преимущества масштабирования стадий:**
- Можно масштабировать отдельные стадии независимо
- Медленные стадии можно ускорить, увеличив параллельный фактор
- Каждая стадия может иметь свой уровень параллелизма

## ⚠️ Ловушки

### 1. Незакрытые каналы

**Проблема:** Если канал не закрыт, следующая стадия будет ждать бесконечно.

**Решение:**
```go
go func() {
    defer close(outputCh)  // Всегда закрываем при завершении
    for value := range inputCh {
        outputCh <- process(value)
    }
}()
```

### 2. Утечки горутин

**Проблема:** Если приёмник не читает из канала, стадии заблокируются.

**Решение:** Всегда читайте из последнего канала в Pipeline.

```go
// Плохо: забыли прочитать
pipeline := process(generate(1, 2, 3), square)
// Горутины зависнут!

// Хорошо: читаем результаты
for value := range pipeline {
    fmt.Println(value)
}
```

### 3. Буферизованные каналы

**Вопрос:** Нужны ли буферы между стадиями?

**Ответ:** Зависит от задачи:
- **Без буфера:** стадии работают синхронно (одна ждёт другую)
- **С буфером:** стадии могут работать асинхронно (одна обрабатывает, другая уже читает следующее)

```go
// Без буфера (синхронно)
outputCh := make(chan T)

// С буфером (асинхронно)
outputCh := make(chan T, 10)
```

### 4. Ошибки в стадиях

**Проблема:** Если стадия паникует, канал не закроется, и Pipeline зависнет.

**Решение:** Обрабатывать ошибки и закрывать каналы.

```go
go func() {
    defer close(outputCh)
    defer func() {
        if r := recover(); r != nil {
            // Логируем ошибку, но не паникуем
            log.Printf("Stage panicked: %v", r)
        }
    }()
    for value := range inputCh {
        outputCh <- process(value)
    }
}()
```

### 5. Замыкание в циклах (closure)

**Проблема:** При использовании переменной цикла в горутине все горутины могут использовать последнее значение.

```go
// ПЛОХО: все горутины будут использовать последнее значение i (например, 2)
for i := 0; i < n; i++ {
    go func() {
        fmt.Println(i) // Все выведут 2 (или n-1)
    }()
}

// ХОРОШО: передаём i как параметр
for i := 0; i < n; i++ {
    go func(workerID int) {
        fmt.Println(workerID) // Каждая выведет своё значение (0, 1, 2...)
    }(i)
}
```

**Правило:** Всегда передавайте переменные цикла как параметры в горутины.

### 6. Конкуренция vs Изоляция в масштабируемых стадиях

**Проблема:** Если все воркеры читают из одного канала, медленный воркер может заблокировать остальных.

**Решение:** Использовать `PipelineSplitChannel` для изоляции данных по каналам.

```go
// Конкуренция: все читают из одного канала
for data := range inputCh { // Может заблокироваться, если один воркер медленный
    process(data)
}

// Изоляция: каждый читает из своего канала
splitChs := PipelineSplitChannel(inputCh, n)
for data := range splitChs[idx] { // Изолированная обработка
    process(data)
}
```

## 💡 Вопросы для размышления

1. **Можно ли сделать Pipeline с переменным количеством стадий?** (Да, используя слайс функций)
2. **Что произойдёт, если одна стадия работает медленнее остальных?** (Она станет "узким местом", остальные будут ждать. Решение: масштабировать эту стадию через Fan-Out)
3. **Можно ли сделать Pipeline, который обрабатывает ошибки?** (Да, использовать канал `chan Result` с полем `error`)
4. **Как измерить производительность Pipeline?** (Засечь время обработки, посчитать throughput)
5. **Почему в стадии `send` нужно передавать `i` как параметр?** (Чтобы избежать замыкания — все горутины будут использовать последнее значение `i`)
6. **Как определить, какую стадию нужно масштабировать?** (Измерить время обработки каждой стадии, найти узкое место)
7. **В чём разница между конкуренцией и изоляцией в масштабируемых стадиях?** (Конкуренция: все читают из одного канала. Изоляция: каждый читает из своего канала через Fan-Out)
8. **Когда использовать Round Robin, а когда случайное распределение?** (Round Robin — простой и равномерный. Случайное — когда нужна истинная случайность, но сложнее реализовать)

## 🔧 Практические примеры

### Пример 1: Обработка логов
```go
logs := generateLogs()
validated := validate(logs)
masked := maskSecrets(validated)
analyzed := analyze(masked)

for result := range saveToDB(analyzed) {
    fmt.Printf("Saved: %v\n", result)
}
```

### Пример 2: Обработка изображений
```go
images := loadImages()
resized := resize(images, 800, 600)
filtered := applyFilter(resized)
compressed := compress(filtered)

for img := range saveImages(compressed) {
    fmt.Printf("Saved: %s\n", img.Path)
}
```

### Пример 3: ETL Pipeline
```go
rawData := extractFromSource()
cleaned := cleanData(rawData)
transformed := transform(cleaned)
enriched := enrich(transformed)

for record := range loadToDB(enriched) {
    fmt.Printf("Loaded: %v\n", record)
}
```

### Пример 4: Pipeline с несколькими стадиями
```go
// Генерируем → Фильтруем → Преобразуем → Фильтруем → Выводим
pipeline := filter(
    process(
        filter(
            generate(1, 2, 3, 4, 5),
            isEven,
        ),
        square,
    ),
    greaterThan(10),
)

for value := range pipeline {
    fmt.Println(value) // 16 (только 4^2 > 10)
}
```

### Пример 5: Масштабируемый Pipeline (конкуренция за канал)
```go
// Обработка логов: парсинг → отправка (все воркеры читают из одного канала)
func parse(inputCh <-chan string) <-chan string {
    outputCh := make(chan string)
    go func() {
        defer close(outputCh)
        for data := range inputCh {
            outputCh <- fmt.Sprintf("parsed - %s", data)
        }
    }()
    return outputCh
}

func send(inputCh <-chan string, workers int) <-chan string {
    outputCh := make(chan string)
    var wg sync.WaitGroup
    wg.Add(workers)
    
    // Все воркеры конкурируют за чтение из inputCh
    for i := 0; i < workers; i++ {
        go func(workerID int) {
            defer wg.Done()
            for data := range inputCh {
                outputCh <- fmt.Sprintf("sent: %s by worker %d", data, workerID)
            }
        }(i)
    }
    
    go func() {
        wg.Wait()
        close(outputCh)
    }()
    
    return outputCh
}

// Использование: parse → send (5 воркеров для отправки)
logs := generateLogs()
for result := range send(parse(logs), 5) {
    fmt.Println(result)
}
```

### Пример 6: Масштабируемый Pipeline (изоляция через Fan-Out)
```go
// PipelineSplitChannel — распределяет данные по каналам (Round Robin)
func PipelineSplitChannel[T any](inputCh <-chan T, n int) []<-chan T {
    outputChannels := make([]chan T, n)
    for i := 0; i < n; i++ {
        outputChannels[i] = make(chan T)
    }
    
    go func() {
        idx := 0
        for value := range inputCh {
            outputChannels[idx] <- value
            idx = (idx + 1) % n // Round Robin
        }
        for _, ch := range outputChannels {
            close(ch)
        }
    }()
    
    resultChannels := make([]<-chan T, n)
    for i := 0; i < n; i++ {
        resultChannels[i] = outputChannels[i]
    }
    return resultChannels
}

func send(inputCh <-chan string, n int) <-chan string {
    var wg sync.WaitGroup
    wg.Add(n)
    outputCh := make(chan string)
    splitChs := PipelineSplitChannel(inputCh, n) // Каждый воркер получает свой канал
    
    // Каждая горутина читает из своего канала (изоляция)
    for i := 0; i < n; i++ {
        go func(idx int) {
            defer wg.Done()
            for data := range splitChs[idx] {
                outputCh <- fmt.Sprintf("sent: %s by worker %d", data, idx)
            }
        }(i)
    }
    
    go func() {
        wg.Wait()
        close(outputCh)
    }()
    
    return outputCh
}

// Использование: parse → send (каждый воркер изолирован)
logs := generateLogs()
for result := range send(parse(logs), 5) {
    fmt.Println(result)
}
```

**Ключевой момент:** 
- **Пример 5:** Все воркеры конкурируют за данные (проще, но возможна блокировка)
- **Пример 6:** Каждый воркер изолирован через свой канал (сложнее, но предсказуемее)

## 🎯 Преимущества Pipeline

1. **Параллелизм:** стадии работают одновременно
2. **Модульность:** каждая стадия независима
3. **Масштабируемость:** легко добавить/убрать стадии
4. **Читаемость:** декларативный стиль (описываем ЧТО, а не КАК)
5. **Тестируемость:** каждую стадию можно тестировать отдельно
6. **Композиция:** паттерны комбинируются как LEGO-конструктор

## 🧩 Pipeline как LEGO-конструктор

**Философия:** Pipeline — это не просто паттерн, это способ мышления о композиции.

**Базовые блоки:**
- `generate` — источник данных
- `filter` — фильтрация
- `process/decorate` — преобразование
- `fanOut` — распределение
- `fanIn` — сбор
- `tee` — дублирование

**Комбинирование:**
```go
// Простой Pipeline
pipeline := process(generate(1, 2, 3), square)

// Сложный Pipeline (комбинация паттернов)
pipeline := fanIn(
    process(
        filter(
            fanOut(generate(1, 2, 3, 4, 5), 3),
            isEven,
        ),
        heavyProcessing,
    ),
)

// Масштабируемый Pipeline (стадия с параллельным фактором)
pipeline := send(
    parse(
        validate(generateLogs()),
    ),
    5, // 5 воркеров для стадии send
)
```

**Правило:** Каждый паттерн — это функция, которая принимает канал и возвращает канал. Это позволяет бесконечно комбинировать их.

## 🔄 Pipeline vs Worker Pool

| Аспект | Pipeline | Worker Pool |
|--------|----------|-------------|
| **Назначение** | Многостадийная обработка | Параллельная обработка задач |
| **Структура** | Линейная цепочка | Пул воркеров |
| **Данные** | Поток данных | Очередь задач |
| **Когда использовать** | ETL, обработка потоков | Обработка независимых задач |

## 🎯 Round Robin в PipelineSplitChannel

**Round Robin (карусель)** — алгоритм распределения данных по каналам по кругу.

**Как работает:**
```
Данные: [1, 2, 3, 4, 5]
Каналы: [ch0, ch1]

Распределение:
1 → ch0 (idx=0)
2 → ch1 (idx=1)
3 → ch0 (idx=0, снова)
4 → ch1 (idx=1, снова)
5 → ch0 (idx=0, снова)
```

**Реализация:**
```go
idx := 0
for value := range inputCh {
    outputChannels[idx] <- value
    idx = (idx + 1) % n // Переходим к следующему каналу по кругу
}
```

**Почему Round Robin, а не случайное распределение?**
- **Round Robin:** Простой, предсказуемый, равномерный (если данные однородные)
- **Случайное:** Сложнее реализовать, может быть неравномерным
- **По нагрузке:** Требует мониторинг, сложнее реализовать

**Аналогия:** Очередь в банке — клиенты распределяются по окнам по очереди, независимо от того, свободно окно или нет.

**Важно:** Round Robin не проверяет, свободен ли канал. Если один воркер медленный, его канал может переполниться (если не буферизован) или заблокировать распределитель.

---

*"Pipeline — это не просто цепочка функций. Это река, где каждая стадия — порог, который меняет течение воды, но не останавливает поток."* — Мусаси
