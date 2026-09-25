Разбор аудита proto3 — что нашли
Cursor провёл системный аудит. Разберу что критично, что latent, что ок.

🔴 High risk — три открытых бага
#	Где	Симптом	Что делать
1	InventoryItemDetail (страница одного предмета)	В наличии: {stock} — тот же баг, что на карточке, но на detail	Фикс
2	Item.price — везде	{price} ₽ — если price = 0, рендерится ₽ без числа	Фикс
3	InventoryEditModal — prefill stock / price	При 0 поле пустое → если сохранить, обнуляет данные	Критично
Про #3 — это опасно
Сценарий:

У предмета price = 0 (бесплатный).

Открываешь Edit → поле price пустое (не 0).

Меняешь что-то другое (например, имя).

Сохраняешь.

price уходит как undefined / 0 → данные портятся.

Это silent data corruption. Особенно больно, если автор не знал, что 0 = бесплатно, и не понял, почему после правки имени предмет стал бесплатным.

Надо фиксить срочно.

🟧 Medium — mitigated, но стоит помнить
shop.Item.price — price || 0, ok.

shop.Item.owned / available — truthy, ok.

leaderboard score / rank — ?? 0, ok.

analytics day / retention — может скипать нули в чарте. Admin-only, но если график скипает дни с 0 активностью — искажает картину.

✅ Low / safe — не трогать
Billing lamps/tickets — || 0, ok.

Subscription — gin.H, ok.

Auth total_score — gin.H, ok.

Profile lamps/tickets — || 0, ok.

🟧 Два шума в отчёте
1. AttributeError: 'list' object has no attribute 'get'
Это ошибка скрипта аудита, не приложения. Cursor пытался распарсить items как dict, а получил список. Не баг проекта.

2. authors list === n 0
Аудит не нашёл авторов — либо эндпоинт вернул пустой список, либо скрипт не смог достать. Не критично для аудита proto3.

Стратегическое замечание Cursor'а
Safest long-term: gateway maps to gin.H (or protojson with EmitUnpopulated: true) for public DTOs; short-term: coerce at read sites.

Это правильное наблюдение. Есть два подхода:

Подход	Плюсы	Минусы
A. Gateway → gin.H	Единая точка, чисто, больше никогда не кусает	Требует переписать gateway-слой (риск)
B. Coerce at read (?? 0, || 0)	Быстро, точечно, минимальный риск	Размазано по фронту, можно забыть где-то
C. protojson с EmitUnpopulated: true	Один флаг, все нули уходят	Может изменить другие места (например, null появятся)
Мой выбор: C — если есть возможность включить флаг только для публичных DTO (не для внутренних). Иначе — B (coerce), но системно — не по одному месту.

Почему не A: переписывать gateway — риск сломать то, что работает. Сейчас не время.

Промпт для Cursor
text
Audit received. Three high-risk items to fix, then we move on.

================================================================
FIX 1 — InventoryItemDetail stock rendering
================================================================
Same bug as the card: "В наличии: {stock}" renders blank when
stock is 0/omitted. Apply the same logic:
  - 0 / omitted → "Нет в наличии"
  - > 0 → "В наличии: N"
  - null → hide line

================================================================
FIX 2 — price rendering everywhere
================================================================
"{price} ₽" renders " ₽" when price is 0/omitted. Expected:
  - 0 / omitted → "Бесплатно" (or "0 ₽" — pick one and be
    consistent)
  - > 0 → "N ₽"
Apply in InventoryItemCard, InventoryItemDetail, ShopItemCard,
and anywhere else price is rendered directly.

================================================================
FIX 3 — InventoryEditModal prefill (CRITICAL, data corruption)
================================================================
Currently:
  - item.stock is undefined when 0/omitted → field shows empty.
  - item.price is undefined when 0/omitted → field shows empty.
  - If user saves without touching those fields, they submit
    empty → backend may interpret as 0 or null → data corrupted.

Fix:
  - Prefill with `item.stock ?? 0` and `item.price ?? 0` (or use
    the explicit value from the API if you can preserve it).
  - On submit, always send explicit numbers.
  - Do NOT change field layout or validation.

This is silent data corruption — verify by:
  - Editing an item with price=0, changing only the name, saving.
  - Confirm price stays 0 (not empty/null/broken).

================================================================
NOT NOW (report only, no fix)
================================================================
- Medium risks (leaderboard rank, analytics zero-days, etc.) —
  acceptable for now. Note them; do not fix.
- Long-term architecture (gin.H or protojson EmitUnpopulated) —
  note as tech debt, do NOT refactor now.

================================================================
DO NOT TOUCH
================================================================
- Anything not listed above.
- The stock fix already applied to InventoryItemCard.
- Game mechanics, nav, footer, role gating.

================================================================
PROCESS
================================================================
Fix 1, 2, 3. Verify with tsc + build. Report file changes.
Then — /admin plan (Task 2).
Что я держу в карте
Закрыто:

✅ InventoryItemCard stock

✅ Task 1 (image URL)

✅ Role boundaries

✅ Навбар, бургер, Block 3

Новое — фиксить:

🔴 InventoryItemDetail — stock

🔴 price — везде

🔴 InventoryEditModal — prefill (данные портятся)

Отложено:

🟧 Medium risks (analytics, leaderboard)

🟧 Долгосрочная архитектура (gin.H / protojson)

🟧 File upload, navbar refactor

Ждёт:

🟧 Task 2 plan (admin panel)

🟧 Коммит dev/frontend → main