#!/bin/bash

# ============================================================
# Скрипт сбора метрик Event Horizon
# Использование: ./metrics_snapshot.sh [load_duration] [vus]
# Пример: ./metrics_snapshot.sh 30 10
# ============================================================

set -e

# --- Настройки ---
PROMETHEUS_URL="http://localhost:9090"
K6_DIR="./deployments/k6"
OUTPUT_DIR="./metrics_snapshots"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOAD_DURATION=${1:-30}  # По умолчанию 30 секунд
VUS=${2:-10}            # По умолчанию 10 пользователей
OUTPUT_FILE="${OUTPUT_DIR}/snapshot_${TIMESTAMP}.json"

mkdir -p "$OUTPUT_DIR"

# --- Функция для запроса метрики из Prometheus ---
query_metric() {
    local query=$1
    local result=$(curl -s -G --data-urlencode "query=$query" "${PROMETHEUS_URL}/api/v1/query" | jq -r '.data.result[0].value[1] // "null"')
    echo "$result"
}

echo "🚀 Запуск сбора метрик Event Horizon"
echo "📅 Время: $(date)"
echo "⚙️  Нагрузка: ${VUS} пользователей, ${LOAD_DURATION} секунд"
echo "📁 Результат: ${OUTPUT_FILE}"
echo ""

# --- Фаза 1: Нагрузочный тест (в фоне) ---
echo "🔥 Запуск нагрузочного теста k6..."
cd "$K6_DIR"
k6 run --vus "$VUS" --duration "${LOAD_DURATION}s" e2e-test.js > /tmp/k6_output.log 2>&1 &
K6_PID=$!
cd - > /dev/null

# Даем системе время на разогрев перед сбором метрик
echo "⏳ Ожидание 5 секунд для прогрева системы..."
sleep 5

# --- Фаза 2: Сбор метрик ---
echo "📊 Сбор метрик из Prometheus..."

# 1. Базовые метрики
RPS=$(query_metric "sum(rate(gateway_requests_total[1m]))")
RPS_READ=$(query_metric "sum(rate(gateway_requests_total{method='GET'}[1m]))")
RPS_WRITE=$(query_metric "sum(rate(gateway_requests_total{method='POST'}[1m]))")
ERROR_RATE=$(query_metric "sum(rate(gateway_requests_total{status=~'5..'}[1m])) / sum(rate(gateway_requests_total[1m])) * 100")
LATENCY_P99=$(query_metric "histogram_quantile(0.99, sum(rate(gateway_request_duration_seconds_bucket[1m])) by (le))")

# 2. Go Runtime
GOROUTINES=$(query_metric "go_goroutines{job='game'}")
GC_CYCLES=$(query_metric "rate(go_gc_cycles_automatic_gc_cycles_total{job='game'}[1m])")
HEAP=$(query_metric "go_memstats_heap_alloc_bytes{job='game'}")
MUTEX=$(query_metric "rate(go_mutex_wait_total_seconds_total{job='game'}[1m])")

# 3. NATS
NATS_CONNS=$(query_metric "gnatsd_varz_connections")
NATS_IN=$(query_metric "rate(gnatsd_varz_in_msgs[1m])")
NATS_OUT=$(query_metric "rate(gnatsd_varz_out_msgs[1m])")

# 4. Бизнес-метрики
BALANCER_CONNS=$(query_metric "balancer_active_connections")
GAME_SUBMITS=$(query_metric "sum(rate(game_submits_total[1m]))")

# 5. Redis (если есть экспортер)
REDIS_MEMORY=$(query_metric "redis_memory_used_bytes{job='redis'}" || echo "null")
REDIS_CONNS=$(query_metric "redis_connected_clients{job='redis'}" || echo "null")

# 6. PostgreSQL (если есть экспортер)
PG_CONNS=$(query_metric "pg_stat_database_numbackends{datname='eventhorizon'}" || echo "null")
PG_DEADLOCKS=$(query_metric "pg_stat_database_deadlocks{datname='eventhorizon'}" || echo "null")

# Дожидаемся завершения k6
echo "⏳ Ожидание завершения нагрузочного теста..."
wait $K6_PID
echo "✅ Нагрузочный тест завершен."

# --- Фаза 3: Сохранение результатов в JSON ---
echo "💾 Сохранение результатов..."

cat > "$OUTPUT_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "test_config": {
    "duration_seconds": $LOAD_DURATION,
    "vus": $VUS
  },
  "metrics": {
    "rps": {
      "total": $RPS,
      "read": $RPS_READ,
      "write": $RPS_WRITE,
      "error_rate_percent": $ERROR_RATE
    },
    "latency": {
      "p99_seconds": $LATENCY_P99
    },
    "go_runtime": {
      "goroutines": $GOROUTINES,
      "gc_cycles_per_second": $GC_CYCLES,
      "heap_alloc_bytes": $HEAP,
      "mutex_wait_seconds_per_second": $MUTEX
    },
    "nats": {
      "connections": $NATS_CONNS,
      "in_msgs_per_second": $NATS_IN,
      "out_msgs_per_second": $NATS_OUT
    },
    "business": {
      "balancer_active_connections": $BALANCER_CONNS,
      "game_submits_per_second": $GAME_SUBMITS
    },
    "storage": {
      "redis_memory_bytes": $REDIS_MEMORY,
      "redis_connections": $REDIS_CONNS,
      "postgres_connections": $PG_CONNS,
      "postgres_deadlocks": $PG_DEADLOCKS
    }
  }
}
EOF

echo ""
echo "✅ Готово!"
echo "📁 Результаты сохранены в: $OUTPUT_FILE"
echo ""
echo "📊 Для просмотра метрик в консоли, выполните:"
echo "   cat $OUTPUT_FILE | jq '.'"
echo ""
echo "📈 Для просмотра в Grafana, откройте: http://localhost:3000"