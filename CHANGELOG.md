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

# ---------------------------------------------------------------

📅 Релиз v1.0 — 23.06.2026

✅ Что сделано за спринт
1. Полноценный запуск в Docker

Все 5 микросервисов (auth, billing, game, leaderboard, gateway) работают в контейнерах

3 инстанса gateway + балансировщик (round-robin)

Вся инфраструктура в Docker Compose (PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana)

2. Решены проблемы с protobuf

Переход на статическую сборку бинарников (CGO_ENABLED=0, -extldflags=-static)

Образы собираются из готовых бинарников (FROM scratch) — генерация protobuf вынесена из Docker-сборки

Фикс nil pointer dereference в NATS, добавлены retry-механизмы

3. Сети и DNS

Все сервисы в единой Docker-сети event-horizon-net

gRPC-клиенты обращаются по именам контейнеров (auth:50051, billing:50053 и т.д.)

4. Метрики и мониторинг

Prometheus собирает метрики со всех сервисов (8 targets UP)

Добавлены метрики для balancer (порт 9098)

NATS мониторится через /varz (JSON)

5. Миграции

Накачены миграции для auth, billing, game, leaderboard через goose

6. Docker Hub

Все образы опубликованы: eastwesser/*:latest

Автоматический деплой через docker-compose pull && up -d

📊 Текущий статус (23.06.2026)
Компонент	Статус	Порт
Auth	✅	50051 (gRPC), 9091 (metrics)
Billing	✅	50053 (gRPC), 9093 (metrics)
Game	✅	50052 (gRPC), 9092 (metrics)
Leaderboard	✅	50054 (gRPC), 9094 (metrics)
Gateway-1	✅	8081 (HTTP), 9095 (metrics)
Gateway-2	✅	8082 (HTTP), 9096 (metrics)
Gateway-3	✅	8083 (HTTP), 9097 (metrics)
Balancer	✅	8079 (HTTP), 9098 (metrics)
PostgreSQL (4 шт)	✅	5460-5463
Redis (4 шт)	✅	6379-6382
NATS	✅	4222, 8222
Jaeger	✅	16686
Prometheus	✅	9090
Grafana	✅	3000

🚀 Как поднять сейчас
bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Запустить всё
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 3. Проверить health
curl http://localhost:8079/health

# 4. Накатить миграции (если БД пустые)
cd services/auth && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable" up
cd services/billing && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable" up
cd services/game && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5461/eventhorizon_game?sslmode=disable" up
cd services/leaderboard && goose -dir migrations postgres "postgres://eventhorizon:eventhorizon@localhost:5463/eventhorizon_leaderboard?sslmode=disable" up

🐛 Известные проблемы и решения
Проблема	Решение
Gateway падает с slice bounds out of range	Пересобрать статический бинарник, скопировать в FROM scratch
NATS nil pointer	Проверить nc != nil перед вызовом методов
Сервисы не видят друг друга по DNS	Добавить networks: - event-horizon-net для всех сервисов
Billing не подключается к NATS	Добавить retry-цикл при старте
NATS не отдаёт метрики в Prometheus-формате	Использовать /varz для мониторинга (JSON)

📚 Полезные команды
bash
# Статус всех контейнеров
docker-compose -f deployments/docker-compose.cluster.yml ps

# Логи конкретного сервиса
docker-compose -f deployments/docker-compose.cluster.yml logs gateway --tail=30

# Проверить метрики в Prometheus
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Пересобрать один сервис
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker push eastwesser/gateway:latest
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway

# Остановить всё
docker-compose -f deployments/docker-compose.cluster.yml down

🔮 Планы на следующий спринт
Настроить автообновление дашбордов в Grafana

Добавить бизнес-метрики (RPS, latency, ошибки)

Настроить алерты в Telegram

CI/CD через GitHub Actions (сборка и пуш в докерхаб)

Написать e2e-тесты для API

Добавить логирование в Elasticsearch

Настроить горизонтальное масштабирование gateway

💪 Команда
Backend: Денис Матвеев (Golang, gRPC, NATS, Docker)

Архитектура: Микросервисная, событийно-ориентированная

Деплой: Docker Compose (пока), планы на k3s

Версия: 1.0.0
Дата релиза: 23.06.2026
Следующий релиз: TBD

---

Проект Event Horizon (v1.0)
1. Цель проекта
Event Horizon — это pet-проект, который я делаю, чтобы:

Отработать архитектуру микросервисов на реальном проекте.

Разобраться в DevOps-подходе: CI/CD, оркестрация, мониторинг.

Подготовиться к собеседованиям, имея в портфолио живую систему, которую можно показать и объяснить.

Проект — игровой бэкенд с элементами:

Аутентификация (JWT)

Игровая логика (очки, рекорды)

Внутриигровая валюта (лампочки/билетики)

Таблица лидеров (топ-10 в реальном времени)

Уведомления (планируются)

Аналитика (планируется)

2. Архитектура (общая схема)
text
[React Client :5173]
    │ HTTP (JSON)
    ▼
[Balancer :8079] — самописный, Least Connections
    │ HTTP
    ▼
[Gateway 1-3 :8081-8083] — JWT, HTTP→gRPC
    │ gRPC
    ▼
┌──────────────┼──────────────┐
│              │              │
▼              ▼              ▼
Auth :5051     Game :5052     Billing :5053     Leaderboard :5054
│              │              │                  │
▼              ▼              ▼                  ▼
PG :5460       PG :5461       PG :5462          PG :5463 + Redis :6382
(users)        (scores)       (balances)        (leaderboard)
│              │              │                  │
└──────────────┼──────────────┴──────────────────┘
               │
               ▼
          [NATS :4222] — событийная шина (score.updated, user.registered)
               │
               ▼
    Leaderboard подписан → обновляет Redis → WebSocket → клиент
3. Компоненты и порты
Компонент	Протокол	Порт	Назначение
Balancer	HTTP	8079	Least Connections, самописный
Gateway 1	HTTP	8081	Входная точка, JWT, роутинг
Gateway 2	HTTP	8082	Входная точка, JWT, роутинг
Gateway 3	HTTP	8083	Входная точка, JWT, роутинг
Auth	gRPC	5051	Аутентификация, JWT
Game	gRPC	5052	Игровая логика, очки
Billing	gRPC	5053	Внутриигровая валюта
Leaderboard	gRPC	5054	Топ-10, WebSocket
NATS	TCP	4222	Событийная шина (JetStream)
NATS (мониторинг)	HTTP	8222	JSON-метрики (не для Prometheus)
PostgreSQL (Auth)	TCP	5460	Пользователи
PostgreSQL (Game)	TCP	5461	Рекорды, счета
PostgreSQL (Billing)	TCP	5462	Балансы, транзакции
PostgreSQL (Leaderboard)	TCP	5463	Топ-10
Redis (Auth)	TCP	6379	Кеш, JWT сессии
Redis (Game)	TCP	6380	Кеш
Redis (Billing)	TCP	6381	Кеш
Redis (Leaderboard)	TCP	6382	Кеш, топ-10
Prometheus	HTTP	9090	Метрики
Grafana	HTTP	3000	Дашборды
Jaeger	HTTP	16686	Трассировка
OTLP (HTTP)	HTTP	4318	OpenTelemetry
OTLP (gRPC)	gRPC	4317	OpenTelemetry
4. Взаимодействие сервисов
4.1. Клиент → Сервис (синхронный запрос)
text
1. Клиент → Balancer :8079 (HTTP)
2. Balancer выбирает Gateway с наименьшим количеством активных соединений
3. Balancer → Gateway :8081-8083 (HTTP)
4. Gateway проверяет JWT (если есть)
5. Gateway преобразует HTTP → gRPC
6. Gateway → нужный сервис (Auth/Game/Billing/Leaderboard) :5051-5054 (gRPC)
7. Сервис → БД (PostgreSQL) или Redis
8. Ответ → тем же маршрутом обратно
4.2. Сервисы → Сервисы (асинхронные события)
text
1. Game сохраняет рекорд в PostgreSQL
2. Game → NATS :4222 (публикует событие score.updated)
3. NATS → Leaderboard (подписка)
4. Leaderboard обновляет Redis :6382
5. Leaderboard → WebSocket → клиент (push-уведомление)
Сервисы НЕ общаются друг с другом напрямую по gRPC!
Только через NATS.

5. Балансировщик (самописный)
Алгоритм: Least Connections (наименьшее количество активных соединений).

go
func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
    var selected *Backend
    var minConns int32 = 2147483647

    for _, b := range lb.backends {
        conns := atomic.LoadInt32(&b.ActiveConns)
        if conns < minConns {
            minConns = conns
            selected = b
        }
    }
    return selected
}
Особенности:

Самописный, без Consul (пока).

Метрики на :9098.

6. Gateway
Задачи:

Принимает HTTP-запросы от клиента.

Проверяет JWT (если требуется).

Определяет, какой сервис нужен по пути:

/api/auth/* → Auth

/api/game/* → Game

/api/billing/* → Billing

/api/leaderboard/* → Leaderboard

Преобразует HTTP → gRPC.

Вызывает метод сервиса.

Rate Limiter:

Сейчас закомментирован (не мешает разработке).

Включить на проде, когда нагрузка станет > 100 RPS.

Настройки в internal/ratelimit/limiter.go:

AllowSubmit — 10 запросов/сек на пользователя

AllowLogin — 5 запросов/сек с IP

AllowWebSocket — 100 соединений/мин с IP

7. NATS (событийная шина)
Роль: передача событий между сервисами.

Порт: 4222 (клиентский), 8222 (HTTP-мониторинг).

JetStream: включён (сообщения хранятся персистентно).

Кластер:

NATS поддерживает кластеризацию "из коробки".

Для прода — минимум 3 ноды (RAFT).

Настройка: -cluster nats://0.0.0.0:6222 -routes nats://other-node:6222.

События (примеры):

user.registered — Auth → другие сервисы

score.updated — Game → Leaderboard, Billing

payment.completed — Payment → Billing, Analytics

8. Базы данных
8.1. PostgreSQL (каждому сервису своя)
Сервис	БД	Порт	Таблицы
Auth	users	5460	users, sessions
Game	scores	5461	scores, games
Billing	balances	5462	balances, transactions
Leaderboard	leaderboard	5463	top_players, history
Репликация:

План: 1 мастер на запись + 3 слейва на чтение.

Включить при нагрузке > 1000 RPS.

Сейчас — 1 БД на сервис.

8.2. Redis (кеш)
Сервис	Порт	TTL	Назначение
Auth	6379	15 мин	JWT, сессии
Game	6380	5 мин	Кеш игровых данных
Billing	6381	5 мин	Кеш балансов
Leaderboard	6382	1 мин	Топ-10 (обновляется часто)
Схема кеширования (Cache-Aside):

Сервис проверяет Redis.

Если есть — возвращает из кеша.

Если нет — идёт в PostgreSQL, записывает в Redis.

При обновлении — инвалидирует кеш.

9. Мониторинг
9.1. Метрики
Сервис	Метрики	Что собираем
Auth	:9091	JWT errors, registration, login
Game	:9092	Score updates, games played
Billing	:9093	Transactions, balances
Leaderboard	:9094	Top updates, WS connections
Gateway 1-3	:9095-9097	RPS, latency, HTTP errors
Balancer	:9098	Active connections
NATS	:8222	JSON (не используется в Prometheus)
9.2. Инструменты
Инструмент	Порт	Назначение
Prometheus	9090	Сбор метрик
Grafana	3000	Дашборды
Jaeger	16686	Трассировка
OTLP (HTTP)	4318	OpenTelemetry
OTLP (gRPC)	4317	OpenTelemetry
Примечание: NATS на :8222 отдаёт JSON, не Prometheus-формат.
Чтобы собирать метрики — нужен NATS Exporter или встроенный /metrics.

10. DevOps
10.1. Сейчас
Docker Compose — инфраструктура (PostgreSQL, Redis, NATS, Prometheus, Grafana, Jaeger).

Локальные бинарники — сервисы (Auth, Game, Billing, Leaderboard, Gateway).

Makefile — запуск (make start-services).

10.2. План (Ansible + k3s)
Ansible:

Автоматизация деплоя бинарников на VM.

Плейбуки для копирования бинарников и перезапуска systemd-сервисов.

k3s:

Лёгкий Kubernetes (можно поднять на одной VM).

Helm-чарты для микросервисов.

Инфраструктура — либо в Docker Compose, либо в Helm (StatefulSet).

CI/CD:

GitHub Actions → сборка Docker-образов.

Ansible → деплой на VM.

Нагрузочное тестирование (k6) после каждого деплоя.

10.3. Почему не 50 серверов?
Для старта — 1-3 сервера достаточно.
50 серверов — это уровень 1M+ пользователей в месяц.
Пока проект в разработке — Docker Compose + 1 VM — идеально.

11. Будущие сервисы
Сервис	Назначение	БД / Хранилище
Notification	Push-уведомления, email, SMS, Telegram	Firebase FCM, Redis
Analytics	DAU, MAU, Retention, события	ClickHouse (или PostgreSQL)
Payment	Реальные деньги, Boosty, вебхуки	PostgreSQL, Redis
Все новые сервисы:

Общаются через NATS.

Имеют свой Redis и БД.

Запускаются в 2 экземплярах (основной + резервный).

12. Ответы на частые вопросы
Вопрос	Ответ
Gateway обращается к NATS?	Нет, только к сервисам через gRPC.
Как сервисы общаются между собой?	Только через NATS (событийная шина).
Как информация возвращается из кеша/БД?	Тем же маршрутом обратно (сервис → Gateway → клиент).
Когда включать Rate Limiter?	На проде, при >100 RPS.
Нужно ли 50 серверов?	Нет, для старта 1-3, потом k3s.
Мастер + слейвы для БД?	Да, при >1000 RPS.
Ретраи с джиттером уже есть?	В Billing — да, в других — проверить.
Добавлять Consul?	Можно, но пока не критично.
Как реплицируется NATS?	Через кластеризацию (-cluster, -routes).

24.06.2026

Обновлены мэйкфайлы:

# Запустить всё с миграциями (ОДНА КОМАНДА!)
make deploy

# Проверить статус
make status

# Перезапустить
make restart

# Только миграции
make migrate-all

# Только инфраструктура
make up

# Всё стереть
make clean

