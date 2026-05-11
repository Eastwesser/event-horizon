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
