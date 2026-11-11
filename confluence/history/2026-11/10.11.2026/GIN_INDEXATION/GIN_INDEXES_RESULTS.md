## Результат фикса индексов (10.11.2026, P=200)

### Что сделали
- Миграция `20261110000000_add_indexes.sql`: GIN на attributes, btree на type/price/author, GIN trgm на name.
- `EXPLAIN ANALYZE`: `Bitmap Index Scan on idx_inventory_attributes_gin` вместо `Seq Scan` (actual 0.255ms на 20 строк).

### Результат

| Метрика | До индексов | После индексов |
|---|---|---|
| Failed | 3.21% | **1.55%** |
| browse p95 | 11514ms | 10396ms |
| inventory/items p95 | 4455ms | 4173ms |
| status_5xx | 1.96% | **0.00%** |
| inventory errors | 9.34% | **0.00%** |
| circuit inventory | **открывался** | **0** |

### Вывод
- **`circuit inventory` исчез** — прямой эффект GIN-индекса. ✅ Шаг 5 закрыт.
- **`status_5xx` 0%** — все ошибки ушли.
- **browse p95 10.4s** — не изменился радикально, потому что bottleneck **не в Inventory**, а в **общей задержке стека** (Gateway → ValidateToken → gRPC → сервис → PG).
- **Следующий bottleneck — ValidateToken** (Шаг 4, 560 rps в Auth) и **contention на 10 items** (Шаг 3.5 + больше items).

### Что осталось
- Шаг 4: ValidateToken кэш (убьёт 90% вызовов в Auth).
- Шаг 3.5: короткий резерв (уберёт contention, purchase p95 ↓).
- Засев 50+ unowned non-merch items (contention ↓ ×5).