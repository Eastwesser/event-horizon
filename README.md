# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-v1.0.6-brightgreen.svg)]()

---

## 📦 Архитектура (актуально v1.0.6, 03.08.2026)
```text
[GitHub Actions] (в облаке)
        │
        │ SSH + Ansible
        ▼
[Твоя виртуалка]
        │
        │ запускает
        ▼
[docker-compose up -d]
        │
        ▼
[Event Horizon работает]

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

    ┌──────────────────────────────────────────────────────────────┐
    │               NEW: Inventory Service :50059                  │
    │   Управление мерчем авторов, интеграция с Shop через NATS   │
    └──────────────────────────────────────────────────────────────┘
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

- 30+ Docker-контейнеров (PostgreSQL, Redis, NATS, Jaeger, Prometheus, Grafana)
- Миграции баз данных
- Все микросервисы, включая Profile, Shop, Inventory и NATS Hub
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
| WS | /ws/leaderboard | WebSocket обновления |

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
| Game | 50052 | 9092 | PG 5461 | 6380 |
| Billing | 50053 | 9093 | PG 5462 | 6381 |
| Leaderboard | 50054 | 9094 | PG 5463 | 6382 |
| Profile | 50060 | 9099 | PG 5464 | — |
| Shop | 50055 | 9095 | PG 5465 | 6383 |
| Inventory | 50059 | 9096 | PG 5465 | 6383 |
| Gateway | HTTP 8081-8083 | 9095-9097 | — | — |
| Balancer | HTTP 8079 | 9098 | — | — |

### Инфраструктура

| Сервис | Порт(ы) | Назначение |
|--------|---------|------------|
| NATS-1 | 4222, 8222 | Узел кластера |
| NATS-2 | 4223, 8223 | Узел кластера |
| NATS-3 | 4224, 8224 | Узел кластера |
| NATS Hub | — | Создаёт Stream EVENTS |
| Jaeger UI | 16686 | Трассировка |
| Prometheus | 9090 | Метрики |
| Grafana | 3000 | Дашборды |

---

## 🎮 Игры

| Игра | Описание | Скины |
|------|----------|-------|
| Flappy Bird | Лети и не врезайся в трубы | Золотая птичка, Радужные трубы |
| Hexagon | Гексагональный пазл с блинами | Космические блины |
| Towers | Строй башню из падающих блоков | Радужные блоки |
| Memory | Найди пары фруктов | Карточки со зверями |

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

### 🧩 Новые сервисы (1–2 месяца)

| Сервис | Назначение | Порт (gRPC) |
|--------|------------|-------------|
| Payment | Реальные платежи (Boosty/Stripe) | 50058 |
| Notification | Push, Email, Telegram | 50056 |
| Analytics | DAU, MAU, Retention (ClickHouse) | 50057 |

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

Текущая: v1.0.6 (03.08.2026)

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
