# 🎮 Event Horizon

**Игровая платформа** с микросервисной архитектурой на Go, real-time leaderboard через NATS и целевой нагрузкой 10k RPS.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-✓-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📦 Архитектура (актуально v1.0.3, 05.07.2026)

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
    ┌─────────────────────────────────────┐
    │        NATS Hub (инфраструктура)    │
    │   ── создаёт Stream `EVENTS` ──►    │
    │      Subjects: score.updated,       │
    │      user.registered, shop.*        │
    └──────────────────┬──────────────────┘
               │
               ▼
          [NATS :4222] — событийная шина
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

| Сервис | Порт | Назначение |
|--------|------|------------|
| **NATS** | `4222` | Событийная шина |
| **NATS Hub** | — | Создаёт Stream EVENTS (инфраструктурный) |
| **NATS мониторинг** | `8222` | JSON-метрики |
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

**Текущая:** `v1.0.3` (05.07.2026)  

---

## ⭐ Если проект полезен

⭐ Поставь звезду на GitHub  
🐛 Создай Issue  
📬 Напиши мне: [eastwesser@gmail.com](mailto:eastwesser@gmail.com)

---

**Event Horizon — играй, соревнуйся, побеждай!** 🚀