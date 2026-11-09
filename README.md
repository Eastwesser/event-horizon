# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.3.6-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.3.6** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy. 

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **v1.1.0 pack** | [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) · [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md) · [`DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md) |
| **Load (measured)** | [`K6_VERDICT.md`](confluence/history/2026-10/10.10.2026/K6_VERDICT.md) · [`SUITABLE_SERVER.md`](confluence/history/2026-10/09.10.2026/SUITABLE_SERVER.md) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · OpenAPI `/docs` только с `DOCS_PASSWORD` (без пароля адрес отвечает 404) |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

---

## 📦 Архитектура (актуально v1.1.0, 09.10.2026)

Схемы: [`EH_SCHEMAS.md`](confluence/architecture/EH_SCHEMAS.md) · Miro export: [`08.10.2026/miro/`](confluence/history/2026-10/08.10.2026/miro/) · stickers: [`MIRO_STICKERS_PASTE.md`](confluence/history/2026-10/08.10.2026/miro/MIRO_STICKERS_PASTE.md)

```text
[GitHub Actions] ──SSH/Ansible──► [VM] ──docker-compose──► Event Horizon

[React :5173] -Nginx- ──HTTP──► [Balancer :8079] ──HTTP──► [Gateway ×3 :8081–8083]
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

BACKEND PATTERNS REVISION (+ future ideas): event_horizon/confluence/tech_debt/FUTURE_EVENT_HORIZON.md

### Что нового в v1.1.3.6

Патч-релиз: балансировщик, инфра, синхронизация OpenAPI, P0-10 (Inventory merge), P0-11 (admin cancel fulfilled).

**Балансер (Least Connections)**

- Health checks бэкендов: горутина раз в 5s пингует `/ready` каждого gateway, флипает `Healthy atomic.Bool`. `getLeastConnBackend` пропускает `unhealthy`.
- Jitter при равных `ActiveConns`: `math/rand/v2` выбирает случайного из равных — последовательная нагрузка распределяется по пулу.
- Метрики: `balancer_backend_healthy{backend}`, `balancer_backend_active_connections{backend}`.
- `/ready` балансера → 503, если `HealthyBackends() == 0` (было всегда 200).
- `slog` вместо `log.Printf` в hot path; `atomic.LoadInt32` в логах.
- `X-Balancer-Backend` переехал из request в response header.
- `503` при отсутствии healthy backends → `Retry-After: 5`.

**Nginx**

- `proxy_next_upstream error timeout http_502` + `proxy_next_upstream_tries 2`: connection error и 502 ретраятся на второй балансер. 503 и 504 — не ретраятся.
- `max_fails=0` — не помечаем балансер dead по одному connection error.
- Исправлена пропущенная `;` в `location /`.

**Docker**

- `Dockerfile.balancer.bin`: `scratch` → `alpine:3.20` + `wget` — compose healthcheck работает.
- `balancer-2` метрики на хосте `:9190` (`9099` занят Profile, `9098` — `balancer-1`).
- `nginx` `depends_on` — без `condition: service_healthy` (на cold boot блокировал `make deploy`).

**Prometheus**

- `balancer-2:9098` target — оба балансера скрейпятся.

**OpenAPI**

- `services/gateway/api/openapi.yaml` синхронизирован с `docs/openapi.yaml` (копия для Swagger `//go:embed`).

**Инвентарь: мерж полей при `PUT` (P0-10)**

- `UpdateItem` теперь **мержит** входящие поля в существующий товар, а не перезаписывает целиком. Если передать только `name` или `price`, остальные поля сохраняются.
- `protojson` bug: `google.protobuf.Struct` при отсутствии в JSON приходит как `&Struct{}` (пустой), а не `nil`. Фикс: `len(req.Attributes.AsMap()) > 0` перед мержем.
- `version` явно инициализируется `1` в `CreateItem` и `CreateItemWithOutbox`. Инкремент `version + 1` работает без «прыжков на 2».
- `CachedRepository.UpdateItem` делает `DeleteItem` (инвалидацию), а не `warmItem`.

**Магазин: админ отменяет `fulfilled_at` (P0-11)**

- `AdminCancelPurchase` вызывает `cancelPurchaseInternal(..., isAdmin=true)` → `ignoreFulfilled=true` в `RefundPurchase`. Админ может отменить «собранный» товар, игрок — нет. Окно 7 дней остаётся.
- Идемпотентность сохранена: повторный `POST` → `already_refunded: true`, `refunded_amount: 0`.

**Инфраструктура и документация**

- `Makefile`: починен `test-smoke` (сломанный цикл `for`). `make daily` работает.
- Постмортем: `confluence/postmortem/PROTOJSON_BUG.md`.
- Тесты: обновлён мок `RefundPurchase` в `shop_service_test.go`. `make test-all` — зелёный.
- Nginx: `deployments/nginx/nginx.conf` — детали в `NGINX_KEYIDEA.md`.
- TODO: `confluence/history/2026-11/09.11.2026/TODO_LIST_9thNOV2026.md`.

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

- Also, added Nginx before Load Balancers

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
