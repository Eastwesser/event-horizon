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

Measured on ~8 GiB VM (10.10.2026): **20 VU browse PASS** (~50 RPS, p95 429 ms, 0% errors); 50 VU still 100% OK but p95 ~960 ms.  
Verdict: [`K6_VERDICT.md`](confluence/history/2026-10/10.10.2026/K6_VERDICT.md).  
Do **not** run 10k VUs on this VM — that is not “10k DAU”; it will OOM / thrash (see below in version notes).

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
