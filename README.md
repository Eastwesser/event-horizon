# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-v1.0.7-brightgreen.svg)]()

---

## 📦 Архитектура (актуально v1.0.7, 19.08.2026)

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
│ :50057        │ subjects: score.updated, user.registered, shop.*, …           │
│ ClickHouse    │ async ──► Profile / Leaderboard / Notification / Fulfillment  │
│ :8123/:9000   │ Leaderboard Redis Sorted Set ──WS──► Client                   │
└───────────────┴───────────────────────────────────────────────────────────────┘
```
---

## 🚀 Быстрый старт

# 1. Клонировать
git clone https://github.com/Eastwesser/event-horizon.git
cd event-horizon

# 2. Запустить всё одной командой (Docker Compose)
make deploy

# 3. Проверить
make status

# 4. Или запустить в k3s
make deploy-k3s

Готово! Всё поднимется автоматически:

- 30+ Docker-контейнеров (PostgreSQL, Redis, NATS, ClickHouse, Jaeger, Prometheus, Grafana)
- Миграции баз данных
- Микросервисы: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Fulfillment, Notification, NATS Hub, …
- Мониторинг (Prometheus + Grafana)
- Опционально: k3s кластер

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

---

## 🔌 WebSocket

Подключиться к real-time обновлениям лидерборда:

# Через терминал
wscat -c ws://localhost:8079/ws/leaderboard

# В браузере
const ws = new WebSocket('ws://localhost:5173/ws/leaderboard');
ws.onmessage = (e) => console.log('📩', JSON.parse(e.data));

---

## 🐳 Makefile & Docker-команды

# Запустить всё
make deploy

# Посмотреть логи
make logs

# Статус контейнеров
make ps

# Остановить всё
make down

# Полная очистка (удалить volumes)
make clean

# Собрать все образы
make docker-build-all

# Запушить все образы
make docker-push-all

# Деплой в k3s
make deploy-k3s

# Удалить из k3s
make undeploy-k3s

# Локальный деплой через Ansible
make delivery-dev

---

## 🖥️ Мониторинг

| Сервис | Порт | Доступ |
|--------|------|--------|
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3000 | http://localhost:3000 (admin/admin) |
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

- Архитектура и схемы
- История релизов
- Технический долг
- FAQ
- CHANGELOG.md
- Delivery & Deployment

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

Текущая: v1.0.7 (19.08.2026)

### Что нового в v1.0.7

- Payment / Authors / History / Analytics (+ ClickHouse) за Gateway
- Circuit breaker на всех gRPC-клиентах Gateway
- Frontend: подписка, авторы, аналитика, Ханойская башня, merch-gate
- MCP/RAG (stdio) для Cursor
- OpenAPI: auth refresh / whoami / logout / update-role + новые API

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
