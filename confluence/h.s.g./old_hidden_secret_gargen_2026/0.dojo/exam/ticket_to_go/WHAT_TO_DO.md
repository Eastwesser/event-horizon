# 🎫 Ticket to Go — Билеты к собеседованию

**Цель:** Подготовиться к техническому собеседованию по Go через систему модулей.

---

## 📚 Структура курса

**6 модулей** — каждый покрывает важный паттерн для продакшена:

1. **Module 1: Rate Limiter** — Защита от перегрузки
2. **Module 2: Load Balancer** — Распределение нагрузки
3. **Module 3: Circuit Breaker** — Обработка ошибок
4. **Module 4: Health Checker** — Мониторинг
5. **Module 5: Retry with Backoff** — Устойчивость
6. **Module 6: Cache (LRU)** — Оптимизация

---

## 🎯 Формат каждого модуля

Каждый модуль содержит **6 типов задач:**

1. **Разминка** — Концепция Go (замыкания, мапы, паники, и т.д.)
2. **LeetCode** — Алгоритмическая задача
3. **Конкурентность** — Goroutines, Channels, Sync
4. **Паттерн** — Production-ready паттерн (Rate Limiter, Circuit Breaker, и т.д.)
5. **SQL** — База данных (запросы, оптимизация)
6. **Проект** — Применение в реальных проектах

---

## 📂 Как пользоваться

### 1. Открой модуль
```bash
cd modules/module_1_rate_limiters/
```

### 2. Прочитай README
Каждый модуль имеет подробный `README.md` с:
- Описанием билета
- Задачами с решениями
- Примерами кода
- Ссылками на CITADEL

### 3. Реши задачи
Пройди все 6 задач по порядку:
- Разминка
- LeetCode
- Concurrency
- Паттерн
- SQL
- Проект

### 4. Запусти код
```bash
go run main.go
```

---

## 🔥 Ключевые паттерны (обязательно знать!)

### 1. Rate Limiter (Token Bucket)
**Зачем:** Защита API от DDoS, ограничение запросов.

**Алгоритмы:**
- Fixed Window
- Sliding Window
- **Token Bucket** ⭐
- Leaky Bucket

**Use case:** API rate limiting (100 req/min per user)

---

### 2. Load Balancer (Round Robin)
**Зачем:** Распределение нагрузки между серверами.

**Алгоритмы:**
- **Round Robin** ⭐
- Weighted Round Robin
- Least Connections
- IP Hash

**Use case:** Балансировка между воркерами, серверами

---

### 3. Circuit Breaker
**Зачем:** Защита от каскадных отказов при проблемах с внешними сервисами.

**Состояния:**
- **Closed** (закрыт) → запросы проходят
- **Open** (открыт) → запросы блокируются
- **Half-Open** (полуоткрыт) → тестовые запросы

**Use case:** Защита от падения внешних API (поставщики, оплата)

---

### 4. Health Checker
**Зачем:** Мониторинг здоровья сервисов.

**Фичи:**
- Периодические проверки (`time.Ticker`)
- Таймауты (`context.WithTimeout`)
- Exponential backoff при ошибках
- Метрики (uptime, error rate)

**Use case:** Мониторинг микросервисов, SLA расчёты

---

### 5. Retry with Backoff
**Зачем:** Повторять неудачные запросы с умной стратегией.

**Стратегии:**
- **Exponential Backoff** (1s, 2s, 4s, 8s...)
- **Jitter** (случайная добавка для распределения нагрузки)
- **Max Retries** (лимит попыток)

**Use case:** Retry для внешних API, очереди, БД

---

### 6. LRU Cache
**Зачем:** Оптимизация производительности через кэширование.

**Структура:**
- **Double Linked List** (порядок использования)
- **HashMap** (быстрый доступ O(1))

**Стратегии:**
- Write-Through (запись в кэш + БД)
- Write-Back (запись только в кэш)
- TTL (время жизни)

**Use case:** Кэширование результатов запросов, сессии

---

## 📊 Таблица модулей

| Модуль | Паттерн | LeetCode | Ключевая концепция |
|--------|---------|----------|-------------------|
| 1 | Rate Limiter | Two Sum (1) | Token Bucket |
| 2 | Load Balancer | Valid Parentheses (20) | Round Robin |
| 3 | Circuit Breaker | Merge Lists (21) | 3 состояния |
| 4 | Health Checker | Linked List Cycle (141) | Periodic checks |
| 5 | Retry with Backoff | LRU Cache (146) | Exponential + Jitter |
| 6 | LRU Cache | Implement Trie (208) | Double LL + HashMap |

---

## 🎯 План подготовки к собесу

### Неделя 1: Основы
- **Day 1-2:** Module 1 (Rate Limiter)
- **Day 3-4:** Module 2 (Load Balancer)
- **Day 5-6:** Module 3 (Circuit Breaker)
- **Day 7:** Повторение, практика

### Неделя 2: Продвинутое
- **Day 1-2:** Module 4 (Health Checker)
- **Day 3-4:** Module 5 (Retry)
- **Day 5-6:** Module 6 (Cache)
- **Day 7:** Повторение всех паттернов

### Неделя 3: Проекты
- Применение паттернов в реальных проектах
- System Design вопросы
- Mock interviews

---

## 💡 Советы для собеса

1. **Паттерны важнее алгоритмов**
   - На продакшен собесах спрашивают **паттерны** (Rate Limiter, Circuit Breaker)
   - LeetCode — только для разминки

2. **Объясняй trade-offs**
   - "Token Bucket лучше Fixed Window, потому что..."
   - "Round Robin проще Weighted RR, но..."

3. **Связывай с реальными проектами**
   - "В AdTime мы использовали бы Rate Limiter для..."
   - "Circuit Breaker защитит от падения API поставщиков"

4. **Знай метрики**
   - Uptime, Error Rate, Latency
   - SLA расчёты

5. **Практикуй SQL**
   - Оптимизация запросов
   - Индексы, шардирование
   - Геолокационные запросы

---

## 📝 Чек-лист перед собесом

**Go концепции:**
- ✅ Goroutines, Channels, Select
- ✅ Mutex, RWMutex, atomic
- ✅ Context (WithCancel, WithTimeout)
- ✅ WaitGroup, Semaphore
- ✅ Panic & Recover
- ✅ Defer, порядок выполнения
- ✅ Замыкания в горутинах

**Паттерны:**
- ✅ Rate Limiter (Token Bucket)
- ✅ Load Balancer (Round Robin)
- ✅ Circuit Breaker (3 состояния)
- ✅ Health Checker (periodic checks)
- ✅ Retry with Backoff (exponential + jitter)
- ✅ LRU Cache (Double LL + HashMap)

**SQL:**
- ✅ Joins, Group By, Having
- ✅ Индексы, EXPLAIN ANALYZE
- ✅ Шардирование, репликация
- ✅ Транзакции, изоляция

**Проект:**
- ✅ Adtime — генерация подарков, мониторинг

---

**Удачи на собесе!** 🎯
