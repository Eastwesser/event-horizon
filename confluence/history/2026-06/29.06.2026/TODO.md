[denismatveev@c0der event_horizon]$ chmod +x metrics_collector.sh
./metrics_collector.sh
📊 Сбор метрик Event Horizon (без нагрузки, только текущие значения)
📅 Время: Пн 29 июн 2026 19:21:01 MSK
✅ Метрики сохранены в: ./metrics_snapshots/snapshot_20260629_192101.json

📊 Содержимое:
{
  "timestamp": "2026-06-29T19:21:01+03:00",
  "metrics": {
    "game": {
      "submits_total": 122,
      "goroutines": 29,
      "heap_alloc_bytes": 4987112,
      "score_count": 122
    },
    "nats": {
      "connections": 6,
      "in_msgs_total": 2576,
      "out_msgs_total": 503
    },
    "gateway": {
      "instance_1_requests": 247,
      "instance_2_requests": 229,
      "instance_3_requests": 208
    },
    "balancer": {
      "active_connections": null
    },
    "storage": {
      "redis_memory_bytes": null,
      "postgres_connections": null
    }
  }
}
[denismatveev@c0der event_horizon]$ 

НАМ НУЖНО СОБРАТЬ ВСЕ НЕДОСТАЮЩИЕ МЕТРИКИ!

📝 TODO — 29.06.2026 (Цель: Полный мониторинг)
1. Починить Gateway метрики (RPS, Latency)
Проверить docker-compose.cluster.yml — у всех Gateway есть METRICS_PORT=9095-9097.

Проверить prometheus.yml — есть ли job для gateway с тремя таргетами.

Добавить в services/gateway/cmd/main.go регистрацию метрик: prometheus.MustRegister(gatewayRequestsTotal, gatewayRequestDuration).

Пересобрать Gateway и перезапустить.

2. Добавить метрики в Balancer
Создать файл services/balancer/internal/metrics/metrics.go с кастомными метриками.

Зарегистрировать balancer_active_connections (Gauge) и balancer_requests_total (Counter).

Инкрементить метрики в least_conn.go при проксировании запросов.

Пересобрать Balancer и перезапустить.

3. Настроить Redis Exporter
Добавить сервис redis-exporter в docker-compose.cluster.yml.

Подключить его к Redis (порт 6379).

Добавить job в prometheus.yml.

4. Настроить PostgreSQL Exporter
Добавить сервис postgres-exporter в docker-compose.cluster.yml.

Подключить к PostgreSQL (порт 5432).

Добавить job в prometheus.yml.

5. Обновить скрипт сбора метрик
Добавить в metrics_collector.sh запросы для новых метрик (Redis, PostgreSQL, Balancer).

Добавить в JSON-вывод новые секции.

Проверить, что все метрики не null.

6. Создать дашборд в Grafana
Создать папку Event Horizon.

Добавить панели для:

RPS (по Gateway).

Latency (P50, P95, P99).

Go Runtime (Goroutines, Heap, GC).

NATS (Connections, In/Out msgs).

Game Submits (по играм).

Balancer Connections.

Redis Memory/Connections.

PostgreSQL Connections/Deadlocks.

🔥 Приступаем к реализации
Я разбил задачи на логические блоки. Мы можем делать их по порядку или выбрать самые важные. 
Я предлагаю начать с пункта 1 и 2, потому что это даст нам RPS и Latency — основу любого мониторинга.

Задача 1: Починить Gateway метрики
Шаг 1.1: Проверь текущие метрики Gateway
bash
# Проверь, что метрики вообще регистрируются
curl -s http://localhost:9095/metrics | head -20
Если там нет gateway_requests_total, значит, метрики не зарегистрированы.

Шаг 1.2: Добавь регистрацию в services/gateway/cmd/main.go
Найди в файле main() строчки, где создаются метрики:

go
gatewayRequestsTotal := promauto.NewCounterVec(...)
gatewayRequestDuration := promauto.NewHistogramVec(...)
Добавь после них:

go
prometheus.MustRegister(gatewayRequestsTotal)
prometheus.MustRegister(gatewayRequestDuration)
Шаг 1.3: Пересобери и перезапусти Gateway
bash
cd ~/event_horizon/services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker push eastwesser/gateway:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3
Шаг 1.4: Проверь метрики снова
bash
curl -s http://localhost:9095/metrics | grep gateway_requests_total
Задача 2: Добавить метрики в Balancer
Шаг 2.1: Создай файл services/balancer/internal/metrics/metrics.go
go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

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
Шаг 2.2: Используй метрики в least_conn.go
Найди место, где обрабатывается запрос. Добавь:

go
import "github.com/Eastwesser/event-horizon/services/balancer/internal/metrics"

func (lb *LeastConnBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    metrics.RequestsTotal.Inc()
    metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
    // ... остальной код
}
Шаг 2.3: Пересобери и перезапусти Balancer
bash
cd ~/event_horizon/services/balancer
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o balancer ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.balancer.bin -t eastwesser/balancer:latest .
docker push eastwesser/balancer:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d balancer
Задача 3: Настроить Redis Exporter
Добавь в docker-compose.cluster.yml:

yaml
redis-exporter:
  image: oliver006/redis_exporter:latest
  container_name: event-horizon-redis-exporter
  environment:
    REDIS_ADDR: "redis://event-horizon-redis:6379"
  ports:
    - "9121:9121"
  networks:
    - event-horizon-net
В prometheus.yml добавь job:

yaml
- job_name: 'redis'
  static_configs:
    - targets: ['event-horizon-redis-exporter:9121']
Задача 4: Обновить скрипт сбора метрик
Обнови metrics_collector.sh, добавив запросы:

bash
# Redis
REDIS_MEM=$(query_metric "redis_memory_used_bytes")
REDIS_CONNS=$(query_metric "redis_connected_clients")

# PostgreSQL
PG_CONNS=$(query_metric "pg_stat_database_numbackends{datname='eventhorizon'}")

# Balancer
BALANCER_CONNS=$(query_metric "balancer_active_connections")
Задача 5: Создать дашборд в Grafana
Зайди в Grafana (http://localhost:3000).

Создай папку Event Horizon.

Создай дашборд с панелями:

RPS (Gateway): sum(rate(gateway_requests_total[1m])) by (instance)

Latency P99: histogram_quantile(0.99, sum(rate(gateway_request_duration_seconds_bucket[1m])) by (le))

Goroutines: go_goroutines{job="game"}

NATS Connections: gnatsd_varz_connections

Game Submits: sum(game_submits_total) by (game_id)

🚀 План действий на сегодня
Сделай пункт 1 (Gateway метрики) — это база.

Сделай пункт 2 (Balancer метрики) — это важно.

Обнови скрипт и снова запусти его, чтобы увидеть новые метрики.

Настрой Redis Exporter — добавим метрики по кешу.

Создай дашборд в Grafana — чтобы всё видеть глазами.

ПЕРЕСБОРКА - REBUILD
cd ~/event_horizon/services/balancer
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o balancer ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.balancer.bin -t eastwesser/balancer:latest .
docker push eastwesser/balancer:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d balancer

cd ~/event_horizon/services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker push eastwesser/gateway:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3