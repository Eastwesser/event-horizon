# EventHorizon — День мониторинга и метрик
**Дата:** 16 июня 2026  
**Автор:** Капитан и Хохо  
**Статус:** ✅ Завершено

---

## 🎯 Цель дня
Развернуть полноценный стек мониторинга для EventHorizon:
- Prometheus (сбор метрик)
- Grafana (визуализация)
- Jaeger (трейсинг — начат, но отложен)
- Алерты в Telegram
- Бизнес-метрики (RPS, latency, ошибки)

---

## 📦 Итоговый стек

| Компонент | Версия | Порт | Статус |
|-----------|--------|------|--------|
| Prometheus | latest | 9090 | ✅ |
| Grafana | latest | 3000 | ✅ |
| Jaeger | all-in-one | 16686 | 🟡 (UI готов, трейсы позже) |
| NATS | 2.10 | 4222 / 8222 | ✅ |
| Auth | go1.25.7 | 50051 / 9091 | ✅ |
| Billing | go1.25.7 | 50053 / 9093 | ✅ |
| Game | go1.25.7 | 50052 / 9092 | ✅ |
| Gateway | go1.25.7 | 8080 / 9095 | ✅ |
| Leaderboard | go1.25.7 | 50054 / 9094 | ✅ |

---

## 🔧 Порты мониторинга

| Сервис | Метрики (Prometheus) | Назначение |
|--------|----------------------|------------|
| Auth | `http://localhost:9091/metrics` | gRPC + метрики |
| Game | `http://localhost:9092/metrics` | Игровая логика |
| Billing | `http://localhost:9093/metrics` | Награды и транзакции |
| Leaderboard | `http://localhost:9094/metrics` | Топ игроков |
| Gateway | `http://localhost:9095/metrics` | HTTP + WebSocket |
| NATS | `http://localhost:8222/metrics` | Брокер сообщений |
| Prometheus | `http://localhost:9090` | UI + API |
| Grafana | `http://localhost:3000` | Дашборды |
| Jaeger UI | `http://localhost:16686` | Трейсинг |

---

## 📊 Бизнес-метрики (кастомные)

### Gateway
| Метрика | Тип | Лейблы | Назначение |
|---------|-----|--------|------------|
| `gateway_requests_total` | Counter | `method`, `path`, `status` | RPS по эндпоинтам |
| `gateway_request_duration_seconds` | Histogram | `method`, `path` | Latency (P50/P95/P99) |

### Game
| Метрика | Тип | Лейблы | Назначение |
|---------|-----|--------|------------|
| `game_submits_total` | Counter | `game_id`, `status` | Количество игр |
| `game_score_histogram` | Histogram | `game_id` | Распределение очков |

---

## 🔍 Тестирование метрик (curl-команды)

### 1. Проверить, что метрики доступны
```bash
# Auth
curl -s http://localhost:9091/metrics | head -10

# Game
curl -s http://localhost:9092/metrics | head -10

# Billing
curl -s http://localhost:9093/metrics | head -10

# Leaderboard
curl -s http://localhost:9094/metrics | head -10

# Gateway
curl -s http://localhost:9095/metrics | head -10

# NATS
curl -s http://localhost:8222/metrics | head -10
2. Проверить кастомные метрики

bash
# Gateway RPS
curl -s http://localhost:9095/metrics | grep gateway_requests_total

# Game submits
curl -s http://localhost:9092/metrics | grep game_submits_total

# Game scores
curl -s http://localhost:9092/metrics | grep game_score_histogram
3. Проверить Prometheus

bash
# Статус всех сервисов (up/down)
curl -s http://localhost:9090/api/v1/query?query=up | jq '.data.result'

# Gateway requests total
curl -s "http://localhost:9090/api/v1/query?query=gateway_requests_total" | jq .

# Game submits total
curl -s "http://localhost:9090/api/v1/query?query=game_submits_total" | jq .

# RPS по эндпоинтам
curl -s "http://localhost:9090/api/v1/query?query=rate(gateway_requests_total[$__rate_interval])" | jq .
📈 Дашборды в Grafana

Импортированные дашборды

ID	Название	Источник	Статус
153	Go Metrics	Prometheus	✅
1860	Node Exporter Full	Prometheus	✅
13707	NATS Server Dashboard	Prometheus	✅
-	EventHorizon Business Metrics	Prometheus	✅ (кастомный)
Кастомный дашборд: EventHorizon Business Metrics

JSON для импорта:

json
{
  "title": "EventHorizon Business Metrics",
  "tags": ["eventhorizon", "business"],
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "panels": [
    {
      "title": "RPS by endpoint",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(gateway_requests_total[$__rate_interval])) by (path)",
          "legendFormat": "{{path}}"
        }
      ]
    },
    {
      "title": "Game submissions",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(game_submits_total[$__rate_interval])) by (game_id, status)",
          "legendFormat": "{{game_id}} - {{status}}"
        }
      ]
    },
    {
      "title": "Score distribution (P95)",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(game_score_histogram_bucket[$__rate_interval])) by (le, game_id))",
          "legendFormat": "{{game_id}}"
        }
      ]
    },
    {
      "title": "P95 latency",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(gateway_request_duration_seconds_bucket[$__rate_interval])) by (le, method, path))",
          "legendFormat": "{{method}} {{path}}"
        }
      ]
    }
  ]
}
🚨 Алерты в Grafana

Настроенные алерты (все в папке EventHorizon)

Имя правила	Условие	Действие	Статус
Gateway Down - Alert	up{job="gateway"} == 0	Telegram	✅
Auth Service Down - Alert	up{job="auth"} == 0	Telegram	✅
Game Service Down - Alert	up{job="game"} == 0	Telegram	✅
Billing Service Down - Alert	up{job="billing"} == 0	Telegram	✅
Leaderboard Service Down - Alert	up{job="leaderboard"} == 0	Telegram	✅
Настройка алерта (шаблон)

Запрос: up{job="<service>"}
Условие: IS BELOW 1 FOR 1m
Pending period: 1m
Contact point: Telegram (бот)
🤖 Telegram бот для алертов

Настройка

Создать бота у @BotFather
Получить токен: 1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
Добавить бота в группу
Получить Chat ID: -1001234567890
В Grafana → Alerting → Contact points → добавить Telegram
Формат уведомления

text
{{ $labels.job }} service is DOWN!
Service {{ $labels.job }} has been down for more than 1 minute.
Check logs: tail -f /tmp/{{ $labels.job }}.log
🧪 Проверка работоспособности

1. Проверить, что все сервисы запущены

bash
make status
2. Проверить Gateway

bash
curl http://localhost:8080/health
# Должен вернуть {"status":"ok"}
3. Проверить Prometheus targets

bash
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[].labels.job'
# Должны быть: auth, billing, game, gateway, leaderboard, nats
4. Проверить, что метрики собираются

bash
curl -s http://localhost:9090/api/v1/query?query=up | jq '.data.result[].value[1]'
# Должны быть "1" для всех сервисов (кроме NATS — может быть 0)
5. Проверить Grafana

bash
curl -s http://localhost:3000/api/health
# Должен вернуть {"commit":"...","database":"ok","version":"..."}
📁 Ключевые файлы

Файл	Назначение
deployments/docker-compose.cluster.yml	Docker Compose с Prometheus, Grafana, Jaeger
deployments/prometheus/prometheus.yml	Конфиг Prometheus (таргеты)
deployments/grafana/	Дашборды и datasources (в плане)
services/*/cmd/main.go	Метрики + Jaeger (в плане)
🚀 Как перезапустить всё вместе

bash
cd /home/denismatveev/event_horizon
./restart.sh
Или пошагово:

bash
make stop-services
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d
make all
make status
📝 Что отложено (Jaeger / трейсинг)

Задача	Статус	План
Добавить OpenTelemetry в сервисы	🟡 Начато	17 июня
Настроить экспорт в Jaeger	🟡 Начато	17 июня
Интеграция с Grafana (Tempo)	⬜️	Позже
🧠 Итог

Компонент	Статус	Комментарий
Prometheus	✅	Собирает метрики со всех сервисов
Grafana	✅	4 дашборда (Go Metrics, Node Exporter, NATS, Business)
Алерты	✅	5 правил, уведомления в Telegram
NATS дашборд	✅	ID 13707
Бизнес-метрики	✅	RPS, latency, игры, очки
Jaeger UI	✅	Запущен, трейсы позже
EventHorizon теперь виден и слышен! 🔥

text
Дата завершения: 16 июня 2026, 02:00
Статус: ✅ MVP мониторинга готов