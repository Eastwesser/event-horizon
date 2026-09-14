# Урок 23 — Avito unmet demand + health/cache

## Тема
Суммарная неудовлетворённость покупателей (ближайший товар), health-check HTTP urls в `map[string]bool`, партиционированный кэш.

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_unmet_demand` | sort goods + SearchInts, сумма abs(need-goods) |
| `02_healthcheck` | 3 retry, mutex, статус 200 |
| `03_partitioned_cache` | Get/Set по шардам |

## Как решать / что говорить
- **Unmet:** для каждой потребности бинарный поиск; сравнить `g[i]` и `g[i-1]`.
- **Health:** снимок ключей под `RLock`, писать статус под `Lock` — иначе concurrent map fatal.
- **Cache:** hash(key)%N → shard RWMutex.

## Собес-советы
- Уточни: товар бесконечный (как в условии) vs limited stock.
- Не `range` по map во время write без копии ключей.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

