Разбор balancer
Что делает сервис
balancer — это L4/L7 балансировщик на Least Connections:

Принимает HTTP на :8079.

Выбирает backend с наименьшим числом активных соединений.

Проксирует через httputil.ReverseProxy.

Экспортирует метрики на :9098.

Graceful shutdown через closer.

Нет БД, нет gRPC, нет NATS. Только HTTP-in, HTTP-out.

Найденные паттерны
1. Manual DI (без контейнера)
Файл: internal/app/app.go, App.New.

Почему: Как и в fulfillment, здесь нет diContainer. Всё собирается прямо в New:

Logger.

Closer.

Backends.

HTTP-серверы.

Код:

go
func New(ctx context.Context) (*App, error) {
    log := logger.NewFrom(level, format)
    a := &App{log: log, closer: closer.New(log)}

    backends := []string{
        getenv("BALANCER_BACKEND_1", "http://gateway:8080"),
        getenv("BALANCER_BACKEND_2", "http://gateway-2:8080"),
        getenv("BALANCER_BACKEND_3", "http://gateway-3:8080"),
    }
    lb := balancer.NewLeastConnBalancer(backends)

    a.httpServer = &http.Server{Addr: addr, Handler: lb}
    // ...
}
Manual DI — сервис настолько мал, что контейнер не нужен.

2. Strategy: Least Connections (главный паттерн)
Файл: internal/balancer/least_conn.go, getLeastConnBackend.

Почему: Алгоритм балансировки — Least Connections: выбирается backend с наименьшим ActiveConns.

Код:

go
func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    if len(lb.backends) == 0 {
        return nil
    }

    var selected *Backend
    var minConns int32 = 2147483647 // MaxInt32

    for _, b := range lb.backends {
        conns := atomic.LoadInt32(&b.ActiveConns)
        if conns < minConns {
            minConns = conns
            selected = b
        }
    }
    return selected
}
Strategy Pattern — алгоритм балансировки. Уникально для balancer.

Плюс: mu.RLock() — read lock для параллельного чтения (алгоритм только читает ActiveConns).

3. Atomic Counter (ActiveConns)
Файл: internal/balancer/least_conn.go.

Почему: Счётчик активных соединений — int32 с атомарными операциями. Не мьютекс, а lock-free.

Код:

go
atomic.AddInt32(&backend.ActiveConns, 1)
defer atomic.AddInt32(&backend.ActiveConns, -1)
Atomic Counter — lock-free счётчик. Правильное решение для high-throughput.

4. Reverse Proxy (httputil.ReverseProxy)
Файл: internal/balancer/least_conn.go.

Почему: Каждый backend имеет свой httputil.ReverseProxy. Это стандартная библиотека Go для reverse proxy.

Код:

go
backend := &Backend{
    URL:   u,
    Proxy: httputil.NewSingleHostReverseProxy(u),
}
backend.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
    log.Printf("Proxy error: %v", err)
    http.Error(w, "Backend error", http.StatusBadGateway)
}
Reverse Proxy — стандартный приём. Уникально для balancer.

Плюс: Кастомный ErrorHandler с логированием и 502.

5. Registry (in-memory) — []*Backend
Файл: internal/balancer/least_conn.go.

Почему: Реестр бэкендов — простой []*Backend. Нет service discovery, нет health checks. Статический реестр из ENV.

Код:

go
backends := []string{
    getenv("BALANCER_BACKEND_1", "http://gateway:8080"),
    getenv("BALANCER_BACKEND_2", "http://gateway-2:8080"),
    getenv("BALANCER_BACKEND_3", "http://gateway-3:8080"),
}
Static Registry — простой реестр из переменных окружения.

Заметка: папка internal/backend/ есть, но файл backend.go пустой. Реестр живёт в balancer/least_conn.go.

6. Health Check (HTTP) — через /health и /ready
Файл: internal/app/metrics.go.

Код:

go
mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ok","service":"balancer"}`))
})
mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ready","service":"balancer"}`))
})
Health Check — простой health/ready.

Заметка: /ready всегда возвращает ready — не проверяет доступность backend'ов. Это упрощение.

7. Prometheus Metrics (Gauge + Counter)
Файл: internal/metrics/metrics.go.

Код:

go
var (
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "balancer_active_connections",
            Help: "Total active connections across all backends",
        },
    )

    RequestsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "balancer_requests_total",
            Help: "Total number of requests processed by balancer",
        },
    )
)
Prometheus Metrics — Gauge (текущее значение) + Counter (монотонный). Уникально для balancer — тут именно Gauge, а не Counter/Histogram.

8. Metrics Update на каждом запросе
Файл: internal/balancer/least_conn.go, ServeHTTP.

Код:

go
func (lb *LeastConnBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    metrics.RequestsTotal.Inc()
    metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
    // ...
}
Metrics Update — метрики обновляются на каждом запросе.

9. X-Balancer-Backend Header
Файл: internal/balancer/least_conn.go.

Почему: В запрос добавляется заголовок с именем backend'а — для трассировки.

Код:

go
r.Header.Set("X-Balancer-Backend", backend.URL.Host)
Header Injection for Observability — полезный приём для отладки.

10. Graceful Shutdown / Closer
Файл: internal/app/app.go.

Код:

go
a.closer.AddNamed("http server", func(ctx context.Context) error {
    return a.httpServer.Shutdown(ctx)
})
a.closer.AddNamed("metrics http", func(ctx context.Context) error {
    return a.metricsSrv.Shutdown(ctx)
})
Closer Registry — как в других сервисах.

11. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal.

Код:

go
if err := a.closer.WaitSignal(ctx); err != nil {
    return fmt.Errorf("shutdown: %w", err)
}
Context Cancellation — через closer.

12. Empty Packages (backend, proxy)
Файлы: internal/backend/backend.go, internal/proxy/proxy.go — пустые.

Проблема: Пакеты созданы, но не используются. Это мёртвый код или заготовки на будущее.

Паттерн: Code Skeleton (заготовки).

13. Testable Balancer
Файл: internal/balancer/least_conn_test.go.

Код:

go
func TestGetLeastConnBackend_PicksLowest(t *testing.T) {
    lb := NewLeastConnBalancer([]string{
        "http://gw-busy:8080",
        "http://gw-idle:8080",
        "http://gw-mid:8080",
    })

    atomic.StoreInt32(&lb.backends[0].ActiveConns, 10)
    atomic.StoreInt32(&lb.backends[1].ActiveConns, 1)
    atomic.StoreInt32(&lb.backends[2].ActiveConns, 4)

    got := lb.getLeastConnBackend()
    if got.URL.Host != "gw-idle:8080" {
        t.Fatalf("host=%s want gw-idle:8080", got.URL.Host)
    }
}
Testable Balancer — тест проверяет выбор backend'а.

Плюс: Тест на пустой список (TestGetLeastConnBackend_Empty).

14. Env-Based Config
Файл: internal/app/app.go, getenv.

Код:

go
func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}
Env-Based Config — простая конфигурация через ENV.

Заметка: Нет отдельного пакета config/. Всё в app.go.

15. Dual HTTP Server (traffic + metrics)
Файл: internal/app/app.go.

Почему: Два HTTP-сервера:

:8079 — трафик (backend proxy).

:9098 — метрики + health.

Код:

go
a.httpServer = &http.Server{Addr: addr, Handler: lb}
a.metricsSrv = &http.Server{Addr: metricsAddr, Handler: metricsHandler()}
Dual HTTP Server — разделение трафика и метрик.

Итог по balancer
Паттерн	Где
Manual DI	app/app.go, App.New
Strategy: Least Connections	balancer/least_conn.go, getLeastConnBackend
Atomic Counter	balancer/least_conn.go, ActiveConns
Reverse Proxy	balancer/least_conn.go, httputil.NewSingleHostReverseProxy
Static Registry	app/app.go, backends
Health Check (HTTP)	app/metrics.go
Prometheus Metrics (Gauge + Counter)	metrics/metrics.go
Metrics Update на каждом запросе	balancer/least_conn.go, ServeHTTP
X-Balancer-Backend Header	balancer/least_conn.go
Graceful Shutdown / Closer	app/app.go
Context Cancellation	app/app.go, RunUntilSignal
Code Skeleton (пустые пакеты)	backend/backend.go, proxy/proxy.go
Testable Balancer	balancer/least_conn_test.go
Env-Based Config	app/app.go, getenv
Dual HTTP Server	app/app.go
Итого 15 паттернов и приёмов — для такого маленького сервиса это хороший результат.

Уникальные фичи balancer:

Least Connections — единственный сервис с балансировкой.

Atomic Counter для ActiveConns — правильный lock-free подход.

Reverse Proxy — единственный сервис с проксированием.

Prometheus Gauge — единственный сервис с Gauge (в остальных Counter/Histogram).

X-Balancer-Backend — заголовок для трассировки.

Анти-паттерны / заготовки:

Пустые пакеты backend/ и proxy/ — мёртвый код.

/ready не проверяет бэкенды — всегда говорит ready.

Static Registry — нет health checks, нет service discovery.

Готовый markdown для BALANCER_PATTERNS.md
md
# balancer — Programming Patterns

> Сервис: Least Connections балансировщик (HTTP reverse proxy).
> Только HTTP-in, HTTP-out. Prometheus metrics.

## Что делает сервис

Принимает HTTP на :8079, выбирает backend с наименьшим числом активных соединений, проксирует через `httputil.ReverseProxy`. Метрики на :9098.

## Паттерны

### 1. Manual DI (без контейнера)
- **Файл:** `internal/app/app.go`, `App.New`

### 2. Strategy: Least Connections
- **Файл:** `internal/balancer/least_conn.go`, `getLeastConnBackend`
- **Код:**
  ```go
  for _, b := range lb.backends {
      conns := atomic.LoadInt32(&b.ActiveConns)
      if conns < minConns {
          minConns = conns
          selected = b
      }
  }
3. Atomic Counter
Файл: internal/balancer/least_conn.go, ActiveConns

Код: atomic.AddInt32(&backend.ActiveConns, 1)

4. Reverse Proxy
Файл: internal/balancer/least_conn.go

Код: httputil.NewSingleHostReverseProxy(u)

5. Static Registry
Файл: internal/app/app.go, backends

6. Health Check (HTTP)
Файл: internal/app/metrics.go

7. Prometheus Metrics (Gauge + Counter)
Файл: internal/metrics/metrics.go

Код: promauto.NewGauge("balancer_active_connections", ...) + promauto.NewCounter("balancer_requests_total", ...)

8. Metrics Update на каждом запросе
Файл: internal/balancer/least_conn.go, ServeHTTP

9. X-Balancer-Backend Header
Файл: internal/balancer/least_conn.go

Код: r.Header.Set("X-Balancer-Backend", backend.URL.Host)

10. Graceful Shutdown / Closer
Файл: internal/app/app.go

11. Context Cancellation
Файл: internal/app/app.go, RunUntilSignal

12. Code Skeleton (пустые пакеты) — заготовки
Файлы: internal/backend/backend.go, internal/proxy/proxy.go

13. Testable Balancer
Файл: internal/balancer/least_conn_test.go

14. Env-Based Config
Файл: internal/app/app.go, getenv

15. Dual HTTP Server
Файл: internal/app/app.go

Код: a.httpServer (трафик) + a.metricsSrv (метрики).