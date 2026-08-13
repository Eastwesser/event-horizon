# Apache Kafka для senior backend (red_mad_robot)

Kafka — распределённый commit log. Сообщения пишутся в **топики**, дробятся на **партиции**, читаются **consumer groups** с отслеживанием **offset**. Сильная сторона: throughput, retention, replay. В Event Horizon Kafka используется для purchase-flow Week 5: топик `purchase.paid` (Shop → Fulfillment/Notification).

## Топики и партиции

- **Topic** — логический поток событий (например `purchase.paid`).
- **Partition** — упорядоченный append-only лог; единица параллелизма и масштабирования.
- Producer выбирает партицию: round-robin, или по **key** (hash). Один ключ → одна партиция → порядок для этого ключа.
- Больше партиций → выше parallelism consumers, но больше накладных расходов (файлы, rebalance).
- Replication factor / ISR: отказоустойчивость записи; `acks=all` + `min.insync.replicas` на проде.

## Producer: семантика записи

- **acks=0/1/all**: компромисс latency vs durability.
- Идемпотентный producer (`enable.idempotence`) снижает дубли при ретраях на уровне брокера.
- Batching + compression (lz4/zstd) — ключ к throughput.
- Ключ в EH purchase-flow: логично `purchase_uuid` или `user_uuid` — чтобы события одной покупки/пользователя шли в одну партицию.

## Consumer groups и offset

- **Consumer group**: несколько инстансов делят партиции топика; одна партиция → не более одного consumer в группе.
- **Offset** — позиция в партиции; commit после обработки (at-least-once) или до (риск потери).
- Rebalance: при добавлении/удалении consumer'а партиции переназначаются — «stop-the-world» эффекты, sticky/cooperative assignors смягчают.
- **Consumer lag** = high watermark − committed offset; главный SLO-сигнал здоровья пайплайна.
- Несколько групп на один топик = независимый fan-out (Fulfillment и Analytics могут читать раздельно).

## Семантика доставки и идемпотентность

- Стандарт: **at-least-once** (обработал → commit offset). Падение между side-effect и commit → дубль.
- Exactly-once в Kafka — узкий сценарий (transactions / EOS); в микросервисах обычно: **дедуп по `event_uuid`** + идемпотентный handler.
- Poison message: retry limit → DLQ-топик / side storage, иначе бесконечный reprocess одной партиции блокирует порядок.

## Replay и retention

- Retention по времени/размеру (или compaction по ключу для changelog).
- Consumer может сбросить offset на earliest/timestamp и **переиграть** историю — сильный аргумент vs RabbitMQ.
- Log compaction: последнее значение по ключу (удобно для состояния, не для полного аудита).

## Event Horizon: purchase.paid

Контракт: `contracts/events/PURCHASE_KAFKA.md`, типы в `contracts/events/purchase.go`.

| Topic | Producer | Consumer |
|-------|----------|----------|
| `purchase.paid` | shop | fulfillment, notification |
| `purchase.fulfilled` | fulfillment | shop (optional), notification |

- Payload `PurchasePaid`: `event_uuid`, `purchase_uuid`, `user_uuid`, `item_uuid`, `price`.
- Env: `KAFKA_BROKERS=kafka:9092`; клиент Sarama (`-tags kafka`).
- Важно: Kafka **не заменяет** NATS в EH. NATS JetStream — основная доменная шина (`shop.purchased`, `inventory.item.created`, …). Kafka — отдельный контур fulfillment после оплаты.
- Shop пишет в outbox/Kafka после успешной покупки; Fulfillment обрабатывает идемпотентно по `event_uuid` / `purchase_uuid`.

## Ops: что мониторят на проде

- Consumer lag по группам/партициям.
- Under-replicated partitions, offline partitions.
- Disk usage / retention; request latency produce/fetch.
- Rebalance rate, failed fetch/produce.
- Схема эволюции: Avro/Protobuf + Schema Registry (на собесе упомянуть даже если в EH JSON).

## Когда Kafka — правильный выбор

- Высокий throughput event streams, analytics, CDC.
- Нужен replay и долгий retention.
- Много независимых consumer groups на один поток.
- Избыточна для простой RPC-замены и лёгкой шины 5–15 сервисов — там часто выигрывает NATS (как в EH).

## Типичные вопросы на собесе

- Что такое партиция и как связан порядок сообщений с ключом?
- Как работает consumer group и что происходит при rebalance?
- Чем offset отличается от ack в RabbitMQ?
- Как бороться с consumer lag?
- Почему at-least-once и как сделать обработку идемпотентной?
- Чем log compaction отличается от time-based retention?
- Как в EH разделены роли Kafka (`purchase.paid`) и NATS?
