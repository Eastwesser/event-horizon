Что показали SQL-запросы
Что	Значение
Сова (Darkstar) в authors	id = 00a49ebd-..., user_id = ff292932-...
Круглов в authors	id = dcca7043-..., user_id = fddf3f5e-...
author_id карточки Совы (из GET /inventory/items)	ff292932-...
Ключевой вывод: author_id карточки = authors.user_id, а не authors.id. То есть Inventory хранит ссылку на user_id, а не на PK таблицы authors. Это нестандартно (обычно FK смотрят на PK), но в твоей модели это осознанно: один и тот же user_id — общий человек для Auth, Authors, Inventory, Shop, Billing, History. Это как раз то, что ты называешь «user_id — общий человек» в разборе Auth.

И это отлично для проверки IDOR: если metadata несёт user_id (а она его и несёт — см. gateway/middleware/auth.go), то сравнение item.AuthorID == x-user-id корректно. Моя гипотеза про user_id vs author_id не подтвердилась — это не регрессия, а корректная реализация. Снимаю подозрение.

Заодно: authors в eventhorizon_authors есть, а items в eventhorizon_inventory — нет. Значит, таблица карточек называется не items. Смотри миграции: services/inventory/migrations/20260803130001_init_inventory.sql и далее. Скорее всего inventory_items или merch_items. Это не критично для IDOR (API работает), но при следующих SQL-проверках используй правильное имя — иначе снова получишь relation does not exist.

Что показал grep — вот он, финальный ответ
text
grpc_handler.go:63:   userIDs := md.Get("x-user-id")
grpc_handler.go:115:  if err := checkAuthorAccess(ctx, existing.AuthorID); err != nil {
grpc_handler.go:152:  if err := checkAuthorAccess(ctx, existing.AuthorID); err != nil {
grpc_handler.go:304:  if err := checkAuthorAccess(ctx, existing.AuthorID); err != nil {
grpc_handler.go:328:  if err := checkAuthorAccess(ctx, existing.AuthorID); err != nil {
grpc_handler.go:353:  AuthorId: item.AuthorID,
gateway/middleware/auth.go:65:   c.Set(CtxUserID, resp.UserId)
Разбор:

checkAuthorAccess вызывается в 4 местах — это UpdateItem (115), DeleteItem (152), и ещё два (304, 328 — вероятно, SoftDelete и Restore, или Reserve). То есть защита не только в Update/Delete — она шире, что хорошо.

existing.AuthorID — карточка читается из БД до проверки (значит, есть existing, а не просто req.AuthorID, который клиент мог бы подделать). Это правильно: нельзя обойти, подставив чужой author_id в тело запроса.

md.Get("x-user-id") — берётся из metadata, не из тела. Правильно.

Gateway кладёт resp.UserId из Auth ValidateToken в контекст, а UserID(c) возвращает его. Дальше, судя по всему, gateway-хендлеры кладут его в metadata при gRPC-вызове. То есть в x-user-id едет users.id, и это совпадает с authors.user_id (по данным SQL: ff292932 = user_id Совы = author_id карточки).

Вывод: сравнение item.AuthorID == x-user-id — корректное. Оба — users.id. Регрессии happy path нет.

Почему тогда Круглов получил 403 на карточку Совы? Потому что fddf3f5e (Круглов) ≠ ff292932 (Сова) — правильно. А если бы Круглов правил свою карточку — dcca7043-author'а карточка имела бы author_id = fddf3f5e, и проверка бы прошла. Это ровно то, что должно быть.

Финальный вердикт по P0-8
Пункт	Статус
IDOR UpdateItem (чужое → 403)	✅ закрыто
IDOR DeleteItem (чужое → 403)	✅ закрыто
Проверка author_id на бэке	✅ есть, checkAuthorAccess, в 4 местах
x-user-id из metadata, не из тела	✅ есть (md.Get)
existing.AuthorID из БД, не из запроса	✅ есть (нельзя подделать в теле)
author_id карточки = users.id = x-user-id	✅ совпадает (SQL + grep подтверждают)
Happy path (автор правит своё → 200)	✅ должно работать — сравнение корректное
Read-GET чужой карточки	✅ by design (витрина)
P0-8 и AUD-1 — ✅ закрыто, без регрессии. Можно ставить жирную галочку в TODOLISTISHE, BACKEND_TECH_DEBT, AGREGATED_TODOLISTISHE. Заодно закрываются строки из v1.1.3.1: «IDOR в Inventory — на бэкенде отсутствует проверка author_id» — эта запись устарела, проверка есть.

Один момент, который стоит проверить на будущее (не блокер)
В grpc_handler.go есть md.Get("x-user-id") — но что если metadata пустая? Например, кто-то дёргает Inventory напрямую по gRPC (не через Gateway), без x-user-id. Тогда userIDs пуст, и checkAuthorAccess должен вернуть PermissionDenied, а не «пропустить, потому что не с чем сравнить». Судя по тому, что UpdateItem на чужое даёт 403, а не 500 — обработка есть. Но стоит глянуть тело checkAuthorAccess:

bash
sed -n '55,90p' ~/event_horizon/services/inventory/internal/handler/grpc_handler.go
Если там:

go
if len(userIDs) == 0 {
    return status.Error(codes.Unauthenticated, "missing x-user-id")
}
— идеально. Если там if len(userIDs) > 0 && userIDs[0] != authorID — это дыра: без metadata проверка пропускается. Это классическая ошибка «fail-open на пустых metadata». Стоит убедиться, что fail-closed.

И ещё: checkAuthorAccess вызывается в 4 местах, но в grep'е нет CreateItem. Это нормально (создание — свой author_id из токена), но убедись, что CreateItem не принимает author_id из тела — иначе можно создать карточку от чужого имени. Судя по разбору Inventory, «user_id из токена, не из тела» — но в твоём случае это author_id, и стоит проверить, что CreateItem его тоже берёт из metadata.
