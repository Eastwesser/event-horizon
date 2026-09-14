# 🎫 Модуль 1: Rate Limiter + Базовые алгоритмы

**Цель:** Научиться ограничивать нагрузку на систему и защищать от DDoS.

---

## 📋 Билет (Темы модуля)

1. **Разминка:** Замыкания в горутинах (a, b, c вывод)
2. **LeetCode:** Two Sum (LeetCode 1)
3. **Конкурентность:** Ограничение одновременных запросов (50 URL, ≤5 одновременно)
4. **Паттерн:** Token Bucket Rate Limiter (`Allow() bool`)
5. **SQL:** Поиск пользователей с >100 запросов/час
6. **Проект:** Rate limiting в Roolz/Adtime

---

## 🎯 Задачи (в порядке выполнения)

### 1. Разминка: Замыкания в горутинах
**Папка:** `1.task_warmup/`

**Задача:**
```go
// Что выведет этот код? Почему?
func main() {
    for i := 0; i < 3; i++ {
        go func() {
            fmt.Println(i)
        }()
    }
    time.Sleep(time.Second)
}
```

**Правильный ответ:** `3 3 3` (или в другом порядке)

**Почему?**
- Замыкание захватывает **переменную** `i`, а не **значение**
- Когда горутины запускаются, цикл уже завершён → `i = 3`

**Как исправить:**
```go
for i := 0; i < 3; i++ {
    go func(id int) {  // Передаём значение!
        fmt.Println(id)
    }(i)
}
```

---

### 2. LeetCode: Two Sum
**Папка:** `2.task_leetcode/`

**Задача:** [LeetCode 1 - Two Sum](https://leetcode.com/problems/two-sum/)

Дан массив `nums` и число `target`. Найди два индекса, сумма которых равна `target`.

**Пример:**
```
Input: nums = [2,7,11,15], target = 9
Output: [0,1] (nums[0] + nums[1] = 2 + 7 = 9)
```

**Решение (O(n)):**
```go
func twoSum(nums []int, target int) []int {
    seen := make(map[int]int)
    
    for i, num := range nums {
        complement := target - num
        if j, ok := seen[complement]; ok {
            return []int{j, i}
        }
        seen[num] = i
    }
    
    return nil
}
```

---

### 3. Конкурентность: Ограничение одновременных запросов
**Папка:** `3.task_concurrency/`

**Задача:** 50 URL, но не больше 5 запросов одновременно.

**Решение: Semaphore**
```go
func fetchURLs(urls []string) {
    sem := make(chan struct{}, 5)  // Семафор на 5 слотов
    var wg sync.WaitGroup
    
    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            
            sem <- struct{}{}  // Взять слот
            defer func() { <-sem }()  // Освободить слот
            
            resp, _ := http.Get(u)
            fmt.Println(resp.Status)
        }(url)
    }
    
    wg.Wait()
}
```

---

### 4. Паттерн: Token Bucket Rate Limiter
**Папка:** `4.task_pattern/`

**Подпапки:**
- `1.fixed_window/` — Фиксированное окно (100 req/min)
- `2.sliding_window/` — Скользящее окно
- `3.token_bucket/` — Token Bucket (основной!)
- `4.leaky_bucket/` — Leaky Bucket

**Token Bucket — основной паттерн:**
```go
type TokenBucket struct {
    mu         sync.Mutex
    tokens     int
    capacity   int
    refillRate int  // токенов в секунду
    lastRefill time.Time
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    // Пополнить токены
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill).Seconds()
    newTokens := int(elapsed * float64(tb.refillRate))
    
    tb.tokens = min(tb.capacity, tb.tokens + newTokens)
    tb.lastRefill = now
    
    // Проверить доступность
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    return false
}
```

**Use case:** Защита API от DDoS.

---

### 5. SQL: Поиск пользователей с >100 запросов/час
**Папка:** `5.task_sql/`

**Задача:**
```sql
-- Таблица requests: user_id, timestamp
-- Найти пользователей с >100 запросов в последний час

SELECT user_id, COUNT(*) as request_count
FROM requests
WHERE timestamp >= NOW() - INTERVAL '1 hour'
GROUP BY user_id
HAVING COUNT(*) > 100
ORDER BY request_count DESC;
```

---

### 6. Проект: Rate limiting в Roolz/Adtime
**Папка:** `6.task_project/`

**Вопросы:**
1. Где применить rate limiting в Roolz?
   - API для поставщиков (защита от spam)
   - Логистические расчёты (ограничение ресурсов)
   - Внутренние микросервисы

2. Какой алгоритм выбрать?
   - **Token Bucket** — для API (гибкость + burst)
   - **Fixed Window** — для простых случаев
   - **Sliding Window** — для точности

3. Как реализовать?
   - Middleware для HTTP
   - Redis для distributed rate limiting
   - Prometheus для мониторинга

---

## 🔗 Связь с CITADEL

- **CITADEL_COMPLETE_ANSWERS.md** — ЧАСТЬ 1: Goroutines, Channels, Mutex
- **CITADEL_COMPLETE_TASKS.md** — ЧАСТЬ 2-3: Concurrency patterns
- **6.six-stickers/1.golang_main/2.sema_context/** — Semaphore example

---

## 🚀 Порядок изучения

1. Реши разминку (`1.task_warmup`)
2. Реши Two Sum (`2.task_leetcode`)
3. Реализуй ограничение запросов (`3.task_concurrency`)
4. Изучи все 4 rate limiter паттерна (`4.task_pattern`)
5. Напиши SQL запрос (`5.task_sql`)
6. Подумай над проектом (`6.task_project`)

---

**Удачи на собесе!** 🎯
