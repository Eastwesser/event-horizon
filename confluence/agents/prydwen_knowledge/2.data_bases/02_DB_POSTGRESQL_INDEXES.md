# PostgreSQL: индексы

Индекс ускоряет поиск ценой записи и места. Senior-ответ: не «везде B-tree», а селективность, partial/covering и безопасный деплой через `CONCURRENTLY`.

## B-tree (по умолчанию)

- Подходит для `=`, `<`, `>`, `BETWEEN`, `IN`, `ORDER BY`, `LIKE 'prefix%'` (не `%suffix`).
- Составной индекс `(a, b, c)` — работает слева направо: условие по `a`, или `a+b`, …; «только `b`» обычно индекс не использует эффективно.
- Правило большого пальца: **равенства → диапазон → сортировка** в порядке колонок.

Другие типы (кратко для собеса):

- **Hash** — почти только `=`; редко нужен.
- **GIN** — jsonb, массивы, full-text.
- **GiST / SP-GiST** — гео, range, trigram.
- **BRIN** — огромные append-only таблицы по корреляции с физическим порядком.

## Partial indexes

```sql
CREATE INDEX ON orders (user_id) WHERE status = 'pending';
```

- Меньше размер, быстрее обновления, если запросы всегда с тем же предикатом.
- В EH-сценариях: «активные сессии», «неоплаченные платежи», outbox `WHERE published_at IS NULL`.

## Covering / Index-Only Scan

- Include-колонки (PG 11+): `CREATE INDEX ON t (id) INCLUDE (status, updated_at)`.
- Heap fetch не нужен, если все нужные поля в индексе и visibility map позволяет (после VACUUM).
- Не тащите в INCLUDE огромные TEXT без нужды — раздуете индекс.

## Unique и FK

- `PRIMARY KEY` / `UNIQUE` → B-tree уникальный.
- FK выигрывают от индекса на **referencing** колонке (иначе DELETE/UPDATE родителя = Seq Scan дочерней).

## Когда индекс вреден

- Низкая селективность (`boolean`, «пол») без partial.
- Таблица с редким чтением и частой записью — каждый индекс = доп. random I/O на INSERT/UPDATE.
- Дубликаты индексов `(a)` и `(a,b)` — первый часто избыточен.

## CREATE INDEX CONCURRENTLY

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS ...
```

- Не блокирует запись длинным `ACCESS EXCLUSIVE` как обычный CREATE INDEX (но занимает больше шагов, может упасть → invalid index).
- Нельзя внутри транзакционного миграционного блока в некоторых инструментах — проверяйте migrator.
- После фейла: `INVALID` индекс → `DROP INDEX CONCURRENTLY` и повтор.
- В проде EH-сервисов: тяжёлые индексы на больших таблицах — только CONCURRENTLY + мониторинг replication lag.

## Эксплуатация

- `pg_stat_user_indexes` — unused indexes (низкий `idx_scan`).
- `EXPLAIN (ANALYZE, BUFFERS)` — используется ли индекс, сколько heap fetches.
- Раздутый индекс → `REINDEX CONCURRENTLY` (PG 12+).

## Типичные вопросы на собесе

- Почему составной `(a,b)` не помогает фильтру только по `b`?
- Что такое partial index и когда применять?
- Зачем INCLUDE и Index-Only Scan?
- Чем опасен обычный `CREATE INDEX` на большой таблице в проде?
- Нужен ли индекс на колонке FK?
- Как понять, что индекс не используется?
