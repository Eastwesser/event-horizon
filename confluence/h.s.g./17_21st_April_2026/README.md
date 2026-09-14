# Урок 17 — 21 апреля 2026 (Postgres / ClickHouse / интеграция)

Конспекты БД + легенды Roolz. Код — иллюстрации концептов, **не** реальная БД.

## Что говорить на собесе

### PostgreSQL
- **MVCC:** UPDATE = xmax на старой версии + новая строка → bloat; **VACUUM** помечает место внутри файла.
- **WAL:** сначала журнал → data files; реплики читают WAL; PITR.
- **Кэш:** shared_buffers + OS page cache.

### ClickHouse (легенда)
«В Roolz CH для аналитики событий 3M+/час — OLAP, допускаем дубли/потерю. Exact write — в PostgreSQL. Debezium (WAL) → Kafka → CH. Для дашбордов 99.9% хватало.»

### Интеграция
- **Антипаттерн:** Dual Writes, 2PC.
- **Outbox:** бизнес + outbox в одной SQL-транзакции; воркер → Kafka; consumer идемпотентен.
- **CDC:** Debezium на WAL — без polling, replay возможен.

### Выбор инструмента (одна фраза)
PG — OLTP; TimeScale — time-series чанки; CH — колоночный OLAP/логи; Cassandra — широкий write; Neo4j — граф; ES — полнотекст; Dragonfly — многопоточный Redis-compatible (эксперимент).

## Задачи (демо-сниппеты)

### 1. Outbox demo
Канал = «брокер»; журнал outbox в слайсе; poller шлёт события.

### 2. MVCC versions
Имитация xmin/xmax: update не затирает, а создаёт версию.

### 3. DB choice
Таблица сценарий → инструмент (для устного ответа).

## Как запускать

```bash
export GOWORK=off
cd code/01_outbox_demo && go run main.go
```
