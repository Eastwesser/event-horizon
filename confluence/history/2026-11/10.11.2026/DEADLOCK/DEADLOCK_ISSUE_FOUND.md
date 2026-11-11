1. Deadlock в Shop — корневая причина найдена
Смотрю postgres_repo.go:

Purchase (строка 156–169):

go
tx, err := r.db.BeginTx(ctx, nil)           // 156
_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, ...)  // 163 — лок на inventory
_, err = tx.ExecContext(ctx, `INSERT INTO purchases (user_id, item_id, price) VALUES ($1, $2, $3)`, ...)  // 169 — лок на purchases
Cancel (строка 278–344):

go
tx, err := r.db.BeginTx(ctx, nil)           // 278
err = tx.QueryRowContext(ctx, `... FOR UPDATE OF p ...`)  // 289-297 — лок на purchases
qerr := tx.QueryRowContext(ctx, `...`)      // 302 — ещё один лок
res, err := tx.ExecContext(ctx, `...`)      // 326 — лок на inventory
Вот оно. purchase идёт: inventory → purchases. cancel идёт: purchases → inventory. Противоположный порядок блокировок → deadlock при одновременном выполнении на одной паре (user_id, item_id). Это классический deadlock, ровно как в учебнике.

Плюс cancel использует FOR UPDATE OF p в запросе с JOIN (строка 297) — если JOIN задевает inventory и items, порядок блокировок зависит от плана запроса, что делает его непредсказуемым.

Фикс (без переписывания саги):

В cancel сначала брать лок на inventory (или items), потом на purchases. Или наоборот, но одинаково в обоих методах.

Или: в purchase перед INSERT INTO inventory сделать SELECT ... FROM items WHERE id = $1 FOR UPDATE — тогда оба метода начинают с items (общий лок), и deadlock исчезает.

Или: обернуть оба в SET LOCAL lock_timeout = '2s' — deadlock не исчезнет, но будет быстрый fail вместо 30s зависания.

Приоритет: P0. Deadlock на 20 VU — это не потолок, это баг. В проде при 5–10 одновременных покупках одного item — тоже словят.

## Результат фикса (10.11.2026, после патча RefundPurchase)

### Что сделали
- `RefundPurchase`: якорь `SELECT 1 FROM items WHERE id = $1 FOR UPDATE` **первым**,
  `JOIN items` убран из `SELECT purchases` (category — отдельным `SELECT`),
  `lock_timeout = 2s`.
- `PurchaseItemWithStock` — не трогали (уже правильный порядок).

### Проверка (P=50 / P=100 / P=200)

| Профиль | До deadlock fix | После deadlock fix |
|---|---|---|
| P=50  | 0.67% failed, browse 1894ms, purchase 2264ms | 1.11% failed, browse 4222ms, purchase 4397ms |
| P=100 | **9.35% failed**, browse **68674ms**, 5xx 7.17% | **0.69% failed**, browse 5057ms, 5xx **0%** |
| P=200 | 3.42% failed, browse 17309ms, **deadlock 40P01** | 2.87% failed, browse 9477ms, **deadlock 0** |

### Выводы
- **Deadlock 40P01 исчез полностью** — все профили без единого deadlock.
- **P=100 — самый большой выигрыш**: 9.35% → 0.69% failed, browse 68s → 5s, `No available backends` 0.
- **P=200 browse ×1.8 лучше** (17.3s → 9.5s), purchase ×1.5 лучше.
- **Новый сигнал — `55P03 lock_timeout`** (1 раз в P=200): `cancel` ждал `items FOR UPDATE` > 2s, отвалился за 2s. Это **лучше** deadlock: контролируемо, сага не рвётся.
- **P=50 стал чуть хуже по latency** (browse 1.9s → 4.2s) — это **цена сериализации** через `items FOR UPDATE`. 10 items на 5 purchase-VU = 2 VU на item, contention. Лечится **больше items** (Шаг 5).

### Что осталось (не в этом фиксе)
- **`bulkhead full`** (Shop) — Шаг 3 (MaxConcurrent 32 → 128+).
- **`circuit inventory`** — Шаг 3/5 (bulkhead + GIN-индексы).
- **`purchase p95` растёт** из-за contention на 10 items — Шаг 5 (засеять 50+ unowned non-merch).