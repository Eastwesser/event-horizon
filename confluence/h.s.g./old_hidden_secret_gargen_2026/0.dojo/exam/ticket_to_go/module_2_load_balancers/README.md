# 🎫 Модуль 2: Load Balancer + Структуры данных

**Цель:** Научиться распределять нагрузку между серверами.

---

## 📋 Билет

1. **Разминка:** Работа с мапами (порядок итерации, конкурентность)
2. **LeetCode:** Valid Parentheses (LeetCode 20)
3. **Конкурентность:** Worker Pool с graceful shutdown
4. **Паттерн:** Round Robin Load Balancer
5. **SQL:** Шардирование таблицы запросов по диапазонам
6. **Проект:** Распределение нагрузки в Roolz (логистика)

---

## 🎯 Задачи

### 1. Разминка: Мапы

**Вопросы:**
```go
// 1. Какой порядок итерации?
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Println(k, v)  // Случайный порядок! (security feature)
}

// 2. Concurrent map writes?
for i := 0; i < 10; i++ {
    go func(id int) {
        m[fmt.Sprintf("key%d", id)] = id  // ❌ Data race!
    }(i)
}
```

**Решение:**
- Используй `sync.RWMutex` или `sync.Map`

---

### 2. LeetCode: Valid Parentheses

**Задача:** [LeetCode 20](https://leetcode.com/problems/valid-parentheses/)

```go
func isValid(s string) bool {
    stack := []rune{}
    pairs := map[rune]rune{')': '(', '}': '{', ']': '['}
    
    for _, ch := range s {
        if opening, ok := pairs[ch]; ok {
            // Closing bracket
            if len(stack) == 0 || stack[len(stack)-1] != opening {
                return false
            }
            stack = stack[:len(stack)-1]  // Pop
        } else {
            // Opening bracket
            stack = append(stack, ch)
        }
    }
    
    return len(stack) == 0
}
```

---

### 3. Worker Pool с Graceful Shutdown

**См.:** `6.six-stickers/1.golang_main/3.worker_pool/`

**Ключевые моменты:**
- `close(jobs)` → воркеры завершаются
- `wg.Wait()` → ждём завершения
- `context.WithCancel()` → отмена

---

### 4. Round Robin Load Balancer

**Паттерн:**
```go
type RoundRobin struct {
    servers []string
    current int
    mu      sync.Mutex
}

func (rr *RoundRobin) Next() string {
    rr.mu.Lock()
    defer rr.mu.Unlock()
    
    server := rr.servers[rr.current]
    rr.current = (rr.current + 1) % len(rr.servers)
    return server
}
```

**Другие алгоритмы:**
- **Weighted Round Robin** — с весами
- **Least Connections** — наименее загруженный
- **IP Hash** — по IP клиента

---

### 5. SQL: Шардирование

```sql
-- Создать шарды по диапазонам user_id
CREATE TABLE requests_shard_1 (
    CHECK (user_id >= 0 AND user_id < 1000000)
) INHERITS (requests);

CREATE TABLE requests_shard_2 (
    CHECK (user_id >= 1000000 AND user_id < 2000000)
) INHERITS (requests);

-- Автоматическая маршрутизация
CREATE RULE requests_insert_shard_1 AS
ON INSERT TO requests
WHERE (user_id >= 0 AND user_id < 1000000)
DO INSTEAD
INSERT INTO requests_shard_1 VALUES (NEW.*);
```

---

### 6. Проект: Roolz логистика

**Вопросы:**
- Как распределить поставщиков по регионам?
- Как балансировать между складами?
- Какой алгоритм выбрать?

**Решение:**
- **Геолокация** + **Round Robin**
- **Capacity-aware** балансировка
- Мониторинг через Prometheus

---

## 🔗 См. также

- **CITADEL** — Load Balancing strategies
- **6.six-stickers/3.worker_pool** — Worker Pool example

Удачи! 🎯
