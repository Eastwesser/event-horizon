# Отчёт: Закрытие P0 (07.11.2026)

Что такое IDOR (в трёх предложениях)

IDOR (Insecure Direct Object Reference) — это уязвимость, при которой приложение позволяет пользователю получить доступ к чужому объекту (карточке, заказу, профилю), просто подставив его ID в запрос, потому что сервер не проверяет, принадлежит ли этот объект текущему пользователю. 

Если фронтенд скрывает кнопку «Редактировать» для чужих товаров, но API по-прежнему принимает PUT /items/{чужой_id}, любой злоумышленник может через curl изменить или удалить чужой контент. 

Лечится это только на бэкенде: перед выполнением операции сервер должен проверить, что author_id объекта совпадает с user_id из токена (или что у пользователя есть роль admin).

## 1. P0-8: IDOR в Inventory

**Проблема:** `PUT /api/v1/inventory/items/{id}` (и DELETE) позволял любому авторизованному юзеру редактировать/удалять чужие карточки. Фронт прятал кнопки, но API был открыт.

**Решение:** В `grpc_handler.go` добавлен `checkAuthorAccess`. Перед мутацией сервер достает `x-user-id` и `x-user-role` из gRPC-metadata (их прокидывает Gateway), загружает карточку из БД и сравнивает `item.AuthorID` с `x-user-id`. Если не совпало и это не админ — `PermissionDenied` (403).

**Тест:** Curl от Круглова на карточку Совы → `403 "you can only edit your own items"`. Своя карточка → 200. Запись в БД не прошла.

## 2. P0-6: Authors, hard delete → soft-delete

**Проблема:** `RevertApplication` (компенсация при ошибке `Auth.UpdateRole`) делал `DELETE FROM authors`. FK внутри базы нет, поэтому DELETE тихо проходил, оставляя висячие ссылки в других базах: `inventory.author_id`, `payment.revenue_splits`, `shop.purchases`.

**Решение:** 
- `repo.RevertApplication`: `DELETE` заменен на `UPDATE authors SET active = false, updated_at = now()`. Строка жива, межбазовые ссылки целы.
- Добавлен `repo.GetActiveByUserID` (только `active = true`) для публичного профиля.
- В proto добавлен новый метод `GetMyAuthor` (без фильтра `active`), чтобы разделить семантику: `GET /authors/:user_id` отдает 404 для снятых, а `GET /authors/me` отдает 200 всегда (чтобы автор видел, что его сняли, MOD-3).
- `gateway.go`: ручка `/authors/me` переведена на `GetMyAuthor`.
- `mapErr`: `AlreadyReviewed` → `FailedPrecondition` (409), `AlreadyPrivileged` → `PermissionDenied` (403). Раньше оба были 400.

**Тест:** `TestRevertApplication_SoftDeletesAuthor` (testcontainers). Проверяет, что после отката строка существует (`COUNT(*) = 1`) и `active = false`. `make test-integration` зелёный.

## 3. P0-5: Billing, идемпотентность наград

**Проблема:** При redelivery события награда начислялась дважды, так как `reference_id` генерировался через `time.Now().UnixNano()` (каждый раз новый).

**Решение:**
- В `di.go` `reference_id` теперь берется из `event.Timestamp` (стабильный ключ события). Fallback на `UnixNano` оставлен только для legacy-событий, где timestamp равен 0.
- В `outbox_worker.go` добавлен ретрай (5 попыток) и `forceStreamRecreation()` при ошибке `no response from stream` (лечит залипшего лидера NATS JetStream, issue #6391).
- Добавлен `SetNX` анти-спам в Redis: `score_lock:{user_id}:{game_id}:{score}` с TTL 30s.

## 4. Инфраструктура и тесты (фоновое)

- **CI (go.sum):** Починен `make test-unit`. В `balancer`, `nats-hub`, `inventory` обновлены `go.sum` (`go mod tidy`), из-за отсутствия записей для `go-spew`, `compress` и др. Также поправлены права на `go/pkg/mod`.
- **Mongo-тесты:** В `inventory` тесты Mongo честно скипаются (`t.Skipf`), если база не поднята. Раньше падали с `nil pointer`.
- **Диск (инцидент):** `/mnt/shared` был забит на 100%, гит дважды «умирал» из-за пустых объектов в `.git`. Восстановлено через `git fsck --lost-found`. Урок: перед пушем проверять `df -h`.

**Ключевой вывод для собеса:** 
«Фронт скрывает кнопку — это защита от дурака, а не защита данных. Настоящая защита — это проверка прав в хендлере после загрузки объекта (для IDOR) и понимание, что hard delete в микросервисах без FK ломает соседние базы (для Authors). Плюс классический урок про NATS в кластере: клиентский ретрай с пересозданием стрима — обязателен, если лидер стрима залип на одном subject».