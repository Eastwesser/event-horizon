# EventHorizon 🎮

Игровая платформа с микросервисной архитектурой, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

## Архитектура (актуальная на 11.05.2026)

```text
┌─────────┐    HTTP     ┌─────────┐    gRPC    ┌─────────┐
│  curl   │ ──────────► │ Gateway │ ──────────►│  Game   │
└─────────┘             └─────────┘            └─────────┘
                             │                      │
                             │ WebSocket            │ NATS publish
                             ▼                      ▼
                        ┌─────────┐            ┌─────────┐
                        │  React  │            │  NATS   │
                        │ Client  │            │JetStream│
                        └─────────┘            └────┬────┘
                                                     │
                                    ┌────────────────┼────────────────┐
                                    │                │                │
                                    ▼                ▼                ▼
                              ┌───────────┐      ┌─────────┐      ┌─────────┐
                              │Leaderboard│      │ Billing │      │  Auth   │
                              │ :50054    │      │ :50053  │      │ :50051  │
                              └────┬──────┘      └────┬────┘      └────┬────┘
                                   │                  │                │
                                   ▼                  ▼                ▼
                              ┌─────────┐      ┌──────────┐      ┌───────────┐
                              │  Redis  │      │PostgreSQL│      │PostgreSQL │
                              │ :6382   │      │ :5462    │      │ :5460     │
                              └─────────┘      └──────────┘      └───────────┘
```

## Сервисы (актуально 11.05.2026)

Сервис	    Порт(gRPC)	  PostgreSQL	Redis	  Статус
Auth	      50051	        5460	      6379	  ✅
Game	      50052	        5461	      6380	  ✅
Billing	    50053	        5462	      6381	  ✅
Leaderboard	50054	        5463	      6382	  ✅
Gateway	    8080(HTTP)	  -	          -	      ✅

## Быстрый старт

1. Запустить всё одной командой

```bash
cd ~/event_horizon
make all
```

2. Проверить, что всё работает

```bash
# Статус Docker контейнеров
make ps

# Проверить порты
ss -tlnp | grep -E "50051|50052|50053|50054|8080"
```

3. Тестирование API

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Отправить рекорд
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<UUID>",
    "game_id": "hexagon",
    "level": 3,
    "seed": "test_seed",
    "moves": []
  }'
```

## Команды Make

```bash
make all           # Запустить всё (Docker + сервисы)
make stop          # Остановить все сервисы
make restart       # Перезапустить всё
make up            # Поднять только Docker контейнеры
make down          # Остановить Docker контейнеры
make ps            # Статус контейнеров
make logs          # Логи контейнеров
make clean         # Остановить и удалить volumes

# Логи сервисов
tail -f /tmp/auth.log
tail -f /tmp/game.log
tail -f /tmp/billing.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log

# Порты (host)

Сервис	                Порт	    Назначение
PostgreSQL(auth)	      5460	    Основная БД
PostgreSQL(game)	      5461	    Игровая БД
PostgreSQL(billing)	    5462	    Платёжная (внутриигровая) БД
PostgreSQL(leaderboard)	5463	    БД для топа
Redis	                  6379-6382	Кеши/сессии
NATS	                  4222	    Событийная шина
NATS monitoring	        8222	    Метрики
Gateway	                8080	    HTTP API + WebSocket
```

## Статус (11.05.2026)

Сервис	          Статус
Auth	            ✅ JWT, регистрация, логин
Gateway	          ✅ HTTP → gRPC, WebSocket, NATS
Game	            ✅ Честная валидация, подсчёт очков
Billing	          ✅ Лампочки, билетики, транзакции
Leaderboard	      ✅ Redis Sorted Set, NATS
WebSocket	        ✅ Real-time broadcast
NATS JetStream	  ✅ Событийная шина
Graceful shutdown	✅ Все сервисы

## Документация

- История и планы       [event_horizon/confluence/history]
- Технический долг      [event_horizon/confluence/tech_debt]
- FAQ                   [event_horizon/confluence/faq]
- OpenAPI спецификация  [event_horizon/docs/openapi.yaml] 

## Будущие заметки:

- Graceful shutdown для всех сервисов
- NATS кластер из 3 нод
- Prometheus + Grafana мониторинг
- CQRS для leaderboard (улучшение)
- Envoy как API gateway (опционально)

## Техдолг (основное):

- NATS кластер из 3 нод
- Prometheus + Grafana + Jaeger мониторинг
- Load balancer (nginx + самописный)
- Envoy как API gateway
- CQRS для leaderboard (улучшение)

## Статус на 30.05.2026 (Что теперь работает)

```text
  ✅ Авторизация (JWT)
  ✅ Игра (drag-n-drop, слияние стопок)
  ✅ Leaderboard (суммирование очков, email)
  ✅ Billing (лампочки, билетики)
  ✅ WebSocket (real-time обновления)
  ✅ Все миграции
  ✅ Баланс на фронтенде
```

# EventHorizon 🎮

Игровая платформа с микросервисной архитектурой, 
real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

## Архитектура (актуальная на 30.05.2026)

```text
┌─────────┐    HTTP     ┌─────────┐    gRPC    ┌─────────┐
│  React  │ ──────────► │ Gateway │ ──────────►│  Game   │
└─────────┘             └─────────┘            └─────────┘
                             │                      │
                             │ WebSocket            │ NATS publish
                             ▼                      ▼
                        ┌─────────┐            ┌─────────┐
                        │ Client  │            │  NATS   │
                        │ (wscat) │            │JetStream│
                        └─────────┘            └────┬────┘
                                                     │
                                    ┌────────────────┼────────────────┐
                                    │                │                │
                                    ▼                ▼                ▼
                              ┌───────────┐      ┌─────────┐      ┌─────────┐
                              │Leaderboard│      │ Billing │      │  Auth   │
                              │ :50054    │      │ :50053  │      │ :50051  │
                              └────┬──────┘      └────┬────┘      └────┬────┘
                                   │                  │                │
                                   ▼                  ▼                ▼
                              ┌─────────┐      ┌──────────┐      ┌───────────┐
                              │  Redis  │      │PostgreSQL│      │PostgreSQL │
                              │ :6382   │      │ :5462    │      │ :5460     │
                              └─────────┘      └──────────┘      └───────────┘
```

## Сервисы (актуально 30.05.2026)

```text
Сервис	        Порт (gRPC)	        PostgreSQL	    Redis	    Статус

Auth	          50051	              5460	          6379	    ✅
Game	          50052	              5461	          6380	    ✅
Billing	        50053	              5462	          6381	    ✅
Leaderboard	    50054	              5463	          6382	    ✅
Gateway	        8080 (HTTP)	        -	              -	        ✅
```

## Быстрый старт

1. Запустить всё одной командой

```bash
cd ~/event_horizon
make all
```

2. Проверить, что всё работает

```bash
# Статус Docker контейнеров
make ps

# Проверить порты
ss -tlnp | grep -E "50051|50052|50053|50054|8080"
```

3. Тестирование API

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Получить баланс
curl -X GET "http://localhost:8080/api/billing/balance/all" \
  -H "Authorization: Bearer <токен>"

# Отправить рекорд
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<UUID>",
    "game_id": "hexagon",
    "level": 3,
    "score": 100,
    "seed": "test_seed",
    "moves": []
  }'
```

# Команды Make

```bash
make all           # Запустить всё (Docker + сервисы)
make stop          # Остановить все сервисы
make restart       # Перезапустить всё
make up            # Поднять только Docker контейнеры
make down          # Остановить Docker контейнеры
make ps            # Статус контейнеров
make logs          # Логи контейнеров
make clean         # Остановить и удалить volumes

# Миграции баз данных
make migrate-all   # Применить все миграции

# Логи сервисов
tail -f /tmp/auth.log
tail -f /tmp/game.log
tail -f /tmp/billing.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log
```

# Порты (host)

```text
Сервис	                  Порт	    Назначение

PostgreSQL (auth)	        5460	    Основная БД
PostgreSQL (game)	        5461	    Игровая БД
PostgreSQL (billing)	    5462	    Платёжная (внутриигровая) БД
PostgreSQL (leaderboard)	5463	    БД для топа
Redis	                    6379-6382	Кеши/сессии
NATS	                    4222	    Событийная шина
NATS monitoring	          8222	    Метрики
Gateway	                  8080	    HTTP API + WebSocket
```

# Статус (30.05.2026)

```text
Компонент	          Статус

Auth	              ✅ JWT, регистрация, логин
Gateway	            ✅ HTTP → gRPC, WebSocket, NATS
Game	              ✅ Честная валидация, подсчёт очков
Billing	            ✅ Лампочки, билетики, транзакции
Leaderboard	        ✅ Redis Sorted Set, NATS, суммирование очков
WebSocket	          ✅ Real-time broadcast
NATS JetStream	    ✅ Событийная шина
Graceful shutdown	  ✅ Все сервисы
Goose миграции	    ✅ Auth, Game, Billing, Leaderboard
Баланс на фронтенде	✅ Отображается
```

# Что работает (30.05.2026)

```text
✅ Авторизация (JWT)
✅ Игра (drag-n-drop, слияние стопок)
✅ Leaderboard (суммирование очков, email)
✅ Billing (лампочки, билетики)
✅ WebSocket (real-time обновления)
✅ Все миграции
✅ Баланс на фронтенде
```
# Планы на следующий спринт

- Выбор уровня сложности (1-20, множитель очков)
- Анимация пуф-эффекта при очистке стопки
- Prometheus + Grafana + Jaeger мониторинг
- Блинопекарня (магазин за лампочки)
- NATS кластер из 3 нод
- Load balancer (nginx)
- tech debt

## 🖥️ Мониторинг (добавлено 16.06.2026)

### Стек
- **Prometheus** — сбор метрик (порт 9090)
- **Grafana** — визуализация (порт 3000, admin/admin)
- **Jaeger** — трейсинг (порт 16686, в разработке)
- **Alertmanager** — уведомления в Telegram

### Метрики
Каждый сервис отдаёт метрики на своём порту:

| Сервис | Метрики | Порт |
|--------|---------|------|
| Auth | `http://localhost:9091/metrics` | 9091 |
| Game | `http://localhost:9092/metrics` | 9092 |
| Billing | `http://localhost:9093/metrics` | 9093 |
| Leaderboard | `http://localhost:9094/metrics` | 9094 |
| Gateway | `http://localhost:9095/metrics` | 9095 |
| NATS | `http://localhost:8222/metrics` | 8222 |

### Дашборды Grafana
| ID | Название |
|----|----------|
| 153 | Go Metrics |
| 1860 | Node Exporter Full |
| 13707 | NATS Server Dashboard |
| - | **EventHorizon Business Metrics** (кастомный) |

### Алерты
Настроены 5 алертов в Telegram:
- Gateway Down
- Auth Down
- Game Down
- Billing Down
- Leaderboard Down

### Бизнес-метрики
- `gateway_requests_total` — RPS по эндпоинтам
- `gateway_request_duration_seconds` — Latency (P50/P95/P99)
- `game_submits_total` — Количество игр по типам
- `game_score_histogram` — Распределение очков

### Проверка
```bash
# Проверить все метрики
curl -s http://localhost:9095/metrics | grep gateway_

# Проверить Prometheus
curl -s http://localhost:9090/api/v1/query?query=up | jq '.data.result'