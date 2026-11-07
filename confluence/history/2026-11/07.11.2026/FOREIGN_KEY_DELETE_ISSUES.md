Что показал SQL по FK
text
referencing_table | referencing_column | referenced_table | referenced_column | delete_rule
------------------+--------------------+------------------+-------------------+-------------
(0 rows)
В eventhorizon_authors на таблицу authors не ссылается ни одна FK. Ни authors.id, ни authors.user_id. То есть в этой базе DELETE не упадёт по FK. Но:

authors существует в других базах. В eventhorizon_inventory карточка хранит author_id = authors.user_id (это мы уже установили). В eventhorizon_payment — revenue_splits. В eventhorizon_shop — purchases/purchase_sagas. Это межбазовые ссылки, FK между базами в Postgres нет. Значит, DELETE в eventhorizon_authors не упадёт по FK вообще — он просто оставит висячие ссылки в других базах. Это хуже, чем падение: карточки остаются с author_id, которого больше нет в authors, и GET /api/v1/authors/:user_id вернёт 404, а витрина магазина покажет товар с несуществующим автором.

authors.user_id — text, не uuid. Это отдельная деталь: в authors user_id text not null, а в users (Auth) — uuid. Значит, author_id в Inventory — тоже text (или uuid? мы видели ff292932-..., но тип не смотрели). Если где-то uuid, а где-то text — сравнение через = в Go (string == string) работает, а в SQL между базами — не работает. На IDOR это не влияет (там сравнение в Go), но при любых JOIN между базами — сломается.

В authors нет deleted_at — только active boolean not null default true + индекс idx_authors_active. То есть soft-delete уже спроектирован: колонка есть, индекс есть. Но RevertApplication им не пользуется — делает DELETE.

Что это значит для P0-6
Формулировка в TODO — «RevertApplication hard delete — FK падает» — неточна для текущей схемы. По факту:

FK на authors внутри eventhorizon_authors — нет (SQL подтверждает 0 rows).

Значит, DELETE не упадёт. Он пройдёт молча.

Но он оставит висячие ссылки в других базах (Inventory, Payment, Shop), и это хуже, чем падение: данные консистентно не ломаются, но витрина и история начинают ссылаться в никуда.

Это классический случай, когда «FK падает» в TODO — это страх, а не факт. Реальность: FK нет, поэтому DELETE тихо проходит, и проблема не в падении, а в тихой порче межбазовых ссылок.