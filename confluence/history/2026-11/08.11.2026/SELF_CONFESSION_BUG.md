⚠️ Стоп. Вижу баг в моём же диффе. Не запускай.
Ты скинул лог до фикса (или это старый билд?). Но главное — я вижу, что мой дифф неполный. Разберём.

🔍 Что показывает твой лог
До фикса (то, что ты скинул)
text
PUT {"name":"Тест v2","version":0}   →  attributes: {}, stock: 0, version: 2
attributes → {} (обнулились) ← баг P0-10

stock → 0 (обнулился) ← баг P0-10

price → 100 (сохранился!) ← странно

description → "desc" (сохранился!) ← странно

images → ["/uploads/x.png"] (сохранились!) ← странно

version → 2 (было 0 → стало 2) ← странно

Стоп. Почему price, description, images сохранились, а attributes и stock — нет?

Это значит, что в converter.ItemFromUpdateRequest что-то уже мержит. Проверим по коду, который ты прислал:

go
func ItemFromUpdateRequest(req *pb.UpdateItemRequest) *model.Item {
	attrs := map[string]interface{}{}
	if req.Attributes != nil {
		attrs = req.Attributes.AsMap()
	}
	return &model.Item{
		ID:          req.Id,
		AuthorID:    req.AuthorId,      // "" → перезапишет!
		Type:        req.Type,          // "" → перезапишет!
		Name:        req.Name,          // "Тест v2"
		Description: req.Description,   // "" → перезапишет!
		Price:       req.Price,         // 0 → перезапишет!
		Stock:       int(req.Stock),    // 0 → перезапишет!
		Attributes:  attrs,             // {} → перезапишет!
		Images:      req.Images,        // nil → перезапишет!
		Version:     int(req.Version),  // 0
	}
}
Этот конвертер точно обнуляет всё. Значит, price, description, images не должны были сохраниться. Но в логе они сохранились. Два варианта:

Ты уже применил мой дифф (или что-то похожее), и лог — после фикса, но с неполным мержем: attributes и stock всё ещё обнуляются.

В repository.UpdateItem есть частичный мерж (COALESCE / NULLIF) — но тогда attributes и stock тоже должны были сохраниться.

Смотри на version: 2. Было 0, стало 2. Это не мой дифф (version из req был 0, значит service.UpdateItem подтянул бы текущую version из БД и repo.UpdateItem увеличил бы на 1 → стало бы 1, не 2). Значит, version уже была 1 к моменту PUT (ты делал PUT раньше?), или repo.UpdateItem инкрементит дважды.

🚨 Мой дифф неполный. Что я пропустил
В моём диффе attributes мержатся через if req.Attributes != nil. Но structpb.Struct в proto — это *Struct, и когда фронт не шлёт attributes, req.Attributes == nil. Тогда item.Attributes остаётся из existing — должно работать.

А stock — if req.Stock >= 0. req.Stock == 0 (дефолт proto для int32). 0 >= 0 = true → item.Stock = 0. Вот баг в моём диффе!

То же с price: if req.Price > 0. req.Price == 0 → 0 > 0 = false → item.Price остаётся из existing (правильно). Но если existing.Price == 0 (бесплатный товар) — мы не сможем его выставить. Известное ограничение, я его отметил.

А version: 2 — это отдельный вопрос. Может, ты делал PUT до этого, и version уже была 1. Или repo.UpdateItem делает version = version + 1 дважды. Это BUG-2026-11-08-4, надо проверить.

✅ Исправленный дифф (мой предыдущий был с багом)
Файл: services/inventory/internal/handler/grpc_handler.go → UpdateItem

Стало (исправленная версия):

go
func (h *GRPCHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.ItemResponse, error) {

    existing, err := h.service.GetItem(ctx, req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, "item not found")
    }
    if err := checkAuthorAccess(ctx, existing.AuthorID); err != nil {
        return nil, err
    }

    // P0-10: merge into existing so partial PUT (e.g. only name)
    // does not wipe attributes / images / price / stock.
    item := *existing

    if req.Type != "" {
        item.Type = req.Type
    }
    if req.Name != "" {
        item.Name = req.Name
    }
    if req.Description != "" {
        item.Description = req.Description
    }
    // Price: proto3 int32/float64 default 0. We can't distinguish "not sent" from "sent 0".
    // Accept only positive; 0 means "keep existing" (known limitation, BUG-2026-11-08-2).
    if req.Price > 0 {
        item.Price = req.Price
    }
    // Stock: same issue. 0 means "keep existing" (known limitation, BUG-2026-11-08-3).
    // NOTE: this also means "set stock to 0" is impossible via PATCH. Accept for now.
    if req.Stock > 0 {
        item.Stock = int(req.Stock)
    }
    if req.Images != nil {
        item.Images = req.Images
    }
    if req.Attributes != nil {
        item.Attributes = req.Attributes.AsMap()
    }
    if req.Version > 0 {
        item.Version = int(req.Version)
    }

    if err := validateItem(&item); err != nil {
        return nil, err
    }

    if err := h.service.UpdateItem(ctx, &item); err != nil {
        return nil, mapInventoryErr(err, "failed to update item")
    }

    protoItem, err := converter.ItemToProto(&item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}
Что изменилось от моего предыдущего диффа:

if req.Stock >= 0 → if req.Stock > 0. Иначе stock=0 (дефолт proto) затирает существующее значение. Это компромисс: нельзя через PATCH выставить stock=0. Записываем как BUG-2026-11-08-3.

if req.Price > 0 — оставил как было (тоже компромисс, BUG-2026-11-08-2).

Про attributes: if req.Attributes != nil — оставил. structpb.Struct — указатель, nil когда не прислали. Должно работать. Если в твоём логе attributes всё ещё {} после фикса — значит, req.Attributes не nil, а &Struct{} (пустая структура). Это возможно, если фронт шлёт "attributes": {} явно. Это отдельный баг фронта, BUG-2026-11-08-5.

Про version: 2: это не от моего фикса. Мой фикс ставит item.Version = int(req.Version) если > 0. Если req.Version == 0 — service.UpdateItem подтягивает из БД и repo.UpdateItem инкрементит. Было 0 (или 1?) → стало 2. Проверь в БД:

bash
docker exec -it event-horizon-postgres-inventory psql -U eventhorizon -d eventhorizon_inventory \
  -c "SELECT id, name, stock, version, attributes, updated_at FROM inventory_items WHERE id='45d731da-590c-43fa-8153-b9c3c2744124';"
Если version = 2 — либо было 1 до PUT, либо repo.UpdateItem инкрементит дважды. BUG-2026-11-08-4.

🐛 Баги, найденные по ходу (готовые блоки для CLEAR_TODO)
markdown
### 🐛 BUGS FOUND DURING EXECUTION

#### BUG-2026-11-08-1 — `UpdateItem` не выставляет `updated_at`
- **Файл:** `services/inventory/internal/repository/postgres_repo.go` → `UpdateItem`
- **Симптом:** в ответе PUT `updated_at` изменился (`17:59:39` → `17:59:52`), значит **работает** через SQL. Снять с учёта.
- **Статус:** [закрыто — работает]

#### BUG-2026-11-08-2 — `price=0` неотличим от «не передали»
- **Где:** `handler.UpdateItem`, `if req.Price > 0`
- **Симптом:** нельзя через PATCH сделать товар бесплатным.
- **Решение (потом):** `*float64` в proto `UpdateItemRequest` (proto3 `optional double price`).
- **Статус:** [известное ограничение]

#### BUG-2026-11-08-3 — `stock=0` неотличим от «не передали»
- **Где:** `handler.UpdateItem`, `if req.Stock > 0`
- **Симптом:** нельзя через PATCH выставить `stock=0`, если текущий `stock=5`.
- **Решение (потом):** `*int32` в proto `UpdateItemRequest` (proto3 `optional int32 stock`).
- **Статус:** [известное ограничение]

#### BUG-2026-11-08-4 — `version` прыгает на 2 за один PUT
- **Где:** `repository.UpdateItem` / `service.UpdateItem`
- **Симптом:** в логе `version` было `0`, стало `2`. Ожидалось `1`.
- **Как проверить:** `SELECT version FROM inventory_items WHERE id='...'` до и после PUT.
- **Гипотезы:** (a) `version` уже была `1` до PUT; (b) `repo.UpdateItem` делает `version = version + 1` дважды; (c) `service.UpdateItem` инкрементит при `Version <= 0`, а `repo` — ещё раз.
- **Статус:** [проверить]

#### BUG-2026-11-08-5 — `attributes: {}` от фронта затирает существующие
- **Где:** фронт (`InventoryEditModal.tsx`?) шлёт `"attributes": {}` явно
- **Симптом:** если фронт отправляет пустой объект, `req.Attributes != nil` (это `&Struct{}`, не `nil`), и мерж затирает существующие атрибуты.
- **Решение:** либо фронт не шлёт `attributes` если не менял; либо в handler проверять `len(req.Attributes.AsMap()) > 0` перед мержем (но тогда нельзя «очистить все атрибуты» — компромисс).
- **Статус:** [проверить]
