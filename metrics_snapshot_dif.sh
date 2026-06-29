#!/bin/bash

# ============================================================
# Скрипт сбора метрик Event Horizon с АВТОМАТИЧЕСКИМ СРАВНЕНИЕМ
# Использование: ./metrics_snapshot_dif.sh [load_duration] [vus]
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
ALERT_THRESHOLD=30      # Процент, при превышении которого скрипт ругается

mkdir -p "$OUTPUT_DIR"

# --- Функция для запроса метрики из Prometheus ---
query_metric() {
    local query=$1
    local result=$(curl -s -G --data-urlencode "query=$query" "${PROMETHEUS_URL}/api/v1/query" | jq -r '.data.result[0].value[1] // "null"')
    echo "$result"
}

# --- Функция для сравнения двух чисел и подсчета процента ---
calc_diff() {
    local current=$1
    local previous=$2
    
    # Если предыдущее значение null или 0, не считаем разницу
    if [[ "$previous" == "null" || "$previous" == "0" || "$previous" == "" ]]; then
        echo "N/A (no baseline)"
        return
    fi

    # Если текущее значение null, считаем его как 0 для арифметики
    if [[ "$current" == "null" || "$current" == "" ]]; then
        current=0
    fi

    # Вычисляем разницу в процентах с помощью Python (для точности с плавающей точкой)
    local diff_percent=$(python3 -c "
prev = float('$previous')
curr = float('$current')
if prev == 0:
    print('N/A (zero baseline)')
else:
    diff = ((curr - prev) / prev) * 100
    sign = '+' if diff > 0 else ''
    print(f'{sign}{diff:.1f}%')
")
    echo "$diff_percent"
}

# --- Функция для красивого вывода строки с метрикой и диффом ---
print_metric() {
    local name=$1
    local current=$2
    local previous=$3
    local format=$4  # 'raw', 'bytes', 'ms', 'percent'

    # Форматируем текущее значение
    case $format in
        bytes)
            current_fmt=$(numfmt --to=iec --suffix=B "$current" 2>/dev/null || echo "${current} B")
            previous_fmt=$(numfmt --to=iec --suffix=B "$previous" 2>/dev/null || echo "${previous} B")
            ;;
        ms)
            current_fmt=$(printf "%.2f ms" "$current")
            previous_fmt=$(printf "%.2f ms" "$previous")
            ;;
        percent)
            current_fmt=$(printf "%.2f%%" "$current")
            previous_fmt=$(printf "%.2f%%" "$previous")
            ;;
        *)
            current_fmt=$(printf "%.2f" "$current")
            previous_fmt=$(printf "%.2f" "$previous")
            ;;
    esac

    diff=$(calc_diff "$current" "$previous")
    
    # Проверяем, не превышает ли изменение порог (для P99 и Error Rate)
    local alert_msg=""
    if [[ "$name" == *"P99"* || "$name" == *"Error Rate"* ]]; then
        if [[ "$diff" != *"N/A"* ]]; then
            local diff_num=$(echo "$diff" | sed 's/%//')
            if (( $(echo "$diff_num > $ALERT_THRESHOLD" | bc -l) )); then
                alert_msg=" ⚠️ ПРЕВЫШЕНИЕ ПОРОГА!"
            elif (( $(echo "$diff_num < -$ALERT_THRESHOLD" | bc -l) )); then
                alert_msg=" 🎉 УЛУЧШЕНИЕ!"
            fi
        fi
    fi

    printf "  %-35s %18s (было: %12s) [изм: %8s]%s\n" \
        "$name" \
        "$current_fmt" \
        "$previous_fmt" \
        "$diff" \
        "$alert_msg"
}

echo "🚀 Запуск сбора метрик Event Horizon (с диффом)"
echo "📅 Время: $(date)"
echo "⚙️  Нагрузка: ${VUS} пользователей, ${LOAD_DURATION} секунд"
echo "📁 Результат: ${OUTPUT_FILE}"
echo ""

# --- Ищем последний снапшот для сравнения ---
LAST_SNAPSHOT=$(ls -t "$OUTPUT_DIR"/snapshot_*.json 2>/dev/null | head -n1)

if [[ -n "$LAST_SNAPSHOT" ]]; then
    echo "📂 Найден предыдущий снапшот для сравнения: $(basename "$LAST_SNAPSHOT")"
    PREVIOUS_DATA=$(cat "$LAST_SNAPSHOT")
else
    echo "📂 Предыдущих снапшотов не найдено. Первый запуск."
    PREVIOUS_DATA="{}"
fi
echo ""

# --- Фаза 1: Нагрузочный тест (в фоне) ---
echo "🔥 Запуск нагрузочного теста k6..."
cd "$K6_DIR"
k6 run --vus "$VUS" --duration "${LOAD_DURATION}s" e2e-test.js > /tmp/k6_output.log 2>&1 &
K6_PID=$!
cd - > /dev/null

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

# --- Фаза 4: Красивый вывод с диффом ---
echo ""
echo "📊 СВОДКА МЕТРИК (сравнение с предыдущим запуском)"
echo "═══════════════════════════════════════════════════════════════════════════════════"

# Достаем предыдущие значения
PREV_RPS=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.rps.total // "null"')
PREV_LATENCY=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.latency.p99_seconds // "null"')
PREV_ERROR_RATE=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.rps.error_rate_percent // "null"')
PREV_GOROUTINES=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.go_runtime.goroutines // "null"')
PREV_HEAP=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.go_runtime.heap_alloc_bytes // "null"')
PREV_NATS_CONNS=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.nats.connections // "null"')
PREV_NATS_IN=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.nats.in_msgs_per_second // "null"')
PREV_NATS_OUT=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.nats.out_msgs_per_second // "null"')
PREV_GAME_SUBMITS=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.business.game_submits_per_second // "null"')
PREV_BALANCER_CONNS=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.business.balancer_active_connections // "null"')
PREV_REDIS_MEMORY=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.storage.redis_memory_bytes // "null"')
PREV_PG_CONNS=$(echo "$PREVIOUS_DATA" | jq -r '.metrics.storage.postgres_connections // "null"')

# Печатаем метрики
print_metric "Total RPS" "$RPS" "$PREV_RPS" "raw"
print_metric "Read RPS" "$RPS_READ" "$PREV_RPS" "raw"
print_metric "Write RPS" "$RPS_WRITE" "$PREV_RPS" "raw"
print_metric "Error Rate" "$ERROR_RATE" "$PREV_ERROR_RATE" "percent"
print_metric "Latency P99" "$LATENCY_P99" "$PREV_LATENCY" "ms"
print_metric "Goroutines (Game)" "$GOROUTINES" "$PREV_GOROUTINES" "raw"
print_metric "Heap (Game)" "$HEAP" "$PREV_HEAP" "bytes"
print_metric "NATS Connections" "$NATS_CONNS" "$PREV_NATS_CONNS" "raw"
print_metric "NATS In msgs/s" "$NATS_IN" "$PREV_NATS_IN" "raw"
print_metric "NATS Out msgs/s" "$NATS_OUT" "$PREV_NATS_OUT" "raw"
print_metric "Game Submits/s" "$GAME_SUBMITS" "$PREV_GAME_SUBMITS" "raw"
print_metric "Balancer Connections" "$BALANCER_CONNS" "$PREV_BALANCER_CONNS" "raw"
print_metric "Redis Memory" "$REDIS_MEMORY" "$PREV_REDIS_MEMORY" "bytes"
print_metric "PG Connections" "$PG_CONNS" "$PREV_PG_CONNS" "raw"

echo "═══════════════════════════════════════════════════════════════════════════════════"
echo ""
echo "✅ Готово!"
echo "📁 Результаты сохранены в: $OUTPUT_FILE"
echo ""
echo "📊 Для просмотра детального JSON, выполните:"
echo "   cat $OUTPUT_FILE | jq '.'"
echo ""
echo "📈 Для просмотра в Grafana, откройте: http://localhost:3000"