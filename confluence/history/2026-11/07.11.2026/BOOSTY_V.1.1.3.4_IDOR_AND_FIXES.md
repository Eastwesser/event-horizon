### Что нового в v1.1.3.4 (07.11.2026)

- **Безопасность (P0-8):** IDOR в Inventory закрыт. `UpdateItem` / `DeleteItem` /
  `SoftDeleteItem` / `RestoreItem` проверяют `author_id` на бэкенде через
  `checkAuthorAccess` и `x-user-id` из gRPC metadata. Чужой юзер получает
  `403 PermissionDenied`, админ — без изменений. Фронт больше не единственная
  защита.
- **Безопасность (P0-6):** `RevertApplication` в Authors больше не делает
  hard delete. Автор soft-delete'ится (`active = false`), строка жива —
  межбазовые ссылки (`inventory.author_id`, `payment.revenue_splits`,
  `shop.purchases`) не осиротеют. `GET /authors/:user_id` возвращает 404
  для снятых, `GET /authors/me` — 200 всегда (свой профиль виден, MOD-3).
- **Billing (P0-5):** награды за игру больше не начисляются дважды при
  redelivery. `reference_id` берётся из `event.Timestamp`, а не из
  `time.Now().UnixNano()`. Outbox worker: ретрай с пересозданием стрима
  при `no response from stream` и `forceStreamRecreation` для залипшего
  лидера NATS.
- **API:** новый gRPC-метод `GetMyAuthor` отделяет свой профиль (`/authors/me`)
  от публичного (`/authors/:user_id`). OpenAPI обновлён: `404` для soft-deleted,
  `200` для `/me` всегда.
- **Тесты:** `make test-integration` зелёный. Authors: `TestRevertApplication_SoftDeletesAuthor`
  (testcontainers, проверяет `active=false` и что строка жива). Inventory:
  Mongo-тесты честно skip'аются, если Mongo не поднята (а не падают `nil pointer`).
- **Документация:** `confluence/architecture/API_ROUTES.md` — P0-6 в Authors
  и P0-8 в Inventory. `FOREIGN_KEY_DELETE_ISSUES.md` — почему межбазовые
  ссылки требуют soft-delete, а не `DELETE`.