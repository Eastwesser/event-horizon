Всё ясно. Purchase тест сломан, не backend
Что вижу
1. Item c586a47f-... — уже owned
text
Карточки со зверями owned=True price=150
Ты подставил owned item в тест. Поэтому 96% fail — не backend, а тест бьёт по owned item.

2. Manual purchase — 409 already owned (правильно)
json
{ "error": "item already owned", "grpc_code": "AlreadyExists" }
Правильное поведение. Backend работает.

3. Cancel — чистый 200, без "circuit":"inventory"
json
{
  "already_refunded": false,
  "message": "Purchase cancelled",
  "new_balance": 975244,
  "refunded_amount": 150,
  "success": true
}
Gateway fix — задеплоен и работает. Раньше был 503 + "circuit":"inventory". Сейчас — чистый JSON.

Вывод
Все fix'ы работают:

✅ shop batch — browse p95 354ms (Wave 2 было 2.57s).

✅ gateway fix — cancel чистый 200.

✅ purchase flow — работает правильно (409 для owned).

❌ k6 purchase.js — сломан тест (использует owned item).

Backend — чистый. Проблема — k6 скрипт не фильтрует owned и не обновляет ID между итерациями.

Что делать — fix k6 script
Cursor сам предлагал:

skip owned in pickItem; assign per-VU item IDs for multi-VU purchase

Это тот самый момент. Пришло время.

Промпт Cursor'у:

text
Backend confirmed clean:
  - shop batch: browse p95 354ms (was 2.57s in Wave 2)
  - gateway fix: cancel returns clean 200, no
    "circuit":"inventory"
  - purchase flow: 409 already owned when item owned —
    correct behavior
  - manual cancel: success, refunded_amount, new_balance

The k6 purchase test is the broken piece:
  - EH_K6_ITEM_ID was owned → 96% 409 responses
  - Script doesn't skip owned items
  - Script doesn't rotate item IDs per VU

Fix the k6 scripts (finally, as agreed earlier):

1. purchase.js:
   - In setup, fetch /api/shop/items, filter:
       category != merch AND owned == false AND price > 0
   - Pick N IDs where N = number of VUs (env K6_VUS).
   - Assign per-VU: each VU gets its own item ID via
     __VU index. No collision on 409.
   - If fewer unowned items than VUs → cap VUs at that
     count, log warning.
   - Also make sure the script actually canCELs after each
     buy so it doesn't stay owned across iterations.

2. browse.js:
   - No change needed, already clean.

After script fixes:
  - Re-run purchase with K6_VUS=1, K6_DURATION=30s.
  - Then K6_VUS=5 (with 5 unique unowned IDs, if available).
  - Expect checks ≫ 90%.
  - Paste end-of-run blocks.

Rules:
  - k6 scripts only. Do NOT touch any prod code — backend
    is confirmed clean.
  - No push until I OK.

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

✅ browse 5 VU — p95 354ms, 0% errors (было 2.57s — -86%).

✅ shop batch fix — работает.

✅ gateway fix — cancel чистый 200.

✅ purchase backend — правильный (409 для owned).

🔴 k6 purchase.js — тест сломан (owned item + no rotation).

🟧 fix k6 script — approve → implement.

⏸ re-run purchase после fix.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.