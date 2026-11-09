# Changelog

## [v1.1.0] — 2026-10-09

Polish wave **v1.0.9 → v1.1.0** (07–09.10.2026). Public HTTP now under **`/api/v1/*`**.

### Added
- Canonical REST prefix `/api/v1` (Gateway + OpenAPI + FE + k6); legacy `/api/*` rewrite
- Companion daily gift: `POST /api/v1/game/companion/daily-gift` (+1000 tickets, idempotent)
- Record-beaten in-app notifications → `/leaderboard?game=&level=&highlight=`
- Nick refresh across LB via NATS `event.user.nickname.updated`
- Avatar upload (any auth user) on Profile + Authors
- Seeds: `make seed-lb-demo`, `make seed-card-artists`, `make seed-v110`
- Gears flower palettes; Tamagotchi room chrome; shop skin/theme filters
- Docs: SECURITY_AUDIT_RU, Miro stickers/review, API_V1, LOAD_POSTURE, PARKED_WISHES

### Changed
- FE nickname helper (`getNickname`) for all game submits
- Inventory uploads no longer author-only
- About / admin analytics copy; ticklist closed for v1.1.0 wave

### Security / load
- CSRF/XSS/SQL checklist + rate-limit notes (`SECURITY_AUDIT_RU.md`)
- CORE k6 browse/purchase green; legacy 500 VU login blast = expected fail

### Decision-locked (not bugs)
- C4 payouts deferred (lock D); full 3D/Dodo engines out of scope — see `PARKED_WISHES.md`

---

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
```

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

Все заметные изменения в проекте Event Horizon.

## [v1.0.0] — 2026-06-23

### Добавлено
- Первый релиз
- Основные сервисы: Auth, Game, Billing, Leaderboard, Gateway, Balancer
- NATS JetStream
- WebSocket real-time обновления
- Docker Compose
- Prometheus + Grafana
- Миграции через goose
- Статическая сборка бинарников

---

# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📦 Архитектура (актуально v1.0.1, 24.06.2026)

```text
[React Client :5173]
    │ HTTP (JSON)
    ▼
[Balancer :8079] — самописный, Least Connections
    │ HTTP
    ▼
[Gateway 1-3 :8081-8083] — JWT, HTTP → gRPC
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
```

---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Запустить всё одной командой
make deploy

# 3. Проверить
make status
```

**Готово!** Всё поднимется автоматически:
- Docker-контейнеры (PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana)
- Миграции баз данных
- Все микросервисы

---

## 📍 Эндпоинты (доступны через балансировщик :8079)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/auth/register` | Регистрация |
| `POST` | `/api/auth/login` | Логин (JWT) |
| `GET` | `/api/billing/balance/all` | Баланс (лампочки/билетики) |
| `POST` | `/api/game/submit` | Отправить рекорд |
| `GET` | `/api/leaderboard` | Топ-10 (публичный) |
| `WS` | `/ws/leaderboard` | WebSocket обновления |

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# Отправить рекорд
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard

# В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Запустить всё
make deploy

docker-compose -f deployments/docker-compose.cluster.yml up -d

# Посмотреть логи
make logs

docker-compose -f deployments/docker-compose.cluster.yml logs -f

# Статус контейнеров
make ps

docker-compose -f deployments/docker-compose.cluster.yml ps

# Остановить всё
make down

docker-compose -f deployments/docker-compose.cluster.yml down

# Полная очистка (удалить volumes)
make clean

docker-compose -f deployments/docker-compose.cluster.yml down -v
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| **Prometheus** | `9090` | [http://localhost:9090](http://localhost:9090) |
| **Grafana** | `3000` | [http://localhost:3000](http://localhost:3000) (admin/admin) |
| **Jaeger** | `16686` | [http://localhost:16686](http://localhost:16686) |
| **NATS Exporter** | `7777` | [http://localhost:7777/metrics](http://localhost:7777/metrics) |

---

## 🧩 Компоненты и порты

### Микросервисы

| Сервис | gRPC | Metrics | БД | Redis |
|--------|------|---------|-----|-------|
| **Auth** | `5051` | `9091` | PG `5460` | `6379` |
| **Game** | `5052` | `9092` | PG `5461` | `6380` |
| **Billing** | `5053` | `9093` | PG `5462` | `6381` |
| **Leaderboard** | `5054` | `9094` | PG `5463` | `6382` |
| **Gateway** | HTTP `8081-8083` | `9095-9097` | — | — |
| **Balancer** | HTTP `8079` | `9098` | — | — |

### Инфраструктура

| Сервис | Порт | Назначение |
|--------|------|------------|
| NATS | `4222` | Событийная шина |
| NATS мониторинг | `8222` | JSON-метрики |
| Jaeger UI | `16686` | Трассировка |
| Prometheus | `9090` | Метрики |
| Grafana | `3000` | Дашборды |

---

## 📚 Документация

- [Архитектура и схемы](./confluence/architecture/)
- [История релизов](./confluence/history/)
- [Технический долг](./confluence/tech_debt/)
- [FAQ](./confluence/faq/)
- [CHANGELOG.md](./CHANGELOG.md)

---

## 🧪 Тестирование

Запустить E2E-тесты (k6):

```bash
cd deployments/k6
k6 run e2e-test.js
```

---

## 🔮 Планы на следующие спринты

### 🔥 Ближайшие задачи (1–2 недели)

- [ ] **Нагрузочное тестирование (k6)** — прогнать все сценарии, замерить RPS, latency
- [ ] **Рефакторинг кода** — убрать хардкод, добавить структурные логи, комментарии
- [ ] **Rate Limiter** — раскомментировать, настроить лимиты (10/сек на пользователя)
- [ ] **Документация API** — OpenAPI/Swagger, README для каждого сервиса
- [ ] **Тесты** — юнит-тесты (≥70% покрытия), интеграционные тесты (testcontainers)

### ⚙️ DevOps (1–2 недели)

- [ ] **CI/CD** — GitHub Actions: сборка, пуш в Docker Hub, деплой через SSH
- [ ] **Ansible** — автоматизация установки Docker, копирования бинарников
- [ ] **k3s (Kubernetes)** — Helm-чарты, Ingress (Traefik), горизонтальное масштабирование
- [ ] **Service Discovery** — Consul для регистрации сервисов

### 🧩 Новые сервисы (1–2 месяца)

| Сервис | Назначение | Порт (gRPC) |
|--------|------------|-------------|
| **Shop** | Магазин за билетики | `5055` |
| **Notification** | Push, Email, Telegram | `5056` |
| **Analytics** | DAU, MAU, Retention (ClickHouse) | `5057` |
| **Payment** | Реальные платежи (Boosty/Stripe) | `5058` |

### 🎮 Игровой контент

- [ ] Добавить игры: `flappy`, `towers`, `memory`
- [ ] Уровни сложности (1–20)
- [ ] Достижения (achievements)
- [ ] Блинопекарня (магазин за лампочки)

### 🧠 Устойчивость

- [ ] Circuit Breaker + Bulkhead
- [ ] Retry с джиттером
- [ ] Graceful shutdown
- [ ] Алерты в Telegram (Alertmanager)

---

## 🧑‍💻 Команда

- **Backend & DevOps:** Денис Матвеев ([Eastwesser](https://github.com/Eastwesser))
- **Архитектура:** Микросервисная, событийно-ориентированная
- **Деплой:** Docker Compose (сейчас) → k3s (в планах)

---

## 📦 Версия

**Текущая:** `v1.0.1` (24.06.2026)  
**Следующий релиз:** `v1.1.0` (план — 30.06.2026)

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub  
🐛 Создай Issue  
📬 Напиши мне: [eastwesser@gmail.com](mailto:eastwesser@gmail.com)

---

**Event Horizon — играй, соревнуйся, побеждай!** 🚀

## [v1.0.1] — 2026-06-24

### Добавлено
- Jaeger для Gateway
- NATS Exporter для Prometheus
- Автоматические миграции через `make deploy`

### Исправлено
- WebSocket через балансировщик
- Healthcheck NATS

---

## [v1.0.2] — 2026-06-29

### Добавлено
- **Мониторинг (полный стек)**:
  - Gateway метрики: `gateway_requests_total`, `gateway_request_duration_seconds`
  - Balancer метрики: `balancer_active_connections`, `balancer_requests_total`
  - Redis Exporter (`9121`)
  - PostgreSQL Exporter (`9187`)
  - Скрипты сбора метрик: `metrics_collector.sh`, `metrics_snapshot.sh`, `metrics_snapshot_dif.sh`
- **Grafana**: дашборд Event Horizon с панелями (RPS, Latency, Go, NATS, Redis, PG)

### Исправлено
- Jaeger для Gateway (добавлен `initTracer`)
- WebSocket прокси (`vite.config.ts`)
- Healthcheck NATS
- Автоматизация миграций (`make deploy`)

---

## [v1.0.3] — 2026-07-05

### Добавлено
- **Profile Service** — агрегация данных пользователя (CQRS + Read Model)
  - Собирает данные из Auth (через NATS `user.registered`)
  - Собирает данные из Game (через NATS `score.updated`)
  - Хранит профиль в отдельной БД
  - Отдаёт единый профиль через gRPC
- **NATS Hub** — централизованное управление Stream'ами
  - Создаёт Stream `EVENTS` при старте
  - Добавлены subject'ы: `event.>`, `score.updated`, `user.registered`, `shop.purchased`, `payment.completed`
- **Jaeger** — добавлен `JAEGER_ENDPOINT=jaeger:4317` во все сервисы
- **pprof** — добавлен во все сервисы (импорт `net/http/pprof`)
- **Redis Exporter** — настроен на порту `9121`
- **PostgreSQL Exporter** — настроен на порту `9187`
- **Makefile** — добавлены команды:
  - `migrate-profile` — миграции для Profile Service
  - `build-nats-hub` — сборка nats-hub

### Изменено
- **Gateway**:
  - Добавлен эндпоинт `/api/profile` (агрегированный профиль)
  - Добавлен `PROFILE_ADDR=profile:50060`
  - Добавлен импорт `profilePb`
- **Game**:
  - Убрано создание Stream `SCORES` (теперь использует `EVENTS`)
- **Leaderboard**:
  - Убрано создание Stream `SCORES`
  - Добавлен `depends_on` для NATS и nats-hub
- **Profile**:
  - Убрано создание Stream (использует `EVENTS`)
- **Docker Compose**:
  - Добавлен `postgres-profile` (порт `5464`)
  - Добавлен `nats-hub`
  - Добавлен `profile` (порт `50060`, метрики `9099`)
  - Добавлены `depends_on` для корректного порядка запуска

### Исправлено
- Никнейм теперь привязан к `userId` (обновляется через бэкенд)
- Статистика привязана к `userId` в `localStorage`
- WebSocket исправлен (прокси на `8079` вместо `8080`)
- Решена проблема с дублированием Stream'ов в NATS

# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📦 Архитектура (актуально v1.0.4, 05.07.2026)

```text
          [React Client :5173]
               │ HTTP (JSON)
               ▼
           [Balancer :8079] — самописный, Least Connections
               │ HTTP
               ▼
           [Gateway 1-3 :8081-8083] — JWT, HTTP → gRPC
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
    ┌──────────────────────────────────────────────────────────────┐
    │                   NATS КЛАСТЕР (3 ноды)                      │
    │        nats-1 :4222  |  nats-2 :4223  |  nats-3 :4224        │
    │   ── создаёт Stream `EVENTS` через NATS Hub ──►              │
    │      Subjects: score.updated, user.registered, shop.*        │
    └──────────────────────────────────────────────────────────────┘
               │
               ▼
    Profile Service подписан на score.updated, user.registered
               │
               ▼
    Leaderboard подписан → обновляет Redis → WebSocket → клиент
```

---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Запустить всё одной командой
make deploy

# 3. Проверить
make status
```

**Готово!** Всё поднимется автоматически:
- Docker-контейнеры (PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana)
- Миграции баз данных
- Все микросервисы, включая Profile Service и NATS Hub

---

## 📍 Эндпоинты (доступны через балансировщик :8079)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/auth/register` | Регистрация |
| `POST` | `/api/auth/login` | Логин (JWT) |
| `GET` | `/api/billing/balance/all` | Баланс (лампочки/билетики) |
| `POST` | `/api/game/submit` | Отправить рекорд |
| `GET` | `/api/leaderboard` | Топ-10 (публичный) |
| `GET` | `/api/profile` | Полный профиль пользователя (агрегированный) |
| `WS` | `/ws/leaderboard` | WebSocket обновления |

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# Отправить рекорд
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"test-user","game_id":"hexagon","level":1,"score":150,"seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'

# Получить профиль (агрегированный)
curl -X GET http://localhost:8079/api/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard

# В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Запустить всё
make deploy

docker-compose -f deployments/docker-compose.cluster.yml up -d

# Посмотреть логи
make logs

docker-compose -f deployments/docker-compose.cluster.yml logs -f

# Статус контейнеров
make ps

docker-compose -f deployments/docker-compose.cluster.yml ps

# Остановить всё
make down

docker-compose -f deployments/docker-compose.cluster.yml down

# Полная очистка (удалить volumes)
make clean

docker-compose -f deployments/docker-compose.cluster.yml down -v
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| **Prometheus** | `9090` | [http://localhost:9090](http://localhost:9090) |
| **Grafana** | `3000` | [http://localhost:3000](http://localhost:3000) (admin/admin) |
| **Jaeger** | `16686` | [http://localhost:16686](http://localhost:16686) |
| **NATS Exporter** | `7777` | [http://localhost:7777/metrics](http://localhost:7777/metrics) |

---

## 🧩 Компоненты и порты

### Микросервисы

| Сервис | gRPC | Metrics | БД | Redis |
|--------|------|---------|-----|-------|
| **Auth** | `5051` | `9091` | PG `5460` | `6379` |
| **Game** | `5052` | `9092` | PG `5461` | `6380` |
| **Billing** | `5053` | `9093` | PG `5462` | `6381` |
| **Leaderboard** | `5054` | `9094` | PG `5463` | `6382` |
| **Profile** | `50060` | `9099` | PG `5464` | — |
| **Gateway** | HTTP `8081-8083` | `9095-9097` | — | — |
| **Balancer** | HTTP `8079` | `9098` | — | — |

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| **NATS-1** | `4222` (клиент), `8222` (мониторинг) | Узел кластера |
| **NATS-2** | `4223` (клиент), `8223` (мониторинг) | Узел кластера |
| **NATS-3** | `4224` (клиент), `8224` (мониторинг) | Узел кластера |
| **NATS Hub** | — | Создаёт Stream EVENTS (инфраструктурный) |
| **Jaeger UI** | `16686` | Трассировка |
| **Prometheus** | `9090` | Метрики |
| **Grafana** | `3000` | Дашборды |

---

## 📚 Документация

- [Архитектура и схемы](./confluence/architecture/)
- [История релизов](./confluence/history/)
- [Технический долг](./confluence/tech_debt/)
- [FAQ](./confluence/faq/)
- [CHANGELOG.md](./CHANGELOG.md)

---

## 🧪 Тестирование

Запустить E2E-тесты (k6):

```bash
cd deployments/k6
k6 run e2e-test.js
```

---

## 🔮 Планы на следующие спринты

### 🔥 Ближайшие задачи (1–2 недели)

- [ ] **Нагрузочное тестирование (k6)** — прогнать все сценарии, замерить RPS, latency
- [ ] **Рефакторинг кода** — убрать хардкод, добавить структурные логи, комментарии
- [ ] **Rate Limiter** — раскомментировать, настроить лимиты (10/сек на пользователя)
- [ ] **Документация API** — OpenAPI/Swagger, README для каждого сервиса
- [ ] **Тесты** — юнит-тесты (≥70% покрытия), интеграционные тесты (testcontainers)

### ⚙️ DevOps (1–2 недели)

- [ ] **CI/CD** — GitHub Actions: сборка, пуш в Docker Hub, деплой через SSH
- [ ] **Ansible** — автоматизация установки Docker, копирования бинарников
- [ ] **k3s (Kubernetes)** — Helm-чарты, Ingress (Traefik), горизонтальное масштабирование
- [ ] **Service Discovery** — Consul для регистрации сервисов

### 🧩 Новые сервисы (1–2 месяца)

| Сервис | Назначение | Порт (gRPC) |
|--------|------------|-------------|
| **Shop** | Магазин за билетики | `50055` |
| **Notification** | Push, Email, Telegram | `50056` |
| **Analytics** | DAU, MAU, Retention (ClickHouse) | `50057` |
| **Payment** | Реальные платежи (Boosty/Stripe) | `50058` |

### 🎮 Игровой контент

- [ ] Добавить игры: `flappy`, `towers`, `memory`
- [ ] Уровни сложности (1–20)
- [ ] Достижения (achievements)
- [ ] Блинопекарня (магазин за лампочки)

### 🧠 Устойчивость

- [ ] Circuit Breaker + Bulkhead
- [ ] Retry с джиттером
- [ ] Graceful shutdown
- [ ] Алерты в Telegram (Alertmanager)

---

## 🧑‍💻 Команда

- **Backend & DevOps:** Денис Матвеев ([Eastwesser](https://github.com/Eastwesser))
- **Архитектура:** Микросервисная, событийно-ориентированная
- **Деплой:** Docker Compose (сейчас) → k3s (в планах)

---

## 📦 Версия

**Текущая:** `v1.0.4` (05.07.2026)

### Что нового в v1.0.4
- **NATS кластер из 3 нод** — отказоустойчивая событийная шина.
- **Персистентные рекорды** — Game сохраняет рекорды в PostgreSQL.
- **Восстановление Leaderboard** — данные не теряются при перезапуске.

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub  
🐛 Создай Issue  
📬 Напиши мне: [eastwesser@gmail.com](mailto:eastwesser@gmail.com)

---

# 🎮 Event Horizon (ОББНОВЛЕНО 3 августа 2026)

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📦 Архитектура (актуально v1.0.5, 15.07.2026)

```text
          [React Client :5173]
               │ HTTP (JSON)
               ▼
           [Balancer :8079] — самописный, Least Connections
               │ HTTP
               ▼
           [Gateway 1-3 :8081-8083] — JWT, HTTP → gRPC
               │ gRPC
               ▼
┌──────────────┼──────────────┬──────────────┬──────────────┐
│              │              │              │              │
▼              ▼              ▼              ▼              ▼
Auth :5051     Game :5052     Billing :5053  Leaderboard   Shop :50055
│              │              │              :5054          │
▼              ▼              ▼              ▼              ▼
PG :5460       PG :5461       PG :5462       PG :5463      PG :5465 + Redis :6383
(users)        (scores)       (balances)     + Redis :6382 (items, inventory)
│              │              │              (leaderboard)  │
└──────────────┼──────────────┴──────────────┴──────────────┘
               │
               ▼
    ┌──────────────────────────────────────────────────────────────┐
    │                   NATS КЛАСТЕР (3 ноды)                      │
    │        nats-1 :4222  |  nats-2 :4223  |  nats-3 :4224        │
    │   ── создаёт Stream `EVENTS` через NATS Hub ──►              │
    │      Subjects: score.updated, user.registered, shop.*        │
    └──────────────────────────────────────────────────────────────┘
               │
               ▼
    Profile :50060 — подписан на score.updated, user.registered
               │
               ▼
    Leaderboard обновляет Redis → WebSocket → клиент
```

---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Запустить всё одной командой
make deploy

# 3. Проверить
make status
```

**Готово!** Всё поднимется автоматически:
- 30+ Docker-контейнеров (PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana)
- Миграции баз данных
- Все микросервисы, включая Profile, Shop и NATS Hub
- Мониторинг (Prometheus + Grafana)

---

## 📍 Эндпоинты (доступны через балансировщик :8079)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/auth/register` | Регистрация |
| `POST` | `/api/auth/login` | Логин (JWT) |
| `POST` | `/api/game/submit` | Отправить рекорд |
| `GET` | `/api/billing/balance/all` | Баланс (лампочки/билетики) |
| `GET` | `/api/leaderboard` | Топ-10 (публичный) |
| `GET` |	`/api/shop/items` |	Список товаров |
| `POST` |	`/api/shop/purchase` |	Купить товар (списание билетиков) |
| `GET` |	`/api/shop/inventory` |	Инвентарь пользователя |
| `GET` | `/api/profile` | Полный профиль пользователя (агрегированный) |
| `WS` | `/ws/leaderboard` | WebSocket обновления |

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# Получить баланс
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Посмотреть товары в магазине
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Купить товар
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.'

# Посмотреть инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Отправить рекорд
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"7fc8a659-1bb2-4d7c-b60e-c140239d5c62","game_id":"hexagon","level":1,"score":150,"user_email":"test@example.com","seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'

# Получить профиль
curl -X GET http://localhost:8079/api/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'  
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard

# В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Запустить всё
make deploy

docker-compose -f deployments/docker-compose.cluster.yml up -d

# Посмотреть логи
make logs

docker-compose -f deployments/docker-compose.cluster.yml logs -f

# Статус контейнеров
make ps

docker-compose -f deployments/docker-compose.cluster.yml ps

# Остановить всё
make down

docker-compose -f deployments/docker-compose.cluster.yml down

# Полная очистка (удалить volumes)
make clean

docker-compose -f deployments/docker-compose.cluster.yml down -v
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| **Prometheus** | `9090` | [http://localhost:9090](http://localhost:9090) |
| **Grafana** | `3000` | [http://localhost:3000](http://localhost:3000) (admin/admin) |
| **Jaeger** | `16686` | [http://localhost:16686](http://localhost:16686) |
| **NATS Exporter** | `7777` | [http://localhost:7777/metrics](http://localhost:7777/metrics) |

---

## 🧩 Компоненты и порты

### Микросервисы

| Сервис | gRPC | Metrics | БД | Redis |
|--------|------|---------|-----|-------|
| **Auth** | `5051` | `9091` | PG `5460` | `6379` |
| **Game** | `5052` | `9092` | PG `5461` | `6380` |
| **Billing** | `5053` | `9093` | PG `5462` | `6381` |
| **Leaderboard** | `5054` | `9094` | PG `5463` | `6382` |
| **Profile** | `50060` | `9099` | PG `5464` | — |
| **Shop** |	`50055` |	`9095` | PG `5465` |	`6383` |
| **Gateway** | HTTP `8081-8083` | `9095-9097` | — | — |
| **Balancer** | HTTP `8079` | `9098` | — | — |

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| **NATS-1** | `4222` (клиент), `8222` (мониторинг) | Узел кластера |
| **NATS-2** | `4223` (клиент), `8223` (мониторинг) | Узел кластера |
| **NATS-3** | `4224` (клиент), `8224` (мониторинг) | Узел кластера |
| **NATS Hub** | — | Создаёт Stream EVENTS (инфраструктурный) |
| **Jaeger UI** | `16686` | Трассировка |
| **Prometheus** | `9090` | Метрики |
| **Grafana** | `3000` | Дашборды |

### 🎮 Игры

| Игра |	Описание |	Скины |
|--------|---------|------------|
| **Flappy Bird** |	`Лети и не врезайся в трубы` |	`Золотая птичка, Радужные трубы` |
| **Hexagon** |	`Гексагональный пазл с блинами` |	`Космические блины` |
| **Towers** |	`Строй башню из плавающих блоков` |	`Радужные блоки` |
| **Memory** |	`Найди пары фруктов` |	`Карточки со зверями` |

---

## 📚 Документация

- [Архитектура и схемы](./confluence/architecture/)
- [История релизов](./confluence/history/)
- [Технический долг](./confluence/tech_debt/)
- [FAQ](./confluence/faq/)
- [CHANGELOG.md](./CHANGELOG.md)

---

## 🧪 Тестирование

Запустить E2E-тесты (k6):

```bash
cd deployments/k6
k6 run e2e-test.js
```

---

## 🔮 Планы на следующие спринты

### 🔥 Ближайшие задачи (1–2 недели)

- [ ] **Нагрузочное тестирование (k6)** — прогнать все сценарии, замерить RPS, latency
- [ ] **Оптимизация** - индексов в БД всех сервисов
- [ ] **Рефакторинг кода** — убрать хардкод, добавить структурные логи, комментарии
- [ ] **Rate Limiter** — раскомментировать, настроить лимиты (100/сек на пользователя)
- [ ] **Документация API** — OpenAPI/Swagger, README для каждого сервиса
- [ ] **Тесты** — юнит-тесты (≥70% покрытия), интеграционные тесты (testcontainers)

### ⚙️ DevOps (1–2 недели)

- [X] **CI/CD** — GitHub Actions: сборка, пуш в Docker Hub, деплой через SSH
- [X] **Ansible** — автоматизация установки Docker, копирования бинарников
- [ ] **k3s (Kubernetes)** — Helm-чарты, Ingress (Traefik), горизонтальное масштабирование
- [ ] **Service Discovery** — Consul для регистрации сервисов

### 🧩 Новые сервисы (1–2 месяца)

| Сервис | Назначение | Порт (gRPC) |
|--------|------------|-------------|
| **Notification** | Push, Email, Telegram | `50056` |
| **Analytics** | DAU, MAU, Retention (ClickHouse) | `50057` |
| **Payment** | Реальные платежи (Boosty/Stripe) | `50058` |

### 🎮 Игровой контент

- [ ] Лампочки как бусты в играх (замедление, подсказки)
- [ ] Уровни сложности (1–20)
- [ ] Достижения (achievements)

### 🧠 Устойчивость

- [ ] Circuit Breaker + Bulkhead
- [ ] Retry с джиттером
- [X] Graceful shutdown
- [ ] Алерты в Telegram (Alertmanager)

---

## 🧑‍💻 Команда

- **Backend & DevOps:** Денис Матвеев ([Eastwesser](https://github.com/Eastwesser))
- **Архитектура:** Микросервисная, событийно-ориентированная
- **Деплой:** Docker Compose (сейчас) → k3s (в планах)

---

## 📦 Версия

**Текущая:** `v1.0.5` (15.07.2026)

### Что нового в v1.0.5

- **Магазин (Shop Service)** — покупка скинов, инвентарь, интеграция с Billing
- **Космические блины** — полный маппинг 8 типов для Hexagon
- **Memory** — скин "Карточки со зверями" (15 видов)
- **Flappy Bird** — синяя птичка, радужные трубы (сохранена форма)
- **Towers** — дефолтные красные блоки + радужные скины
- **Дата покупки** — добавлено поле purchased_at в /api/shop/inventory
- **NATS кластер** — 3 ноды + мониторинг в Prometheus
- **WebSocket** — исправлено подключение через балансировщик

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub  
🐛 Создай Issue  
📬 Напиши мне: [eastwesser@gmail.com](mailto:eastwesser@gmail.com)

---

**Event Horizon — играй, соревнуйся, побеждай!** 🚀
---

## [v1.0.8] — 2026-08-30

Patch release: secrets hygiene, thin NATS deploy, purchase path on JetStream, Gateway/login hardening.

### Added
- `.env.example` + env substitution across compose / Patroni / k3s secrets
- Thin vs heavy deploy: `make deploy` (NATS + obs + fulfillment/notification/analytics); `make deploy-heavy` (Kafka profile); `make stop-heavy`
- Shop → NATS `purchase.paid`; fulfillment & notification JetStream consumers (Kafka only if `KAFKA_BROKERS` set)
- NATS Hub EVENTS subjects: `purchase.paid`, `purchase.fulfilled`, …
- Gateway response cache (whoami / profile) + `_partial` fallback; Boosty HMAC webhook verify
- Grafana alerting provisioning; rebuild/push/coverage helper scripts
- Frontend Auth: password minLength **8** (aligned with Auth service)

### Fixed
- Empty `POSTGRES_*` / Grafana env on deploy (`${VAR:-eventhorizon}` defaults)
- Gateway wrapped gRPC `InvalidArgument` → HTTP **400** (was 500 / “connection error”)
- Game `depends_on` healthy Postgres (Prometheus scrape)
- Notification `/metrics`; analytics NATS durable names (dashes); ClickHouse readiness race
- `rebuild-all-backend.sh` no longer requires `task` CLI

### Docs
- `confluence/history/2026-08/30.08.2026/Issues.md` — full issue/fix index
- OPTIMIZATION / DB_ISSUES / FRONTEND_STATUS / GitGuardian assessment

# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.0.8-brightgreen.svg)](https://github.com/Eastwesser/event-horizon/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon** — production-style microservices platform: gRPC mesh, NATS events, OpenAPI gateway, observability, and a React game client. Designed for learning and as a reference implementation you can deploy locally in one command.

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture per service · Gateway HTTP `/api/` → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, secrets via env (`.env` gitignored) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · [OpenAPI `/docs`](http://localhost:8079/docs) when running |

---

## 📦 Архитектура (актуально v1.0.8, 30.08.2026)

Полная схема: [`confluence/architecture/EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) · Mermaid: `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.7-system-design.md` · Miro legacy: `confluence/architecture/SYSTEM_DESIGN/event-horizon-v1.0.6.png`

```text
[GitHub Actions] ──SSH/Ansible──► [VM] ──docker-compose──► Event Horizon

[React :5173] ──HTTP──► [Balancer :8079] ──HTTP──► [Gateway ×3 :8081–8083]
                                                      │ JWT · HTTP→gRPC
                                                      ▼
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Auth :50051   │ Game :50052   │ Billing:50053 │ Leaderboard   │ Shop :50055   │
│ PG:5460 Redis │ PG:5461       │ PG:5462 Redis │ :50054        │ PG:5465 Redis │
└───────────────┴───────────────┴───────────────┴─PG:5463+R6382─┴───────────────┘
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Inventory     │ Profile:50060 │ Payment:50058 │ Authors:50061 │ History:50062 │
│ :50059        │ PG:5464       │ sub/merch gate│ PG:5468 Redis │ PG:5469       │
└───────────────┴───────────────┴───────────────┴───────────────┴───────────────┘
┌───────────────┬───────────────────────────────────────────────────────────────┐
│ Analytics     │ NATS JetStream :4222/:4223/:4224  Stream EVENTS (NATS Hub)    │
│ :50057        │ subjects: score.updated, purchase.paid/fulfilled, shop.*, …   │
│ ClickHouse    │ async ──► Profile / Leaderboard / Notification / Fulfillment  │
│ :8123/:9000   │ Leaderboard Redis Sorted Set ──WS──► Client                   │
└───────────────┴───────────────────────────────────────────────────────────────┘
```

**Deploy profiles (v1.0.8):** `make deploy` = thin stack (NATS + apps + ClickHouse + Prometheus/Grafana/Jaeger + fulfillment/notification/analytics). Kafka is opt-in: `make deploy-heavy` / `make stop-heavy`.
---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Локальные секреты — скопируй шаблон и задай свои значения
cp .env.example .env
# Отредактируй: JWT_SECRET, POSTGRES_PASSWORD, GRAFANA_ADMIN_PASSWORD (см. .env.example)

# 3. Запустить thin-стек одной командой (NATS path; без Kafka)
make deploy

# 4. Проверить
make status

# 5. Опционально: Kafka broker на более мощной машине
# make deploy-heavy

# 6. Или запустить в k3s
make deploy-k3s
```

Готово! Всё поднимется автоматически:

- PostgreSQL, Redis, NATS, ClickHouse, Jaeger, Prometheus, Grafana (+ apps)
- Миграции баз данных
- Микросервисы: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Fulfillment, Notification, NATS Hub, …
- Purchase path: Shop → NATS `purchase.paid` → Fulfillment / Notification
- Опционально: Kafka (`make deploy-heavy`), k3s (`make deploy-k3s`)

---

## 📍 Эндпоинты (доступны через балансировщик :8079)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | /api/auth/register | Регистрация |
| POST | /api/auth/login | Логин (JWT) |
| POST | /api/game/submit | Отправить рекорд |
| GET | /api/billing/balance/all | Баланс (лампочки/билетики) |
| GET | /api/leaderboard | Топ-10 (публичный) |
| GET | /api/shop/items | Список товаров |
| POST | /api/shop/purchase | Купить товар (списание билетиков) |
| GET | /api/shop/inventory | Инвентарь пользователя |
| GET | /api/profile | Полный профиль пользователя (агрегированный) |
| GET/POST | /api/payment/… | Подписка / CanPurchaseMerch |
| GET/POST | /api/authors/… | Авторы |
| GET | /api/history | История событий |
| GET | /api/analytics/… | Аналитика (admin) |
| GET | /openapi.yaml · /docs | OpenAPI + Swagger UI |
| WS | /ws/leaderboard | WebSocket обновления |

### API prefix (v1.0.x)

Публичный HTTP API живёт под **`/api/`** — без сегмента версии (`/api/v1/` не используется).

| Кто | Как вызывать |
|-----|----------------|
| **Gateway / curl / OpenAPI** | полный путь: `/api/shop/purchase` |
| **Frontend (axios)** | `baseURL: '/api'` + относительный путь: `api.post('/shop/purchase')` |
| **WebSocket / ops** | вне `/api`: `/ws/leaderboard`, `/health`, `/ready` |

Подробная таблица маршрутов и RBAC: [`confluence/architecture/API_ROUTES.md`](confluence/architecture/API_ROUTES.md).

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# Получить баланс
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Посмотреть товары в магазине
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Купить товар
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.'

# Посмотреть инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Отправить рекорд
curl -X POST http://localhost:8079/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"7fc8a659-1bb2-4d7c-b60e-c140239d5c62","game_id":"hexagon","level":1,"score":150,"user_email":"tuzer@example.com","seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'

# Получить профиль
curl -X GET http://localhost:8079/api/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard
```

```javascript
// В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Thin stack (NATS; Kafka OFF)
make deploy

# Kafka broker only (apps stay on NATS unless KAFKA_BROKERS set)
make deploy-heavy
make stop-heavy

# Посмотреть логи / статус
make logs
make ps

# Остановить всё / полная очистка volumes
make down
make clean

# Собрать / запушить образы
make docker-build-all
make docker-push-all

# k3s / Ansible
make deploy-k3s
make undeploy-k3s
make delivery-dev
```

### Пересборка после изменений в коде

```bash
bash scripts/rebuild-proto.sh              # protoc для gRPC-сервисов (Gateway без proto)
bash scripts/rebuild-services.sh game analytics gateway
bash scripts/docker-push-images.sh game analytics gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d game analytics gateway gateway-2 gateway-3
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3000 | http://localhost:3000 (`GRAFANA_ADMIN_PASSWORD`, default admin) |
| Jaeger | 16686 | http://localhost:16686 |
| NATS Exporter | 7777 | http://localhost:7777/metrics |

---

## 🧩 Компоненты и порты

### Микросервисы

| Сервис | gRPC | Metrics | БД | Redis |
|--------|------|---------|-----|-------|
| Auth | 50051 | 9091 | PG 5460 | 6379 |
| Game | 50052 | 9092 | PG 5461 | — |
| Billing | 50053 | 9093 | PG 5462 | 6381 |
| Leaderboard | 50054 | 9094 | PG 5463 | 6382 |
| Shop | 50055 | 9095 | PG 5465 | 6383 |
| Analytics | 50057 | 9106 | ClickHouse 8123/9000 | — |
| Fulfillment | — | 9101 | — | — |
| Notification | — | 9102 | — | — |
| Payment | 50058 | 9103 | PG 5467 | 6386 |
| Inventory | 50059 | 9096 | PG / Mongo | Redis |
| Profile | 50060 | 9099 | PG 5464 | — |
| Authors | 50061 | 9104 | PG 5468 | 6387 |
| History | 50062 | 9105 | PG 5469 | — |
| Gateway | HTTP 8081-8083 | 9095-9097 | — | — |
| Balancer | HTTP 8079 | 9098 | — | — |

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| NATS-1 | 4222, 8222 | Узел кластера |
| NATS-2 | 4223, 8223 | Узел кластера |
| NATS-3 | 4224, 8224 | Узел кластера |
| NATS Hub | — | Создаёт Stream EVENTS |
| ClickHouse | 8123, 9000 | Analytics OLAP |
| Jaeger UI | 16686 | Трассировка |
| Prometheus | 9090 | Метрики |
| Grafana | 3000 | Дашборды |

---

## 🎮 Игры

| Игра | Описание | Скины |
|------|----------|-------|
| Flappy Bird | Лети и не врезайся в трубы | Золотая птичка, Радужные трубы |
| Hexagon | Гексагональный пазл с блинами | Космические блины |
| Towers (Башенки) | Строй башню из падающих блоков | Радужные блоки |
| Memory | Найди пары фруктов | Карточки со зверями |
| Hanoi (Ханойская башня) | Классика 3 стержня / кольца 3–8 | — |

---

## 📚 Документация

| Тема | Путь |
|------|------|
| Архитектура и схемы | [`confluence/architecture/EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) |
| HTTP API / RBAC | [`confluence/architecture/API_ROUTES.md`](confluence/architecture/API_ROUTES.md) |
| HTTP status codes | [`confluence/architecture/STATUS_CODES.md`](confluence/architecture/STATUS_CODES.md) |
| Load resilience | [`confluence/architecture/LOAD_RESILIENCE.md`](confluence/architecture/LOAD_RESILIENCE.md) |
| Game Outbox | [`confluence/architecture/GAME_OUTBOX.md`](confluence/architecture/GAME_OUTBOX.md) |
| Patroni HA roadmap | [`deployments/patroni/README.md`](deployments/patroni/README.md) |
| Issues & fixes (v1.0.8) | [`confluence/history/2026-08/30.08.2026/Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md) |
| Технический долг | [`confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`](confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md) |
| CHANGELOG | [`CHANGELOG.md`](CHANGELOG.md) |

---

## 🔒 Безопасность и доверие

- **Секреты не коммитятся** — скопируй [`.env.example`](.env.example) → `.env`, задай `JWT_SECRET` и `POSTGRES_PASSWORD`.
- **JWT** — задаётся через `JWT_SECRET`; Auth предупреждает при дефолтном ключе.
- **Пароли в compose** (`eventhorizon`, Patroni stubs) — только для локальной разработки; в k3s — Kubernetes Secrets.
- **OpenAPI** — канонический контракт: [`docs/openapi.yaml`](docs/openapi.yaml), Swagger UI на `/docs`.
- Вопросы по безопасности — Issue или eastwesser@gmail.com.

---

## 🧪 Тестирование

Unit-тесты (Week 2 Clean Architecture + testify):

```bash
task test              # все сервисы
task test-coverage     # покрытие по gRPC-сервисам
task w2-check          # структурный чеклист W2
```

Auth service layer is covered by unit tests with hand-written repository mocks
(`services/auth/internal/service/auth_service_test.go`). Converter layers have
package tests across gRPC services. Mongo inventory test is `//go:build integration`.

When you have network, you can optionally add testify/suite + mockery:

```bash
cd services/auth && go get github.com/stretchr/testify@v1.11.1
# then regenerate mocks via //go:generate on UserRepository
```

Нагрузочное тестирование (k6):

```bash
cd deployments/k6
k6 run loadtest.js
```

---

## 🔮 Планы на следующие спринты

### 🔥 Ближайшие задачи (1–2 недели)

- [ ] Нагрузочное тестирование (k6) — прогнать все сценарии, замерить RPS, latency
- [ ] Оптимизация индексов в БД всех сервисов
- [ ] Rate Limiter — настроить лимиты (100/сек на пользователя)
- [ ] Документация API — OpenAPI/Swagger для всех сервисов
- [ ] Юнит-тесты — покрытие ≥70%

### ⚙️ DevOps (1–2 недели)

- [X] CI/CD — GitHub Actions: сборка → Docker Hub
- [X] Ansible — автоматизация деплоя
- [X] k3s (Kubernetes) — установлен и настроен
- [ ] Helm-чарты — для управления деплоем
- [ ] Service Discovery — Consul

### 🧩 Payment / Notification / Analytics (уже в стеке)

| Сервис | Статус | Порт (gRPC) |
|--------|--------|-------------|
| Payment | ✅ реализовано (hardening: Boosty signature verification) | 50058 |
| Notification | ✅ реализовано (consumer; Telegram optional) | — |
| Analytics | ✅ реализовано (ClickHouse ingest + admin APIs) | 50057 |

### 🎮 Игровой контент

- [ ] Лампочки как бусты в играх (замедление, подсказки)
- [ ] Уровни сложности (1–20)
- [ ] Достижения (achievements)

### 🧠 Устойчивость

- [ ] Circuit Breaker + Bulkhead
- [ ] Retry с джиттером
- [X] Graceful shutdown
- [ ] Алерты в Telegram (Alertmanager)

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

Текущая: **v1.0.8** (30.08.2026)

### Что нового в v1.0.8

- **Secrets:** `.env.example`, compose/Patroni/k3s via `${…}` (no hardcoded DB passwords in yaml)
- **Thin deploy:** `make deploy` = NATS + obs + fulfillment/notification/analytics; Kafka → `make deploy-heavy`
- **Purchase on NATS:** Shop publishes `purchase.paid`; fulfillment & notification consume JetStream
- **Gateway:** response cache + `_partial` profile; gRPC InvalidArgument → 400; Boosty HMAC webhook
- **Frontend Auth:** password min 8 (match Auth); clearer errors
- **Ops:** game waits for healthy PG; notification `/metrics`; analytics durable names + CH readiness
- **Tooling:** `scripts/rebuild-*.sh`, `docker-push-images.sh`, `coverage-gate.sh`
- Issue log: [`Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md)

### Что нового в v1.0.7

- Payment / Authors / History / Analytics (+ ClickHouse) за Gateway
- Circuit breaker на всех gRPC-клиентах Gateway
- Frontend: подписка, авторы, аналитика, Ханойская башня, merch-gate (403 + `subscription_required`)
- Game **Outbox** для `score.updated` (миграция + worker; пересобери образ `game`)
- MCP/RAG (stdio) для Cursor · OpenAPI sync (auth refresh / whoami / logout)
- Patroni Auth **stubs** · ClickHouse Docker-network fix · compose `--env-file .env`
- Документация: API routes, load resilience, status codes, rebuild scripts

### Что нового в v1.0.6

- Inventory Service — управление мерчем авторов, интеграция с Shop через NATS
- CI/CD — GitHub Actions собирает и пушит образы на Docker Hub
- Ansible — автоматический деплой на сервер
- k3s (Kubernetes) — установлен и настроен кластер
- Makefile — новые команды: docker-build-all, docker-push-all, deploy-k3s, undeploy-k3s
- Delivery Pipeline — полный пайплайн для production
- Документация — обновлены README, добавлены гайды по деплою

### Что нового в v1.0.5

- Магазин (Shop Service) — покупка скинов, инвентарь, интеграция с Billing
- Космические блины — полный маппинг 8 типов для Hexagon
- Memory — скин "Карточки со зверями" (15 видов)
- Flappy Bird — синяя птичка, радужные трубы (сохранена форма)
- Towers — дефолтные красные блоки + радужные скины
- Дата покупки — добавлено поле purchased_at в /api/shop/inventory
- NATS кластер — 3 ноды + мониторинг в Prometheus
- WebSocket — исправлено подключение через балансировщик

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub
🐛 Создай Issue
📬 Напиши мне: eastwesser@gmail.com

Event Horizon — играй, соревнуйся, побеждай! 🚀

## [v1.0.9] — 2026-10-03

### Added
- Berserk CCG card seed: 280 cards across sets 4–8
- Attribute filters panel (element, rarity, class, flags, stats with operators, icons multi-select, artist autocomplete)
- Sort controls (name, rarity, price, artist, set, element)
- Pagination with URL state (100 per page)
- Artist pages (/authors + /authors/:id)
- Prev/next navigation on card detail
- Noiz reviews UI block on 8th-set cards (127 rows)
- Cancel purchase (refund at purchase price, idempotent)
- Slim inventory list DTO (InventoryItemsCatalog)

### Fixed
- Element normalization: леса → woods (56 cards)
- Purchase-path RBAC: user allowed on ReserveItem/ReleaseItem
- Data typos: Анная → Анна, Индарри → Индарра
- Detail hero white corners
- Shop card layout (drop description)

### Docs
- REVIEW_RESULT.md (full batch review)
- Berserk history: rules, icons, how-to-read, pricing
- NOIZ source comments

### Files
- See REVIEW_RESULT.md for full file list.

====== v 1.1.3 ======

# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.3-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.3** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy. Frontend polish v1.1.2 + backend patterns catalog v1.1.3.

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **v1.1.0 pack** | [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) · [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md) · [`DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md) |
| **Load (measured)** | [`K6_VERDICT.md`](confluence/history/2026-10/10.10.2026/K6_VERDICT.md) · [`SUITABLE_SERVER.md`](confluence/history/2026-10/09.10.2026/SUITABLE_SERVER.md) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · OpenAPI `/docs` только с `DOCS_PASSWORD` (без пароля адрес отвечает 404) |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

### v1.1.3 highlights (15.10.2026)

- **Patterns catalog:** `confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md` — 17 сервисов, 333 паттерна (Facade, Saga, Outbox, Circuit Breaker, RAG, TF-IDF, Least Conn, WebSocket ticket, etc.)
- **Roadmap:** `confluence/tech_debt/FUTURE_EVENT_HORIZON.md` — i18n (RU/EN/DE/ES), tech debt, games polish, users & prod, observability & security
- **Frontend v1.1.2:** BoostCheckbox per-game effects · Gears canvas fix · Pancaker space emojis · Hanoi disk bore · Memonia layout · ShopItemCard `isDigitalItem` (hide 998 stock for skins/themes) · Flappy hint «Нажмите пробел»

### v1.1.0 highlights (07–10.10.2026)

- `/api/v1` public REST · shop/inventory polish · achievements · history NATS  
- Companion daily tickets · avatar upload · nick→LB · record-beaten deep-link  
- Mobile safe-area + game touch targets · Miro stickers pack · security audit RU  
- Ops close: gateway v1 rewrite, Pancaker score trust, Grafana provision, CORE k6 verdict  

### Future decisions (not blockers)

| Track | Status |
|-------|--------|
| Miro board polish / Boosty posts / presentations | **You IRL** — research tomorrow |
| Optional `git tag v1.1.0` | Ask when ready (marker on commit, not a branch) |
| Unified game chrome · Tamagotchi art pass | Soft polish |
| C4 payouts · 108 login authors · full 3D | **Locked** until product OK |
| Loki / Tempo / node-exporter / SLO burn / multi-DB scrape | Dream list — [`DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md) |
| Next hardware for load demos | Prefer **8 vCPU / 32 GiB** — see SUITABLE_SERVER |

---

## 📦 Архитектура (актуально v1.1.0, 09.10.2026)

Схемы: [`EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) · Miro export: [`08.10.2026/miro/`](confluence/history/2026-10/08.10.2026/miro/) · stickers: [`MIRO_STICKERS_PASTE.md`](confluence/history/2026-10/08.10.2026/miro/MIRO_STICKERS_PASTE.md)

```text
[GitHub Actions] ──SSH/Ansible──► [VM] ──docker-compose──► Event Horizon

[React :5173] ──HTTP──► [Balancer :8079] ──HTTP──► [Gateway ×3 :8081–8083]
                                                      │ JWT · HTTP→gRPC
                                                      ▼
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Auth :50051   │ Game :50052   │ Billing:50053 │ Leaderboard   │ Shop :50055   │
│ PG:5460 Redis │ PG:5461       │ PG:5462 Redis │ :50054        │ PG:5465 Redis │
└───────────────┴───────────────┴───────────────┴─PG:5463+R6382─┴───────────────┘
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Inventory     │ Profile:50060 │ Payment:50058 │ Authors:50061 │ History:50062 │
│ :50059        │ PG:5464       │ sub/merch gate│ PG:5468 Redis │ PG:5469       │
└───────────────┴───────────────┴───────────────┴───────────────┴───────────────┘
┌───────────────┬───────────────────────────────────────────────────────────────┐
│ Analytics     │ NATS JetStream :4222/:4223/:4224  Stream EVENTS (NATS Hub)    │
│ :50057        │ subjects: score.updated, purchase.paid/fulfilled, shop.*, …   │
│ ClickHouse    │ async ──► Profile / Leaderboard / Notification / Fulfillment  │
│ :8123/:9000   │ Leaderboard Redis Sorted Set ──WS──► Client                   │
└───────────────┴───────────────────────────────────────────────────────────────┘

> **Fulfillment** — эмуляция сборки заказа (delay 10s). 
> По факту наличия товара, я просто ищу его, достаю из коробки и упаковываю для отправки.
> Заменится на реальный WMS API в будущем. См. `FUTURE_EVENT_HORIZON.md` → Tech Debt.
```

**Deploy profiles (v1.1.0):** `make deploy` = thin stack (NATS + apps + ClickHouse + Prometheus/Grafana/Jaeger + fulfillment/notification/analytics). Kafka is opt-in: `make deploy-heavy` / `make stop-heavy`. Seeds: `make seed-v110`, `make seed-lb-demo`, `make seed-card-artists`.

---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Локальные секреты — скопируй шаблон и задай свои значения
cp .env.example .env
# Отредактируй: JWT_SECRET, POSTGRES_PASSWORD, GRAFANA_ADMIN_PASSWORD (см. .env.example)

# 3. Запустить thin-стек одной командой (NATS path; без Kafka)
make deploy

# 4. Проверить
make status

# 5. Опционально: Kafka broker на более мощной машине
# make deploy-heavy

# 6. Или запустить в k3s
make deploy-k3s
```

Готово! Всё поднимется автоматически:

- PostgreSQL, Redis, NATS, ClickHouse, Jaeger, Prometheus, Grafana (+ apps)
- Миграции баз данных — после `pg_isready` на всех Postgres (первый `make deploy` после перезагрузки машины дожидается баз, затем goose)
- Микросервисы: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Fulfillment, Notification, NATS Hub, …
- Purchase path: Shop → NATS `purchase.paid` → Fulfillment / Notification
- Опционально: Kafka (`make deploy-heavy`), k3s (`make deploy-k3s`)

---

## 🪟 Windows

На Windows проект запускается из Ubuntu в WSL2. Makefile написан на bash.

1. Поставь [Docker Desktop](https://www.docker.com/products/docker-desktop/) и включи **Use the WSL 2 based engine**. В Settings → Resources → WSL integration включи дистрибутив Ubuntu.
2. Поставь Ubuntu из Microsoft Store, если его ещё нет. Клонируй репозиторий в домашний каталог Linux (`~/event_horizon`). Том Postgres на `/mnt/c` работает очень медленно.
3. В Ubuntu: `sudo apt update && sudo apt install -y make golang-go git`. Миграции применяет goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
```

4. `cp .env.example .env`, затем `make deploy`. Первый старт поднимает Postgres, ждёт `pg_isready` и только потом запускает goose и остальные сервисы. Шлюз открывается в браузере Windows: `http://localhost:8079` (WSL2 отдаёт localhost на хост).

Сборка `nats-hub` и миграции идут через `make deploy` внутри Ubuntu. Из PowerShell контейнеры можно поднять отдельно:

```powershell
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d
```

---

## 📍 Эндпоинты (балансировщик :8079)

Канон: **`/api/v1/*`**. Legacy `/api/*` на gateway один раз переписывается в v1. Микросервисы — gRPC (не HTTP). См. [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md).

| Метод | Путь | Описание |
|-------|------|----------|
| POST | /api/v1/auth/register | Регистрация |
| POST | /api/v1/auth/login | Логин (JWT) |
| POST | /api/v1/game/submit | Отправить рекорд |
| POST | /api/v1/game/companion/daily-gift | Tamagotchi: +1000 билетиков / день |
| GET | /api/v1/billing/balance/all | Баланс (лампы/билетики) |
| GET | /api/v1/leaderboard | Топ (публичный) |
| GET | /api/v1/shop/items | Каталог |
| POST | /api/v1/shop/purchase | Купить (билетики) |
| POST | /api/v1/shop/purchase/:id/cancel | Отмена покупки |
| GET | /api/v1/shop/inventory | Инвентарь |
| GET | /api/v1/profile | Профиль (+ ачивки) |
| POST | /api/v1/uploads | Картинка (auth) |
| GET/POST | /api/v1/payment/… | Подписка / merch gate |
| GET/POST | /api/v1/authors/… | Авторы |
| GET | /api/v1/history | История |
| GET | /api/v1/analytics/… | Аналитика (admin) |
| GET | /openapi.yaml · /docs | OpenAPI + Swagger UI, только если задан `DOCS_PASSWORD` |
| WS | /ws/leaderboard | Real-time LB |

| Кто | Как |
|-----|-----|
| **Gateway / curl / OpenAPI** | `/api/v1/shop/purchase` |
| **Frontend** | `baseURL: '/api/v1'` + `api.post('/shop/purchase')` |
| **WebSocket / ops** | `/ws/leaderboard`, `/health`, `/ready` |

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# Получить баланс
curl -s -X GET "http://localhost:8079/api/v1/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Посмотреть товары в магазине
curl -s -X GET http://localhost:8079/api/v1/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Купить товар
curl -X POST http://localhost:8079/api/v1/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.'

# Посмотреть инвентарь
curl -s -X GET http://localhost:8079/api/v1/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Отправить рекорд
curl -X POST http://localhost:8079/api/v1/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"7fc8a659-1bb2-4d7c-b60e-c140239d5c62","game_id":"hexagon","level":1,"score":150,"user_email":"tuzer@example.com","seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/v1/leaderboard?game_id=hexagon&limit=10" | jq '.'

# Получить профиль
curl -X GET http://localhost:8079/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard
```

```javascript
// В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Thin stack (NATS; Kafka OFF)
make deploy

# Kafka broker only (apps stay on NATS unless KAFKA_BROKERS set)
make deploy-heavy
make stop-heavy

# Посмотреть логи / статус
make logs
make ps

# Остановить всё / полная очистка volumes
make down
make clean

# Собрать / запушить образы
make docker-build-all
make docker-push-all

# k3s / Ansible
make deploy-k3s
make undeploy-k3s
make delivery-dev
```

### Пересборка после изменений в коде

```bash
bash scripts/rebuild-proto.sh              # protoc для gRPC-сервисов (Gateway без proto)
bash scripts/rebuild-services.sh game analytics gateway
bash scripts/docker-push-images.sh game analytics gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d game analytics gateway gateway-2 gateway-3
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3000 | http://localhost:3000 (`GRAFANA_ADMIN_PASSWORD`, default admin) |
| Alertmanager | 9193 | http://localhost:9193 |
| Jaeger | 16686 | http://localhost:16686 |
| NATS Exporter | 7777 | http://localhost:7777/metrics |

Daily maintenance: `make daily` (status + smoke + Prom/Grafana targets).  
Full command kit + Prom/Grafana gap list: [`confluence/history/2026-10/09.10.2026/DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md).

---

## 🧩 Компоненты и порты

### Микросервисы

Порты ниже — те, что проброшены на хост в `deployments/docker-compose.cluster.yml`. Внутри сети контейнеров Postgres слушает 5432, Redis 6379.

| Сервис | gRPC / HTTP | Metrics на хосте | Postgres на хосте | Redis на хосте |
|--------|-------------|------------------|-------------------|----------------|
| Auth | 50051 | 9091 | 5460 `eventhorizon` | 6379 |
| Game | 50052 | 9092 | 5461 `eventhorizon_game` | — |
| Billing | 50053 | 9093 | 5462 `eventhorizon_billing` | 6381 |
| Leaderboard | 50054 | 9094 | 5463 `eventhorizon_leaderboard` | 6382 |
| Shop | 50055 | 9095 | 5465 `eventhorizon_shop` | 6383 |
| Analytics | 50057 | 9106 | ClickHouse 8123 / 9000 | — |
| Fulfillment | — | 9101 | — | — |
| Notification | 50063 | 9102 | 5470 `eventhorizon_notification` | — |
| Payment | 50058 | 9103 | 5467 `eventhorizon_payment` | 6386 |
| Inventory | 50059 | 9096 | 5466 `eventhorizon_inventory` | 6384 |
| Profile | 50060 | 9099 | 5464 `eventhorizon_profile` | 6385 |
| Authors | 50061 | 9104 | 5468 `eventhorizon_authors` | 6387 |
| History | 50062 | 9105 | 5469 `eventhorizon_history` | — |
| Gateway | HTTP 8081, 8082, 8083 | не проброшен | — | 6379 (общий с Auth) |
| Balancer | HTTP 8079 | 9098 | — | — |

Gateway слушает метрики внутри контейнера (`9095`, `9096`, `9097`). На хост они не выведены: `9095` и `9096` на хосте — это Shop и Inventory, `9097` — NATS Hub. Prometheus ходит по именам `gateway:9095`, `shop:9095` внутри docker-сети, это разные контейнеры. Redis `6380` (`redis-game`) поднят, процесс Game к нему не подключён. Mongo в этом compose нет: Inventory пишет в Postgres.

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| NATS-1 | 4222 клиент, 8222 мониторинг, 6222 кластер | Узел |
| NATS-2 | 4223, 8223, 6223 | Узел |
| NATS-3 | 4224, 8224, 6224 | Узел |
| NATS Hub | 9097 | Создаёт Stream EVENTS, свои метрики |
| NATS Exporter | 7777 | `/metrics` |
| ClickHouse | 8123 HTTP, 9000 native | Analytics OLAP |
| Jaeger | 16686 UI, 4317 OTLP gRPC, 4318 OTLP HTTP | Трассировка |
| Prometheus | 9090 | Метрики |
| Grafana | 3000 | Дашборды |
| Alertmanager | 9193 → контейнер 9093 | Алерты |
| Kafka (профиль heavy) | 19092 → контейнер 9092 | Не входит в `make deploy` |
| Postgres exporter | 9187 | Метрики auth-БД |
| Redis exporter | 9121 | Метрики Redis |
| Qdrant (профиль qdrant) | 6333 | Не входит в обычный подъём |

---

## 🎮 Игры

| Игра | Описание | Скины |
|------|----------|-------|
| Flappy Bird | Лети и не врезайся в трубы | Золотая птичка, Радужные трубы |
| Pancaker (Hexagon) | Гексагональный пазл с блинами | Космические блины |
| Builder (Towers) | Строй башню из падающих блоков | Радужные блоки |
| Memonia (Memory) | Найди пары фруктов | Карточки со зверями |
| Hanoi | Классика: 3 стержня, кольца 3–8 | — |
| 2048 | Собирай плитки до 2048 | — |
| Gears | Собирай шестерни на поле | — |
| Tamagotchi | Корми спутника и забирай дневной подарок | — |

---

## 📚 Документация

| Тема | Путь |
|------|------|
| Архитектура и схемы | [`confluence/architecture/EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) |
| HTTP API / RBAC | [`confluence/architecture/API_ROUTES.md`](confluence/architecture/API_ROUTES.md) |
| HTTP status codes | [`confluence/architecture/STATUS_CODES.md`](confluence/architecture/STATUS_CODES.md) |
| Load resilience | [`confluence/architecture/LOAD_RESILIENCE.md`](confluence/architecture/LOAD_RESILIENCE.md) |
| Game Outbox | [`confluence/architecture/GAME_OUTBOX.md`](confluence/architecture/GAME_OUTBOX.md) |
| Patroni HA roadmap | [`deployments/patroni/README.md`](deployments/patroni/README.md) |
| Issues & fixes (v1.0.8) | [`confluence/history/2026-08/30.08.2026/Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md) |
| Что осталось на v1.1 | [`confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) |
| DevOps на потом | [`confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md) |
| Технический долг | [`confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`](confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md) |
| MCP для Cursor | [`services/mcp/README.md`](services/mcp/README.md) |
| Волна 11.10 | [`confluence/history/2026-10/11.10.2026/FINISHERS.md`](confluence/history/2026-10/11.10.2026/FINISHERS.md) |
| CHANGELOG | [`CHANGELOG.md`](CHANGELOG.md) |
| Паттерны (17 сервисов, 333 паттерна) | [`confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md`](confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md) |
| 15th Oct 2026 Roadmap (i18n, tech debt, games) | [`confluence/tech_debt/FUTURE_EVENT_HORIZON.md`](confluence/tech_debt/FUTURE_EVENT_HORIZON.md) |

---

## 🔒 Безопасность и доверие

- **Секреты не коммитятся** — скопируй [`.env.example`](.env.example) → `.env`, задай `JWT_SECRET` и `POSTGRES_PASSWORD`.
- **JWT** — задаётся через `JWT_SECRET`; Auth предупреждает при дефолтном ключе.
- **Пароли в compose** (`eventhorizon`, Patroni stubs) — только для локальной разработки; в k3s — Kubernetes Secrets.
- **OpenAPI** — канонический контракт: [`docs/openapi.yaml`](docs/openapi.yaml). `GET /docs` и `GET /openapi.yaml` отвечают 404, пока не задан `DOCS_PASSWORD`.
- Вопросы по безопасности — Issue или eastwesser@gmail.com.

---

## 🧪 Тестирование

Unit-тесты (Week 2 Clean Architecture + testify):

```bash
task test              # все сервисы
task test-coverage     # покрытие по gRPC-сервисам
task w2-check          # структурный чеклист W2
```

Auth service layer is covered by unit tests with hand-written repository mocks
(`services/auth/internal/service/auth_service_test.go`). Converter layers have
package tests across gRPC services. Mongo inventory test is `//go:build integration`.

When you have network, you can optionally add testify/suite + mockery:

```bash
cd services/auth && go get github.com/stretchr/testify@v1.11.1
# then regenerate mocks via //go:generate on UserRepository
```

Нагрузочное тестирование (k6) — **CORE** (не legacy 500 VU):

```bash
make test-k6                          # browse.js — default ~20 VU
K6_VUS=20 K6_DURATION=60s make test-k6
K6_VUS=5 make test-k6-purchase
```
---

## 🔮 Планы на следующие спринты

Статус на **v1.1.3** (15.10.2026).

Статус на **v1.1.0** (10.10.2026). Ниже — то, что уже в продукте, что осталось лично, и что лежит в парковке.

Живые списки:

- что закрыто и что осталось лично: [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md)
- идеи на потом: [`DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md)

### Уже в продукте

- **v1.0.9** — каталог Берсерк, тонкий list DTO, отмена покупки, отзывы Noiz
- **v1.0.7–v1.0.8** — circuit breaker на gRPC gateway, OpenAPI `/docs`, Payment / Authors / History / Analytics, покупка через NATS, секреты из `.env`
- **v1.0.6** — CI (GitHub Actions), Ansible, манифесты k3s
- **v1.1.0** — публичный `/api/v1`, буст за лампочки, достижения, полировка Hanoi / Gears / Hexagon / Pancaker / компаньон, Grafana, прогон CORE k6

### Инженерный хвост v1.1

Закрыт. Разбор — в `LEFT_TO_FINISH.md`.

### Что ещё не код

- стикеры на Miro
- единый хром мини-игр и арт тамагочи
- витрина на 108 авторов и 3D: нет текстов и ассетов
- смена пароля: отдельный сценарий auth, не недостающий глагол
- WAF и пентест: это контур перед приложением и отчёт человека, не флаг в репозитории
- ротация `JWT_SECRET`

Доля 85/15 пишется в `revenue_splits` при закрытии подписки. Перевода на карту автора нет.

Тег `v1.1.1` стоит на коммите с сагой покупки, билетом сокета, закрытым `/docs` и этой таблицей портов.

### Парковка

Consul, Helm как повседневный деплой, Alertmanager → Telegram, Loki, Tempo, SLO burn. Список — в `DREAM_DEVOPS_IDEAS.md`. Rate limit на gateway уже стоит.

### Если открывать следующий эпик

| Задача | Зачем |
|--------|--------|
| Пагинация каталога на сервере | Имеет смысл до 1000+ карт |
| Уровни лидерборда 2–3 | Сейчас все доски читают `level=1` |
| Порог покрытия в CI | Юнит-тесты есть, порог ≥70% не включён |
| Retry с джиттером, bulkhead | У gateway уже есть circuit breaker |
| i18n (RU/EN/DE/ES) | Смена языка сессии, флаг рядом с бургером |

Подробный roadmap: [`FUTURE_EVENT_HORIZON.md`](confluence/tech_debt/FUTURE_EVENT_HORIZON.md).

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

Текущая: **v1.1.3** (15.10.2026). 

BACKEND PATTERNS REVISION (+ future ideas): event_horizon/confluence/tech_debt/FUTURE_EVENT_HORIZON.md

### Что нового в v1.1.3

Только докуменация и фронтенд-полировка. Бэкенд не менялся.

**Patterns catalog.** 17 сервисов прочитаны, 333 паттерна задокументированы: Facade, Repository, Port, Adapter, Decorator, Strategy, Saga, Outbox, Circuit Breaker, Retry, Rate Limiter, RAG, TF-IDF, Least Connections, WebSocket Ticket, HMAC, SQL Sandbox, и др. Файл: `confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md`.

**Roadmap.** `confluence/tech_debt/FUTURE_EVENT_HORIZON.md` — i18n (RU/EN/DE/ES), tech debt, games polish, users & prod, observability & security.

**Фронтенд v1.1.2 (в составе v1.1.3):**
- BoostCheckbox — per-game эффект буста, `gameId` проброшен.
- Pancaker — космические эмодзи.
- Gears — исправлен рендер канваса (`skinsLoading` в `useEffect`).
- Hanoi — дырки дисков интегрированы в диск.
- Memonia — поле не схлопывается (CSS `aspect-ratio`).
- Flappy — «Нажмите пробел».
- ShopItemCard / ShopItemDetail — `isDigitalItem`, скрытие «В наличии: 998» для скинов и тем.
- `itemIcons` — fallback-иконка из `GAME_ICON`.

### Что нового в v1.1.2

Только фронтенд. Правила, очки и лидерборд те же. Бэкенд и образы сервисов пересобирать не нужно.

**1.1.2.1 Tamagotchi.** Комната — ровный вид сбоку: окно с луной и звёздами за стеклом, кровать, стол. Виды: звезда, кот, дракон, лисичка, зайчик. Серия заботы меняет форму: малыш (1–6 дн.), растёт (7–20), вырос (21+). Сон кладёт на кровать, кормление ставит миску, подарок дня оставляет вещь в комнате.

**1.1.2.2 Builder.** Вид сверху, плиты ходят по двум диагоналям: с северо-запада и с северо-востока. Ровное попадание не режет плиту. Если часть свешивается, этот край падает, а на башне остаётся только нахлёст. Очки считаются как раньше. Камера держит верхушку в кадре. Экран конца партии по центру.

**1.1.2.3 Gears.** Колодец с латунной рамкой, пунктир прицела, лоток следующих шестерней снизу. Две одинаковые вспыхивают при слиянии. Уровни 1–10, опасная линия и физика те же. Подсолнух, Роза и Анютины глазки — скины из магазина, не бесплатные кнопки.

**1.1.2.4 Memonia.** Сетка 6 колонок, найденная пара остаётся на месте. Лица — свой набор Event Horizon, не чужие фрукты. До старта уже лежат рубашки.

**1.1.2.5 Hanoi.** Три деревянных стержня с круглой базой. Диски с верхом, боком и тенью, цифра мелкая. Выбранный диск поднимается. Победа — модалка снизу, башня C остаётся в кадре. Правила те же: 3 стержня, 3–8 колец.

**1.1.2.6 2048.** Светлое поле, плитки теплеют с номиналом, 2048 — золото Horizon. Над полем счёт, рекорд и «Заново». Сохранить — на конце партии. Буст в «Как играть?». Поле 4×4 и спавн 2/4 те же.

**1.1.2.7 Pancaker.** Стопка — слои со сдвигом и своим знаком типа, не эмодзи. Поднос те же фишки. Пустая клетка — тихий контур. Поле в рамке сцены, в шапке уровень и цель.

**1.1.2.8 Flappy.** Пиксельный зверь (крыло вверх / середина / вниз / падение), плоское небо, город, земля, трубы с ободком в палитре Horizon. До старта птица и первая пара труб уже в кадре. Хитбоксы те же.

### Что нового в v1.1.1

- Покупка пишет `purchase.paid` в outbox с настоящим `purchases.id`. Воркер shop и game публикует с `FOR UPDATE SKIP LOCKED`.
- Списание билетиков идёт со стабильным `reference_id` на цикл покупки. Повтор клика не списывает второй раз. Строка `purchase_sagas` дожимает покупку, если процесс умер после списания.
- Повтор отмены снова зовёт `AddCurrency` с ключом `cancel-{id}`. Уникальный ключ не задваивает деньги.
- Сокет `/ws/leaderboard` открывается одноразовым билетом (`POST /api/v1/ws-ticket`), не JWT в query. Origin проверяется.
- `GET /docs` и `GET /openapi.yaml` скрыты, пока не задан `DOCS_PASSWORD`. Swagger UI лежит в репозитории.
- Успешная подписка пишет `revenue_splits`: автору 85%, платформе остаток. Перевода автору нет.
- `GET /api/v1/billing/balance` и `GET /api/v1/billing/transactions` читают id из токена.
- Fulfillment и notification читают Kafka только когда JetStream не поднят.
- Упавший Redis на проверке сессии отдаёт 503, а не «токен жив».
- На дашборде Grafana есть ряд «paid → fulfilled» и счётчик неудачных возвратов.
- Первый `make deploy` после перезагрузки ждёт `pg_isready`, потом goose.

### Что нового в v1.1.0

- **API:** public REST under `/api/v1/*` (legacy `/api/*` rewritten); OpenAPI + FE + k6 aligned  
- **Games / FE:** boost gate, Pancaker score fix (hexagon empty-moves trust), Hanoi multipliers, Gears skins, Companion daily gift, LB level-1 only, profile edit disclosure  
- **Ops:** Grafana dashboard provision fix, Prom alerts expand, `make daily` / `obs-check`, cold-boot healthchecks  
- **Load:** CORE k6 browse/purchase; verdict on 8 GiB VM — 20 VU green ceiling  
- **Docs:** postmortem API v1, DAILY_OPS, SUITABLE_SERVER, DREAM_DEVOPS_IDEAS, interview EH answers  

**10k DAU vs 10k VUs:** planning model ≈ 10k *daily active users* (~30–50 RPS peak), **not** 10 000 concurrent k6 VUs. On an 8 GiB all-in-one compose host, `K6_VUS=10000` will die (RAM + bcrypt/login + PG pools). Use CORE 20–50 VU locally; move load demos to ~8c/32G.

### Что нового в v1.0.9

- **Berserk CCG catalog:** 280 карт засеяно (5 сетов, 4–8), attribute filters (element / rarity / class / flags / stats / icons / artist), URL-shareable state, sort (name / rarity / price / artist / set / element), pagination
- **Purchase + refund:** cancel purchase (`POST /api/shop/purchase/:id/cancel`), refund по цене покупки, idempotent, restore stock, RBAC fix (user allowed on Reserve/Release)
- **Noiz reviews:** «Мнение Noiz» блок на картах 8-го сета (127 rows), `attributes.noiz_review` `{text, rating, verdict, author}`
- **UI polish:** CardImage (5/7, cover/contain), CatalogPager, artist pages + per-artist grid, prev/next navigation, attribute grouping in detail, cancel modal
- **Data:** element normalization (`леса` → `woods`, 56 cards), slim list DTO `InventoryItemsCatalog` (−26% payload)
- **Full review:** [`REVIEW_RESULT.md`](confluence/history/2026-10/03.10.2026/REVIEW_RESULT.md)

### Что нового в v1.0.8

- **Secrets:** `.env.example`, compose/Patroni/k3s via `${…}` (no hardcoded DB passwords in yaml)
- **Thin deploy:** `make deploy` = NATS + obs + fulfillment/notification/analytics; Kafka → `make deploy-heavy`
- **Purchase on NATS:** Shop publishes `purchase.paid`; fulfillment & notification consume JetStream
- **Gateway:** response cache + `_partial` profile; gRPC InvalidArgument → 400; Boosty HMAC webhook
- **Frontend Auth:** password min 8 (match Auth); clearer errors
- **Ops:** game waits for healthy PG; notification `/metrics`; analytics durable names + CH readiness
- **Tooling:** `scripts/rebuild-*.sh`, `docker-push-images.sh`, `coverage-gate.sh`
- Issue log: [`Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md)

### Что нового в v1.0.7

- Payment / Authors / History / Analytics (+ ClickHouse) за Gateway
- Circuit breaker на всех gRPC-клиентах Gateway
- Frontend: подписка, авторы, аналитика, Ханойская башня, merch-gate (403 + `subscription_required`)
- Game **Outbox** для `score.updated` (миграция + worker; пересобери образ `game`)
- MCP/RAG (stdio) для Cursor · OpenAPI sync (auth refresh / whoami / logout)
- Patroni Auth **stubs** · ClickHouse Docker-network fix · compose `--env-file .env`
- Документация: API routes, load resilience, status codes, rebuild scripts

### Что нового в v1.0.6

- Inventory Service — управление мерчем авторов, интеграция с Shop через NATS
- CI/CD — GitHub Actions собирает и пушит образы на Docker Hub
- Ansible — автоматический деплой на сервер
- k3s (Kubernetes) — установлен и настроен кластер
- Makefile — новые команды: docker-build-all, docker-push-all, deploy-k3s, undeploy-k3s
- Delivery Pipeline — полный пайплайн для production
- Документация — обновлены README, добавлены гайды по деплою

### Что нового в v1.0.5

- Магазин (Shop Service) — покупка скинов, инвентарь, интеграция с Billing
- Космические блины — полный маппинг 8 типов для Hexagon
- Memory — скин "Карточки со зверями" (15 видов)
- Flappy Bird — синяя птичка, радужные трубы (сохранена форма)
- Towers — дефолтные красные блоки + радужные скины
- Дата покупки — добавлено поле purchased_at в /api/shop/inventory
- NATS кластер — 3 ноды + мониторинг в Prometheus
- WebSocket — исправлено подключение через балансировщик

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub
🐛 Создай Issue
📬 Напиши мне: eastwesser@gmail.com

Event Horizon — играй, соревнуйся, побеждай! 🚀


======== v.1.1.3.1 ========

# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.3.1-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.3.1** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy. Frontend polish v1.1.2 + backend patterns catalog v1.1.3.

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **v1.1.0 pack** | [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) · [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md) · [`DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md) |
| **Load (measured)** | [`K6_VERDICT.md`](confluence/history/2026-10/10.10.2026/K6_VERDICT.md) · [`SUITABLE_SERVER.md`](confluence/history/2026-10/09.10.2026/SUITABLE_SERVER.md) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · OpenAPI `/docs` только с `DOCS_PASSWORD` (без пароля адрес отвечает 404) |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

### v1.1.3 highlights (15.10.2026)

- **Patterns catalog:** `confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md` — 17 сервисов, 333 паттерна (Facade, Saga, Outbox, Circuit Breaker, RAG, TF-IDF, Least Conn, WebSocket ticket, etc.)
- **Roadmap:** `confluence/tech_debt/FUTURE_EVENT_HORIZON.md` — i18n (RU/EN/DE/ES), tech debt, games polish, users & prod, observability & security
- **Frontend v1.1.2:** BoostCheckbox per-game effects · Gears canvas fix · Pancaker space emojis · Hanoi disk bore · Memonia layout · ShopItemCard `isDigitalItem` (hide 998 stock for skins/themes) · Flappy hint «Нажмите пробел»

### v1.1.0 highlights (07–10.10.2026)

- `/api/v1` public REST · shop/inventory polish · achievements · history NATS  
- Companion daily tickets · avatar upload · nick→LB · record-beaten deep-link  
- Mobile safe-area + game touch targets · Miro stickers pack · security audit RU  
- Ops close: gateway v1 rewrite, Pancaker score trust, Grafana provision, CORE k6 verdict  

### Future decisions (not blockers)

| Track | Status |
|-------|--------|
| Miro board polish / Boosty posts / presentations | **You IRL** — research tomorrow |
| Optional `git tag v1.1.0` | Ask when ready (marker on commit, not a branch) |
| Unified game chrome · Tamagotchi art pass | Soft polish |
| C4 payouts · 108 login authors · full 3D | **Locked** until product OK |
| Loki / Tempo / node-exporter / SLO burn / multi-DB scrape | Dream list — [`DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md) |
| Next hardware for load demos | Prefer **8 vCPU / 32 GiB** — see SUITABLE_SERVER |

---

## 📦 Архитектура (актуально v1.1.0, 09.10.2026)

Схемы: [`EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) · Miro export: [`08.10.2026/miro/`](confluence/history/2026-10/08.10.2026/miro/) · stickers: [`MIRO_STICKERS_PASTE.md`](confluence/history/2026-10/08.10.2026/miro/MIRO_STICKERS_PASTE.md)

```text
[GitHub Actions] ──SSH/Ansible──► [VM] ──docker-compose──► Event Horizon

[React :5173] ──HTTP──► [Balancer :8079] ──HTTP──► [Gateway ×3 :8081–8083]
                                                      │ JWT · HTTP→gRPC
                                                      ▼
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Auth :50051   │ Game :50052   │ Billing:50053 │ Leaderboard   │ Shop :50055   │
│ PG:5460 Redis │ PG:5461       │ PG:5462 Redis │ :50054        │ PG:5465 Redis │
└───────────────┴───────────────┴───────────────┴─PG:5463+R6382─┴───────────────┘
┌───────────────┬───────────────┬───────────────┬───────────────┬───────────────┐
│ Inventory     │ Profile:50060 │ Payment:50058 │ Authors:50061 │ History:50062 │
│ :50059        │ PG:5464       │ sub/merch gate│ PG:5468 Redis │ PG:5469       │
└───────────────┴───────────────┴───────────────┴───────────────┴───────────────┘
┌───────────────┬───────────────────────────────────────────────────────────────┐
│ Analytics     │ NATS JetStream :4222/:4223/:4224  Stream EVENTS (NATS Hub)    │
│ :50057        │ subjects: score.updated, purchase.paid/fulfilled, shop.*, …   │
│ ClickHouse    │ async ──► Profile / Leaderboard / Notification / Fulfillment  │
│ :8123/:9000   │ Leaderboard Redis Sorted Set ──WS──► Client                   │
└───────────────┴───────────────────────────────────────────────────────────────┘

> **Fulfillment** — эмуляция сборки заказа (delay 10s). 
> По факту наличия товара, я просто ищу его, достаю из коробки и упаковываю для отправки.
> Заменится на реальный WMS API в будущем. См. `FUTURE_EVENT_HORIZON.md` → Tech Debt.
```

**Deploy profiles (v1.1.0):** `make deploy` = thin stack (NATS + apps + ClickHouse + Prometheus/Grafana/Jaeger + fulfillment/notification/analytics). Kafka is opt-in: `make deploy-heavy` / `make stop-heavy`. Seeds: `make seed-v110`, `make seed-lb-demo`, `make seed-card-artists`.

---

## 🚀 Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Локальные секреты — скопируй шаблон и задай свои значения
cp .env.example .env
# Отредактируй: JWT_SECRET, POSTGRES_PASSWORD, GRAFANA_ADMIN_PASSWORD (см. .env.example)

# 3. Запустить thin-стек одной командой (NATS path; без Kafka)
make deploy

# 4. Проверить
make status

# 5. Опционально: Kafka broker на более мощной машине
# make deploy-heavy

# 6. Или запустить в k3s
make deploy-k3s
```

Готово! Всё поднимется автоматически:

- PostgreSQL, Redis, NATS, ClickHouse, Jaeger, Prometheus, Grafana (+ apps)
- Миграции баз данных — после `pg_isready` на всех Postgres (первый `make deploy` после перезагрузки машины дожидается баз, затем goose)
- Микросервисы: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Fulfillment, Notification, NATS Hub, …
- Purchase path: Shop → NATS `purchase.paid` → Fulfillment / Notification
- Опционально: Kafka (`make deploy-heavy`), k3s (`make deploy-k3s`)

---

## 🪟 Windows

На Windows проект запускается из Ubuntu в WSL2. Makefile написан на bash.

1. Поставь [Docker Desktop](https://www.docker.com/products/docker-desktop/) и включи **Use the WSL 2 based engine**. В Settings → Resources → WSL integration включи дистрибутив Ubuntu.
2. Поставь Ubuntu из Microsoft Store, если его ещё нет. Клонируй репозиторий в домашний каталог Linux (`~/event_horizon`). Том Postgres на `/mnt/c` работает очень медленно.
3. В Ubuntu: `sudo apt update && sudo apt install -y make golang-go git`. Миграции применяет goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
```

4. `cp .env.example .env`, затем `make deploy`. Первый старт поднимает Postgres, ждёт `pg_isready` и только потом запускает goose и остальные сервисы. Шлюз открывается в браузере Windows: `http://localhost:8079` (WSL2 отдаёт localhost на хост).

Сборка `nats-hub` и миграции идут через `make deploy` внутри Ubuntu. Из PowerShell контейнеры можно поднять отдельно:

```powershell
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d
```

---

## 📍 Эндпоинты (балансировщик :8079)

Канон: **`/api/v1/*`**. Legacy `/api/*` на gateway один раз переписывается в v1. Микросервисы — gRPC (не HTTP). См. [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md).

| Метод | Путь | Описание |
|-------|------|----------|
| POST | /api/v1/auth/register | Регистрация |
| POST | /api/v1/auth/login | Логин (JWT) |
| POST | /api/v1/game/submit | Отправить рекорд |
| POST | /api/v1/game/companion/daily-gift | Tamagotchi: +1000 билетиков / день |
| GET | /api/v1/billing/balance/all | Баланс (лампы/билетики) |
| GET | /api/v1/leaderboard | Топ (публичный) |
| GET | /api/v1/shop/items | Каталог |
| POST | /api/v1/shop/purchase | Купить (билетики) |
| POST | /api/v1/shop/purchase/:id/cancel | Отмена покупки |
| GET | /api/v1/shop/inventory | Инвентарь |
| GET | /api/v1/profile | Профиль (+ ачивки) |
| POST | /api/v1/uploads | Картинка (auth) |
| GET/POST | /api/v1/payment/… | Подписка / merch gate |
| GET/POST | /api/v1/authors/… | Авторы |
| GET | /api/v1/history | История |
| GET | /api/v1/analytics/… | Аналитика (admin) |
| GET | /openapi.yaml · /docs | OpenAPI + Swagger UI, только если задан `DOCS_PASSWORD` |
| WS | /ws/leaderboard | Real-time LB |

| Кто | Как |
|-----|-----|
| **Gateway / curl / OpenAPI** | `/api/v1/shop/purchase` |
| **Frontend** | `baseURL: '/api/v1'` + `api.post('/shop/purchase')` |
| **WebSocket / ops** | `/ws/leaderboard`, `/health`, `/ready` |

---

## 🔧 Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8079/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","nickname":"Test"}'

# Логин (получить токен)
TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# Получить баланс
curl -s -X GET "http://localhost:8079/api/v1/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Посмотреть товары в магазине
curl -s -X GET http://localhost:8079/api/v1/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Купить товар
curl -X POST http://localhost:8079/api/v1/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.'

# Посмотреть инвентарь
curl -s -X GET http://localhost:8079/api/v1/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Отправить рекорд
curl -X POST http://localhost:8079/api/v1/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"user_id":"7fc8a659-1bb2-4d7c-b60e-c140239d5c62","game_id":"hexagon","level":1,"score":150,"user_email":"tuzer@example.com","seed":"test_seed","moves":[]}'

# Посмотреть лидерборд
curl -s "http://localhost:8079/api/v1/leaderboard?game_id=hexagon&limit=10" | jq '.'

# Получить профиль
curl -X GET http://localhost:8079/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

```bash
# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard
```

```javascript
// В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));
```

---

## 🐳 Makefile & Docker-команды

```bash
# Thin stack (NATS; Kafka OFF)
make deploy

# Kafka broker only (apps stay on NATS unless KAFKA_BROKERS set)
make deploy-heavy
make stop-heavy

# Посмотреть логи / статус
make logs
make ps

# Остановить всё / полная очистка volumes
make down
make clean

# Собрать / запушить образы
make docker-build-all
make docker-push-all

# k3s / Ansible
make deploy-k3s
make undeploy-k3s
make delivery-dev
```

### Пересборка после изменений в коде

```bash
bash scripts/rebuild-proto.sh              # protoc для gRPC-сервисов (Gateway без proto)
bash scripts/rebuild-services.sh game analytics gateway
bash scripts/docker-push-images.sh game analytics gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d game analytics gateway gateway-2 gateway-3
```

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3000 | http://localhost:3000 (`GRAFANA_ADMIN_PASSWORD`, default admin) |
| Alertmanager | 9193 | http://localhost:9193 |
| Jaeger | 16686 | http://localhost:16686 |
| NATS Exporter | 7777 | http://localhost:7777/metrics |

Daily maintenance: `make daily` (status + smoke + Prom/Grafana targets).  
Full command kit + Prom/Grafana gap list: [`confluence/history/2026-10/09.10.2026/DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md).

---

## 🧩 Компоненты и порты

### Микросервисы

Порты ниже — те, что проброшены на хост в `deployments/docker-compose.cluster.yml`. Внутри сети контейнеров Postgres слушает 5432, Redis 6379.

| Сервис | gRPC / HTTP | Metrics на хосте | Postgres на хосте | Redis на хосте |
|--------|-------------|------------------|-------------------|----------------|
| Auth | 50051 | 9091 | 5460 `eventhorizon` | 6379 |
| Game | 50052 | 9092 | 5461 `eventhorizon_game` | — |
| Billing | 50053 | 9093 | 5462 `eventhorizon_billing` | 6381 |
| Leaderboard | 50054 | 9094 | 5463 `eventhorizon_leaderboard` | 6382 |
| Shop | 50055 | 9095 | 5465 `eventhorizon_shop` | 6383 |
| Analytics | 50057 | 9106 | ClickHouse 8123 / 9000 | — |
| Fulfillment | — | 9101 | — | — |
| Notification | 50063 | 9102 | 5470 `eventhorizon_notification` | — |
| Payment | 50058 | 9103 | 5467 `eventhorizon_payment` | 6386 |
| Inventory | 50059 | 9096 | 5466 `eventhorizon_inventory` | 6384 |
| Profile | 50060 | 9099 | 5464 `eventhorizon_profile` | 6385 |
| Authors | 50061 | 9104 | 5468 `eventhorizon_authors` | 6387 |
| History | 50062 | 9105 | 5469 `eventhorizon_history` | — |
| Gateway | HTTP 8081, 8082, 8083 | не проброшен | — | 6379 (общий с Auth) |
| Balancer | HTTP 8079 | 9098 | — | — |

Gateway слушает метрики внутри контейнера (`9095`, `9096`, `9097`). На хост они не выведены: `9095` и `9096` на хосте — это Shop и Inventory, `9097` — NATS Hub. Prometheus ходит по именам `gateway:9095`, `shop:9095` внутри docker-сети, это разные контейнеры. Redis `6380` (`redis-game`) поднят, процесс Game к нему не подключён. Mongo в этом compose нет: Inventory пишет в Postgres.

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| NATS-1 | 4222 клиент, 8222 мониторинг, 6222 кластер | Узел |
| NATS-2 | 4223, 8223, 6223 | Узел |
| NATS-3 | 4224, 8224, 6224 | Узел |
| NATS Hub | 9097 | Создаёт Stream EVENTS, свои метрики |
| NATS Exporter | 7777 | `/metrics` |
| ClickHouse | 8123 HTTP, 9000 native | Analytics OLAP |
| Jaeger | 16686 UI, 4317 OTLP gRPC, 4318 OTLP HTTP | Трассировка |
| Prometheus | 9090 | Метрики |
| Grafana | 3000 | Дашборды |
| Alertmanager | 9193 → контейнер 9093 | Алерты |
| Kafka (профиль heavy) | 19092 → контейнер 9092 | Не входит в `make deploy` |
| Postgres exporter | 9187 | Метрики auth-БД |
| Redis exporter | 9121 | Метрики Redis |
| Qdrant (профиль qdrant) | 6333 | Не входит в обычный подъём |

---

## 🎮 Игры

| Игра | Описание | Скины |
|------|----------|-------|
| Flappy Bird | Лети и не врезайся в трубы | Золотая птичка, Радужные трубы |
| Pancaker (Hexagon) | Гексагональный пазл с блинами | Космические блины |
| Builder (Towers) | Строй башню из падающих блоков | Радужные блоки |
| Memonia (Memory) | Найди пары фруктов | Карточки со зверями |
| Hanoi | Классика: 3 стержня, кольца 3–8 | — |
| 2048 | Собирай плитки до 2048 | — |
| Gears | Собирай шестерни на поле | — |
| Tamagotchi | Корми спутника и забирай дневной подарок | — |

---

## 📚 Документация

| Тема | Путь |
|------|------|
| Архитектура и схемы | [`confluence/architecture/EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) |
| HTTP API / RBAC | [`confluence/architecture/API_ROUTES.md`](confluence/architecture/API_ROUTES.md) |
| HTTP status codes | [`confluence/architecture/STATUS_CODES.md`](confluence/architecture/STATUS_CODES.md) |
| Load resilience | [`confluence/architecture/LOAD_RESILIENCE.md`](confluence/architecture/LOAD_RESILIENCE.md) |
| Game Outbox | [`confluence/architecture/GAME_OUTBOX.md`](confluence/architecture/GAME_OUTBOX.md) |
| Patroni HA roadmap | [`deployments/patroni/README.md`](deployments/patroni/README.md) |
| Issues & fixes (v1.0.8) | [`confluence/history/2026-08/30.08.2026/Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md) |
| Что осталось на v1.1 | [`confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) |
| DevOps на потом | [`confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md) |
| Технический долг | [`confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md`](confluence/tech_debt/CURRENT_DEBT/STILL_TECH_DEBT.md) |
| MCP для Cursor | [`services/mcp/README.md`](services/mcp/README.md) |
| Волна 11.10 | [`confluence/history/2026-10/11.10.2026/FINISHERS.md`](confluence/history/2026-10/11.10.2026/FINISHERS.md) |
| CHANGELOG | [`CHANGELOG.md`](CHANGELOG.md) |
| Паттерны (17 сервисов, 333 паттерна) | [`confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md`](confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md) |
| 15th Oct 2026 Roadmap (i18n, tech debt, games) | [`confluence/tech_debt/FUTURE_EVENT_HORIZON.md`](confluence/tech_debt/FUTURE_EVENT_HORIZON.md) |

---

## 🔒 Безопасность и доверие

- **Секреты не коммитятся** — скопируй [`.env.example`](.env.example) → `.env`, задай `JWT_SECRET` и `POSTGRES_PASSWORD`.
- **JWT** — задаётся через `JWT_SECRET`; Auth предупреждает при дефолтном ключе.
- **Пароли в compose** (`eventhorizon`, Patroni stubs) — только для локальной разработки; в k3s — Kubernetes Secrets.
- **OpenAPI** — канонический контракт: [`docs/openapi.yaml`](docs/openapi.yaml). `GET /docs` и `GET /openapi.yaml` отвечают 404, пока не задан `DOCS_PASSWORD`.
- Вопросы по безопасности — Issue или eastwesser@gmail.com.

---

## 🧪 Тестирование

Unit-тесты (Week 2 Clean Architecture + testify):

```bash
task test              # все сервисы
task test-coverage     # покрытие по gRPC-сервисам
task w2-check          # структурный чеклист W2
```

Auth service layer is covered by unit tests with hand-written repository mocks
(`services/auth/internal/service/auth_service_test.go`). Converter layers have
package tests across gRPC services. Mongo inventory test is `//go:build integration`.

When you have network, you can optionally add testify/suite + mockery:

```bash
cd services/auth && go get github.com/stretchr/testify@v1.11.1
# then regenerate mocks via //go:generate on UserRepository
```

Нагрузочное тестирование (k6) — **CORE** (не legacy 500 VU):

```bash
make test-k6                          # browse.js — default ~20 VU
K6_VUS=20 K6_DURATION=60s make test-k6
K6_VUS=5 make test-k6-purchase
```
---

## 🔮 Планы на следующие спринты

Статус на **v1.1.3** (15.10.2026).

Статус на **v1.1.0** (10.10.2026). Ниже — то, что уже в продукте, что осталось лично, и что лежит в парковке.

Живые списки:

- что закрыто и что осталось лично: [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md)
- идеи на потом: [`DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md)

### Уже в продукте

- **v1.0.9** — каталог Берсерк, тонкий list DTO, отмена покупки, отзывы Noiz
- **v1.0.7–v1.0.8** — circuit breaker на gRPC gateway, OpenAPI `/docs`, Payment / Authors / History / Analytics, покупка через NATS, секреты из `.env`
- **v1.0.6** — CI (GitHub Actions), Ansible, манифесты k3s
- **v1.1.0** — публичный `/api/v1`, буст за лампочки, достижения, полировка Hanoi / Gears / Hexagon / Pancaker / компаньон, Grafana, прогон CORE k6

### Инженерный хвост v1.1

Закрыт. Разбор — в `LEFT_TO_FINISH.md`.

### Что ещё не код

- стикеры на Miro
- единый хром мини-игр и арт тамагочи
- витрина на 108 авторов и 3D: нет текстов и ассетов
- смена пароля: отдельный сценарий auth, не недостающий глагол
- WAF и пентест: это контур перед приложением и отчёт человека, не флаг в репозитории
- ротация `JWT_SECRET`

Доля 85/15 пишется в `revenue_splits` при закрытии подписки. Перевода на карту автора нет.

Тег `v1.1.1` стоит на коммите с сагой покупки, билетом сокета, закрытым `/docs` и этой таблицей портов.

### Парковка

Consul, Helm как повседневный деплой, Alertmanager → Telegram, Loki, Tempo, SLO burn. Список — в `DREAM_DEVOPS_IDEAS.md`. Rate limit на gateway уже стоит.

### Если открывать следующий эпик

| Задача | Зачем |
|--------|--------|
| Пагинация каталога на сервере | Имеет смысл до 1000+ карт |
| Уровни лидерборда 2–3 | Сейчас все доски читают `level=1` |
| Порог покрытия в CI | Юнит-тесты есть, порог ≥70% не включён |
| Retry с джиттером, bulkhead | У gateway уже есть circuit breaker |
| i18n (RU/EN/DE/ES) | Смена языка сессии, флаг рядом с бургером |

Подробный roadmap: [`FUTURE_EVENT_HORIZON.md`](confluence/tech_debt/FUTURE_EVENT_HORIZON.md).

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

Текущая: **v1.1.3.1** (03.11.2026).

BACKEND PATTERNS REVISION (+ future ideas): event_horizon/confluence/tech_debt/FUTURE_EVENT_HORIZON.md

### Что нового в v1.1.3.1

Патч-релиз, сфокусированный на критических исправлениях UX в кабинете автора, стабильности редактирования карточек ККИ и безопасности бэкенда.

**Фронтенд**: Кабинет автора и Навигация

- AppNavbar — пункт меню «Автор» переименован в «Кабинет автора» для устранения путаницы с общей страницей «Авторы». Улучшена человеко-ориентированность UI.

- Modal — исправлен конфликт ширины (max-w-lg перебивал max-w-3xl). Добавлен проп align для позиционирования. Заголовок теперь зафиксирован, скроллится только тело формы (решает проблему «уезжающей» шапки при длинных формах).

- InventoryEditModal — починен рендеринг полей ККИ (isCard). Если бэкенд присылал кривой item.type, форма теперь распознает карточку по маркерам в attributes (market_rub, element).

- InventoryEditModal — починено сохранение текста (card_text). Ранее текст терялся из-за неверного условия при формировании attributes.

- InventoryEditModal — починен пересчет цены. Убрана путаница между price и market_rub, цена в билетиках теперь корректно пересчитывается и отправляется на бэк.

- InventoryEditModal — удален дублирующийся вызов updateItem, из-за которого при сохранении улетало два запроса и возникал конфликт версий.

- inventoryApi — добавлено поле version в UpdateItemRequest. Включает оптимистичную блокировку (защита от перезаписи устаревшими данными).

**Бэкенд**: Inventory

- grpc_handler.go — добавлен validateItem. Теперь бэкенд отклоняет создание и обновление товаров с отрицательными price, stock и отрицательными значениями в attributes (закрывает баг с «Совой колдуна» и market_rub: -1).

#### Зафиксировано в техдолг (P0):

IDOR в Inventory — на бэкенде отсутствует проверка author_id при вызове UpdateItem/DeleteItem. Фронт скрывает кнопки, но нужна проверка на уровне gRPC.

### Что нового в v1.1.3

Только докуменация и фронтенд-полировка. Бэкенд не менялся.

**Patterns catalog.** 17 сервисов прочитаны, 333 паттерна задокументированы: Facade, Repository, Port, Adapter, Decorator, Strategy, Saga, Outbox, Circuit Breaker, Retry, Rate Limiter, RAG, TF-IDF, Least Connections, WebSocket Ticket, HMAC, SQL Sandbox, и др. Файл: `confluence/architecture/EVENT_HORIZON_PATTERNS_15.10.2026.md`.

**Roadmap.** `confluence/tech_debt/FUTURE_EVENT_HORIZON.md` — i18n (RU/EN/DE/ES), tech debt, games polish, users & prod, observability & security.

**Фронтенд v1.1.2 (в составе v1.1.3):**
- BoostCheckbox — per-game эффект буста, `gameId` проброшен.
- Pancaker — космические эмодзи.
- Gears — исправлен рендер канваса (`skinsLoading` в `useEffect`).
- Hanoi — дырки дисков интегрированы в диск.
- Memonia — поле не схлопывается (CSS `aspect-ratio`).
- Flappy — «Нажмите пробел».
- ShopItemCard / ShopItemDetail — `isDigitalItem`, скрытие «В наличии: 998» для скинов и тем.
- `itemIcons` — fallback-иконка из `GAME_ICON`.

### Что нового в v1.1.2

Только фронтенд. Правила, очки и лидерборд те же. Бэкенд и образы сервисов пересобирать не нужно.

**1.1.2.1 Tamagotchi.** Комната — ровный вид сбоку: окно с луной и звёздами за стеклом, кровать, стол. Виды: звезда, кот, дракон, лисичка, зайчик. Серия заботы меняет форму: малыш (1–6 дн.), растёт (7–20), вырос (21+). Сон кладёт на кровать, кормление ставит миску, подарок дня оставляет вещь в комнате.

**1.1.2.2 Builder.** Вид сверху, плиты ходят по двум диагоналям: с северо-запада и с северо-востока. Ровное попадание не режет плиту. Если часть свешивается, этот край падает, а на башне остаётся только нахлёст. Очки считаются как раньше. Камера держит верхушку в кадре. Экран конца партии по центру.

**1.1.2.3 Gears.** Колодец с латунной рамкой, пунктир прицела, лоток следующих шестерней снизу. Две одинаковые вспыхивают при слиянии. Уровни 1–10, опасная линия и физика те же. Подсолнух, Роза и Анютины глазки — скины из магазина, не бесплатные кнопки.

**1.1.2.4 Memonia.** Сетка 6 колонок, найденная пара остаётся на месте. Лица — свой набор Event Horizon, не чужие фрукты. До старта уже лежат рубашки.

**1.1.2.5 Hanoi.** Три деревянных стержня с круглой базой. Диски с верхом, боком и тенью, цифра мелкая. Выбранный диск поднимается. Победа — модалка снизу, башня C остаётся в кадре. Правила те же: 3 стержня, 3–8 колец.

**1.1.2.6 2048.** Светлое поле, плитки теплеют с номиналом, 2048 — золото Horizon. Над полем счёт, рекорд и «Заново». Сохранить — на конце партии. Буст в «Как играть?». Поле 4×4 и спавн 2/4 те же.

**1.1.2.7 Pancaker.** Стопка — слои со сдвигом и своим знаком типа, не эмодзи. Поднос те же фишки. Пустая клетка — тихий контур. Поле в рамке сцены, в шапке уровень и цель.

**1.1.2.8 Flappy.** Пиксельный зверь (крыло вверх / середина / вниз / падение), плоское небо, город, земля, трубы с ободком в палитре Horizon. До старта птица и первая пара труб уже в кадре. Хитбоксы те же.

### Что нового в v1.1.1

- Покупка пишет `purchase.paid` в outbox с настоящим `purchases.id`. Воркер shop и game публикует с `FOR UPDATE SKIP LOCKED`.
- Списание билетиков идёт со стабильным `reference_id` на цикл покупки. Повтор клика не списывает второй раз. Строка `purchase_sagas` дожимает покупку, если процесс умер после списания.
- Повтор отмены снова зовёт `AddCurrency` с ключом `cancel-{id}`. Уникальный ключ не задваивает деньги.
- Сокет `/ws/leaderboard` открывается одноразовым билетом (`POST /api/v1/ws-ticket`), не JWT в query. Origin проверяется.
- `GET /docs` и `GET /openapi.yaml` скрыты, пока не задан `DOCS_PASSWORD`. Swagger UI лежит в репозитории.
- Успешная подписка пишет `revenue_splits`: автору 85%, платформе остаток. Перевода автору нет.
- `GET /api/v1/billing/balance` и `GET /api/v1/billing/transactions` читают id из токена.
- Fulfillment и notification читают Kafka только когда JetStream не поднят.
- Упавший Redis на проверке сессии отдаёт 503, а не «токен жив».
- На дашборде Grafana есть ряд «paid → fulfilled» и счётчик неудачных возвратов.
- Первый `make deploy` после перезагрузки ждёт `pg_isready`, потом goose.

### Что нового в v1.1.0

- **API:** public REST under `/api/v1/*` (legacy `/api/*` rewritten); OpenAPI + FE + k6 aligned  
- **Games / FE:** boost gate, Pancaker score fix (hexagon empty-moves trust), Hanoi multipliers, Gears skins, Companion daily gift, LB level-1 only, profile edit disclosure  
- **Ops:** Grafana dashboard provision fix, Prom alerts expand, `make daily` / `obs-check`, cold-boot healthchecks  
- **Load:** CORE k6 browse/purchase; verdict on 8 GiB VM — 20 VU green ceiling  
- **Docs:** postmortem API v1, DAILY_OPS, SUITABLE_SERVER, DREAM_DEVOPS_IDEAS, interview EH answers  

**10k DAU vs 10k VUs:** planning model ≈ 10k *daily active users* (~30–50 RPS peak), **not** 10 000 concurrent k6 VUs. On an 8 GiB all-in-one compose host, `K6_VUS=10000` will die (RAM + bcrypt/login + PG pools). Use CORE 20–50 VU locally; move load demos to ~8c/32G.

### Что нового в v1.0.9

- **Berserk CCG catalog:** 280 карт засеяно (5 сетов, 4–8), attribute filters (element / rarity / class / flags / stats / icons / artist), URL-shareable state, sort (name / rarity / price / artist / set / element), pagination
- **Purchase + refund:** cancel purchase (`POST /api/shop/purchase/:id/cancel`), refund по цене покупки, idempotent, restore stock, RBAC fix (user allowed on Reserve/Release)
- **Noiz reviews:** «Мнение Noiz» блок на картах 8-го сета (127 rows), `attributes.noiz_review` `{text, rating, verdict, author}`
- **UI polish:** CardImage (5/7, cover/contain), CatalogPager, artist pages + per-artist grid, prev/next navigation, attribute grouping in detail, cancel modal
- **Data:** element normalization (`леса` → `woods`, 56 cards), slim list DTO `InventoryItemsCatalog` (−26% payload)
- **Full review:** [`REVIEW_RESULT.md`](confluence/history/2026-10/03.10.2026/REVIEW_RESULT.md)

### Что нового в v1.0.8

- **Secrets:** `.env.example`, compose/Patroni/k3s via `${…}` (no hardcoded DB passwords in yaml)
- **Thin deploy:** `make deploy` = NATS + obs + fulfillment/notification/analytics; Kafka → `make deploy-heavy`
- **Purchase on NATS:** Shop publishes `purchase.paid`; fulfillment & notification consume JetStream
- **Gateway:** response cache + `_partial` profile; gRPC InvalidArgument → 400; Boosty HMAC webhook
- **Frontend Auth:** password min 8 (match Auth); clearer errors
- **Ops:** game waits for healthy PG; notification `/metrics`; analytics durable names + CH readiness
- **Tooling:** `scripts/rebuild-*.sh`, `docker-push-images.sh`, `coverage-gate.sh`
- Issue log: [`Issues.md`](confluence/history/2026-08/30.08.2026/Issues.md)

### Что нового в v1.0.7

- Payment / Authors / History / Analytics (+ ClickHouse) за Gateway
- Circuit breaker на всех gRPC-клиентах Gateway
- Frontend: подписка, авторы, аналитика, Ханойская башня, merch-gate (403 + `subscription_required`)
- Game **Outbox** для `score.updated` (миграция + worker; пересобери образ `game`)
- MCP/RAG (stdio) для Cursor · OpenAPI sync (auth refresh / whoami / logout)
- Patroni Auth **stubs** · ClickHouse Docker-network fix · compose `--env-file .env`
- Документация: API routes, load resilience, status codes, rebuild scripts

### Что нового в v1.0.6

- Inventory Service — управление мерчем авторов, интеграция с Shop через NATS
- CI/CD — GitHub Actions собирает и пушит образы на Docker Hub
- Ansible — автоматический деплой на сервер
- k3s (Kubernetes) — установлен и настроен кластер
- Makefile — новые команды: docker-build-all, docker-push-all, deploy-k3s, undeploy-k3s
- Delivery Pipeline — полный пайплайн для production
- Документация — обновлены README, добавлены гайды по деплою

### Что нового в v1.0.5

- Магазин (Shop Service) — покупка скинов, инвентарь, интеграция с Billing
- Космические блины — полный маппинг 8 типов для Hexagon
- Memory — скин "Карточки со зверями" (15 видов)
- Flappy Bird — синяя птичка, радужные трубы (сохранена форма)
- Towers — дефолтные красные блоки + радужные скины
- Дата покупки — добавлено поле purchased_at в /api/shop/inventory
- NATS кластер — 3 ноды + мониторинг в Prometheus
- WebSocket — исправлено подключение через балансировщик

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub
🐛 Создай Issue
📬 Напиши мне: eastwesser@gmail.com

Event Horizon — играй, соревнуйся, побеждай! 🚀


==================================== FROM 1.1.3.2 and so on I will just use the current version changes

### Что нового в v1.1.3.2

Патч-релиз, закрывающий сквозные баги синхронизации профиля, истории и лидерборда.
Внедрена event-driven синхронизация ника и авторасчет заработка в кабинете автора.

**Фронтенд**: История и Кабинет автора

- HistoryPage.tsx — переработан дизайн. Сырой JSON заменен на человекочитаемые карточки: иконки игр, результаты рекордов, суммы покупок. Добавлен безопасный парсинг (safeParse), чтобы битые события не роняли страницу.

- AuthorSales.tsx — создан новый компонент дашборда продаж для автора. Выводит количество продаж, заработок автора (85%), а также таблицу последних покупок с маскировкой email покупателей.

- AuthorDashboard.tsx — очищен от дублирующегося кода загрузки продаж, теперь использует вынесенный компонент AuthorSales.

**Бэкенд**: Gateway

- game/submit — исправлен баг с отсутствием ника в лидерборде. Теперь Gateway при отправке рекорда самостоятельно достает актуальный ник из Auth (GetUser), а не полагается на фронтенд. Добавлен fallback на email.

Это гарантирует, что для любой игры ник корректно улетает в событие score.updated.

**Бэкенд**: Profile (Синхронизация ника)

- app.go (di.go) — добавлена подписка на событие event.user.nickname.updated через NATS JetStream (Durable Consumer profile-nickname-updated).

- profile_service.go — добавлен метод UpdateNickname, который обновляет только поле ника в базе Profile и инвалидирует Redis-кэш, не затирая очки и ачивки.

Теперь при смене ника в Auth, Profile автоматически синхронизируется, и эндпоинт /profile отдает актуальные данные.

**Бэкенд**: Leaderboard

- redis_repo.go — код сохранения nickname в Redis (хэш ...:info) проверен и оставлен без изменений, так как проблема крылась в отсутствии ника в самом событии от Gateway.

**Бэкенд**: History

- workers.go — исправлена критическая ошибка подписки на NATS. Ранее при попытке создать Durable-консьюмер с точкой в имени (например, history-score.updated) NATS выдавал invalid consumer name, и события не записывались. Теперь точки заменяются на подчеркивания (strings.ReplaceAll).


### Что нового в v1.1.3.3

1. Админка: выдача и отзыв подписки
  - Добавлен API POST /api/v1/admin/payments/grant-subscription (галочка) и revoke-subscription (крестик).
  - В AdminPage.tsx добавлены кнопки подписки (зелёная галочка / красный крестик) под селектом ролей.
  - В payment_service.go добавлены методы GrantSubscription и RevokeSubscription, в postgres_repo.go — CreateSubscription и RevokeSubscription (с погашением старой подписки).

2. Критические баги с начислением валюты
  - Billing: награды не начислялись при no response from stream. Исправлено: добавлен ретрай с пересозданием стрима (forceStreamRecreation) при «залипшем» лидере NATS. Теперь balance.updated публикуется успешно.
  - Анти-спам в Billing. Добавлен SetNX в Redis с ключом score_lock:{user_id}:{game_id}:{score} и TTL 30 секунд. Повторное сохранение того же счета в течение 30 секунд игнорируется.
  - Двойное начисление при redelivery. reference_id для наград теперь берётся из event.Timestamp, а не из time.Now().UnixNano().

3. История (History)
  - Починен Durable-консьюмер: nats.Durable("history-score_updated") (точки заменены на подчёркивания). События снова пишутся в базу.
  - Фронтенд (HistoryPage.tsx) переведен на красивые карточки: иконки игр, результаты рекордов, суммы покупок. Убран сырой JSON.

4. Лидерборд: ники вместо UUID
  - Gateway теперь при POST /game/submit самостоятельно достаёт nickname из Auth (GetUser) и передаёт в Game. Это гарантирует, что для любой игры ник улетает в score.updated.
  - В Leaderboard (Redis) ники сохраняются и корректно отображаются на фронте.

5. Кабинет автора: Дашборд продаж
  - Создан компонент AuthorSales.tsx. Выводит количество продаж, заработок автора (85%) и таблицу покупок с маскировкой email покупателей.
  - AuthorDashboard.tsx очищен от дублирующегося кода.

6. Профиль: синхронизация ника через NATS
  - При смене ника в Auth Gateway публикует event.user.nickname.updated.
  - Profile ловит это событие через NATS (Durable Consumer profile-nickname-updated) и обновляет свою базу и кэш.
  - Теперь /profile всегда отдаёт актуальный ник.

7. Фиксы UI
  - Modal.tsx: исправлен конфликт max-w-lg / max-w-3xl, добавлен фиксированный заголовок и скролл только для тела.
  - HanoiTower.tsx: убран align="bottom" и max-w-sm, модалка победы больше не уезжает.
  - InventoryItemDetail.tsx: исправлен рендер [object Object] в атрибутах (иконки, массивы).
  - Кнопки «Сохранить рекорд» в играх выровнены по центру.

==============

Текущая: **v1.1.3.4** (07.11.2026).

BACKEND PATTERNS REVISION (+ future ideas): event_horizon/confluence/tech_debt/FUTURE_EVENT_HORIZON.md

### Что нового в v1.1.3.4

- **P0-8 Inventory IDOR closed:** `UpdateItem` / `DeleteItem` / `SoftDeleteItem` / `RestoreItem` проверяют `author_id` на бэкенде (`checkAuthorAccess` + `x-user-id` из metadata). Чужой юзер — `403`, админ — без изменений.
- **P0-6 Authors soft-delete:** `RevertApplication` больше не делает `DELETE FROM authors`, а ставит `active = false`. Межбазовые ссылки (`inventory.author_id`, `payment.revenue_splits`, `shop.purchases`) остаются валидными. Новый gRPC `GetMyAuthor` отделяет `/authors/me` (200 всегда, включая inactive) от `/authors/:user_id` (404 для soft-deleted).
- **P0-5 Billing reference_id:** награды за игру берут `reference_id` из `event.Timestamp` вместо `time.Now().UnixNano()` — redelivery не задваивает. Outbox worker: `forceStreamRecreation` для залипшего лидера NATS.
- **Test hygiene:** `make test-integration` зелёный. Authors: `TestRevertApplication_SoftDeletesAuthor` (testcontainers). Inventory: Mongo-тесты `t.Skipf`, если Mongo нет.

DETAILS:

### Что нового в v1.1.3.4

Патч-релиз по безопасности и надёжности. Закрывает два P0 из аудита и один P0 по деньгам.

**Безопасность: Inventory IDOR (P0-8)**

- `grpc_handler.go`: `checkAuthorAccess(ctx, existing.AuthorID)` вызывается в `UpdateItem`, `DeleteItem`, `SoftDeleteItem`, `RestoreItem`. Сравнение идёт с `x-user-id` из metadata, а не с телом запроса.
- `md.Get("x-user-id")` — единственный источник правды о юзере. Подделать `author_id` в body нельзя.
- Чужой юзер получает `403 {"error":"you can only edit your own items"}`. Админ — без изменений.

**Authors: soft-delete вместо hard delete (P0-6)**

- `repo.RevertApplication`: `DELETE FROM authors` → `UPDATE authors SET active = false, updated_at = now()`. Строка жива, `inventory.author_id` / `payment.revenue_splits` / `shop.purchases` не осиротеют.
- `repo.GetActiveByUserID` (новый) — только `active = true`, для публичного `GET /authors/:user_id`.
- `repo.GetByUserID` — без фильтра, для `/authors/me`.
- `service.GetAuthor` — публичный, только активные. `service.GetMyAuthor` (новый) — включая `active = false`.
- `handler.GetAuthor` → `GetActiveByUserID`, `handler.GetMyAuthor` (новый) → `GetByUserID`.
- `proto`: `rpc GetMyAuthor(GetMyAuthorRequest) returns (GetMyAuthorResponse)`.
- `gateway.go`: `GET /api/v1/authors/me` → `GetMyAuthor`. `GET /api/v1/authors/:user_id` — не трогали, фильтр в сервисе.
- `mapErr`: `AlreadyReviewed` → `FailedPrecondition` (409), `AlreadyPrivileged` → `PermissionDenied` (403). Раньше оба были `InvalidArgument` (400).

**Billing: reference_id из события (P0-5)**

- `outbox_worker.go`: `reference_id` для наград — `event.Timestamp`, не `time.Now().UnixNano()`. Redelivery больше не задваивает.
- `forceStreamRecreation`: при `no response from stream` стрим `EVENTS` удаляется и пересоздаётся, чтобы сбросить залипшего лидера NATS. (См. комментарий в `outbox_worker.go`: core-подписка на `balance.updated` глушит `ErrNoResponders`, и `js.Publish` виснет до таймаута.)
- `SetNX` анти-спам: `score_lock:{user_id}:{game_id}:{score}`, TTL 30s.

**Test hygiene**

- `revert_integration_test.go` (Authors): `uuid.NewString()` для `id`/`app_id` (в `authors.id` тип `uuid`, а `"aid-1"` не парсится → `SQLSTATE 22P02`).
- `mongo_repo_test.go` (Inventory): `Ping` + `t.Skipf`, если Mongo нет. Раньше был `nil pointer dereference` при недоступной Mongo.
- `make test-integration` — зелёный по всем 7 сервисам.

**API / docs**

- `docs/openapi.yaml`: `GET /authors/:user_id` — `404` для soft-deleted. `GET /authors/me` — `200` всегда (включая `active = false`).
- `confluence/architecture/API_ROUTES.md`: пометки P0-6 / P0-8.
- `confluence/history/2026-11/07.11.2026/AUTHORS_P0_6.md` — разбор soft-delete.
- `confluence/history/2026-11/07.11.2026/FOREIGN_KEY_DELETE_ISSUES.md` — почему межбазовые ссылки требуют soft-delete.


=====================================================================================

### Что нового в v1.1.3.5

Патч-релиз по UX, карточкам товаров и надёжности. Закрывает P0-6 (soft-delete Authors), P0-8 (IDOR Inventory) и несколько фронтенд-долгов.

**Фронтенд: карточки товаров**

- Стоимость карты ККИ берётся из `attributes.cost` (золото/серебро), а не из `market_rub`.
- Ход карты: если `move = 0` — прочерк, а не «0».
- Класс карты (например, «Компаньон») выводится в карточке товара.
- Бейдж «Карточка» вместо «Товар» — исправлен маппинг в `shopItemMap.ts`.
- Иконки карт — бейджи вместо `[object Object]`.

**Фронтенд: UX и мобилки**

- Убран горизонтальный скролл (`overflow-x-hidden`).
- Исправлены отступы, центрирование модалок и колокольчика.
- Модалка «Подтверждение покупки» — заголовок по центру, кнопки внизу.
- Вкладка «Инвентарь» доступна только авторам и админам.

**Фронтенд: кэш**

- При выходе из аккаунта очищается `shop-storage` и `shop_balance_cache`.
- `catalogPatches` убран из `persist` — галочка «В инвентаре» больше не кэшируется.
- `Owned` сбрасывается в `false` перед установкой (фикс `shop_service.go`).

**Админка**

- Статистика инвентаря: «Уникальных позиций», «Общее кол-во товаров в наличии».
- Тип «Не указан» кликабельный — можно посмотреть список товаров без типа.

**Бэкенд**

- P0-8 (IDOR Inventory): `checkAuthorAccess` проверяет `author_id` на `UpdateItem`, `DeleteItem`, `SoftDeleteItem`, `RestoreItem`.
- P0-6 (Authors soft-delete): `RevertApplication` ставит `active = false`, а не `DELETE`. `GetMyAuthor` отделяет `/me` от публичного `/:user_id`.
- Billing: `reference_id` из `event.Timestamp`, `forceStreamRecreation` для залипшего лидера NATS, анти-спам через Redis `SetNX`.
- History: Durable-консьюмер исправлен (точки → подчёркивания).
- Leaderboard: Gateway тянет ник из Auth.
- Profile: синхронизация ника через NATS.
- Authors: дашборд продаж (85% заработка).

============= v.1.1.3.6 =============

### Что нового в v1.1.3.6

Патч-релиз по балансировщику, инфре и синхронизации OpenAPI.

**Балансер (Least Connections)**

- Health checks бэкендов: горутина раз в 5s пингует `/ready` каждого gateway, флипает `Healthy atomic.Bool`. `getLeastConnBackend` пропускает `unhealthy`.
- Jitter при равных `ActiveConns`: `math/rand/v2` выбирает случайного из равных → последовательная нагрузка распределяется по пулу, а не бьёт в первый.
- Метрики: `balancer_backend_healthy{backend}`, `balancer_backend_active_connections{backend}`.
- `/ready` балансера → 503, если `HealthyBackends() == 0` (было всегда 200).
- `slog` вместо `log.Printf` в hot path; `atomic.LoadInt32` в логах (была гонка чтения).
- `X-Balancer-Backend` переехал из request в response header (было debug-неудобно).
- `503` при отсутствии healthy backends → `Retry-After: 5`.

**Nginx**

- `proxy_next_upstream error timeout http_502` + `proxy_next_upstream_tries 2`: connection error и 502 ретраятся на второй балансер. 503 (CB open) и 504 (timeout после обработки) — не ретраятся.
- `max_fails=0` — не помечаем балансер dead по одному connection error (Nginx и так ретраит через `proxy_next_upstream`).
- Исправлена пропущенная `;` в `location /`.

**Docker**

- `Dockerfile.balancer.bin`: `scratch` → `alpine:3.20` + `wget` — compose healthcheck теперь работает.
- `balancer-2` метрики на хосте `:9190` (`9099` занят Profile, `9098` — `balancer-1`).
- `nginx` `depends_on` — без `condition: service_healthy` (на cold boot блокировал `make deploy`).

**Prometheus**

- `balancer-2:9098` target — оба балансера скрейпятся.

**OpenAPI**

- `services/gateway/api/openapi.yaml` синхронизирован с `docs/openapi.yaml`:
  - `password.minLength: 8` (было 6 — врало против Auth).
  - `enum games` = 8 игр (hexagon, flappy, memory, towers, hanoi, 2048, gears, tamagotchi) — было 4.
  - `include_deleted` в `GET /inventory/items`, `version` в `PUT /inventory/items/{id}`, `idempotency_key` в bulk.
  - `nickname` в `GET /leaderboard` → `entries[]`.
  - `days` в `GET /analytics/dau|mau`, `cohort_days_ago`/`window_days` в `GET /analytics/retention`.
  - `id=all` в `POST /notifications/{id}/read`.
  - `GET /authors/me` — soft-deleted (200), `GET /authors/{user_id}` — 404 для soft-deleted.
  - Добавлены: `POST /admin/shop/purchases/cancel`, `POST /admin/payments/grant-subscription`, `POST /admin/payments/revoke-subscription`.

Патч-релиз по фиксам мержа данных при редактировании (P0-10) и вседозволенности админов при отмене заказов (P0-11). А также закрытие архитектурного бага с `protojson` и починка `Makefile`.

**Инвентарь: мерж полей при `PUT` (P0-10)**

- **`UpdateItem`** (`services/inventory/internal/handler/grpc_handler.go`) теперь **мержит** входящие поля в существующий товар, а не перезаписывает его целиком. Если админ/автор передаёт только `name` или `price`, остальные поля (`description`, `images`, `stock`, `attributes`, `author_id`) сохраняются как были.
- **`protojson` bug** (`PROTOJSON_BUG.md`): выяснилось, что `google.protobuf.Struct` при отсутствии в JSON приходит как `&Struct{}` (пустой), а не `nil`. Из-за этого `attributes` затирались в `{}`. Фикс: `len(req.Attributes.AsMap()) > 0` перед мержем.
- **`version`** теперь явно инициализируется `1` в `CreateItem` и `CreateItemWithOutbox` (совпадает с `DEFAULT 1` в БД). Инкремент `version = version + 1` при апдейте работает корректно (без «прыжков на 2»).
- **Кэш**: `CachedRepository.UpdateItem` теперь делает `DeleteItem` (инвалидацию), а не `warmItem`. Кэш больше не хранит устаревшие данные, если БД их изменила (например, через триггер).

**Магазин: админ отменяет `fulfilled_at` (P0-11)**

- **`AdminCancelPurchase`** (`services/shop/internal/service/shop_service.go`) теперь вызывает `cancelPurchaseInternal(..., isAdmin=true)`, который передаёт `ignoreFulfilled=true` в `RefundPurchase`.
- В `RefundPurchase` (`services/shop/internal/repository/postgres_repo.go`) добавлен флаг `ignoreFulfilled bool`. Для админа проверка `fulfilled_at` пропускается (но проверка окна 7 дней — остаётся). Теперь админ может отменить даже «собранный» товар, игрок — нет.
- Идемпотентность сохранена: повторный `POST` возвращает `already_refunded: true`, `refunded_amount: 0`.

**Инфраструктура и документация**

- **Makefile**: починен `test-smoke` (сломанный цикл `for` из-за лишних пробелов). `make daily` снова работает.
- **Постмортем**: `confluence/postmortem/PROTOJSON_BUG.md` — разбор бага с `google.protobuf.Struct`, чтобы не наступить на него снова в других сервисах.
- **Тесты**: обновлён мок `RefundPurchase` в `shop_service_test.go` (сигнатура с флагом). `make test-all` — зелёный.
- **Nginx**: добавлен `deployments/nginx/nginx.conf` (reverse proxy перед балансерами) — детали в `NGINX_KEYIDEA.md`.
- **Чистый TODO**: агрегированный список задач на 09.11.2026 — в `confluence/history/2026-11/09.11.2026/TODO_LIST_9thNOV2026.md`.
