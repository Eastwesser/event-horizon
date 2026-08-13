# PostgreSQL: VACUUM и bloat

MVCC оставляет мёртвые версии строк. VACUUM их убирает; без него — bloat, раздувание таблиц/индексов и риск **transaction ID wraparound**.

## Что делает VACUUM

- Помечает место мёртвых tuple’ов как переиспользуемое (обычно **не** отдаёт диск ОС — для этого `VACUUM FULL`).
- Обновляет **visibility map** → возможны Index-Only Scans.
- Поддерживает статистику вместе с `ANALYZE` (autovacuum часто делает оба).
- `VACUUM FREEZE` — продвигает xmax/xid freeze, защищает от wraparound.

Обычный VACUUM не блокирует DML сильно; `VACUUM FULL` переписывает таблицу и берёт тяжёлый лок — в проде редкость, чаще `pg_repack`.

## Autovacuum

Фоновый процесс, триггерится по порогам:

- roughly: dead tuples > `threshold + scale_factor * reltuples`
- На горячих таблицах (outbox, сессии, inventory movements) дефолты могут не успевать → **per-table** `autovacuum_vacuum_scale_factor` меньше, отдельный `vacuum_cost_*`.

Следить:

- `pg_stat_user_tables`: `n_dead_tup`, `last_autovacuum`, `last_analyze`
- возраст xid: `age(datfrozenxid)` в `pg_database`

## Bloat

- Таблица/индекс логически маленькие, физически огромные из-за dead space / фрагментации.
- Симптомы: Seq Scan читает гигабайты, медленные UPDATE, раздутый backup.
- Причины: частые UPDATE больших строк, долгие транзакции/слоты репликации, отстающий autovacuum, HOT-update не срабатывает (если меняются индексированные колонки).

**HOT (Heap-Only Tuple):** UPDATE без смены индексных колонок может избежать лишней индексной работы — проектируйте «часто меняющиеся» поля вне лишних индексов.

## XID wraparound

- Transaction ID — 32-bit; при исчерпании Postgres **остановит** запись, чтобы не перепутать видимость.
- Защита: регулярный freeze через vacuum.
- Мониторинг age; алерты задолго до ~2B.
- Долгий `idle in transaction`, забытый prepared transaction, отстающий replication slot — держат горизонт и мешают vacuum.

## Практика для EH

- Коротких транзакций в repository достаточно; не держать tx на время вызова внешних API.
- Outbox-таблица: высокий churn → агрессивнее autovacuum + partial index по unpublished.
- После массовых DELETE/батчей — явный `VACUUM (ANALYZE)` в maintenance window при необходимости.
- Connection `MaxConnLifetime=5m` не заменяет vacuum, но снижает сюрпризы сессий.

## Типичные вопросы на собесе

- Зачем VACUUM, если есть MVCC?
- Чем VACUUM отличается от VACUUM FULL?
- Что такое autovacuum и почему его тюнят на горячих таблицах?
- Что такое bloat и как его обнаружить?
- Что такое xid wraparound и чем грозит?
- Почему долгая транзакция мешает очистке?
