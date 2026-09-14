# 🧠 Difficult Algorithms — Продвинутые алгоритмы для Senior

**Цель:** Разобрать сложные алгоритмы, которые спрашивают на Senior позициях и в FAANG.

---

## 📋 Список алгоритмов

1. **A*** (A-star) — Поиск кратчайшего пути с эвристикой
2. **CQRS** (Command Query Responsibility Segregation) — Паттерн архитектуры
3. **Dijkstra + Ball Tree** — Кратчайший путь + Пространственные структуры
4. **TSP** (Traveling Salesman Problem) — Задача коммивояжёра
5. **VTSP** (Vehicle TSP) — Задача маршрутизации транспорта
6. **Spread** — Распределение нагрузки
7. **KNN** (K-Nearest Neighbors) — k ближайших соседей
8. **Greedy** — Жадные алгоритмы
9. **SCAN** — Spatial Clustering

---

## 🎯 Детальный разбор

### 1. A* (A-star) — Поиск пути с эвристикой

**Что это:**
- Алгоритм поиска кратчайшего пути на графе
- Использует **эвристику** (оценку расстояния до цели)
- Быстрее Dijkstra благодаря приоритизации перспективных путей

**Формула:**
```
f(n) = g(n) + h(n)

где:
g(n) = стоимость пути от старта до n
h(n) = эвристическая оценка от n до цели (например, Euclidean distance)
```

**Реализация:**
```go
type Node struct {
    x, y     int
    g, h, f  float64
    parent   *Node
}

func AStar(start, goal Node, grid [][]int) []Node {
    openSet := &PriorityQueue{}
    heap.Push(openSet, &start)
    
    closedSet := make(map[string]bool)
    
    for openSet.Len() > 0 {
        current := heap.Pop(openSet).(*Node)
        
        if current.x == goal.x && current.y == goal.y {
            return reconstructPath(current)
        }
        
        key := fmt.Sprintf("%d,%d", current.x, current.y)
        closedSet[key] = true
        
        for _, neighbor := range getNeighbors(current, grid) {
            nKey := fmt.Sprintf("%d,%d", neighbor.x, neighbor.y)
            if closedSet[nKey] {
                continue
            }
            
            tentativeG := current.g + distance(current, neighbor)
            
            if tentativeG < neighbor.g {
                neighbor.parent = current
                neighbor.g = tentativeG
                neighbor.h = heuristic(neighbor, goal)
                neighbor.f = neighbor.g + neighbor.h
                
                heap.Push(openSet, neighbor)
            }
        }
    }
    
    return nil  // Путь не найден
}

// Эвристика: Euclidean distance
func heuristic(a, b Node) float64 {
    dx := float64(a.x - b.x)
    dy := float64(a.y - b.y)
    return math.Sqrt(dx*dx + dy*dy)
}
```

**Use case:**
- Навигация (карты, GPS)
- Игры (движение NPC)
- Роботика

---

### 2. CQRS — Command Query Responsibility Segregation

**Что это:**
- Паттерн разделения **записи** (Command) и **чтения** (Query)
- Разные модели данных для записи и чтения
- Часто используется с **Event Sourcing**

**Архитектура:**
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ├───── Command ────> ┌──────────────┐
       │                    │ Write Model  │
       │                    │ (Normalized) │
       │                    └──────┬───────┘
       │                           │
       │                      Event Store
       │                           │
       │                    ┌──────▼───────┐
       └───── Query ─────>  │  Read Model  │
                            │(Denormalized)│
                            └──────────────┘
```

**Пример:**
```go
// Command (запись)
type CreateOrderCommand struct {
    UserID    int
    ProductID int
    Quantity  int
}

type OrderCommandHandler struct {
    repo *OrderRepository
}

func (h *OrderCommandHandler) Handle(cmd CreateOrderCommand) error {
    order := &Order{
        UserID:    cmd.UserID,
        ProductID: cmd.ProductID,
        Quantity:  cmd.Quantity,
        CreatedAt: time.Now(),
    }
    
    if err := h.repo.Save(order); err != nil {
        return err
    }
    
    // Публикуем событие
    publishEvent(OrderCreatedEvent{OrderID: order.ID})
    
    return nil
}

// Query (чтение)
type GetOrdersByUserQuery struct {
    UserID int
}

type OrderQueryHandler struct {
    readDB *ReadDB  // Денормализованная БД (например, MongoDB)
}

func (h *OrderQueryHandler) Handle(q GetOrdersByUserQuery) ([]OrderView, error) {
    return h.readDB.FindOrdersByUser(q.UserID)
}
```

**Преимущества:**
- Независимое масштабирование записи и чтения
- Оптимизация моделей данных под задачу
- Event Sourcing для аудита

**Use case:**
- E-commerce (заказы, платежи)
- Финансовые системы
- Микросервисы

---

### 3. Dijkstra + Ball Tree

**Dijkstra:** Кратчайший путь на графе (без эвристики)

```go
func Dijkstra(graph Graph, start, end int) ([]int, float64) {
    dist := make(map[int]float64)
    prev := make(map[int]int)
    visited := make(map[int]bool)
    
    pq := &PriorityQueue{}
    heap.Push(pq, &Item{node: start, priority: 0})
    dist[start] = 0
    
    for pq.Len() > 0 {
        current := heap.Pop(pq).(*Item).node
        
        if current == end {
            return reconstructPath(prev, end), dist[end]
        }
        
        if visited[current] {
            continue
        }
        visited[current] = true
        
        for neighbor, weight := range graph[current] {
            newDist := dist[current] + weight
            
            if oldDist, ok := dist[neighbor]; !ok || newDist < oldDist {
                dist[neighbor] = newDist
                prev[neighbor] = current
                heap.Push(pq, &Item{node: neighbor, priority: newDist})
            }
        }
    }
    
    return nil, math.Inf(1)  // Путь не найден
}
```

**Ball Tree:** Структура данных для k-NN в многомерном пространстве

```go
type BallTree struct {
    center   []float64
    radius   float64
    left     *BallTree
    right    *BallTree
    points   []Point
}

func (bt *BallTree) KNN(query []float64, k int) []Point {
    pq := &PriorityQueue{}  // Max heap
    bt.search(query, k, pq)
    
    result := make([]Point, pq.Len())
    for i := pq.Len() - 1; i >= 0; i-- {
        result[i] = heap.Pop(pq).(Point)
    }
    return result
}

func (bt *BallTree) search(query []float64, k int, pq *PriorityQueue) {
    if bt.isLeaf() {
        for _, point := range bt.points {
            dist := euclidean(query, point.coords)
            if pq.Len() < k {
                heap.Push(pq, Item{point: point, priority: dist})
            } else if dist < pq.Top().priority {
                heap.Pop(pq)
                heap.Push(pq, Item{point: point, priority: dist})
            }
        }
        return
    }
    
    // Рекурсивный поиск в ближайшем поддереве
    distLeft := euclidean(query, bt.left.center)
    distRight := euclidean(query, bt.right.center)
    
    if distLeft < distRight {
        bt.left.search(query, k, pq)
        if distRight - bt.right.radius < pq.Top().priority {
            bt.right.search(query, k, pq)
        }
    } else {
        bt.right.search(query, k, pq)
        if distLeft - bt.left.radius < pq.Top().priority {
            bt.left.search(query, k, pq)
        }
    }
}
```

**Use case:**
- Геолокация (поиск ближайших поставщиков в Roolz)
- Recommendation systems
- Clustering

---

### 4. TSP (Traveling Salesman Problem)

**Задача:** Найти кратчайший маршрут через N городов.

**Сложность:** NP-hard (точное решение за O(n!))

**Heuristic решение (Nearest Neighbor):**
```go
func TSPNearestNeighbor(cities []City) []int {
    n := len(cities)
    visited := make([]bool, n)
    path := []int{0}  // Начинаем с города 0
    visited[0] = true
    
    current := 0
    for len(path) < n {
        nearest := -1
        minDist := math.Inf(1)
        
        for i := 0; i < n; i++ {
            if !visited[i] {
                dist := distance(cities[current], cities[i])
                if dist < minDist {
                    minDist = dist
                    nearest = i
                }
            }
        }
        
        path = append(path, nearest)
        visited[nearest] = true
        current = nearest
    }
    
    return path
}
```

**Genetic Algorithm для TSP:**
```go
func TSPGA(cities []City, popSize, generations int) []int {
    population := initPopulation(popSize, len(cities))
    
    for gen := 0; gen < generations; gen++ {
        // 1. Оценка fitness
        fitness := make([]float64, popSize)
        for i, individual := range population {
            fitness[i] = 1.0 / calculateTotalDistance(individual, cities)
        }
        
        // 2. Селекция (выбираем лучших)
        newPop := selection(population, fitness)
        
        // 3. Кроссовер
        newPop = crossover(newPop)
        
        // 4. Мутация
        newPop = mutate(newPop, 0.01)
        
        population = newPop
    }
    
    // Возвращаем лучшего
    best := population[0]
    bestDist := calculateTotalDistance(best, cities)
    
    for _, individual := range population {
        dist := calculateTotalDistance(individual, cities)
        if dist < bestDist {
            best = individual
            bestDist = dist
        }
    }
    
    return best
}
```

**Use case:**
- Логистика (маршруты доставки)
- Производство (последовательность операций)
- Circuit board drilling

---

### 5. VTSP (Vehicle Routing Problem)

**Расширение TSP:** Несколько машин, capacity constraints.

**Формулировка:**
- N клиентов, M машин
- Каждая машина имеет capacity
- Минимизировать total distance

**Greedy решение:**
```go
type Vehicle struct {
    ID       int
    Capacity int
    Load     int
    Route    []int
}

func VTSP(customers []Customer, vehicles []Vehicle) {
    unvisited := make([]int, len(customers))
    for i := range customers {
        unvisited[i] = i
    }
    
    for _, vehicle := range vehicles {
        for len(unvisited) > 0 && vehicle.Load < vehicle.Capacity {
            // Найти ближайшего клиента
            nearest := findNearest(vehicle.Route[len(vehicle.Route)-1], unvisited, customers)
            
            if customers[nearest].Demand + vehicle.Load <= vehicle.Capacity {
                vehicle.Route = append(vehicle.Route, nearest)
                vehicle.Load += customers[nearest].Demand
                unvisited = remove(unvisited, nearest)
            } else {
                break  // Машина заполнена
            }
        }
    }
}
```

**Use case:**
- Доставка (Яндекс.Лавка, Ozon)
- Логистика в Roolz (распределение заказов по машинам)
- Waste collection

---

### 6. KNN (K-Nearest Neighbors)

**Задача:** Найти k ближайших соседей точки.

**Brute Force (O(n)):**
```go
func KNN(query Point, points []Point, k int) []Point {
    distances := make([]struct {
        point Point
        dist  float64
    }, len(points))
    
    for i, p := range points {
        distances[i].point = p
        distances[i].dist = euclidean(query, p)
    }
    
    sort.Slice(distances, func(i, j int) bool {
        return distances[i].dist < distances[j].dist
    })
    
    result := make([]Point, k)
    for i := 0; i < k; i++ {
        result[i] = distances[i].point
    }
    return result
}
```

**Оптимизация (Ball Tree / KD-Tree):** См. раздел Ball Tree выше.

**Use case:**
- Recommendation systems
- Поиск похожих товаров
- Геолокация

---

### 7. Greedy Algorithms

**Принцип:** Выбирай локально оптимальное решение на каждом шаге.

**Activity Selection:**
```go
type Activity struct {
    start, end int
}

func activitySelection(activities []Activity) []Activity {
    // Сортируем по времени окончания
    sort.Slice(activities, func(i, j int) bool {
        return activities[i].end < activities[j].end
    })
    
    selected := []Activity{activities[0]}
    lastEnd := activities[0].end
    
    for _, act := range activities[1:] {
        if act.start >= lastEnd {
            selected = append(selected, act)
            lastEnd = act.end
        }
    }
    
    return selected
}
```

**Huffman Coding:**
```go
func huffmanCoding(freq map[rune]int) map[rune]string {
    pq := &PriorityQueue{}
    
    for char, f := range freq {
        heap.Push(pq, &Node{char: char, freq: f})
    }
    
    for pq.Len() > 1 {
        left := heap.Pop(pq).(*Node)
        right := heap.Pop(pq).(*Node)
        
        parent := &Node{
            freq:  left.freq + right.freq,
            left:  left,
            right: right,
        }
        heap.Push(pq, parent)
    }
    
    root := heap.Pop(pq).(*Node)
    codes := make(map[rune]string)
    generateCodes(root, "", codes)
    
    return codes
}
```

---

### 8. SCAN (Spatial Clustering)

**DBSCAN (Density-Based Spatial Clustering):**

```go
func DBSCAN(points []Point, eps float64, minPts int) [][]Point {
    visited := make(map[int]bool)
    clusters := [][]Point{}
    
    for i, point := range points {
        if visited[i] {
            continue
        }
        visited[i] = true
        
        neighbors := regionQuery(point, points, eps)
        
        if len(neighbors) < minPts {
            // Noise point
            continue
        }
        
        // Start new cluster
        cluster := []Point{point}
        
        for j := 0; j < len(neighbors); j++ {
            nIdx := neighbors[j]
            
            if !visited[nIdx] {
                visited[nIdx] = true
                
                nNeighbors := regionQuery(points[nIdx], points, eps)
                if len(nNeighbors) >= minPts {
                    neighbors = append(neighbors, nNeighbors...)
                }
            }
            
            cluster = append(cluster, points[nIdx])
        }
        
        clusters = append(clusters, cluster)
    }
    
    return clusters
}

func regionQuery(point Point, points []Point, eps float64) []int {
    neighbors := []int{}
    for i, p := range points {
        if euclidean(point, p) <= eps {
            neighbors = append(neighbors, i)
        }
    }
    return neighbors
}
```

**Use case:**
- Геолокационная кластеризация (группы поставщиков в Roolz)
- Anomaly detection
- Image segmentation

---

## 🎯 Советы для собеса

### 1. Объясняй сложность
- **Time Complexity:** O(n), O(n log n), O(n²)
- **Space Complexity:** O(1), O(n)
- **Trade-offs:** Точность vs Скорость

### 2. Обсуждай real-world применение
- "В Roolz мы использовали бы Ball Tree для поиска ближайших поставщиков"
- "VTSP подходит для оптимизации маршрутов доставки"
- "CQRS для разделения write/read в микросервисах"

### 3. Знай ограничения
- **TSP:** NP-hard → нужны эвристики
- **DBSCAN:** чувствителен к параметрам eps, minPts
- **Greedy:** не всегда optimal solution

---

**Удачи на Senior собесе!** 🧠
