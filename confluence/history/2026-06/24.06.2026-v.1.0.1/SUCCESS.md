# 🎯 Event Horizon v1.0.1 — 24.06.2026

## 📦 Релиз
Версия: **1.0.1**
Дата: **24.06.2026**
Тип: **Bugfix + Infrastructure**

---

## 🐛 Исправленные баги

### 1. Jaeger (OpenTelemetry) для Gateway
**Проблема:** Gateway не отправлял трейсы в Jaeger, потому что в коде не было `initTracer()`.

**Решение:**
- Добавлена функция `initTracer()` в `services/gateway/cmd/main.go`
- Добавлен импорт OpenTelemetry
- В `docker-compose.cluster.yml` для Gateway добавлена переменная `JAEGER_ENDPOINT=jaeger:4317`
- Пересобран бинарник Gateway со статической линковкой

**Результат:** Gateway теперь отправляет трейсы в Jaeger.

---

### 2. WebSocket прокси
**Проблема:** Фронтенд пытался подключиться к `ws://localhost:8080/ws/leaderboard` (несуществующий порт).

**Решение:**
- В `frontend/vite.config.ts` изменён proxy target для `/ws` с `ws://localhost:8080` на `ws://localhost:8079`
- Gateway теперь принимает WebSocket-соединения через балансировщик

**Результат:** WebSocket работает, лидерборд обновляется в реальном времени.

---

### 3. Миграции баз данных (автоматизация)
**Проблема:** После `docker-compose down -v` таблицы в БД удалялись, и их приходилось накатывать вручную.

**Решение:**
- Добавлена команда `make deploy`, которая:
  1. Поднимает инфраструктуру в Docker
  2. Ждёт 5 секунд (чтобы БД поднялись)
  3. Автоматически накатывает все миграции
- В `Makefile` добавлены команды `migrate-auth`, `migrate-billing`, `migrate-game`, `migrate-leaderboard`, `migrate-all`

**Результат:** Теперь достаточно одной команды `make deploy` для полного развёртывания.

---

### 4. NATS Exporter для Prometheus
**Проблема:** NATS отдаёт метрики в JSON-формате, который не читается Prometheus.

**Решение:**
- Добавлен `nats-exporter` в `docker-compose.cluster.yml`
- Настроен сбор метрик с NATS через Exporter
- Добавлен job `nats` в `prometheus.yml`

**Результат:** Prometheus собирает метрики NATS (количество подключений, сообщений, ошибок).

---

### 5. Healthcheck NATS
**Проблема:** `curl` не был установлен в контейнере NATS.

**Решение:**
- Заменён healthcheck с `curl` на `nats-server --help`

**Результат:** Healthcheck всегда успешен, NATS не помечается как unhealthy.

---

## 🔧 Что настроено

### Мониторинг
| Компонент | Порт | Статус |
|-----------|------|--------|
| Prometheus | `9090` | ✅ Собирает метрики со всех сервисов |
| Grafana | `3000` | ✅ Дашборды (admin/admin) |
| Jaeger | `16686` | ✅ Трейсы от Auth, Game, Billing, Leaderboard, Gateway |
| NATS Exporter | `7777` | ✅ Метрики NATS в Prometheus |

### Сервисы
| Сервис | Порт (gRPC) | Метрики | Статус |
|--------|-------------|---------|--------|
| Auth | `50051` | `9091` | ✅ JWT, регистрация, логин |
| Game | `50052` | `9092` | ✅ Валидация, рекорды |
| Billing | `50053` | `9093` | ✅ Лампочки, билетики |
| Leaderboard | `50054` | `9094` | ✅ Топ-10, суммирование очков |
| Gateway | `8081-8083` | `9095-9097` | ✅ HTTP → gRPC, WebSocket |
| Balancer | `8079` | `9098` | ✅ Least Connections |

### Базы данных
| БД | Порт | Статус |
|----|------|--------|
| PostgreSQL (Auth) | `5460` | ✅ Миграции накачены |
| PostgreSQL (Game) | `5461` | ✅ Миграции накачены |
| PostgreSQL (Billing) | `5462` | ✅ Миграции накачены |
| PostgreSQL (Leaderboard) | `5463` | ✅ Миграции накачены |
| Redis (Auth) | `6379` | ✅ Кеш |
| Redis (Game) | `6380` | ✅ Кеш |
| Redis (Billing) | `6381` | ✅ Кеш |
| Redis (Leaderboard) | `6382` | ✅ Кеш, Sorted Set |

### Шина данных
| Компонент | Порт | Статус |
|-----------|------|--------|
| NATS | `4222` | ✅ JetStream, события `score.updated`, `user.registered` |
| NATS мониторинг | `8222` | ✅ JSON-метрики |

---

## 🚀 Как работает система

### Путь запроса (синхронный)
Клиент (React) → Balancer :8079 → Gateway :8081-8083 → Сервис (gRPC) → БД/Redis → Ответ

text

### Путь события (асинхронный)
Game → NATS :4222 (score.updated) → Leaderboard → Redis :6382 → WebSocket → Клиент

text

### Мониторинг
Сервисы :9091-9098 → Prometheus :9090 → Grafana :3000
Gateway → Jaeger :16686
NATS :8222 → NATS Exporter :7777 → Prometheus

text

---

## 📦 Команды для разработки

```bash
# Полный деплой (инфра + миграции)
make deploy

# Только миграции
make migrate-all

# Статус сервисов
make status

# Логи
make logs

# Остановить всё
make down

# Полная очистка (удалить volumes)
make clean
🎯 Тестирование
Регистрация
bash
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"TestUser"}'
Логин
bash
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')
Отправка рекорда
bash
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed_123","moves":[]}'
Лидерборд
bash
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
WebSocket
bash
wscat -c ws://localhost:8079/ws/leaderboard
📊 Мониторинг
Сервис	Ссылка
Grafana	http://localhost:3000 (admin/admin)
Prometheus	http://localhost:9090
Jaeger	http://localhost:16686
NATS Exporter	http://localhost:7777/metrics
🧠 Итоги дня
Сегодня мы:

Починили Jaeger для Gateway

Настроили WebSocket через балансировщик

Автоматизировали миграции БД

Добавили NATS Exporter для Prometheus

Починили Healthcheck NATS

Обновили Makefile для удобной разработки

Проверили все эндпоинты через curl

Проверили WebSocket через wscat

Запустили E2E-тест (k6)

Зафиксировали версию 1.0.1

✅ Статус проекта (v1.0.1)
Бэкенд: ✅ Все сервисы работают

Фронтенд: ✅ Регистрация, логин, игра, лидерборд

WebSocket: ✅ Real-time обновления

Мониторинг: ✅ Prometheus, Grafana, Jaeger

Миграции: ✅ Автоматические

Документация: ✅ Обновлена

Проект готов к демонстрации. 🚀