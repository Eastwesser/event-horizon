## Cold start после рестарта (10.11.2026)

### Что было
- P=50 **первый** прогон после рестарта Inventory (миграция индексов):
  browse 5212ms, 2.26% errors (`circuit inventory` открылся), 5xx 0.49%.
- P=50 **второй/третий**: browse 3137ms, 0% errors. **Вернулось** к baseline (3166ms).

### Причина
- PG cold buffer (inventory_items не в shared_buffers).
- Redis cold cache (inventory list кэш пустой).
- Gateway 3 реплики × пустой sync.Map (TTL 1s, 10 токенов — прогревается за 1s).

### Вывод
- **Это не баг кэша** — это cold start **любого** сервиса после рестарта.
- **`circuit inventory` открылся** — 5 подряд timeout на cold PG, CB защитил (half-open через 10s, закрылся).
- **Кэш ValidateToken полезен**: ручной тест 680→66ms (×10), P=50 warm вернулся к baseline.

### Улучшения (не в этом спринте)
- `pg_prewarm` для inventory_items.
- Warmup-запросы в `make deploy` после старта.
- Cold start **всегда** будет — **не** паниковать, **подождать** 30–60s.