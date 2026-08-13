# MongoDB

Документоориентированная БД: BSON-документы, гибкая схема, горизонтальное масштабирование через шардирование. В Event Horizon — **экспериментальный** контур inventory; Postgres остаётся источником истины для транзакционных сценариев. Не удалять Mongo из inventory без явного решения.

## Модель данных

- **Документ** ≈ JSON (BSON): вложенные объекты, массивы.
- **Коллекция** ≈ таблица без фиксированной схемы (схема на практике живёт в коде/`json` tags).
- `_id` обязателен (ObjectId или ваш UUID/string).
- Денормализация нормальна: читаете одним запросом то, что в SQL было бы JOIN’ом.

Проектирование: моделируйте под **запросы**, не под ER-диаграмму 1:1. Избыточность vs сложность обновлений в многих местах.

## CRUD и драйвер Go

- Официальный `mongo-driver`: client → database → collection.
- Всегда с `context.Context` и timeout.
- `Find` / `FindOne` / `UpdateOne` / transactions (multi-doc — только при replica set).

Multi-document ACID слабее привычного Postgres для сложной доменной логики: сессии транзакций есть, но стоимость и ограничения (retryable writes, snapshot) другие. Outbox + строгая инварианта инвентаря в EH опираются на **Postgres**.

## Индексы

- B-tree по умолчанию; составные — тот же left-prefix принцип.
- Unique, TTL indexes (expire sessions/cache docs), partial filter expressions, text, wildcard (осторожно).
- `explain("executionStats")` — COLLSCAN vs IXSCAN.
- Лишние индексы бьют по write throughput так же, как в Postgres.

## Когда Mongo уместен

- Гибкие/полуструктурированные документы, частая эволюция полей.
- Read-heavy каталоги с денормализацией.
- Горизонтальный scale шардированием по ключу доступа.
- Временные/экспериментальные проекции рядом с основной системой.

## Когда Mongo не нужен (и EH)

- Строгие многосущностные транзакции, деньги, биллинг → **PostgreSQL**.
- Простые реляционные связи с кучей ad-hoc JOIN/reporting → SQL/OLAP.
- «Схема потом» без индексов и без бюджета на consistency → хаос в проде.

**Event Horizon inventory:** Mongo допускается как experimental store/проекция; критичные операции (резерв, списание, outbox) — ориентир на Postgres + транзакции. Не «переносим всё в Mongo ради моды».

## Питфолы

- Неограниченный рост документов (16MB limit) — массивы без границ.
- Hot shard key (`created_at` только) — неравномерная нагрузка.
- Forget index → COLLSCAN на миллионах.
- Сравнение с SQL «просто JSON в jsonb» — Postgres jsonb + GIN часто закрывает гибкость без второй БД.

## Типичные вопросы на собесе

- Чем документная модель отличается от реляционной?
- Есть ли multi-document транзакции и какие ограничения?
- Как выбрать shard key?
- Когда TTL index уместен?
- Почему в EH inventory Mongo experimental, а деньги в Postgres?
- COLLSCAN vs IXSCAN — как проверить?
