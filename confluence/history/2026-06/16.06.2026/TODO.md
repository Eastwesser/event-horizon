🎯 План на сегодня: Мониторинг

Цель

Развернуть полноценный стек мониторинга для EventHorizon:

Prometheus — сбор метрик
Grafana — визуализация
Jaeger — распределённый трейсинг (если успеем)
Что даст мониторинг:

Компонент	Что даёт	Для чего нужно
Prometheus	Сбор RPS, latency, ошибок	Видеть нагрузку на сервисы
Grafana	Дашборды с графиками	Визуально следить за состоянием системы
Jaeger	Трейсинг запросов	Отслеживать, где тормозит (Gateway → gRPC → БД)
📋 План действий

1. Подготовка инфраструктуры

Добавим в docker-compose.cluster.yml новые сервисы:

yaml
services:
  # ... существующие сервисы (postgres, redis, nats, etc.)

  prometheus:
    image: prom/prometheus:latest
    container_name: event-horizon-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./deployments/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - event-horizon-net

  grafana:
    image: grafana/grafana:latest
    container_name: event-horizon-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_SECURITY_ADMIN_USER=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./deployments/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./deployments/grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - event-horizon-net

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: event-horizon-jaeger
    ports:
      - "16686:16686"   # UI
      - "4318:4318"     # OTLP HTTP
      - "4317:4317"     # OTLP gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - event-horizon-net

volumes:
  prometheus_data:
  grafana_data:

networks:
  event-horizon-net:
    driver: bridge
2. Конфиг Prometheus

Создадим deployments/prometheus.yml:

yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gateway'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'

  - job_name: 'auth'
    static_configs:
      - targets: ['host.docker.internal:50051']
    metrics_path: '/metrics'

  - job_name: 'game'
    static_configs:
      - targets: ['host.docker.internal:50052']
    metrics_path: '/metrics'

  - job_name: 'billing'
    static_configs:
      - targets: ['host.docker.internal:50053']
    metrics_path: '/metrics'

  - job_name: 'leaderboard'
    static_configs:
      - targets: ['host.docker.internal:50054']
    metrics_path: '/metrics'

  - job_name: 'nats'
    static_configs:
      - targets: ['host.docker.internal:8222']
    metrics_path: '/metrics'
3. Добавить метрики в Go сервисы

В каждый сервис (auth, game, billing, leaderboard, gateway) добавим эндпоинт /metrics:

go
// В main.go каждого сервиса
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Метрики
var (
    // Счётчики запросов
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "service_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "endpoint", "status"},
    )

    // Latency гистограмма
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "service_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Активные соединения (WebSocket, gRPC)
    ActiveConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "service_active_connections",
            Help: "Active connections",
        },
        []string{"type"},
    )
)

// В main добавить эндпоинт /metrics
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
4. Дашборды в Grafana

Создадим дашборды для:

Дашборд	Что показывает
Overview	Общий RPS, ошибки, latency по всем сервисам
Gateway	RPS по эндпоинтам, WebSocket соединения
Game	Количество игр, сохранений, ошибок валидации
Billing	Транзакции в секунду, балансы
Leaderboard	Обновления топа, размер Redis
NATS	Длина очереди, publish/subscribe rate
Jaeger	Трейсинг запросов
✅ Сегодняшний чек-лист

Шаг	Задача	Статус
1	Добавить Prometheus и Grafana в docker-compose	⬜️
2	Настроить конфиг Prometheus	⬜️
3	Добавить /metrics во все сервисы	⬜️
4	Добавить базовые метрики (RPS, latency, errors)	⬜️
5	Запустить и проверить сбор метрик	⬜️
6	Настроить Grafana дашборды	⬜️
7	(Опционально) Добавить Jaeger	⬜️
8	Написать алерты (если успеем)	⬜️
