# FINAL TASKS — Единый документ с задачами

*Этот документ объединяет все задачи для практики из teachers/dondo, kata_mastermind, include_in_stickers и других источников.*

---

## ЧАСТЬ 1: БАЗОВЫЕ ЗАДАЧИ (MAP, SLICE, MATH)

### Задача 1: Find Duplicates (Найти дубликаты)

**Условие:**
Дан массив чисел. Вернуть `map[int]int`, где ключ — элемент массива, значение — количество раз, которое этот элемент встречается.

**Пример:**
```go
Input: nums = [1, 2, 2, 2, 6, 7, 8, 9, 10, 3, 4, 5]
Output: map[1:1 2:3 3:1 4:1 5:1 6:1 7:1 8:1 9:1 10:1]
```

**Сигнатура:**
```go
func findDupes(nums []int) map[int]int
```

**Подсказка:**
- Используй `map[int]int` для подсчёта
- `m[nums[i]]++` работает, потому что при обращении к несуществующему ключу Go возвращает 0 для `int`

**Решение:** (не смотри сразу!)

<details>
<summary>Показать решение</summary>

```go
func findDupes(nums []int) map[int]int {
    m := make(map[int]int)
    for i := 0; i < len(nums); i++ {
        m[nums[i]]++
    }
    return m
}
```

</details>

---

### Задача 2: 643. Maximum Average Subarray I (LeetCode)

**Условие:**
Дан массив целых чисел `nums` и целое число `k`. Найти непрерывный подмассив длины `k`, который имеет максимальное среднее значение, и вернуть это среднее.

**Пример 1:**
```
Input: nums = [1,12,-5,-6,50,3], k = 4
Output: 12.75000
Explanation: Maximum average is (12 - 5 - 6 + 50) / 4 = 51 / 4 = 12.75
```

**Пример 2:**
```
Input: nums = [5], k = 1
Output: 5.00000
```

**Сигнатура:**
```go
func findMaxAverage(nums []int, k int) float64
```

**Подход:** Скользящее окно (Sliding Window)
- Посчитай сумму первых `k` элементов
- Двигай окно вправо: вычитай левый элемент, добавляй правый
- Отслеживай максимальную сумму
- В конце раздели на `k`

**Решение:** (не смотри сразу!)

<details>
<summary>Показать решение</summary>

```go
func findMaxAverage(nums []int, k int) float64 {
    var res float64
    var maxSum int
    window := 0

    // Считаем сумму первого окна
    for j := 0; j < k; j++ {
        window += nums[j]
    }

    maxSum = window

    // Двигаем окно
    for i := 0; i < len(nums) - k; i++ {
        newSum := window - nums[i] + nums[i+k]

        if newSum > maxSum {
            maxSum = newSum
        }

        window = newSum
    }

    res = float64(maxSum) / float64(k)

    return res
}
```

</details>

---

### Задача 3: 1. Two Sum (LeetCode)

**Условие:**
Дан массив целых чисел `nums` и целое число `target`. Вернуть индексы двух элементов, которые в сумме дают `target`.

**Пример:**
```
Input: nums = [2,7,11,15], target = 9
Output: [0,1]
Explanation: nums[0] + nums[1] == 9, вернём [0,1]
```

**Сигнатура:**
```go
func twoSum(nums []int, target int) []int
```

**Подход:** Hash Map (O(n))
- Создай `map[int]int` для хранения `значение → индекс`
- Для каждого элемента `num` проверяй, есть ли в мапе `target - num`
- Если есть → вернуть индексы
- Если нет → добавить `num` в мапу

**Решение:** (не смотри сразу!)

<details>
<summary>Показать решение</summary>

```go
func twoSum(nums []int, target int) []int {
    seen := make(map[int]int)

    for i, num := range nums {
        res := target - num
        if j, ok := seen[res]; ok {
            return []int{j, i}
        }
        seen[num] = i
    }

    return nil
}
```

**Объяснение:**
- `seen` — мапа, где ключ — значение элемента, значение — индекс
- `res = target - num` — что нам нужно найти
- `if j, ok := seen[res]` — comma-ok идиома: проверяем, есть ли `res` в мапе
- Если есть → вернуть индексы `[j, i]`
- Если нет → добавить текущий элемент `seen[num] = i`

</details>

---

## ЧАСТЬ 2: CONCURRENCY (ГОРУТИНЫ, КАНАЛЫ, SYNC)

### Задача 4: Weather API с кэшированием (Mutex + Sync.Once)

**Условие:**
Есть ручка `/weather`, которая отрабатывает за ~1 секунду (вызов `aiWeatherForecast()`). Нужно ускорить выполнение этой ручки, используя кэширование.

**Требования:**
- Кэшировать результат `aiWeatherForecast()` на 5 секунд (TTL)
- Использовать `sync.RWMutex` для защиты кэша
- Использовать `sync.Once` для инициализации первого значения
- Использовать `context.WithCancel` для graceful shutdown

**Подход 1: RWMutex + Double-Check Locking**

```go
package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

var (
    cacheTemp   int
    cacheTime   time.Time
    cacheMutex  sync.RWMutex
    cacheTTL    = 5 * time.Second
)

func aiWeatherForecast() int {
    time.Sleep(1 * time.Second)
    return 20 + time.Now().Second()%10
}

func getCachedWeather() int {
    cacheMutex.RLock()
    
    // Если кэш свежий - возвращаем
    if time.Since(cacheTime) < cacheTTL {
        temp := cacheTemp
        cacheMutex.RUnlock()
        return temp
    }
    cacheMutex.RUnlock()
    
    // Обновляем кэш
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    
    // Double-check (на случай если другая горутина уже обновила)
    if time.Since(cacheTime) < cacheTTL {
        return cacheTemp
    }
    
    cacheTemp = aiWeatherForecast()
    cacheTime = time.Now()
    return cacheTemp
}

func main() {
    http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
        temp := getCachedWeather()
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"temperature":%d}`, temp)
    })
    
    fmt.Println("Server starting on :3333...")
    http.ListenAndServe(":3333", nil)
}
```

**Подход 2: Sync.Once + Background Update**

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"
)

var (
    temperature int
    mu          sync.RWMutex
    once        sync.Once
)

func aiWeatherForecast() int {
    time.Sleep(1 * time.Second)
    return 20 + time.Now().Second()%10
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Background updater
    go func(ctx context.Context) {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                temp := aiWeatherForecast()
                mu.Lock()
                temperature = temp
                mu.Unlock()
                fmt.Printf("Temperature updated: %d°C\n", temp)  
            case <-ctx.Done():
                return
            }
        }
    }(ctx)

    http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
        mu.RLock()
        temp := temperature
        mu.RUnlock()

        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"temperature":%d}\n`, temp)
    })

    // Инициализация первого значения (Sync.Once)
    once.Do(func() {
        temp := aiWeatherForecast()
        mu.Lock()
        temperature = temp
        mu.Unlock()
    })
    
    if err := http.ListenAndServe(":3333", nil); err != nil {
        panic(err)
    }
}
```

**Вопросы для размышления:**
- Зачем нужен Double-Check Locking в подходе 1?
- Почему в подходе 2 background updater лучше для highload?
- Что будет, если убрать `defer cancel()` в подходе 2?

---

### Задача 5: Fan-Out Pattern (распределение работы)

**Условие:**
Дан слайс данных. Нужно обработать каждый элемент в отдельной горутине и собрать результаты.

**Пример:** Скачивание 10 видео с YouTube параллельно.

**Подход:** Fan-Out / Fan-In

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    
    s1 := []int{1, 2, 3, 4, 5}
    ch1 := make(chan int, len(s1))

    /*
                    or here
                    /
        data ->  FANOUT -> or here
                    \
                    or here
    */
    for i := 0; i < len(s1); i++ {
        wg.Add(1)

        go func(val int) {  // <- запускаем много горутин
            defer wg.Done()
            // Имитация обработки (например, загрузка видео)
            // processed := val * 2 // пример обработки
            ch1 <- val      // <- каждая пишет в общий канал
        }(s1[i])

    }

    go func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Println("Recovered:", r)
            }
        }()
        wg.Wait()
        close(ch1)
    }()

    // Это FAN-IN - сбор результатов
    for v := range ch1 {    // <- читаем из одного канала
        fmt.Println(v)      //    результаты всех горутин
    }
    fmt.Println("Все данные обработаны!")
}
```

**Вопросы:**
- Зачем нужен буферизированный канал с размером `len(s1)`?
- Что будет, если забыть `close(ch1)`?
- Зачем `recover()` в горутине, которая закрывает канал?

---

### Задача 6: Fan-Out с дженериками (Split Channel)

**Условие:**
Написать функцию `SplitChannel[T any](inputCh <-chan T, n int) []<-chan T`, которая принимает один входной канал и распределяет данные по `n` выходным каналам (Round Robin).

**Решение:**

```go
package main

import (
    "fmt"
    "sync"
)

func SplitChannel[T any](inputCh <-chan T, n int) []<-chan T {
    outputChannels := make([]chan T, n)
    for i := 0; i < n; i++ {
        outputChannels[i] = make(chan T)
    }

    go func() {
        idx := 0
        for value := range inputCh {
            outputChannels[idx] <- value // блокирующая запись
            idx = (idx + 1) % n // Round Robin
        }

        for _, ch := range outputChannels {
            close(ch)
        }
    }()

    // Typecast to []<-chan T (read-only channels)
    resultChannels := make([]<-chan T, n)
    for i := 0; i < n; i++ {
        resultChannels[i] = outputChannels[i]
    }
    
    return resultChannels
}

func main() {
    channel := make(chan int)

    go func() {
        defer close(channel)
        for i := 0; i < 10; i++ {
            channel <- i
        }
    }()

    channels := SplitChannel(channel, 2)
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        for value := range channels[0] {
            fmt.Println("ch1: ", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range channels[1] {
            fmt.Println("ch2: ", value)
        }
    }()

    wg.Wait()
}
```

**Вопросы:**
- Что такое Round Robin и зачем он нужен?
- Почему возвращаем `[]<-chan T` (read-only) вместо `[]chan T`?
- Что будет, если один из выходных каналов не читается?

---

### Задача 7: Worker Pool с Error Group

**Условие:**
Создать пул из N воркеров, которые обрабатывают задачи из канала. Если хотя бы один воркер возвращает ошибку — остановить всех и вернуть ошибку.

**Подход:** `golang.org/x/sync/errgroup`

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "golang.org/x/sync/errgroup"
    "time"
)

func worker(ctx context.Context, id int, jobs <-chan int, results chan<- int) error {
    for {
        select {
        case job, ok := <-jobs:
            if !ok {
                return nil // Канал закрыт
            }
            // Имитация обработки
            time.Sleep(100 * time.Millisecond)
            
            // Симуляция ошибки для воркера 3
            if id == 3 && job == 5 {
                return errors.New("worker 3 failed on job 5")
            }
            
            results <- job * 2
        case <-ctx.Done():
            return ctx.Err() // Контекст отменён
        }
    }
}

func main() {
    jobs := make(chan int, 10)
    results := make(chan int, 10)

    g, ctx := errgroup.WithContext(context.Background())

    // Запускаем 5 воркеров
    for i := 1; i <= 5; i++ {
        workerID := i
        g.Go(func() error {
            return worker(ctx, workerID, jobs, results)
        })
    }

    // Producer: отправляем задачи
    g.Go(func() error {
        defer close(jobs)
        for i := 1; i <= 10; i++ {
            select {
            case jobs <- i:
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        return nil
    })

    // Consumer: собираем результаты
    go func() {
        g.Wait() // Ждём завершения всех горутин
        close(results)
    }()

    // Собираем результаты
    for result := range results {
        fmt.Println("Result:", result)
    }

    // Проверяем ошибки
    if err := g.Wait(); err != nil {
        fmt.Println("Error:", err)
    }
}
```

**Вопросы:**
- Как Error Group останавливает всех воркеров при ошибке?
- Зачем нужен `context.Context` в Error Group?
- Что будет, если забыть `close(jobs)` в producer?

---

## ЧАСТЬ 3: ЗАДАЧИ НА АЛГОРИТМЫ

### Задача 8: Valid Parentheses (Валидные скобки)

**Условие:**
Дана строка, содержащая только символы `'(', ')', '{', '}', '[', ']'`. Определить, является ли строка корректной.

**Пример:**
```
Input: s = "()[]{}"
Output: true

Input: s = "(]"
Output: false
```

**Подход:** Stack (стек)

**Решение:** (попробуй сам!)

<details>
<summary>Показать решение</summary>

```go
func isValid(s string) bool {
    stack := []rune{}
    pairs := map[rune]rune{
        ')': '(',
        '}': '{',
        ']': '[',
    }

    for _, char := range s {
        if char == '(' || char == '{' || char == '[' {
            // Открывающая скобка — добавляем в стек
            stack = append(stack, char)
        } else {
            // Закрывающая скобка
            if len(stack) == 0 {
                return false // Стек пуст, но пришла закрывающая
            }
            // Проверяем, соответствует ли последняя открывающая
            if stack[len(stack)-1] != pairs[char] {
                return false
            }
            // Убираем последнюю открывающую скобку
            stack = stack[:len(stack)-1]
        }
    }

    // Стек должен быть пуст в конце
    return len(stack) == 0
}
```

</details>

---

## ЧАСТЬ 4: SYSTEM DESIGN / ARCHITECTURE

### Задача 9: Design Rate Limiter

**Условие:**
Реализовать rate limiter, который ограничивает количество запросов до N в секунду.

**Подход 1:** Token Bucket (семафор)

```go
package main

import (
    "context"
    "fmt"
    "time"
)

type RateLimiter struct {
    tokens chan struct{}
}

func NewRateLimiter(rps int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, rps),
    }
    
    // Заполняем bucket токенами
    for i := 0; i < rps; i++ {
        rl.tokens <- struct{}{}
    }
    
    // Background refiller
    go func() {
        ticker := time.NewTicker(time.Second / time.Duration(rps))
        defer ticker.Stop()
        
        for range ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
                // Bucket полон
            }
        }
    }()
    
    return rl
}

func (rl *RateLimiter) Allow(ctx context.Context) bool {
    select {
    case <-rl.tokens:
        return true
    case <-ctx.Done():
        return false
    default:
        return false
    }
}

func main() {
    rl := NewRateLimiter(10) // 10 RPS
    
    for i := 0; i < 100; i++ {
        if rl.Allow(context.Background()) {
            fmt.Println("Request allowed:", i)
        } else {
            fmt.Println("Request denied:", i)
        }
        time.Sleep(50 * time.Millisecond)
    }
}
```

**Вопросы:**
- Почему используем буферизированный канал размером `rps`?
- Что делает background refiller?
- Как изменить реализацию для burst capacity (всплеск)?

---

### Задача 10: Design Load Balancer

**Условие:**
Реализовать load balancer, который распределяет запросы между несколькими серверами.

**Подход:** Round Robin

**Решение:** (попробуй сам, используя атомики или мьютексы!)

---

## ЧАСТЬ 5: SQL ЗАДАЧИ

*Практические SQL-задачи для собеседований*

---

### Задача 5.1: Покупки до бана

**Условие:**
Вывести уникальные комбинации пользователя и SKU товара для всех покупок, совершенных пользователями **до того, как их забанили**. Отсортировать сначала по имени пользователя, потом по SKU.

**Таблицы:**
```sql
-- user (пользователи)
id | firstname | lastname | birth
1  | Ivan      | Petrov   | 1996-05-01
2  | Anna      | Petrova  | 1999-06-01
3  | Anna      | Petrova  | 1990-10-02

-- purchase (покупки)
sku| price | user_id | date
1  | 5500  | 1       | 2021-02-15
1  | 5700  | 1       | 2021-01-15
2  | 4000  | 1       | 2021-02-14
3  | 8000  | 2       | 2021-03-01
4  | 4000  | 2       | 2021-03-02

-- ban_list (список бана)
user_id | date_from
1       | 2021-03-08
```

<details>
<summary>Решение</summary>

```sql
SELECT DISTINCT u.firstname, p.sku
FROM user u
JOIN purchase p
    ON p.user_id = u.id
LEFT JOIN ban_list b
    ON b.user_id = u.id
WHERE b.user_id IS NULL  -- пользователь не забанен
   OR p.date < b.date_from  -- или покупка до бана
ORDER BY u.firstname, p.sku;
```

**Альтернативное решение (с подзапросом):**
```sql
SELECT DISTINCT u.firstname, p.sku
FROM user u
JOIN purchase p ON p.user_id = u.id
LEFT JOIN (
    SELECT user_id, MIN(date_from) as first_ban
    FROM ban_list
    GROUP BY user_id
) b ON b.user_id = u.id
WHERE p.date < COALESCE(b.first_ban, '9999-12-31')
ORDER BY u.firstname, p.sku;
```
</details>

---

### Задача 5.2: Заказы до первого возврата

**Условие:**
Вывести уникальные комбинации имени пользователя и номера заказа для всех заказов, совершенных пользователями **до их первого возврата**. Отсортировать сначала по имени, потом по номеру заказа.

**Таблицы:**
```sql
-- users (пользователи)
id | name     | email
1  | Ivan     | ivan@mail.com
2  | Anna     | anna@mail.com
3  | Petr     | petr@mail.com

-- orders (заказы)
id | user_id | order_date | amount
1  | 1       | 2024-01-15 | 5000
2  | 1       | 2024-02-10 | 7000
3  | 2       | 2024-01-20 | 3000
4  | 2       | 2024-02-15 | 4000
5  | 3       | 2024-01-25 | 6000

-- returns (возвраты)
id | user_id | return_date | reason
1  | 1       | 2024-02-20  | Не подошел размер
2  | 2       | 2024-02-25  | Брак товара
```

**Ожидаемый результат:**
```
name | order_id
Ivan | 1
Ivan | 2
Anna | 3
Anna | 4
Petr | 5
```

<details>
<summary>Решение 1: LEFT JOIN с проверкой NULL</summary>

```sql
SELECT DISTINCT u.name, o.id
FROM users u
JOIN orders o 
    ON o.user_id = u.id
LEFT JOIN returns r
    ON r.user_id = u.id
WHERE r.id IS NULL  -- пользователь не делал возвратов
   OR o.order_date < r.return_date  -- или заказ до возврата
ORDER BY u.name, o.id;
```
</details>

<details>
<summary>Решение 2: CTE с MIN(return_date)</summary>

```sql
WITH first_returns AS (
    SELECT 
        user_id,
        MIN(return_date) AS first_return_date
    FROM returns
    GROUP BY user_id
)
SELECT DISTINCT u.name, o.id
FROM users u
JOIN orders o 
    ON o.user_id = u.id
LEFT JOIN first_returns r 
    ON r.user_id = u.id
WHERE r.first_return_date IS NULL 
   OR o.order_date < r.first_return_date
ORDER BY u.name, o.id;
```
</details>

<details>
<summary>Решение 3: Подзапрос с COALESCE</summary>

```sql
SELECT DISTINCT u.name, o.id
FROM users u
JOIN orders o ON o.user_id = u.id
LEFT JOIN (
    SELECT user_id, MIN(return_date) as first_return 
    FROM returns 
    GROUP BY user_id
) r ON r.user_id = u.id
WHERE o.order_date < COALESCE(r.first_return, '9999-12-31')
ORDER BY u.name, o.id;
```
</details>

---

### Задача 5.3: Успешные доставки по категориям

**Условие:**
Вывести уникальные пары "город - категория товара" для всех заказов, где доставка была **успешной** (`status = 'delivered'`).

**Таблицы:**
```sql
-- customers (клиенты)
id | name | city
1  | Alex | Moscow
2  | Kate | SPb

-- products (товары)
id | name       | category
1  | iPhone     | Electronics
2  | T-shirt    | Clothing
3  | Book       | Education

-- deliveries (доставки)
id | customer_id | product_id | status    | delivery_date
1  | 1           | 1          | delivered | 2024-01-15
2  | 1           | 1          | failed    | 2024-01-20
3  | 1           | 2          | delivered | 2024-01-18
4  | 2           | 3          | delivered | 2024-01-22
```

**Ожидаемый результат:**
```
city    | category
Moscow  | Electronics
Moscow  | Clothing
SPb     | Education
```

<details>
<summary>Решение</summary>

```sql
SELECT DISTINCT c.city, p.category
FROM deliveries d
JOIN customers c ON d.customer_id = c.id
JOIN products p ON d.product_id = p.id
WHERE d.status = 'delivered'
ORDER BY c.city, p.category;
```
</details>

---

### Задача 5.4: Траты на подарки

**Условие:**
Вывести, сколько пользователь `'gamer_ivan'` потратил на подарки в декабре. Показать: username, общую сумму, количество подарков, список купленных игр.

**Таблицы:** Board Games (см. nightcall_training_sql_golang.txt)

<details>
<summary>Решение</summary>

```sql
SELECT 
    u.username,
    SUM(pi.quantity * pi.price_per_item * (1 - pi.discount_percent/100)) as total_spent,
    COUNT(DISTINCT p.id) as gift_count,
    STRING_AGG(DISTINCT bg.title, ', ') as purchased_games
FROM users u
JOIN purchases p ON p.user_id = u.id
JOIN purchase_items pi ON pi.purchase_id = p.id
JOIN board_games bg ON bg.id = pi.game_id
WHERE u.username = 'gamer_ivan'
    AND p.is_gift = TRUE
    AND EXTRACT(MONTH FROM p.purchase_date) = 12
    AND EXTRACT(YEAR FROM p.purchase_date) = EXTRACT(YEAR FROM CURRENT_DATE)
GROUP BY u.id, u.username;
```
</details>

---

### Задача 5.5: Топ самых популярных игр для подарков

**Условие:**
Вывести топ-10 самых популярных игр, которые дарили в декабре. Показать: название игры, количество раз подарено, общее количество экземпляров, список дарителей.

<details>
<summary>Решение</summary>

```sql
SELECT 
    bg.title,
    COUNT(pi.id) as times_gifted,
    SUM(pi.quantity) as total_copies_gifted,
    ARRAY_AGG(DISTINCT u.username) as gift_givers
FROM purchases p
JOIN purchase_items pi ON pi.purchase_id = p.id
JOIN board_games bg ON bg.id = pi.game_id
JOIN users u ON u.id = p.user_id
WHERE p.is_gift = TRUE
    AND EXTRACT(MONTH FROM p.purchase_date) = 12
GROUP BY bg.id, bg.title
ORDER BY times_gifted DESC
LIMIT 10;
```
</details>

---

### Задача 5.6: Популярные игры по возрастным группам

**Условие:**
Вывести самые популярные игры для подарков, сгруппированные по возрастным группам получателей.

<details>
<summary>Решение</summary>

```sql
SELECT 
    bg.title,
    CASE 
        WHEN bg.min_age BETWEEN 6 AND 12 THEN 'Дети (6-12)'
        WHEN bg.min_age BETWEEN 13 AND 17 THEN 'Подростки (13-17)'
        WHEN bg.min_age >= 18 THEN 'Взрослые (18+)'
    END as target_age_group,
    COUNT(*) as gift_count
FROM purchases p
JOIN purchase_items pi ON pi.purchase_id = p.id
JOIN board_games bg ON bg.id = pi.game_id
WHERE p.is_gift = TRUE
GROUP BY bg.id, bg.title, target_age_group
ORDER BY gift_count DESC;
```
</details>

---

### Задача 5.7: Топ дарителей

**Условие:**
Вывести пользователей, которые потратили больше всего на подарки. Показать: username, общую сумму, количество уникальных получателей, дату последнего подарка.

<details>
<summary>Решение</summary>

```sql
SELECT 
    u.username,
    SUM(pi.quantity * pi.price_per_item) as total_spent,
    COUNT(DISTINCT p.gift_for) as unique_recipients,
    MAX(p.purchase_date) as last_gift_date
FROM users u
JOIN purchases p ON p.user_id = u.id
JOIN purchase_items pi ON pi.purchase_id = p.id
WHERE p.is_gift = TRUE
GROUP BY u.id, u.username
ORDER BY total_spent DESC;
```
</details>

---

### Задача 5.8: Средний чек на подарки по месяцам

**Условие:**
Вывести статистику по подаркам: год, месяц, количество подарков, средний чек, минимальный и максимальный чек.

<details>
<summary>Решение</summary>

```sql
SELECT 
    EXTRACT(YEAR FROM p.purchase_date) as year,
    EXTRACT(MONTH FROM p.purchase_date) as month,
    COUNT(*) as total_gifts,
    AVG(p.total_amount) as avg_gift_amount,
    MIN(p.total_amount) as min_gift,
    MAX(p.total_amount) as max_gift
FROM purchases p
WHERE p.is_gift = TRUE
GROUP BY EXTRACT(YEAR FROM p.purchase_date), EXTRACT(MONTH FROM p.purchase_date)
ORDER BY year DESC, month DESC;
```
</details>

---

## ЧАСТЬ 6: АЛГОРИТМИЧЕСКИЕ ЗАДАЧИ (30 шт)

*Источник: algo_task_examples_kozyrev.txt*  
*80% алгоритмических задач с собесов решаются этими паттернами*

---

### РАЗДЕЛ 1: МАССИВЫ И СЛАЙСЫ

#### Задача 6.1: Слияние отсортированных массивов

**Условие:**
Даны два отсортированных массива `nums1` и `nums2`. Слить их **in-place** в `nums1`, который имеет достаточно места.

**Функция:**
```go
func merge(nums1 []int, m int, nums2 []int, n int)
```

**Подход:** Два указателя с конца обоих массивов, заполнение `nums1` с конца.

<details>
<summary>Решение</summary>

```go
func merge(nums1 []int, m int, nums2 []int, n int) {
    i, j, k := m-1, n-1, m+n-1
    
    // Сливаем с конца
    for j >= 0 {
        if i >= 0 && nums1[i] > nums2[j] {
            nums1[k] = nums1[i]
            i--
        } else {
            nums1[k] = nums2[j]
            j--
        }
        k--
    }
}
```

**Сложность:**
- Время: O(m + n)
- Память: O(1)
</details>

---

#### Задача 6.2: Удаление нулей

**Условие:**
Дан массив `nums`, переместить все нули в конец, сохранив порядок ненулевых элементов.

**Функция:**
```go
func moveZeroes(nums []int)
```

**Подход:** Два указателя — один для ненулевых, второй для текущего элемента.

<details>
<summary>Решение</summary>

```go
func moveZeroes(nums []int) {
    slow := 0  // указатель на позицию для следующего ненулевого
    
    for fast := 0; fast < len(nums); fast++ {
        if nums[fast] != 0 {
            nums[slow], nums[fast] = nums[fast], nums[slow]
            slow++
        }
    }
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.3: Произведение элементов кроме текущего

**Условие:**
Дан массив `nums`, вернуть массив `answer`, где `answer[i]` равен произведению всех элементов `nums` кроме `nums[i]`. **Без использования деления.**

**Функция:**
```go
func productExceptSelf(nums []int) []int
```

**Подход:** Два прохода — префиксное и суффиксное произведение.

<details>
<summary>Решение</summary>

```go
func productExceptSelf(nums []int) []int {
    n := len(nums)
    answer := make([]int, n)
    
    // Префиксное произведение (слева направо)
    answer[0] = 1
    for i := 1; i < n; i++ {
        answer[i] = answer[i-1] * nums[i-1]
    }
    
    // Суффиксное произведение (справа налево)
    suffix := 1
    for i := n - 1; i >= 0; i-- {
        answer[i] *= suffix
        suffix *= nums[i]
    }
    
    return answer
}
```

**Сложность:**
- Время: O(n)
- Память: O(1) (не считая выходной массив)
</details>

---

#### Задача 6.4: Поиск пропущенных чисел

**Условие:**
Дан массив `nums` размера `n`, содержащий числа из диапазона `[0, n]`. Найти **все** пропущенные числа.

**Функция:**
```go
func findDisappearedNumbers(nums []int) []int
```

**Подход:** Маркировка индексов (меняем знак числа по индексу).

<details>
<summary>Решение</summary>

```go
func findDisappearedNumbers(nums []int) []int {
    // Маркируем присутствующие числа
    for i := 0; i < len(nums); i++ {
        index := abs(nums[i]) - 1
        if nums[index] > 0 {
            nums[index] = -nums[index]
        }
    }
    
    // Собираем пропущенные (положительные индексы)
    result := []int{}
    for i := 0; i < len(nums); i++ {
        if nums[i] > 0 {
            result = append(result, i+1)
        }
    }
    
    return result
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.5: Вращение массива

**Условие:**
Дан массив `nums`, повернуть его вправо на `k` позиций.

**Функция:**
```go
func rotate(nums []int, k int)
```

**Подход:** Три реверса — весь массив, затем первые k, затем остальные.

<details>
<summary>Решение</summary>

```go
func rotate(nums []int, k int) {
    k %= len(nums)  // на случай если k > len
    
    // Реверс всего массива
    reverse(nums, 0, len(nums)-1)
    // Реверс первых k элементов
    reverse(nums, 0, k-1)
    // Реверс остальных
    reverse(nums, k, len(nums)-1)
}

func reverse(nums []int, left, right int) {
    for left < right {
        nums[left], nums[right] = nums[right], nums[left]
        left++
        right--
    }
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

### РАЗДЕЛ 2: СВЯЗНЫЕ СПИСКИ

#### Задача 6.6: Удаление N-го элемента с конца

**Условие:**
Дан связный список, удалить N-й элемент с конца за **один проход**.

**Функция:**
```go
func removeNthFromEnd(head *ListNode, n int) *ListNode
```

**Подход:** Два указателя с разницей в N узлов.

<details>
<summary>Решение</summary>

```go
type ListNode struct {
    Val  int
    Next *ListNode
}

func removeNthFromEnd(head *ListNode, n int) *ListNode {
    dummy := &ListNode{Next: head}
    slow, fast := dummy, dummy
    
    // Двигаем fast на n+1 шагов вперёд
    for i := 0; i <= n; i++ {
        fast = fast.Next
    }
    
    // Двигаем оба указателя до конца
    for fast != nil {
        slow = slow.Next
        fast = fast.Next
    }
    
    // Удаляем узел
    slow.Next = slow.Next.Next
    
    return dummy.Next
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.7: Разворот связного списка

**Условие:**
Развернуть односвязный список.

**Функция:**
```go
func reverseList(head *ListNode) *ListNode
```

**Подход:** Итеративно меняем указатели.

<details>
<summary>Решение (итеративное)</summary>

```go
func reverseList(head *ListNode) *ListNode {
    var prev *ListNode
    curr := head
    
    for curr != nil {
        next := curr.Next
        curr.Next = prev
        prev = curr
        curr = next
    }
    
    return prev
}
```

**Решение (рекурсивное):**
```go
func reverseList(head *ListNode) *ListNode {
    if head == nil || head.Next == nil {
        return head
    }
    
    newHead := reverseList(head.Next)
    head.Next.Next = head
    head.Next = nil
    
    return newHead
}
```

**Сложность:**
- Время: O(n)
- Память: O(1) итеративно, O(n) рекурсивно (стек)
</details>

---

#### Задача 6.8: Поиск цикла в списке

**Условие:**
Определить, есть ли цикл в односвязном списке (Floyd's Cycle Detection).

**Функция:**
```go
func hasCycle(head *ListNode) bool
```

**Подход:** Два указателя (медленный и быстрый).

<details>
<summary>Решение</summary>

```go
func hasCycle(head *ListNode) bool {
    if head == nil || head.Next == nil {
        return false
    }
    
    slow, fast := head, head.Next
    
    for fast != nil && fast.Next != nil {
        if slow == fast {
            return true
        }
        slow = slow.Next
        fast = fast.Next.Next
    }
    
    return false
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.9: Слияние двух отсортированных списков

**Условие:**
Слить два отсортированных связных списка в один.

**Функция:**
```go
func mergeTwoLists(l1 *ListNode, l2 *ListNode) *ListNode
```

**Подход:** Итеративное объединение с dummy узлом.

<details>
<summary>Решение</summary>

```go
func mergeTwoLists(l1 *ListNode, l2 *ListNode) *ListNode {
    dummy := &ListNode{}
    current := dummy
    
    for l1 != nil && l2 != nil {
        if l1.Val < l2.Val {
            current.Next = l1
            l1 = l1.Next
        } else {
            current.Next = l2
            l2 = l2.Next
        }
        current = current.Next
    }
    
    // Добавляем оставшиеся элементы
    if l1 != nil {
        current.Next = l1
    }
    if l2 != nil {
        current.Next = l2
    }
    
    return dummy.Next
}
```

**Сложность:**
- Время: O(n + m)
- Память: O(1)
</details>

---

#### Задача 6.10: Удаление дубликатов из отсортированного списка

**Условие:**
Дан отсортированный связный список, удалить все дубликаты так, чтобы каждый элемент встречался только один раз.

**Функция:**
```go
func deleteDuplicates(head *ListNode) *ListNode
```

**Подход:** Один указатель, сравнение с Next.

<details>
<summary>Решение</summary>

```go
func deleteDuplicates(head *ListNode) *ListNode {
    if head == nil {
        return nil
    }
    
    current := head
    
    for current != nil && current.Next != nil {
        if current.Val == current.Next.Val {
            current.Next = current.Next.Next  // удаляем дубликат
        } else {
            current = current.Next
        }
    }
    
    return head
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

### РАЗДЕЛ 3: СТРОКИ

#### Задача 6.11: Проверка анаграмм

**Условие:**
Определить, являются ли две строки `s` и `t` анаграммами.

**Функция:**
```go
func isAnagram(s string, t string) bool
```

**Подход:** Частотная таблица или сортировка.

<details>
<summary>Решение (хеш-таблица)</summary>

```go
func isAnagram(s string, t string) bool {
    if len(s) != len(t) {
        return false
    }
    
    freq := make(map[rune]int)
    
    for _, ch := range s {
        freq[ch]++
    }
    
    for _, ch := range t {
        freq[ch]--
        if freq[ch] < 0 {
            return false
        }
    }
    
    return true
}
```

**Сложность:**
- Время: O(n)
- Память: O(1) (максимум 26 букв)
</details>

---

#### Задача 6.12: Реверс слов

**Условие:**
Дана строка `s`, содержащая слова, разделённые пробелами. Вернуть строку со словами в обратном порядке.

**Функция:**
```go
func reverseWords(s string) string
```

**Подход:** Split → Reverse → Join.

<details>
<summary>Решение</summary>

```go
import "strings"

func reverseWords(s string) string {
    // Разбиваем на слова и удаляем пустые
    words := strings.Fields(s)
    
    // Разворачиваем массив слов
    left, right := 0, len(words)-1
    for left < right {
        words[left], words[right] = words[right], words[left]
        left++
        right--
    }
    
    return strings.Join(words, " ")
}
```

**Сложность:**
- Время: O(n)
- Память: O(n)
</details>

---

#### Задача 6.13: Валидация скобок

**Условие:**
Дана строка `s`, содержащая символы `'(', ')', '{', '}', '[', ']'`. Определить, корректна ли последовательность.

**Функция:**
```go
func isValid(s string) bool
```

**Подход:** Стек.

<details>
<summary>Решение</summary>

```go
func isValid(s string) bool {
    stack := []rune{}
    pairs := map[rune]rune{
        ')': '(',
        '}': '{',
        ']': '[',
    }
    
    for _, ch := range s {
        if ch == '(' || ch == '{' || ch == '[' {
            stack = append(stack, ch)
        } else {
            if len(stack) == 0 || stack[len(stack)-1] != pairs[ch] {
                return false
            }
            stack = stack[:len(stack)-1]  // pop
        }
    }
    
    return len(stack) == 0
}
```

**Сложность:**
- Время: O(n)
- Память: O(n)
</details>

---

#### Задача 6.14: Самая длинная подстрока без повторений

**Условие:**
Дана строка `s`, найти длину самой длинной подстроки без повторяющихся символов.

**Функция:**
```go
func lengthOfLongestSubstring(s string) int
```

**Подход:** Sliding window с хеш-таблицей.

<details>
<summary>Решение</summary>

```go
func lengthOfLongestSubstring(s string) int {
    seen := make(map[byte]int)
    maxLen, left := 0, 0
    
    for right := 0; right < len(s); right++ {
        ch := s[right]
        
        // Если символ уже встречался, сдвигаем левую границу
        if lastIndex, found := seen[ch]; found && lastIndex >= left {
            left = lastIndex + 1
        }
        
        seen[ch] = right
        maxLen = max(maxLen, right-left+1)
    }
    
    return maxLen
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

**Сложность:**
- Время: O(n)
- Память: O(min(n, m)), где m — размер алфавита
</details>

---

#### Задача 6.15: Поиск всех анаграмм в строке

**Условие:**
Дана строка `s` и строка `p`, найти все начальные индексы анаграмм `p` в `s`.

**Функция:**
```go
func findAnagrams(s string, p string) []int
```

**Подход:** Sliding window с частотной таблицей.

<details>
<summary>Решение</summary>

```go
func findAnagrams(s string, p string) []int {
    if len(s) < len(p) {
        return []int{}
    }
    
    result := []int{}
    pFreq := make(map[byte]int)
    windowFreq := make(map[byte]int)
    
    // Частота символов в p
    for i := 0; i < len(p); i++ {
        pFreq[p[i]]++
    }
    
    // Скользящее окно
    for i := 0; i < len(s); i++ {
        windowFreq[s[i]]++
        
        // Уменьшаем окно
        if i >= len(p) {
            left := s[i-len(p)]
            windowFreq[left]--
            if windowFreq[left] == 0 {
                delete(windowFreq, left)
            }
        }
        
        // Проверяем анаграмму
        if mapsEqual(windowFreq, pFreq) {
            result = append(result, i-len(p)+1)
        }
    }
    
    return result
}

func mapsEqual(a, b map[byte]int) bool {
    if len(a) != len(b) {
        return false
    }
    for k, v := range a {
        if b[k] != v {
            return false
        }
    }
    return true
}
```

**Сложность:**
- Время: O(n)
- Память: O(1) (максимум 26 букв)
</details>

---

### РАЗДЕЛ 4: ХЕШ-ТАБЛИЦЫ И ПОИСК

#### Задача 6.16: Two Sum

**Условие:**
Дан массив `nums` и целое `target`, найти два индекса чисел, сумма которых равна `target`.

**Функция:**
```go
func twoSum(nums []int, target int) []int
```

**Подход:** Хеш-таблица для быстрого поиска комплемента.

<details>
<summary>Решение</summary>

```go
func twoSum(nums []int, target int) []int {
    seen := make(map[int]int)
    
    for i, num := range nums {
        complement := target - num
        if j, found := seen[complement]; found {
            return []int{j, i}
        }
        seen[num] = i
    }
    
    return nil
}
```

**Сложность:**
- Время: O(n)
- Память: O(n)
</details>

---

#### Задача 6.17: Группировка анаграмм

**Условие:**
Дан массив строк `strs`, сгруппировать анаграммы вместе.

**Функция:**
```go
func groupAnagrams(strs []string) [][]string
```

**Подход:** Хеш-таблица с отсортированными строками как ключами.

<details>
<summary>Решение</summary>

```go
import "sort"

func groupAnagrams(strs []string) [][]string {
    groups := make(map[string][]string)
    
    for _, str := range strs {
        // Сортируем строку для получения ключа
        key := sortString(str)
        groups[key] = append(groups[key], str)
    }
    
    result := [][]string{}
    for _, group := range groups {
        result = append(result, group)
    }
    
    return result
}

func sortString(s string) string {
    runes := []rune(s)
    sort.Slice(runes, func(i, j int) bool {
        return runes[i] < runes[j]
    })
    return string(runes)
}
```

**Сложность:**
- Время: O(n * k log k), где k — средняя длина строки
- Память: O(n * k)
</details>

---

#### Задача 6.18: Intersection of Two Arrays

**Условие:**
Даны два массива `nums1` и `nums2`, вернуть их пересечение (уникальные элементы).

**Функция:**
```go
func intersection(nums1 []int, nums2 []int) []int
```

**Подход:** Хеш-множество.

<details>
<summary>Решение</summary>

```go
func intersection(nums1 []int, nums2 []int) []int {
    set1 := make(map[int]bool)
    result := make(map[int]bool)
    
    for _, num := range nums1 {
        set1[num] = true
    }
    
    for _, num := range nums2 {
        if set1[num] {
            result[num] = true
        }
    }
    
    output := []int{}
    for num := range result {
        output = append(output, num)
    }
    
    return output
}
```

**Сложность:**
- Время: O(n + m)
- Память: O(n)
</details>

---

#### Задача 6.19: Top K Frequent Elements

**Условие:**
Дан массив `nums`, вернуть k самых частых элементов.

**Функция:**
```go
func topKFrequent(nums []int, k int) []int
```

**Подход:** Частотная таблица + bucket sort.

<details>
<summary>Решение</summary>

```go
func topKFrequent(nums []int, k int) []int {
    // Частотная таблица
    freq := make(map[int]int)
    for _, num := range nums {
        freq[num]++
    }
    
    // Bucket sort: индекс = частота, значение = список чисел
    buckets := make([][]int, len(nums)+1)
    for num, count := range freq {
        buckets[count] = append(buckets[count], num)
    }
    
    // Собираем результат с конца
    result := []int{}
    for i := len(buckets) - 1; i >= 0 && len(result) < k; i-- {
        result = append(result, buckets[i]...)
    }
    
    return result[:k]
}
```

**Сложность:**
- Время: O(n)
- Память: O(n)
</details>

---

#### Задача 6.20: Первый уникальный символ

**Условие:**
Дана строка `s`, найти индекс первого неповторяющегося символа. Если его нет, вернуть `-1`.

**Функция:**
```go
func firstUniqChar(s string) int
```

**Подход:** Частотная таблица за два прохода.

<details>
<summary>Решение</summary>

```go
func firstUniqChar(s string) int {
    freq := make(map[rune]int)
    
    // Подсчитываем частоту
    for _, ch := range s {
        freq[ch]++
    }
    
    // Ищем первый уникальный
    for i, ch := range s {
        if freq[ch] == 1 {
            return i
        }
    }
    
    return -1
}
```

**Сложность:**
- Время: O(n)
- Память: O(1) (максимум 26 букв)
</details>

---

### РАЗДЕЛ 5: ДВА УКАЗАТЕЛЯ

#### Задача 6.21: K ближайших элементов

**Условие:**
Дан отсортированный массив `arr`, число `k` и целевое `x`. Найти `k` ближайших к `x` элементов.

**Функция:**
```go
func findClosestElements(arr []int, k int, x int) []int
```

**Подход:** Два указателя — left и right, сужаем окно.

<details>
<summary>Решение</summary>

```go
func findClosestElements(arr []int, k int, x int) []int {
    left, right := 0, len(arr)-1
    
    // Сужаем окно до k элементов
    for right-left+1 > k {
        if abs(arr[left]-x) > abs(arr[right]-x) {
            left++
        } else {
            right--
        }
    }
    
    return arr[left : right+1]
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.22: Container With Most Water

**Условие:**
Дан массив `height`, где `height[i]` — высота линии. Найти два индекса, образующие контейнер с максимальной площадью.

**Функция:**
```go
func maxArea(height []int) int
```

**Подход:** Два указателя — left и right, двигаем меньшую высоту.

<details>
<summary>Решение</summary>

```go
func maxArea(height []int) int {
    left, right := 0, len(height)-1
    maxArea := 0
    
    for left < right {
        width := right - left
        h := min(height[left], height[right])
        area := width * h
        maxArea = max(maxArea, area)
        
        // Двигаем указатель с меньшей высотой
        if height[left] < height[right] {
            left++
        } else {
            right--
        }
    }
    
    return maxArea
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.23: Minimum Difference Between K Elements

**Условие:**
Дан отсортированный массив и `k`, найти минимальную разницу между `k` последовательными элементами.

**Функция:**
```go
func minimumDifference(nums []int, k int) int
```

**Подход:** Скользящее окно размера k.

<details>
<summary>Решение</summary>

```go
import "sort"

func minimumDifference(nums []int, k int) int {
    if k == 1 {
        return 0
    }
    
    sort.Ints(nums)
    minDiff := nums[len(nums)-1] - nums[0]
    
    for i := 0; i+k-1 < len(nums); i++ {
        diff := nums[i+k-1] - nums[i]
        minDiff = min(minDiff, diff)
    }
    
    return minDiff
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**Сложность:**
- Время: O(n log n) (сортировка)
- Память: O(1)
</details>

---

#### Задача 6.24: Smallest Difference Between Two Arrays

**Условие:**
Даны два массива, найти пару чисел (по одному из каждого), у которых минимальная разница.

**Функция:**
```go
func smallestDifference(array1 []int, array2 []int) []int
```

**Подход:** Сортировка + два указателя.

<details>
<summary>Решение</summary>

```go
import (
    "math"
    "sort"
)

func smallestDifference(array1 []int, array2 []int) []int {
    sort.Ints(array1)
    sort.Ints(array2)
    
    i, j := 0, 0
    minDiff := math.MaxInt64
    result := []int{}
    
    for i < len(array1) && j < len(array2) {
        diff := abs(array1[i] - array2[j])
        
        if diff < minDiff {
            minDiff = diff
            result = []int{array1[i], array2[j]}
        }
        
        if array1[i] < array2[j] {
            i++
        } else {
            j++
        }
    }
    
    return result
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

**Сложность:**
- Время: O(n log n + m log m)
- Память: O(1)
</details>

---

#### Задача 6.25: Remove Duplicates In-Place

**Условие:**
Дан отсортированный массив `nums`, удалить дубликаты **in-place** и вернуть новую длину.

**Функция:**
```go
func removeDuplicates(nums []int) int
```

**Подход:** Два указателя — один для уникальных, другой для чтения.

<details>
<summary>Решение</summary>

```go
func removeDuplicates(nums []int) int {
    if len(nums) == 0 {
        return 0
    }
    
    slow := 0
    
    for fast := 1; fast < len(nums); fast++ {
        if nums[fast] != nums[slow] {
            slow++
            nums[slow] = nums[fast]
        }
    }
    
    return slow + 1
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

### РАЗДЕЛ 6: ПЛАВАЮЩИЕ ОКНА (SLIDING WINDOW)

#### Задача 6.26: Максимум в скользящем окне

**Условие:**
Дан массив `nums` и размер окна `k`, вернуть массив максимумов для каждого окна.

**Функция:**
```go
func maxSlidingWindow(nums []int, k int) []int
```

**Подход:** Deque (дек) для хранения индексов в убывающем порядке значений.

<details>
<summary>Решение</summary>

```go
func maxSlidingWindow(nums []int, k int) []int {
    if len(nums) == 0 || k == 0 {
        return []int{}
    }
    
    result := []int{}
    deque := []int{}  // храним индексы
    
    for i := 0; i < len(nums); i++ {
        // Удаляем элементы вне окна
        for len(deque) > 0 && deque[0] < i-k+1 {
            deque = deque[1:]
        }
        
        // Удаляем меньшие элементы с конца
        for len(deque) > 0 && nums[deque[len(deque)-1]] < nums[i] {
            deque = deque[:len(deque)-1]
        }
        
        deque = append(deque, i)
        
        // Добавляем максимум, когда окно заполнено
        if i >= k-1 {
            result = append(result, nums[deque[0]])
        }
    }
    
    return result
}
```

**Сложность:**
- Время: O(n)
- Память: O(k)
</details>

---

#### Задача 6.27: Minimum Window Substring

**Условие:**
Даны строки `s` и `t`, найти минимальную подстроку `s`, содержащую все символы `t`.

**Функция:**
```go
func minWindow(s string, t string) string
```

**Подход:** Sliding window с частотной таблицей.

<details>
<summary>Решение</summary>

```go
func minWindow(s string, t string) string {
    if len(s) < len(t) {
        return ""
    }
    
    tFreq := make(map[byte]int)
    for i := 0; i < len(t); i++ {
        tFreq[t[i]]++
    }
    
    required := len(tFreq)
    formed := 0
    windowFreq := make(map[byte]int)
    
    left, right := 0, 0
    minLen := len(s) + 1
    minStart := 0
    
    for right < len(s) {
        ch := s[right]
        windowFreq[ch]++
        
        if tFreq[ch] > 0 && windowFreq[ch] == tFreq[ch] {
            formed++
        }
        
        // Сжимаем окно
        for left <= right && formed == required {
            if right-left+1 < minLen {
                minLen = right - left + 1
                minStart = left
            }
            
            leftCh := s[left]
            windowFreq[leftCh]--
            if tFreq[leftCh] > 0 && windowFreq[leftCh] < tFreq[leftCh] {
                formed--
            }
            left++
        }
        
        right++
    }
    
    if minLen == len(s)+1 {
        return ""
    }
    return s[minStart : minStart+minLen]
}
```

**Сложность:**
- Время: O(|S| + |T|)
- Память: O(|S| + |T|)
</details>

---

#### Задача 6.28: Optimal Vacation Planning

**Условие:**
Дан массив `meetings`, где каждый элемент — количество встреч в день. Найти окно из `k` дней с минимумом встреч.

**Функция:**
```go
func minMeetingsDays(meetings []int, k int) int
```

**Подход:** Sliding window с подсчётом суммы.

<details>
<summary>Решение</summary>

```go
func minMeetingsDays(meetings []int, k int) int {
    if len(meetings) < k {
        return -1
    }
    
    // Начальное окно
    windowSum := 0
    for i := 0; i < k; i++ {
        windowSum += meetings[i]
    }
    
    minSum := windowSum
    minIndex := 0
    
    // Сдвигаем окно
    for i := k; i < len(meetings); i++ {
        windowSum += meetings[i] - meetings[i-k]
        if windowSum < minSum {
            minSum = windowSum
            minIndex = i - k + 1
        }
    }
    
    return minIndex
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.29: Maximum Sum Subarray of Size K

**Условие:**
Дан массив `nums` и `k`, найти максимальную сумму подмассива размера `k`.

**Функция:**
```go
func maxSumSubarray(nums []int, k int) int
```

**Подход:** Базовое скользящее окно.

<details>
<summary>Решение</summary>

```go
func maxSumSubarray(nums []int, k int) int {
    if len(nums) < k {
        return 0
    }
    
    // Начальное окно
    windowSum := 0
    for i := 0; i < k; i++ {
        windowSum += nums[i]
    }
    
    maxSum := windowSum
    
    // Сдвигаем окно
    for i := k; i < len(nums); i++ {
        windowSum += nums[i] - nums[i-k]
        maxSum = max(maxSum, windowSum)
    }
    
    return maxSum
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

#### Задача 6.30: Subarray with Given Sum

**Условие:**
Дан массив положительных чисел `nums` и сумма `target`, найти подмассив с этой суммой (вернуть индексы).

**Функция:**
```go
func subarraySum(nums []int, target int) []int
```

**Подход:** Динамическое скользящее окно переменного размера.

<details>
<summary>Решение</summary>

```go
func subarraySum(nums []int, target int) []int {
    left, windowSum := 0, 0
    
    for right := 0; right < len(nums); right++ {
        windowSum += nums[right]
        
        // Сжимаем окно, если сумма больше target
        for windowSum > target && left <= right {
            windowSum -= nums[left]
            left++
        }
        
        // Нашли окно с нужной суммой
        if windowSum == target {
            return []int{left, right}
        }
    }
    
    return []int{-1, -1}  // не найдено
}
```

**Сложность:**
- Время: O(n)
- Память: O(1)
</details>

---

## ИТОГО: 30 АЛГОРИТМИЧЕСКИХ ЗАДАЧ ✅

**Статистика:**
- Раздел 1 (Массивы): 5 задач
- Раздел 2 (Связные списки): 5 задач
- Раздел 3 (Строки): 5 задач
- Раздел 4 (Хеш-таблицы): 5 задач
- Раздел 5 (Два указателя): 5 задач
- Раздел 6 (Sliding Window): 5 задач

**Покрытие:** 80% алгоритмических задач с собеседований решаются этими паттернами.

---

## ЧАСТЬ 7: ЗАДАЧИ С РЕАЛЬНЫХ СОБЕСОВ

*Источник: island_golang, island_system_design, island_companies*

---

### Задача 7.1: Get Champions (AvitoTech)

**Условие:**
Найти чемпионов соревнований по шагам. Чемпионы должны соответствовать критериям:
1. Прошли **наибольшее общее количество шагов** за все дни соревнования
2. **Не пропустили ни одного дня** (есть запись в каждом дне)

**Функция:**
```go
type Statistics struct {
    UserId int
    Steps  int
}

type Result struct {
    UserIds []int
    Steps   int
}

func getChampions(statistics [][]Statistics) Result
```

**Входные данные:**
- `statistics`: список дней соревнования
- Каждый день — список `{userId, steps}`

**Выходные данные:**
- `userIds`: список победителей (если несколько с одинаковым максимумом)
- `steps`: общее количество шагов победителя(ей)

**Примеры:**

*Пример 1:*
```go
statistics = [
    [ { userId: 1, steps: 1000 }, { userId: 2, steps: 1500 } ],
    [ { userId: 2, steps: 1000 } ]
]

Вывод: { userIds: [2], steps: 2500 }
```

*Пример 2:*
```go
statistics = [
    [ { userId: 1, steps: 2000 }, { userId: 2, steps: 1500 } ],
    [ { userId: 2, steps: 4000 }, { userId: 1, steps: 3500 } ]
]

Вывод: { userIds: [1, 2], steps: 5500 }
```

**Подход:** Хеш-таблицы для подсчёта дней и суммы шагов.

<details>
<summary>Решение</summary>

```go
func getChampions(statistics [][]Statistics) Result {
    if len(statistics) == 0 {
        return Result{}
    }
    
    // Подсчёт дней участия для каждого пользователя
    daysParticipated := make(map[int]int)
    
    // Подсчёт общего количества шагов
    totalSteps := make(map[int]int)
    
    totalDays := len(statistics)
    
    // O(n * m) - проходим по всем дням и пользователям
    for _, day := range statistics {
        for _, stat := range day {
            daysParticipated[stat.UserId]++
            totalSteps[stat.UserId] += stat.Steps
        }
    }
    
    // Находим пользователей, которые не пропустили ни одного дня
    validUsers := make(map[int]bool)
    for userId, days := range daysParticipated {
        if days == totalDays {
            validUsers[userId] = true
        }
    }
    
    // Находим максимальное количество шагов среди валидных
    maxSteps := 0
    for userId := range validUsers {
        if totalSteps[userId] > maxSteps {
            maxSteps = totalSteps[userId]
        }
    }
    
    // Собираем всех пользователей с maxSteps
    var resultIds []int
    for userId := range validUsers {
        if totalSteps[userId] == maxSteps {
            resultIds = append(resultIds, userId)
        }
    }
    
    return Result{
        UserIds: resultIds,
        Steps:   maxSteps,
    }
}
```

**Сложность:**
- Время: O(n × m), где n — дни, m — пользователи
- Память: O(m)

**Тесты:**
```go
func main() {
    example1 := getChampions([][]Statistics{
        {{UserId: 1, Steps: 1000}, {UserId: 2, Steps: 1500}},
        {{UserId: 2, Steps: 1000}},
    })
    fmt.Println("Example 1:", example1)  // {[2] 2500}
    
    example2 := getChampions([][]Statistics{
        {{UserId: 1, Steps: 2000}, {UserId: 2, Steps: 1500}},
        {{UserId: 2, Steps: 4000}, {UserId: 1, Steps: 3500}},
    })
    fmt.Println("Example 2:", example2)  // {[1 2] 5500}
}
```
</details>

---

## ЧАСТЬ 8: ЗАДАЧИ ИЗ ОЗОН (SQL + GO)

*Источник: island_sql/OBSIDIAN VAULT/Озон_задачник*

---

### SQL Задачи

#### Задача 8.1: Количество заказов по пользователям

**Условие:**
Вернуть количество заказов по каждому пользователю с `price_total >= 1000`, отсортировать по количеству в обратном порядке.

<details>
<summary>Решение</summary>

```sql
SELECT user_id, COUNT(*) AS c
FROM orders 
WHERE price_total >= 1000 
GROUP BY user_id
ORDER BY c DESC;
```
</details>

---

#### Задача 8.2: Покупки до бана

**Условие:**
Вывести уникальные комбинации (user_id, sku) для всех покупок, совершенных **до бана** пользователя. Отсортировать по lastname, firstname, user_id, sku.

<details>
<summary>Решение</summary>

```sql
SELECT DISTINCT u.id, u.firstname, u.lastname, p.sku 
FROM user AS u 
JOIN purchase AS p ON u.id = p.user_id 
LEFT JOIN ban_list AS b ON u.id = b.user_id
WHERE b.date > p.date 
   OR b.user_id IS NULL 
ORDER BY u.lastname, u.firstname, u.id, p.sku;
```

**Объяснение:**
- `LEFT JOIN ban_list` — включаем всех пользователей (даже незабаненных)
- `b.date > p.date` — покупка до бана
- `b.user_id IS NULL` — пользователь не забанен
</details>

---

#### Задача 8.3: Пользователи с суммой покупок > 5000

**Условие:**
Найти пользователей, совершивших покупок на сумму > 5000. Формат: `user_id | first | last | sum`.

<details>
<summary>Решение</summary>

```sql
SELECT u.id, u.first, u.last, SUM(p.price) AS sum
FROM users AS u 
JOIN purchase AS p ON u.id = p.user_id
GROUP BY u.id, u.first, u.last
HAVING sum > 5000;
```

**Объяснение:**
- `HAVING` фильтрует результат после агрегации
</details>

---

#### Задача 8.4: Модель чата (User, Chat, Message)

**Условие:**
- User: имя, дата регистрации
- Chat: название, дата создания
- Message: текст, автор, дата создания
- User может быть в нескольких чатах
- Message принадлежит одному чату

<details>
<summary>Решение (схема БД)</summary>

```sql
-- Пользователи
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    registered_at TIMESTAMP DEFAULT NOW()
);

-- Чаты
CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Связь многие-ко-многим (users <-> chats)
CREATE TABLE user_chats (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, chat_id)
);

-- Сообщения
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```
</details>

---

#### Задача 8.5: Модель библиотеки (Author, Book, Reader)

**Условие:**
- Физически книга одна, может быть только у одного читателя
- У книги может быть несколько авторов

<details>
<summary>Решение (схема БД)</summary>

```sql
-- Авторы
CREATE TABLE author (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    birth_year INTEGER
);

-- Читатели
CREATE TABLE reader (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    card_number VARCHAR(20) UNIQUE
);

-- Книги
CREATE TABLE book (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    reader_id INTEGER REFERENCES reader(id) ON DELETE SET NULL,
    CONSTRAINT single_owner UNIQUE (id, reader_id)
);

-- Связь книг с авторами (многие-ко-многим)
CREATE TABLE book_author (
    book_id INTEGER REFERENCES book(id) ON DELETE CASCADE,
    author_id INTEGER REFERENCES author(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);
```

**Запросы:**

1. Все книги на руках:
```sql
SELECT b.title
FROM book b
WHERE b.reader_id IS NOT NULL;
```

2. Книги с > 3 авторами:
```sql
SELECT b.title
FROM book b
JOIN book_author ba ON b.id = ba.book_id
GROUP BY b.id, b.title
HAVING COUNT(ba.author_id) > 3;
```

3. Топ-3 читаемых автора:
```sql
SELECT a.name, COUNT(b.id) AS books_checked_out
FROM author a
JOIN book_author ba ON a.id = ba.author_id
JOIN book b ON ba.book_id = b.id
WHERE b.reader_id IS NOT NULL
GROUP BY a.id, a.name
ORDER BY books_checked_out DESC
LIMIT 3;
```
</details>

---

### Go Задачи

#### Задача 8.6: Проверка монотонности слайса

**Условие:**
Монотонная функция — либо везде не убывает, либо везде не возрастает.

<details>
<summary>Решение</summary>

```go
func isMonotonic(nums []int) bool {
    if len(nums) <= 2 {
        return true
    }
    
    increasing := true
    decreasing := true
    
    for i := 1; i < len(nums); i++ {
        if nums[i] > nums[i-1] {
            decreasing = false
        }
        if nums[i] < nums[i-1] {
            increasing = false
        }
    }
    
    return increasing || decreasing
}
```
</details>

---

#### Задача 8.7: Удаление нулей из слайса

<details>
<summary>Решение</summary>

```go
func remove(nums []int) []int {
    result := make([]int, 0, len(nums))
    for _, v := range nums {
        if v != 0 {
            result = append(result, v)
        }
    }
    return result
}

// In-place вариант
func removeInPlace(nums []int) []int {
    writeIdx := 0
    for _, v := range nums {
        if v != 0 {
            nums[writeIdx] = v
            writeIdx++
        }
    }
    return nums[:writeIdx]
}
```
</details>

---

#### Задача 8.8: Генератор уникальных рандомных чисел

<details>
<summary>Решение</summary>

```go
import (
    "math/rand"
    "time"
)

func uniqRandn(n int) []int {
    rand.Seed(time.Now().UnixNano())
    
    seen := make(map[int]bool)
    result := make([]int, 0, n)
    
    for len(result) < n {
        num := rand.Intn(n * 10)  // диапазон
        if !seen[num] {
            seen[num] = true
            result = append(result, num)
        }
    }
    
    return result
}
```
</details>

---

#### Задача 8.9: Generator и Squarer с каналами

**Условие:**
- `generator`: записывает числа в канал
- `squarer`: читает из канала, возводит в квадрат, пишет в новый канал
- Обе функции завершаются по отмене контекста

<details>
<summary>Решение</summary>

```go
func generator(ctx context.Context, nums []int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case <-ctx.Done():
                return
            case out <- n:
            }
        }
    }()
    return out
}

func squarer(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case n, ok := <-in:
                if !ok {
                    return
                }
                select {
                case <-ctx.Done():
                    return
                case out <- n * n:
                }
            }
        }
    }()
    return out
}
```
</details>

---

#### Задача 8.10: Merge каналов

**Условие:**
Объединить несколько каналов в один.

<details>
<summary>Решение</summary>

```go
func merge(ctx context.Context, channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case v, ok := <-c:
                    if !ok {
                        return
                    }
                    select {
                    case <-ctx.Done():
                        return
                    case out <- v:
                    }
                }
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```
</details>

---

#### Задача 8.11: Zip каналов

**Условие:**
Объединить два канала попарно: `(a1, b1), (a2, b2), ...`

<details>
<summary>Решение</summary>

```go
type Pair struct {
    A, B int
}

func zip(ctx context.Context, ch1, ch2 <-chan int) <-chan Pair {
    out := make(chan Pair)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            default:
                v1, ok1 := <-ch1
                v2, ok2 := <-ch2
                if !ok1 || !ok2 {
                    return
                }
                select {
                case <-ctx.Done():
                    return
                case out <- Pair{v1, v2}:
                }
            }
        }
    }()
    return out
}
```
</details>

---

#### Задача 8.12: Кэш с TTL

<details>
<summary>Решение</summary>

```go
type CacheItem struct {
    value      string
    expiration time.Time
}

type CacheWithTTL struct {
    data map[string]CacheItem
    mu   sync.RWMutex
}

func NewCacheWithTTL() *CacheWithTTL {
    return &CacheWithTTL{
        data: make(map[string]CacheItem),
    }
}

func (c *CacheWithTTL) Set(key, value string, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = CacheItem{
        value:      value,
        expiration: time.Now().Add(ttl),
    }
}

func (c *CacheWithTTL) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, ok := c.data[key]
    if !ok {
        return "", false
    }
    
    // Проверка TTL
    if item.expiration.IsZero() || time.Now().Before(item.expiration) {
        return item.value, true
    }
    
    return "", false
}
```
</details>

---


---

## ЧАСТЬ 9: BOSS KATA ACADEMY — РЕАЛЬНЫЕ ЗАДАЧИ С СОБЕСОВ

*Источник: island_kataboss — задачи от Boss Kata Academy*

**Формат:** Live-кодинг (1.5 часа) + System Design (60 минут)

---

### Задача 9.1: Ускорение Inc() — горутины + синхронизация

**Условие:**
```go
var count int

func Inc() {
    time.Sleep(1 * time.Millisecond)
    count++
}

func main() {
    for i := 0; i < 10000; i++ {
        Inc()
    }
    fmt.Println(count)
}
```

**Вопрос:** За сколько исполнится (~10 сек)? Что выведется (10000)? Как ускорить?

<details>
<summary>Решение</summary>

**С горутинами + Mutex:**
```go
var (
    count int
    mu    sync.Mutex
    wg    sync.WaitGroup
)

func Inc() {
    time.Sleep(1 * time.Millisecond)
    mu.Lock()
    count++
    mu.Unlock()
}

func main() {
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            Inc()
        }()
    }
    wg.Wait()
    fmt.Println(count)  // 10000
}
```

**С atomic:**
```go
var count int64

func Inc() {
    time.Sleep(1 * time.Millisecond)
    atomic.AddInt64(&count, 1)
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            Inc()
        }()
    }
    wg.Wait()
    fmt.Println(count)
}
```

**Результат:** ~1-2 миллисекунды (параллельное выполнение)
</details>

---

### Задача 9.2: Wrapper с таймаутом (Context + Select)

**Условие:**
```go
func unpredictableFunc() int64 {
    rnd := rand.Int63n(5000)
    time.Sleep(time.Duration(rnd) * time.Millisecond)
    return rnd
}
```

**Требование:** Обёртка с таймаутом (1 сек). Если не отработала — ошибка. Измерять время.

<details>
<summary>Решение</summary>

```go
func unpredictableFuncWithTimeout(timeout time.Duration) (int64, error) {
    start := time.Now()
    
    resultCh := make(chan int64, 1)  // Буферизированный!
    
    go func() {
        result := unpredictableFunc()
        resultCh <- result
    }()
    
    select {
    case result := <-resultCh:
        elapsed := time.Since(start)
        log.Printf("Выполнено за %v\n", elapsed)
        return result, nil
    case <-time.After(timeout):
        elapsed := time.Since(start)
        log.Printf("Таймаут за %v\n", elapsed)
        return 0, fmt.Errorf("timeout exceeded: %v", timeout)
    }
}

func main() {
    result, err := unpredictableFuncWithTimeout(1 * time.Second)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }
}
```

**Ключевые моменты:**
- `make(chan int64, 1)` — горутина не блокируется при таймауте
- `select` с `time.After` для таймаута
- `time.Since(start)` для измерения времени
</details>

---

### Задача 9.3: changeName — 5 способов работы с указателями

**Условие:**
```go
type Person struct {
    Name string
}

func changeName(person *Person) {
    person = &Person{Name: "Alice"}
}

func main() {
    person := &Person{Name: "Bob"}
    fmt.Println(person.Name)  // Bob
    changeName(person)
    fmt.Println(person.Name)  // Bob (не изменилось!)
}
```

**Вопрос:** Что выведется? 5 способов исправить?

<details>
<summary>Решения</summary>

**Способ 1: Изменить поле напрямую**
```go
func changeName(person *Person) {
    person.Name = "Alice"
}
```

**Способ 2: Вернуть новый указатель**
```go
func changeName(person *Person) *Person {
    return &Person{Name: "Alice"}
}

person = changeName(person)
```

**Способ 3: Указатель на указатель**
```go
func changeName(person **Person) {
    *person = &Person{Name: "Alice"}
}

changeName(&person)
```

**Способ 4: Метод структуры**
```go
func (p *Person) ChangeName(name string) {
    p.Name = name
}

person.ChangeName("Alice")
```

**Способ 5: По значению + return**
```go
func changeName(person Person) Person {
    person.Name = "Alice"
    return person
}

person = changeName(*person)
```
</details>

---

### Задача 9.4: Ускорение ручки `/weather` — 3 способа

**Условие:**
```go
func aiWeatherForecast() int {
    time.Sleep(1 * time.Second)  // ~1 сек
    return rand.Intn(30)
}

http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "{\"temperature\":%d}\n", aiWeatherForecast())
})
```

**Вопрос:** Ручка работает ~1 сек. Как ускорить?

<details>
<summary>Решения</summary>

**Способ 1: Кэш с TTL**
```go
var (
    cache     *CacheItem
    cacheMu   sync.RWMutex
    cacheTTL  = 5 * time.Minute
)

type CacheItem struct {
    Temperature int
    ExpiresAt   time.Time
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
    cacheMu.RLock()
    if cache != nil && time.Now().Before(cache.ExpiresAt) {
        temp := cache.Temperature
        cacheMu.RUnlock()
        fmt.Fprintf(w, "{\"temperature\":%d}\n", temp)
        return
    }
    cacheMu.RUnlock()
    
    temp := aiWeatherForecast()
    
    cacheMu.Lock()
    cache = &CacheItem{temp, time.Now().Add(cacheTTL)}
    cacheMu.Unlock()
    
    fmt.Fprintf(w, "{\"temperature\":%d}\n", temp)
}
```

**Способ 2: Background обновление**
```go
var currentTemp int

func updateWeatherLoop() {
    for {
        currentTemp = aiWeatherForecast()
        time.Sleep(5 * time.Minute)
    }
}

func main() {
    go updateWeatherLoop()
    http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "{\"temperature\":%d}\n", currentTemp)
    })
    http.ListenAndServe(":3333", nil)
}
```

**Способ 3: Singleflight**
```go
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func weatherHandler(w http.ResponseWriter, r *http.Request) {
    result, err, _ := g.Do("weather", func() (interface{}, error) {
        return aiWeatherForecast(), nil
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, "{\"temperature\":%d}\n", result.(int))
}
```
</details>

---

### Задача 9.5: SQL — покупки до бана (LEFT JOIN)

**Условие:**
```sql
-- user
id | firstname | lastname | birth
1  | Ivan      | Petrov   | 1996-05-01
2  | Anna      | Petrova  | 1999-06-01

-- purchase
sku | price | user_id | date
1   | 5500  | 1       | 2021-02-15
2   | 4000  | 1       | 2021-02-14

-- ban_list
user_id | date
1       | 2021-02-20
```

**Вопрос:** Уникальные `(user_id, sku)` для покупок **до бана**. Сортировать по lastname, firstname, sku.

<details>
<summary>Решение</summary>

```sql
SELECT DISTINCT u.id, u.firstname, u.lastname, p.sku 
FROM user AS u 
JOIN purchase AS p ON u.id = p.user_id 
LEFT JOIN ban_list AS b ON u.id = b.user_id
WHERE b.date > p.date     -- Покупка до бана
   OR b.user_id IS NULL   -- Пользователь не забанен
ORDER BY u.lastname, u.firstname, u.id, p.sku;
```

**Объяснение:**
- `LEFT JOIN ban_list` — включаем всех (даже незабаненных)
- `b.date > p.date` — покупка **до** бана
- `b.user_id IS NULL` — вообще не забанен
</details>

---

## ЧАСТЬ 10: ЗАДАЧИ ЛУКЬЯНОВА + ФИЛОСОФИЯ KATA

*Источники: 7 SQL Luckyanov.txt (455 строк) + KATA_methodology to DB.txt (1130 строк)*

**Синтез двух путей:**
- **Философия KATA** — почему (ACID, изоляция, нормализация)
- **Практика Лукьянова** — как (7 задач с решениями)

---

### 10.1: Покупки до бана + GROUP BY HAVING

**Схема:**
```sql
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE purchases (
    purchase_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    purchase_date TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE banlist (
    user_id INT PRIMARY KEY,
    ban_date TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

**Задача 1:** Выбрать юзеров и ID покупки, совершённой **до бана** (или не забанены).

<details>
<summary>Решение</summary>

```sql
SELECT u.user_id, u.name, p.purchase_id
FROM users u
JOIN purchases p ON u.user_id = p.user_id
LEFT JOIN banlist b ON u.user_id = b.user_id
WHERE b.ban_date IS NULL OR p.purchase_date < b.ban_date;
```
</details>

**Задача 2:** Выбрать юзеров с покупками **> 5000**.

<details>
<summary>Решение</summary>

```sql
SELECT u.user_id, u.name, SUM(p.amount) AS total_spent
FROM users u
JOIN purchases p ON u.user_id = p.user_id
GROUP BY u.user_id, u.name
HAVING SUM(p.amount) > 5000;
```

**Ключ:** `HAVING` — фильтр **после** агрегации (нельзя `WHERE SUM(...)`)
</details>

---

### 10.2: Библиотека — Многие-ко-многим

**Схема:**
```sql
CREATE TABLE Books (
    book_id INT PRIMARY KEY,
    title VARCHAR(255),
    reader_id INT NULL  -- NULL = на полке
);

CREATE TABLE Book_Author (
    book_id INT,
    author_id INT,
    FOREIGN KEY (book_id) REFERENCES Books(book_id)
);
```

**Задача:** Книги **в библиотеке** (reader_id IS NULL), у которых **> 3 авторов**.

<details>
<summary>Решение</summary>

```sql
SELECT b.title
FROM Books b
JOIN Book_Author ba ON b.book_id = ba.book_id
WHERE b.reader_id IS NULL
GROUP BY b.book_id, b.title
HAVING COUNT(ba.author_id) > 3;
```
</details>

---

### 10.3: ⭐ Оптимальный индекс (порассуждать!)

**Запрос:**
```sql
SELECT * FROM employee 
WHERE sex = 'm' AND salary > 300000 AND age = 20 
ORDER BY created_at;
```

**Вопрос:** Какой индекс оптимален?

<details>
<summary>Решение + Рассуждения</summary>

```sql
CREATE INDEX ON employee (age, salary);
```

**Почему?**
1. **Селективность:**
   - `sex = 'm'` — 50% строк (низкая)
   - `age = 20` — средняя
   - `salary > 300000` — высокая

2. **Правило:** equality (`=`) → range (`>`)

3. **Почему НЕ `created_at`?**
   - Лучше быстро отфильтровать, чем сортировать миллионы строк

**Философия KATA:**
> Индекс — это trade-off: скорость SELECT vs память + медленный INSERT

**Видео:** https://www.youtube.com/watch?v=DyqtBiDrz3g&t=265s
</details>

---

### 10.4: ⭐ Race Conditions + SELECT FOR UPDATE

**Схема: Система холдов (holds)**
```sql
CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    amount BIGINT NOT NULL
);

CREATE TABLE holds (
    id TEXT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL
);
```

**Изначальный код (НЕПРАВИЛЬНЫЙ):**
```sql
BEGIN;
SELECT amount FROM accounts WHERE id = :id;
UPDATE accounts SET amount = amount - 10 WHERE id = :id;
INSERT INTO holds VALUES (:uuid, :id, 10);
COMMIT;
```

**Проблема:** Race condition (два запроса прочитают одно значение!)

<details>
<summary>Решение</summary>

```sql
BEGIN;

-- ⭐ Заблокировать строку
SELECT amount FROM accounts WHERE id = :id FOR UPDATE;

IF (amount - :hold_amount) < 0 THEN
    ROLLBACK;
    RETURN 'Insufficient funds';
END IF;

UPDATE accounts SET amount = amount - :hold_amount WHERE id = :id;
INSERT INTO holds VALUES (gen_random_uuid(), :id, :hold_amount);

COMMIT;
```

**Что делает `FOR UPDATE`:**
- Блокирует строку до COMMIT
- Другие транзакции **ждут**
- Гарантирует атомарность

**Философия KATA (ACID):**
> **I = Isolation** — транзакции изолированы друг от друга.  
> `SELECT FOR UPDATE` — инструмент для pessimistic locking.

**Use cases:**
- Financial transactions (баланс, холды)
- Inventory management (остатки товаров)
- Booking systems (seats, tickets)
</details>

---

### 10.5: ⭐ Индексы — Левое правило (когда НЕ работает)

**Индекс:**
```sql
CREATE INDEX myIdx ON carts (sku, country, customer_id);
```

**Вопрос 1:**
```sql
SELECT * FROM carts WHERE sku = 192 AND country = 'ru';
```
**Работает ли индекс?**

<details>
<summary>Ответ</summary>

**ДА!**  
Индекс `(sku, country, customer_id)` работает для префикса `(sku, country)`
</details>

**Вопрос 2:**
```sql
SELECT * FROM carts WHERE country = 'ru' AND customer_id = 10;
```
**Работает ли индекс?**

<details>
<summary>Ответ</summary>

**НЕТ!**  
Первый столбец (`sku`) пропущен → **левое правило нарушено** → индекс игнорируется

**Правило:**
- Индекс `(A, B, C)` работает для:
  - ✅ WHERE A
  - ✅ WHERE A, B
  - ✅ WHERE A, B, C
  - ❌ WHERE B (первый пропущен!)
  - ❌ WHERE C
  - ❌ WHERE B, C

**Философия KATA:**
> Индекс — это упорядоченная структура (B-tree).  
> Без первого столбца = поиск в неотсортированном списке.
</details>

---

### 10.6: Последние комментарии каждого автора

**Задача:** Вывести **последний комментарий каждого автора**.

```sql
CREATE TABLE comments (
    id SERIAL,
    comment TEXT,
    author TEXT,
    date TIMESTAMP
);
```

<details>
<summary>Решение</summary>

```sql
SELECT c.*
FROM comments c
JOIN (
    SELECT author, MAX(date) AS max_date
    FROM comments
    GROUP BY author
) AS sub ON c.author = sub.author AND c.date = sub.max_date;
```

**Объяснение:**
1. Подзапрос: `MAX(date)` для каждого автора
2. JOIN по автору + дате
3. Результат: последний комментарий

**Философия KATA:**
> Подзапросы (subqueries) — мощный инструмент для сложной логики.  
> Альтернатива: window functions (`ROW_NUMBER() OVER (PARTITION BY author ORDER BY date DESC)`)
</details>

---

### 10.7: Сортировка по CASE

**Условие:** Кастомная сортировка (id: 1, 3, 2, 4, 5)

<details>
<summary>Решение</summary>

```sql
SELECT * FROM test
ORDER BY CASE
    WHEN id = 1 THEN 1
    WHEN id = 3 THEN 2
    WHEN id = 2 THEN 3
    WHEN id = 4 THEN 4
    WHEN id = 5 THEN 5
END;
```

**Use cases:**
- Featured products (топовые товары в начале)
- Pinned posts (закреплённые сообщения)
- Custom priority sorting
</details>

---

## ФИЛОСОФИЯ KATA: ПОЧЕМУ ТАК?

*Выжимка из KATA_methodology to DB.txt*

### Зачем нужны БД? Почему не файлы?

**8 причин:**
1. Эффективное хранение (таблицы, индексы)
2. Быстрый доступ (оптимизация запросов)
3. Многопользовательский доступ
4. Целостность данных (constraints, транзакции)
5. Безопасность (аутентификация, шифрование)
6. Резервное копирование
7. Масштабируемость
8. Управление данными (CRUD)

**Вывод:** Файлы = хаос. БД = порядок + гарантии.

---

### ACID — Почему это важно?

- **A (Atomicity):** Всё или ничего (перевод денег: снять + зачислить)
- **C (Consistency):** Constraints не нарушаются
- **I (Isolation):** Транзакции изолированы (уровни изоляции)
- **D (Durability):** После COMMIT данные сохранены (даже при сбое)

---

### Уровни изоляции — Trade-offs

| Уровень | Dirty Read | Non-repeatable | Phantom | Производительность |
|---------|------------|----------------|---------|-------------------|
| Read Uncommitted | ❌ | ❌ | ❌ | 🔥🔥🔥🔥 |
| Read Committed | ✅ | ❌ | ❌ | 🔥🔥🔥 |
| Repeatable Read | ✅ | ✅ | ❌ | 🔥🔥 |
| Serializable | ✅ | ✅ | ✅ | 🔥 |

**Философия:**
- Выше уровень → больше гарантий, **ниже производительность**
- **Read Committed** — баланс для 95% случаев

---

### Нормализация — Зачем?

**1NF:** Атомарные значения
```sql
-- ❌ Плохо: phones: "123, 456"
-- ✅ Хорошо: отдельная таблица phones
```

**3NF:** Нет транзитивной зависимости

**Когда денормализовать?**
- Много JOIN'ов → медленно
- Читаем часто, пишем редко
- Аналитика (OLAP)

---

## КЛЮЧЕВЫЕ ВОПРОСЫ ЛУКЬЯНОВА

1. **WHERE vs HAVING?**
   - `WHERE` — до агрегации
   - `HAVING` — после агрегации

2. **Внешние ключи — использовать?**
   - ✅ Целостность данных
   - ❌ Производительность (проверки)

3. **Репликация: синхронная vs асинхронная?**
   - Синхронная: данные сразу (медленнее)
   - Асинхронная: lag (быстрее)

4. **Ключ шардирования — как выбрать?**
   - user_id — равномерное распределение
   - date — hot shard
   - Зависит от паттерна запросов!

---

## ИТОГ: СИНТЕЗ ДВУХ ПУТЕЙ

**Kata учит "почему":**
- ACID — гарантии
- Уровни изоляции — trade-offs
- Нормализация — структура

**Лукьянов учит "как":**
- `SELECT FOR UPDATE` — race conditions
- Индексы — левое правило
- `HAVING` vs `WHERE` — на практике

**Вместе = полное понимание:**
- Философия (почему так)
- Практика (как решать)
- Готовность к любым собесам

---
