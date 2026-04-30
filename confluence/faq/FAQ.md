# FAQ

## Project Tree (start)

```text
eventhorizon/
├── services/
│   ├── auth/           # авторизация (JWT, OAuth, "забыли пароль")
│   │   ├── cmd/main.go
│   │   ├── internal/...
│   │   ├── proto/      # gRPC .proto файлы
│   │   └── Dockerfile
│   ├── game/           # игровая логика (шестиугольники, флаппи)
│   │   ├── cmd/main.go
│   │   ├── internal/...
│   │   ├── proto/      # gRPC .proto файлы
│   │   └── Dockerfile
│   ├── leaderboard/    # highscore, топ-10, обновления через NATS
│   │   ├── cmd/main.go
│   │   ├── internal/...
│   │   ├── proto/
│   │   └── Dockerfile
│   ├── billing/        # лампочки, билетики, этажи
│   │   ├── cmd/main.go
│   │   ├── internal/...
│   │   ├── proto/
│   │   └── Dockerfile
│   └── gateway/        # API Gateway (Gin, WebSocket, маршрутизация)
│       ├── cmd/main.go
│       ├── internal/...
│       └── Dockerfile
├── monitoring/         # Prometheus + Grafana + Loki
│   ├── prometheus.yml
│   ├── grafana/
│   └── docker-compose.monitoring.yml
├── deployments/        # Kubernetes или docker-compose для кластера
│   ├── docker-compose.cluster.yml
│   └── configs/
├── scripts/
│   ├── loadtest.js
│   └── seed.sql
├── frontend/           # Angular
├── Makefile            # make up-cluster, make down-cluster, make test
└── README.md
```

# RUN:

```bash
docker-compose -f deployments/docker-compose.cluster.yml up -d
```

# CHECK HEALTH:

```bash
# Postgres
docker exec event-horizon-postgres pg_isready -U eventhorizon
# Ожидаем: /var/run/postgresql:5432 - accepting connections

# Проверить Redis
docker exec event-horizon-redis redis-cli ping
# → PONG

# NATS
docker exec event-horizon-nats nats-server --version
# Ожидаем: nats-server version 2.10.x

# Проверить, что NATS включил JetStream
docker logs event-horizon-nats | head -20
# Должно быть что-то про "JetStream" в логах
```

# CHECK SERVICES:

```bash
# Остановить всё
docker-compose -f deployments/docker-compose.cluster.yml down

# Запустить заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверить
docker ps
```

# Что мы имеем в итоге
```text
Сервис	    Порт (host)	    Статус	    Особенности

PostgreSQL	5460	        healthy	    Пользователь eventhorizon, БД eventhorizon
Redis	    6379	        healthy	    Стандартный
NATS	    4222, 8222	    healthy	    JetStream включён
```

# FIND AND KILL THE PORT IN USE:
```bash
# Check if in use:
sudo lsof -i :5432

# KILL PORT (WARNING!!!)
sudo kill -9 1234
# OR
sudo fuser -k 5432/tcp 
```

# PRUNING OPTION (NUKE):
```bash
# Удалить остановленные контейнеры
docker container prune -f

# Удалить неиспользуемые образы
docker image prune -f

# Удалить всё, что не используется (аккуратно!)
docker system prune -f
```

# 1. Базы данных: "сразу разделить" — правильное решение

## у каждого микросервиса своя PostgreSQL + свой Redis.

```text
PostgreSQL Auth    :5460 + Redis Auth    :6379 (сессии, кеш)
PostgreSQL Game    :5461 + Redis Game    :6380 (игровые состояния)
PostgreSQL Billing :5462 + Redis Billing :6381 (идемпотентность)
PostgreSQL Leader  :5463 + Redis Leader  :6382 (топ-10)
```

## Почему это правильно для цели (10k RPS):

Аспект	                Общая БД	                        Раздельные БД

Масштабирование	        ❌ Один инстанс узкое место	        ✅ Каждый сервис скейлится отдельно
Отказоустойчивость	    ❌ Падение БД — всё падает	        ✅ Auth упал — игра продолжается
Схемы	                ❌ Трудно менять, все зависят	    ✅ Меняешь схему auth независимо
DevOps	                ✅ Одна БД проще	                    ❌ 4+ БД, 4+ Redis

"Мы выбрали Database per Service, потому что даже в pet-проекте важно эмулировать реальную архитектуру. 
Данные консистентны через события (NATS), не через shared database."

# ПРОВЕРКА РАБОТЫ БАЗ ДАННЫХ

## Auth Postgres (порт 5460)
docker exec event-horizon-postgres pg_isready -U eventhorizon

## Game Postgres (порт 5461)
docker exec event-horizon-postgres-game pg_isready -U eventhorizon

## Billing Postgres (порт 5462)
docker exec event-horizon-postgres-billing pg_isready -U eventhorizon

## Leaderboard Postgres (порт 5463)
docker exec event-horizon-postgres-leaderboard pg_isready -U eventhorizon

## Redis'ы (все должны ответить PONG)
docker exec event-horizon-redis redis-cli ping
docker exec event-horizon-redis-game redis-cli ping
docker exec event-horizon-redis-billing redis-cli ping
docker exec event-horizon-redis-leaderboard redis-cli ping

## NATS
docker exec event-horizon-nats nats-server --version

# NATS:

MAIN LIBRARY: nats-io/netscli/nats@latest

To run NATS:
```bash
nats sub "event.>" -- server localhost:4222
```

For more info check (
    event_horizon/confluence/history/01.april/30.04.2026/NATS_SUCCESS.md
    event_horizon/confluence/history/01.april/30.04.2026/NATS_TEST.md
)