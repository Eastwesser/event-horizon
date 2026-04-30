# EventHorizon 🎮

Игровая платформа с микросервисной архитектурой, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

## Архитектура (30.04.2026)
```text
┌─────────┐ HTTP        ┌─────────┐ gRPC      ┌─────────┐
│ Client  │ ──────────► │ Gateway │ ────────► │ Auth    │
└─────────┘             └─────────┘           └─────────┘
                            │                       │
                            │ NATS                  │ PostgreSQL
                            ▼                       ▼
                        ┌─────────┐             ┌─────────┐
                        │ NATS    │             │ DB      │
                        │JetStream│             │ :5460   │
                        └─────────┘             └─────────┘
                            │
                            │ Subscribe
                            ▼
                    ┌─────────────────┐
                    │   Leaderboard   │
                    │ Redis Sorted Set│
                    └─────────────────┘
```

## Сервисы
```text
| Сервис      | Порт (gRPC) | PostgreSQL | Redis |
|-------------|-------------|------------|-------|
| Auth        | 50051       | 5460       | 6379  |
| Game        | 50052       | 5461       | 6380  |
| Billing     | 50053       | 5462       | 6381  |
| Leaderboard | 50054       | 5463       | 6382  |
| Gateway     | 8080 (HTTP) | -          | -     |
```

## Быстрый старт

### 1. Поднять инфраструктуру

```bash
make up
```

### 2. Запустить сервисы

```bash
# Терминал 1 - Auth
cd services/auth && go run cmd/main.go

# Терминал 2 - Gateway
cd services/gateway && go run cmd/main.go

# Терминал 3 - NATS subscriber (опционально)
nats sub "event.>" --server localhost:4222
```

### 3. Тестирование

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```

## Команды Make

```bash
make up          # Поднять инфраструктуру (Docker)
make down        # Остановить инфраструктуру
make ps          # Статус контейнеров
make logs        # Логи инфраструктуры
make clean       # Остановить и удалить volumes
```

## Порты (host)

```text
Сервис	                    Порт	    Назначение

PostgreSQL (auth)	        5460	    Основная БД
PostgreSQL (game)	        5461	    Игровая БД
PostgreSQL (billing)	    5462	    Платёжная БД
PostgreSQL (leaderboard)	5463	    БД для топа

Redis	                    6379-6382	Кеши/сессии

NATS	                    4222	    Событийная шина
NATS monitoring	            8222	    Метрики

Gateway	8080	HTTP API
```

## Документация

- История и планы     [event_horizon/confluence/history]
- Технический долг    [event_horizon/confluence/tech_debt]
- FAQ                 [event_horizon/confluence/faq] 

## Будущие заметки:
- Graceful shutdown для всех сервисов
- NATS кластер из 3 нод
- Prometheus + Grafana мониторинг
- CQRS для leaderboard (улучшение)
- Envoy как API gateway (опционально)

## Статус
```text
✅ Auth service (JWT, регистрация, логин)
✅ Gateway (HTTP → gRPC прокси)
✅ NATS JetStream (события публикуются)
⏳ Game service (в разработке)
⏳ Leaderboard (в разработке)
⏳ Billing service (в плане)
```
