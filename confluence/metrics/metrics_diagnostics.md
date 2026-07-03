почти все значения — null. Это значит, что либо:

Метрики не регистрируются в коде.

Prometheus не scrap'ит их (не настроен job).

Имена метрик не совпадают с тем, что я написал в запросах.

Давай разберемся по порядку.

🔍 Шаг 1: Проверяем, что Prometheus видит все таргеты
Выполни эту команду, чтобы увидеть список активных таргетов:

bash
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health, scrapePool: .scrapePool}'
Что должно быть:

auth, billing, game, leaderboard, gateway, balancer, nats — все должны быть "health": "up".

🔍 Шаг 2: Проверяем доступность метрик напрямую
Проверим, отдает ли Game сервис метрики:

bash
curl -s http://localhost:9092/metrics | grep -E "go_goroutines|gateway|game"
Что мы хотим увидеть:

go_goroutines{...} — есть.

game_submits_total{...} — должна быть, если мы её зарегистрировали.

gateway_requests_total{...} — должна быть в Gateway.

🔍 Шаг 3: Проверяем NATS Exporter
bash
curl -s http://localhost:7777/metrics | grep -E "gnatsd_varz_connections"
Если есть ответ — NATS Exporter работает.

🛠️ Что сломалось и как починить
1. Метрики в Go (Game, Auth, Billing, Leaderboard)
Проблема: В коде нет регистрации кастомных метрик (например, game_submits_total).

Решение: Добавить в services/game/internal/service/game_service.go:

go
var (
    gameSubmitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_submits_total",
            Help: "Total number of game submissions",
        },
        []string{"game_id", "status"},
    )
)

// В методе SubmitScore:
gameSubmitsTotal.WithLabelValues(req.GameID, "success").Inc()
2. Метрики в Gateway (RPS, Latency)
Проблема: Я вижу в services/gateway/cmd/main.go у тебя есть:

go
gatewayRequestsTotal := promauto.NewCounterVec(...)
gatewayRequestDuration := promauto.NewHistogramVec(...)
Они должны работать. Но в JSON они null — значит, либо на них нет данных, либо я неправильно написал запрос.

Проверь вручную:

bash
curl -s http://localhost:9095/metrics | grep gateway_requests_total
3. Метрики в Balancer
Проблема: В JSON balancer_active_connections — null. Значит, либо метрика не регистрируется, либо я ошибся в имени.

Проверь:

bash
curl -s http://localhost:9098/metrics | grep balancer
4. NATS
nats метрики — 6 соединений, in_msgs_per_second и out_msgs_per_second — 0. Это нормально, если нагрузка была маленькая, но мы запускали 50 пользователей на 60 секунд — должно быть больше 0.

Проверь NATS Exporter вручную:

bash
curl -s http://localhost:7777/metrics | grep gnatsd_varz_in_msgs
5. Redis и PostgreSQL
У тебя нет Redis Exporter и PostgreSQL Exporter в docker-compose.cluster.yml. Поэтому метрики null.

Решение: Добавить их, если нужны метрики БД.

✅ План действий (прямо сейчас)
Проверь, что Prometheus scrap'ит все таргеты (команда выше).

Проверь метрики вручную через curl для Gateway, Game, NATS.

Если метрики есть, но в JSON null — я перепишу запросы в скрипте, чтобы они точно совпадали с именами в твоем коде.

📝 Обновленный скрипт (с правильными запросами)
Я перепишу скрипт, используя динамический поиск метрик через Prometheus API, чтобы он сам находил правильные имена.

bash
#!/bin/bash

# Функция для поиска метрики по шаблону и возврата первого значения
query_metric_flexible() {
    local pattern=$1
    # Ищем все метрики, подходящие под шаблон
    local metric_names=$(curl -s "http://localhost:9090/api/v1/label/__name__/values" | jq -r ".data[] | select(contains(\"$pattern\"))" | head -1)
    
    if [[ -z "$metric_names" ]]; then
        echo "null"
        return
    fi
    
    # Берем первую найденную метрику и делаем запрос
    local query="rate(${metric_names}[1m])"
    local result=$(curl -s -G --data-urlencode "query=$query" "http://localhost:9090/api/v1/query" | jq -r '.data.result[0].value[1] // "null"')
    echo "$result"
}

# Использование:
# RPS=$(query_metric_flexible "gateway_requests")
# LATENCY=$(query_metric_flexible "gateway_request_duration_seconds")
# GOROUTINES=$(query_metric_flexible "go_goroutines")

---

### FIX ###

🔥 Что я вижу в твоих логах
Gateway метрики (RPS, latency) — НЕ РАБОТАЮТ. curl http://localhost:9095/metrics | grep gateway_requests_total ничего не вывел. 
Значит, Gateway не отдаёт метрики. 
Это потому что у тебя три инстанса Gateway, и они висят на портах 9095, 9096, 9097, но Prometheus почему-то не видит их метрики (или они не зарегистрированы).

Game метрики — РАБОТАЮТ! ✅

game_submits_total{game_id="hexagon",status="success"} 122 — есть!

go_goroutines 29 — есть!

game_score_histogram — есть!

NATS метрики — РАБОТАЮТ! ✅

gnatsd_varz_connections 6 — есть!

gnatsd_varz_in_msgs 2576 — есть!

Balancer метрики — НЕ РАБОТАЮТ. curl http://localhost:9098/metrics | grep balancer ничего не вывел.

Redis и PostgreSQL — НЕТ экспортеров, поэтому метрики null.

🛠️ Починим скрипт: исправим имена метрик и добавим новые
Я перепишу скрипт с правильными именами, которые ты увидел в логах.

Создай новый файл metrics_collector.sh (или перезапиши старый):

bash
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
PG_CONNS=$(query_metric "pg_stat_database_numbackends{datname='eventhorizon'}" || echo "null")

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
      "postgres_connections": $PG_CONNS
    }
  }
}
EOF

echo "✅ Метрики сохранены в: $OUTPUT_FILE"
echo ""
echo "📊 Содержимое:"
cat "$OUTPUT_FILE" | jq '.'
🚀 Запускаем новый скрипт
bash
chmod +x metrics_collector.sh
./metrics_collector.sh

Продолжение тут:
event_horizon/confluence/history/2026-06/29.06.2026