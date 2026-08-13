# PostgreSQL: производительность

Как искать тормоза: EXPLAIN, N+1, пулы соединений. Цифры EH — не магия, а дисциплина.

## EXPLAIN / EXPLAIN ANALYZE

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT ...;
```

Смотреть:

- **Seq Scan vs Index Scan / Bitmap** — ожидаемо ли на объёме;
- **Rows estimated vs actual** — расхождение → устаревшая статистика (`ANALYZE`), кривой planner;
- **Buffers hit/read** — упираемся в диск или shared_buffers;
- **Sort / HashAggregate** — memory vs disk (`work_mem`);
- **Nested Loop** с огромным inner — часто забытый индекс или N×M.

План ≠ приговор: на маленьких таблицах Seq Scan нормален.

## Классические антипаттерны

### N+1

В сервисе: один `List`, потом в цикле `GetByID` / lazy load. В БД: ORM без prefetch.

Лечение: `JOIN`, `WHERE id = ANY($1)`, batch-запрос, DataLoader на gateway — но в EH лучше правильный SQL в repository, не магия в handler.

### SELECT *

Тянет TOAST/лишние колонки, мешает covering index. Для списков — узкий projection.

### Сортировка без индекса

`ORDER BY created_at DESC LIMIT 20` без индекса → sort всего набора. Пагинация: keyset (`WHERE (created_at, id) < (…)`) лучше OFFSET на больших страницах.

### Неявные преобразования

`WHERE indexed_col = $1` с другим типом → seq scan. UUID/text/int — совпадайте типы с колонкой.

## Пул соединений (EH)

Во всех основных сервисах:

| Параметр | Значение |
|----------|----------|
| MaxOpen / MaxConns | **25** |
| MinIdle / MinConns | **10** |
| MaxConnLifetime | **5 minutes** |

Почему это влияет на perf:

- Слишком большой пул → очереди на Postgres, context switch, рост latency (не throughput).
- Слишком маленький → `timeout acquiring connection` при всплесках; смотрите saturation метрик пула.
- Lifetime 5m — равномерный reconnect, меньше «липких» проблем после сетевых blip.

Сумма `replicas * 25` должна укладываться в `max_connections` (с запасом на админов/миграции) или идите в **PgBouncer** (transaction mode осторожно с prepared statements / session features).

## Контекст и таймауты

- `QueryContext` + deadline из gRPC ctx — не копить запросы после отмены клиента.
- Statement timeout на роли/сессии — защита от «случайно» тяжёлого запроса.
- `/ready` ping не должен ходить тяжёлым SQL.

## Прикладной чеклист EH

1. Медленный RPC → trace_id → SQL в логах / slow query.
2. EXPLAIN ANALYZE на копии данных.
3. Индекс / перепись запроса / убрать N+1 в service.
4. Проверить pool saturation и locks (`pg_stat_activity`, `wait_event`).
5. Не чинить увеличением MaxConns «до 200», пока не доказали, что bottleneck — пул, а не запросы.

## Типичные вопросы на собесе

- Как читать EXPLAIN ANALYZE? Что такое nested loop trap?
- Что такое N+1 и как лечить на уровне SQL и кода?
- Почему увеличение pool size часто ухудшает latency?
- Зачем в EH MaxConns=25, MinConns=10, MaxConnLifetime=5m?
- Keyset vs OFFSET пагинация?
- Как statement timeout связан с context в Go?
