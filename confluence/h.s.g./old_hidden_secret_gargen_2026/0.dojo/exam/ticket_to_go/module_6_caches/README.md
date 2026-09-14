# 🎫 Модуль 6: Кэширование (LRU) + Геолокация

**Цель:** Оптимизировать производительность через кэширование и пространственные структуры.

---

## 📋 Билет

1. **Разминка:** Срезы vs массивы, capacity
2. **LeetCode:** Implement Trie (LeetCode 208) или Ball Tree задача
3. **Конкурентность:** Thread-safe LRU кэш
4. **Паттерн:** Write-through vs write-back кэш
5. **SQL:** Геолокационные запросы (ближайшие точки)
6. **Проект:** Ball Tree в Roolz для поиска поставщиков

---

## 🎯 Задачи

### 1. Slice vs Array, Capacity

```go
// Array (фиксированный размер)
arr := [5]int{1, 2, 3, 4, 5}

// Slice (динамический)
slice := []int{1, 2, 3}
fmt.Println(len(slice), cap(slice))  // 3, 3

slice = append(slice, 4)
fmt.Println(len(slice), cap(slice))  // 4, 6 (реаллокация!)
```

---

### 2. LeetCode: LRU Cache

**Задача:** [LeetCode 146](https://leetcode.com/problems/lru-cache/)

**Структура:**
- **Double Linked List** — порядок использования
- **HashMap** — быстрый доступ O(1)

```go
type LRUCache struct {
    capacity int
    cache    map[int]*Node
    head     *Node
    tail     *Node
}

type Node struct {
    key, value int
    prev, next *Node
}

func (c *LRUCache) Get(key int) int {
    if node, ok := c.cache[key]; ok {
        c.moveToFront(node)
        return node.value
    }
    return -1
}

func (c *LRUCache) Put(key, value int) {
    if node, ok := c.cache[key]; ok {
        node.value = value
        c.moveToFront(node)
        return
    }
    
    // Новый элемент
    node := &Node{key: key, value: value}
    c.cache[key] = node
    c.addToFront(node)
    
    if len(c.cache) > c.capacity {
        c.removeLRU()
    }
}
```

---

### 3. Thread-safe LRU

```go
type SafeLRU struct {
    mu    sync.RWMutex
    cache *LRUCache
}

func (s *SafeLRU) Get(key int) int {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.cache.Get(key)
}
```

---

### 4. Write-Through vs Write-Back

**Write-Through:**
- Запись в кэш **И** в БД одновременно
- Медленнее, но консистентно

**Write-Back:**
- Запись только в кэш
- Асинхронная запись в БД
- Быстрее, но риск потери данных

---

### 5. SQL: Геолокация

```sql
-- Ближайшие поставщики (PostGIS)
SELECT 
    name,
    ST_Distance(location, ST_MakePoint(37.6156, 55.7522)) as distance_km
FROM suppliers
ORDER BY location <-> ST_MakePoint(37.6156, 55.7522)  -- KNN оператор
LIMIT 10;

-- Без PostGIS (Haversine formula)
SELECT 
    name,
    (6371 * acos(cos(radians(55.7522)) * cos(radians(lat)) 
    * cos(radians(lon) - radians(37.6156)) 
    + sin(radians(55.7522)) * sin(radians(lat)))) AS distance_km
FROM suppliers
ORDER BY distance_km
LIMIT 10;
```

---

### 6. Ball Tree для Roolz

**Идея:** Быстрый поиск ближайших соседей в многомерном пространстве.

**Use case:**
- Поиск ближайших поставщиков
- Логистическая оптимизация
- k-NN алгоритмы

**Библиотеки:**
- `github.com/golang/geo` (S2 geometry)
- Custom Ball Tree implementation

---

**Паттерны кэширования:**
- **LRU** (Least Recently Used)
- **LFU** (Least Frequently Used)
- **TTL** (Time To Live)
- **Write-Through / Write-Back**

Удачи! 🚀
