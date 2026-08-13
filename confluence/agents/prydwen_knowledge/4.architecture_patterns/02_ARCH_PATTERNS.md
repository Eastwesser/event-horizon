# Архитектурные паттерны: Clean Architecture, DI, Repository, CQRS lite

Senior-ожидание на собесе (red_mad_robot): уметь объяснить слои, направление зависимостей, тестируемость и когда «облегчённый CQRS» уместен. В Event Horizon эталон — Clean Architecture по сервисам (Inventory как reference).

## Clean Architecture (слои EH)

1. **Handler / API** (`internal/handler/`) — gRPC (или HTTP) адаптеры; маппинг proto ↔ model; **без бизнес-логики**.
2. **Service** (`internal/service/`) — use-cases, правила, оркестрация, транзакции на уровне домена.
3. **Repository** (`internal/repository/`) — доступ к PG/Mongo/Redis/ClickHouse.
4. **Model** (`internal/model/`) — доменные типы, независимые от транспорта.

- Зависимости только внутрь: Handler → Service → Repository. **Никогда наоборот**.
- Инфраструктура (БД-драйверы, NATS client) живёт снаружи и внедряется через интерфейсы.
- Выигрыш: unit-тесты service с fake repo; смена Mongo→PG не ломает handler.

## Dependency Injection (DI)

- Конструкторы явно принимают зависимости (`NewShopService(billingClient, repo, js)`).
- Сборка графа — в `internal/app` / `di.go` (composition root), не по всему коду.
- Избегай service locator и глобальных синглтонов с скрытым состоянием.
- Интерфейсы объявляй там, где нужны потребителю (или рядом с service), не «интерфейс на каждый struct заранее».
- В тестах подменяй интерфейсы моками; интеграционные тесты поднимают реальный PG/Redis.

## Repository pattern

- Репозиторий скрывает SQL/драйвер: `Create`, `GetByID`, `Update`, `List` в терминах домена.
- Не протекай `*sql.Rows` / `pgx.Row` в service — мапь в model.
- Транзакции: либо `WithTx` на repo, либо Unit of Work / передача `tx` аккуратно без context в struct.
- EH: Inventory — reference для Outbox + Redis decorator + транзакции; Billing outbox в том же tx, что изменение баланса.
- Антипаттерн: «God repository» на 40 таблиц или repo, который ходит в другие сервисы по сети.

## CQRS lite (не полный CQRS/ES)

- Полный CQRS: раздельные модели записи и чтения, часто разные хранилища.
- **CQRS lite** в EH:
  - **Leaderboard**: запись рекордов через NATS/PG-путь, чтение топа из **Redis**.
  - **Billing**: запись в PG (source of truth), горячее чтение баланса часто из Redis.
  - **Analytics**: запись событий в **ClickHouse**, чтение DAU/MAU/retention отдельно от OLTP History (Postgres trail).
- Когда достаточно: разные load-профили write/read, денормализованные проекции.
- Когда рано: простое CRUD без bottleneck'ов — лишняя сложность.

## Сопутствующие паттерны (шпаргалка собеса)

- **Factory / Builder**: сборка сложных объектов/конфигов без телескопических конструкторов.
- **Facade**: упростить работу с подсистемой (клиент биллинга за фасадом).
- **Strategy**: сменная политика (разные pricing/discount).
- **Adapter**: обертка внешней системы под ваш порт.
- **Decorator**: Inventory Redis cache вокруг repo; не путать с бизнес-логикой в cache-слое.
- **Outbox / Saga**: см. `03_ARCH_INTEGRATION_PATTERNS.md` — критичны для Shop/Billing.

## C4 / HLD / LLD (как рассказывать)

- **C4 Context**: пользователи и внешние системы.
- **Container**: gateway, сервисы, NATS, Kafka, PG, Redis, CH.
- **Component**: handler/service/repo внутри сервиса.
- **HLD**: границы, протоколы, data flow покупки.
- **LLD**: конкретные структуры, SQL, consumer durable names.

## Антипаттерны слоёв

- Бизнес-правила в gRPC handler или в SQL-триггерах «потому что так быстрее».
- Циклические импорты service↔repo через concrete types.
- `context.Context` в полях структуры (нарушение EH rules) — только аргумент методов.
- Анемичная модель + вся логика в «менеджерах» без границ — допустимо в Go pragmatic-стиле, но объясняй trade-off.

## Типичные вопросы на собесе

- Нарисуйте слои Clean Architecture и направление зависимостей.
- Где должна жить валидация: proto validate, service, или оба?
- Как тестировать service без реальной БД?
- Что такое composition root в DI?
- Чем CQRS lite отличается от event sourcing?
- Почему нельзя вызывать repository из другого bounded context напрямую?
- Как в EH устроен reference-сервис Inventory (outbox, cache, слои)?
