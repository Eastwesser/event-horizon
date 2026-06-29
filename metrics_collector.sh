#!/bin/bash

# ============================================================
# Сборщик метрик Event Horizon (рабочая версия с правильными именами)
# ============================================================

PROMETHEUS_URL="http://localhost:9090"
OUTPUT_DIR="./metrics_snapshots"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="${OUTPUT_DIR}/snapshot_${TIMESTAMP}.json"

mkdir -p "$OUTPUT_DIR"

# Функция запроса
query_metric() {
    local query=$1
    local result=$(curl -s -G --data-urlencode "query=$query" "${PROMETHEUS_URL}/api/v1/query" | jq -r '.data.result[0].value[1] // "null"')
    echo "$result"
}

echo "📊 Сбор метрик Event Horizon (без нагрузки, только текущие значения)"
echo "📅 Время: $(date)"

# --- Сбор метрик с правильными именами ---

# 1. Game метрики (работают!)
GAME_SUBMITS=$(query_metric "sum(game_submits_total)")
GO_GOROUTINES=$(query_metric "go_goroutines{job='game'}")
GO_HEAP=$(query_metric "go_memstats_heap_alloc_bytes{job='game'}")
GAME_SCORE_COUNT=$(query_metric "game_score_histogram_count{game_id='hexagon'}")

# 2. NATS метрики (работают!)
NATS_CONNS=$(query_metric "gnatsd_varz_connections")
NATS_IN_MSGS=$(query_metric "gnatsd_varz_in_msgs")
NATS_OUT_MSGS=$(query_metric "gnatsd_varz_out_msgs")

# 3. Gateway метрики (проверим все три инстанса)
GW1=$(query_metric "gateway_requests_total{job='gateway', instance='gateway:9095'}")
GW2=$(query_metric "gateway_requests_total{job='gateway', instance='gateway-2:9096'}")
GW3=$(query_metric "gateway_requests_total{job='gateway', instance='gateway-3:9097'}")

# 4. Balancer метрики
BALANCER_CONNS=$(query_metric "balancer_active_connections")

# 5. Redis и PostgreSQL (если есть экспортеры)
REDIS_MEM=$(query_metric "redis_memory_used_bytes{job='redis'}" || echo "null")
REDIS_MEM=$(query_metric "redis_memory_used_bytes")
REDIS_CONNS=$(query_metric "redis_connected_clients")
PG_CONNS=$(query_metric "pg_stat_database_numbackends{datname='eventhorizon'}" || echo "null")
PG_CONNS=$(query_metric "pg_stat_database_numbackends{datname='eventhorizon'}")

# --- Сохраняем в JSON ---
cat > "$OUTPUT_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "metrics": {
    "game": {
      "submits_total": $GAME_SUBMITS,
      "goroutines": $GO_GOROUTINES,
      "heap_alloc_bytes": $GO_HEAP,
      "score_count": $GAME_SCORE_COUNT
    },
    "nats": {
      "connections": $NATS_CONNS,
      "in_msgs_total": $NATS_IN_MSGS,
      "out_msgs_total": $NATS_OUT_MSGS
    },
    "gateway": {
      "instance_1_requests": $GW1,
      "instance_2_requests": $GW2,
      "instance_3_requests": $GW3
    },
    "balancer": {
        "active_connections": $BALANCER_CONNS
    },
    "storage": {
        "redis_memory_bytes": $REDIS_MEM,
        "redis_connections": $REDIS_CONNS,
        "postgres_connections": $PG_CONNS
    },
}
EOF

echo "✅ Метрики сохранены в: $OUTPUT_FILE"
echo ""
echo "📊 Содержимое:"
cat "$OUTPUT_FILE" | jq '.'