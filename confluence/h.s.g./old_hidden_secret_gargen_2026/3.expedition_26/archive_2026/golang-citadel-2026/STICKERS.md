
---

## 🔍 Объяснение двух ловушек

### Ловушка 1: nil интерфейс vs интерфейс с nil

```go
type MyErr struct{}
func (*MyErr) Error() string { return "x" }

var e error = (*MyErr)(nil)
// e != nil  // почему?
```

**Суть:** Интерфейс в Go — это структура `(type, value)`.
- **nil интерфейс**: `(nil, nil)` → `e == nil` ✅
- **интерфейс с nil**: `(*MyErr, nil)` → `e != nil` ❌ (тип есть, значение nil)

**Почему:** Интерфейс знает тип `*MyErr`, даже если значение nil. Проверка `e == nil` вернёт `false`, потому что тип не nil.

**Как проверить значение внутри:**
```go
if e != nil {
    v := reflect.ValueOf(e)
    if v.IsNil() { /* реально nil */ }
}
```

### Ловушка 2: LEFT JOIN + WHERE убивает LEFT

```sql
-- ❌ ПЛОХО: WHERE убивает LEFT JOIN
SELECT u.id, r.id
FROM a_user u
LEFT JOIN d_return r ON r.user_id = u.id
WHERE r.created_at >= :from;  -- это превращает LEFT в INNER!

-- ✅ ХОРОШО: фильтр в ON сохраняет LEFT
SELECT u.id, r.id
FROM a_user u
LEFT JOIN d_return r 
  ON r.user_id = u.id 
 AND r.created_at >= :from;  -- фильтр в ON
```

**Суть:** `WHERE` применяется **после** JOIN. Если фильтруешь по правой таблице (`r.created_at`), все строки без правой части отбрасываются → LEFT JOIN становится INNER JOIN.

**Правило:** Фильтр по правой таблице в `ON` сохраняет LEFT, в `WHERE` — убивает.

---

## Боевые “супер-стикеры” (A–F) — твоя структура

## (A) nil интерфейсы + type assertion + Generics [T]

```go
// type assertion (без паники)
v, ok := x.(T)
if !ok { /* handle */ }

// type switch
switch v := x.(type) {
case T:
    _ = v
default:
}

// nil-ловушка
var e error = (*MyErr)(nil)
// e != nil  // тип есть, значение nil

// Generics [T]
func Max[T comparable](a, b T) T {
    if a > b { return a }
    return b
}
```

**Ловушка:** `(*MyErr)(nil)` в интерфейсе → `e != nil` (тип есть, значение nil). Generics: `comparable` для `==`, `constraints.Ordered` для `< >`.

## (B) select + context

```go
// context + select
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel() // всегда!

select {
case <-ctx.Done():
    return ctx.Err()
case v := <-ch:
    _ = v
case <-time.After(200 * time.Millisecond):
    // timeout
default:
    // non-blocking (осторожно: busy loop!)
}
```

**Ловушка:** Вызывающий создаёт `ctx` и владеет `cancel()`. Забытый `cancel()` = утечка. `default` в `select` = busy spinning (100% CPU).

## (C) sync.* (mutex, RWMutex, Cond, Once, Pool, Map)


















```go
// Mutex
mu.Lock()
state[key]++
mu.Unlock()

// RWMutex (много чтения, мало записи)
rw.RLock()
_ = state[key]
rw.RUnlock()

// sync.Cond (ожидание события)
cond := sync.NewCond(&mu)
cond.Wait()  // ждёт сигнал
cond.Signal() // будит одного

// sync.Once (один раз)
var once sync.Once
once.Do(func() { /* init */ })

// sync.Pool (переиспользование объектов)
pool := &sync.Pool{
    New: func() interface{} { return &Buffer{} },
}
buf := pool.Get().(*Buffer)
defer pool.Put(buf)

// sync.Map (thread-safe map)
var m sync.Map
m.Store("key", "value")
v, ok := m.Load("key")
```

**Ловушка:** `sync.Map` — только для read-heavy или когда ключи не пересекаются. Обычная `map + mutex` быстрее для write-heavy.














## (D) SQL агрегатные функции (SUM, COALESCE, GROUP BY)

```sql
-- SUM, COUNT, AVG, MAX, MIN
SELECT 
    user_id,
    SUM(amount) AS total,
    COUNT(*) AS orders,
    COALESCE(SUM(discount), 0) AS total_discount
FROM b_purchase
WHERE created_at >= :from
GROUP BY user_id
HAVING SUM(amount) > 1000;  -- фильтр после GROUP BY

-- COALESCE (первое не-NULL)
SELECT COALESCE(name, email, 'unknown') AS display_name;

-- CASE WHEN
SELECT 
    CASE 
        WHEN amount > 1000 THEN 'high'
        WHEN amount > 100 THEN 'medium'
        ELSE 'low'
    END AS category
FROM b_purchase;
```

**Ловушка:** `WHERE` фильтрует строки до GROUP BY, `HAVING` — после. `COALESCE` возвращает первое не-NULL значение.
















## (E) Сложные CREATE TABLE (CHECK, PARTITION, UNIQUE)

```sql
-- CHECK constraint
CREATE TABLE b_purchase (
    id bigserial PRIMARY KEY,
    amount numeric(10,2) NOT NULL CHECK (amount > 0),
    status text NOT NULL CHECK (status IN ('pending', 'paid', 'refunded')),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- UNIQUE constraint
CREATE TABLE a_user (
    id bigserial PRIMARY KEY,
    email text UNIQUE NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Composite index
CREATE INDEX idx_purchase_user_time 
ON b_purchase(user_id, created_at DESC);

-- Partial index (только для активных)
CREATE INDEX idx_purchase_active 
ON b_purchase(user_id) 
WHERE status = 'paid';
```





**Ловушка:** CHECK валидирует при INSERT/UPDATE. Partial index экономит место и ускоряет запросы с фильтром.
























## (F) Сложные EXPLAIN ANALYZE + продакшн-запросы

```sql
-- EXPLAIN ANALYZE (BUFFERS показывает I/O)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT u.id, SUM(p.amount)
FROM a_user u
JOIN b_purchase p ON p.user_id = u.id
WHERE u.created_at >= :from
GROUP BY u.id
HAVING SUM(p.amount) > 1000;

-- Что смотреть:
-- 1. Seq Scan vs Index Scan
-- 2. Nested Loop vs Hash Join
-- 3. Sort (Memory vs Disk)
-- 4. Actual rows vs Planned rows

-- CTE (Common Table Expression)
WITH recent_users AS (
    SELECT id FROM a_user 
    WHERE created_at >= :from
)
SELECT u.id, COUNT(p.id)
FROM recent_users u
LEFT JOIN b_purchase p ON p.user_id = u.id
GROUP BY u.id;

-- Window functions
SELECT 
    user_id,
    amount,
    SUM(amount) OVER (PARTITION BY user_id) AS user_total,
    ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) AS rn
FROM b_purchase;
```

**Ловушка:** `BUFFERS` показывает реальный I/O. Если `Actual rows >> Planned rows` → устаревшая статистика (`ANALYZE table`). Window functions не группируют строки, а добавляют вычисляемые столбцы.
