Что нашли
Bottleneck — не индексы, а N+1 запросы
shop_service.go Ln 93–97:

go
// Проверяем owned для каждого товара
for i := range items {
    owned, _ := s.pgRepo.IsItemOwned(ctx, userID, items[i].ID)
    items[i].Owned = owned
}
305 round-trips на /api/shop/items, даже при Redis hit.
EXPLAIN подтверждает: индексы в порядке (GetItems 3–38ms, IsItemOwned 0.07ms). Проблема — количество запросов.

Fix — batch SELECT:

sql
SELECT item_id FROM inventory WHERE user_id = $1
Один запрос вместо 305. Redis или in-memory set для проверки — по item_id.

Purchase fail — двойная запись response
Gateway вызывает inventory после shop buy/cancel. Если inventory CB открыт:

throughBreaker пишет 503.

Handler всё равно пишет success JSON.

Клиент видит 503 (fail), но покупка закоммичена.

k6 skip'ает cancel → stuck already owned.

Fix — если inventory CB открыт после shop success:

НЕ писать 503.

Логировать, вернуть shop 200 (та же семантика, как другие inventory errors).

Логика: shop уже сделал своё дело. Inventory — вторичный (owned обновление). Ошибка inventory не должна отменять purchase клиенту.

Ответ Cursor'у
text
AGREE on (1)+(2). Implement both locally. No push.

1. Shop fix — batch owned lookup:
   - Replace the per-item loop with one query:
       SELECT item_id FROM inventory WHERE user_id = $1
     (or equivalent repo method).
   - Build an in-memory set of owned item_ids.
   - Set items[i].Owned = set.Has(items[i].ID).
   - No index migration.
   - Keep Redis cache path intact if present.
   - Confirm the shop test suite still passes (the shop
     service is already at ~84% — don't break it).

2. Gateway fix — do NOT override shop success on
   inventory CB failure:
   - After shop returns success (buy or cancel), if
     inventory reserve/release fails (CB open, timeout,
     ErrBulkheadFull, Unavailable), log the error and
     return the shop result unchanged (200 + success body).
   - Never write a 503 after shop has committed.
   - Same for cancel path.

After both fixes:
  - Re-run k6 browse (5 VU) and purchase (1 VU with
    unowned item).
  - Compare:
      browse: p95 before 2.57s → expect big drop
      purchase: errors before ~96% → expect near 0
  - Paste end-of-run blocks.
  - Then we do the optional k6 script fixes (skip owned,
    per-VU IDs) as a separate commit.

Rules:
  - One PR for shop fix, one for gateway fix (or two
    commits if one PR — your call, but separable).
  - tsc + go test clean.
  - Do NOT touch indexes, migrations, C1–C3, C4, hero,
    nav, games, seed, refund.
  - Do NOT touch k6 scripts yet.

Standing rule: you don't push; you write commit messages.
Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

PR #2 — открыт (CI coverage gate).

Wave 4 (2) — k6:

✅ browse 5 VU baseline — p95 2.57s, ~4.1 RPS, 0% errors.

✅ shop curl — 0.67s / 52KB.

✅ EXPLAIN — индексы в порядке.

🔴 bottleneck найден: N+1 (IsItemOwned в цикле).

🔴 purchase — 96% fail из-за gateway double-write.

🟧 fix (1) + (2) — approve → implement.

🟧 script fixes — после clean baseline.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.