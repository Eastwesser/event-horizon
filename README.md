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
