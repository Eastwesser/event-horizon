# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.0-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.0** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy.

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **v1.1.0 pack** | [`09.10.2026/YET_TO_DO_ITEMS.md`](confluence/history/2026-10/09.10.2026/YET_TO_DO_ITEMS.md) · [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · OpenAPI [`/docs`](http://localhost:8079/docs) when running |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

### v1.1.0 highlights (07–09.10.2026)

- `/api/v1` public REST · shop/inventory polish · achievements · history NATS  
- Companion daily tickets · avatar upload · nick→LB · record-beaten deep-link  
- Mobile safe-area + game touch targets · Miro stickers pack · security audit RU  

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
- Миграции баз данных
- Микросервисы: Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Fulfillment, Notification, NATS Hub, …
- Purchase path: Shop → NATS `purchase.paid` → Fulfillment / Notification
- Опционально: Kafka (`make deploy-heavy`), k3s (`make deploy-k3s`)

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
| GET | /openapi.yaml · /docs | OpenAPI + Swagger UI |
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
| Pre-prod debt checklist | [`confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md`](confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md) |
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

**Чеклист (tick boxes):** [`confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md`](confluence/architecture/FINAL_DETAILS.md/TODO_FINAL_PRE_PROD_DETAILS.md)

Классификация — 3 оси
Для каждой задачи:

Сложность: XS / S / M / L / XL.

Риск: низкий / средний / высокий (что может сломаться).

Зависимость: что должно быть до.

🟢 Быстрые победы (XS–S, 1–3 дня)
Здесь и сейчас, минимальный риск.

#	Задача	Сложность	Риск	Зачем
1	PUT 403 investigation	XS	низкий	30 минут. Выяснить причину.
2	Emoji → SVG / PNG в UI	S	низкий	1-2 дня. Заменить ~20 эмодзи в навигации / кнопках / флагах.
3	Rate Limiter (100/сек per user)	S	низкий	Стандартный middleware.
4	Retry с джиттером в gateway	S	низкий	gRPC-клиенты. 1 день.
5	Alerts в Telegram (Alertmanager)	S	низкий	Готовый Alertmanager. 1 день.
6	Circuit Breaker + Bulkhead	S	низкий	У тебя уже есть circuit breaker. Bulkhead — добавить.
Итого: ~5-7 дней, низкий риск, каждая задача закрывается отдельно.

🟡 Средние (M, 1 неделя)
Требуют плана, но предсказуемы.

#	Задача	Сложность	Риск	Зависимость
7	/shop/items thin DTO	M	низкий	—
8	Bottleneck fix — server-side page	M	средний	Нужен bottleneck если 1000+
9	Laмпочки как бусты в играх	M	средний	Game service + UI
10	Уровни сложности (1–20)	M	средний	Game service
11	Достижения (achievements)	M	средний	Новая фича, БД
12	Юнит-тесты ≥70%	M	низкий	Долгая задача, но параллельно
13	OpenAPI/Swagger для всех сервисов	M	низкий	Docs, но нужно по всем сервисам
14	Service Discovery (Consul)	M	средний	Инфраструктура
15	Helm-чарты для k3s	M	средний	DevOps
Итого: ~2-3 недели, средний риск, часть — параллельно.

🔴 Крупные (L–XL)
Требуют декомпозиции на под-волны.

🟠 C. Author registration (XL)
Разбить на 4 под-волны:

C1. Заявка (S).

Route /register-author (auth user).

Форма: имя, портфолио, причина.

POST /api/authors/apply.

Запись в БД: pending.

C2. Admin approval (M).

/admin → вкладка «Заявки».

Список pending, кнопки approve / reject.

Approve → role=author, отправка уведомления.

C3. Author dashboard (L).

/author/dashboard.

Мои карты (list + filter).

Создать карту (уже есть /inventory, но только свои).

Продажи (из shop).

C4. Payouts / analytics (L).

Сколько заработал (лампочки / билетики).

Выплаты.

Требует обсуждения модели монетизации — что автор получает? % от продаж?

Итого: 3-4 недели, высокий риск (новые эндпоинты, БД, роли, монетизация).

🟠 Bottleneck (L)
Зависимость: сделать до того, как каталог вырастет до 1000+.

Server-side pagination / filter / sort на /inventory/items.

Клиент запрашивает по странице, не весь каталог.

Плюс: меньший DTO.

Минус: сложнее UI (URL state).

Итого: 1-2 недели.

🟠 Реальные товары (M)
Заменить placeholder (Ключница Дракон → Берсерк ККИ).

Удалить старые тестовые items.

Оставить только карточки ККИ.

Плюс: Cursor может сделать SQL/script.

Минус: ручное решение — какие товары оставить, какие удалить.

Итого: 2-3 дня.

🟠 Полиш игр (M-L per game)
5 игр, каждую — отдельно.

Flappy — текстуры, звук?

Towers — анимации, GAME OVER.

Hanoi — детали drag.

Memory — флип, скины.

Hexagon — геймплей.

Итого: ~1 неделя на все 5, если делать параллельно.

🟠 Deploy / k3s / Ansible (L)
Уже частично сделано:

✅ CI/CD GitHub Actions.

✅ Ansible.

✅ k3s.

🟧 Helm-чарты.

🟧 Service Discovery.

Итого: 2-3 недели на полный prod-ready.

🟠 Нагрузочное тестирование + оптимизация БД (M-L)
k6 + индексы.

Прогнать сценарии под нагрузкой.

Замерить RPS / latency.

Оптимизировать индексы.

Итого: 1-2 недели.

📋 Единый план — по приоритетам
Волна 1 — Быстрые победы (1-2 недели, XS–S)
Начать сразу, минимум риска.

PUT 403 investigation (XS).

Emoji → SVG (S).

Retry + jitter (S).

Alerts в Telegram (S).

Circuit Breaker + Bulkhead (S).

Rate Limiter (S).

Итого: ~5-7 дней, каждая независима.

Волна 2 — Техдолг + контент (2-3 недели, M)
/shop/items thin DTO (M).

Реальные товары — placeholder → Берсерк (M).

Полиш игр (M).

Laмпочки, уровни (M).

Волна 3 — Крупная фича (3-4 недели, L-XL)
C. Author registration — разбить на под-волны:

C1. Заявка.

C2. Admin approval.

C3. Dashboard.

C4. Payouts / monetization.

Перед C — обсудить монетизацию — что автор получает?

Волна 4 — Инфраструктура (параллельно)
Helm-чарты, Service Discovery (M).

k6 нагрузка + оптимизация БД (M-L).

Юнит-тесты ≥70% (M) — параллельно всему.

Волна 5 — Долгосрочное
Bottleneck fix — перед 1000+ картами.

Observability — уже работает, добавить алерты.

NATS cluster — работает.

🎯 Что я рекомендую начать с
Три задачи — как «разогрев»:

PUT 403 investigation (30 минут).

Emoji → SVG (1-2 дня).

Retry + jitter (1 день).

Почему:

PUT 403 — быстрый вопрос, ты знаешь ответ сразу.

Emoji → SVG — визуально заметно, контент.

Retry + jitter — стабилизирует gateway.

Потом — /shop/items thin DTO + реальные товары — логическое завершение v1.0.9.

Потом — C. Author registration (большая фича).

Что я держу в карте
Закрыто (v1.0.9 на GitHub):

См. предыдущий разбор.

Backlog — единый план:

Волна 1 — XS-S:

🟧 PUT 403 investigation (XS)

🟧 Emoji → SVG (S)

🟧 Retry + jitter (S)

🟧 Alerts Telegram (S)

🟧 Circuit Breaker + Bulkhead (S)

🟧 Rate Limiter (S)

Волна 2 — M:

🟧 /shop/items thin DTO

🟧 Реальные товары

🟧 Полиш игр

🟧 Laмпочки / уровни

Волна 3 — L-XL:

🟧 C. Author registration (4 под-волны)

Волна 4 — инфра:

🟧 Helm, Consul

🟧 k6, индексы

🟧 Юнит-тесты

Волна 5 — долгосрочное:

🟧 Bottleneck fix (перед 1000+)

🟧 Observability alerts

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

Текущая: **v1.0.9** (03.10.2026)

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
