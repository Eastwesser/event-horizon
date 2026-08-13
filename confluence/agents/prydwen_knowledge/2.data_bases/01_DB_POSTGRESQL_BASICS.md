# PostgreSQL: основы (MVCC, транзакции, пул)

OLTP-сердце Event Horizon (auth, inventory, billing, shop…). На собесе ждут MVCC, уровни изоляции и грамотный connection pool.

## MVCC в двух словах

- **Multi-Version Concurrency Control:** UPDATE/DELETE не затирают строку in-place — создаётся новая версия; старая видна транзакциям, начавшимся раньше (в зависимости от isolation).
- Читатели не блокируют писателей (и наоборот) на обычных SELECT — в отличие от грубых read locks.
- Цена: dead tuples → нужен VACUUM; длинные транзакции держат «снимки» и раздувают bloat.

## Транзакции

```sql
BEGIN;
-- ...
COMMIT; -- или ROLLBACK
```

- Атомарность: all-or-nothing.
- В Go: `BeginTx(ctx, opts)` / pgx `Begin` — **всегда** `defer rollback`, commit только при успехе; передавайте `ctx` для отмены.
- **EH inventory** — эталон: бизнес-изменения + outbox в **одной** транзакции (без dual-write в Kafka «снаружи»).

## Уровни изоляции

Postgres по умолчанию: **Read Committed**.

| Уровень | Что даёт | Аномалии |
|---------|----------|----------|
| Read Uncommitted | в PG = Read Committed | — |
| Read Committed | каждый statement видит последний commit | non-repeatable read, phantom |
| Repeatable Read | снимок на всю транзакцию | serialization failure при конфликте записи |
| Serializable | строгая сериализация (SSI) | больше `40001`, нужен retry |

- **Lost update:** два Read Committed прочитали одно, оба записали — последний выиграл. Лечится: `SELECT … FOR UPDATE`, optimistic version column, или Repeatable Read + retry.
- На собесе уточняйте: в PostgreSQL RR **сильнее**, чем в SQL-стандарте (нет phantom для снимка), но write skew возможен без Serializable.

## Блокировки (кратко)

- Row-level на UPDATE/DELETE; `FOR UPDATE` / `FOR SHARE`.
- Deadlock → одна транзакция откатывается; клиент должен **retry**.
- Не держите открытую транзакцию на время HTTP к внешнему сервису — только локальная работа + outbox.

## Connection pool (Event Horizon)

Стандарт EH:

- **MaxOpen / MaxConns = 25**
- **MinIdle / MinConns = 10**
- **MaxConnLifetime = 5m**

Зачем:

- Postgres `max_connections` конечен; N сервисов × max pool легко упираются в лимит.
- Lifetime ротация сбрасывает «жирные» сессии, помогает после DDL/prepared statements quirks.
- Min idle — меньше cold-start latency на `/ready` и первом запросе.

Пулы: `database/sql` (`SetMaxOpenConns`…) или **pgxpool**. Не открывайте новый TCP на каждый RPC. Не используйте один глобальный conn без пула.

**Формула порядка:** `(CPU cores * 2) + spindles` — эвристика; в k8s важнее суммарный бюджет подов vs `max_connections` и PgBouncer.

## /health и /ready

- Liveness: процесс жив.
- Readiness: `Ping` к Postgres (и другим зависимостям) — как в EH metrics HTTP. Не путайте: падение БД должно выводить под из балансировки, не обязательно рестартить бесконечно.

## Типичные вопросы на собесе

- Как MVCC позволяет читать без блокировки писателей?
- Чем Read Committed отличается от Repeatable Read в PostgreSQL?
- Как избежать lost update при списании баланса (billing)?
- Почему длинная транзакция опасна?
- Зачем MaxConns=25 и MaxConnLifetime=5m в микросервисах EH?
- Где границы транзакции при Outbox pattern?
