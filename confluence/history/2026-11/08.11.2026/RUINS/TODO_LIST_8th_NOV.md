1. Что осталось (сводка по TODO)
Я вижу, что P0 полностью закрыт. Осталось:

Категория	Задач	Что это
MOD-1..4 (P1)	4	Уведомления автору при удалении/редактировании товара админом
MOD-2 (P1)	1	Админ вводит причину удаления/редактирования (модалка + soft-delete)
HIST-1..5 (P1)	5	История: почему пусто на фронте, 10 последних записей, статистика под вкладку
CASE-1..3 (P1)	3	Кейсы «30 пицц», «лаги», «два игрока купили последнюю карту»
P1 (бэкенд)	~20	Разные техдолги (Retry Jitter, Bulkhead, Rate Limit на record_beaten, reuse detection и т.д.)
P2 (фронтенд)	~10	FE-CARD-1..4 (частично закрыты), FE-MOB-1..4 (частично), Auth UX (глазик, автозаполнение, смена пароля), ник в профиле
P2 (админка)	3	Статистика инвентаря (частично), «По автору» (дедупликация 108 авторов)
P3 (roadmap)	8	i18n, пагинация каталога, уровни лидерборда, порог покрытия в CI
P4 (парковка)	6	Loki, WAF, C4 payouts, 3D и т.д.
Итого незакрытых: ~60 задач. Но критичных (P0) — 0.










2. Задача на завтра (твоя идея)
Ты сказал:

«Давай завтра еще проверим, удаляются ли у Раи карточки из инвентаря, если автор удаляет товар. Или админ. Админские права надо сделать вседозволенными.»

Это MOD-1 + MOD-2 + проверка удаления. Давай разобьём на шаги:

Шаг 1. Проверить, что происходит сейчас.

Автор удаляет карточку (DELETE /inventory/items/{id}) → что происходит с покупкой у Раисы? (Остаётся ли shop.purchases? Показывается ли товар в её инвентаре?)

Админ удаляет карточку → то же самое.

Шаг 2. Сделать админские права вседозволенными.

Сейчас CancelPurchase работает только для своих покупок. Для админа нужно API POST /admin/shop/purchases/{id}/cancel (или что-то похожее).

Или разрешить админу отменять чужие покупки через существующий CancelPurchase (но с проверкой role = admin).

Шаг 3. MOD-1 (уведомления).

При удалении товара админом (или автором) — отправить уведомление автору: «Ваш товар снят с продажи. Причина: ...».

При редактировании цены/описания — уведомление автору.

В уведомлении — кнопка «Подробнее» → модалка с причиной.




Шаг 4. Soft-delete.

Убедиться, что при удалении товара он не удаляется физически, а ставится deleted_at.

Тогда покупки у Раисы остаются, но товар не показывается в каталоге.



3. Что я подготовлю к завтра
SQL-проверку: что происходит с shop.purchases и shop.inventory при удалении карточки из inventory.

Код для MOD-1: событие inventory.item.rejected в NATS, слушатель в notification, запись в БД.

Код для админских прав: POST /api/v1/admin/shop/purchases/{id}/cancel с проверкой role = admin.

Фронтенд: модалка «Причина удаления» в админке, уведомление с кнопкой «Подробнее».





UPDATED:


2. Про shop.inventory — хранить снимок
Ты согласен: «фото должно оставаться, как снимок памяти». Правильно.

Что нужно сделать (MVP):

Добавить в shop.inventory колонки item_name и item_image_url (или JSONB с snapshot).

При покупке (PurchaseItemWithStock) — сохранять name и image_url в shop.inventory.

При чтении инвентаря (GetUserInventory) — если item_id удалён, использовать снимок.

Миграция:

sql
ALTER TABLE inventory ADD COLUMN item_name TEXT;
ALTER TABLE inventory ADD COLUMN item_image_url TEXT;
Код:

go
// В PurchaseItemWithStock
_, err = tx.ExecContext(ctx, `
    INSERT INTO inventory (user_id, item_id, item_name, item_image_url)
    VALUES ($1, $2, $3, $4)`,
    userID, itemID, item.Name, item.ImageURL)
В GetUserInventory:

sql
SELECT i.id, i.name, i.description, i.price, i.category, i.game_id, i.image_url, i.available,
       inv.purchased_at,
       COALESCE(p.price, 0) AS purchase_price,
       COALESCE(p.id::text, '') AS purchase_id,
       p.refundable_until,
       p.fulfilled_at,
       inv.item_name,       -- снимок
       inv.item_image_url   -- снимок
FROM inventory inv
LEFT JOIN items i ON inv.item_id = i.id
LEFT JOIN LATERAL (...) p ON true
WHERE inv.user_id = $1
Но это изменение схемы + код. Пока запиши в TODO_LIST_8th_NOV.md:






markdown
## Shop: снимок товара в inventory
- [ ] При покупке сохранять `name` и `image_url` в `shop.inventory` (снимок на момент покупки).
- [ ] При чтении инвентаря — использовать снимок, если товар удалён.








Что произошло
PUT с version: 4 → 200 OK. Значит, checkAuthorAccess пропустил админа. Проверка roles[0] == "admin" работает.

Версия стала 5, имя изменилось на «Фея чащи (edited by admin)».

Откат с version: 5 → 200 OK, имя вернулось на «Фея чащи», версия стала 6.

Важно: в ответе видно "author_id": "e9b1caab-..." — это твой admin_id, а не автор Феи. Но Inventory сохранил author_id исходного автора (Феи), а в ответе author_id — это author_id из запроса (который ты не передавал, поэтому он пустой или admin). Это нормально — UpdateItem не меняет author_id.

Но есть нюанс: в ответе "attributes": {}, "images": [], "price": 0, "stock": 0. Это потому, что ты передал только name в PUT. Остальные поля обнулились. Это проблема — UpdateItem должен мержить поля, а не перезаписывать.

Но это отдельный баг (не Шаг 2.1). Запиши в TODO: «UpdateItem обнуляет поля, если их не передали. Нужно мержить.»




## UpdateItem: мержить поля
- [ ] `UpdateItem` обнуляет `attributes`, `images`, `price`, `stock`, если их не передали. Нужно мержить.


Шаг 7. Фронт (опционально, потом)
На фронте это будет кнопка «Отменить покупку» в админке (в списке покупок или в карточке пользователя). Но это потом. Сначала — бэкенд.







# ITEM_SHOPCARD_FIX.md (дополнение)
- [ ] При удалении товара из inventory — картинка в shop.inventory должна остаться (хранить image_url в shop.inventory или soft-delete).





## AdminCancelPurchase: обход fulfilled_at
- [ ] Админ должен мочь отменить покупку, даже если `fulfilled_at` стоит.
- [ ] Для этого в `RefundPurchase` (или в `AdminCancelPurchase`) убрать проверку `fulfilled_at` для админа.







+ ADD SENTINEL FOR REDIS

Про «критичные Redis»:
Критичные — те, без которых сервис не работает:

Auth (:6379) — сессии, rate limit. Если упадёт — все получат 503.

Billing (:6381) — кэш балансов. Если упадёт — баланс читается из PG (медленнее, но работает).

Shop (:6383) — кэш каталога. Если упадёт — каталог читается из PG.

Inventory (:6384) — кэш карточек.

Leaderboard (:6382) — Sorted Set для топа. Если упадёт — топ читается из PG.

Profile (:6385), Authors (:6387), Payment (:6386) — кэши.

Для MVP достаточно Sentinel для Auth (самый критичный) и Billing (деньги). Остальные можно оставить как есть (они не SPOF, потому что есть fallback на PG).