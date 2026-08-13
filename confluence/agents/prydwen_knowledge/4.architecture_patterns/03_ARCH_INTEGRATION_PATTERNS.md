# Integration patterns: Outbox, Saga, Idempotency (Event Horizon)

Распределённые системы не имеют атомарных кросс-сервисных транзакций «из коробки». Senior обязан объяснить **Transactional Outbox**, **Saga/компенсацию** и **идемпотентность**. В EH это видно на Shop/Billing/Inventory/Payment.

## Проблема dual write

- Нельзя надёжно сделать: `UPDATE db` + `Publish(broker)` как одну атомарную операцию без паттерна.
- Если сначала publish — событие может уйти без commit.
- Если сначала commit, потом publish — процесс может упасть → событие потеряно.
- Решение: **Outbox** (или inbox/CDC из WAL) — событие в БД в той же транзакции.

## Transactional Outbox (как в EH)

1. В одной PG-транзакции меняем бизнес-таблицу **и** вставляем строку в `outbox` (`payload` JSON, `subject`/`event_type`, `created_at`).
2. Коммитим.
3. Фоновый worker читает unpublished, `Publish` в JetStream/Kafka, помечает processed / удаляет.

Где смотреть в коде:

- **Shop** — покупка + `shop.purchased` / Kafka `purchase.paid` (`shop_service`, outbox).
- **Billing** — `AddBalance`/`SpendBalance` → `balance.updated` в том же tx (`postgres_repo`).
- **Inventory** — create item → `inventory.item.created` (Shop подписан, создаёт merch).
- **Payment** — `payment.completed` после confirm.
- **Authors** — `author.upserted`.

Практика:

- Worker идемпотентен к повторной публикации; consumer переживает дубли.
- Батчинг polling outbox + индекс по `processed`/`created_at`.
- Не класть `context.Context` в worker struct; `Start(ctx)`.

## Saga / компенсация (лёгкая модель EH)

- Классическая Saga: длинная бизнес-транзакция = цепочка локальных транзакций + компенсации.
- Оркестрация vs хореография: оркестратор знает шаги; хореография — реакция на события.
- **Shop Purchase в EH** (локальная компенсация, не тяжёлый orchestrator):
  1. Billing `SpendCurrency`.
  2. `PurchaseItemWithStock`.
  3. Если сток падает → refund через `AddCurrency` + событие `shop.purchase.failed`.
- Важно: компенсация тоже должна быть идемпотентной; частичные сбои логируй и алерти.
- 2PC обычно избегают в микросервисах (availability, locking).

## Idempotency (обязательный спутник at-least-once)

- Повторы приходят от: client retry, gateway retry, broker redelivery, outbox republish.
- Ключи идемпотентности: `Idempotency-Key` на HTTP, `event_uuid` на событиях, бизнес-ключ (`purchase_uuid`).
- Паттерны: unique constraint + upsert; таблица processed_events; сравнение версий.
- Consumers History/Analytics/Fulfillment: durable + ack **после** успешной записи; повтор не должен двоить деньги/метрики критичным образом.
- Для Spend/Add: повторный Spend с тем же key не должен списать дважды.

## Inbox pattern (симметрично)

- Consumer в одной tx: сохранить `event_id` как обработанный + применить эффект.
- Если `event_id` уже есть — skip. Сильный способ exactly-once **на стороне приложения**.

## Cache invalidation рядом с интеграцией

- Shop после покупки сбрасывает `shop:items:*`, `balance:*:tickets`.
- Inventory decorator инвалидирует Redis на Update/Delete.
- Auth logout бьёт session/refresh в Redis.
- Invalidate-after-write проще, чем dual write в кэш без outbox; stale допустим с TTL.

## Circuit breaker (связь с интеграцией)

- Gateway оборачивает критичные gRPC (`internal/circuit`): после N ошибок — open, быстрый 503; half-open Timeout≈10s, MaxRequests=3.
- Не ретраить без jitter бесконечно — усилишь каскадный отказ.

## Как рассказывать на собесе end-to-end покупку EH

1. Client → Gateway → Shop.Purchase (gRPC).
2. Shop синхронно: Billing.Spend → сток/покупка; при ошибке стока — компенсация Add + `shop.purchase.failed`.
3. Outbox → NATS `shop.purchased` (History/Analytics) и/или Kafka `purchase.paid` (Fulfillment).
4. Fulfillment идемпотентно обрабатывает выдачу → `purchase.fulfilled`.
5. Метрики/трейсы связывают всё по trace/correlation id.

## Типичные вопросы на собесе

- Почему dual write опасен и как работает transactional outbox?
- Чем оркестрация saga отличается от хореографии?
- Как спроектировать идемпотентный SpendBalance?
- Что делать, если компенсация сама упала?
- Зачем inbox, если уже есть outbox?
- Как в EH устроена покупка Shop↔Billing и зачем `shop.purchase.failed`?
- Где ack/commit offset относительно записи в БД consumer'а?
