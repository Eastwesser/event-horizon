# До и после — v1.1.1

Список из эссе 10.10.2026. Все восемь пунктов закрыты в коде. Миграция `services/shop/migrations/20261010234100_purchase_claims.sql` применяется обычным `make deploy` / `make migrate-all`.

| # | Было | Стало |
|---|------|--------|
| 1 | `purchase.paid` нёс случайный id. `fulfilled_at` не находил строку покупки. | В событии `PurchaseUUID` = `purchases.id`, который записан в той же транзакции. |
| 2 | Ключ списания содержал наносекунды. Два клика — два списания. Компенсация жила только в памяти процесса. | Ключ `shop:{user}:{item}:{cycle}`. Повтор того же цикла упирается в `UNIQUE` и не списывает снова. Если запись покупки упала, строка `billing.refund` остаётся в outbox, воркер вызывает `AddCurrency` с тем же ключом. После возврата cycle увеличивается, следующая покупка получает новый ключ. |
| 3 | Повтор отмены видел `AlreadyRefunded` и больше не начислял билетики. | Повтор снова зовёт `AddCurrency` с `cancel-{purchaseID}`. Второй раз уникальный ключ не проводит деньги, баланс читается как есть. |
| 4 | `purchase.paid` публиковался после commit, мимо outbox. Воркер брал строки без блокировки. | `shop.purchased` и `purchase.paid` пишутся в outbox внутри транзакции покупки. Воркер берёт пачку `FOR UPDATE SKIP LOCKED` и помечает строку только после успешной доставки. |
| 5 | `CheckOrigin` всегда `true`, сокет без токена. | Origin должен совпасть с хостом страницы, localhost или `WS_ALLOWED_ORIGINS`. До upgrade нужен `?token=` или `Authorization: Bearer`. Фронт передаёт access-токен. |
| 6 | При заданном `KAFKA_BROKERS` заказ читали и NATS, и Kafka. | Если JetStream подключён, Kafka-consumer у fulfillment и notification не стартует. `purchase.fulfilled` уходит в один транспорт. |
| 7 | Ошибка Redis на `ValidateToken` логировалась, подпись считалась достаточной. | Auth возвращает `Unavailable`. Gateway отвечает 503 `session store unavailable`. То же на refresh. |
| 8 | Дашборд не показывал путь покупки и срыв возврата. | Панели «Purchase paid → fulfilled» (`orders_total`, `purchases_fulfilled_total`, p95 `assembly_duration_seconds`) и «Shop refund still pending» (`shop_purchase_refund_failures_total`). Loki и Tempo не добавлялись. |

Проверено: `go test` shop/service, auth/service, fulfillment/service, gateway/middleware. Сборка пакетов gateway, shop, notification, fulfillment.
