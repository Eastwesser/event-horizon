# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.3.8-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.3.8** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy. 

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

---

## 📦 Архитектура (актуально v1.1.3.7, 11.11.2026)

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

## 📚 Документация (hidden since 11.11.2026)

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
# 1. Полный деплой с новыми restart:
make ride-on

# 2. Гигиена (status + smoke + obs)
make daily

# 3. Unit-тесты
make test-all

# 4. (опционально)
make test-smoke
make test-k6
make test-integration

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

Живые списки:

- что закрыто и что осталось лично: [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md)
- идеи на потом: [`DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md)

### Парковка

| Задача | 
|--------|
- Порог покрытия в CI | Юнит-тесты есть, порог ≥70% не включён |
- Retry с джиттером, bulkhead | У gateway уже есть circuit breaker |
- Уровни лидерборда 2–3 | Сейчас все доски читают `level=1` |
- BILLING FOR GAMES STRICT TICKET-PER-GAME RULES
- единый хром мини-игр и арт тамагочи
- Batch owned, pgbouncer, 6 Gateway — v1.2.
- смена пароля: отдельный сценарий auth, не недостающий глагол
- витрина на 108 авторов и 3D: нет текстов и ассетов
- ротация `JWT_SECRET`
- Про ORDER-сервис Записал в идеи: корзина, заказы, доставка, резерв на время корзины. v1.2+. Не сейчас.
- Пагинация каталога на сервере | Имеет смысл до 1000+ карт |
- i18n (RU/EN/DE/ES) | Смена языка сессии, флаг рядом с бургером |
- API GATEWAY GOD OBJECT DECOMPOSITION
- WAF и пентест: это контур перед приложением и отчёт человека, не флаг в репозитории
- Доля 85/15 пишется в `revenue_splits` при закрытии подписки. Перевода на карту автора нет.
- Consul, Helm как повседневный деплой, Alertmanager → Telegram, Loki, Tempo, SLO burn. Список — в `DREAM_DEVOPS_IDEAS.md`. Rate limit на gateway уже стоит.
- 200ms — цель для P=50 на 8 GiB, не для P=200. P=200 на 8 GiB — ~1–10s browse, это потолок железа, не кода. Для 200ms на P=200 — 32 GiB, pgbouncer, batch owned, 6 Gateway.
- Upload to prod -> "eastwessers-event-horizon.com"

Подробный roadmap: [`FUTURE_EVENT_HORIZON.md`](confluence/tech_debt/FUTURE_EVENT_HORIZON.md). Some tasks has already been done.

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

BACKEND PATTERNS REVISION (+ future ideas): event_horizon/confluence/tech_debt/FUTURE_EVENT_HORIZON.md

### Что нового в v1.1.3.8

Патч-релиз по **Auth**, **Inventory**, **NATS** и **UI карточек**. Закрывает инцидент с bcrypt, восстановление карточек, MERGE атрибутов и стабильность лидерборда.

**Auth**
- **bcrypt `$2y$`** — нормализация `$2y$` → `$2b$` в `Login`. PHP/htpasswd-хеши теперь работают.
- **Тест:** `TestLogin_AcceptsPHP2yHash` (проверяет `$2y$` → успешный логин).

**Inventory**
- **`RestoreItem`** — работает (soft-delete + restore, `deleted_at: NULL`).
- **`UpdateItem` MERGE** — `card_text`, `market_rub`, `images` не затираются (P0-10).
- **Оптимистичный UI** — `deleteItem` убирает карточку мгновенно, при ошибке — возвращает.
- **`GetItemIncludingDeleted`** — для проверки прав в `SoftDelete`/`Restore` (Postgres + Mongo + Cached).

**Gateway**
- **Полные `attributes`** для author/admin в `GET /api/v1/inventory/items` (форма редактирования).
- **`Images=[]`** — защита от затирания (`if len(req.Images) > 0`).

**NATS**
- **`Replicas: 3`** — stream `EVENTS` реплицирован на 3 ноды (`jsz` везде `1`).
- **`--cluster_advertise`** — стабильный кластер (без дублей маршрутов).

**Leaderboard**
- **Ники** отображаются (Renoir, Raya) — `SaveUserInfo` + `UpdateScoreOnly` в Redis.

**UI**
- **Бейдж «Карточка»** — вниз, к стоку (без переносов «Придворны / й»).
- **Кнопка «Нет в наличии»** — не вылезает за рамку (`overflow-hidden`, `flex-wrap`).
- **`line-clamp-2`** — «Придворный чародей» в 2 строки, «Контрабандист» в 1.

**Итог:** Auth-инцидент закрыт, восстановление карточек работает, NATS стабилен, лидерборд показывает ники. Массовый засев 108+ — v1.1.3.9.

### Что нового в v1.1.3.7

Патч-релиз по **highload** **k6**, **deadlock**, **N+1**, **pool**, **warmup**.

**Highload k6 (P=50 → P=200)**
- `10k-probe.js` — 3 сценария (browse 70%, submit 20%, purchase 10%), `PROFILE` env, мягкие thresholds, метрики по ручкам.
- `seed-users.js` + `scripts/k6-seed-and-run.sh` — пул 10 юзеров, balance 100k, JSON, run.
- **P=100** — 6044ms browse, 0% errors, 0% 5xx. **P=150** — 7359ms, 0% errors, 0% 5xx. **P=200** — 14027ms, 4% errors (contention), 0.94% 5xx.

**Shop**
- **Deadlock fix** — `RefundPurchase` lock order items → inventory → purchases (было purchases → inventory → items). `lock_timeout = 2s`.
- **Purchase резерв** — `UPDATE items SET stock = stock - 1 WHERE stock >= 1 RETURNING` вместо `SELECT FOR UPDATE`. Лок микросекунды, не 2s.
- **Batch owned** — `applyOwnedFlags` 2 SQL → 1 SQL (`ListOwnedItemIDsByItemIDs`). `circuit shop` исчез, `purchase_owned` 72% → 5.63%.
- Pool 100/50, `pg_prewarm('items'/'purchases'/'inventory')`.

**Inventory**
- Pool 100/50, `pg_prewarm('inventory_items')`.
- **GIN индексы** — attributes, type, price, author, name trgm. `inventory/items` ×3.4 быстрее.
- `JAEGER_ENDPOINT=jaeger:4317`.

**Gateway**
- **ValidateToken кэш 1s** — `sync.Map`, fast path при hit. ×10 быстрее (ручной тест 680→66ms).
- **Bulkhead** — 256 (shop/inventory), 128 (auth/profile/leaderboard/game), 64 (billing/payment/authors/history/notification), 32 (analytics). `bulkhead full` исчез.

**Nginx**
- `proxy_read_timeout 30s → 5s`, `proxy_next_upstream error timeout` (без `http_502`), `worker_connections 4096`. p99 36s → 8s.

**Pool 11 сервисов**
- auth/profile 100/50, billing/leaderboard 64/32, payment/authors/history/notification 32/16, game 100/50.

**Docs**
- `confluence/postmortem/` — 8GiB_ISSUE, DEADLOCK_ISSUE_#1, LOW_CONNECTION_POOL_ISSUE, COLD_START_AFTER_RESTART, CACHE_P100, UNDERSTANDING_RELIABILITY.
- `confluence/architecture/` — RESERVE_MARKETPLACE_PATTERN, RULES_FOR_PG_POOL, WARMUP_COMMANDS.

**Итог:** P=100 — baseline 8 GiB. P=200 — потолок (contention на 9 items). Для 200ms на P=200 — 32 GiB (roadmap v1.2).

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub
🐛 Создай Issue
📬 Напиши мне: eastwesser@gmail.com

Event Horizon — играй, соревнуйся, побеждай! 🚀
