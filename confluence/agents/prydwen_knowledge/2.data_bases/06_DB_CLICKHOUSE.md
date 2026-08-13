# ClickHouse

Колоночная OLAP СУБД для аналитики и событий. В Event Horizon сервис **analytics** ходит в ClickHouse по **HTTP** (порт 8123), без native-драйвера в базовом варианте.

## OLAP vs OLTP

| | OLTP (Postgres) | OLAP (ClickHouse) |
|--|-----------------|-------------------|
| Нагрузка | много мелких RW транзакций | крупные сканы, агрегаты |
| Хранение | строки | колонки + сжатие |
| Изменения | частые UPDATE/DELETE | append-heavy, редкие мутации |
| Запросы | point lookup, JOIN умеренно | GROUP BY, time-series, funnel |

Не используйте ClickHouse как primary для балансов/инвентаря. Не используйте Postgres как единственное хранилище тяжёлой аналитики на сырых событиях.

## MergeTree семейство

Базовый движок: **MergeTree** (и Replacing-/Summing-/Aggregating-/ReplicatedMergeTree…).

- Данные партиционируются (часто по месяцу/дню: `toYYYYMM(ts)`).
- **ORDER BY** (ключ сортировки) = главный «индекс» для primary key / sparse index — выбирайте под фильтры (`user_id, ts`).
- Фоновые **merge** кусков; много мелких insert → слишком много parts → «too many parts».
- Батч-insert (тысячи строк) предпочтительнее построчного.

Мутации `ALTER UPDATE/DELETE` — тяжёлые, асинхронные; для аналитики чаще append + ReplacingMergeTree / версионирование.

## HTTP-интерфейс и параметры

EH analytics (`services/analytics/.../clickhouse`):

- URL вида `http://clickhouse:8123`, БД `eventhorizon` (env `CLICKHOUSE_URL`, `CLICKHOUSE_DB`).
- Запросы через HTTP; параметры — **named** `param_<name>` (безопаснее конкатенации SQL).
- Таймауты своего `http.Client` (не `DefaultClient`).
- Health/ready — ping к CH наряду с другими зависимостями.

Практика: read-only пользователь для сервиса, лимиты `max_execution_time`, не светить 8123 наружу без auth/сети.

## Роль в Event Horizon

- События/метрики приходят из стрима (NATS и др.) → воркеры analytics пишут в CH.
- gRPC API analytics читает агрегаты для фронта/админки.
- Postgres остаётся для операционных данных; CH — отчёты, воронки, time-series.

Схема: широкие факты, низкая кардинальность измерений в ORDER BY где возможно; Materialized View для предоaggregатов при росте объёма.

## Питфолы

- Крошечные inserts каждую миллисекунду.
- `SELECT *` на миллиардах строк без LIMIT/фильтров по partition key.
- JOIN как в Postgres «на лету» огромных таблиц — лучше денормализация/словари (dictionaries).
- Ожидание мгновенной consistency после insert (есть задержка видимости/merge в некоторых сценариях).
- Хранение PII без политики retention.

## Типичные вопросы на собесе

- Чем колоночное хранение помогает агрегатам?
- Что такое MergeTree и зачем PARTITION BY / ORDER BY?
- Почему плохи частые мелкие INSERT?
- OLAP vs OLTP — куда поставить billing ledger, а куда clickstream?
- Как EH analytics параметризует HTTP-запросы и почему не string concat?
- Чем ReplacingMergeTree отличается от обычного MergeTree?
