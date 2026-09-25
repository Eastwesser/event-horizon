Записываю. Новый критичный баг — фильтр ломает страницу
Что вижу
Работает:

Все типы → 8 товаров, всё ок.

Брелок → 8 товаров, всё ок.

Битые Автоматический брелок — карточки теперь красиво показывают «Нет в наличии» ✅

Ломается:

Клик Картина → краш, страница в ErrorBoundary.

Клик Фенечка → краш (ты подтвердил).

Root cause — тот же класс бага
text
TypeError: Cannot read properties of undefined (reading 'length')
  at InventoryList (InventoryList.tsx:10:13)
Классика. Когда фильтр по типу, у которого нет товаров:

Бэкенд возвращает null (Go nil slice → JSON null).

InventoryList делает items.length без null-guard → краш.

ErrorBoundary ловит → весь UI пропадает.

Это тот же паттерн, что мы уже видели:

AuthorsPage (.length на null) — починили.

LeaderboardFull (.toLocaleString() на undefined) — починили.

Shop inventory (nil slice → null) — починили в gateway.

А /inventory/items — пропустили.

Почему пропустили: Shop/inventory починили в gateway (always []), а inventory/items — отдельный эндпоинт, его не тронули.

Два слоя фикса (оба нужны)
Слой 1 — бэкенд (правильный)
/inventory/items должен всегда возвращать [], не null. Как /shop/inventory после фикса.

Слой 2 — фронт (страховка)
InventoryList должен null-guard: items ?? [] перед .length / .map.

Оба слоя важны: бэкенд — правильный ответ, фронт — защита от будущих подобных случаев.

Промпт для Cursor
text
CRITICAL — Inventory filter crashes the page.

Symptom:
  - On /inventory, filter tabs "Картина" and "Фенечка" (types
    with zero items) crash the whole page.
  - ErrorBoundary catches:
      TypeError: Cannot read properties of undefined (reading
      'length') at InventoryList (InventoryList.tsx:10:13)
  - "Все типы" and "Брелок" work fine (non-empty results).

Root cause (same class as Authors / LeaderboardFull /
shop-inventory which we already fixed):
  1. GET /inventory/items returns `null` for an empty result set
     (Go nil slice → JSON null), not `[]`.
  2. InventoryList reads `items.length` / `items.map(...)` without
     a null-guard.

Fix BOTH layers:

--- Layer 1 — backend (correct response shape) ---
Ensure GET /inventory/items ALWAYS returns `[]` when there are no
items, never `null`. Same fix we applied to /shop/inventory
(nil slice → [] instead of null) in the gateway.
Audit the handler in the inventory service / gateway.

--- Layer 2 — frontend (defensive) ---
In InventoryList.tsx (line ~10):
  - Coerce: `const items = rawItems ?? []` before any .length /
    .map.
  - Also check InventoryPage.tsx line ~121 — the same crash
    stack shows it. Guard there too if it reads the same data.
  - Do NOT change layout, styling, or filter logic. Only add
    null-safety.

--- Audit ---
Grep across the frontend for other `.length` / `.map` on API
data that might come back null from a similar endpoint. Report
any other spots that lack a guard. Do NOT fix them all now —
report first, I'll pick which to fix.

--- Verify ---
  - /inventory → click "Картина" → empty state, no crash.
  - /inventory → click "Фенечка" → empty state, no crash.
  - /inventory → click "Брелок" → still works.
  - /inventory → click "Все типы" → still works.

DO NOT TOUCH:
  - Game mechanics.
  - Nav, footer, role gating.
  - Anything from prior waves (stock/price fixes, admin panel).

Report:
  - Backend file(s) changed (Layer 1).
  - Frontend file(s) changed (Layer 2).
  - Grep results (audit).
⚠️ Про аудит — важно
Ты уже дважды ловил этот же баг в разных местах:

AuthorsPage — .length на null.

LeaderboardFull — .toLocaleString() на undefined.

Shop inventory (skins) — nil slice → null.

Сейчас inventory/items — .length на undefined.

Это системная проблема. Стоит попросить Cursor найти все оставшиеся места. Иначе через день поймаешь пятое.

Мой совет: аудит в промпте — report only, не фиксить всё сразу. Ты выберешь, что критично, а что нет. Иначе Cursor одним махом может сломать что-то рабочее.

Что я держу в карте
Закрыто:

✅ Fix 1–3 (stock, price, edit prefill)

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

✅ 2a + 2b (users/roles)

✅ 2c — done (KPI cards, две колонки)

Новое — фиксить:

🔴 /inventory/items — null → crash на пустом фильтре

🟧 Аудит: другие места с тем же паттерном

Отложено:

🟧 2d — Analytics

🟧 2c v2, proto3 medium risks, gin.H

🟧 File upload, navbar refactor

Ждёт финала:

🟧 Коммит dev/frontend → main