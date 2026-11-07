# Authors P0-6 — RevertApplication: hard delete → soft delete

## Проблема
`RevertApplication` делал `DELETE FROM authors WHERE user_id = $1`. Hard delete
осиротяет межбазовые ссылки (FK между базами в Postgres нет):
- `inventory.author_id` (карточки мерча)
- `payment.revenue_splits` (журнал 85/15)
- `shop.purchases` / `purchase_sagas`

FK внутри `eventhorizon_authors` на `authors` нет (SQL: 0 rows), поэтому
`DELETE` не падал — он тихо проходил и оставлял висячие ссылки в других базах.
Это хуже падения: витрина и история ссылаются в никуда.

## Решение
`RevertApplication`: `DELETE` → `UPDATE authors SET active = false, updated_at = now() WHERE user_id = $1`.
Строка жива, `active = false`, межбазовые ссылки целы.

## Разделение `/me` и `/:user_id` на уровне gRPC
Раньше оба звали `GetAuthor`, и фильтр `active` в нём сломал бы `/me`
(автор не увидел бы свой снятый профиль — MOD-3). Добавлен `rpc GetMyAuthor`:
- `GetAuthor` — только `active = true`, публичный `GET /authors/:user_id` → 404 для soft-deleted.
- `GetMyAuthor` — без фильтра, `GET /authors/me` → всегда 200, включая `active = false`.

## Что НЕ менялось
- `repo.GetByUserID` — без фильтра `active` (для `/me`).
- `List` — уже фильтровал `active = true`, не трогали.
- `ApproveApplicationInTx` — уже умеет upsert (`UPDATE active=true`), не трогали.
- `gateway.go` `GET /authors/:user_id` — не трогали, фильтр в сервисе.
- `gateway.go` `POST /authors/applications/:id/approve` — уже вызывает
  `RevertApplication` при ошибке `UpdateRole`, теперь это soft-delete. Не трогали.

## Тесты
- Unit: `TestGetAuthor_HidesInactive`, `TestGetMyAuthor_ReturnsInactive`.
- Integration (`//go:build integration`): `TestRevertApplication_SoftDeletesAuthor`
  — `active = false`, `COUNT(*) = 1`.

API_ROUTES.md
В секцию Authors добавь:

markdown
| GET | /api/v1/authors/:user_id | — | Публичный профиль. **404**, если `active = false` (soft-deleted, P0-6). |
| GET | /api/v1/authors/me | author, admin | Свой профиль. **200 всегда**, включая `active = false` (MOD-3). |

TODOLISTISHE / BACKEND_TECH_DEBT
В P0-6 поменяй:

markdown
P0-6	Authors: RevertApplication hard delete — FK падает. Soft delete + check FKs	Бэкенд	...	✅ Закрыто — soft-delete через `active=false`  



---

Итог по P0-6 — что сделано
Компонент	Статус
repo.RevertApplication: DELETE → UPDATE active=false	✅
repo.GetActiveByUserID (новый, WHERE active = true)	✅
repo.GetByUserID (без фильтра, для /me)	✅
service.GetAuthor (публичный, только активные)	✅
service.GetMyAuthor (свой, включая active=false)	✅
handler.GetAuthor → GetActiveByUserID	✅
handler.GetMyAuthor (новый)	✅
handler.mapErr: AlreadyReviewed → FailedPrecondition (409), AlreadyPrivileged → PermissionDenied (403)	✅
handler.toPB: zero-time guard	✅
proto: rpc GetMyAuthor + два message	✅
gateway.go: /authors/me → GetMyAuthor (старый блок закомментирован)	✅
gateway.go: /authors/:user_id — не трогали, фильтр в сервисе	✅
docs/openapi.yaml: 404 для soft-deleted, 200 для /me	✅
Unit-тесты: TestGetAuthor_HidesInactive, TestGetMyAuthor_ReturnsInactive	✅
Integration: TestRevertApplication_SoftDeletesAuthor	✅
make test-integration	✅ зелёный
mongo_repo_test.go skip	✅
revert_integration_test.go uuid	✅
P0-6 закрыт полностью. Можно ставить ✅.