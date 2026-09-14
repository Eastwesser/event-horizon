# 🗄️ SQL Insights — Базовые задачи для собеса

**Цель:** Отработать SQL запросы, которые часто спрашивают на собесах.

---

## 📋 Источники задач

- **SQL Academy** — интерактивные задачи
- **Hobby Games** (настольные игры) — реальные данные из проекта

---

## 🎯 Топ-20 SQL вопросов на собесах

### 1. Основные запросы (CRUD)

**SELECT с WHERE:**
```sql
-- Найти всех пользователей старше 25 лет
SELECT * FROM users WHERE age > 25;

-- С несколькими условиями
SELECT * FROM users 
WHERE age > 25 AND city = 'Moscow' AND is_active = true;
```

**JOIN (LEFT, RIGHT, INNER):**
```sql
-- INNER JOIN (только совпадения)
SELECT u.name, o.order_id, o.total
FROM users u
INNER JOIN orders o ON u.user_id = o.user_id;

-- LEFT JOIN (все из левой таблицы)
SELECT u.name, COUNT(o.order_id) as order_count
FROM users u
LEFT JOIN orders o ON u.user_id = o.user_id
GROUP BY u.user_id, u.name;
```

---

### 2. Агрегатные функции

**COUNT, SUM, AVG, MIN, MAX:**
```sql
-- Количество заказов по пользователям
SELECT user_id, COUNT(*) as order_count
FROM orders
GROUP BY user_id;

-- Средний чек
SELECT AVG(total) as avg_order_value FROM orders;

-- Топ-10 клиентов по сумме заказов
SELECT user_id, SUM(total) as total_spent
FROM orders
GROUP BY user_id
ORDER BY total_spent DESC
LIMIT 10;
```

---

### 3. GROUP BY + HAVING

**Разница WHERE vs HAVING:**
```sql
-- WHERE — фильтрует строки ДО группировки
-- HAVING — фильтрует группы ПОСЛЕ группировки

-- Найти пользователей с >5 заказами
SELECT user_id, COUNT(*) as order_count
FROM orders
WHERE status = 'completed'  -- Фильтр ДО группировки
GROUP BY user_id
HAVING COUNT(*) > 5;        -- Фильтр ПОСЛЕ группировки
```

---

### 4. Подзапросы (Subqueries)

**Scalar Subquery:**
```sql
-- Найти пользователей со средним чеком выше общего среднего
SELECT user_id, AVG(total) as avg_order
FROM orders
GROUP BY user_id
HAVING AVG(total) > (SELECT AVG(total) FROM orders);
```

**IN Subquery:**
```sql
-- Найти пользователей, которые заказывали товар с ID=5
SELECT * FROM users
WHERE user_id IN (
    SELECT DISTINCT user_id 
    FROM orders 
    WHERE product_id = 5
);
```

**EXISTS:**
```sql
-- Найти пользователей, у которых есть хотя бы один заказ
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o WHERE o.user_id = u.user_id
);
```

---

### 5. Window Functions (Оконные функции)

**ROW_NUMBER:**
```sql
-- Пронумеровать заказы каждого пользователя
SELECT 
    user_id, 
    order_id, 
    total,
    ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) as order_number
FROM orders;
```

**RANK:**
```sql
-- Ранжирование пользователей по сумме заказов
SELECT 
    user_id,
    SUM(total) as total_spent,
    RANK() OVER (ORDER BY SUM(total) DESC) as rank
FROM orders
GROUP BY user_id;
```

**LAG/LEAD:**
```sql
-- Сравнить текущий заказ с предыдущим
SELECT 
    order_id,
    total,
    LAG(total) OVER (ORDER BY created_at) as prev_total,
    total - LAG(total) OVER (ORDER BY created_at) as diff
FROM orders;
```

---

### 6. Common Table Expressions (CTE)

**WITH clause:**
```sql
-- Найти пользователей с наибольшим числом заказов
WITH user_orders AS (
    SELECT user_id, COUNT(*) as order_count
    FROM orders
    GROUP BY user_id
)
SELECT u.name, uo.order_count
FROM users u
JOIN user_orders uo ON u.user_id = uo.user_id
WHERE uo.order_count > 10
ORDER BY uo.order_count DESC;
```

**Recursive CTE (для деревьев):**
```sql
-- Иерархия сотрудников
WITH RECURSIVE employee_hierarchy AS (
    -- Базовый случай (CEO)
    SELECT employee_id, name, manager_id, 1 as level
    FROM employees
    WHERE manager_id IS NULL
    
    UNION ALL
    
    -- Рекурсивная часть
    SELECT e.employee_id, e.name, e.manager_id, eh.level + 1
    FROM employees e
    JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy ORDER BY level, name;
```

---

### 7. CASE WHEN (Условная логика)

**Simple CASE:**
```sql
SELECT 
    product_name,
    price,
    CASE 
        WHEN price < 100 THEN 'Cheap'
        WHEN price < 500 THEN 'Medium'
        ELSE 'Expensive'
    END as price_category
FROM products;
```

**CASE в GROUP BY:**
```sql
-- Группировка по возрастным категориям
SELECT 
    CASE 
        WHEN age < 18 THEN 'Child'
        WHEN age < 60 THEN 'Adult'
        ELSE 'Senior'
    END as age_group,
    COUNT(*) as user_count
FROM users
GROUP BY age_group;
```

---

### 8. UNION vs UNION ALL

**UNION (убирает дубликаты):**
```sql
SELECT name FROM users WHERE city = 'Moscow'
UNION
SELECT name FROM users WHERE city = 'SPb';
```

**UNION ALL (быстрее, с дубликатами):**
```sql
SELECT name FROM users WHERE city = 'Moscow'
UNION ALL
SELECT name FROM users WHERE city = 'SPb';
```

---

### 9. Индексы и производительность

**EXPLAIN ANALYZE:**
```sql
-- Анализ плана выполнения
EXPLAIN ANALYZE
SELECT * FROM orders WHERE user_id = 123;
```

**Создание индекса:**
```sql
-- B-tree индекс (по умолчанию)
CREATE INDEX idx_user_id ON orders(user_id);

-- Составной индекс
CREATE INDEX idx_user_status ON orders(user_id, status);

-- Частичный индекс
CREATE INDEX idx_active_orders ON orders(user_id) 
WHERE status = 'active';

-- Уникальный индекс
CREATE UNIQUE INDEX idx_email ON users(email);
```

---

### 10. Транзакции и изоляция

**ACID:**
```sql
BEGIN;

UPDATE accounts SET balance = balance - 100 WHERE account_id = 1;
UPDATE accounts SET balance = balance + 100 WHERE account_id = 2;

COMMIT;  -- или ROLLBACK в случае ошибки
```

**Isolation Levels:**
```sql
-- Read Uncommitted (грязное чтение)
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

-- Read Committed (по умолчанию в PostgreSQL)
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Repeatable Read (защита от non-repeatable read)
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- Serializable (самый строгий)
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

---

## 🔥 Задачи из Hobby Games (Настольные игры)

### Задача 1: Популярные игры

```sql
-- Топ-10 игр по количеству заказов
SELECT 
    g.game_name,
    COUNT(DISTINCT o.order_id) as order_count,
    SUM(oi.quantity) as total_sold
FROM games g
JOIN order_items oi ON g.game_id = oi.game_id
JOIN orders o ON oi.order_id = o.order_id
WHERE o.status = 'completed'
GROUP BY g.game_id, g.game_name
ORDER BY order_count DESC
LIMIT 10;
```

---

### Задача 2: Когортный анализ

```sql
-- Первая покупка каждого пользователя
WITH first_purchase AS (
    SELECT 
        user_id,
        MIN(DATE_TRUNC('month', created_at)) as cohort_month
    FROM orders
    GROUP BY user_id
)

-- Количество пользователей по когортам
SELECT 
    cohort_month,
    COUNT(DISTINCT user_id) as cohort_size
FROM first_purchase
GROUP BY cohort_month
ORDER BY cohort_month;
```

---

### Задача 3: RFM анализ (Recency, Frequency, Monetary)

```sql
WITH rfm AS (
    SELECT 
        user_id,
        -- Recency (дней с последнего заказа)
        CURRENT_DATE - MAX(created_at::date) as recency,
        -- Frequency (количество заказов)
        COUNT(*) as frequency,
        -- Monetary (сумма заказов)
        SUM(total) as monetary
    FROM orders
    WHERE status = 'completed'
    GROUP BY user_id
)

SELECT 
    user_id,
    recency,
    frequency,
    monetary,
    -- Сегментация
    CASE 
        WHEN recency <= 30 AND frequency >= 5 AND monetary >= 10000 THEN 'VIP'
        WHEN recency <= 90 AND frequency >= 3 THEN 'Active'
        WHEN recency > 180 THEN 'Churned'
        ELSE 'Regular'
    END as segment
FROM rfm
ORDER BY monetary DESC;
```

---

### Задача 4: Геолокационные запросы

```sql
-- Ближайшие магазины (с PostGIS)
SELECT 
    store_name,
    ST_Distance(
        location, 
        ST_MakePoint(37.6156, 55.7522)  -- Координаты пользователя
    ) as distance_meters
FROM stores
ORDER BY location <-> ST_MakePoint(37.6156, 55.7522)  -- KNN оператор
LIMIT 5;

-- Без PostGIS (Haversine formula)
SELECT 
    store_name,
    (6371 * acos(
        cos(radians(55.7522)) * cos(radians(lat)) * 
        cos(radians(lon) - radians(37.6156)) + 
        sin(radians(55.7522)) * sin(radians(lat))
    )) AS distance_km
FROM stores
ORDER BY distance_km
LIMIT 5;
```

---

## 🎯 Советы для SQL на собесах

### 1. Объясняй план выполнения
- Используй `EXPLAIN ANALYZE`
- Говори про индексы (когда они нужны, когда нет)
- Обсуждай сканирование (Seq Scan vs Index Scan)

### 2. Знай разницу
- **WHERE vs HAVING** (до vs после группировки)
- **UNION vs UNION ALL** (с дубликатами vs без)
- **INNER JOIN vs LEFT JOIN** (только совпадения vs все из левой)
- **DELETE vs TRUNCATE** (логирование vs быстрое удаление)

### 3. Оптимизация
- Избегай `SELECT *` (только нужные колонки)
- Используй `LIMIT` для больших выборок
- Индексы на часто используемые колонки (WHERE, JOIN, ORDER BY)
- Частичные индексы для фильтрации

### 4. Пиши читаемый SQL
```sql
-- ❌ ПЛОХО
select u.name,sum(o.total) from users u join orders o on u.user_id=o.user_id where o.status='completed' group by u.user_id;

-- ✅ ХОРОШО
SELECT 
    u.name,
    SUM(o.total) as total_spent
FROM users u
JOIN orders o ON u.user_id = o.user_id
WHERE o.status = 'completed'
GROUP BY u.user_id, u.name
ORDER BY total_spent DESC;
```

---

## 🔗 См. также

- **CITADEL_COMPLETE_ANSWERS.md** — ЧАСТЬ 6: SQL (Индексы, Транзакции, ACID)
- **CITADEL_COMPLETE_TASKS.md** — ЧАСТЬ 10: SQL задачи Лукьянова
- **SQL Academy** — практика запросов

---

**Удачи на собесе!** 🗄️
