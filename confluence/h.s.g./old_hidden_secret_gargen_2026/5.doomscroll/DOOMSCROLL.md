# ЛЕГЕНДА ДЛЯ ПОДГОТОВКИ К СОБЕСЕДОВАНИЮ
## Backend Developer (Golang) - Middle+/Senior

> *"Я слышу — и забываю, я вижу — и запоминаю, я делаю — и понимаю."* — Конфуций

---

## 📋 ОГЛАВЛЕНИЕ

### ФАЗА 1: Базовые концепции и основы (Junior → Middle)
- [Go Language Core](#фаза-1-go-language-core)
- [Базы данных - Основы](#фаза-1-базы-данных)
- [HTTP и REST API](#фаза-1-http-и-rest-api)

### ФАЗА 2: Продвинутые темы и архитектура (Middle → Senior)
- [Брокеры сообщений (Kafka, RabbitMQ)](#фаза-2-брокеры-сообщений)
- [Kubernetes и DevOps](#фаза-2-kubernetes-и-devops)
- [Архитектурные паттерны](#фаза-2-архитектурные-паттерны)
- [Мониторинг и Observability](#фаза-2-мониторинг)

### ФАЗА 3: Экспертные темы и системный дизайн (Senior)
- [Распределенные системы](#фаза-3-распределенные-системы)
- [Алгоритмы и оптимизация](#фаза-3-алгоритмы)
- [Безопасность](#фаза-3-безопасность)
- [Операционные системы и сети](#фаза-3-ос-и-сети)

---

# ФАЗА 1: Базовые концепции и основы

## 🔷 ФАЗА 1: Go Language Core

### 1. Что такое горутины и как они соотносятся с потоками ОС?

**Ответ:**
Горутины — это легковесные потоки выполнения, управляемые runtime Go, которые работают поверх системных потоков операционной системы по принципу M:N (много горутин на несколько OS-потоков). В отличие от тяжеловесных системных потоков, горутины имеют динамический стек размером всего 2 КБ при создании, который может расти и уменьшаться по мере необходимости. Переключение контекста между горутинами происходит гораздо быстрее, чем между OS-потоками, так как выполняется в пространстве пользователя без вызова системных функций ядра.

Планировщик Go использует модель G-M-P, где G (goroutine) — это горутина, M (machine) — системный поток, P (processor) — логический процессор с локальной очередью горутин. Благодаря этой архитектуре один процесс может эффективно управлять тысячами горутин на небольшом количестве системных потоков. В моем проекте платформы генерации подарков я использовал пул из 50+ горутин для параллельной обработки запросов к API генерации изображений, и все они работали на 8 системных потоках (по числу CPU-ядер).

**Пример на Golang:**
```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Горутина %d запущена на потоке OS\n", id)
    time.Sleep(100 * time.Millisecond)
}

func main() {
    // Ограничиваем количество OS-потоков
    runtime.GOMAXPROCS(2)
    
    var wg sync.WaitGroup
    
    // Запускаем 100 горутин на 2 OS-потоках
    for i := 1; i <= 100; i++ {
        wg.Add(1)
        go worker(i, &wg)
    }
    
    wg.Wait()
    fmt.Printf("Всего OS-потоков: %d\n", runtime.GOMAXPROCS(0))
}
```

---

### 2. Чем отличаются каналы (buffered vs unbuffered)?

**Ответ:**
Unbuffered каналы (без буфера) блокируют отправителя до тех пор, пока получатель не заберет данные, обеспечивая синхронную передачу сообщений и гарантируя, что данные переданы в момент отправки. Buffered каналы имеют внутренний буфер заданного размера, позволяя отправителю продолжать работу без блокировки, пока буфер не заполнится полностью. Это создает асинхронную передачу данных и может улучшить производительность, но требует аккуратного управления размером буфера.

Основное различие проявляется в паттернах использования: unbuffered каналы идеальны для синхронизации между горутинами (например, сигнализация о завершении работы), а buffered каналы хороши для балансировки нагрузки между производителями и потребителями с разной скоростью работы. В моем сервисе обработки заказов я использовал buffered канал размером 100 для очереди задач рендеринга изображений, что позволило сглаживать всплески нагрузки и избежать блокировки HTTP-хендлеров.

**Пример на Golang:**
```go
package main

import (
    "fmt"
    "time"
)

func unbufferedExample() {
    ch := make(chan int) // unbuffered канал
    
    go func() {
        fmt.Println("Отправляю 1...")
        ch <- 1 // блокируется здесь до чтения
        fmt.Println("Отправил 1")
    }()
    
    time.Sleep(2 * time.Second)
    fmt.Println("Читаю:", <-ch)
}

func bufferedExample() {
    ch := make(chan int, 3) // buffered канал с буфером на 3 элемента
    
    go func() {
        for i := 1; i <= 3; i++ {
            fmt.Printf("Отправляю %d...\n", i)
            ch <- i // не блокируется, пока буфер не заполнен
            fmt.Printf("Отправил %d (без блокировки)\n", i)
        }
    }()
    
    time.Sleep(2 * time.Second)
    for i := 1; i <= 3; i++ {
        fmt.Println("Читаю:", <-ch)
    }
}

func main() {
    fmt.Println("=== Unbuffered канал ===")
    unbufferedExample()
    
    fmt.Println("\n=== Buffered канал ===")
    bufferedExample()
}
```

---

### 3. Что произойдет при попытке записи в закрытый канал?

**Ответ:**
При попытке записи в закрытый канал произойдет паника (panic), которая приведет к аварийной остановке программы, если не будет обработана через механизм recover. Это одна из наиболее распространенных ошибок при работе с каналами в Go, особенно в сценариях с множественными производителями. Чтение из закрытого канала, напротив, безопасно — оно возвращает нулевое значение типа канала и признак того, что канал закрыт.

Чтобы избежать паники, необходимо гарантировать, что только один компонент (обычно владелец канала или координатор) закрывает канал, и это происходит после того, как все производители завершили свою работу. В моем пайплайне обработки изображений я использовал sync.WaitGroup для отслеживания всех горутин-производителей, и только после wg.Wait() закрывал канал результатов, что полностью исключило возможность записи в закрытый канал.

**Пример на Golang:**
```go
package main

import (
    "fmt"
    "sync"
)

func panicExample() {
    ch := make(chan int, 2)
    
    ch <- 1
    ch <- 2
    close(ch) // закрываем канал
    
    // Попытка записи вызовет панику
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Поймали панику:", r)
        }
    }()
    
    ch <- 3 // PANIC: send on closed channel
}

func safeClosePattern() {
    ch := make(chan int, 10)
    var wg sync.WaitGroup
    
    // Запускаем несколько производителей
    for i := 1; i <= 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 1; j <= 5; j++ {
                ch <- id*10 + j
            }
        }(i)
    }
    
    // Горутина для безопасного закрытия канала
    go func() {
        wg.Wait()  // Ждем завершения всех производителей
        close(ch)  // Только теперь безопасно закрываем
    }()
    
    // Читаем до закрытия канала
    for val := range ch {
        fmt.Println("Получили:", val)
    }
    
    // Чтение из закрытого канала безопасно
    val, ok := <-ch
    fmt.Printf("Закрытый канал: значение=%d, открыт=%v\n", val, ok)
}

func main() {
    fmt.Println("=== Пример паники ===")
    panicExample()
    
    fmt.Println("\n=== Безопасное закрытие ===")
    safeClosePattern()
}
```

---

### 4. Что такое sync.Map и когда его использовать?

**Ответ:**
sync.Map — это потокобезопасная реализация map, встроенная в стандартную библиотеку Go, которая оптимизирована для двух специфических сценариев: когда ключи записываются один раз и многократно читаются (read-heavy workload), и когда множество горутин читают, записывают и удаляют непересекающиеся множества ключей. В отличие от обычной map с обертыванием в sync.RWMutex, sync.Map использует внутренние оптимизации с разделением read и dirty map, что позволяет выполнять большинство операций чтения без блокировок через атомарные операции.

Однако sync.Map не является универсальным решением — для сценариев с частыми записями в одни и те же ключи, обычная map с RWMutex может показать лучшую производительность за счет более простой логики. В моем auth-сервисе я использовал sync.Map для кэширования JWT-токенов пользователей, где типичный паттерн — это запись токена при логине и многократное чтение при валидации запросов, что идеально подходит под оптимизации sync.Map.

**Пример на Golang:**
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Сравнение sync.Map и обычной map с мьютексом
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}

func (sm *SafeMap) Store(key string, value int) {
    sm.mu.Lock()
    sm.m[key] = value
    sm.mu.Unlock()
}

func (sm *SafeMap) Load(key string) (int, bool) {
    sm.mu.RLock()
    val, ok := sm.m[key]
    sm.mu.RUnlock()
    return val, ok
}

func benchmarkSyncMap() {
    var sm sync.Map
    var wg sync.WaitGroup
    
    // Записываем один раз
    for i := 0; i < 100; i++ {
        sm.Store(fmt.Sprintf("key_%d", i), i)
    }
    
    start := time.Now()
    
    // Читаем многократно из множества горутин (read-heavy)
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                sm.Load(fmt.Sprintf("key_%d", j))
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("sync.Map: %v\n", time.Since(start))
}

func benchmarkRegularMap() {
    sm := &SafeMap{m: make(map[string]int)}
    var wg sync.WaitGroup
    
    // Записываем один раз
    for i := 0; i < 100; i++ {
        sm.Store(fmt.Sprintf("key_%d", i), i)
    }
    
    start := time.Now()
    
    // Читаем многократно
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                sm.Load(fmt.Sprintf("key_%d", j))
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("RWMutex Map: %v\n", time.Since(start))
}

func main() {
    fmt.Println("=== Бенчмарк read-heavy сценария ===")
    benchmarkSyncMap()
    benchmarkRegularMap()
    
    // Пример использования в кэше сессий
    var sessionCache sync.Map
    
    // Сохраняем сессию
    sessionCache.Store("user_123", map[string]interface{}{
        "id":       123,
        "username": "alice",
        "expires":  time.Now().Add(24 * time.Hour),
    })
    
    // Читаем сессию
    if val, ok := sessionCache.Load("user_123"); ok {
        session := val.(map[string]interface{})
        fmt.Printf("Сессия: %+v\n", session)
    }
    
    // Удаляем сессию
    sessionCache.Delete("user_123")
}
```

---

### 5. Почему порядок итерации элементов в map не гарантирован?

**Ответ:**
Порядок итерации элементов в map намеренно рандомизирован в Go начиная с версии 1.0, чтобы предотвратить зависимость кода от случайного порядка хранения элементов, который может измениться между версиями runtime или при изменении реализации. Внутренняя структура map основана на хеш-таблице с корзинами (buckets), где элементы распределяются по корзинам на основе хеша ключа, и при итерации Go намеренно начинает с случайной корзины и случайного смещения внутри нее. Это защищает разработчиков от написания кода, который неявно полагается на определенный порядок элементов.

Под капотом map организована как массив корзин, каждая из которых может содержать до 8 пар ключ-значение, а при переполнении создается цепочка overflow-корзин. При росте map происходит эвакуация данных (incremental evacuation) в новый массив корзин большего размера, что также влияет на порядок элементов. В моем сервисе обработки метаданных заказов я столкнулся с багом, когда пытался сравнивать два map напрямую через итерацию, и решил проблему, используя sorted slice ключей для детерминированного обхода.

**Пример на Golang:**
```go
package main

import (
    "fmt"
    "sort"
)

func demonstrateRandomOrder() {
    m := map[string]int{
        "apple":  1,
        "banana": 2,
        "cherry": 3,
        "date":   4,
        "elderberry": 5,
    }
    
    fmt.Println("=== Первая итерация ===")
    for k, v := range m {
        fmt.Printf("%s: %d\n", k, v)
    }
    
    fmt.Println("\n=== Вторая итерация (порядок может отличаться) ===")
    for k, v := range m {
        fmt.Printf("%s: %d\n", k, v)
    }
}

func deterministicIteration() {
    m := map[string]int{
        "apple":  1,
        "banana": 2,
        "cherry": 3,
        "date":   4,
        "elderberry": 5,
    }
    
    // Извлекаем ключи
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    
    // Сортируем ключи
    sort.Strings(keys)
    
    fmt.Println("=== Детерминированная итерация (отсортировано) ===")
    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, m[k])
    }
}

func explainInternalStructure() {
    fmt.Println("\n=== Внутренняя структура map ===")
    fmt.Println("Map в Go реализована как:")
    fmt.Println("1. Массив корзин (buckets), каждая вмещает 8 пар ключ-значение")
    fmt.Println("2. Хеш-функция определяет, в какую корзину попадет элемент")
    fmt.Println("3. При переполнении корзины создается overflow bucket")
    fmt.Println("4. При достижении load factor > 6.5 происходит эвакуация в больший массив")
    fmt.Println("5. Go намеренно начинает итерацию со случайной корзины")
}

func main() {
    demonstrateRandomOrder()
    fmt.Println()
    deterministicIteration()
    explainInternalStructure()
}
```

---

### 6. В чем разница между массивами и слайсами? Как работает append?

**Ответ:**
Массивы (arrays) в Go — это типы значений фиксированного размера, где размер является частью типа ([3]int и [5]int — это разные типы), и при передаче в функцию массив полностью копируется. Слайсы (slices) — это динамические представления над массивом, состоящие из трех компонентов: указателя на underlying array, длины (len) и емкости (capacity). Слайс передается в функцию по значению, но это значение включает указатель, поэтому изменения элементов видны в оригинале, хотя изменения самой структуры слайса (например, через append) не влияют на оригинал.

Функция append работает по следующей логике: если len < cap, элемент добавляется в существующий underlying array и возвращается новый slice header с увеличенным len; если len == cap, выделяется новый массив с емкостью примерно в 2 раза больше (коэффициент роста зависит от размера), данные копируются, и возвращается слайс с новым указателем. В моем парсере конфигурационных файлов я столкнулся с багом, когда передавал слайс в функцию, которая делала append, но не возвращала результат — в итоге при реаллокации изменения терялись, и я исправил это, всегда присваивая результат append обратно.

**Пример на Golang:**
```go
package main

import "fmt"

func demonstrateArrayVsSlice() {
    // Массив - фиксированный размер, тип значения
    arr := [3]int{1, 2, 3}
    fmt.Printf("Массив: %v, тип: %T\n", arr, arr)
    
    modifyArray(arr)
    fmt.Printf("После modifyArray: %v (не изменился)\n", arr)
    
    // Слайс - динамический, reference type
    slice := []int{1, 2, 3}
    fmt.Printf("\nСлайс: %v, len=%d, cap=%d\n", slice, len(slice), cap(slice))
    
    modifySlice(slice)
    fmt.Printf("После modifySlice: %v (элементы изменились)\n", slice)
    
    slice = appendToSlice(slice)
    fmt.Printf("После appendToSlice: %v (слайс расширился)\n", slice)
}

func modifyArray(arr [3]int) {
    arr[0] = 999 // изменяет копию
}

func modifySlice(s []int) {
    s[0] = 999 // изменяет underlying array
}

func appendToSlice(s []int) []int {
    return append(s, 4, 5, 6) // может вызвать реаллокацию
}

func demonstrateAppendMechanism() {
    fmt.Println("\n=== Механизм работы append ===")
    
    s := make([]int, 3, 5) // len=3, cap=5
    fmt.Printf("Исходный: %v, len=%d, cap=%d\n", s, len(s), cap(s))
    
    // Append в рамках capacity - реаллокации нет
    s = append(s, 10)
    fmt.Printf("После append(10): %v, len=%d, cap=%d\n", s, len(s), cap(s))
    
    s = append(s, 20)
    fmt.Printf("После append(20): %v, len=%d, cap=%d\n", s, len(s), cap(s))
    
    // Следующий append превысит capacity - произойдет реаллокация
    s = append(s, 30)
    fmt.Printf("После append(30): %v, len=%d, cap=%d (реаллокация!)\n", s, len(s), cap(s))
}

func demonstrateAppendPitfall() {
    fmt.Println("\n=== Ловушка append ===")
    
    original := []int{1, 2, 3}
    fmt.Printf("Оригинал: %v (cap=%d)\n", original, cap(original))
    
    // Создаем два слайса из одного underlying array
    slice1 := original[:2]  // [1, 2]
    slice2 := original[1:]  // [2, 3]
    
    fmt.Printf("slice1: %v, slice2: %v\n", slice1, slice2)
    
    // Append может перезаписать данные!
    slice1 = append(slice1, 999)
    fmt.Printf("После append в slice1:\n")
    fmt.Printf("  slice1: %v\n", slice1)
    fmt.Printf("  slice2: %v (изменился!)\n", slice2)
    fmt.Printf("  original: %v (изменился!)\n", original)
}

func main() {
    demonstrateArrayVsSlice()
    demonstrateAppendMechanism()
    demonstrateAppendPitfall()
}
```

---

### 7. Что такое context в Go и зачем он нужен?

**Ответ:**
Context — это механизм в Go для передачи дедлайнов, сигналов отмены и request-scoped значений через границы API и между горутинами в рамках одного запроса. Основные функции context — это управление жизненным циклом операций (отмена долгих операций при отмене родительского контекста), установка таймаутов для предотвращения зависших запросов, и передача метаданных запроса (например, request ID, trace ID) без явных параметров. Context передается как первый параметр функции по конвенции и должен проверяться через ctx.Done() в долгих операциях для возможности прерывания.

context.WithTimeout создает контекст, который автоматически отменяется через заданное время, а context.WithDeadline — в конкретный момент времени; разница в том, что WithTimeout удобнее для относительных интервалов, а WithDeadline — для абсолютных меток времени. Важно помнить, что context.CancelFunc нужно обязательно вызывать через defer, чтобы освободить ресурсы и избежать утечек горутин. В моем API-шлюзе я использовал context для каскадной отмены: если клиент разорвал HTTP-соединение, контекст отменялся, и все downstream-запросы к микросервисам тут же прерывались, экономя ресурсы.

**Пример на Golang:**
```go
package main

import (
    "context"
    "fmt"
    "time"
)

// Пример 1: Context с таймаутом
func processWithTimeout(ctx context.Context, data string) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel() // ОБЯЗАТЕЛЬНО вызываем cancel
    
    // Симулируем долгую операцию
    select {
    case <-time.After(3 * time.Second):
        fmt.Println("Обработка завершена")
        return nil
    case <-ctx.Done():
        fmt.Println("Операция отменена:", ctx.Err())
        return ctx.Err()
    }
}

// Пример 2: Каскадная отмена в микросервисной архитектуре
func apiGatewayHandler(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Делаем параллельные запросы к микросервисам
    errCh := make(chan error, 3)
    
    go func() { errCh <- callAuthService(ctx) }()
    go func() { errCh <- callOrderService(ctx) }()
    go func() { errCh <- callGenerationService(ctx) }()
    
    // Если хотя бы один сервис вернул ошибку или таймаут
    for i := 0; i < 3; i++ {
        if err := <-errCh; err != nil {
            cancel() // Отменяем все остальные запросы
            return err
        }
    }
    
    return nil
}

func callAuthService(ctx context.Context) error {
    select {
    case <-time.After(1 * time.Second):
        fmt.Println("Auth service: OK")
        return nil
    case <-ctx.Done():
        fmt.Println("Auth service: cancelled")
        return ctx.Err()
    }
}

func callOrderService(ctx context.Context) error {
    select {
    case <-time.After(2 * time.Second):
        fmt.Println("Order service: OK")
        return nil
    case <-ctx.Done():
        fmt.Println("Order service: cancelled")
        return ctx.Err()
    }
}

func callGenerationService(ctx context.Context) error {
    select {
    case <-time.After(1500 * time.Millisecond):
        fmt.Println("Generation service: OK")
        return nil
    case <-ctx.Done():
        fmt.Println("Generation service: cancelled")
        return ctx.Err()
    }
}

func main() {
    fmt.Println("=== Пример с таймаутом ===")
    ctx := context.Background()
    processWithTimeout(ctx, "важные данные")
    
    fmt.Println("\n=== Каскадная отмена ===")
    apiGatewayHandler(context.Background())
}
```

---

## 🔷 ФАЗА 1: Базы данных - Основы

### 8. Что такое индексы в PostgreSQL и какие типы существуют?

**Ответ:**
Индексы в PostgreSQL — это дополнительные структуры данных, которые хранят упорядоченные копии части данных таблицы для ускорения операций поиска, сортировки и соединения, работая как "указатель" на строки в таблице. Основные типы индексов: B-tree (сбалансированное дерево, используется по умолчанию для операторов сравнения =, <, >, <=, >=), Hash (для точного совпадения =, но редко используется), GiST (обобщенное дерево поиска для геометрических данных, full-text search), GIN (инвертированный индекс для массивов, JSONB, полнотекстового поиска), BRIN (block range index для очень больших таблиц с естественным порядком данных). Индексы ускоряют чтение, но замедляют операции вставки, обновления и удаления, так как требуют поддержания актуальности.

B-tree индекс сбалансирован за счет того, что все листовые узлы находятся на одной глубине от корня, и при вставке/удалении дерево автоматически перебалансируется через операции split и merge, гарантируя время поиска O(log N). В моем сервисе заказов я создал составной B-tree индекс на (user_id, created_at DESC) для быстрого получения последних заказов пользователя; порядок колонок критичен — первая колонка должна быть та, по которой идет фильтрация, а вторая — для сортировки.

**SQL примеры:**
```sql
-- B-tree индекс (по умолчанию)
CREATE INDEX idx_orders_user ON orders(user_id);

-- Составной индекс с порядком сортировки
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);

-- Частичный индекс (только для активных заказов)
CREATE INDEX idx_orders_active ON orders(user_id) 
WHERE status IN ('pending', 'processing');

-- GIN индекс для JSONB
CREATE INDEX idx_orders_metadata ON orders USING GIN(metadata);

-- Функциональный индекс
CREATE INDEX idx_users_email_lower ON users(LOWER(email));

-- Проверка использования индекса
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM orders 
WHERE user_id = 123 
ORDER BY created_at DESC 
LIMIT 10;
```

**Пример на Golang:**
```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

type OrderRepository struct {
    db *sql.DB
}

// Запрос использует индекс idx_orders_user_created
func (r *OrderRepository) GetUserOrdersOptimized(ctx context.Context, userID int, limit int) ([]Order, error) {
    query := `
    SELECT id, user_id, total, status, created_at
    FROM orders
    WHERE user_id = $1
    ORDER BY created_at DESC
    LIMIT $2
    `
    
    rows, err := r.db.QueryContext(ctx, query, userID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var orders []Order
    for rows.Next() {
        var o Order
        if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
            return nil, err
        }
        orders = append(orders, o)
    }
    
    return orders, rows.Err()
}
```

---

### 9. Что такое ACID и уровни изоляции транзакций?

**Ответ:**
ACID — это набор принципов, гарантирующих надежность транзакций в базе данных: Atomicity (атомарность — транзакция выполняется полностью или откатывается полностью), Consistency (консистентность — транзакция переводит БД из одного валидного состояния в другое), Isolation (изоляция — параллельные транзакции не видят промежуточных результатов друг друга), Durability (долговечность — зафиксированные изменения сохраняются даже при сбое системы). Эти принципы обеспечивают предсказуемое поведение в условиях конкурентного доступа и сбоев.

Уровни изоляции в PostgreSQL: Read Uncommitted (практически не используется, ведет себя как Read Committed), Read Committed (по умолчанию — видны только зафиксированные изменения других транзакций), Repeatable Read (повторное чтение в рамках транзакции возвращает те же данные, защита от non-repeatable read), Serializable (наивысший уровень, эмулирует последовательное выполнение транзакций, но может приводить к ошибкам сериализации). Чем выше уровень изоляции, тем больше защита от аномалий чтения, но ниже производительность из-за увеличения блокировок и конфликтов. В моем payment-сервисе я использовал Serializable уровень для критических операций списания баланса, чтобы исключить race condition при одновременных платежах одного пользователя, даже несмотря на возможные retries при serialization failures.

**Пример на Golang:**
```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

type PaymentService struct {
    db *sql.DB
}

// Пример с Read Committed (по умолчанию)
func (s *PaymentService) ProcessPaymentReadCommitted(ctx context.Context, userID int, amount float64) error {
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Проверяем баланс
    var balance float64
    err = tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
    if err != nil {
        return err
    }
    
    if balance < amount {
        return fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance, amount)
    }
    
    // Списываем средства
    _, err = tx.ExecContext(ctx, "UPDATE users SET balance = balance - $1 WHERE id = $2", amount, userID)
    if err != nil {
        return err
    }
    
    // Создаем запись о платеже
    _, err = tx.ExecContext(ctx, 
        "INSERT INTO payments (user_id, amount, status) VALUES ($1, $2, $3)",
        userID, amount, "completed")
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

// Пример с Serializable (максимальная защита)
func (s *PaymentService) ProcessPaymentSerializable(ctx context.Context, userID int, amount float64) error {
    maxRetries := 3
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
            Isolation: sql.LevelSerializable,
        })
        if err != nil {
            return err
        }
        
        // Проверяем баланс с блокировкой строки
        var balance float64
        err = tx.QueryRowContext(ctx, 
            "SELECT balance FROM users WHERE id = $1 FOR UPDATE", 
            userID).Scan(&balance)
        if err != nil {
            tx.Rollback()
            return err
        }
        
        if balance < amount {
            tx.Rollback()
            return fmt.Errorf("insufficient balance")
        }
        
        // Списываем средства
        _, err = tx.ExecContext(ctx, 
            "UPDATE users SET balance = balance - $1 WHERE id = $2", 
            amount, userID)
        if err != nil {
            tx.Rollback()
            return err
        }
        
        // Создаем запись о платеже
        _, err = tx.ExecContext(ctx, 
            "INSERT INTO payments (user_id, amount, status) VALUES ($1, $2, $3)",
            userID, amount, "completed")
        if err != nil {
            tx.Rollback()
            return err
        }
        
        err = tx.Commit()
        if err != nil {
            // Проверяем, является ли ошибка serialization failure
            if isSerializationError(err) && attempt < maxRetries-1 {
                log.Printf("Serialization conflict, retry %d/%d", attempt+1, maxRetries)
                continue
            }
            return err
        }
        
        // Успех!
        return nil
    }
    
    return fmt.Errorf("max retries exceeded")
}

func isSerializationError(err error) bool {
    // PostgreSQL error code 40001 = serialization_failure
    return err != nil && err.Error() == "pq: could not serialize access due to concurrent update"
}

// Демонстрация аномалий чтения
func demonstratePhantomRead(db *sql.DB) {
    ctx := context.Background()
    
    fmt.Println("=== Демонстрация Phantom Read ===")
    
    // Транзакция 1: Read Committed
    tx1, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    
    // Первое чтение
    rows1, _ := tx1.QueryContext(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'pending'")
    var count1 int
    rows1.Next()
    rows1.Scan(&count1)
    rows1.Close()
    fmt.Printf("TX1: Первое чтение - %d pending заказов\n", count1)
    
    // Транзакция 2: добавляет новый pending заказ
    tx2, _ := db.BeginTx(ctx, nil)
    tx2.ExecContext(ctx, "INSERT INTO orders (user_id, status) VALUES (1, 'pending')")
    tx2.Commit()
    fmt.Println("TX2: Добавлен новый pending заказ")
    
    // Второе чтение в TX1
    rows2, _ := tx1.QueryContext(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'pending'")
    var count2 int
    rows2.Next()
    rows2.Scan(&count2)
    rows2.Close()
    fmt.Printf("TX1: Второе чтение - %d pending заказов (Phantom Read!)\n", count2)
    
    tx1.Commit()
}

func main() {
    connStr := "postgres://user:pass@localhost/dbname?sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    service := &PaymentService{db: db}
    
    ctx := context.Background()
    
    // Обработка платежа с Serializable уровнем
    err = service.ProcessPaymentSerializable(ctx, 123, 99.99)
    if err != nil {
        log.Printf("Payment failed: %v", err)
    } else {
        fmt.Println("Payment successful")
    }
}
```

---

### 10. Что такое EXPLAIN ANALYZE и как читать его вывод?

**Ответ:**
EXPLAIN ANALYZE — это команда PostgreSQL, которая не только показывает план выполнения запроса (как EXPLAIN), но и реально выполняет запрос, измеряя фактическое время выполнения каждого узла плана и количество обработанных строк. Ключевые метрики в выводе: cost (оценочная стоимость в условных единицах, где первое число — startup cost, второе — total cost), rows (ожидаемое/фактическое количество строк), width (средний размер строки в байтах), actual time (фактическое время в миллисекундах), loops (количество повторений узла). Разница между оценочными (estimated) и фактическими (actual) значениями указывает на проблемы со статистикой или устаревшими индексами.

Красные флаги в EXPLAIN: Seq Scan на больших таблицах (миллионы строк) вместо Index Scan, большая разница между estimated и actual rows (признак устаревшей статистики — нужен ANALYZE), Nested Loop с большим количеством итераций (может быть признаком отсутствия нужного индекса), высокое значение Buffers (shared hit, read, written — показывает интенсивность работы с памятью и диском). В моем аналитическом сервисе я оптимизировал запрос, который делал Seq Scan по таблице orders (5M+ строк), добавив составной индекс на (status, created_at), что изменило план на Index Scan и снизило время выполнения с 3.5 секунд до 45 мс.

**SQL примеры:**
```sql
-- Базовый EXPLAIN
EXPLAIN 
SELECT * FROM orders 
WHERE user_id = 123 
ORDER BY created_at DESC 
LIMIT 10;

-- EXPLAIN ANALYZE (реально выполняет запрос!)
EXPLAIN (ANALYZE, BUFFERS, TIMING) 
SELECT o.id, o.total, u.email
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.status = 'pending' 
AND o.created_at > NOW() - INTERVAL '7 days'
ORDER BY o.created_at DESC;

-- Пример хорошего вывода (используется индекс):
/*
Index Scan using idx_orders_user_created on orders  
  (cost=0.43..12.48 rows=10 width=64) 
  (actual time=0.025..0.042 rows=10 loops=1)
  Index Cond: (user_id = 123)
Planning Time: 0.123 ms
Execution Time: 0.067 ms
*/

-- Пример плохого вывода (Seq Scan):
/*
Seq Scan on orders  
  (cost=0.00..182537.40 rows=500000 width=64) 
  (actual time=1523.234..3421.567 rows=485231 loops=1)
  Filter: (status = 'pending'::text)
  Rows Removed by Filter: 4514769
Planning Time: 0.234 ms
Execution Time: 3456.789 ms  -- ОЧЕНЬ МЕДЛЕННО!
*/
```

**Пример на Golang:**
```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "strings"
    _ "github.com/lib/pq"
)

type QueryAnalyzer struct {
    db *sql.DB
}

// Анализ производительности запроса
func (qa *QueryAnalyzer) AnalyzeQuery(ctx context.Context, query string, args ...interface{}) {
    explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT) %s", query)
    
    rows, err := qa.db.QueryContext(ctx, explainQuery, args...)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    fmt.Println("=== EXPLAIN ANALYZE ===")
    for rows.Next() {
        var line string
        rows.Scan(&line)
        fmt.Println(line)
        
        // Анализируем красные флаги
        if strings.Contains(line, "Seq Scan") && strings.Contains(line, "actual time") {
            fmt.Println("⚠️  КРАСНЫЙ ФЛАГ: Sequential Scan обнаружен!")
        }
        if strings.Contains(line, "rows=") {
            // Парсим estimated vs actual rows
            // Если разница > 10x, это проблема
        }
    }
}

// Сравнение производительности до и после индекса
func (qa *QueryAnalyzer) ComparePerformance(ctx context.Context, userID int) {
    query := `
    SELECT id, user_id, total, status, created_at
    FROM orders
    WHERE user_id = $1
    ORDER BY created_at DESC
    LIMIT 10
    `
    
    fmt.Println("\n=== ДО создания индекса ===")
    qa.AnalyzeQuery(ctx, query, userID)
    
    // Создаем индекс
    _, err := qa.db.ExecContext(ctx, `
        CREATE INDEX IF NOT EXISTS idx_orders_user_created 
        ON orders(user_id, created_at DESC)
    `)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("\n=== ПОСЛЕ создания индекса ===")
    qa.AnalyzeQuery(ctx, query, userID)
}

// Объяснение типов сканирования
func explainScanTypes() {
    fmt.Println("\n=== Типы сканирования в PostgreSQL ===")
    
    fmt.Println(`
1. Seq Scan (Sequential Scan)
   - Полное сканирование таблицы
   - Используется когда индекса нет или он неэффективен
   - ❌ Плохо для больших таблиц

2. Index Scan
   - Поиск через B-tree индекс
   - Затем чтение heap для получения всех колонок
   - ✅ Хорошо для селективных запросов (< 5% строк)

3. Index Only Scan
   - Все нужные колонки есть в индексе
   - Не нужно обращаться к heap
   - ✅✅ Оптимально!

4. Bitmap Index Scan + Bitmap Heap Scan
   - Сначала строит bitmap из индекса
   - Затем читает heap упорядоченно
   - ✅ Хорошо для средней селективности (5-25% строк)

5. Nested Loop Join
   - Вложенные циклы по двум таблицам
   - ❌ Плохо при большом количестве строк в обеих таблицах

6. Hash Join
   - Строит хеш-таблицу из одной таблицы
   - Проходит по второй, ища совпадения
   - ✅ Хорошо для больших таблиц

7. Merge Join
   - Обе таблицы отсортированы
   - Одновременный проход по обеим
   - ✅ Оптимально для отсортированных данных
    `)
}

func main() {
    connStr := "postgres://user:pass@localhost/dbname?sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    analyzer := &QueryAnalyzer{db: db}
    ctx := context.Background()
    
    // Сравниваем производительность
    analyzer.ComparePerformance(ctx, 123)
    
    // Объясняем типы сканирования
    explainScanTypes()
}
```

---

# ФАЗА 2: Продвинутые темы и архитектура

## 🔷 ФАЗА 2: Брокеры сообщений (Kafka, RabbitMQ)

### 11. Что такое Kafka и RabbitMQ? В чем их принципиальная разница?

**Ответ:**
Apache Kafka — это распределенная платформа потоковой обработки данных, оптимизированная для высокопроизводительной обработки больших объемов событий с возможностью их длительного хранения и перечитывания. RabbitMQ — это традиционный брокер сообщений, реализующий паттерны point-to-point и pub-sub через очереди и exchanges, где сообщения обычно удаляются после подтверждения обработки консьюмером. Основное различие в философии: Kafka — это append-only log с сохранением всех событий на диск, где консьюмеры сами управляют смещением (offset), а RabbitMQ — это умный брокер с очередями, который активно управляет доставкой и подтверждением сообщений.

Когда использовать Kafka: event sourcing, обработка потоков данных в реальном времени, агрегация логов, CDC, сценарии где нужно replay событий или множественные консьюмеры читают одни и те же данные. Когда использовать RabbitMQ: task queues с приоритетами, сложная маршрутизация сообщений через exchanges, сценарии с низкой латентностью и небольшим объемом сообщений. В моем проекте я использовал RabbitMQ для очереди задач рендеринга изображений, так как нагрузка была умеренной (~1000 задач/час), и мне требовалась простая логика retry с dead-letter exchange.

**Сравнительная таблица:**
```
╔═══════════════════╦═════════════════════════════╦═════════════════════════════╗
║ Характеристика    ║ Kafka                       ║ RabbitMQ                    ║
╠═══════════════════╬═════════════════════════════╬═════════════════════════════╣
║ Модель            ║ Распределенный лог          ║ Очереди + exchanges         ║
║ Хранение          ║ Долгое (дни/недели)         ║ До подтверждения            ║
║ Throughput        ║ Миллионы msg/сек            ║ Десятки тысяч msg/сек       ║
║ Latency           ║ Средняя (ms)                ║ Низкая (µs)                 ║
║ Гарантии          ║ At least once/Exactly once  ║ At least once               ║
║ Порядок           ║ В рамках партиции           ║ В рамках очереди            ║
║ Replay            ║ ✅ Да (через offset)        ║ ❌ Нет                     ║
║ Маршрутизация     ║ Простая (topic-based)       ║ Сложная (exchanges)         ║
║ Use case          ║ Event streaming, logs       ║ Task queues, RPC            ║
╚═══════════════════╩═════════════════════════════╩═════════════════════════════╝
```

---

### 12. Что такое партиции в Kafka? Как они влияют на производительность?

**Ответ:**
Партиции (partitions) в Kafka — это упорядоченные, неизменяемые последовательности сообщений, на которые разбивается топик для параллельной обработки и горизонтального масштабирования. Каждая партиция хранится на отдельном broker'е, и сообщения внутри партиции строго упорядочены по offset, но порядок между партициями не гарантируется. Ключ сообщения (message key) определяет, в какую партицию попадет сообщение через хеш-функцию, что позволяет гарантировать порядок обработки для сообщений с одинаковым ключом.

Количество партиций влияет на производительность: больше партиций = больше параллелизма для консьюмеров, быстрее throughput для продюсеров, но выше накладные расходы на репликацию. Важно: нельзя уменьшить количество партиций после создания топика, и количество консьюмеров в группе не должно превышать количество партиций, иначе лишние консьюмеры будут простаивать. В моем сервисе событий я создал топик "user-activity" с 12 партициями и настроил партиционирование по user_id, чтобы все события одного пользователя обрабатывались последовательно одним консьюмером.

**Пример на Golang:**
```go
package main

import (
    "context"
    "fmt"
    "hash/fnv"
    "github.com/segmentio/kafka-go"
)

// Кастомный партиционер по user_id
type UserIDPartitioner struct{}

func (p *UserIDPartitioner) Balance(messages []kafka.Message, partitions ...int) {
    for i := range messages {
        userID := string(messages[i].Key)
        partition := hashUserID(userID) % len(partitions)
        messages[i].Partition = partitions[partition]
    }
}

func hashUserID(userID string) int {
    h := fnv.New32a()
    h.Write([]byte(userID))
    return int(h.Sum32())
}

// Producer с партиционированием
type PartitionedProducer struct {
    writer *kafka.Writer
}

func NewPartitionedProducer(brokers []string, topic string) *PartitionedProducer {
    return &PartitionedProducer{
        writer: kafka.NewWriter(kafka.WriterConfig{
            Brokers:  brokers,
            Topic:    topic,
            Balancer: &UserIDPartitioner{}, // кастомный партиционер
        }),
    }
}

func (pp *PartitionedProducer) PublishUserEvent(ctx context.Context, userID, eventType, data string) error {
    msg := kafka.Message{
        Key:   []byte(userID), // ВАЖНО: ключ определяет партицию
        Value: []byte(fmt.Sprintf("%s:%s", eventType, data)),
    }
    
    return pp.writer.WriteMessages(ctx, msg)
}

func main() {
    ctx := context.Background()
    producer := NewPartitionedProducer([]string{"localhost:9092"}, "user-events")
    
    // Все события user-1 попадут в одну партицию
    producer.PublishUserEvent(ctx, "user-1", "click", "button-A")
    producer.PublishUserEvent(ctx, "user-1", "click", "button-B")
    
    // События user-2 могут быть в другой партиции
    producer.PublishUserEvent(ctx, "user-2", "page-view", "home")
    
    fmt.Println("Порядок событий user-1 гарантирован внутри его партиции")
}
```

---

### 13. Что такое Transactional Outbox Pattern?

**Ответ:**
Transactional Outbox Pattern — это архитектурный паттерн, гарантирующий атомарность между изменением состояния в базе данных и публикацией события в брокер сообщений (Kafka/RabbitMQ) через использование единой транзакции. Суть паттерна: при изменении бизнес-данных в БД в рамках той же транзакции записывается событие в специальную таблицу outbox, затем асинхронный процесс-реле (Message Relay) или CDC-инструмент (например, Debezium) забирает события из outbox и публикует их в брокер. Это решает проблему dual-write, когда сервис падает после commit в БД, но до отправки сообщения, или наоборот — отправил сообщение, но транзакция откатилась.

Два подхода реализации: polling-based (реле периодически опрашивает таблицу outbox через SELECT FOR UPDATE SKIP LOCKED и публикует новые события) и log-based (Debezium читает WAL базы данных и публикует изменения в таблице outbox как события в Kafka). Преимущества: строгие гарантии консистентности, события никогда не потеряются, естественная защита от дублей через идемпотентность консьюмеров. В моем payment-сервисе я реализовал Outbox Pattern для публикации события "PaymentCompleted": при успешном списании баланса в транзакции записывалось событие в outbox, а отдельная горутина каждые 100ms публиковала новые события в Kafka, гарантируя, что ни одно подтвержденное списание не потеряется.

**Пример на Golang:**
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"
    _ "github.com/lib/pq"
)

// Схема таблицы outbox
const outboxSchema = `
CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished 
ON outbox(created_at) WHERE published_at IS NULL;
`

type OutboxEvent struct {
    ID            int64
    AggregateType string
    AggregateID   string
    EventType     string
    Payload       json.RawMessage
    CreatedAt     time.Time
    PublishedAt   *time.Time
}

type PaymentService struct {
    db *sql.DB
}

// Бизнес-операция с Outbox Pattern
func (s *PaymentService) ProcessPayment(ctx context.Context, userID int, amount float64) error {
    // Начинаем транзакцию
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 1. Списываем средства
    result, err := tx.ExecContext(ctx,
        "UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, userID)
    if err != nil {
        return err
    }
    
    affected, _ := result.RowsAffected()
    if affected == 0 {
        return fmt.Errorf("insufficient balance")
    }
    
    // 2. Записываем событие в outbox (в той же транзакции!)
    eventPayload, _ := json.Marshal(map[string]interface{}{
        "user_id": userID,
        "amount":  amount,
        "timestamp": time.Now(),
    })
    
    _, err = tx.ExecContext(ctx, `
        INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
        VALUES ($1, $2, $3, $4)
    `, "Payment", fmt.Sprintf("%d", userID), "PaymentCompleted", eventPayload)
    if err != nil {
        return err
    }
    
    // 3. Коммитим обе операции атомарно
    return tx.Commit()
}

// Message Relay: публикует события из outbox в Kafka
type MessageRelay struct {
    db             *sql.DB
    kafkaProducer  KafkaProducer
    pollInterval   time.Duration
}

func NewMessageRelay(db *sql.DB, kafkaProducer KafkaProducer) *MessageRelay {
    return &MessageRelay{
        db:            db,
        kafkaProducer: kafkaProducer,
        pollInterval:  100 * time.Millisecond,
    }
}

func (mr *MessageRelay) Start(ctx context.Context) {
    ticker := time.NewTicker(mr.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mr.processOutboxBatch(ctx)
        }
    }
}

func (mr *MessageRelay) processOutboxBatch(ctx context.Context) {
    // Используем SELECT FOR UPDATE SKIP LOCKED для конкурентной обработки
    query := `
        SELECT id, aggregate_type, aggregate_id, event_type, payload
        FROM outbox
        WHERE published_at IS NULL
        ORDER BY created_at
        LIMIT 100
        FOR UPDATE SKIP LOCKED
    `
    
    tx, err := mr.db.BeginTx(ctx, nil)
    if err != nil {
        log.Printf("Failed to begin tx: %v", err)
        return
    }
    defer tx.Rollback()
    
    rows, err := tx.QueryContext(ctx, query)
    if err != nil {
        log.Printf("Failed to query outbox: %v", err)
        return
    }
    defer rows.Close()
    
    var eventIDs []int64
    
    for rows.Next() {
        var event OutboxEvent
        err := rows.Scan(&event.ID, &event.AggregateType, &event.AggregateID, 
                        &event.EventType, &event.Payload)
        if err != nil {
            log.Printf("Failed to scan row: %v", err)
            continue
        }
        
        // Публикуем в Kafka
        err = mr.kafkaProducer.Publish(ctx, event.EventType, event.Payload)
        if err != nil {
            log.Printf("Failed to publish event %d: %v", event.ID, err)
            continue
        }
        
        eventIDs = append(eventIDs, event.ID)
    }
    
    if len(eventIDs) == 0 {
        return
    }
    
    // Помечаем события как опубликованные
    _, err = tx.ExecContext(ctx, `
        UPDATE outbox 
        SET published_at = NOW() 
        WHERE id = ANY($1)
    `, eventIDs)
    if err != nil {
        log.Printf("Failed to mark events as published: %v", err)
        return
    }
    
    tx.Commit()
    fmt.Printf("Published %d events from outbox\n", len(eventIDs))
}

// Заглушка Kafka Producer
type KafkaProducer interface {
    Publish(ctx context.Context, eventType string, payload []byte) error
}

type SimpleKafkaProducer struct{}

func (p *SimpleKafkaProducer) Publish(ctx context.Context, eventType string, payload []byte) error {
    fmt.Printf("Publishing to Kafka: %s -> %s\n", eventType, string(payload))
    return nil
}

// Архитектурная диаграмма
func explainOutboxPattern() {
    fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║ TRANSACTIONAL OUTBOX PATTERN                                      ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  ┌─────────────┐                                                  ║
║  │   Service   │                                                  ║
║  └──────┬──────┘                                                  ║
║         │                                                         ║
║         │ BEGIN TRANSACTION                                       ║
║         ├──────────────────────────────────────┐                  ║
║         │                                      │                  ║
║         ▼                                      ▼                  ║
║  ┌─────────────┐                        ┌────────────┐            ║
║  │ Business    │                        │  Outbox    │            ║
║  │ Table       │                        │  Table     │            ║
║  │ (UPDATE)    │                        │  (INSERT)  │            ║
║  └─────────────┘                        └────────────┘            ║
║         │                                      │                  ║
║         └──────────────────────────────────────┘                  ║
║                    │                                              ║
║                    │ COMMIT                                       ║
║                    ▼                                              ║
║              ✅ Атомарность                                       ║
║                                                                   ║
║  ┌─────────────────────────────────────────────────┐              ║
║  │         Message Relay (async)                   │              ║
║  │  1. SELECT FROM outbox WHERE published_at NULL  │              ║
║  │  2. Publish to Kafka                            │              ║
║  │  3. UPDATE outbox SET published_at = NOW()      │              ║
║  └─────────────────────────────────────────────────┘              ║
║                    │                                              ║
║                    ▼                                              ║
║             ┌──────────────┐                                      ║
║             │    Kafka     │                                      ║
║             └──────────────┘                                      ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

ПРЕИМУЩЕСТВА:
✅ Гарантия доставки события при commit БД
✅ Не нужна распределенная транзакция (2PC)
✅ Простая реализация

НЕДОСТАТКИ:
❌ Дополнительная таблица в БД
❌ Задержка публикации (polling interval)
❌ Нужен отдельный процесс-реле
    `)
}

func main() {
    explainOutboxPattern()
    
    // Подключаемся к БД
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Создаем схему
    db.Exec(outboxSchema)
    
    // Инициализируем сервисы
    paymentService := &PaymentService{db: db}
    kafkaProducer := &SimpleKafkaProducer{}
    relay := NewMessageRelay(db, kafkaProducer)
    
    ctx := context.Background()
    
    // Запускаем relay в фоне
    go relay.Start(ctx)
    
    // Обрабатываем платеж
    err = paymentService.ProcessPayment(ctx, 123, 99.99)
    if err != nil {
        log.Printf("Payment failed: %v", err)
    } else {
        fmt.Println("Payment successful, event will be published by relay")
    }
    
    // Даем время relay'у обработать событие
    time.Sleep(200 * time.Millisecond)
}
```

---

## 🔷 ФАЗА 2: Kubernetes и DevOps

### 14. Что такое под (Pod) в Kubernetes? Что такое liveness, readiness и startup пробы?

**Ответ:**
Pod — это минимальная развертываемая единица в Kubernetes, представляющая собой группу из одного или нескольких контейнеров, которые совместно используют сетевое пространство (один IP-адрес), хранилище (volumes) и контекст выполнения. Поды — это эфемерные сущности, которые могут быть пересозданы в любой момент при сбоях или обновлениях, поэтому состояние приложения не должно храниться внутри пода без использования Persistent Volumes. Kubernetes управляет подами через контроллеры (Deployment, StatefulSet, DaemonSet), которые обеспечивают желаемое количество реплик и политики обновления.

Пробы (probes) — это механизм проверки здоровья контейнеров: **Liveness Probe** проверяет, жив ли контейнер, и при неудачах kubelet перезапускает контейнер; **Readiness Probe** определяет, готов ли контейнер принимать трафик, и при неудачах под исключается из endpoints Service; **Startup Probe** используется для медленно стартующих приложений, отключая liveness/readiness пробы до первого успешного старта. В моем Go-сервисе API я настроил readiness probe на `/health/ready` (проверяет подключение к БД и Kafka), liveness на `/health/alive` (простая проверка HTTP 200), и startup probe с initialDelaySeconds=30 для прогрева кэшей при старте.

**Kubernetes YAML с пробами:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: api-service:v1.0.0
        ports:
        - containerPort: 8080
        
        # Startup Probe: для медленного старта
        startupProbe:
          httpGet:
            path: /health/startup
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          failureThreshold: 12  # 60 секунд на старт
        
        # Liveness Probe: перезапуск если зависло
        livenessProbe:
          httpGet:
            path: /health/alive
            port: 8080
          periodSeconds: 10
          failureThreshold: 3
        
        # Readiness Probe: убирает из Service если не готов
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          periodSeconds: 5
          failureThreshold: 2
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## 🔷 ФАЗА 2: Архитектурные паттерны

### 15. Что такое CQRS (Command Query Responsibility Segregation)?

**Ответ:**
CQRS (Command Query Responsibility Segregation) — это архитектурный паттерн, разделяющий модель данных на две части: команды (commands) для изменения состояния системы и запросы (queries) для чтения данных, каждая со своей оптимизированной моделью данных и, часто, отдельной БД. Команды выполняют бизнес-логику, валидацию и обновляют write-модель (обычно нормализованную реляционную БД), после чего публикуют события, которые асинхронно обновляют read-модель (денормализованную БД, оптимизированную для быстрого чтения — MongoDB, Elasticsearch, Redis). Это позволяет независимо масштабировать операции чтения и записи, упрощает сложные запросы через пред-агрегированные представления, и естественно интегрируется с Event Sourcing.

Преимущества CQRS: независимое масштабирование (read-реплики могут быть на отдельных серверах), оптимизация для разных паттернов доступа (write-модель для консистентности, read-модель для производительности), упрощение сложной бизнес-логики через разделение ответственности. Недостатки: eventual consistency между моделями (пользователь может не сразу увидеть свои изменения), увеличенная сложность инфраструктуры (две БД, event bus), необходимость идемпотентности обработчиков событий. В моем e-commerce проекте я использовал CQRS для каталога товаров: команды (CreateProduct, UpdatePrice) изменяли PostgreSQL с строгой консистентностью, а события обновляли Elasticsearch для быстрого полнотекстового поиска и фасетной фильтрации, что дало 60% ускорение запросов поиска при сохранении ACID-гарантий для изменений.

**Пример на Golang:**
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

// ========== COMMAND SIDE (Write Model) ==========

type Product struct {
    ID          string
    Name        string
    Price       float64
    Stock       int
    Version     int       // Оптимистичная блокировка
    UpdatedAt   time.Time
}

type CreateProductCommand struct {
    ID    string
    Name  string
    Price float64
    Stock int
}

type UpdatePriceCommand struct {
    ID       string
    NewPrice float64
    Version  int
}

type CommandHandler struct {
    db        *sql.DB
    eventBus  EventBus
}

func (h *CommandHandler) HandleCreateProduct(ctx context.Context, cmd CreateProductCommand) error {
    // Валидация
    if cmd.Price <= 0 {
        return fmt.Errorf("invalid price")
    }
    
    // Начинаем транзакцию
    tx, _ := h.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // Сохраняем в write-модель (PostgreSQL)
    _, err := tx.ExecContext(ctx, `
        INSERT INTO products (id, name, price, stock, version, updated_at)
        VALUES ($1, $2, $3, $4, 1, NOW())
    `, cmd.ID, cmd.Name, cmd.Price, cmd.Stock)
    if err != nil {
        return err
    }
    
    // Публикуем событие через Outbox Pattern
    event := ProductCreatedEvent{
        ProductID: cmd.ID,
        Name:      cmd.Name,
        Price:     cmd.Price,
        Stock:     cmd.Stock,
        Timestamp: time.Now(),
    }
    
    eventData, _ := json.Marshal(event)
    _, err = tx.ExecContext(ctx, `
        INSERT INTO outbox (event_type, aggregate_id, payload)
        VALUES ($1, $2, $3)
    `, "ProductCreated", cmd.ID, eventData)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

func (h *CommandHandler) HandleUpdatePrice(ctx context.Context, cmd UpdatePriceCommand) error {
    tx, _ := h.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // Оптимистичная блокировка через version
    result, err := tx.ExecContext(ctx, `
        UPDATE products 
        SET price = $1, version = version + 1, updated_at = NOW()
        WHERE id = $2 AND version = $3
    `, cmd.NewPrice, cmd.ID, cmd.Version)
    if err != nil {
        return err
    }
    
    affected, _ := result.RowsAffected()
    if affected == 0 {
        return fmt.Errorf("concurrent modification detected")
    }
    
    // Публикуем событие
    event := PriceUpdatedEvent{
        ProductID: cmd.ID,
        NewPrice:  cmd.NewPrice,
        Timestamp: time.Now(),
    }
    
    eventData, _ := json.Marshal(event)
    tx.ExecContext(ctx, `
        INSERT INTO outbox (event_type, aggregate_id, payload)
        VALUES ($1, $2, $3)
    `, "PriceUpdated", cmd.ID, eventData)
    
    return tx.Commit()
}

// ========== QUERY SIDE (Read Model) ==========

type ProductSearchResult struct {
    ID          string
    Name        string
    Price       float64
    Stock       int
    Description string
    Category    string
    Rating      float64
    UpdatedAt   time.Time
}

type QueryHandler struct {
    readDB ReadDatabase // Elasticsearch/MongoDB/Redis
}

func (h *QueryHandler) SearchProducts(ctx context.Context, query string, filters map[string]interface{}) ([]ProductSearchResult, error) {
    // Быстрый поиск по денормализованной read-модели
    // Здесь могут быть пред-агрегированные данные, full-text search, фасеты
    return h.readDB.Search(ctx, query, filters)
}

func (h *QueryHandler) GetProductDetails(ctx context.Context, productID string) (*ProductSearchResult, error) {
    // Детали продукта из read-модели (может включать агрегаты)
    return h.readDB.GetByID(ctx, productID)
}

// ========== EVENT HANDLERS (Синхронизация моделей) ==========

type ProductCreatedEvent struct {
    ProductID string
    Name      string
    Price     float64
    Stock     int
    Timestamp time.Time
}

type PriceUpdatedEvent struct {
    ProductID string
    NewPrice  float64
    Timestamp time.Time
}

type ReadModelUpdater struct {
    readDB ReadDatabase
}

func (u *ReadModelUpdater) HandleProductCreated(ctx context.Context, event ProductCreatedEvent) error {
    // Обновляем read-модель (Elasticsearch)
    doc := ProductSearchResult{
        ID:        event.ProductID,
        Name:      event.Name,
        Price:     event.Price,
        Stock:     event.Stock,
        UpdatedAt: event.Timestamp,
    }
    
    return u.readDB.Index(ctx, event.ProductID, doc)
}

func (u *ReadModelUpdater) HandlePriceUpdated(ctx context.Context, event PriceUpdatedEvent) error {
    // Частичное обновление в read-модели
    return u.readDB.UpdateField(ctx, event.ProductID, "price", event.Price)
}

// ========== ИНТЕРФЕЙСЫ ==========

type EventBus interface {
    Publish(ctx context.Context, event interface{}) error
}

type ReadDatabase interface {
    Search(ctx context.Context, query string, filters map[string]interface{}) ([]ProductSearchResult, error)
    GetByID(ctx context.Context, id string) (*ProductSearchResult, error)
    Index(ctx context.Context, id string, doc interface{}) error
    UpdateField(ctx context.Context, id string, field string, value interface{}) error
}

// ========== АРХИТЕКТУРНАЯ ДИАГРАММА ==========

func explainCQRS() {
    fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║ CQRS PATTERN                                                      ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  CLIENT                                                           ║
║    │                                                              ║
║    ├──── Commands ────┐                    ┌──── Queries ────┐    ║
║    │                  ▼                    │                 ▼    ║
║    │         ┌──────────────────┐          │    ┌───────────────┐ ║
║    │         │ Command Handler  │          │    │Query Handler  │ ║
║    │         └─────────┬────────┘          │    └───────┬───────┘ ║
║    │                   │                   │            │         ║
║    │                   ▼                   │            ▼         ║
║    │         ┌──────────────────┐          │    ┌───────────────┐ ║
║    │         │  Write Model     │          │    │  Read Model   │ ║
║    │         │  (PostgreSQL)    │          │    │(Elasticsearch)│ ║
║    │         │  Normalized      │          │    │ Denormalized  │ ║
║    │         └─────────┬────────┘          │    └───────────────┘ ║
║    │                   │                   │                      ║
║    │                   │ Публикует события │                      ║
║    │                   └────────┬──────────┘                      ║
║    │                            │                                 ║
║    │                            ▼                                 ║
║    │                   ┌─────────────────┐                        ║
║    │                   │   Event Bus     │                        ║
║    │                   │   (Kafka)       │                        ║
║    │                   └────────┬────────┘                        ║
║    │                            │                                 ║
║    │                            │ Подписка                        ║
║    │                            ▼                                 ║
║    │                   ┌──────────────────┐                       ║
║    │                   │ Event Handlers   │                       ║
║    │                   │ (Read Updaters)  │                       ║
║    │                   └────────┬─────────┘                       ║
║    │                            │                                 ║
║    │                            │ Обновляет                       ║
║    └────────────────────────────┼────────────────────────────────┘
║                                 │                                 ║
║                                 ▼                                 ║
║                      Read Model обновлена                         ║
║                      (Eventually Consistent)                      ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

ПРЕИМУЩЕСТВА:
✅ Независимое масштабирование read/write
✅ Оптимизация для разных паттернов доступа
✅ Упрощение сложных запросов

НЕДОСТАТКИ:
❌ Eventual consistency
❌ Увеличенная сложность
❌ Необходимость идемпотентности
    `)
}

func main() {
    explainCQRS()
}
```

---

## 🔷 ФАЗА 2: gRPC и HTTP/REST

### 16. В чем разница между gRPC и REST API? Когда использовать каждый?

**Ответ:**
gRPC — это высокопроизводительный RPC-фреймворк от Google, использующий Protocol Buffers для сериализации данных и HTTP/2 для транспорта, что обеспечивает бинарную передачу, мультиплексирование запросов по одному TCP-соединению, и нативную поддержку стриминга (server, client, bidirectional). REST API — это архитектурный стиль на основе HTTP/1.1 с текстовыми форматами (JSON/XML), где ресурсы адресуются через URL, а действия определяются HTTP-методами (GET, POST, PUT, DELETE). Основные отличия: gRPC быстрее за счет бинарного protobuf (~7-10x меньше размер по сравнению с JSON) и HTTP/2 multiplexing, но требует кодогенерации и плохо работает в браузерах без grpc-web proxy.

Когда использовать gRPC: межсервисная коммуникация в микросервисах (низкая латентность критична), real-time стриминг данных (логи, метрики), polyglot-системы (protobuf поддерживает автогенерацию клиентов для всех языков), mobile-приложения с плохим интернетом (меньше трафика). Когда использовать REST: публичные API для сторонних разработчиков (простота интеграции), веб-приложения в браузере (без дополнительных прокси), простые CRUD-операции, где важна человеко-читаемость запросов. В моем проекте я использовал gRPC для внутренней коммуникации между микросервисами (auth ↔ order ↔ generation), где latency p99 снизилась с 45ms (REST) до 12ms (gRPC), и REST API для фронтенда, так как gRPC требовал бы настройки Envoy proxy для работы в браузере.

**Пример на Golang:**

```protobuf
// order.proto
syntax = "proto3";
package order;

option go_package = "github.com/myproject/api/order";

service OrderService {
  // Unary RPC: простой запрос-ответ
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  
  // Server streaming: сервер отправляет поток данных
  rpc StreamOrderUpdates(StreamOrderRequest) returns (stream OrderUpdate);
  
  // Client streaming: клиент отправляет поток данных
  rpc UploadOrderItems(stream OrderItem) returns (UploadResponse);
  
  // Bidirectional streaming: двунаправленный поток
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}

message CreateOrderRequest {
  int64 user_id = 1;
  repeated OrderItem items = 2;
  string delivery_address = 3;
}

message CreateOrderResponse {
  int64 order_id = 1;
  string status = 2;
  double total_amount = 3;
}

message StreamOrderRequest {
  int64 order_id = 1;
}

message OrderUpdate {
  int64 order_id = 1;
  string status = 2;
  string timestamp = 3;
}

message OrderItem {
  int64 product_id = 1;
  int32 quantity = 2;
}

message UploadResponse {
  int32 items_count = 1;
  string message = 2;
}

message ChatMessage {
  string user_id = 1;
  string message = 2;
  string timestamp = 3;
}
```

**Golang сервер:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    pb "github.com/myproject/api/order"
)

type orderServer struct {
    pb.UnimplementedOrderServiceServer
}

// Unary RPC: простой запрос-ответ
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // Валидация
    if req.UserId == 0 {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }
    
    if len(req.Items) == 0 {
        return nil, status.Error(codes.InvalidArgument, "items cannot be empty")
    }
    
    // Бизнес-логика создания заказа
    orderID := int64(12345) // сгенерированный ID
    totalAmount := calculateTotal(req.Items)
    
    log.Printf("Created order %d for user %d with %d items", orderID, req.UserId, len(req.Items))
    
    return &pb.CreateOrderResponse{
        OrderId:     orderID,
        Status:      "created",
        TotalAmount: totalAmount,
    }, nil
}

// Server streaming: отправляем обновления заказа клиенту
func (s *orderServer) StreamOrderUpdates(req *pb.StreamOrderRequest, stream pb.OrderService_StreamOrderUpdatesServer) error {
    orderID := req.OrderId
    
    // Симулируем обновления статуса заказа
    statuses := []string{"created", "paid", "processing", "shipped", "delivered"}
    
    for _, status := range statuses {
        update := &pb.OrderUpdate{
            OrderId:   orderID,
            Status:    status,
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        if err := stream.Send(update); err != nil {
            return err
        }
        
        log.Printf("Sent update: order %d -> %s", orderID, status)
        time.Sleep(2 * time.Second)
    }
    
    return nil
}

// Client streaming: принимаем поток товаров от клиента
func (s *orderServer) UploadOrderItems(stream pb.OrderService_UploadOrderItemsServer) error {
    var itemCount int32
    
    for {
        item, err := stream.Recv()
        if err == io.EOF {
            // Клиент закончил отправку
            return stream.SendAndClose(&pb.UploadResponse{
                ItemsCount: itemCount,
                Message:    fmt.Sprintf("Received %d items", itemCount),
            })
        }
        if err != nil {
            return err
        }
        
        itemCount++
        log.Printf("Received item: product_id=%d, quantity=%d", item.ProductId, item.Quantity)
    }
}

// Bidirectional streaming: чат
func (s *orderServer) Chat(stream pb.OrderService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        
        log.Printf("Received chat message from %s: %s", msg.UserId, msg.Message)
        
        // Отправляем ответ
        response := &pb.ChatMessage{
            UserId:    "bot",
            Message:   fmt.Sprintf("Echo: %s", msg.Message),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}

func calculateTotal(items []*pb.OrderItem) float64 {
    // Заглушка
    return 999.99
}

func main() {
    // Слушаем на порту 50051
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    // Создаем gRPC сервер
    grpcServer := grpc.NewServer(
        grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
        grpc.MaxSendMsgSize(10 * 1024 * 1024),
    )
    
    // Регистрируем сервис
    pb.RegisterOrderServiceServer(grpcServer, &orderServer{})
    
    log.Println("gRPC server listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

**Golang клиент:**

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/myproject/api/order"
)

func main() {
    // Подключаемся к серверу
    conn, err := grpc.Dial("localhost:50051", 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()
    
    client := pb.NewOrderServiceClient(conn)
    ctx := context.Background()
    
    // 1. Unary RPC
    fmt.Println("=== Unary RPC ===")
    createResp, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
        UserId: 123,
        Items: []*pb.OrderItem{
            {ProductId: 1, Quantity: 2},
            {ProductId: 2, Quantity: 1},
        },
        DeliveryAddress: "Moscow, Red Square, 1",
    })
    if err != nil {
        log.Fatalf("CreateOrder failed: %v", err)
    }
    fmt.Printf("Order created: ID=%d, Status=%s, Total=%.2f\n", 
        createResp.OrderId, createResp.Status, createResp.TotalAmount)
    
    // 2. Server streaming
    fmt.Println("\n=== Server Streaming ===")
    stream, err := client.StreamOrderUpdates(ctx, &pb.StreamOrderRequest{
        OrderId: createResp.OrderId,
    })
    if err != nil {
        log.Fatalf("StreamOrderUpdates failed: %v", err)
    }
    
    for {
        update, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("stream error: %v", err)
        }
        fmt.Printf("Order update: %s -> %s\n", update.Status, update.Timestamp)
    }
    
    // 3. Client streaming
    fmt.Println("\n=== Client Streaming ===")
    uploadStream, err := client.UploadOrderItems(ctx)
    if err != nil {
        log.Fatalf("UploadOrderItems failed: %v", err)
    }
    
    items := []*pb.OrderItem{
        {ProductId: 10, Quantity: 5},
        {ProductId: 20, Quantity: 3},
        {ProductId: 30, Quantity: 7},
    }
    
    for _, item := range items {
        if err := uploadStream.Send(item); err != nil {
            log.Fatalf("upload error: %v", err)
        }
        fmt.Printf("Uploaded item: product_id=%d\n", item.ProductId)
    }
    
    uploadResp, err := uploadStream.CloseAndRecv()
    if err != nil {
        log.Fatalf("CloseAndRecv error: %v", err)
    }
    fmt.Printf("Upload complete: %s\n", uploadResp.Message)
}
```

**Сравнительная таблица:**

```
╔═══════════════════╦═══════════════════════════╦═══════════════════════════╗
║ Характеристика    ║ gRPC                      ║ REST API                  ║
╠═══════════════════╬═══════════════════════════╬═══════════════════════════╣
║ Транспорт         ║ HTTP/2                    ║ HTTP/1.1                  ║
║ Сериализация      ║ Protobuf (binary)         ║ JSON (text)               ║
║ Производительность║ Высокая (7-10x быстрее)   ║ Средняя                   ║
║ Размер payload    ║ Маленький (~70% меньше)    ║ Большой                   ║
║ Стриминг          ║ ✅ Native (bi-directional) ║ ❌ SSE/WebSocket отдельно  ║
║ Браузер           ║ ❌ Нужен grpc-web proxy    ║ ✅ Нативная поддержка      ║
║ Кодогенерация     ║ ✅ Обязательна             ║ ❌ Опционально             ║
║ Human-readable    ║ ❌ Бинарный формат         ║ ✅ JSON читабелен          ║
║ Кэширование       ║ ❌ Сложнее (HTTP/2 POST)   ║ ✅ HTTP-кэш работает       ║
║ Latency           ║ Низкая (2-5ms)             ║ Средняя (10-50ms)         ║
║ Use case          ║ Микросервисы, стриминг     ║ Публичные API, веб        ║
╚═══════════════════╩════════════════════════════╩═══════════════════════════╝
```

---

## 🔷 ФАЗА 2: Мониторинг и Observability

### 17. Что такое Prometheus и как собирать кастомные метрики в Go?

**Ответ:**
Prometheus — это open-source система мониторинга и алертинга, разработанная в SoundCloud, которая использует pull-модель (scraping) для сбора метрик через HTTP-endpoints и хранит time-series данные в локальной TSDB с мощным языком запросов PromQL. Prometheus периодически опрашивает настроенные targets (обычно `/metrics` endpoint), где приложения экспонируют метрики в текстовом формате, и сохраняет их с временными метками. Основные типы метрик: Counter (монотонно растущее значение, например, количество запросов), Gauge (изменяющееся значение, например, текущее количество горутин), Histogram (распределение значений по бакетам, например, latency), Summary (как Histogram, но с квантилями, вычисляемыми на клиенте).

Кастомные метрики в Go создаются через библиотеку `prometheus/client_golang`: определяются глобальные переменные метрик, регистрируются через `prometheus.MustRegister()`, обновляются в бизнес-логике (Inc/Add/Set), и экспонируются через `promhttp.Handler()`. Критично правильно выбирать labels для метрик — они создают новые time-series, и высокая cardinality (много уникальных комбинаций labels) приводит к проблемам с памятью Prometheus. В моем API-сервисе я добавил кастомные метрики: `http_requests_total` (Counter с labels method, path, status для подсчета запросов), `order_processing_duration_seconds` (Histogram для latency обработки заказов), `active_websocket_connections` (Gauge для текущего количества WebSocket-соединений), что позволило обнаружить и оптимизировать медленный endpoint `/api/orders` с p99 latency 850ms через alert в Prometheus.

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// ========== МЕТРИКИ ==========

var (
    // Counter: монотонно растущее значение
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"}, // labels
    )
    
    // Histogram: распределение значений (latency)
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}, // бакеты в секундах
        },
        []string{"method", "path"},
    )
    
    // Gauge: изменяющееся значение
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Current number of active connections",
        },
    )
    
    // Gauge: количество горутин
    goroutineCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutine_count",
            Help: "Current number of goroutines",
        },
    )
    
    // Counter: бизнес-метрика
    ordersProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_processed_total",
            Help: "Total number of orders processed",
        },
        []string{"status"}, // success, failed
    )
    
    // Histogram: время обработки заказа
    orderProcessingDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "order_processing_duration_seconds",
            Help:    "Order processing duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
    )
)

// ========== MIDDLEWARE ДЛЯ АВТОМАТИЧЕСКОГО СБОРА МЕТРИК ==========

func prometheusMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Увеличиваем счетчик активных соединений
        activeConnections.Inc()
        defer activeConnections.Dec()
        
        // Создаем ResponseWriter для перехвата статус-кода
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Вызываем хендлер
        next(rw, r)
        
        // Записываем метрики
        duration := time.Since(start).Seconds()
        
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", rw.statusCode)).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// ========== БИЗНЕС-ЛОГИКА С МЕТРИКАМИ ==========

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Симуляция обработки заказа
    processingTime := time.Duration(rand.Intn(1000)) * time.Millisecond
    time.Sleep(processingTime)
    
    // Случайно успешно или с ошибкой
    success := rand.Float32() > 0.1
    
    if success {
        ordersProcessed.WithLabelValues("success").Inc()
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte(`{"order_id": 12345, "status": "created"}`))
    } else {
        ordersProcessed.WithLabelValues("failed").Inc()
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error": "processing failed"}`))
    }
    
    // Записываем время обработки
    orderProcessingDuration.Observe(time.Since(start).Seconds())
}

func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
    // Симуляция запроса
    time.Sleep(50 * time.Millisecond)
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"orders": []}`))
}

// ========== ФОНОВЫЙ СБОРЩИК СИСТЕМНЫХ МЕТРИК ==========

func collectSystemMetrics() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Собираем количество горутин
        goroutineCount.Set(float64(runtime.NumGoroutine()))
    }
}

// ========== PROMETHEUS ALERTING RULES ==========

func explainAlertingRules() {
    fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║ PROMETHEUS ALERTING RULES (prometheus.rules.yml)                  ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║ groups:                                                           ║
║ - name: api_alerts                                                ║
║   interval: 30s                                                   ║
║   rules:                                                          ║
║   # Alert при высоком p99 latency                                 ║
║   - alert: HighLatency                                            ║
║     expr: |                                                       ║
║       histogram_quantile(0.99,                                    ║
║         rate(http_request_duration_seconds_bucket[5m])            ║
║       ) > 1.0                                                     ║
║     for: 2m                                                       ║
║     labels:                                                       ║
║       severity: warning                                           ║
║     annotations:                                                  ║
║       summary: "High p99 latency"                                 ║
║       description: "p99 latency > 1s for 2 minutes"               ║
║                                                                   ║
║   # Alert при высокой частоте ошибок                              ║
║   - alert: HighErrorRate                                          ║
║     expr: |                                                       ║
║       rate(http_requests_total{status=~"5.."}[5m])                ║
║       / rate(http_requests_total[5m]) > 0.05                      ║
║     for: 1m                                                       ║
║     labels:                                                       ║
║       severity: critical                                          ║
║     annotations:                                                  ║
║       summary: "High error rate"                                  ║
║       description: "Error rate > 5% for 1 minute"                 ║
║                                                                   ║
║   # Alert при падении throughput                                  ║
║   - alert: LowThroughput                                          ║
║     expr: |                                                       ║
║       rate(http_requests_total[5m]) < 10                          ║
║     for: 5m                                                       ║
║     labels:                                                       ║
║       severity: warning                                           ║
║     annotations:                                                  ║
║       summary: "Low request throughput"                           ║
║       description: "RPS < 10 for 5 minutes"                       ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
    `)
}

func main() {
    // Запускаем сборщик системных метрик
    go collectSystemMetrics()
    
    // HTTP хендлеры с метриками
    http.HandleFunc("/api/orders", prometheusMiddleware(createOrderHandler))
    http.HandleFunc("/api/orders/list", prometheusMiddleware(getOrdersHandler))
    
    // Endpoint для Prometheus scraping
    http.Handle("/metrics", promhttp.Handler())
    
    // Health check
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    explainAlertingRules()
    
    fmt.Println("Server starting on :8080")
    fmt.Println("Metrics available at http://localhost:8080/metrics")
    http.ListenAndServe(":8080", nil)
}
```

**Prometheus конфигурация (prometheus.yml):**

```yaml
global:
  scrape_interval: 15s      # Как часто собирать метрики
  evaluation_interval: 15s  # Как часто проверять alerting rules

# Targets для scraping
scrape_configs:
  - job_name: 'api-service'
    static_configs:
      - targets: ['localhost:8080']
        labels:
          service: 'api'
          environment: 'production'
    
    # Здоровье target'а
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:8080

# Alerting rules
rule_files:
  - 'prometheus.rules.yml'

# Alertmanager для отправки уведомлений
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

**Полезные PromQL запросы:**

```promql
# RPS (requests per second)
rate(http_requests_total[5m])

# p50, p95, p99 latency
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Error rate (%)
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100

# Top 5 самых медленных endpoints
topk(5, 
  histogram_quantile(0.99, 
    rate(http_request_duration_seconds_bucket[5m])
  ) by (path)
)

# Количество успешных заказов за последний час
increase(orders_processed_total{status="success"}[1h])
```

---

# ФАЗА 3: Экспертные темы и системный дизайн

## 🔷 ФАЗА 3: Go Advanced Topics

### 18. Что такое интерфейсы в Go? Как работает duck typing?

**Ответ:**
Интерфейсы в Go — это коллекции сигнатур методов, определяющие поведение без привязки к конкретной реализации, и типы реализуют интерфейсы неявно (implicitly) просто через наличие всех требуемых методов, что называется "duck typing" или structural typing. В отличие от явной реализации интерфейсов в Java/C# (implements/extends), в Go если тип имеет все методы интерфейса, он автоматически удовлетворяет этому интерфейсу, что обеспечивает слабую связанность и простоту рефакторинга. Пустой интерфейс `interface{}` (или `any` с Go 1.18+) не требует никаких методов и может содержать значение любого типа, используется для generic-кода до появления дженериков.

Интерфейсы в Go хранятся как пара (type, value) — динамический тип конкретного значения и указатель на само значение, что позволяет проверять тип через type assertion `v.(Type)` или type switch. Важная особенность: nil interface != interface содержащий nil pointer, что часто приводит к багам при возврате nil из функций. В моей архитектуре микросервисов я активно использовал интерфейсы для dependency injection: определил интерфейсы `UserRepository`, `OrderRepository` с методами `Create`, `Get`, `Update`, что позволило легко подменять реальные реализации на mock'и в unit-тестах и менять PostgreSQL на MongoDB без изменения бизнес-логики.

**Пример на Golang:**

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// ========== ОПРЕДЕЛЕНИЕ ИНТЕРФЕЙСОВ ==========

// Интерфейс для хранилища пользователей
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

// Интерфейс для отправки уведомлений
type Notifier interface {
    Send(ctx context.Context, message string) error
}

// Композиция интерфейсов
type NotificationService interface {
    Notifier
    SendBulk(ctx context.Context, messages []string) error
}

// ========== РЕАЛИЗАЦИИ ==========

type User struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
}

// PostgreSQL реализация UserRepository
type PostgresUserRepository struct {
    // db connection
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *User) error {
    fmt.Printf("PostgreSQL: Creating user %s\n", user.Email)
    user.ID = 123 // генерированный ID
    user.CreatedAt = time.Now()
    return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    fmt.Printf("PostgreSQL: Getting user %d\n", id)
    return &User{ID: id, Email: "user@example.com", Name: "John Doe"}, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *User) error {
    fmt.Printf("PostgreSQL: Updating user %d\n", user.ID)
    return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
    fmt.Printf("PostgreSQL: Deleting user %d\n", id)
    return nil
}

// MongoDB реализация UserRepository (duck typing - автоматически реализует интерфейс!)
type MongoUserRepository struct {
    // mongo connection
}

func (r *MongoUserRepository) Create(ctx context.Context, user *User) error {
    fmt.Printf("MongoDB: Inserting user %s\n", user.Email)
    return nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    fmt.Printf("MongoDB: Finding user %d\n", id)
    return &User{ID: id, Email: "mongo@example.com"}, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, user *User) error {
    fmt.Printf("MongoDB: Updating document for user %d\n", user.ID)
    return nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id int64) error {
    fmt.Printf("MongoDB: Removing user %d\n", id)
    return nil
}

// Email Notifier
type EmailNotifier struct{}

func (n *EmailNotifier) Send(ctx context.Context, message string) error {
    fmt.Printf("Sending email: %s\n", message)
    return nil
}

// SMS Notifier - также реализует Notifier
type SMSNotifier struct{}

func (n *SMSNotifier) Send(ctx context.Context, message string) error {
    fmt.Printf("Sending SMS: %s\n", message)
    return nil
}

// ========== БИЗНЕС-ЛОГИКА С DEPENDENCY INJECTION ==========

type UserService struct {
    repo     UserRepository  // интерфейс, а не конкретная реализация
    notifier Notifier
}

func NewUserService(repo UserRepository, notifier Notifier) *UserService {
    return &UserService{
        repo:     repo,
        notifier: notifier,
    }
}

func (s *UserService) RegisterUser(ctx context.Context, email, name string) error {
    user := &User{
        Email: email,
        Name:  name,
    }
    
    // Создаем пользователя (не важно, PostgreSQL или MongoDB)
    if err := s.repo.Create(ctx, user); err != nil {
        return err
    }
    
    // Отправляем уведомление (не важно, Email или SMS)
    message := fmt.Sprintf("Welcome, %s!", name)
    if err := s.notifier.Send(ctx, message); err != nil {
        return err
    }
    
    return nil
}

// ========== TYPE ASSERTION И TYPE SWITCH ==========

func demonstrateTypeAssertion() {
    var i interface{} = "hello"
    
    // Type assertion
    s, ok := i.(string)
    if ok {
        fmt.Printf("i is a string: %s\n", s)
    }
    
    // Panic если тип не совпадает (без ok)
    // s := i.(string)
    
    // Type assertion для вызова специфичных методов
    var notifier Notifier = &EmailNotifier{}
    
    if emailNotifier, ok := notifier.(*EmailNotifier); ok {
        fmt.Printf("It's an EmailNotifier: %T\n", emailNotifier)
    }
}

func processValue(v interface{}) {
    // Type switch
    switch val := v.(type) {
    case string:
        fmt.Printf("String: %s (length: %d)\n", val, len(val))
    case int:
        fmt.Printf("Integer: %d (doubled: %d)\n", val, val*2)
    case bool:
        fmt.Printf("Boolean: %v\n", val)
    case User:
        fmt.Printf("User: %s (%s)\n", val.Name, val.Email)
    case *User:
        fmt.Printf("User pointer: %s (%s)\n", val.Name, val.Email)
    case nil:
        fmt.Println("Nil value")
    default:
        fmt.Printf("Unknown type: %T\n", val)
    }
}

// ========== NIL INTERFACE TRAP ==========

func demonstrateNilInterfaceTrap() {
    fmt.Println("\n=== Nil Interface Trap ===")
    
    var repo UserRepository
    fmt.Printf("repo == nil: %v (type: %T, value: %v)\n", repo == nil, repo, repo)
    
    // Опасность: присваивание nil pointer
    var pgRepo *PostgresUserRepository = nil
    repo = pgRepo  // repo теперь НЕ nil!
    
    fmt.Printf("repo == nil: %v (type: %T, value: %v)\n", repo == nil, repo, repo)
    fmt.Println("Почему? Interface содержит (type=*PostgresUserRepository, value=nil)")
    
    // Правильная проверка
    if repo == nil || repo.(*PostgresUserRepository) == nil {
        fmt.Println("Repository is truly nil")
    }
}

// ========== MOCK ДЛЯ ТЕСТИРОВАНИЯ ==========

type MockUserRepository struct {
    users map[int64]*User
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        users: make(map[int64]*User),
    }
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    user.ID = int64(len(m.users) + 1)
    m.users[user.ID] = user
    return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    user, exists := m.users[id]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *User) error {
    m.users[user.ID] = user
    return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
    delete(m.users, id)
    return nil
}

func main() {
    ctx := context.Background()
    
    fmt.Println("=== Duck Typing - PostgreSQL ===")
    pgRepo := &PostgresUserRepository{}
    emailNotifier := &EmailNotifier{}
    service1 := NewUserService(pgRepo, emailNotifier)
    service1.RegisterUser(ctx, "alice@example.com", "Alice")
    
    fmt.Println("\n=== Duck Typing - MongoDB ===")
    mongoRepo := &MongoUserRepository{}
    smsNotifier := &SMSNotifier{}
    service2 := NewUserService(mongoRepo, smsNotifier)
    service2.RegisterUser(ctx, "bob@example.com", "Bob")
    
    fmt.Println("\n=== Mock для тестирования ===")
    mockRepo := NewMockUserRepository()
    service3 := NewUserService(mockRepo, emailNotifier)
    service3.RegisterUser(ctx, "test@example.com", "Test User")
    
    user, _ := mockRepo.GetByID(ctx, 1)
    fmt.Printf("Retrieved from mock: %+v\n", user)
    
    fmt.Println("\n=== Type Assertion ===")
    demonstrateTypeAssertion()
    
    fmt.Println("\n=== Type Switch ===")
    processValue("hello")
    processValue(42)
    processValue(true)
    processValue(User{Name: "John", Email: "john@test.com"})
    
    demonstrateNilInterfaceTrap()
}
```

---

### 19. Что такое generics в Go 1.18+? Как их использовать?

**Ответ:**
Generics (дженерики) — это механизм параметрического полиморфизма, добавленный в Go 1.18, позволяющий писать функции и типы данных, работающие с различными типами без дублирования кода и без потери типобезопасности через type parameters в квадратных скобках. До Go 1.18 общий код писался либо через `interface{}` с потерей типобезопасности и необходимостью type assertion, либо через code generation (go generate), что было громоздко. Generics используют синтаксис `func Name[T constraint](param T) T`, где T — type parameter, ограниченный constraint'ом (например, `any`, `comparable`, или custom interface).

Type constraints определяют допустимые типы для type parameter через интерфейсы с union types (pipe `|`) или методами: `interface{ ~int | ~float64 }` означает "любой тип с underlying type int или float64", что позволяет работать как с `int`, так и с `type MyInt int`. Основные встроенные constraints: `any` (=`interface{}`), `comparable` (типы, поддерживающие == и !=), и в пакете `golang.org/x/exp/constraints` есть `Ordered`, `Integer`, `Float`. В моей библиотеке утилит я создал generic функции `Map[T, U]`, `Filter[T]`, `Reduce[T]` для работы со слайсами любых типов, что устранило дублирование 15+ функций для разных типов и сделало код более выразительным без потери производительности (generics в Go мономорфизируются на этапе компиляции).

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "golang.org/x/exp/constraints"
)

// ========== БАЗОВЫЕ GENERIC ФУНКЦИИ ==========

// Generic функция для поиска минимума
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// Generic функция для поиска в слайсе
func Contains[T comparable](slice []T, value T) bool {
    for _, item := range slice {
        if item == value {
            return true
        }
    }
    return false
}

// Generic Map: преобразование слайса
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, item := range slice {
        result[i] = fn(item)
    }
    return result
}

// Generic Filter: фильтрация слайса
func Filter[T any](slice []T, predicate func(T) bool) []T {
    result := make([]T, 0)
    for _, item := range slice {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

// Generic Reduce: свертка слайса
func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, item := range slice {
        result = fn(result, item)
    }
    return result
}

// ========== GENERIC ТИПЫ ДАННЫХ ==========

// Generic Stack
type Stack[T any] struct {
    items []T
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{
        items: make([]T, 0),
    }
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) IsEmpty() bool {
    return len(s.items) == 0
}

// Generic Cache
type Cache[K comparable, V any] struct {
    data map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
    return &Cache[K, V]{
        data: make(map[K]V),
    }
}

func (c *Cache[K, V]) Set(key K, value V) {
    c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
    value, exists := c.data[key]
    return value, exists
}

func (c *Cache[K, V]) Delete(key K) {
    delete(c.data, key)
}

// ========== CUSTOM CONSTRAINTS ==========

// Constraint для числовых типов
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}

// Generic функция для суммирования чисел
func Sum[T Number](numbers []T) T {
    var sum T
    for _, num := range numbers {
        sum += num
    }
    return sum
}

// Generic функция для среднего значения
func Average[T Number](numbers []T) float64 {
    if len(numbers) == 0 {
        return 0
    }
    sum := Sum(numbers)
    return float64(sum) / float64(len(numbers))
}

// Constraint с методами
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// ========== GENERIC МЕТОДЫ ==========

type Repository[T any] struct {
    items map[int64]T
    nextID int64
}

func NewRepository[T any]() *Repository[T] {
    return &Repository[T]{
        items: make(map[int64]T),
        nextID: 1,
    }
}

func (r *Repository[T]) Create(item T) int64 {
    id := r.nextID
    r.items[id] = item
    r.nextID++
    return id
}

func (r *Repository[T]) GetByID(id int64) (T, bool) {
    item, exists := r.items[id]
    return item, exists
}

func (r *Repository[T]) Update(id int64, item T) bool {
    if _, exists := r.items[id]; !exists {
        return false
    }
    r.items[id] = item
    return true
}

func (r *Repository[T]) Delete(id int64) bool {
    if _, exists := r.items[id]; !exists {
        return false
    }
    delete(r.items, id)
    return true
}

func (r *Repository[T]) GetAll() []T {
    result := make([]T, 0, len(r.items))
    for _, item := range r.items {
        result = append(result, item)
    }
    return result
}

// ========== ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ ==========

type User struct {
    ID    int64
    Name  string
    Email string
}

func (u User) String() string {
    return fmt.Sprintf("User{ID: %d, Name: %s, Email: %s}", u.ID, u.Name, u.Email)
}

type Product struct {
    ID    int64
    Name  string
    Price float64
}

func (p Product) String() string {
    return fmt.Sprintf("Product{ID: %d, Name: %s, Price: %.2f}", p.ID, p.Name, p.Price)
}

func main() {
    fmt.Println("=== Generic функции ===")
    
    // Min с разными типами
    fmt.Println("Min(3, 5):", Min(3, 5))
    fmt.Println("Min(3.14, 2.71):", Min(3.14, 2.71))
    fmt.Println("Min(\"apple\", \"banana\"):", Min("apple", "banana"))
    
    // Contains
    numbers := []int{1, 2, 3, 4, 5}
    fmt.Println("Contains(numbers, 3):", Contains(numbers, 3))
    fmt.Println("Contains(numbers, 10):", Contains(numbers, 10))
    
    strings := []string{"go", "rust", "python"}
    fmt.Println("Contains(strings, \"go\"):", Contains(strings, "go"))
    
    // Map
    doubled := Map(numbers, func(n int) int { return n * 2 })
    fmt.Println("Doubled:", doubled)
    
    lengths := Map(strings, func(s string) int { return len(s) })
    fmt.Println("String lengths:", lengths)
    
    // Filter
    evens := Filter(numbers, func(n int) bool { return n%2 == 0 })
    fmt.Println("Even numbers:", evens)
    
    // Reduce
    sum := Reduce(numbers, 0, func(acc, n int) int { return acc + n })
    fmt.Println("Sum:", sum)
    
    fmt.Println("\n=== Generic типы данных ===")
    
    // Stack
    intStack := NewStack[int]()
    intStack.Push(1)
    intStack.Push(2)
    intStack.Push(3)
    fmt.Println("Stack pop:", intStack.Pop())
    fmt.Println("Stack pop:", intStack.Pop())
    
    stringStack := NewStack[string]()
    stringStack.Push("hello")
    stringStack.Push("world")
    fmt.Println("String stack pop:", stringStack.Pop())
    
    // Cache
    userCache := NewCache[int64, User]()
    userCache.Set(1, User{ID: 1, Name: "Alice", Email: "alice@example.com"})
    userCache.Set(2, User{ID: 2, Name: "Bob", Email: "bob@example.com"})
    
    if user, exists := userCache.Get(1); exists {
        fmt.Printf("User from cache: %+v\n", user)
    }
    
    fmt.Println("\n=== Custom constraints ===")
    
    intNums := []int{1, 2, 3, 4, 5}
    floatNums := []float64{1.1, 2.2, 3.3}
    
    fmt.Println("Sum of ints:", Sum(intNums))
    fmt.Println("Sum of floats:", Sum(floatNums))
    fmt.Println("Average of ints:", Average(intNums))
    fmt.Println("Average of floats:", Average(floatNums))
    
    fmt.Println("\n=== Generic Repository ===")
    
    userRepo := NewRepository[User]()
    id1 := userRepo.Create(User{Name: "Alice", Email: "alice@test.com"})
    id2 := userRepo.Create(User{Name: "Bob", Email: "bob@test.com"})
    
    fmt.Printf("Created users with IDs: %d, %d\n", id1, id2)
    
    if user, exists := userRepo.GetByID(id1); exists {
        fmt.Printf("Found user: %+v\n", user)
    }
    
    allUsers := userRepo.GetAll()
    fmt.Printf("All users: %+v\n", allUsers)
    
    productRepo := NewRepository[Product]()
    productRepo.Create(Product{Name: "Laptop", Price: 999.99})
    productRepo.Create(Product{Name: "Mouse", Price: 29.99})
    
    allProducts := productRepo.GetAll()
    fmt.Printf("All products: %+v\n", allProducts)
    
    fmt.Println("\n=== Stringer constraint ===")
    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@test.com"},
        {ID: 2, Name: "Bob", Email: "bob@test.com"},
    }
    PrintAll(users)
}
```

**Когда использовать generics:**
```
╔═══════════════════════════════════════════════════════════════════╗
║ КОГДА ИСПОЛЬЗОВАТЬ GENERICS                                       ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║ ✅ ХОРОШИЕ КЕЙСЫ:                                                  ║
║   • Контейнеры данных (Stack, Queue, Tree, Graph)                ║
║   • Функции высшего порядка (Map, Filter, Reduce)                ║
║   • Репозитории и DAO (однотипные CRUD операции)                 ║
║   • Кэши и пулы объектов                                          ║
║   • Утилитарные функции (Min, Max, Sort)                         ║
║                                                                   ║
║ ❌ ПЛОХИЕ КЕЙСЫ:                                                   ║
║   • Бизнес-логика (лучше интерфейсы)                             ║
║   • HTTP хендлеры (слишком специфичные)                          ║
║   • Когда нужны методы на type parameter                         ║
║   • Reflection более уместен                                      ║
║                                                                   ║
║ ⚠️  ОГРАНИЧЕНИЯ:                                                   ║
║   • Нельзя объявить методы на type parameter                     ║
║   • Нельзя использовать type assertions внутри generic функций   ║
║   • Увеличивают время компиляции (мономорфизация)                ║
║   • Ошибки компиляции могут быть сложными для понимания          ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

## 🔷 ФАЗА 3: Security и аутентификация

### 20. Что такое JWT (JSON Web Token)? Как работает JWT-аутентификация?

**Ответ:**
JWT (JSON Web Token) — это компактный, URL-safe токен, представляющий claims (утверждения) в формате JSON, подписанный криптографически для обеспечения целостности и опционально зашифрованный для конфиденциальности. JWT состоит из трех частей, разделенных точками: Header (алгоритм и тип токена в base64url), Payload (claims - данные о пользователе, expiration time, issuer в base64url), Signature (HMAC или RSA подпись для проверки целостности = sign(base64(header) + "." + base64(payload), secret)). JWT часто используется для stateless аутентификации: сервер генерирует JWT при логине, клиент сохраняет в localStorage/cookie и отправляет в Authorization header при каждом запросе, сервер проверяет подпись и извлекает данные пользователя без обращения к БД.

Преимущества JWT: stateless (не нужна сессия на сервере), масштабируемость (любой сервер может валидировать токен зная secret), кросс-доменность (работает между разными сервисами), компактность (можно передавать в URL/header). Недостатки: нельзя отозвать до истечения (expires), размер больше чем session ID, уязвим к XSS если хранится в localStorage (лучше httpOnly cookie), нужно аккуратно выбирать expiration time (короткий = частый relogin, длинный = риск). В моем auth-сервисе я реализовал JWT с access token (5 минут) и refresh token (7 дней): access token используется для API-запросов, а refresh token (хранится в httpOnly cookie) обновляет access token без повторного логина, и храню список отозванных refresh tokens в Redis для logout функциональности.

**Пример на Golang:**

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// ========== JWT CLAIMS ==========

type Claims struct {
    UserID   int64  `json:"user_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// ========== JWT SERVICE ==========

type JWTService struct {
    accessSecret  []byte
    refreshSecret []byte
    accessTTL     time.Duration
    refreshTTL    time.Duration
}

func NewJWTService(accessSecret, refreshSecret string) *JWTService {
    return &JWTService{
        accessSecret:  []byte(accessSecret),
        refreshSecret: []byte(refreshSecret),
        accessTTL:     5 * time.Minute,   // Access token живет 5 минут
        refreshTTL:    7 * 24 * time.Hour, // Refresh token живет 7 дней
    }
}

// Генерация Access Token
func (s *JWTService) GenerateAccessToken(userID int64, email, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "my-app",
            Subject:   fmt.Sprintf("%d", userID),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.accessSecret)
}

// Генерация Refresh Token
func (s *JWTService) GenerateRefreshToken(userID int64) (string, error) {
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "my-app",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.refreshSecret)
}

// Валидация Access Token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Проверяем алгоритм подписи
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.accessSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token")
}

// Валидация Refresh Token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.refreshSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid refresh token")
}

// ========== AUTH SERVICE ==========

type User struct {
    ID       int64
    Email    string
    Password string // hashed
    Role     string
}

type AuthService struct {
    jwtService *JWTService
    users      map[string]*User // email -> user (в реальности - БД)
}

func NewAuthService(jwtService *JWTService) *AuthService {
    return &AuthService{
        jwtService: jwtService,
        users: map[string]*User{
            "alice@example.com": {ID: 1, Email: "alice@example.com", Password: "hashed_password", Role: "admin"},
            "bob@example.com":   {ID: 2, Email: "bob@example.com", Password: "hashed_password", Role: "user"},
        },
    }
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Проверяем credentials (в реальности - bcrypt.CompareHashAndPassword)
    user, exists := s.users[req.Email]
    if !exists || user.Password != "hashed_password" {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Генерируем токены
    accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)
    if err != nil {
        http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
        return
    }
    
    refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
    if err != nil {
        http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
        return
    }
    
    // Отправляем refresh token в httpOnly cookie (защита от XSS)
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Path:     "/",
        Expires:  time.Now().Add(7 * 24 * time.Hour),
        HttpOnly: true,  // Недоступен для JavaScript
        Secure:   true,  // Только HTTPS
        SameSite: http.SameSiteStrictMode,
    })
    
    // Отправляем access token в response body
    response := LoginResponse{
        AccessToken: accessToken,
        ExpiresIn:   int64(s.jwtService.accessTTL.Seconds()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *AuthService) RefreshToken(w http.ResponseWriter, r *http.Request) {
    // Получаем refresh token из cookie
    cookie, err := r.Cookie("refresh_token")
    if err != nil {
        http.Error(w, "Refresh token not found", http.StatusUnauthorized)
        return
    }
    
    // Валидируем refresh token
    claims, err := s.jwtService.ValidateRefreshToken(cookie.Value)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }
    
    // Генерируем новый access token
    user := s.users["alice@example.com"] // В реальности - GetByID из БД
    accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role)
    if err != nil {
        http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
        return
    }
    
    response := LoginResponse{
        AccessToken: accessToken,
        ExpiresIn:   int64(s.jwtService.accessTTL.Seconds()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
    
    fmt.Printf("Refreshed token for user %d\n", claims.UserID)
}

// ========== JWT MIDDLEWARE ==========

type contextKey string

const userContextKey contextKey = "user"

func (s *AuthService) JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Извлекаем токен из Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        // Формат: "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }
        
        tokenString := parts[1]
        
        // Валидируем токен
        claims, err := s.jwtService.ValidateAccessToken(tokenString)
        if err != nil {
            http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
            return
        }
        
        // Добавляем claims в контекст
        ctx := context.WithValue(r.Context(), userContextKey, claims)
        
        // Передаем управление следующему хендлеру
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Middleware для проверки роли
func (s *AuthService) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value(userContextKey).(*Claims)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            if claims.Role != role {
                http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// ========== PROTECTED ENDPOINTS ==========

func profileHandler(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value(userContextKey).(*Claims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    response := map[string]interface{}{
        "user_id": claims.UserID,
        "email":   claims.Email,
        "role":    claims.Role,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Admin only content"))
}

func main() {
    jwtService := NewJWTService("my-secret-key", "my-refresh-secret")
    authService := NewAuthService(jwtService)
    
    mux := http.NewServeMux()
    
    // Public endpoints
    mux.HandleFunc("/api/auth/login", authService.Login)
    mux.HandleFunc("/api/auth/refresh", authService.RefreshToken)
    
    // Protected endpoints
    mux.Handle("/api/profile", authService.JWTMiddleware(http.HandlerFunc(profileHandler)))
    mux.Handle("/api/admin", authService.JWTMiddleware(
        authService.RequireRole("admin")(http.HandlerFunc(adminHandler)),
    ))
    
    fmt.Println("Server starting on :8080")
    fmt.Println("Try:")
    fmt.Println("  curl -X POST http://localhost:8080/api/auth/login -d '{\"email\":\"alice@example.com\",\"password\":\"password\"}'")
    fmt.Println("  curl http://localhost:8080/api/profile -H 'Authorization: Bearer <token>'")
    
    http.ListenAndServe(":8080", mux)
}
```

---

## 🔷 ФАЗА 3: Алгоритмы и структуры данных

### 21. Алгоритм A* (A-star) - поиск кратчайшего пути

**Ответ:**
A* (A-star) — это эвристический алгоритм поиска кратчайшего пути на графе, комбинирующий идеи Dijkstra (учет реальной стоимости пути g(n)) и жадного best-first search (эвристическая оценка до цели h(n)) через функцию f(n) = g(n) + h(n), где g(n) — стоимость пути от старта до узла n, h(n) — эвристическая оценка от n до цели (например, евклидово расстояние или манхэттенское расстояние). A* гарантирует нахождение оптимального пути, если эвристика допустима (никогда не переоценивает реальную стоимость, h(n) ≤ h*(n)) и консистентна (монотонна, h(n) ≤ cost(n, n') + h(n')). Алгоритм использует две структуры данных: open set (приоритетная очередь узлов для исследования, сортированных по f(n)) и closed set (множество уже исследованных узлов).

Преимущества A* над Dijkstra: в среднем исследует меньше узлов благодаря эвристике, направляющей поиск к цели, что критично для больших графов (карты, игровые уровни). Преимущества A* над greedy best-first: гарантирует оптимальность пути, а не просто быстрое нахождение любого пути. В моем проекте маршрутизации я реализовал A* для поиска оптимальных путей в графе дорог города: использовал манхэттенское расстояние как эвристику для сетки улиц, что ускорило поиск на графе в 50k узлов с ~7 секунд (Dijkstra) до ~800ms (A*), сохранив оптимальность маршрутов.

**Пример на Golang:**

```go
package main

import (
    "container/heap"
    "fmt"
    "math"
)

// ========== ГРАФ И УЗЛЫ ==========

type Point struct {
    X, Y int
}

type Node struct {
    Point    Point
    G        float64 // Стоимость от старта до этого узла
    H        float64 // Эвристическая оценка до цели
    F        float64 // G + H
    Parent   *Node
    index    int     // Для priority queue
}

type Graph struct {
    Grid     [][]int  // 0 = проходимо, 1 = препятствие
    Width    int
    Height   int
}

// ========== PRIORITY QUEUE (MIN-HEAP) ==========

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].F < pq[j].F // Min-heap по F
}

func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    node := x.(*Node)
    node.index = n
    *pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    node := old[n-1]
    old[n-1] = nil
    node.index = -1
    *pq = old[0 : n-1]
    return node
}

// ========== ЭВРИСТИЧЕСКИЕ ФУНКЦИИ ==========

// Манхэттенское расстояние (для сетки без диагоналей)
func manhattanDistance(a, b Point) float64 {
    return math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y))
}

// Евклидово расстояние (для любых графов)
func euclideanDistance(a, b Point) float64 {
    dx := float64(a.X - b.X)
    dy := float64(a.Y - b.Y)
    return math.Sqrt(dx*dx + dy*dy)
}

// Диагональное расстояние (Chebyshev) для сетки с диагоналями
func diagonalDistance(a, b Point) float64 {
    dx := math.Abs(float64(a.X - b.X))
    dy := math.Abs(float64(a.Y - b.Y))
    return math.Max(dx, dy)
}

// ========== A* АЛГОРИТМ ==========

func (g *Graph) AStar(start, goal Point) []*Node {
    // Open set: узлы для исследования
    openSet := &PriorityQueue{}
    heap.Init(openSet)
    
    // Closed set: уже исследованные узлы
    closedSet := make(map[Point]bool)
    
    // Стартовый узел
    startNode := &Node{
        Point:  start,
        G:      0,
        H:      manhattanDistance(start, goal),
        F:      0 + manhattanDistance(start, goal),
        Parent: nil,
    }
    
    heap.Push(openSet, startNode)
    
    // Для быстрого поиска узлов в open set
    inOpenSet := make(map[Point]*Node)
    inOpenSet[start] = startNode
    
    for openSet.Len() > 0 {
        // Берем узел с минимальным F
        current := heap.Pop(openSet).(*Node)
        delete(inOpenSet, current.Point)
        
        // Достигли цели?
        if current.Point == goal {
            return reconstructPath(current)
        }
        
        // Добавляем в closed set
        closedSet[current.Point] = true
        
        // Исследуем соседей
        for _, neighbor := range g.getNeighbors(current.Point) {
            // Пропускаем если уже исследован
            if closedSet[neighbor] {
                continue
            }
            
            // Стоимость перехода (обычно 1, но может зависеть от типа дороги)
            tentativeG := current.G + g.getCost(current.Point, neighbor)
            
            // Проверяем, есть ли сосед в open set
            neighborNode, exists := inOpenSet[neighbor]
            
            if !exists {
                // Создаем новый узел
                neighborNode = &Node{
                    Point:  neighbor,
                    G:      tentativeG,
                    H:      manhattanDistance(neighbor, goal),
                    F:      tentativeG + manhattanDistance(neighbor, goal),
                    Parent: current,
                }
                heap.Push(openSet, neighborNode)
                inOpenSet[neighbor] = neighborNode
            } else if tentativeG < neighborNode.G {
                // Нашли лучший путь к этому узлу
                neighborNode.G = tentativeG
                neighborNode.F = tentativeG + neighborNode.H
                neighborNode.Parent = current
                heap.Fix(openSet, neighborNode.index)
            }
        }
    }
    
    // Путь не найден
    return nil
}

// Восстановление пути от цели к старту
func reconstructPath(node *Node) []*Node {
    path := make([]*Node, 0)
    for node != nil {
        path = append([]*Node{node}, path...) // Prepend
        node = node.Parent
    }
    return path
}

// Получение соседей (4 направления или 8 с диагоналями)
func (g *Graph) getNeighbors(p Point) []Point {
    neighbors := make([]Point, 0, 8)
    
    // 4 основных направления
    directions := []Point{
        {0, -1},  // вверх
        {1, 0},   // вправо
        {0, 1},   // вниз
        {-1, 0},  // влево
    }
    
    // Добавляем диагонали (опционально)
    // directions = append(directions, Point{1, -1}, Point{1, 1}, Point{-1, 1}, Point{-1, -1})
    
    for _, dir := range directions {
        next := Point{p.X + dir.X, p.Y + dir.Y}
        
        // Проверяем границы
        if next.X >= 0 && next.X < g.Width && next.Y >= 0 && next.Y < g.Height {
            // Проверяем проходимость
            if g.Grid[next.Y][next.X] == 0 {
                neighbors = append(neighbors, next)
            }
        }
    }
    
    return neighbors
}

// Стоимость перехода (можно варьировать для разных типов поверхности)
func (g *Graph) getCost(from, to Point) float64 {
    // Простая стоимость = 1
    // Можно сделать сложнее: дороги = 1, грязь = 2, вода = 3
    return 1.0
}

// ========== ВИЗУАЛИЗАЦИЯ ==========

func (g *Graph) PrintPath(path []*Node) {
    // Копируем сетку
    display := make([][]rune, g.Height)
    for i := range display {
        display[i] = make([]rune, g.Width)
        for j := range display[i] {
            if g.Grid[i][j] == 1 {
                display[i][j] = '█' // Препятствие
            } else {
                display[i][j] = '·' // Пустое поле
            }
        }
    }
    
    // Рисуем путь
    if path != nil {
        for i, node := range path {
            if i == 0 {
                display[node.Point.Y][node.Point.X] = 'S' // Старт
            } else if i == len(path)-1 {
                display[node.Point.Y][node.Point.X] = 'G' // Цель
            } else {
                display[node.Point.Y][node.Point.X] = '*' // Путь
            }
        }
    }
    
    // Выводим
    for _, row := range display {
        fmt.Println(string(row))
    }
}

// ========== ПРИМЕР ИСПОЛЬЗОВАНИЯ ==========

func main() {
    // Создаем карту (0 = проходимо, 1 = препятствие)
    grid := [][]int{
        {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        {0, 1, 1, 1, 0, 0, 0, 1, 1, 0},
        {0, 0, 0, 1, 0, 0, 0, 0, 1, 0},
        {0, 0, 0, 1, 0, 1, 1, 0, 1, 0},
        {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        {0, 1, 1, 1, 1, 1, 0, 0, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 1, 1, 0},
        {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    }
    
    graph := &Graph{
        Grid:   grid,
        Width:  len(grid[0]),
        Height: len(grid),
    }
    
    start := Point{0, 0}
    goal := Point{9, 9}
    
    fmt.Println("Поиск пути от (0,0) до (9,9) используя A*")
    fmt.Println("\nКарта (█ = препятствие):")
    graph.PrintPath(nil)
    
    // Запускаем A*
    path := graph.AStar(start, goal)
    
    if path != nil {
        fmt.Printf("\n✅ Путь найден! Длина: %d шагов\n", len(path)-1)
        fmt.Println("\nПуть (S = старт, G = цель, * = путь):")
        graph.PrintPath(path)
        
        fmt.Println("\nПодробный путь:")
        for i, node := range path {
            fmt.Printf("Шаг %d: (%d, %d) F=%.1f, G=%.1f, H=%.1f\n", 
                i, node.Point.X, node.Point.Y, node.F, node.G, node.H)
        }
    } else {
        fmt.Println("\n❌ Путь не найден!")
    }
}
```

**Сравнение алгоритмов поиска пути:**

```
╔═══════════════════╦═══════════════════╦═══════════════════╦═══════════════════╗
║ Алгоритм          ║ Время             ║ Память            ║ Оптимальность     ║
╠═══════════════════╬═══════════════════╬═══════════════════╬═══════════════════╣
║ BFS               ║ O(V + E)          ║ O(V)              ║ ✅ (граф без весов)║
║ Dijkstra          ║ O((V+E) log V)    ║ O(V)              ║ ✅ Всегда          ║
║ A*                ║ O(E) best case    ║ O(V)              ║ ✅ (эвристика h≤h*)║
║ Greedy Best-First ║ O(E)              ║ O(V)              ║ ❌ Не гарантирует  ║
║ Bellman-Ford      ║ O(VE)             ║ O(V)              ║ ✅ (отриц. веса)   ║
╚═══════════════════╩═══════════════════╩═══════════════════╩═══════════════════╝

ПРИМЕНЕНИЕ A*:
• GPS навигация и маршрутизация
• Игровой AI (pathfinding для NPC)
• Роботика (планирование траектории)
• Puzzle solving (15-puzzle, Rubik's cube)
```

---

### 22. Алгоритм Dijkstra - кратчайший путь

**Ответ:**
Алгоритм Dijkstra находит кратчайшие пути от одной вершины (источника) до всех остальных вершин в взвешенном графе с неотрицательными весами ребер, используя жадную стратегию: на каждом шаге выбирается необработанная вершина с минимальным расстоянием от источника и обновляются расстояния до ее соседей через релаксацию ребер (если dist[u] + weight(u,v) < dist[v], то dist[v] = dist[u] + weight(u,v)). Алгоритм использует приоритетную очередь (min-heap) для эффективного выбора вершины с минимальным расстоянием, обеспечивая сложность O((V+E) log V) с бинарной кучей или O(V log V + E) с fibonacci heap. Dijkstra гарантирует нахождение оптимального пути для графов с неотрицательными весами, но не работает с отрицательными весами (для этого используется Bellman-Ford).

Разница между Dijkstra и A*: Dijkstra — это частный случай A* с нулевой эвристикой (h(n)=0), что делает его "слепым" поиском без направления к цели, исследующим все направления равномерно. A* использует эвристику для направленного поиска к конкретной цели, что быстрее для point-to-point запросов, но Dijkstra лучше когда нужны пути от источника ко всем вершинам. В моем сервисе логистики я использовал Dijkstra для предварительного расчета расстояний от склада до всех точек доставки (single-source shortest paths), что заняло ~3 секунды для графа в 10k узлов, и затем эти расстояния кэшировались для быстрых запросов оптимизации маршрутов.

**Пример на Golang:**

```go
package main

import (
    "container/heap"
    "fmt"
    "math"
)

// ========== ГРАФ ==========

type Edge struct {
    To     int
    Weight float64
}

type Graph struct {
    Vertices int
    Edges    [][]Edge // Adjacency list
}

func NewGraph(vertices int) *Graph {
    return &Graph{
        Vertices: vertices,
        Edges:    make([][]Edge, vertices),
    }
}

func (g *Graph) AddEdge(from, to int, weight float64) {
    g.Edges[from] = append(g.Edges[from], Edge{To: to, Weight: weight})
}

func (g *Graph) AddUndirectedEdge(from, to int, weight float64) {
    g.AddEdge(from, to, weight)
    g.AddEdge(to, from, weight)
}

// ========== PRIORITY QUEUE ==========

type Item struct {
    Vertex   int
    Distance float64
    index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].Distance < pq[j].Distance
}

func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*Item)
    item.index = n
    *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    old[n-1] = nil
    item.index = -1
    *pq = old[0 : n-1]
    return item
}

// ========== DIJKSTRA АЛГОРИТМ ==========

type DijkstraResult struct {
    Distances []float64  // Кратчайшие расстояния от источника
    Parents   []int      // Родители для восстановления путей
}

func (g *Graph) Dijkstra(source int) *DijkstraResult {
    // Инициализация
    distances := make([]float64, g.Vertices)
    parents := make([]int, g.Vertices)
    visited := make([]bool, g.Vertices)
    
    for i := range distances {
        distances[i] = math.Inf(1) // Бесконечность
        parents[i] = -1
    }
    distances[source] = 0
    
    // Priority queue
    pq := &PriorityQueue{}
    heap.Init(pq)
    heap.Push(pq, &Item{Vertex: source, Distance: 0})
    
    for pq.Len() > 0 {
        // Берем вершину с минимальным расстоянием
        current := heap.Pop(pq).(*Item)
        u := current.Vertex
        
        // Пропускаем если уже обработана
        if visited[u] {
            continue
        }
        visited[u] = true
        
        // Релаксация ребер
        for _, edge := range g.Edges[u] {
            v := edge.To
            weight := edge.Weight
            
            // Если нашли более короткий путь
            if distances[u]+weight < distances[v] {
                distances[v] = distances[u] + weight
                parents[v] = u
                heap.Push(pq, &Item{Vertex: v, Distance: distances[v]})
            }
        }
    }
    
    return &DijkstraResult{
        Distances: distances,
        Parents:   parents,
    }
}

// Восстановление пути от источника до цели
func (r *DijkstraResult) GetPath(target int) []int {
    if r.Parents[target] == -1 && r.Distances[target] == math.Inf(1) {
        return nil // Путь не найден
    }
    
    path := make([]int, 0)
    for v := target; v != -1; v = r.Parents[v] {
        path = append([]int{v}, path...) // Prepend
    }
    return path
}

// ========== ПРИМЕР ИСПОЛЬЗОВАНИЯ ==========

func main() {
    // Создаем граф
    //       (0)
    //      /   \
    //    4/     \2
    //    /       \
    //  (1)--7--(2)
    //   |        |
    //   1        3
    //   |        |
    //  (3)--2--(4)
    
    graph := NewGraph(5)
    graph.AddUndirectedEdge(0, 1, 4)
    graph.AddUndirectedEdge(0, 2, 2)
    graph.AddUndirectedEdge(1, 2, 7)
    graph.AddUndirectedEdge(1, 3, 1)
    graph.AddUndirectedEdge(2, 4, 3)
    graph.AddUndirectedEdge(3, 4, 2)
    
    source := 0
    fmt.Printf("Поиск кратчайших путей от вершины %d\n\n", source)
    
    // Запускаем Dijkstra
    result := graph.Dijkstra(source)
    
    // Выводим кратчайшие расстояния
    fmt.Println("Кратчайшие расстояния:")
    for v := 0; v < graph.Vertices; v++ {
        if result.Distances[v] == math.Inf(1) {
            fmt.Printf("Вершина %d: недостижима\n", v)
        } else {
            fmt.Printf("Вершина %d: расстояние = %.1f\n", v, result.Distances[v])
        }
    }
    
    // Выводим пути
    fmt.Println("\nПути:")
    for target := 0; target < graph.Vertices; target++ {
        path := result.GetPath(target)
        if path != nil {
            fmt.Printf("%d → %d: ", source, target)
            for i, v := range path {
                if i > 0 {
                    fmt.Print(" → ")
                }
                fmt.Print(v)
            }
            fmt.Printf(" (расстояние: %.1f)\n", result.Distances[target])
        }
    }
}
```

---

### 23. K-Nearest Neighbors (KNN) - метод ближайших соседей

**Ответ:**
K-Nearest Neighbors (KNN) — это простой алгоритм машинного обучения для классификации и регрессии, основанный на принципе "похожие объекты находятся рядом": для классификации нового объекта находятся K ближайших соседей в обучающей выборке (по метрике расстояния, обычно евклидово), и новый объект получает класс большинства среди них (voting), а для регрессии — среднее значение целевой переменной соседей. KNN — это lazy learning (ленивое обучение) алгоритм: нет явной фазы обучения модели, все вычисления происходят во время предсказания, что делает его медленным для больших датасетов, но простым в реализации и обновлении (просто добавить новые примеры).

Ключевой параметр K (количество соседей): маленький K (1-3) делает модель чувствительной к шуму и выбросам (overfitting), большой K сглаживает границы классов но может смешивать разные классы (underfitting), типичные значения K=5-10 с нечетным K для избежания ничьих. Для ускорения поиска соседей используются пространственные индексы: KD-Tree для низких размерностей (d<20), Ball Tree для средних размерностей, LSH (Locality-Sensitive Hashing) для высоких размерностей. В моем рекомендательном движке я использовал KNN для поиска похожих товаров: векторизовал товары через embeddings (128 измерений), построил Ball Tree индекс, что позволило находить 10 ближайших соседей за ~5ms вместо ~200ms при линейном поиске по 50k товаров.

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "math"
    "sort"
)

// ========== ТИПЫ ДАННЫХ ==========

type Point struct {
    Features []float64
    Label    string  // Для классификации
    Value    float64 // Для регрессии
}

type Neighbor struct {
    Point    *Point
    Distance float64
}

type KNN struct {
    K           int
    TrainData   []*Point
    DistanceFunc func(a, b []float64) float64
}

// ========== МЕТРИКИ РАССТОЯНИЯ ==========

// Евклидово расстояние
func euclideanDistance(a, b []float64) float64 {
    if len(a) != len(b) {
        panic("vectors must have same length")
    }
    
    sum := 0.0
    for i := range a {
        diff := a[i] - b[i]
        sum += diff * diff
    }
    return math.Sqrt(sum)
}

// Манхэттенское расстояние (L1)
func manhattanDistance(a, b []float64) float64 {
    if len(a) != len(b) {
        panic("vectors must have same length")
    }
    
    sum := 0.0
    for i := range a {
        sum += math.Abs(a[i] - b[i])
    }
    return sum
}

// Косинусное расстояние (для текстов, embeddings)
func cosineDistance(a, b []float64) float64 {
    if len(a) != len(b) {
        panic("vectors must have same length")
    }
    
    dotProduct := 0.0
    normA := 0.0
    normB := 0.0
    
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    
    if normA == 0 || normB == 0 {
        return 1.0
    }
    
    similarity := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
    return 1.0 - similarity // Расстояние = 1 - similarity
}

// ========== KNN АЛГОРИТМ ==========

func NewKNN(k int, distanceFunc func(a, b []float64) float64) *KNN {
    if distanceFunc == nil {
        distanceFunc = euclideanDistance
    }
    
    return &KNN{
        K:            k,
        TrainData:    make([]*Point, 0),
        DistanceFunc: distanceFunc,
    }
}

// Обучение (просто сохраняем данные)
func (knn *KNN) Fit(data []*Point) {
    knn.TrainData = data
}

// Поиск K ближайших соседей
func (knn *KNN) FindNeighbors(query []float64) []Neighbor {
    // Вычисляем расстояния до всех точек
    neighbors := make([]Neighbor, len(knn.TrainData))
    for i, point := range knn.TrainData {
        distance := knn.DistanceFunc(query, point.Features)
        neighbors[i] = Neighbor{
            Point:    point,
            Distance: distance,
        }
    }
    
    // Сортируем по расстоянию
    sort.Slice(neighbors, func(i, j int) bool {
        return neighbors[i].Distance < neighbors[j].Distance
    })
    
    // Берем первые K
    k := knn.K
    if k > len(neighbors) {
        k = len(neighbors)
    }
    
    return neighbors[:k]
}

// Классификация (voting)
func (knn *KNN) Predict(query []float64) string {
    neighbors := knn.FindNeighbors(query)
    
    // Подсчитываем голоса
    votes := make(map[string]int)
    for _, neighbor := range neighbors {
        label := neighbor.Point.Label
        votes[label]++
    }
    
    // Находим класс с максимальным количеством голосов
    maxVotes := 0
    predictedLabel := ""
    for label, count := range votes {
        if count > maxVotes {
            maxVotes = count
            predictedLabel = label
        }
    }
    
    return predictedLabel
}

// Регрессия (среднее значение)
func (knn *KNN) PredictValue(query []float64) float64 {
    neighbors := knn.FindNeighbors(query)
    
    sum := 0.0
    for _, neighbor := range neighbors {
        sum += neighbor.Point.Value
    }
    
    return sum / float64(len(neighbors))
}

// Взвешенная классификация (closer neighbors have more weight)
func (knn *KNN) PredictWeighted(query []float64) string {
    neighbors := knn.FindNeighbors(query)
    
    // Подсчитываем взвешенные голоса
    votes := make(map[string]float64)
    for _, neighbor := range neighbors {
        label := neighbor.Point.Label
        // Вес = 1 / distance (closer = higher weight)
        weight := 1.0 / (neighbor.Distance + 1e-10) // избегаем деления на 0
        votes[label] += weight
    }
    
    // Находим класс с максимальным весом
    maxWeight := 0.0
    predictedLabel := ""
    for label, weight := range votes {
        if weight > maxWeight {
            maxWeight = weight
            predictedLabel = label
        }
    }
    
    return predictedLabel
}

// ========== ПРИМЕР: КЛАССИФИКАЦИЯ ИРИСОВ ==========

func irisExample() {
    fmt.Println("=== KNN для классификации ирисов ===\n")
    
    // Обучающие данные (sepal_length, sepal_width, petal_length, petal_width)
    trainData := []*Point{
        {Features: []float64{5.1, 3.5, 1.4, 0.2}, Label: "setosa"},
        {Features: []float64{4.9, 3.0, 1.4, 0.2}, Label: "setosa"},
        {Features: []float64{4.7, 3.2, 1.3, 0.2}, Label: "setosa"},
        {Features: []float64{7.0, 3.2, 4.7, 1.4}, Label: "versicolor"},
        {Features: []float64{6.4, 3.2, 4.5, 1.5}, Label: "versicolor"},
        {Features: []float64{6.9, 3.1, 4.9, 1.5}, Label: "versicolor"},
        {Features: []float64{6.3, 3.3, 6.0, 2.5}, Label: "virginica"},
        {Features: []float64{5.8, 2.7, 5.1, 1.9}, Label: "virginica"},
        {Features: []float64{7.1, 3.0, 5.9, 2.1}, Label: "virginica"},
    }
    
    // Создаем KNN классификатор
    knn := NewKNN(3, euclideanDistance)
    knn.Fit(trainData)
    
    // Тестовые примеры
    tests := []struct {
        features []float64
        expected string
    }{
        {[]float64{5.0, 3.6, 1.4, 0.2}, "setosa"},
        {[]float64{6.5, 3.0, 4.8, 1.4}, "versicolor"},
        {[]float64{6.2, 3.4, 5.4, 2.3}, "virginica"},
    }
    
    for i, test := range tests {
        predicted := knn.Predict(test.features)
        neighbors := knn.FindNeighbors(test.features)
        
        fmt.Printf("Тест %d:\n", i+1)
        fmt.Printf("  Признаки: %v\n", test.features)
        fmt.Printf("  Предсказан: %s\n", predicted)
        fmt.Printf("  Ожидается: %s\n", test.expected)
        fmt.Printf("  Ближайшие соседи:\n")
        for j, n := range neighbors {
            fmt.Printf("    %d. %s (расстояние: %.3f)\n", j+1, n.Point.Label, n.Distance)
        }
        fmt.Println()
    }
}

// ========== ПРИМЕР: РЕГРЕССИЯ ==========

func regressionExample() {
    fmt.Println("=== KNN для регрессии (предсказание цены дома) ===\n")
    
    // Обучающие данные (площадь, комнаты, цена)
    trainData := []*Point{
        {Features: []float64{50, 1}, Value: 100000},
        {Features: []float64{70, 2}, Value: 150000},
        {Features: []float64{90, 2}, Value: 180000},
        {Features: []float64{100, 3}, Value: 220000},
        {Features: []float64{120, 3}, Value: 250000},
        {Features: []float64{150, 4}, Value: 300000},
    }
    
    knn := NewKNN(3, euclideanDistance)
    knn.Fit(trainData)
    
    // Предсказание для нового дома
    query := []float64{85, 2} // 85 кв.м, 2 комнаты
    predicted := knn.PredictValue(query)
    neighbors := knn.FindNeighbors(query)
    
    fmt.Printf("Дом: площадь=%.0f кв.м, комнат=%.0f\n", query[0], query[1])
    fmt.Printf("Предсказанная цена: $%.0f\n\n", predicted)
    fmt.Println("Ближайшие соседи:")
    for i, n := range neighbors {
        fmt.Printf("%d. Площадь=%.0f, Комнат=%.0f, Цена=$%.0f (расстояние: %.2f)\n",
            i+1, n.Point.Features[0], n.Point.Features[1], n.Point.Value, n.Distance)
    }
}

func main() {
    irisExample()
    fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
    regressionExample()
}
```

---

## 🔷 ФАЗА 3: Observability и Мониторинг

### 24. SLI, SLO, SLA - метрики надежности сервиса

**Ответ:**
SLI (Service Level Indicator) — это количественная метрика, измеряющая определенный аспект качества сервиса, например uptime (доступность), latency (задержка), error rate (частота ошибок), throughput (пропускная способность); SLI всегда измеряется как доля "хороших" событий (например, 99.5% запросов выполнились за <100ms). SLO (Service Level Objective) — это целевое значение SLI, определяющее приемлемый уровень качества сервиса для бизнеса, например "99.9% запросов должны выполняться за <200ms" или "99.95% uptime за месяц", и SLO формирует error budget (бюджет ошибок = 100% - SLO = допустимое количество сбоев до нарушения SLO). SLA (Service Level Agreement) — это контракт с клиентами с юридическими и финансовыми последствиями при нарушении (штрафы, компенсации), всегда менее строгий чем SLO (например, SLA = 99.5%, SLO = 99.9%), чтобы был буфер для внутренних incident.

SLI выбираются исходя из user experience: для API критичны latency и availability (пользователь хочет быстрый ответ без ошибок), для batch processing — throughput и freshness (данные должны обрабатываться вовремя), для storage — durability и availability (данные не должны теряться). Error budget используется для баланса между скоростью разработки и надежностью: если error budget исчерпан (SLO нарушен), прекращается деплой новых фичей и фокус смещается на reliability (bagfix, refactoring, добавление мониторинга). В моем SaaS-проекте я определил SLI/SLO: API latency p99 < 500ms (SLO 99%), uptime > 99.9% за месяц, error rate < 0.1%, и настроил мониторинг в Prometheus с алертами при приближении к error budget (alerting rule: если за 1 час error rate > 0.05%, значит burn rate слишком высокий и error budget закончится раньше месяца).

**Пример SLI/SLO:**

```
╔═══════════════════════════════════════════════════════════════════╗
║ ПРИМЕРЫ SLI/SLO                                                   ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║ 📊 AVAILABILITY (Доступность)                                     ║
║   SLI: % успешных запросов (HTTP 200-299, 300-399)               ║
║   SLO: 99.9% за rolling window 30 дней                            ║
║   Error Budget: 0.1% = 43 минуты downtime в месяц                ║
║   Измерение: (successful_requests / total_requests) * 100        ║
║                                                                   ║
║ ⏱️ LATENCY (Задержка)                                              ║
║   SLI: % запросов быстрее threshold                               ║
║   SLO: 99% запросов < 200ms, 99.9% < 500ms                        ║
║   Error Budget: 1% slow requests = 43K из 4.3M req/month         ║
║   Измерение: histogram_quantile(0.99, request_duration)          ║
║                                                                   ║
║ 🔥 ERROR RATE (Частота ошибок)                                    ║
║   SLI: % запросов без ошибок (HTTP 200-499, exclude 5xx)         ║
║   SLO: 99.95% запросов без server errors                          ║
║   Error Budget: 0.05% = 2.15K errors из 4.3M req/month           ║
║   Измерение: 1 - (http_errors_5xx / total_requests)              ║
║                                                                   ║
║ 📦 THROUGHPUT (Пропускная способность)                            ║
║   SLI: % событий обработанных вовремя                             ║
║   SLO: 99% сообщений обрабатываются за < 5 секунд                 ║
║   Измерение: (processed_within_5s / total_processed)             ║
║                                                                   ║
║ 💾 DURABILITY (Долговечность данных)                               ║
║   SLI: % объектов не потерянных                                   ║
║   SLO: 99.999999999% (11 девяток) за год                          ║
║   Error Budget: 1 объект из 100 миллиардов может быть потерян    ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

ERROR BUDGET BURN RATE:
• Если SLO = 99.9% за 30 дней → error budget = 43 минуты
• При error rate 1% весь budget сгорит за 72 минуты (CRITICAL!)
• При error rate 0.1% budget сгорит за 12 часов (WARNING)
• При error rate 0.01% budget сгорит ровно за 30 дней (OK)

ALERTING STRATEGY:
• 2% burn rate за 1 час → Page (критично)
• 5% burn rate за 6 часов → Ticket (нужен fix)
• 10% burn rate за 3 дня → Review (оптимизация)
```

**Prometheus запросы для SLI:**

```yaml
# Availability SLI
- record: sli:availability:ratio
  expr: |
    sum(rate(http_requests_total{status=~"2..|3.."}[5m]))
    /
    sum(rate(http_requests_total[5m]))

# Latency SLI (% requests < 200ms)
- record: sli:latency_200ms:ratio
  expr: |
    histogram_quantile(0.99,
      rate(http_request_duration_seconds_bucket[5m])
    ) < 0.2

# Error Rate SLI
- record: sli:error_rate:ratio
  expr: |
    1 - (
      sum(rate(http_requests_total{status=~"5.."}[5m]))
      /
      sum(rate(http_requests_total[5m]))
    )

# Error Budget оставшийся (за 30 дней)
- record: slo:error_budget:remaining
  expr: |
    1 - (
      (1 - sli:availability:ratio) / (1 - 0.999)
    )
```

---

### 25. WebSocket - двунаправленная коммуникация

**Ответ:**
WebSocket — это протокол полнодуплексной (full-duplex) двунаправленной связи поверх одного TCP-соединения, позволяющий серверу отправлять данные клиенту без явного запроса (server push), в отличие от HTTP request-response модели; соединение устанавливается через HTTP Upgrade handshake (клиент отправляет `Connection: Upgrade, Upgrade: websocket`), после чего переключается на WebSocket protocol (frame-based, не текстовый как HTTP). WebSocket идеален для real-time приложений: чаты, live обновления (биржевые котировки, спортивные счета), collaborative editing (Google Docs), multiplayer игры, push-уведомления, где latency критична и polling неэффективен (HTTP long-polling создает новое соединение на каждый запрос, Server-Sent Events (SSE) однонаправленны).

Масштабирование WebSocket серверов сложнее чем stateless HTTP API: нужен sticky sessions (клиент всегда подключается к одному серверу) или pub/sub через Redis/Kafka для синхронизации сообщений между серверами (клиент A на сервере 1 отправляет сообщение клиенту B на сервере 2 → сервер 1 публикует в Redis → сервер 2 получает и отправляет клиенту B). Нужен graceful shutdown: при деплое новой версии сервера нельзя просто убить connections, сначала закрывается accept новых соединений, затем отправляется close frame существующим клиентам с reconnect hint, и дается время на переподключение. В моем проекте real-time аналитики я реализовал WebSocket сервер с pub/sub через Redis: каждый сервер подписан на каналы событий, при получении события из Redis пробрасывает его всем подключенным WebSocket клиентам, что позволило масштабировать до 10k одновременных соединений на 3 серверах с latency <50ms.

**Пример на Golang:**

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/go-redis/redis/v8"
)

// ========== WEBSOCKET ТИПЫ ==========

type Message struct {
    Type      string      `json:"type"`
    Channel   string      `json:"channel"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}

type Client struct {
    ID       string
    Conn     *websocket.Conn
    Send     chan *Message
    Hub      *Hub
    Channels map[string]bool // подписки клиента
    mu       sync.RWMutex
}

// ========== HUB (центральный менеджер соединений) ==========

type Hub struct {
    clients    map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
    redis      *redis.Client
    mu         sync.RWMutex
}

func NewHub(redisAddr string) *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *Message, 256),
        redis: redis.NewClient(&redis.Options{
            Addr: redisAddr,
        }),
    }
}

func (h *Hub) Run(ctx context.Context) {
    // Подписка на Redis pub/sub для координации между серверами
    pubsub := h.redis.Subscribe(ctx, "websocket:broadcast")
    defer pubsub.Close()

    go func() {
        for {
            msg, err := pubsub.ReceiveMessage(ctx)
            if err != nil {
                return
            }

            var message Message
            if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
                continue
            }

            // Отправляем всем локальным клиентам
            h.broadcastToClients(&message)
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            log.Printf("Client %s connected. Total: %d", client.ID, len(h.clients))

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.Send)
            }
            h.mu.Unlock()
            log.Printf("Client %s disconnected. Total: %d", client.ID, len(h.clients))

        case message := <-h.broadcast:
            // Публикуем в Redis для других серверов
            msgJSON, _ := json.Marshal(message)
            h.redis.Publish(ctx, "websocket:broadcast", msgJSON)

            // Отправляем локальным клиентам
            h.broadcastToClients(message)
        }
    }
}

func (h *Hub) broadcastToClients(message *Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    for client := range h.clients {
        // Проверяем подписку клиента на канал
        client.mu.RLock()
        subscribed := client.Channels[message.Channel]
        client.mu.RUnlock()

        if subscribed || message.Channel == "" {
            select {
            case client.Send <- message:
            default:
                // Канал клиента заполнен, закрываем соединение
                close(client.Send)
                delete(h.clients, client)
            }
        }
    }
}

// ========== CLIENT ОБРАБОТКА ==========

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // В продакшене проверяйте origin!
    },
}

func (c *Client) ReadPump() {
    defer func() {
        c.Hub.unregister <- c
        c.Conn.Close()
    }()

    c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    for {
        var msg Message
        err := c.Conn.ReadJSON(&msg)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        // Обрабатываем команды клиента
        switch msg.Type {
        case "subscribe":
            c.mu.Lock()
            c.Channels[msg.Channel] = true
            c.mu.Unlock()
            log.Printf("Client %s subscribed to %s", c.ID, msg.Channel)

        case "unsubscribe":
            c.mu.Lock()
            delete(c.Channels, msg.Channel)
            c.mu.Unlock()
            log.Printf("Client %s unsubscribed from %s", c.ID, msg.Channel)

        case "message":
            // Клиент отправил сообщение → broadcast другим клиентам
            msg.Timestamp = time.Now()
            c.Hub.broadcast <- &msg
        }
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                // Hub закрыл канал
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := c.Conn.WriteJSON(message); err != nil {
                return
            }

        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// ========== HTTP HANDLERS ==========

func serveWs(hub *Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Println(err)
            return
        }

        client := &Client{
            ID:       fmt.Sprintf("client_%d", time.Now().UnixNano()),
            Conn:     conn,
            Send:     make(chan *Message, 256),
            Hub:      hub,
            Channels: make(map[string]bool),
        }

        hub.register <- client

        go client.WritePump()
        go client.ReadPump()
    }
}

// Endpoint для публикации сообщений через REST API
func publishHandler(hub *Hub) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var msg Message
        if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        msg.Timestamp = time.Now()
        hub.broadcast <- &msg

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
    }
}

// ========== ГЛАВНЫЙ СЕРВЕР ==========

func main() {
    hub := NewHub("localhost:6379")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go hub.Run(ctx)

    http.HandleFunc("/ws", serveWs(hub))
    http.HandleFunc("/api/publish", publishHandler(hub))

    // Статический файл для теста
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(testHTML))
    })

    log.Println("WebSocket server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// ========== HTML CLIENT ДЛЯ ТЕСТИРОВАНИЯ ==========

const testHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <h1>WebSocket Real-Time Chat</h1>
    <div>
        <input id="channel" placeholder="Channel name" value="general">
        <button onclick="subscribe()">Subscribe</button>
        <button onclick="unsubscribe()">Unsubscribe</button>
    </div>
    <div>
        <input id="message" placeholder="Message">
        <button onclick="sendMessage()">Send</button>
    </div>
    <div id="messages" style="border: 1px solid #ccc; height: 300px; overflow-y: auto; margin-top: 10px; padding: 10px;">
    </div>

    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        
        ws.onopen = () => {
            console.log('Connected to WebSocket');
            addMessage('System', 'Connected to server');
        };
        
        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            addMessage(msg.channel, JSON.stringify(msg.data));
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        ws.onclose = () => {
            console.log('Disconnected from WebSocket');
            addMessage('System', 'Disconnected from server');
        };
        
        function subscribe() {
            const channel = document.getElementById('channel').value;
            ws.send(JSON.stringify({
                type: 'subscribe',
                channel: channel
            }));
            addMessage('System', 'Subscribed to ' + channel);
        }
        
        function unsubscribe() {
            const channel = document.getElementById('channel').value;
            ws.send(JSON.stringify({
                type: 'unsubscribe',
                channel: channel
            }));
            addMessage('System', 'Unsubscribed from ' + channel);
        }
        
        function sendMessage() {
            const channel = document.getElementById('channel').value;
            const message = document.getElementById('message').value;
            ws.send(JSON.stringify({
                type: 'message',
                channel: channel,
                data: message
            }));
            document.getElementById('message').value = '';
        }
        
        function addMessage(channel, text) {
            const div = document.getElementById('messages');
            const time = new Date().toLocaleTimeString();
            div.innerHTML += '<div><strong>[' + time + '] ' + channel + ':</strong> ' + text + '</div>';
            div.scrollTop = div.scrollHeight;
        }
    </script>
</body>
</html>
`
```

**Сравнение real-time технологий:**

```
╔══════════════════╦══════════════╦══════════════╦══════════════════╗
║ Технология       ║ Направление  ║ Overhead     ║ Применение       ║
╠══════════════════╬══════════════╬══════════════╬══════════════════╣
║ HTTP Polling     ║ Client→Server║ Высокий      ║ Не рекомендуется ║
║ Long Polling     ║ Client→Server║ Средний      ║ Legacy браузеры  ║
║ Server-Sent      ║ Server→Client║ Низкий       ║ Уведомления      ║
║ Events (SSE)     ║              ║              ║                  ║
║ WebSocket        ║ Bidirectional║ Очень низкий ║ Real-time apps   ║
║ gRPC Streaming   ║ Bidirectional║ Низкий       ║ Service-to-      ║
║                  ║              ║              ║ service          ║
╚══════════════════╩══════════════╩══════════════╩══════════════════╝
```

---


### 26. Traveling Salesman Problem (TSP) - задача коммивояжера

**Ответ:**
Traveling Salesman Problem (TSP) — это классическая NP-трудная задача комбинаторной оптимизации: найти кратчайший маршрут, проходящий через N городов ровно один раз и возвращающийся в начальную точку, минимизируя общую стоимость (расстояние) пути. TSP имеет (N-1)!/2 возможных маршрутов для N городов, что делает полный перебор невозможным даже для небольших N (для 20 городов это ~60 квинтиллионов вариантов), поэтому используются приближенные алгоритмы и эвристики: жадный алгоритм (greedy nearest neighbor), динамическое программирование (Held-Karp O(N²*2^N)), генетические алгоритмы, имитация отжига (simulated annealing), метод ветвей и границ (branch and bound). Жадная эвристика nearest neighbor: начинаем с произвольного города, на каждом шаге выбираем ближайший непосещенный город, гарантирует решение не более чем в 2 раза хуже оптимального, работает за O(N²).

TSP применяется не только для маршрутизации курьеров и транспорта, но и для оптимизации сверления печатных плат (drill path optimization), планирования телескопных наблюдений (минимизация времени перехода между объектами), DNA секвенирования (упорядочивание фрагментов). Для практических задач часто достаточно "хорошего" решения (within 10% от оптимального), что достижимо за разумное время через эвристики. В моем проекте логистики я решал вариацию TSP для оптимизации маршрутов доставки: использовал greedy nearest neighbor + 2-opt улучшение (swap ребер для уменьшения пересечений), что для 50 точек доставки сократило маршрут на 22% относительно наивного порядка и работало за ~100ms, достаточно быстро для динамического пересчета при добавлении новых заказов.

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "math"
    "math/rand"
)

// ========== СТРУКТУРЫ ДАННЫХ ==========

type City struct {
    ID   int
    X, Y float64
}

type Tour struct {
    Cities   []int
    Distance float64
}

// ========== РАССТОЯНИЕ МЕЖДУ ГОРОДАМИ ==========

func distance(c1, c2 City) float64 {
    dx := c1.X - c2.X
    dy := c1.Y - c2.Y
    return math.Sqrt(dx*dx + dy*dy)
}

func tourDistance(cities []City, tour []int) float64 {
    dist := 0.0
    for i := 0; i < len(tour)-1; i++ {
        dist += distance(cities[tour[i]], cities[tour[i+1]])
    }
    // Возврат в начальный город
    dist += distance(cities[tour[len(tour)-1]], cities[tour[0]])
    return dist
}

// ========== GREEDY NEAREST NEIGHBOR ==========

func nearestNeighbor(cities []City) Tour {
    n := len(cities)
    visited := make([]bool, n)
    tour := make([]int, 0, n)
    
    // Начинаем с города 0
    current := 0
    tour = append(tour, current)
    visited[current] = true
    
    for len(tour) < n {
        nearest := -1
        minDist := math.Inf(1)
        
        // Находим ближайший непосещенный город
        for i := 0; i < n; i++ {
            if !visited[i] {
                dist := distance(cities[current], cities[i])
                if dist < minDist {
                    minDist = dist
                    nearest = i
                }
            }
        }
        
        tour = append(tour, nearest)
        visited[nearest] = true
        current = nearest
    }
    
    return Tour{
        Cities:   tour,
        Distance: tourDistance(cities, tour),
    }
}

// ========== 2-OPT УЛУЧШЕНИЕ ==========

func twoOptSwap(tour []int, i, k int) []int {
    newTour := make([]int, len(tour))
    copy(newTour, tour[:i])
    
    // Reverse segment [i, k]
    for j := i; j <= k; j++ {
        newTour[j] = tour[k-(j-i)]
    }
    
    copy(newTour[k+1:], tour[k+1:])
    return newTour
}

func twoOpt(cities []City, initialTour []int, maxIterations int) Tour {
    tour := make([]int, len(initialTour))
    copy(tour, initialTour)
    bestDistance := tourDistance(cities, tour)
    improved := true
    iterations := 0
    
    for improved && iterations < maxIterations {
        improved = false
        iterations++
        
        for i := 0; i < len(tour)-1; i++ {
            for k := i + 1; k < len(tour); k++ {
                newTour := twoOptSwap(tour, i, k)
                newDistance := tourDistance(cities, newTour)
                
                if newDistance < bestDistance {
                    tour = newTour
                    bestDistance = newDistance
                    improved = true
                }
            }
        }
    }
    
    return Tour{
        Cities:   tour,
        Distance: bestDistance,
    }
}

// ========== SIMULATED ANNEALING ==========

func simulatedAnnealing(cities []City, initialTemp, coolingRate float64, iterations int) Tour {
    n := len(cities)
    
    // Начальный случайный тур
    tour := make([]int, n)
    for i := range tour {
        tour[i] = i
    }
    rand.Shuffle(n, func(i, j int) {
        tour[i], tour[j] = tour[j], tour[i]
    })
    
    currentDistance := tourDistance(cities, tour)
    bestTour := make([]int, n)
    copy(bestTour, tour)
    bestDistance := currentDistance
    
    temp := initialTemp
    
    for iter := 0; iter < iterations; iter++ {
        // Случайный swap двух городов
        i := rand.Intn(n)
        j := rand.Intn(n)
        
        // Создаем новый тур
        newTour := make([]int, n)
        copy(newTour, tour)
        newTour[i], newTour[j] = newTour[j], newTour[i]
        
        newDistance := tourDistance(cities, newTour)
        delta := newDistance - currentDistance
        
        // Принимаем новое решение если оно лучше ИЛИ с вероятностью e^(-delta/temp)
        if delta < 0 || rand.Float64() < math.Exp(-delta/temp) {
            tour = newTour
            currentDistance = newDistance
            
            if currentDistance < bestDistance {
                copy(bestTour, tour)
                bestDistance = currentDistance
            }
        }
        
        // Охлаждение
        temp *= coolingRate
    }
    
    return Tour{
        Cities:   bestTour,
        Distance: bestDistance,
    }
}

// ========== GENETIC ALGORITHM ==========

type Population struct {
    Tours      []Tour
    BestTour   Tour
    Generation int
}

func createRandomTour(n int) []int {
    tour := make([]int, n)
    for i := range tour {
        tour[i] = i
    }
    rand.Shuffle(n, func(i, j int) {
        tour[i], tour[j] = tour[j], tour[i]
    })
    return tour
}

func crossover(parent1, parent2 []int) []int {
    n := len(parent1)
    child := make([]int, n)
    
    // Order crossover (OX)
    start := rand.Intn(n)
    end := rand.Intn(n)
    if start > end {
        start, end = end, start
    }
    
    // Копируем сегмент из parent1
    used := make(map[int]bool)
    for i := start; i <= end; i++ {
        child[i] = parent1[i]
        used[parent1[i]] = true
    }
    
    // Заполняем остальное из parent2
    childIdx := 0
    for i := 0; i < n; i++ {
        if childIdx == start {
            childIdx = end + 1
        }
        if childIdx >= n {
            break
        }
        
        if !used[parent2[i]] {
            child[childIdx] = parent2[i]
            childIdx++
        }
    }
    
    return child
}

func mutate(tour []int, mutationRate float64) []int {
    if rand.Float64() < mutationRate {
        i := rand.Intn(len(tour))
        j := rand.Intn(len(tour))
        tour[i], tour[j] = tour[j], tour[i]
    }
    return tour
}

func geneticAlgorithm(cities []City, popSize, generations int, mutationRate float64) Tour {
    n := len(cities)
    
    // Инициализация популяции
    population := make([]Tour, popSize)
    for i := range population {
        tour := createRandomTour(n)
        population[i] = Tour{
            Cities:   tour,
            Distance: tourDistance(cities, tour),
        }
    }
    
    bestTour := population[0]
    
    for gen := 0; gen < generations; gen++ {
        // Сортируем по fitness (короче = лучше)
        for i := 0; i < popSize-1; i++ {
            for j := i + 1; j < popSize; j++ {
                if population[j].Distance < population[i].Distance {
                    population[i], population[j] = population[j], population[i]
                }
            }
        }
        
        // Обновляем лучший тур
        if population[0].Distance < bestTour.Distance {
            bestTour = population[0]
        }
        
        // Создаем новое поколение
        newPopulation := make([]Tour, popSize)
        
        // Элитизм: сохраняем 10% лучших
        eliteSize := popSize / 10
        copy(newPopulation[:eliteSize], population[:eliteSize])
        
        // Crossover + mutation
        for i := eliteSize; i < popSize; i++ {
            parent1 := population[rand.Intn(popSize/2)]  // Выбираем из лучшей половины
            parent2 := population[rand.Intn(popSize/2)]
            
            child := crossover(parent1.Cities, parent2.Cities)
            child = mutate(child, mutationRate)
            
            newPopulation[i] = Tour{
                Cities:   child,
                Distance: tourDistance(cities, child),
            }
        }
        
        population = newPopulation
    }
    
    return bestTour
}

// ========== ПРИМЕРЫ И СРАВНЕНИЕ ==========

func main() {
    // Генерируем случайные города
    rand.Seed(42)
    n := 20
    cities := make([]City, n)
    for i := range cities {
        cities[i] = City{
            ID: i,
            X:  rand.Float64() * 100,
            Y:  rand.Float64() * 100,
        }
    }
    
    fmt.Printf("Решение TSP для %d городов\n\n", n)
    
    // 1. Greedy Nearest Neighbor
    fmt.Println("=== Greedy Nearest Neighbor ===")
    greedyTour := nearestNeighbor(cities)
    fmt.Printf("Расстояние: %.2f\n", greedyTour.Distance)
    fmt.Printf("Маршрут: %v\n\n", greedyTour.Cities[:5]) // показываем первые 5 городов
    
    // 2. 2-Opt улучшение greedy решения
    fmt.Println("=== 2-Opt улучшение ===")
    twoOptTour := twoOpt(cities, greedyTour.Cities, 100)
    fmt.Printf("Расстояние: %.2f (улучшение: %.1f%%)\n", 
        twoOptTour.Distance, 
        (greedyTour.Distance-twoOptTour.Distance)/greedyTour.Distance*100)
    
    // 3. Simulated Annealing
    fmt.Println("\n=== Simulated Annealing ===")
    saTour := simulatedAnnealing(cities, 10000, 0.995, 10000)
    fmt.Printf("Расстояние: %.2f\n", saTour.Distance)
    
    // 4. Genetic Algorithm
    fmt.Println("\n=== Genetic Algorithm ===")
    gaTour := geneticAlgorithm(cities, 100, 500, 0.01)
    fmt.Printf("Расстояние: %.2f\n", gaTour.Distance)
    
    // Сравнение
    fmt.Println("\n=== Сравнение алгоритмов ===")
    fmt.Printf("Greedy:             %.2f\n", greedyTour.Distance)
    fmt.Printf("2-Opt:              %.2f (%.1f%% лучше)\n", 
        twoOptTour.Distance, 
        (greedyTour.Distance-twoOptTour.Distance)/greedyTour.Distance*100)
    fmt.Printf("Simulated Annealing: %.2f\n", saTour.Distance)
    fmt.Printf("Genetic Algorithm:   %.2f\n", gaTour.Distance)
}
```

---

### 27. DBSCAN - кластеризация на основе плотности

**Ответ:**
DBSCAN (Density-Based Spatial Clustering of Applications with Noise) — это алгоритм кластеризации, группирующий точки по плотности: точки с большим количеством соседей в радиусе ε (epsilon) объединяются в кластеры, а изолированные точки помечаются как шум (outliers), в отличие от K-means, не требует задавать количество кластеров заранее и находит кластеры произвольной формы (не только выпуклые). DBSCAN использует два параметра: ε (epsilon) — радиус соседства, minPts — минимальное количество точек для формирования кластера; точка классифицируется как core point если имеет ≥ minPts соседей в радиусе ε, border point если не core но находится в радиусе ε от core point, noise point если не удовлетворяет ни одному условию. Алгоритм: для каждой непосещенной точки проверяем, является ли она core point (region query), если да — создаем новый кластер и расширяем его через все достижимые core points (density-connected), сложность O(N log N) с пространственным индексом (KD-Tree) или O(N²) без него.

DBSCAN превосходит K-means для данных с шумом, кластерами разной плотности и сложной геометрии (например, полумесяцы, кольца), но чувствителен к выбору ε и minPts (слишком маленький ε → много маленьких кластеров, слишком большой → все в одном кластере), плохо работает с кластерами сильно отличающейся плотности (используйте HDBSCAN в этом случае). Применения: обнаружение аномалий (outliers), географическая кластеризация (hotspots преступлений, эпидемий), сегментация изображений, анализ сетевого трафика. В моем проекте анализа логов я использовал DBSCAN для группировки похожих ошибок по векторным представлениям (embeddings) текстов ошибок: это автоматически выделило 15 кластеров типичных проблем и пометило ~5% уникальных ошибок как шум для ручного разбора, что сократило время анализа инцидентов на 60%.

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "math"
    "math/rand"
)

// ========== СТРУКТУРЫ ДАННЫХ ==========

type Point struct {
    ID       int
    X, Y     float64
    Cluster  int  // -1 = noise, 0 = unvisited, >0 = cluster ID
    Visited  bool
}

type DBSCAN struct {
    Epsilon float64
    MinPts  int
    Points  []*Point
}

// ========== DBSCAN АЛГОРИТМ ==========

func NewDBSCAN(epsilon float64, minPts int) *DBSCAN {
    return &DBSCAN{
        Epsilon: epsilon,
        MinPts:  minPts,
        Points:  make([]*Point, 0),
    }
}

func (db *DBSCAN) AddPoint(x, y float64) {
    point := &Point{
        ID:      len(db.Points),
        X:       x,
        Y:       y,
        Cluster: 0,
        Visited: false,
    }
    db.Points = append(db.Points, point)
}

func (db *DBSCAN) distance(p1, p2 *Point) float64 {
    dx := p1.X - p2.X
    dy := p1.Y - p2.Y
    return math.Sqrt(dx*dx + dy*dy)
}

// Находим соседей точки в радиусе epsilon
func (db *DBSCAN) regionQuery(point *Point) []*Point {
    neighbors := make([]*Point, 0)
    for _, p := range db.Points {
        if db.distance(point, p) <= db.Epsilon {
            neighbors = append(neighbors, p)
        }
    }
    return neighbors
}

// Расширяем кластер
func (db *DBSCAN) expandCluster(point *Point, neighbors []*Point, clusterID int) {
    point.Cluster = clusterID
    
    for i := 0; i < len(neighbors); i++ {
        neighbor := neighbors[i]
        
        if !neighbor.Visited {
            neighbor.Visited = true
            neighborNeighbors := db.Region Query(neighbor)
            
            // Если neighbor - core point, добавляем его соседей
            if len(neighborNeighbors) >= db.MinPts {
                neighbors = append(neighbors, neighborNeighbors...)
            }
        }
        
        // Добавляем в кластер если еще не в кластере
        if neighbor.Cluster == 0 {
            neighbor.Cluster = clusterID
        }
    }
}

// Запуск DBSCAN
func (db *DBSCAN) Fit() {
    clusterID := 0
    
    for _, point := range db.Points {
        if point.Visited {
            continue
        }
        
        point.Visited = true
        neighbors := db.regionQuery(point)
        
        if len(neighbors) < db.MinPts {
            // Noise point
            point.Cluster = -1
        } else {
            // Core point - создаем новый кластер
            clusterID++
            db.expandCluster(point, neighbors, clusterID)
        }
    }
}

// Получить статистику кластеризации
func (db *DBSCAN) GetStats() map[string]interface{} {
    clusterCounts := make(map[int]int)
    for _, point := range db.Points {
        clusterCounts[point.Cluster]++
    }
    
    numClusters := len(clusterCounts)
    if _, hasNoise := clusterCounts[-1]; hasNoise {
        numClusters-- // Не считаем шум как кластер
    }
    
    return map[string]interface{}{
        "num_points":   len(db.Points),
        "num_clusters": numClusters,
        "noise_points": clusterCounts[-1],
        "cluster_sizes": clusterCounts,
    }
}

// Визуализация (ASCII art)
func (db *DBSCAN) Visualize() {
    minX, maxX := math.Inf(1), math.Inf(-1)
    minY, maxY := math.Inf(1), math.Inf(-1)
    
    for _, p := range db.Points {
        minX = math.Min(minX, p.X)
        maxX = math.Max(maxX, p.X)
        minY = math.Min(minY, p.Y)
        maxY = math.Max(maxY, p.Y)
    }
    
    width, height := 60, 20
    grid := make([][]rune, height)
    for i := range grid {
        grid[i] = make([]rune, width)
        for j := range grid[i] {
            grid[i][j] = '.'
        }
    }
    
    symbols := []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J'}
    
    for _, p := range db.Points {
        x := int((p.X - minX) / (maxX - minX) * float64(width-1))
        y := int((p.Y - minY) / (maxY - minY) * float64(height-1))
        
        if p.Cluster == -1 {
            grid[y][x] = 'X' // Noise
        } else if p.Cluster > 0 && p.Cluster <= len(symbols) {
            grid[y][x] = symbols[p.Cluster-1]
        }
    }
    
    for i := range grid {
        fmt.Println(string(grid[i]))
    }
}

// ========== ПРИМЕРЫ ==========

func example1TwoClusters() {
    fmt.Println("=== Пример 1: Два четких кластера ===")
    
    db := NewDBSCAN(3.0, 4)
    
    // Кластер 1
    for i := 0; i < 20; i++ {
        x := rand.Float64()*10 + 10
        y := rand.Float64()*10 + 10
        db.AddPoint(x, y)
    }
    
    // Кластер 2
    for i := 0; i < 20; i++ {
        x := rand.Float64()*10 + 40
        y := rand.Float64()*10 + 40
        db.AddPoint(x, y)
    }
    
    // Шум
    for i := 0; i < 5; i++ {
        x := rand.Float64() * 60
        y := rand.Float64() * 60
        db.AddPoint(x, y)
    }
    
    db.Fit()
    
    stats := db.GetStats()
    fmt.Printf("Точек: %d\n", stats["num_points"])
    fmt.Printf("Кластеров: %d\n", stats["num_clusters"])
    fmt.Printf("Шума: %d\n\n", stats["noise_points"])
    
    db.Visualize()
}

func example2MoonShapes() {
    fmt.Println("\n=== Пример 2: Полумесяцы (K-means не справится) ===")
    
    db := NewDBSCAN(2.5, 5)
    
    // Верхний полумесяц
    for i := 0; i < 30; i++ {
        angle := math.Pi * float64(i) / 30
        x := math.Cos(angle)*20 + 30
        y := math.Sin(angle)*20 + 30
        db.AddPoint(x, y)
    }
    
    // Нижний полумесяц (перевернутый)
    for i := 0; i < 30; i++ {
        angle := -math.Pi * float64(i) / 30
        x := math.Cos(angle)*20 + 30
        y := math.Sin(angle)*20 + 20
        db.AddPoint(x, y)
    }
    
    db.Fit()
    
    stats := db.GetStats()
    fmt.Printf("Точек: %d\n", stats["num_points"])
    fmt.Printf("Кластеров: %d\n", stats["num_clusters"])
    fmt.Printf("Шума: %d\n\n", stats["noise_points"])
    
    db.Visualize()
}

func main() {
    rand.Seed(42)
    
    example1TwoClusters()
    example2MoonShapes()
    
    fmt.Println("\n=== Сравнение DBSCAN vs K-means ===")
    fmt.Println(`
DBSCAN:
  ✅ Находит кластеры произвольной формы
  ✅ Не нужно задавать количество кластеров
  ✅ Обрабатывает шум (outliers)
  ❌ Чувствителен к выбору epsilon и minPts
  ❌ Плохо с кластерами разной плотности

K-means:
  ✅ Быстрый и простой
  ✅ Гарантированная сходимость
  ❌ Нужно задавать K
  ❌ Только выпуклые кластеры
  ❌ Чувствителен к выбросам
    `)
}
```

---


## 🔷 ФАЗА 3: Экспертные темы - продолжение

### 28. Reflection в Go - интроспекция типов

**Ответ:**
Reflection (отражение) в Go — это механизм интроспекции и манипуляции типами и значениями во время выполнения через пакет `reflect`, позволяющий получать информацию о типах (reflect.Type), читать и изменять значения (reflect.Value), вызывать методы динамически, что критично для библиотек, работающих с произвольными типами: ORM (gorm), JSON/XML сериализация (encoding/json), dependency injection контейнеры, тестовые фреймворки. Основные функции: `reflect.TypeOf(v)` возвращает тип значения, `reflect.ValueOf(v)` возвращает значение с методами для чтения полей (`Field`, `Index`), изменения через pointer (`Elem`, `Set`), проверки типа (`Kind`, `Implements`). Reflection медленный (в 10-100 раз медленнее прямого доступа) из-за динамической диспетчеризации и проверок типов, используйте только когда нет альтернативы.

Важные паттерны reflection: для изменения значения нужен pointer (`reflect.ValueOf(&x).Elem().SetInt(42)`), проверка нулевых указателей (`v.IsNil()`), итерация по структурам (`t.NumField()`, `v.Field(i)`), чтение тегов (`field.Tag.Get("json")`). Reflection часто используется с `interface{}` для generic-кода до Go 1.18, но с появлением дженериков предпочтительнее использовать type parameters где возможно. В моем ORM-подобном фреймворке я использовал reflection для автоматического mapping между SQL rows и Go structs: парсил struct tags `db:"column_name"`, создавал prepared statements динамически, и заполнял поля через `reflect.Value.Set()`, что позволило писать `db.Query(&users, "SELECT * FROM users WHERE age > ?", 18)` вместо ручного scan'инга каждой колонки.

**Пример на Golang:**

```go
package main

import (
    "fmt"
    "reflect"
    "strings"
)

// ========== БАЗОВЫЙ REFLECTION ==========

type User struct {
    ID       int    `json:"id" db:"user_id" validate:"required"`
    Name     string `json:"name" db:"user_name" validate:"required,min=3"`
    Email    string `json:"email" db:"email" validate:"email"`
    Age      int    `json:"age,omitempty" db:"age"`
    IsActive bool   `json:"is_active" db:"is_active"`
}

func inspectType() {
    user := User{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30}
    
    // Получаем тип
    t := reflect.TypeOf(user)
    fmt.Printf("Type: %v\n", t)
    fmt.Printf("Name: %s\n", t.Name())
    fmt.Printf("Kind: %v\n", t.Kind())
    fmt.Printf("NumFields: %d\n\n", t.NumField())
    
    // Итерация по полям структуры
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fmt.Printf("Field %d:\n", i)
        fmt.Printf("  Name: %s\n", field.Name)
        fmt.Printf("  Type: %v\n", field.Type)
        fmt.Printf("  Tag json: %s\n", field.Tag.Get("json"))
        fmt.Printf("  Tag db: %s\n", field.Tag.Get("db"))
        fmt.Printf("  Tag validate: %s\n\n", field.Tag.Get("validate"))
    }
}

func inspectValue() {
    user := User{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30}
    
    v := reflect.ValueOf(user)
    t := v.Type()
    
    fmt.Println("Field values:")
    for i := 0; i < v.NumField(); i++ {
        fieldValue := v.Field(i)
        fieldType := t.Field(i)
        fmt.Printf("  %s (%v) = %v\n", fieldType.Name, fieldType.Type, fieldValue.Interface())
    }
}

// ========== ИЗМЕНЕНИЕ ЗНАЧЕНИЙ ==========

func modifyValue() {
    user := User{ID: 1, Name: "Alice"}
    fmt.Printf("Before: %+v\n", user)
    
    // Для изменения нужен pointer!
    v := reflect.ValueOf(&user).Elem()
    
    // Изменяем поле Name
    nameField := v.FieldByName("Name")
    if nameField.CanSet() {
        nameField.SetString("Bob")
    }
    
    // Изменяем Age
    ageField := v.FieldByName("Age")
    if ageField.CanSet() {
        ageField.SetInt(25)
    }
    
    fmt.Printf("After: %+v\n", user)
}

// ========== ДИНАМИЧЕСКИЙ ВЫЗОВ МЕТОДОВ ==========

type Calculator struct{}

func (c Calculator) Add(a, b int) int {
    return a + b
}

func (c Calculator) Multiply(a, b int) int {
    return a * b
}

func (c Calculator) Greet(name string) string {
    return "Hello, " + name
}

func callMethodDynamically() {
    calc := Calculator{}
    v := reflect.ValueOf(calc)
    
    // Вызов Add(5, 3)
    addMethod := v.MethodByName("Add")
    args := []reflect.Value{
        reflect.ValueOf(5),
        reflect.ValueOf(3),
    }
    result := addMethod.Call(args)
    fmt.Printf("Add(5, 3) = %v\n", result[0].Int())
    
    // Вызов Greet("Alice")
    greetMethod := v.MethodByName("Greet")
    args = []reflect.Value{reflect.ValueOf("Alice")}
    result = greetMethod.Call(args)
    fmt.Printf("Greet(\"Alice\") = %v\n", result[0].String())
}

// ========== STRUCT TO MAP КОНВЕРТЕР ==========

func structToMap(s interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    
    v := reflect.ValueOf(s)
    t := v.Type()
    
    // Если pointer, берем Elem
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = v.Type()
    }
    
    if v.Kind() != reflect.Struct {
        return result
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        value := v.Field(i)
        
        // Используем json tag как ключ, иначе имя поля
        key := field.Tag.Get("json")
        if key == "" {
            key = field.Name
        } else {
            // Убираем ",omitempty" и т.п.
            key = strings.Split(key, ",")[0]
        }
        
        result[key] = value.Interface()
    }
    
    return result
}

// ========== GENERIC VALIDATOR ==========

func validate(s interface{}) []string {
    errors := make([]string, 0)
    
    v := reflect.ValueOf(s)
    t := v.Type()
    
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = v.Type()
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        value := v.Field(i)
        
        validateTag := field.Tag.Get("validate")
        if validateTag == "" {
            continue
        }
        
        rules := strings.Split(validateTag, ",")
        for _, rule := range rules {
            switch rule {
            case "required":
                if value.IsZero() {
                    errors = append(errors, fmt.Sprintf("%s is required", field.Name))
                }
            case "email":
                if value.Kind() == reflect.String {
                    email := value.String()
                    if !strings.Contains(email, "@") {
                        errors = append(errors, fmt.Sprintf("%s must be a valid email", field.Name))
                    }
                }
            }
            
            if strings.HasPrefix(rule, "min=") {
                minLen := 0
                fmt.Sscanf(rule, "min=%d", &minLen)
                if value.Kind() == reflect.String && len(value.String()) < minLen {
                    errors = append(errors, fmt.Sprintf("%s must be at least %d characters", field.Name, minLen))
                }
            }
        }
    }
    
    return errors
}

// ========== ПРОСТОЙ ORM-LIKE MAPPER ==========

func mapRowToStruct(row map[string]interface{}, dest interface{}) error {
    v := reflect.ValueOf(dest)
    if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
        return fmt.Errorf("dest must be a pointer to struct")
    }
    
    v = v.Elem()
    t := v.Type()
    
    for i := 0; i < v.NumField(); i++ {
        field := t.Field(i)
        dbTag := field.Tag.Get("db")
        if dbTag == "" {
            dbTag = field.Name
        }
        
        // Ищем значение из row
        if rowValue, ok := row[dbTag]; ok {
            fieldValue := v.Field(i)
            if fieldValue.CanSet() {
                // Конвертируем тип если нужно
                rowVal := reflect.ValueOf(rowValue)
                if rowVal.Type().AssignableTo(fieldValue.Type()) {
                    fieldValue.Set(rowVal)
                } else {
                    // Попытка конвертации
                    if rowVal.Type().ConvertibleTo(fieldValue.Type()) {
                        fieldValue.Set(rowVal.Convert(fieldValue.Type()))
                    }
                }
            }
        }
    }
    
    return nil
}

// ========== MAIN ==========

func main() {
    fmt.Println("=== Inspect Type ===")
    inspectType()
    
    fmt.Println("\n=== Inspect Value ===")
    inspectValue()
    
    fmt.Println("\n=== Modify Value ===")
    modifyValue()
    
    fmt.Println("\n=== Call Method Dynamically ===")
    callMethodDynamically()
    
    fmt.Println("\n=== Struct to Map ===")
    user := User{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30}
    m := structToMap(user)
    fmt.Printf("Map: %+v\n", m)
    
    fmt.Println("\n=== Validator ===")
    invalidUser := User{} // Empty struct
    errors := validate(invalidUser)
    if len(errors) > 0 {
        fmt.Println("Validation errors:")
        for _, err := range errors {
            fmt.Printf("  - %s\n", err)
        }
    }
    
    validUser := User{ID: 1, Name: "Alice", Email: "alice@test.com"}
    errors = validate(validUser)
    if len(errors) == 0 {
        fmt.Println("Validation passed!")
    }
    
    fmt.Println("\n=== ORM-like Mapper ===")
    // Симуляция SQL row
    row := map[string]interface{}{
        "user_id":   123,
        "user_name": "Bob",
        "email":     "bob@test.com",
        "age":       25,
        "is_active": true,
    }
    
    var mappedUser User
    if err := mapRowToStruct(row, &mappedUser); err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Printf("Mapped user: %+v\n", mappedUser)
    }
}
```

**Когда использовать reflection:**

```
╔═══════════════════════════════════════════════════════════════════╗
║ КОГДА ИСПОЛЬЗОВАТЬ REFLECTION                                     ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║ ✅ ХОРОШИЕ ПРИМЕНЕНИЯ:                                             ║
║   • Сериализация/десериализация (JSON, XML, Protocol Buffers)    ║
║   • ORM маппинг (SQL rows → structs)                              ║
║   • Валидация на основе тегов                                     ║
║   • Dependency injection контейнеры                                ║
║   • Тестовые assertion библиотеки                                 ║
║   • Генерация кода на основе структур                             ║
║                                                                   ║
║ ❌ ПЛОХИЕ ПРИМЕНЕНИЯ:                                              ║
║   • Обычная бизнес-логика (используйте interfaces)                ║
║   • Когда можно использовать generics (Go 1.18+)                  ║
║   • В горячих циклах (performance critical code)                  ║
║   • Когда тип известен на этапе компиляции                        ║
║                                                                   ║
║ ⚠️  ПРОИЗВОДИТЕЛЬНОСТЬ:                                            ║
║   • Прямой доступ:    1x (базовая линия)                          ║
║   • Interface call:   ~1.2x медленнее                              ║
║   • Reflection:       10-100x медленнее                            ║
║   • Кэшируйте reflect.Type где возможно                           ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

### 29. OAuth 2.0 - делегированная авторизация

**Ответ:**
OAuth 2.0 — это протокол делегированной авторизации, позволяющий приложению получить ограниченный доступ к ресурсам пользователя на другом сервисе (Resource Server) без раскрытия учетных данных пользователя приложению, через выдачу access token'ов Authorization Server'ом после согласия пользователя. Основные роли: Resource Owner (пользователь), Client (приложение запрашивающее доступ), Authorization Server (выдает токены, например Google OAuth), Resource Server (API с защищенными ресурсами, например Google Drive API). OAuth 2.0 определяет 4 grant types (способа получения токена): Authorization Code (для веб-приложений с backend, самый безопасный), Implicit (устаревший, для SPA), Resource Owner Password Credentials (для доверенных приложений), Client Credentials (server-to-server без пользователя).

Authorization Code Flow с PKCE (Proof Key for Code Exchange): 1) Client редиректит пользователя на Authorization Server с параметрами (client_id, redirect_uri, scope, state, code_challenge), 2) пользователь аутентифицируется и соглашается предоставить доступ, 3) Authorization Server редиректит обратно на redirect_uri с authorization code, 4) Client обменивает code на access token через backend запрос с client_secret и code_verifier, 5) Client использует access token для запросов к Resource Server. PKCE защищает от authorization code interception attack для mobile/SPA приложений. В моем сервисе интеграций я реализовал OAuth 2.0 client для подключения к Google Calendar и Slack: сохранял access/refresh tokens в БД (encrypted), автоматически обновлял access token при истечении через refresh token, и отзывал токены при disconnect интеграции, что позволило пользователям безопасно подключать внешние сервисы без передачи паролей.

**Пример на Golang:**

```go
package main

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
    
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

// ========== OAUTH 2.0 CONFIG ==========

var googleOAuthConfig = &oauth2.Config{
    ClientID:     "YOUR_CLIENT_ID.apps.googleusercontent.com",
    ClientSecret: "YOUR_CLIENT_SECRET",
    RedirectURL:  "http://localhost:8080/auth/callback",
    Scopes: []string{
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/userinfo.profile",
    },
    Endpoint: google.Endpoint,
}

// ========== STATE MANAGEMENT (CSRF protection) ==========

var stateStore = make(map[string]time.Time) // В продакшене используйте Redis

func generateState() string {
    b := make([]byte, 32)
    rand.Read(b)
    state := base64.URLEncoding.EncodeToString(b)
    stateStore[state] = time.Now().Add(10 * time.Minute)
    return state
}

func validateState(state string) bool {
    expiry, exists := stateStore[state]
    if !exists {
        return false
    }
    
    delete(stateStore, state)
    return time.Now().Before(expiry)
}

// ========== OAUTH 2.0 HANDLERS ==========

func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Генерируем state для CSRF protection
    state := generateState()
    
    // Создаем authorization URL
    authURL := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    
    // Редиректим пользователя на Google
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
    // Проверяем state (CSRF protection)
    state := r.URL.Query().Get("state")
    if !validateState(state) {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    
    // Получаем authorization code
    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "Code not found", http.StatusBadRequest)
        return
    }
    
    // Обмениваем code на token
    token, err := googleOAuthConfig.Exchange(context.Background(), code)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Получаем информацию о пользователе
    userInfo, err := getUserInfo(token.AccessToken)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Сохраняем токен в сессии (или БД)
    // В продакшене: сохраните в БД с шифрованием, создайте свою сессию
    
    response := map[string]interface{}{
        "user":          userInfo,
        "access_token":  token.AccessToken,
        "refresh_token": token.RefreshToken,
        "expiry":        token.Expiry,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// ========== ПОЛУЧЕНИЕ ДАННЫХ ПОЛЬЗОВАТЕЛЯ ==========

type UserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
}

func getUserInfo(accessToken string) (*UserInfo, error) {
    req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+accessToken)
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get user info: %s", string(body))
    }
    
    var userInfo UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, err
    }
    
    return &userInfo, nil
}

// ========== REFRESH TOKEN ==========

func refreshAccessToken(refreshToken string) (*oauth2.Token, error) {
    tokenSource := googleOAuthConfig.TokenSource(context.Background(), &oauth2.Token{
        RefreshToken: refreshToken,
    })
    
    return tokenSource.Token()
}

// ========== REVOKE TOKEN (LOGOUT) ==========

func revokeToken(token string) error {
    revokeURL := "https://oauth2.googleapis.com/revoke"
    
    data := url.Values{}
    data.Set("token", token)
    
    resp, err := http.PostForm(revokeURL, data)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to revoke token: status %d", resp.StatusCode)
    }
    
    return nil
}

// ========== MIDDLEWARE ДЛЯ PROTECTED ENDPOINTS ==========

func oauthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        // Формат: "Bearer <token>"
        var accessToken string
        fmt.Sscanf(authHeader, "Bearer %s", &accessToken)
        
        // Проверяем токен через Google API
        userInfo, err := getUserInfo(accessToken)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Добавляем user info в context
        ctx := context.WithValue(r.Context(), "user", userInfo)
        next(w, r.WithContext(ctx))
    }
}

// ========== PROTECTED ENDPOINT ==========

func handleProfile(w http.ResponseWriter, r *http.Request) {
    userInfo := r.Context().Value("user").(*UserInfo)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(userInfo)
}

// ========== MAIN ==========

func main() {
    // Public endpoints
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(indexHTML))
    })
    http.HandleFunc("/auth/login", handleLogin)
    http.HandleFunc("/auth/callback", handleCallback)
    
    // Protected endpoints
    http.HandleFunc("/api/profile", oauthMiddleware(handleProfile))
    
    fmt.Println("OAuth 2.0 server starting on :8080")
    fmt.Println("Visit http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

// ========== HTML PAGE ==========

const indexHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>OAuth 2.0 Example</title>
</head>
<body>
    <h1>OAuth 2.0 Google Login Example</h1>
    <a href="/auth/login">
        <button>Login with Google</button>
    </a>
</body>
</html>
`
```

**OAuth 2.0 Flow diagram:**

```
╔═══════════════════════════════════════════════════════════════════╗
║ AUTHORIZATION CODE FLOW (с PKCE)                                  ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  1. USER clicks "Login with Google"                               ║
║     │                                                              ║
║     ▼                                                              ║
║  2. CLIENT redirects to Authorization Server                      ║
║     GET /authorize?                                                ║
║         client_id=...                                              ║
║         redirect_uri=http://myapp.com/callback                    ║
║         scope=email profile                                        ║
║         state=random_csrf_token                                    ║
║         code_challenge=hash(code_verifier)  [PKCE]                ║
║     │                                                              ║
║     ▼                                                              ║
║  3. AUTHORIZATION SERVER показывает consent screen                ║
║     "MyApp хочет доступ к вашему email и profile"                 ║
║     [Allow] [Deny]                                                 ║
║     │                                                              ║
║     ▼ (User clicks Allow)                                          ║
║  4. AUTHORIZATION SERVER redirects back                           ║
║     GET http://myapp.com/callback?                                ║
║         code=AUTHORIZATION_CODE                                    ║
║         state=random_csrf_token                                    ║
║     │                                                              ║
║     ▼                                                              ║
║  5. CLIENT validates state, exchanges code for token              ║
║     POST /token                                                    ║
║         code=AUTHORIZATION_CODE                                    ║
║         client_id=...                                              ║
║         client_secret=...                                          ║
║         redirect_uri=http://myapp.com/callback                    ║
║         code_verifier=original_random_string  [PKCE]              ║
║     │                                                              ║
║     ▼                                                              ║
║  6. AUTHORIZATION SERVER returns tokens                           ║
║     {                                                              ║
║       "access_token": "ya29.a0AfH6SMC...",                         ║
║       "refresh_token": "1//0gKpR...",                              ║
║       "expires_in": 3600,                                          ║
║       "token_type": "Bearer"                                       ║
║     }                                                              ║
║     │                                                              ║
║     ▼                                                              ║
║  7. CLIENT uses access_token to call Resource Server              ║
║     GET /api/userinfo                                              ║
║     Authorization: Bearer ya29.a0AfH6SMC...                        ║
║     │                                                              ║
║     ▼                                                              ║
║  8. RESOURCE SERVER validates token and returns data              ║
║     {                                                              ║
║       "email": "user@example.com",                                 ║
║       "name": "John Doe"                                           ║
║     }                                                              ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝

БЕЗОПАСНОСТЬ:
• state parameter - защита от CSRF
• PKCE - защита от authorization code interception
• redirect_uri validation - только whitelist domains
• HTTPS обязателен для production
• Храните client_secret в секретах (не в коде!)
• Шифруйте refresh tokens в БД
```

---

# ФИНАЛЬНЫЙ ЧИТ-ЛИСТ: Быстрое повторение перед собеседованием

## ⚡ Go Core Concepts

**Goroutines & Channels:**
- `go func()` - запуск горутины (легковесный поток, ~2KB стека)
- `ch := make(chan int, 10)` - буферизованный канал
- `select` с `default` - non-blocking операции
- Закрывайте channels от отправителя: `close(ch)`
- WaitGroup для синхронизации горутин

**Context:**
- `context.WithTimeout` - авто-отмена по таймауту
- `context.WithCancel` - ручная отмена
- `ctx.Done()` - канал для проверки отмены
- Передавайте контекст первым параметром

**sync.Map:**
- Thread-safe map, оптимизирована для read-heavy
- Не нужен explicit lock для read/write
- Хуже обычной map с RWMutex для write-heavy

**Slices:**
- `len` vs `cap` - длина vs емкость
- `append` может вызвать реаллокацию (cap удваивается)
- Ловушка: shared underlying array между слайсами

**Interfaces:**
- Duck typing - implicit implementation
- Nil interface ≠ interface со значением nil
- Type assertion: `v.(Type)` или `v, ok := i.(Type)`

**Generics (Go 1.18+):**
- `func Name[T constraint](x T) T`
- Constraints: `any`, `comparable`, custom `interface{}`
- Используйте для контейнеров данных, не для бизнес-логики

**Reflection:**
- `reflect.TypeOf`, `reflect.ValueOf`
- Медленный (10-100x), используйте только когда необходимо
- Для изменения: `reflect.ValueOf(&x).Elem().Set(...)`

## 🗄️ Databases

**PostgreSQL Indexes:**
- B-tree (по умолчанию) - для `=`, `<`, `>`, `ORDER BY`
- GIN - для JSONB, arrays, full-text search
- Partial index: `WHERE status = 'active'`
- Compound: `(user_id, created_at DESC)`

**ACID & Isolation:**
- **Read Committed** (по умолчанию) - видит committed изменения
- **Repeatable Read** - snapshot на начало транзакции
- **Serializable** - самый строгий, retry при conflicts
- Используйте `FOR UPDATE` для pessimistic locking

**EXPLAIN ANALYZE:**
- Seq Scan ❌ - полное сканирование таблицы
- Index Scan ✅ - использует индекс
- Index Only Scan ✅✅ - все данные из индекса
- Смотрите на actual time vs estimated rows

## 📨 Message Brokers

**Kafka Partitions:**
- Message key определяет партицию
- Порядок гарантирован только внутри партиции
- Партиций >= количества consumers для параллелизма

**Transactional Outbox Pattern:**
- Записываем бизнес-данные + event в одной транзакции
- Message Relay публикует события из outbox в Kafka
- Гарантия доставки = eventual consistency

**Idempotency:**
- Используйте unique message IDs
- Храните processed IDs (Redis Set или DB)
- `INSERT ... ON CONFLICT DO NOTHING`

## ☸️ Kubernetes

**Probes:**
- **Startup** - для медленного старта (12 × 5s = 60s)
- **Liveness** - перезапуск если зависло
- **Readiness** - убирает из Service если не готов

**Resources:**
- `requests` - гарантированные ресурсы (для scheduling)
- `limits` - максимум (OOMKilled если превышен)
- CPU throttling может замедлить приложение

## 🏗️ Architectural Patterns

**CQRS:**
- Разделяем Write Model (commands) и Read Model (queries)
- Events синхронизируют модели (eventual consistency)
- Optimistic locking для concurrency

**Event Sourcing:**
- Храним события, а не состояние
- Текущее состояние = replay всех событий
- Audit log из коробки, сложность в миграциях

## 🔐 Security

**JWT:**
- Header.Payload.Signature (base64url encoded)
- Stateless, но нельзя отозвать до expiry
- Access token (5 min) + Refresh token (7 days)
- Храните refresh в httpOnly cookie

**OAuth 2.0:**
- Authorization Code Flow - для веб-приложений
- PKCE - защита от code interception
- state parameter - CSRF protection
- Никогда не храните пароли пользователей

## 📊 Monitoring

**SLI/SLO/SLA:**
- **SLI** - метрика (% успешных запросов)
- **SLO** - цель (99.9% за месяц)
- **SLA** - контракт (штрафы при нарушении)
- Error Budget = 100% - SLO

**Prometheus Metrics:**
- **Counter** - монотонно растет (requests_total)
- **Gauge** - текущее значение (active_connections)
- **Histogram** - распределение (latency buckets)
- **Summary** - квантили на клиенте

**PromQL Queries:**
```promql
rate(http_requests_total[5m])  # RPS
histogram_quantile(0.99, rate(http_duration_bucket[5m]))  # p99 latency
increase(errors_total[1h])  # Total errors in 1h
```

## 🧮 Algorithms

**A* (A-star):**
- f(n) = g(n) + h(n)
- g(n) = cost from start, h(n) = heuristic to goal
- Optimal if heuristic admissible (h(n) ≤ h*(n))

**Dijkstra:**
- Shortest paths from single source
- A* с h(n) = 0 (слепой поиск)
- O((V+E) log V) с priority queue

**KNN:**
- Lazy learning - no training phase
- K=5-10 типично, нечетное для tie-breaking
- Используйте KD-Tree/Ball Tree для ускорения

**TSP:**
- NP-hard, (N-1)!/2 вариантов
- Greedy Nearest Neighbor - O(N²), не более 2x от оптимального
- 2-Opt улучшение - swap ребер для уменьшения пересечений

**DBSCAN:**
- Находит кластеры по плотности
- Parameters: ε (radius), minPts (min points in cluster)
- Обрабатывает шум, произвольная форма кластеров

## 🌐 Real-Time

**WebSocket:**
- Full-duplex over single TCP connection
- HTTP Upgrade handshake
- Sticky sessions или pub/sub (Redis) для масштабирования
- Ping/Pong для keep-alive (54s интервал)

## 💡 Quick Tips

- Всегда закрывайте resources: `defer file.Close()`
- Используйте `context.Context` для cancellation
- Indexes != magic bullet - измеряйте через EXPLAIN
- Мониторинг > логи для production issues
- Eventual consistency - норма для distributed systems
- Graceful shutdown - дайте время завершить запросы

---

**Конец документа. Удачи на собеседовании! 🚀**


### 30. Ball Tree - пространственная структура для KNN

**Ответ:**
Ball Tree — это бинарное дерево для ускорения поиска ближайших соседей, где каждый узел содержит гиперсферу (ball) минимального радиуса, охватывающую все точки поддерева, что позволяет отсекать целые поддеревья при поиске соседей если их ball находится дальше текущих K кандидатов. Ball Tree превосходит KD-Tree в высоких размерностях (d>20) и неравномерно распределенных данных, строится за O(N log N) через recursive partitioning по центроиду и pivot axis. Используется в scikit-learn для KNN и DBSCAN, особенно эффективен для метрик типа Haversine (географические координаты) и косинусного расстояния (text embeddings).

```go
// Упрощенная Ball Tree node структура
type BallNode struct {
    Center    []float64  // Центр гиперсферы
    Radius    float64    // Радиус охватывающей сферы
    Points    []*Point   // Точки (только в листьях)
    Left      *BallNode
    Right     *BallNode
}

// Поиск K ближайших соседей
func (bn *BallNode) QueryKNN(query []float64, k int, heap *PriorityQueue) {
    distToCenter := distance(query, bn.Center)
    
    // Отсечение: если ball дальше чем k-й сосед, пропускаем
    if heap.Len() >= k && distToCenter - bn.Radius > heap.Peek().Distance {
        return
    }
    
    if bn.IsLeaf() {
        for _, p := range bn.Points {
            dist := distance(query, p.Features)
            heap.Push(&Neighbor{Point: p, Distance: dist})
            if heap.Len() > k {
                heap.Pop()
            }
        }
    } else {
        // Рекурсивно обходим ближайшее поддерево первым
        if distance(query, bn.Left.Center) < distance(query, bn.Right.Center) {
            bn.Left.QueryKNN(query, k, heap)
            bn.Right.QueryKNN(query, k, heap)
        } else {
            bn.Right.QueryKNN(query, k, heap)
            bn.Left.QueryKNN(query, k, heap)
        }
    }
}
```

---

### 31. Greedy Algorithms - жадные алгоритмы

**Ответ:**
Greedy algorithms делают локально оптимальный выбор на каждом шаге в надежде найти глобальный оптимум, не пересматривая прошлые решения, что делает их быстрыми O(N log N) но не всегда корректными (работают только для задач с matroid/greedy-choice property). Классические примеры: Dijkstra, Prim's MST, Huffman coding, Activity Selection (выбор max непересекающихся интервалов), Fractional Knapsack (можно брать дроби предметов). Greedy НЕ работает для 0/1 Knapsack (нужен DP), TSP (нужны эвристики), Graph Coloring (NP-complete). В production используйте greedy как fast approximation с последующим улучшением через local search (2-opt для TSP).

```go
// Activity Selection: max непересекающихся интервалов
type Activity struct {
    Start, End int
}

func selectActivities(activities []Activity) []Activity {
    // Сортируем по времени окончания (greedy choice)
    sort.Slice(activities, func(i, j int) bool {
        return activities[i].End < activities[j].End
    })
    
    selected := []Activity{activities[0]}
    lastEnd := activities[0].End
    
    for _, act := range activities[1:] {
        if act.Start >= lastEnd {  // Не пересекается
            selected = append(selected, act)
            lastEnd = act.End
        }
    }
    return selected  // Оптимально!
}

// Fractional Knapsack: можно брать дроби
type Item struct {
    Value, Weight float64
}

func fractionalKnapsack(items []Item, capacity float64) float64 {
    // Сортируем по value/weight ratio (greedy)
    sort.Slice(items, func(i, j int) bool {
        return items[i].Value/items[i].Weight > items[j].Value/items[j].Weight
    })
    
    totalValue := 0.0
    for _, item := range items {
        if capacity >= item.Weight {
            totalValue += item.Value
            capacity -= item.Weight
        } else {
            totalValue += item.Value * (capacity / item.Weight)
            break
        }
    }
    return totalValue
}
```

---

### 32. ELK Stack - централизованный логгинг

**Ответ:**
ELK Stack (Elasticsearch + Logstash + Kibana) — это популярный стек для сбора, хранения, поиска и визуализации логов: Logstash собирает логи из multiple sources (files, syslog, kafka) и парсит через grok patterns, Elasticsearch индексирует для быстрого full-text search, Kibana предоставляет UI для dashboards и поиска. Альтернативы: Filebeat/Fluentd вместо Logstash (легковеснее), Grafana Loki (метки вместо full-text index, дешевле), Clickhouse (columnar, быстрее для аналитики). В production структурируйте логи как JSON, добавляйте correlation IDs (trace_id для distributed tracing), ротируйте indices по дням (logs-2024-01-15), используйте ILM policies для удаления старых логов.

```yaml
# Filebeat config (легковеснее Logstash)
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/app/*.log
  json.keys_under_root: true
  json.add_error_key: true
  fields:
    service: adtime-backend
    environment: production

output.elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  index: "logs-%{[service]}-%{+yyyy.MM.dd}"

# Logstash grok pattern
filter {
  grok {
    match => { "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{DATA:module} - %{GREEDYDATA:message}" }
  }
  date {
    match => [ "timestamp", "ISO8601" ]
  }
}
```

```go
// Structured logging для ELK
import "github.com/sirupsen/logrus"

log := logrus.New()
log.SetFormatter(&logrus.JSONFormatter{})

log.WithFields(logrus.Fields{
    "user_id":      userID,
    "generation_id": genID,
    "trace_id":     ctx.Value("trace_id"),
    "latency_ms":   latency,
}).Info("Generation completed")

// В Kibana можно искать: user_id:"123" AND latency_ms:>1000
```

---

### 33. Central Logging в Clickhouse

**Ответ:**
Clickhouse для логгирования — это columnar OLAP база, оптимизированная для INSERT и аналитических запросов (aggregations, GROUP BY), в 10-100x быстрее Elasticsearch для time-series analytics при меньшем потреблении ресурсов благодаря компрессии. Создайте MergeTree таблицу с PARTITION BY toYYYYMM(timestamp) для автоматической ротации, ORDER BY (timestamp, level, service) для эффективных range queries, TTL для автоудаления старых данных. Используйте Vector/Fluentd для буферизации и батчинга INSERT'ов (batch size 10k-100k), избегайте UPDATE/DELETE (Clickhouse не OLTP база). Для full-text search добавьте ngram tokenbf_v1 index или интегрируйте с Elasticsearch hybrid подход.

```sql
-- Clickhouse таблица для логов
CREATE TABLE logs (
    timestamp DateTime,
    level LowCardinality(String),
    service LowCardinality(String),
    message String,
    user_id UUID,
    generation_id Nullable(UUID),
    trace_id String,
    latency_ms UInt32,
    status_code UInt16,
    metadata String  -- JSON as String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, level, service)
TTL timestamp + INTERVAL 30 DAY  -- Автоудаление через 30 дней
SETTINGS index_granularity = 8192;

-- Быстрые аналитические запросы
SELECT 
    service,
    countIf(level = 'ERROR') as errors,
    avg(latency_ms) as avg_latency_ms,
    quantile(0.99)(latency_ms) as p99_latency_ms
FROM logs
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY service
ORDER BY errors DESC;

-- Error rate за последние 5 минут
SELECT 
    toStartOfMinute(timestamp) as minute,
    countIf(level = 'ERROR') * 100.0 / count() as error_rate_percent
FROM logs
WHERE timestamp >= now() - INTERVAL 5 MINUTE
GROUP BY minute
ORDER BY minute;
```

```go
// Go client для Clickhouse логов
import "github.com/ClickHouse/clickhouse-go/v2"

type LogEntry struct {
    Timestamp    time.Time
    Level        string
    Service      string
    Message      string
    UserID       uuid.UUID
    GenerationID uuid.NullUUID
    TraceID      string
    LatencyMs    uint32
    StatusCode   uint16
}

func (l *Logger) BatchInsert(entries []LogEntry) error {
    batch, _ := l.conn.PrepareBatch(ctx, "INSERT INTO logs")
    
    for _, entry := range entries {
        batch.Append(
            entry.Timestamp,
            entry.Level,
            entry.Service,
            entry.Message,
            entry.UserID,
            entry.GenerationID,
            entry.TraceID,
            entry.LatencyMs,
            entry.StatusCode,
        )
    }
    
    return batch.Send()
}
```

---

# 🎯 ФАЗА 3: ПРИВЯЗКА К ПРОЕКТУ ADTIME

## Контекст проекта

**AdTime Backend MVP** - платформа генерации изображений через Kandinsky AI с последующей печатью на мерче.

**Ключевые модели:**
- `User` (UUID, email, role: user/designer/admin, subscription)
- `Generation` (UUID, user_id, prompt, status, model_version, result_url, retry_count)
- `Order` (UUID, user_id, generation_id, status, amount, factory_id)
- `Payment` (YooKassa интеграция)
- `Subscription` (тарифные планы)

**Сервисы:**
- `KandinskyService` - генерация через API
- `CircuitBreaker` - отказоустойчивость
- `MultiProvider` - fallback между провайдерами (Kandinsky, Leonardo, StableDiffusion)
- `S3Storage` - хранение изображений
- `WebSocket` - real-time статусы генераций

---

## Примеры на моделях AdTime

### 1. Transactional Outbox для Generation Events

```go
// Атомарная запись генерации + event в AdTime
func (s *GenerationService) CreateGeneration(ctx context.Context, userID uuid.UUID, prompt string) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. Создаем Generation
    genID := uuid.New()
    _, err := tx.ExecContext(ctx, `
        INSERT INTO generations (id, user_id, prompt, status, model_version, created_at)
        VALUES ($1, $2, $3, 'pending', 'kandinsky-2.1', NOW())
    `, genID, userID, prompt)
    if err != nil {
        return err
    }
    
    // 2. Записываем event в outbox (в той же транзакции!)
    eventPayload, _ := json.Marshal(map[string]interface{}{
        "generation_id": genID,
        "user_id":       userID,
        "prompt":        prompt,
    })
    
    _, err = tx.ExecContext(ctx, `
        INSERT INTO outbox (event_type, aggregate_id, payload)
        VALUES ('GenerationCreated', $1, $2)
    `, genID, eventPayload)
    if err != nil {
        return err
    }
    
    // 3. Коммит - оба INSERT'а атомарно!
    return tx.Commit()
}

// Message Relay публикует в Kafka → KandinskyWorker получает → начинает генерацию
```

---

### 2. Circuit Breaker для Kandinsky API

```go
// Ваш реальный CircuitBreaker service
type KandinskyService struct {
    client  *http.Client
    breaker *CircuitBreaker
}

func (ks *KandinskyService) Generate(ctx context.Context, gen *Generation) error {
    // Circuit Breaker защищает от каскадных сбоев
    err := ks.breaker.Call(func() error {
        return ks.callKandinskyAPI(ctx, gen)
    })
    
    if err == ErrCircuitOpen {
        // Fallback на другой провайдер
        log.Warn("Kandinsky circuit open, using Leonardo fallback")
        return ks.multiProvider.GenerateWithFallback(ctx, gen)
    }
    
    return err
}

// Параметры Circuit Breaker для AdTime
var kandinskyBreaker = NewCircuitBreaker(CircuitBreakerConfig{
    MaxFailures:    5,              // 5 ошибок подряд
    ResetTimeout:   60 * time.Second, // Через 1 мин пробуем снова
    HalfOpenMax:    3,              // 3 пробных запроса в half-open
})
```

---

### 3. CQRS для Generation History

```go
// WRITE MODEL: PostgreSQL (source of truth)
type GenerationCommandHandler struct {
    db *sql.DB
}

func (h *GenerationCommandHandler) UpdateStatus(ctx context.Context, genID uuid.UUID, status string) error {
    tx, _ := h.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // Update с optimistic locking через retry_count
    result, _ := tx.ExecContext(ctx, `
        UPDATE generations 
        SET status = $1, 
            completed_at = CASE WHEN $1 = 'completed' THEN NOW() ELSE NULL END,
            retry_count = retry_count + 1
        WHERE id = $2 AND retry_count = $3
    `, status, genID, currentRetryCount)
    
    if affected, _ := result.RowsAffected(); affected == 0 {
        return fmt.Errorf("concurrent modification detected")
    }
    
    // Publish event
    tx.ExecContext(ctx, `INSERT INTO outbox ...`)
    
    return tx.Commit()
}

// READ MODEL: Elasticsearch (для быстрого поиска и фильтрации)
type GenerationQueryHandler struct {
    es *elasticsearch.Client
}

func (h *GenerationQueryHandler) SearchGenerations(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]*Generation, error) {
    // Быстрый поиск по денормализованным данным
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "must": []map[string]interface{}{
                    {"term": {"user_id": userID.String()}},
                    {"terms": {"status": filters["statuses"]}},
                    {"range": {"created_at": map[string]string{
                        "gte": filters["from"].(string),
                    }}},
                },
            },
        },
        "sort": [{"created_at": "desc"}],
    }
    
    // В Elasticsearch храним денормализованные данные:
    // {generation_id, user_email, prompt, status, result_url, created_at}
    return h.es.Search("generations", query)
}

// Event Handler синхронизирует модели
func (u *ReadModelUpdater) HandleGenerationCompleted(ctx context.Context, event GenerationCompletedEvent) error {
    doc := map[string]interface{}{
        "generation_id": event.GenerationID,
        "user_id":       event.UserID,
        "user_email":    event.UserEmail,  // Денормализация!
        "prompt":        event.Prompt,
        "status":        "completed",
        "result_url":    event.ResultURL,
        "completed_at":  time.Now(),
    }
    
    return u.es.Index(ctx, "generations", event.GenerationID, doc)
}
```

---

### 4. Prometheus Metrics для AdTime

```go
// Метрики специфичные для AdTime
var (
    generationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "adtime_generations_total",
            Help: "Total number of generation requests",
        },
        []string{"model_version", "status"},
    )
    
    generationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "adtime_generation_duration_seconds",
            Help:    "Generation processing time",
            Buckets: []float64{5, 10, 30, 60, 120, 300},  // Kandinsky занимает 10-60 секунд
        },
        []string{"model_version"},
    )
    
    kandinskyAPIErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "adtime_kandinsky_api_errors_total",
            Help: "Kandinsky API errors by type",
        },
        []string{"error_type"},  // timeout, 5xx, rate_limit
    )
    
    activeWebsocketConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "adtime_websocket_connections_active",
            Help: "Current number of WebSocket connections",
        },
    )
)

// В вашем GenerationService
func (s *GenerationService) ProcessGeneration(ctx context.Context, gen *Generation) {
    start := time.Now()
    
    err := s.kandinsky.Generate(ctx, gen)
    
    duration := time.Since(start).Seconds()
    generationDuration.WithLabelValues(gen.ModelVersion).Observe(duration)
    
    if err != nil {
        status := "failed"
        errorType := classifyError(err)
        kandinskyAPIErrors.WithLabelValues(errorType).Inc()
    } else {
        status := "completed"
    }
    
    generationsTotal.WithLabelValues(gen.ModelVersion, status).Inc()
}

// Alerting rules для AdTime
/*
- alert: HighGenerationFailureRate
  expr: |
    rate(adtime_generations_total{status="failed"}[5m])
    / rate(adtime_generations_total[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High generation failure rate (>10%)"

- alert: KandinskyAPIDown
  expr: up{job="kandinsky-api"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Kandinsky API is down"
*/
```

---

### 5. WebSocket Real-Time для Generation Status

```go
// Ваш реальный WebSocketManager
type GenerationStatusHub struct {
    clients    map[uuid.UUID]map[*websocket.Conn]bool  // userID -> connections
    register   chan *Client
    unregister chan *Client
    broadcast  chan *GenerationUpdate
    redis      *redis.Client
}

type GenerationUpdate struct {
    GenerationID uuid.UUID `json:"generation_id"`
    Status       string     `json:"status"`
    ResultURL    *string    `json:"result_url,omitempty"`
    Progress     int        `json:"progress"`  // 0-100%
}

func (h *GenerationStatusHub) NotifyUser(userID uuid.UUID, update *GenerationUpdate) {
    // Публикуем в Redis для multi-server support
    updateJSON, _ := json.Marshal(update)
    h.redis.Publish(ctx, fmt.Sprintf("generation:updates:%s", userID), updateJSON)
    
    // Отправляем всем WebSocket подключениям пользователя
    if connections, ok := h.clients[userID]; ok {
        for conn := range connections {
            conn.WriteJSON(update)
        }
    }
}

// В GenerationService после каждого статуса
func (s *GenerationService) UpdateGenerationStatus(ctx context.Context, genID uuid.UUID, status string) {
    // UPDATE в БД
    s.repo.UpdateStatus(ctx, genID, status)
    
    // Real-time уведомление через WebSocket
    gen, _ := s.repo.GetByID(ctx, genID)
    s.wsHub.NotifyUser(gen.UserID, &GenerationUpdate{
        GenerationID: genID,
        Status:       status,
        Progress:     calculateProgress(status),
        ResultURL:    gen.ResultURL,
    })
}
```

---

### 6. SLI/SLO для AdTime

```yaml
# Service Level Indicators для AdTime
SLIs:
  # Availability: % successful generation requests
  - name: generation_availability
    query: |
      sum(rate(adtime_generations_total{status!="failed"}[5m]))
      / sum(rate(adtime_generations_total[5m]))
    target: 99.5%
    
  # Latency: % generations completed within 60 seconds
  - name: generation_latency
    query: |
      histogram_quantile(0.95,
        rate(adtime_generation_duration_seconds_bucket[5m])
      ) < 60
    target: 95%
    
  # Success Rate: % generations without errors
  - name: generation_success_rate
    query: |
      sum(rate(adtime_generations_total{status="completed"}[5m]))
      / sum(rate(adtime_generations_total{status=~"completed|failed"}[5m]))
    target: 99%

SLOs:
  - service: AdTime Generation API
    availability: 99.5%  # 3.6 часа downtime в месяц
    latency_p95: 60s
    success_rate: 99%
    error_budget: 0.5%  # = 21.6K failed из 4.3M generations/month

# При нарушении SLO → freeze deployments, focus on reliability
```

---

### 7. Distributed Tracing для Multi-Provider Fallback

```go
// OpenTelemetry tracing для AdTime
import "go.opentelemetry.io/otel"

func (s *MultiProviderService) GenerateWithFallback(ctx context.Context, gen *Generation) error {
    ctx, span := otel.Tracer("adtime").Start(ctx, "GenerateWithFallback")
    defer span.End()
    
    // Пробуем провайдеров по приоритету
    providers := []Provider{
        s.kandinsky,   // Primary
        s.leonardo,    // Fallback 1
        s.craiyon,     // Fallback 2
    }
    
    for i, provider := range providers {
        providerCtx, providerSpan := otel.Tracer("adtime").Start(ctx, fmt.Sprintf("Provider:%s", provider.Name()))
        
        err := provider.Generate(providerCtx, gen)
        
        if err == nil {
            providerSpan.SetAttributes(attribute.Bool("success", true))
            providerSpan.End()
            span.SetAttributes(attribute.String("used_provider", provider.Name()))
            return nil
        }
        
        providerSpan.SetAttributes(
            attribute.Bool("success", false),
            attribute.String("error", err.Error()),
        )
        providerSpan.End()
        
        log.WithFields(logrus.Fields{
            "trace_id":   span.SpanContext().TraceID().String(),
            "provider":   provider.Name(),
            "attempt":    i + 1,
            "error":      err,
        }).Warn("Provider failed, trying next")
    }
    
    return fmt.Errorf("all providers failed")
}

// В Jaeger UI видно всю цепочку: GenerateWithFallback → Kandinsky (failed) → Leonardo (success)
```

---

## �� Финальные рекомендации для AdTime

**Что уже хорошо:**
✅ Circuit Breaker для Kandinsky API
✅ WebSocket для real-time статусов
✅ Multi-provider fallback
✅ Structured logging

**Что улучшить:**

1. **Добавить Transactional Outbox** для гарантированной публикации events при изменении Generation.status
2. **CQRS для истории** - Elasticsearch для быстрого поиска generations с фильтрами
3. **SLI/SLO мониторинг** - дашборд с error budget и alerting по burn rate
4. **Distributed tracing** - OpenTelemetry для анализа latency в multi-provider chain
5. **Graceful shutdown** - дать время завершить активные генерации перед деплоем
6. **Idempotency keys** - для retry запросов к Kandinsky API без дублирования

**Архитектура будущего:**

```
User Request → API Gateway (rate limiting)
           ↓
    FastAPI Backend → PostgreSQL (generations, orders)
           ↓
    Kafka/Redis (event bus)
           ↓
    Generation Workers (Celery/Go)
      ├─→ Kandinsky API (primary)
      ├─→ Leonardo API (fallback 1)
      └─→ Craiyon API (fallback 2)
           ↓
    S3 Storage (result images)
           ↓
    WebSocket Hub → Browser (real-time updates)
    
Monitoring: Prometheus + Grafana + Jaeger
Logging: Clickhouse (analytics) + Elasticsearch (search)
```

---

**🎉 ДОКУМЕНТ ЗАВЕРШЕН! Удачи на собеседовании в Gcore! 🚀**

