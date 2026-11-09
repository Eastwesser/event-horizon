# PROTOJSON_BUG — `google.protobuf.Struct` не бывает `nil`

**Дата:** 09.11.2026
**Сервис:** Inventory
**Симптом:** `PUT /api/v1/inventory/items/:id` без поля `attributes` в теле — обнулял `attributes` в БД. `POST` работал, `PUT` — нет.
**Диагноз:** Gateway сериализует proto3 `google.protobuf.Struct` через `protojson`. Отсутствующее поле в JSON → **не `nil`, а `&structpb.Struct{}`** (пустая структура). В Go-коде проверка `if req.Attributes != nil` проходит, `AsMap()` возвращает `map{}`, `item.Attributes = map{}` затирает существующее.
**Фикс:** `if req.Attributes != nil && len(req.Attributes.AsMap()) > 0 { ... }`.
**Ограничение:** очистить `attributes` через PUT нельзя (пустой `&Struct{}` = «не трогать»). Для полной очистки — отдельный PATCH или `optional` в proto. `BUG-2026-11-08-5`.
**Файл:** `services/inventory/internal/handler/grpc_handler.go` → `UpdateItem`.

## Урок

`google.protobuf.Struct`, `google.protobuf.Timestamp`, `google.protobuf.StringValue` и другие well-known types в proto3 — **не имеют `nil`-семантики при `protojson` unmarshal**. Отсутствующее поле становится **нулевым значением структуры** (`&Struct{}`, `&Timestamp{}`, `&StringValue{Value: ""}`).

Проверка `if req.Field != nil` **не отличает** «не передали» от «передали пустое». Для PATCH-семантики нужны **`optional` поля** (proto3 `optional`) или явные указатели в DTO.

## Что проверить в других местах

- `Images` (`repeated string`) — `nil` vs `[]`: proto3 `repeated` **всегда `[]`**, не `nil`. Проверка `if req.Images != nil` **может быть всегда `true`**. В нашем тесте `images` сохранились — значит, Gateway прислал `nil` (не `[]`). Но это **не гарантировано** для всех клиентов.
- `AuthorID` в `UpdateItemRequest` — если Gateway присылает пустую строку, может затирать `author_id`. Сейчас не трогаем (мерж в handler), но при рефакторинге — учесть.

## TODO

- [ ] `Images` в `UpdateItem` — проверить, приходит ли `nil` или `[]`. Если `[]` — фикс `len(req.Images) > 0`.
- [ ] `AuthorID` в `UpdateItem` — проверить, не затирает ли.
- [ ] При рефакторинге proto → `optional` поля для PATCH-семантики.