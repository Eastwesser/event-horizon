Где мы
Закрыто:

✅ proto3 фиксы (stock, price, edit prefill)

✅ Image URL (карточки, превью, clear)

✅ Inventory filter crash — фронт + бэкенд

✅ Admin panel 2a + 2b + 2c

✅ Nav, burger, Block 3

✅ Все 3 gateway пересозданы на свежем образе

Cursor в работе (TODO_3.md):

🟧 Подтвердить бэкенд-фиксы в образе (inventory / shop)

🟧 Timeout 15000ms — проверить, исчез ли

🟧 shop/items twin — фикс

🟧 Аудит других proto3 null-мест

Отложено:

🟧 2d — Analytics

🟧 2c v2, proto3 medium, gin.H

🟧 File upload, navbar refactor

🟧 DNS fix для docker push

Ждёт:

🟧 Коммит dev/frontend → потом main


-----

UPDATED:

✅ Что подтвердил
1. Backend-фиксы в источнике
Инвентарь — уже в живом образе:

go
// gateway.go Ln 1380–1390
items := resp.GetItems()
if items == nil {
    items = []*inventoryPb.Item{}
}
c.JSON(http.StatusOK, gin.H{"items": items, "total": resp.GetTotal()})
Shop — в источнике, но НЕ в живом образе:

go
// gateway.go Ln 799–805
items := resp.GetItems()
if items == nil {
    items = []*shopPb.Item{}
}
c.JSON(http.StatusOK, items)
⚠️ Важно: c.JSON(http.StatusOK, items) — это не gin.H{"items": items}. То есть shop возвращает голый массив, а не объект. Это может ломать фронт, если он ждёт {items: [...]}. Cursor пишет «Frontend coerced» — значит фронт это уже обрабатывает. Но стоит удостовериться, что это намеренно, а не случайно.

2. Timeout — ✅ исчез
Фильтр	HTTP	Время
ALL	200	~0.05–0.7s
брелок	200	~0.07s
картина	200	~0.05s
фенечка	200	~0.04s (len=0, не null)
15s таймаут ушёл. Все три gateway свежие → балансер работает. ✅

3. Shop twin — фронт уже защищён
api.ts → getShopItems() нормализует null → []

shopStore.ts → не падает, использует []

ShopWithInfiniteScroll → использует inventoryApi.searchItems (уже coerce)

До rebuild фронт не крашится, но API всё равно возвращает null.

4. Аудит — ещё 5 потенциальных twin'ов
Endpoint	Empty → null?	FE guard?
shop/items	было yes, теперь fix в source	yes (новый)
shop/inventory	нет	yes
inventory/items	нет	yes
authors	yes если Authors nil	?? []
history/events	yes	?? []
analytics/dau	yes	|| []
analytics/retention	yes	retention?.points
leaderboard/...	yes	Array.isArray
inventory/stats	maps могут быть null	?? {}
5 twin'ов. Все защищены на фронте, но бэкенд всё ещё возвращает null.

Что осталось сделать
🔴 Шаг 1 — Rebuild gateway (это финальный шаг)
Ты говорил, что делаешь make rebuild-services SVC=gateway, потом up -d gateway. Теперь то же самое, но для всех трёх:

bash
make rebuild-services SVC=gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d --force-recreate gateway gateway-2 gateway-3
Что это даст:

shop/items fix попадёт в живой образ.

Пустые категории магазина будут возвращать [], не null.

После rebuild — проверить:

bash
curl 'http://localhost:8079/api/shop/items?category=nonexistent_cat_xyz' \
  -H "Authorization: Bearer <admin_token>"
Ожидание: [] или {"items":[]}, не null.

🟧 Шаг 2 — Разобраться с c.JSON(http.StatusOK, items) в shop
Это важно. Shop возвращает голый массив, а inventory — gin.H{"items": ...}. Разная форма для похожих эндпоинтов.

Вопрос к Cursor'у:

text
Clarify the shop GetItems response shape:

  Current fix:
    c.JSON(http.StatusOK, items)   // bare array

  But inventory uses:
    c.JSON(http.StatusOK, gin.H{"items": items, "total": ...})

  Was a bare array intentional for shop? The frontend coerces
  it now, but is this the contract we want long-term?

  If it should match inventory's pattern (wrapped in gin.H),
  fix it in the same rebuild.

  Report the intended contract before changing anything.
Это может быть намеренно (если фронт так ждёт) или ошибкой. Стоит уточнить.

🟧 Шаг 3 — Аудит 5 twin'ов
Cursor уже составил таблицу. Все 5 защищены на фронте. Но бэкенд — источник null.

Что делать: не срочно, но стоит зафиксировать:

authors → if nil { Authors = []*authorsPb.Author{} }

history → то же

analytics/dau → то же

analytics/retention → то же

leaderboard → то же

Это отдельная волна. Можно после коммита.

Промпт для Cursor — финальный
text
Two things before the checkpoint commit.

1. REBUILD + RECREATE all three gateways with the shop fix:
     make rebuild-services SVC=gateway
     docker compose --env-file .env -f
       deployments/docker-compose.cluster.yml
       up -d --force-recreate gateway gateway-2 gateway-3

   Verify:
     curl 'http://localhost:8079/api/shop/items?category=profile_theme'
       with admin bearer → must return [] or {"items":[]}, not null.
     curl 'http://localhost:8079/api/shop/items?category=nonexistent_cat_xyz'
       → same expectation.

   Report the actual response bodies.

2. CLARIFY the shop response shape:
   Current fix uses `c.JSON(http.StatusOK, items)` — bare array.
   Inventory uses `gin.H{"items": items, "total": ...}`.

   Question: is the bare array the intended contract for shop,
   or should it be wrapped to match inventory?

   Report your recommendation. Do NOT change the shape unless
   we agree — the frontend is already coerced to handle both.

DO NOT TOUCH
  - Game mechanics.
  - Nav, footer, role gating.
  - Anything from prior waves.

OUTPUT
  - Rebuild log (last 20 lines).
  - Response bodies from the two curl checks.
  - Your recommendation on the shop response shape.
Что я держу в карте
Закрыто:

✅ proto3 фиксы (stock, price, edit prefill)

✅ Image URL

✅ Inventory filter crash — бэкенд + фронт

✅ Admin panel 2a + 2b + 2c

✅ Nav, burger, Block 3

✅ Все 3 gateway на свежем образе

✅ Timeout исчез — 0.04–0.07s на все фильтры

✅ shop/items fix в source — осталось rebuild

В работе:

🔴 Rebuild gateway — применить shop fix в живом образе

🟧 Уточнить форму ответа shop (голый массив vs gin.H)

Отложено:

🟧 5 backend twin'ов (authors, history, analytics, leaderboard)

🟧 2d — Analytics

🟧 2c v2, proto3 medium, gin.H

🟧 File upload, navbar refactor

🟧 DNS fix

Ждёт:

🟧 Коммит dev/frontend → потом main