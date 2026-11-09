Что важно:

AuthorID не трогаем вообще — берётся из existing. Это закрывает побочный баг: раньше author_id обнулялся, если фронт его не передал.

CreatedAt / UpdatedAt — берутся из existing. repository.UpdateItem сам поставит updated_at = now() (проверить — если нет, это отдельный баг, пишем в BUG-...).

Attributes / Images — мержим только если клиент прислал не-nil. nil = «не трогать», {} = «очистить». Это честная PATCH-семантика.

Stock >= 0 — 0 легален (последний товар). «Не передали» — это 0 в proto, что неотличимо. Пока принимаем: если фронт хочет «оставить stock», пусть шлёт текущее значение. Это известное ограничение, записываем в BUG-YYYY-MM-DD-N (см. ниже).

Price > 0 — та же история. Бесплатный товар (price=0) через PATCH не выставить. Известное ограничение.

Version — если клиент прислал > 0, используем его (оптимистичная блокировка). Если 0 — service.UpdateItem сам подтянет из БД (там уже есть эта логика).

🧪 Как проверить
bash
# 1. Создать товар
TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')

ITEM=$(curl -s -X POST http://localhost:8079/api/v1/inventory/items \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"type":"карточка","name":"Тест","description":"desc","price":100,"stock":5,"attributes":{"market_rub":25},"images":["/uploads/x.png"]}')
echo "$ITEM" | jq .

ID=$(echo "$ITEM" | jq -r '.item.id')
VER=$(echo "$ITEM" | jq -r '.item.version')

# 2. PUT только с name
curl -s -X PUT "http://localhost:8079/api/v1/inventory/items/$ID" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"name\":\"Тест v2\",\"version\":$VER}" | jq .
Ожидаем:

name = "Тест v2"

description = "desc" (сохранилось)

price = 100 (сохранилось)

stock = 5 (сохранилось)

attributes.market_rub = 25 (сохранилось)

images = ["/uploads/x.png"] (сохранилось)

author_id — не пустой, тот же, что был

version — увеличился на 1

До фикса: всё, кроме name, станет "" / 0 / {} / [] / null. Это мы уже видели в логе.


====

🐛 Баги, которые я вижу по пути (записываем, не фиксим сейчас)
Ты сказал: «Если будем находить баги — записываем, не фиксим моментально». Вот готовые блоки для CLEAR_TODO_08.11.2026.md, секция 🐛 BUGS FOUND DURING EXECUTION:

markdown
### 🐛 BUGS FOUND DURING EXECUTION

#### BUG-2026-11-08-1 — `UpdateItem` не выставляет `updated_at`
- **Файл:** `services/inventory/internal/repository/postgres_repo.go` → `UpdateItem`
- **Симптом:** после `PUT /inventory/items/:id` поле `updated_at` не меняется (или меняется неявно через триггер — надо проверить).
- **Как проверить:** `SELECT updated_at FROM inventory_items WHERE id='...'` до и после PUT.
- **Статус:** [проверить]

#### BUG-2026-11-08-2 — `stock=0` неотличим от «не передали» в `UpdateItem`
- **Где:** `handler.UpdateItem`, `if req.Stock >= 0`
- **Симптом:** нельзя через PATCH выставить `stock=0`, если текущий `stock=5`. `0` = «не передали».
- **Решение (потом):** `*int32` в `UpdateItemRequest` (proto `optional int32 stock`).
- **Статус:** [известное ограничение]

#### BUG-2026-11-08-3 — `price=0` неотличим от «не передали» в `UpdateItem`
- **Где:** `handler.UpdateItem`, `if req.Price > 0`
- **Симптом:** нельзя через PATCH сделать товар бесплатным (`price=0`).
- **Решение (потом):** `*float64` в `UpdateItemRequest`.
- **Статус:** [известное ограничение]

📝 Пока ждёшь — вот готовый блок для CLEAR_TODO (BUG-2026-11-08-4)
markdown
#### BUG-2026-11-08-4 — `version` прыгает на 2 за один PUT
- **Где:** `repository.UpdateItem` / `service.UpdateItem` / `repo.CreateItem`
- **Симптом:** в логе после создания `version: 0`, после одного PUT — `version: 2`. Ожидалось `1`.
- **Как проверить:** 
  1. `SELECT version FROM inventory_items WHERE id='...'` сразу после `POST` — это `CreateItem`. Если `0` — идём дальше.
  2. `SELECT version FROM inventory_items WHERE id='...'` после `PUT` — это `UpdateItem`. Если `2` — двойной инкремент.
- **Гипотезы:**
  - (A) `repo.UpdateItem` делает `version = version + 1`, `service.UpdateItem` тоже → двойной.
  - (B) `repo.CreateItem` ставит `version = 1`, а не `0` → `expected` был неверный.
- **Статус:** [проверить]

📝 Пока скидываешь — вот готовые блоки BUG-... для CLEAR_TODO
markdown
#### BUG-2026-11-08-5 — `attributes` обнуляются в `repo.UpdateItem`, хотя `handler` их не трогает
- **Где:** `services/inventory/internal/repository/postgres_repo.go` → `UpdateItem` (гипотеза)
- **Симптом:** в `handler.UpdateItem` `item.Attributes` сохраняется из `existing` (не трогаем, потому что `req.Attributes == nil`). Но в БД после PUT `attributes = {}`.
- **Что проверить:** `repo.UpdateItem` — возможно, `UPDATE ... SET attributes = $N` с `$N = {}` (пересобирает мапу), или есть `jsonb_set` / `COALESCE` с неверной логикой.
- **Диагностика:** `SELECT attributes FROM inventory_items WHERE id='...'` **до** PUT (должно быть `{"market_rub":25}`), **после** PUT (стало `{}`). Если до — `{"market_rub":25}`, после — `{}`, значит `handler` передал `{"market_rub":25}`, а `repo` его обнулил. Тогда баг в `repo`.
- **Статус:** [проверить]

#### BUG-2026-11-08-4 (обновление) — `version` прыгает на 2 за один PUT, гипотеза A подтверждена
- **Факты:** создание → `version: 0`. Один PUT → `version: 2`. `req.Version = 0`.
- **Где:** `services/inventory/internal/repository/postgres_repo.go` → `UpdateItem` (гипотеза)
- **Что проверить:** `UPDATE ... SET version = version + 1` — есть ли **два** таких инкремента (в двух `UPDATE` подряд, или в одном `UPDATE` дважды).
- **Статус:** [проверить]

#### BUG-2026-11-08-6 — `converter.ItemFromCreateRequest` кладёт `map{}` при `req.Attributes == nil`
- **Где:** `services/inventory/internal/converter/item.go` → `ItemFromCreateRequest`
- **Симптом:** при создании товара **без** `attributes` в БД пишется `attributes = {}`, а не `NULL`. Это может мешать логике «пустой vs отсутствующий» в `UpdateItem`.
- **Решение (потом):** `var attrs map[string]interface{}` (nil) вместо `map[string]interface{}{}`. Но это **сломает** `structpb.NewStruct(nil)` — надо проверить. Не срочно.
- **Статус:** [известное ограничение]