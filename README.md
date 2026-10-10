# 🎮 Event Horizon

**Игровая микросервисная платформа** на Go + React: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд (NATS → Redis → WS). Целевая модель нагрузки — **~10k DAU** (см. METRICS / `LOAD_POSTURE`).

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose%20%7C%20k3s-blue.svg)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Release-v1.1.1-brightgreen.svg)](https://github.com/Eastwesser/event-horizon)
[![CI](https://img.shields.io/github/actions/workflow/status/Eastwesser/event-horizon/main.yml?branch=main&label=CI)](https://github.com/Eastwesser/event-horizon/actions)

**Event Horizon v1.1.1** — Clean Architecture per service, gRPC mesh, NATS JetStream, OpenAPI gateway (`/api/v1`), observability (Prometheus / Grafana / Jaeger), React client with mobile-safe shells. One-command local deploy.

| | |
|---|---|
| **Repository** | [github.com/Eastwesser/event-horizon](https://github.com/Eastwesser/event-horizon) |
| **Architecture** | Clean Architecture · Gateway HTTP **`/api/v1/*`** → gRPC |
| **Security** | JWT in Redis, bcrypt cost 12, rate limits, secrets via env |
| **v1.1.0 pack** | [`LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) · [`API_V1.md`](confluence/history/2026-10/08.10.2026/API_V1.md) · [`DAILY_OPS.md`](confluence/history/2026-10/09.10.2026/DAILY_OPS.md) |
| **Load (measured)** | [`K6_VERDICT.md`](confluence/history/2026-10/10.10.2026/K6_VERDICT.md) · [`SUITABLE_SERVER.md`](confluence/history/2026-10/09.10.2026/SUITABLE_SERVER.md) |
| **Docs** | [`confluence/architecture/`](confluence/architecture/) · OpenAPI [`/docs`](http://localhost:8079/docs) when running |
| **Boosty** | [boosty.to/eastwesser](https://boosty.to/eastwesser) |

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
| Что осталось на v1.1 | [`confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md`](confluence/history/2026-10/09.10.2026/LEFT_TO_FINISH.md) |
| DevOps на потом | [`confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md`](confluence/history/2026-10/09.10.2026/DREAM_DEVOPS_IDEAS.md) |
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

### Осталось снаружи кода

- стикеры на Miro
- `git tag v1.1.0`
- единый хром мини-игр
- арт тамагочи по референсу
- выплаты авторам (C4), витрина на 108 авторов и 3D остаются закрытыми

### Парковка

Consul, Helm как повседневный деплой, Alertmanager → Telegram, Loki, Tempo, SLO burn. Список — в `DREAM_DEVOPS_IDEAS.md`.

### Если открывать следующий эпик

Отдельное решение, когда появится время:

| Задача | Зачем |
|--------|--------|
| Пагинация каталога на сервере | Имеет смысл до 1000+ карт |
| Уровни лидерборда 2–3 | Сейчас все доски читают `level=1` |
| Юнит-тесты, порог покрытия | В CI порог ≥70% не включён |
| Retry с джиттером, bulkhead | У gateway уже есть circuit breaker |
| Rate limit | Продуктовый лимит, не баг |
| Emoji → SVG | Когда текущие иконки мешают в интерфейсе |

---

## 🧑‍💻 Команда

Backend & DevOps: Денис Матвеев (Eastwesser)
Архитектура: Микросервисная, событийно-ориентированная
Деплой: Docker Compose → k3s (внедряется)

---

## 📦 Версия

Текущая: **v1.1.1** (10.10.2026)

### Что нового в v1.1.1

- Покупка пишет `purchase.paid` в outbox с настоящим `purchases.id`. Воркер публикует с `FOR UPDATE SKIP LOCKED`.
- Списание билетиков идёт со стабильным `reference_id` на цикл покупки. Повтор клика не списывает второй раз. Упавший возврат дожимает воркер.
- Повтор отмены снова зовёт `AddCurrency` с ключом `cancel-{id}`. Уникальный ключ не задваивает деньги.
- Сокет `/ws/leaderboard` требует токен и свой Origin.
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
