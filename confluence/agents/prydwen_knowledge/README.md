# Prydwen — база знаний Event Horizon (индекс)

Prydwen — markdown-корпус для людей и для RAG (`search_prydwen` в MCP). Цель: быстрые, точные ответы по Go, архитектуре и **реального** проекта EH + легенды для собеса.

## Как готовиться к собесе завтра (red_mad_robot)

1. **30–40 мин — каркас проекта:** `8.legend_projects/03_PROJECT_EVENT_HORIZON.md` (порты, Outbox, Redis, RBAC, circuit, MCP).
2. **20 мин — интеграции:** `4.architecture_patterns/03_ARCH_INTEGRATION_PATTERNS.md` + `3.message_brokers/04_BROKERS_NATS.md`.
3. **20 мин — безопасность и HTTP:** `9.common_backend/02_STATUS_CODES.md`, `03_SECURITY.md`, `01_QUERIES.md`.
4. **15 мин — тесты:** `6.testing/01_UNIT_TESTING.md`, `04_INTEGRATIONAL_TESTING.md`.
5. **15 мин — AI/MCP (если спросят):** `7.ai_engineering/*`.
6. **По желанию легенды:** `01_LEGEND_ADTIME.md`, `02_LEGEND_ROOLZ.md` — только как каркас, без выдуманных личных KPI.
7. Прогон вслух: purchase flow (Shop→Billing→Outbox→NATS→History/Analytics) и «401 vs 403».

Формат файлов: структурированная шпаргалка, в конце блок **«Типичные вопросы на собесе»** — удобно для self-check.

## Карта разделов → темы интервью

| Раздел | О чём спрашивают |
|--------|------------------|
| `1.golang_fundamentials/` | типы, slices/maps, goroutines, channels, sync, context, GMP, GC, errors |
| `2.data_bases/` | Postgres основы/индексы/vacuum/perf, Mongo, ClickHouse |
| `3.message_brokers/` | сравнение брокеров, Rabbit, Kafka, **NATS JetStream** |
| `4.architecture_patterns/` | микросервисы, паттерны, **Outbox/Saga/CB**, сеть |
| `5.devops/` | Docker, k8s, observability, CI/CD |
| `6.testing/` | unit table-driven/mocks, integration/e2e/k6/smoke |
| `7.ai_engineering/` | LLM tokens/галлюцинации, RAG, MCP-агенты |
| `8.legend_projects/` | AdTime/Roolz (легенды), **EH — правда репо** |
| `9.common_backend/` | пагинация/идемпотентность, статусы, security |

## Быстрые якоря по сервисам EH

- **Auth** → JWT, bcrypt 12, Redis sessions + `9.common_backend/03_SECURITY.md`
- **Shop/Billing/Payment** → Outbox, Saga-лайт, merch gate + integration patterns
- **Inventory** → эталон Cache-Aside + Outbox + Mongo + optimistic version
- **Gateway** → RBAC, circuit breaker → 503
- **Analytics** → ClickHouse parameterized queries
- **MCP** → `7.ai_engineering/03_AI_AGENTS_MCP.md`

## Для LLM-агентов

1. Сначала `search_prydwen`, не выдумывать порты.
2. Source of truth по EH — `8.legend_projects/03_PROJECT_EVENT_HORIZON.md`.
3. Писать в корпус так, чтобы retrieval работал: один топик, имена сервисов и subjects.

## Правила поддержки корпуса

- Русский язык в телах шпаргалок для собеса.
- Не коммитить секреты; не индексировать `.env`.
- Противоречия по портам/статусам — править EH-файл, не плодить форки правды.

Удачи на собесе: говори архитектуру EH уверенно, легенды — честно как narrative, безопасность и статусы — коротко и с примерами.
