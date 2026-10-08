Первый блок (общие практики и процессы)
1. Тестирование

Какие критичные сценарии вы покрывали unit-, интеграционными и e2e-тестами? Приведите 1–2 примера, где тест действительно поймал регрессию или позволил безопасно изменить сложную логику.

2. Code Review

В code review вы даёте обратную связь по архитектуре и качеству кода. Приведите 2 конкретных примера замечаний, которые повлияли на решение: что было предложено изначально, какой риск вы увидели и что команда изменила после вашего комментария.

3. Декомпозиция задач

Какого масштаба задачи вы сейчас умеете декомпозировать самостоятельно? Приведите пример задачи крупнее 2–3 дней: как разбили её на этапы, оценили риски и зависимости и как фактический срок соотнёсся с оценкой.

4. Ответственность после деплоя

Как устроена ваша ответственность после деплоя: участвуете ли вы в релизах, мониторинге, разборе алертов, откатах и поддержке? Что делаете лично и как часто возникают ситуации, где решение нужно принимать вам?

5. Документация

Какие документы вы создавали для Event Horizon и кто ими пользуется? Что именно вы документировали: архитектуру, API, deployment, мониторинг, troubleshooting? Приведите пример, когда документация помогла другому человеку решить задачу без вашего участия.

Второй блок (глубокая техническая часть)
6. PostgreSQL

Какой у вас самый сильный реальный кейс с PostgreSQL в Event Horizon? Опишите, что делали лично: проектировали таблицы или связи, меняли схему, писали миграции, создавали индексы, оптимизировали запросы, работали с транзакциями/блокировками. Если искали медленный запрос, как нашли причину и что изменилось после оптимизации?

7. NATS/Kafka

Как именно Event Horizon работает с NATS/Кафка и какую часть этой схемы вы проектировали или настраивали сами? Опишите применимые вещи: consumer group, партиции и порядок событий, offsets, повторная доставка, retries/DLQ, идемпотентность, consumer lag, rebalance. Важно указать не теорию, а что реально было в вашем сервисе и какие решения принимали лично вы.

8. API Design (REST/gRPC)

Кроме уже упомянутой, вы проектировали REST/gRPC API. Возьмите один наиболее сложный контракт и опишите, что именно проектировали лично: структуру методов/эндпоинтов и данных, protobuf/Swagger, правила ошибок, timeout/deadline, версионирование и обратную совместимость. Был ли случай, когда нужно было изменить контракт так, чтобы не сломать потребителей?

---

## Ответы (черновик для собеседования — опирается на реальный код EH)

*Контекст: Event Horizon — учебный/portfolio микросервисный проект; я sole developer + владелец продукта. «Команда» = я + ассистенты (Cursor) / ревью паттернов по курсу Kozirev. Формулировки ниже — от первого лица.*

---

### 1. Тестирование

**Что покрываем**

| Уровень | Что критично | Где |
|--------|----------------|-----|
| **Unit** | бизнес-правила shop cancel/refund, billing spend/add, physical vs digital merch, achievements tiers, gateway DTO map | `services/shop/.../shop_service_test.go`, billing/profile tests |
| **Integration** (`-tags=integration`) | Postgres: balance + outbox row после purchase; inventory/shop repos через testcontainers или `*_TEST_DATABASE_URL` | `make test-integration` |
| **e2e / load** | CORE k6: browse catalog + purchase→cancel; не 500 VU login blast | `make test-k6`, `make test-k6-purchase` |

**Пример A — тест поймал опасную логику refund.**  
В unit-тестах `CancelPurchase`: сценарии «already refunded» (нельзя второй раз `AddCurrency`), «refund window expired», «billing refund failed после пометки». Это позволило менять порядок шагов cancel (каталог пропал / orphan purchase) и всё равно гарантировать: либо билетики вернулись ровно один раз, либо явная ошибка — без двойного начисления. Без этих тестов легко сломать идемпотентность отмены.

**Пример B — load/e2e отделил регрессию от capacity.**  
Прогон legacy `loadtest.js` (~500 VU) дал ~99% fail на login. Параллельно shop/inventory/submit/LB оставались green. Разбор: bcrypt cost **12** + Redis rate limit login **5 req/s на IP** (и global ~100 rps) — это ожидаемый потолок, не «auth сломан». CORE suite (`browse.js` 20 VU) после фикса env в Makefile: **100% checks**, p95 &lt; 800 ms. Документировали в `LOAD_RESULTS/` и `SECURITY_BASELINE.md`, чтобы следующий «красный» 500 VU blast не воспринимался как регресс.

---

### 2. Code Review

Sole-dev, поэтому «ревью» = ревью своих/агентных PR против архитектурных правил курса и HARD RULES продукта.

**Пример 1 — cleanup SQL на чужой БД.**  
Предложение: один файл `cleanup-shop-content-v2.sql` гнать в shop Postgres.  
Риск: SECTION A трогает `inventory_items.deleted_at` → на shop DB падает транзакция; SECTION B (скрытие junk в `items`) может не примениться предсказуемо.  
Что изменили: `make seed-shop` / `seed-shop-inventory` режут файл по маркерам SECTION A/B и шлют в **разные** контейнеры (`event-horizon-postgres-inventory` / `-shop`). Berserk `карточка` (~281) остаются под явными UUID-only правилами.

**Пример 2 — FE `/api/api/profile` и history «пустая».**  
Предложение: FE звал `api.get('/api/profile')` при `baseURL: '/api'` → 404. History слушала `user.registered`, gateway публиковал только `event.user.registered`.  
Риск: «сломан профиль/история» в проде при живых сервисах.  
Что изменили: относительный путь `/profile`; dual-publish + history worker нормализует `event.user.registered` → `user.registered` с Nak при ошибке записи. Это закрыло пустой экран истории без ломки profile consumer.

---

### 3. Декомпозиция задач

Умею декомпозировать эпики **недели+** (пример: v1.1.0 polish wave с ticklist на десятки пунктов).

**Крупная задача:** «v1.1.0 polish» (insights 07.10 + ticklist 08.10) — сайт/магазин/игры/mobile/load/docs.

| Этап | Содержание | Риски / зависимости |
|------|------------|---------------------|
| 1 Cleanup | scripts/, FE stores по играм | CI/Makefile refs на Dockerfiles |
| 2 Shop/FE | merch chip, inventory ×qty, tickets в админке | HARD RULE: не трогать Berserk cards |
| 3 History/NATS | dual subject | согласовать с profile consumer |
| 4 Games polish | boost unranked, cosmic skins, Memonia map | не ломать ranked submit |
| 5 Mobile | safe-area, touch targets | после FE cleanup |
| 6 Load | CORE k6 + seeds docker exec | не путать с 500 VU blast |
| 7 Docs/Miro/Boosty | порты, Boosty draft; Miro — IRL | Miro/Boosty нельзя «закоммитить» за меня |

**Оценка vs факт:** code+ops уложились в интенсивный 1–2 дня сессий; runtime (deploy/seeds/k6) — отдельный вечер; Miro/Boosty публикация остаются IRL. Риски (Berserk, пустой `$SHOP_DSN` → unix socket) заранее выписаны в ticklist / `WHAT_TO_DO_IRL.md` — это сократило хаос при seeds.

---

### 4. Ответственность после деплоя

В EH я **единственный**, кто релизит и отвечает за стенд.

- **Релиз:** `make deploy` / `deploy-heavy` (Kafka opt-in), `make status` (health/ready), goose migrations, `make seed-v110`.
- **Мониторинг:** Prometheus / Grafana / Jaeger / Alertmanager в compose; сервисы отдают `/health` и `/ready`.
- **Разбор:** k6 отчёты, логи gateway/auth при login pressure, проверка card count 281 после shop cleanup.
- **Откат:** docker image/tag + git tag `v1.1.0` как rollback point (после Miro/Boosty); destructive SQL только idempotent + backup note.
- **Поддержка:** баги из F12/voice insights (двойной `/api`, LB `undefined`, пустая history) — чиню сам end-to-end.

Ситуации «решение только на мне» — постоянно: трогать ли Berserk, считать ли 500 VU fail регрессией, что считать CORE для отчёта.

---

### 5. Документация

**Что создавал / веду (пользуюсь сам; для интервью / ментора / будущего себя):**

| Документ | Тип |
|----------|-----|
| `README.md` | architecture sketch, ports, quick start |
| `docs/openapi.yaml` + gateway `/docs` | REST contract |
| `confluence/architecture/EH_SCHEMAS.md`, Mermaid SYSTEM_DESIGN | topology |
| `INTERVIEW_QUESTIONS_EH.md`, `INTERVIEW_PATTERNS.md` | ports, patterns, WS truth, SQL practice |
| `V1_1_0_FINAL_STATUS.md`, `WHAT_TO_DO_IRL.md`, ticklist | release / IRL checklist |
| `LOAD_RESULTS/`, `K6_TESTS.md`, `SECURITY_BASELINE.md` | load + security posture |
| `BOOSTY_FIX.md` | public product copy |
| Per-service README (billing и др.) | local API how-to |

**Пример «без меня»:** инструкция seeds через `docker exec` / `make seed-v110` в `LOAD_RESULTS/README.md` и `WHAT_TO_DO_IRL.md` — после ошибки `psql` на unix socket `:5432` можно накатить shop/history seeds, не спрашивая «какой DSN». Аналогично: «не гоняй legacy 500 VU как CORE» — снимает ложный инцидент.

---

### 6. PostgreSQL

**Сильный кейс: transactional purchase + outbox + идемпотентный refund (shop/billing) и schema evolution через goose.**

Лично:

- Миграции goose на сервисных БД (отдельный Postgres на сервис: shop `:5465`, inventory `:5466`, profile `:5464`, history `:5469`, …) — изоляция схем.
- Shop: purchase пишет покупку + outbox `shop.purchased` в одной логике; integration test проверяет `outbox` row `processed=false`.
- Cancel/refund: окно 7 дней, флаги already refunded / fulfilled — unit-покрыто; spend/refund билетиков через billing с инвалидацией Redis cache.
- Inventory: `CreateItemWithOutbox` — item + outbox event в PG-пути (Mongo path — без outbox).
- Profile: миграция ачивок `amateur/pro/hero` × 8 игр (`20261008120000_…`), evaluate от `best_scores`.
- Операционно: cleanup SQL с **явными UUID**, assert `inventory_cards_active = 281` — защита реальных Berserk карточек.
- Оптимизация «медленного» в смысле продукта: FE группировка inventory по `item_id` (N покупок одного скина), чтобы не лечить «дубли» тяжёлым SQL на каждый list — UX fix при корректной нормализованной модели purchases.

Индексы: GIN на `inventory_items.attributes`, индексы по `author_id`/`type` из init migrations; правило команды — не плодить индексы до EXPLAIN после measured k6.

---

### 7. NATS / Kafka

**Факт топологии:** основной bus — **NATS JetStream**. Kafka — **opt-in** (`make deploy-heavy`), не обязателен для thin deploy.

**Что настраивал / держал руками:**

- **nats-hub:** создаёт/обновляет stream `EVENTS` (FileStorage, MaxAge 7d, MaxBytes 1GiB) с явным списком subjects (`score.updated`, `user.registered`, `shop.purchased`, `purchase.paid`/`fulfilled`, `inventory.item.*`, …). Durable listener `nats-hub-listener` на `event.>` с Ack.
- **History ingest:** отдельные durable consumers `history-<subject>`, **ManualAck**; при ошибке `RecordEvent` → **Nak** (повторная доставка), при успехе Ack. Нормализация subject `event.user.registered` → тип `user.registered` (идемпотентность на уровне доменного типа для FE-фильтров).
- **Outbox workers** (inventory/shop/billing pattern): at-least-once publish из таблицы `outbox`, mark `processed` после успешного JetStream publish — защита от «событие ушло, а транзакция откатилась» наоборот.
- **Gateway register:** dual-publish `event.user.registered` + `user.registered`, чтобы не ломать старых/новых подписчиков.

**Чего нет «как в Kafka textbook» и как говорим честно:**  
JetStream ≈ stream + durable consumers (аналог consumer group/offsets внутри NATS), не Kafka partitions/rebalance API. Отдельного классического DLQ-топика в коде history нет — retry через Nak; poison messages смотрим по логам. Consumer lag — через `StreamInfo` / мониторинг, не отдельный Lag API как в Kafka. Порядок: полагаемся на subject design + single-writer семантики домена, не на global total order across all subjects.

---

### 8. API Design (REST / gRPC)

**Самый сложный контракт: Shop purchase + cancel (+ merch gate через Payment), проксируемый Gateway REST → gRPC Shop/Billing/Payment.**

Что проектировал / держал в синхроне:

- **REST (публичный):** `POST /api/shop/purchase`, `POST /api/shop/purchase/:id/cancel`, `GET /api/shop/items`, `GET /api/shop/inventory` — зафиксированы в `docs/openapi.yaml`, отдаются через Gin gateway `:8079` → gateway replicas. Версионирование: **без `/api/v1`** — стабильный префикс `/api/`; ломающие изменения избегаем, добавляем поля (например `can_cancel` в proto/DTO).
- **gRPC Shop:** сообщения с `validate.rules` / hand-written `Validate()` до полного `task gen-proto`; handler тонкий, логика в service.
- **Ошибки:** доменные (`ErrRefundWindowExpired`, `ErrAlreadyFulfilled`) → осмысленный HTTP/gRPC status через gateway mapping; клиент FE показывает toast, не «500 всё подряд».
- **Timeouts / deadlines:** HTTP clients не `http.DefaultClient`; gRPC context с deadline на межсервисных вызовах; k6 thresholds как SLO-smoke (p95).
- **Совместимость:** пример — добавление ingest subject `event.user.registered` **без** удаления `user.registered`; FE `can_cancel` + группировка inventory на клиенте не требовали breaking change API list purchases.
- **Смежный контракт:** Payment `CanPurchaseMerch` — shop/FE спрашивают до покупки физического мерча (подписка), цена пола ≥ 100_000 tickets — продуктовое правило на границе API + seed/SQL.

**Случай «не сломать потребителей»:** смена копирайта/иконок и dual NATS publish; OpenAPI остаётся source of truth для REST, proto — для mesh. FE axios `baseURL: '/api'` + относительные пути — контрактная дисциплина, чтобы не поймать `/api/api/*` снова.

---

### Короткие ссылки на код (если попросят показать)

- Rate limit: `services/gateway/internal/ratelimit/limiter.go` (login 5/s IP, global 100/s, submit 10/s)
- bcrypt 12: `services/auth/internal/service/auth_service.go`
- Shop cancel tests: `services/shop/internal/service/shop_service_test.go`
- History Nak/Ack: `services/history/internal/worker/workers.go`
- EVENTS stream: `services/nats-hub/internal/app/app.go`
- OpenAPI: `docs/openapi.yaml`
