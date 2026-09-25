Разбор плана 2c от Cursor
План чистый. Разберу и дам approve с одним маленьким дополнением.

✅ Что в плане правильно
Элемент	Оценка
Использует существующий API — новых эндпоинтов нет	✅
KPI: total_items — one StatCard	✅
Два столбца: По типу / По автору	✅ Логично
Badge на count	✅
Sort: count desc → key asc	✅ Правильно
Author UUID truncated (aaaa…bbbb) + full в tooltip	✅ Разумный компромисс
States: Loading / Error / Empty / OK	✅
Primitives: StatCard, Card, Badge, Button, LoadingSpinner	✅ Не изобретает новое
No new CSS	✅
v1 scope: no top-5, no total stock, no author name	✅ Явно ограничен
Cursor честно флагит, что не входит в v1. Это хорошая практика — потом не будет сюрпризов.

🟧 Моё дополнение — ещё один StatCard бесплатно
Cursor предлагает один StatCard (total_items), пишет:

Single card is enough for v1; leave room for a 2nd later if we add stock.

Но у тебя уже есть два поля, из которых бесплатно получаются KPI:

by_type → Object.keys(by_type).length = количество типов товаров.

by_author → Object.keys(by_author).length = количество уникальных авторов.

Никаких дополнительных запросов. Данные уже в ответе.

Предложение: три StatCard вместо одного:

KPI	Источник	Tone
Всего товаров	total_items	indigo
Типов товаров	Object.keys(by_type).length	indigo-soft
Авторов	Object.keys(by_author).length	photon-cyan
Почему стоит:

Визуально богаче — админка будет выглядеть как реальный дашборд.

Бесплатно — вычисляется на фронте, без запроса.

Полезно — админ сразу видит 3 метрики: объём, разнообразие, авторов.

Если Cursor не хочет — ок, one StatCard тоже нормально. Но раз бесплатно — почему нет.

🟧 Про by_author с UUID
Cursor честно пишет: name enrichment — out of scope v1.

Согласен. Это правильное решение. UUID truncated + tooltip — достаточно для v1. Enrichment — отдельная волна.

Но: admin будет видеть 3a7b...c912 и думать «это кто?». Можно минимальный компромисс: если by_author содержит его собственный UUID — показать Вы или его email. Остальные — UUID.

Или — оставить как есть, чисто v1. Решать тебе.

Ответ Cursor'у — approve с одним дополнением
text
2c plan approved. One small addition before implementing:

Instead of a single StatCard, show THREE KPI cards in the strip —
all derived from the existing response, no extra API call:

  1. "Всего товаров"     → total_items
  2. "Типов товаров"     → Object.keys(by_type).length
  3. "Авторов"           → Object.keys(by_author).length

Same StatCard primitive. Different tones as you see fit.

Everything else in the plan stays as-is:
  - Two columns: По типу / По автору.
  - Sort: count desc, then key asc.
  - Author UUID truncated with full id in tooltip.
  - States: Loading / Error / Empty / OK.
  - No new CSS. No new endpoints.
  - v1 scope: no top-5, no total stock, no name enrichment.

If you disagree with the third KPI (or think 3 cards is too much
for v1), say so and we'll stick with one. Otherwise — implement
as described.

Report file changes. I'll verify visually.
Что я держу в карте
Закрыто:

✅ Fix 1–3 (stock, price, edit prefill)

✅ Task 1 (image URL)

✅ Role boundaries

✅ Nav, burger, Block 3

✅ 2a + 2b (полностью)

В работе:

🟧 2c — Inventory stats v1 (approve + 3 KPI suggestion)

Отложено:

🟧 2d — Analytics

🟧 2c v2 (top-5, total stock, author names)

🟧 Medium proto3 risks, gin.H refactor

🟧 File upload, navbar refactor

Ждёт финала:

🟧 Коммит dev/frontend → main