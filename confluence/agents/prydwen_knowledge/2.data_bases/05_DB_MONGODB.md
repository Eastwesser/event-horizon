# MongoDB

Документоориентированная БД: BSON-документы, гибкая схема, горизонтальное масштабирование через шардирование. В Event Horizon — **экспериментальный** контур inventory; Postgres остаётся источником истины для транзакционных сценариев. Не удалять Mongo из inventory без явного решения.

**Собес-банк (DesChat / AdTime):** `confluence/interview/2.database/mongodb_adtime_chat.md`  
**Легенда:** `8.legend_projects/01_LEGEND_ADTIME.md` (GIFTS + deschat, не ad-tech)  
**Mermaid кейсы:** `unrelated/kata-lectors/5.system_design/sd_schemas.md` (§1 AdTime, §5 Messenger)

## Модель данных

- **Документ** ≈ JSON (BSON): вложенные объекты, массивы.
- **Коллекция** ≈ таблица без фиксированной схемы (схема на практике живёт в коде/`bson` tags).
- `_id` обязателен (ObjectId или ваш UUID/string).
- Денормализация нормальна: читаете одним запросом то, что в SQL было бы JOIN’ом.

Проектирование: моделируйте под **запросы**, не под ER-диаграмму 1:1. Избыточность vs сложность обновлений в многих местах.

## CRUD и драйвер Go

- Официальный `mongo-driver` (`go.mongodb.org/mongo-driver`): client → database → collection.
- Всегда с `context.Context` и timeout.
- `Find` / `FindOne` / `UpdateOne` / `DeleteOne`; cursor закрывать (`All` / `Next`).
- Multi-doc transactions — только при **replica set** (`session` / `WithTransaction`).
- Теги: `bson:"title"`, `ID primitive.ObjectID \`bson:"_id,omitempty"\``.
- Индексы: `mongo.IndexModel` + `Indexes().CreateOne`.

Multi-document ACID слабее привычного Postgres для сложной доменной логики: сессии есть, но стоимость и ограничения (retryable writes, snapshot, cross-shard) другие. Outbox + строгая инварианта инвентаря в EH опираются на **Postgres**.

## Индексы

- B-tree по умолчанию; составные — тот же left-prefix принцип.
- Unique, TTL (expire sessions / archive messages), partial filter, text, wildcard (осторожно).
- `explain("executionStats")` — COLLSCAN vs IXSCAN.
- Лишние индексы бьют по write throughput так же, как в Postgres.

### DesChat (шпаргалка индексов)

| Запрос | Индекс |
|--------|--------|
| История чата | `{ chatId: 1, createdAt: -1 }` |
| Сообщения юзера в чате | `{ chatId: 1, userId: 1 }` |
| Список чатов | `{ "participants.userId": 1, updatedAt: -1 }` |
| Архив | TTL на `messages.createdAt` (конспект: 90d) |

Пагинация: предпочитай cursor (`createdAt` + `_id`), не огромный `$skip`. Агрегация + `$lookup` — только если нужно; часто проще отдельный read users.

## Транзакция «отправить сообщение»

1. Start session / `WithTransaction`  
2. Insert → `messages`  
3. Update → `chats.lastMessage` + `updatedAt`  
4. Inc unread у других participants  
5. Commit  

Optimistic lock на `participants` / `version`, чтобы два юзера не влезли гонкой. Cross-shard TX дороже → shard key ≈ `chatId`.

## Когда Mongo уместен

- Гибкие/полуструктурированные документы, частая эволюция полей (чат, notes, экспериментальный inventory).
- Append-heavy чаты с денорм `lastMessage`.
- Горизонтальный scale шардированием по ключу доступа.
- Временные/экспериментальные проекции рядом с основной системой.

## Когда Mongo не нужен (и EH / Lime)

- Строгие многосущностные транзакции, деньги, биллинг, oversell → **PostgreSQL** (`FOR UPDATE` / optimistic version).
- Каталог с фасетами + full-text витрина → PG + **Elasticsearch** (Mongo text — ок для простого, не для Lime-scale фильтров).
- Простые реляционные связи с кучей ad-hoc JOIN/reporting → SQL/OLAP.
- Write-by-key на огромном RPS без сложных агрегаций → Cassandra/Scylla (путь Discord), не «Mongo forever».

**Event Horizon inventory:** Mongo допускается как experimental store/проекция; критичные операции (резерв, списание, outbox) — Postgres. Не «переносим всё в Mongo ради моды».

## Миграция Mongo → Postgres (+ jsonb)

Dual-write → backfill → read shadow → cutover → stop Mongo writes.  
jsonb + GIN / expression indexes на частые пути. Смысл ухода: JOIN, отчёты, привычный TX, команда SQL-first.

## Питфолы

- Неограниченный рост документов (16MB) — массивы без границ.
- Hot shard key (`createdAt` только) — неравномерная нагрузка.
- Forget index → COLLSCAN на миллионах.
- Сравнение с SQL «просто JSON в jsonb» — Postgres jsonb + GIN часто закрывает гибкость без второй БД.
- Версии: multi-doc TX с **4.0**; в учебных compose часто `mongo:7.0.5`; актуальный major — смотри release notes перед собесом (не врать «у нас 8», если образ 7).

## Docker (учебный якорь)

`mongo:7.0.5`, auth из env, volume `/data/db`, healthcheck `mongosh` ping, порт 27017. Секреты — только env, не в git.

## Типичные вопросы на собесе

- Чем документная модель отличается от реляционной?
- Есть ли multi-document транзакции и какие ограничения?
- Как выбрать shard key?
- Когда TTL index уместен?
- Почему в EH inventory Mongo experimental, а деньги в Postgres?
- COLLSCAN vs IXSCAN — как проверить?
- Как атомарно обновить `lastMessage` при новом сообщении? (код `WithTransaction`)
- Когда Cassandra/Scylla вместо Mongo?
- Как zero-downtime уйти на Postgres + jsonb?
- Почему Mongo ок для AdTime DesChat, но плох как единственный стор Lime-каталога?

Сырьё голосовой: `unrelated/kata-lectors/MONGO/mongo-db-sobes.md`.
