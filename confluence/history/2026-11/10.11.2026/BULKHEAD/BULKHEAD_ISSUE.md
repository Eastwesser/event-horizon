2. Bulkhead full — корневая причина: MaxConcurrent=32 на все сервисы
gateway.go:106:

go
MaxConcurrent: 32, // bulkhead: cap in-flight calls per downstream
32 одновременных запроса на Shop (и на Inventory, и на всё остальное). А у нас 140 browse-VU × 4 ручки = 560 запросов, из них ~140 идут в Shop (shop/items), ~140 в Inventory (inventory/items), ~140 в Leaderboard, ~140 в Profile. Bulkhead 32 на Shop → 108 из 140 сразу получают 503 bulkhead full. Это и есть «Shop не тянет 140 browse-VU» — он не получает больше 32 одновременно.

Фикс:

Поднять MaxConcurrent до 128–256 на Shop/Inventory (read path), оставить 32 на write (purchase).

Или добавить MaxQueue (сейчас нет) — запросы ждут в очереди, а не 503 сразу.

Или разделить bulkhead на read/write: shop_read (256) и shop_write (32).

Приоритет: P1. Не баг, а настройка. Но 32 — это очень мало для read path под 140 VU.

## Результат фикса (10.11.2026, bulkhead 256/128/64/32 + Nginx 5s + deadlock fix)

### Что сделали
- `newServiceBreaker(name, maxConcurrent)` — параметр лимита.
- shop=256, inventory=256, leaderboard/profile/game/auth=128, billing/payment/authors/notification/history=64, analytics=32.
- Nginx: `worker_connections 4096`, `proxy_read_timeout 5s`, `proxy_next_upstream error timeout` (без http_502).

### Проверка (P=50 / P=100 / P=200)

| Профиль | Baseline | После deadlock | **После всех 3 фиксов** |
|---|---|---|---|
| P=50  | 0.67% failed, browse 1894ms | 1.11% failed, browse 4222ms | **0.30% failed, browse 3166ms** |
| P=100 | 9.35% failed, browse 68674ms | 0.69% failed, browse 5057ms | **1.72% failed, browse 5900ms** |
| P=200 | 3.42% failed, browse 17309ms, deadlock+bulkhead | 2.87% failed, browse 9477ms, bulkhead full | **3.21% failed, browse 11515ms, inventory CB** |

### Выводы
- **`deadlock 40P01` — 0** на всех профилях. ✅ Шаг 1 закрыт.
- **`bulkhead full` — 0** на всех профилях. ✅ Шаг 3 закрыт.
- **`No available backends` — 0**. ✅ Шаг 2 закрыт.
- **P=50 — лучший результат**: 0.30% failed, purchase_owned 12% (было 39%).
- **Новый сигнал — `circuit inventory`** (P=200): Inventory **не тянет** 140 browse-VU × `inventory/items` (Seq Scan, нет индексов). CB открывается → browse errors 9.34%.
- **P=200 browse 11.5s** — всё ещё много, но это **следующий** bottleneck (Inventory Seq Scan).

### Что осталось (Шаги 4, 5, 3.5)
- Шаг 4: **ValidateToken кэш** — уберёт избыток коллов в Auth.
- Шаг 5: **GIN-индексы Inventory** — уберёт `circuit inventory` в P=200.
- Шаг 3.5: **короткий резерв** — уберёт contention на 10 items (purchase p95).