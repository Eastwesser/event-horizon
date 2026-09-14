# Паттерн: Load Balancer — Распределение нагрузки

## 🎯 Задача

Распределить **N задач между M воркерами** равномерно, чтобы избежать перегрузки одного воркера.

## 🗡️ Стратегии балансировки

### 1. Round Robin — по кругу

```go
type RoundRobinBalancer struct {
	workers []chan Job
	current int
	mu      sync.Mutex
}

func (rb *RoundRobinBalancer) Next() chan Job {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	worker := rb.workers[rb.current]
	rb.current = (rb.current + 1) % len(rb.workers)
	return worker
}

func main() {
	balancer := NewRoundRobinBalancer(5)  // 5 воркеров
	
	for _, job := range jobs {
		balancer.Next() <- job  // Распределение по кругу
	}
}
```

**Характеристики**:
- Самый простой алгоритм
- Равномерное распределение **по количеству**
- Не учитывает нагрузку воркеров

### 2. Least Connections — наименьшая нагрузка

```go
type LeastConnBalancer struct {
	workers []Worker
	mu      sync.Mutex
}

type Worker struct {
	jobs   chan Job
	active int32  // atomic counter
}

func (lb *LeastConnBalancer) Next() *Worker {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	minWorker := &lb.workers[0]
	minLoad := atomic.LoadInt32(&minWorker.active)
	
	for i := 1; i < len(lb.workers); i++ {
		load := atomic.LoadInt32(&lb.workers[i].active)
		if load < minLoad {
			minLoad = load
			minWorker = &lb.workers[i]
		}
	}
	
	atomic.AddInt32(&minWorker.active, 1)
	return minWorker
}

func (w *Worker) process(job Job) {
	defer atomic.AddInt32(&w.active, -1)
	// Process job
}
```

**Характеристики**:
- Выбирает наименее загруженный воркер
- Лучше для задач с разной длительностью
- Больше overhead (atomic операции)

### 3. Random — случайный выбор

```go
type RandomBalancer struct {
	workers []chan Job
}

func (rb *RandomBalancer) Next() chan Job {
	return rb.workers[rand.Intn(len(rb.workers))]
}
```

**Характеристики**:
- Простейшая реализация
- Статистически равномерно
- Нет state (без mutex)

### 4. Weighted — взвешенное распределение

```go
type WeightedBalancer struct {
	workers []Worker
	weights []int  // [1, 2, 3] → worker 2 получит в 2 раза больше задач
	current int
	gcd     int  // greatest common divisor
}

func (wb *WeightedBalancer) Next() chan Job {
	for {
		wb.current = (wb.current + 1) % len(wb.workers)
		
		if wb.current == 0 {
			wb.currentWeight -= wb.gcd
			if wb.currentWeight <= 0 {
				wb.currentWeight = wb.maxWeight
			}
		}
		
		if wb.weights[wb.current] >= wb.currentWeight {
			return wb.workers[wb.current].jobs
		}
	}
}
```

**Характеристики**:
- Учитывает мощность воркеров (CPU, память)
- Сложная логика
- Используется в Nginx

## 💡 Правило Додзё

> **"Load Balancer — это мост с 5 дорожками. Round Robin отправляет повозки по очереди. Least Connections отправляет на самую свободную. Weighted отдает предпочтение широким дорожкам."**

## 🧪 Полная реализация Round Robin

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID       int
	Duration time.Duration
}

type Worker struct {
	ID   int
	jobs chan Job
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range w.jobs {
		fmt.Printf("Worker %d processing job %d\n", w.ID, job.ID)
		time.Sleep(job.Duration)
		fmt.Printf("Worker %d finished job %d\n", w.ID, job.ID)
	}
}

type LoadBalancer struct {
	workers []*Worker
	current int
	mu      sync.Mutex
}

func NewLoadBalancer(numWorkers int) *LoadBalancer {
	workers := make([]*Worker, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = &Worker{
			ID:   i,
			jobs: make(chan Job, 10),
		}
	}
	return &LoadBalancer{workers: workers}
}

func (lb *LoadBalancer) Next() chan Job {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	worker := lb.workers[lb.current]
	lb.current = (lb.current + 1) % len(lb.workers)
	return worker.jobs
}

func (lb *LoadBalancer) Start() {
	var wg sync.WaitGroup
	for _, worker := range lb.workers {
		wg.Add(1)
		go worker.Start(&wg)
	}
	wg.Wait()
}

func (lb *LoadBalancer) Stop() {
	for _, worker := range lb.workers {
		close(worker.jobs)
	}
}

func main() {
	lb := NewLoadBalancer(3)
	
	go lb.Start()
	
	// Отправляем 10 задач
	for i := 0; i < 10; i++ {
		job := Job{
			ID:       i,
			Duration: 100 * time.Millisecond,
		}
		lb.Next() <- job
	}
	
	time.Sleep(2 * time.Second)
	lb.Stop()
}
```

**Вывод**:
```
Worker 0 processing job 0
Worker 1 processing job 1
Worker 2 processing job 2
Worker 0 processing job 3
Worker 1 processing job 4
...
```

## 🛠️ Реальные примеры

### 1. HTTP Reverse Proxy

```go
type Server struct {
	URL    string
	Alive  bool
	mu     sync.RWMutex
}

type LoadBalancer struct {
	servers []*Server
	current int
	mu      sync.Mutex
}

func (lb *LoadBalancer) GetNextServer() *Server {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	start := lb.current
	for {
		server := lb.servers[lb.current]
		lb.current = (lb.current + 1) % len(lb.servers)
		
		server.mu.RLock()
		alive := server.Alive
		server.mu.RUnlock()
		
		if alive {
			return server
		}
		
		if lb.current == start {
			return nil  // Все сервера недоступны
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := lb.GetNextServer()
	if server == nil {
		http.Error(w, "Service Unavailable", 503)
		return
	}
	
	proxy := httputil.NewSingleHostReverseProxy(server.URL)
	proxy.ServeHTTP(w, r)
}
```

### 2. Database Connection Pool

```go
type ConnectionPool struct {
	conns   []*sql.DB
	current int
	mu      sync.Mutex
}

func (cp *ConnectionPool) GetConnection() *sql.DB {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	conn := cp.conns[cp.current]
	cp.current = (cp.current + 1) % len(cp.conns)
	return conn
}

func (cp *ConnectionPool) Query(query string) (*sql.Rows, error) {
	conn := cp.GetConnection()
	return conn.Query(query)
}
```

### 3. gRPC Load Balancer

```go
type GRPCBalancer struct {
	clients []pb.ServiceClient
	current int
	mu      sync.Mutex
}

func (gb *GRPCBalancer) Call(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	gb.mu.Lock()
	client := gb.clients[gb.current]
	gb.current = (gb.current + 1) % len(gb.clients)
	gb.mu.Unlock()
	
	return client.Process(ctx, req)
}
```

## 🎓 Сравнение стратегий

| Стратегия | Сложность | Равномерность | Overhead | Use Case |
|-----------|-----------|---------------|----------|----------|
| Round Robin | O(1) | Хорошая | Минимальный | Одинаковые задачи |
| Least Connections | O(N) | Отличная | Средний | Разные задачи |
| Random | O(1) | Статистическая | Нулевой | Stateless |
| Weighted | O(N) | Настраиваемая | Высокий | Разная мощность |

## 🧪 Эксперименты

### Тест 1: Round Robin vs Random

```go
func testDistribution(balancer Balancer, numJobs int) {
	counts := make(map[int]int)
	
	for i := 0; i < numJobs; i++ {
		workerID := balancer.Next()
		counts[workerID]++
	}
	
	fmt.Println(counts)
}

// Round Robin: {0: 34, 1: 33, 2: 33} ← идеально равномерно
// Random:      {0: 37, 1: 29, 2: 34} ← статистически равномерно
```

### Тест 2: Least Connections под нагрузкой

```go
// 3 воркера, задачи с разной длительностью
jobs := []Job{
	{ID: 1, Duration: 1 * time.Second},
	{ID: 2, Duration: 100 * time.Millisecond},
	{ID: 3, Duration: 100 * time.Millisecond},
}

// Round Robin: Worker 0 перегружен (1s задача)
// Least Conn:  Workers 1, 2 обработают больше (100ms задачи)
```

## 🎓 Вопросы для медитации

1. Почему Round Robin плох для задач с разной длительностью?
2. Как реализовать Sticky Sessions (запросы от одного клиента на один воркер)?
3. Чем отличается client-side load balancing от server-side?
4. Как обрабатывать сбой одного из воркеров?

## 🔗 Связанные паттерны

- **Health Checker**: проверка доступности воркеров
- **Circuit Breaker**: отключение недоступных воркеров
- **Worker Pool**: Load Balancer распределяет задачи между воркерами

---

**Задание**: Реализуй Load Balancer с поддержкой:
1. Health checks (периодическая проверка воркеров)
2. Автоматическое исключение недоступных воркеров
3. Sticky sessions (по clientID)
4. Метрики (сколько задач обработал каждый воркер)

Объясни, почему Least Connections лучше для HTTP-сервера, а Round Robin — для batch обработки.
