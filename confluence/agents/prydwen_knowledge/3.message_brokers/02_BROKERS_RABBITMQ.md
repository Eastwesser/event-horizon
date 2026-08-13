# RabbitMQ для senior backend (red_mad_robot)

RabbitMQ — классический message broker на AMQP 0-9-1. Сильная сторона — гибкая маршрутизация через exchanges и очереди, удобный management UI, паттерны RPC и delayed/priority messaging. В Event Horizon основной брокер — NATS; RabbitMQ нужно знать как альтернативу и для собеса.

## Базовая модель: Exchange → Binding → Queue → Consumer

- **Producer** публикует не «в очередь», а в **exchange** с routing key.
- **Exchange** решает, в какие **queues** положить сообщение по bindings.
- **Consumer** читает из queue; при competing consumers — load balancing round-robin (с учётом prefetch).
- Сообщение живёт в queue до ack (или до TTL/DLX, если настроено).

## Типы exchanges

- **direct**: точное совпадение routing key → queue. Классика для команд и точечной доставки.
- **topic**: pattern matching (`order.*.created`, `payment.#`). Удобно для иерархии событий.
- **fanout**: копирует во все связанные queues, routing key игнорируется. Broadcast/notifications.
- **headers**: маршрутизация по заголовкам (реже на практике, тяжелее рассуждать).
- Default exchange (`""`): routing key = имя queue — упрощённый путь «прямо в очередь».

## Queues: свойства и паттерны

- **Durable queue** + **persistent message** — переживают рестарт брокера (не путать: оба флага нужны).
- **Exclusive / auto-delete** — для временных RPC reply-очередей.
- **Prefetch (QoS)**: сколько unacked сообщений может держать consumer. Слишком большой → неравномерная нагрузка; слишком маленький → idle.
- **Priority queues**: поле priority; используй осторожно (ограниченный диапазон, влияние на fairness).
- **Quorum queues** (современный prod-default) vs classic mirrored: лучше consistency в кластере.

## Acknowledgement и семантика доставки

- **auto-ack**: брокер считает доставленным сразу → риск потери при падении consumer'а (at-most-once).
- **manual ack**: ack после успешной обработки → at-least-once; при падении — requeue.
- **nack / reject**: с requeue=true — повтор; с requeue=false — в DLX/DLQ или дроп.
- **Idempotency**: consumer обязан переживать повторную доставку (unique key операции, upsert).
- На собесе: «где ack?» — после commit БД / успешного side-effect, не «после парсинга JSON».

## DLX / DLQ и отложенные сообщения

- **Dead Letter Exchange (DLX)**: куда уходит сообщение после reject, TTL expire, max-length.
- **DLQ** — очередь для разбора poison messages; обязательный ops-паттерн на проде.
- Delayed messages: плагины / TTL + DLX hop (сообщение «созревает» и возвращается в рабочую очередь).
- Retry с backoff: отдельные retry-очереди с растущим TTL, лимит попыток → DLQ.

## RPC over RabbitMQ

- Client публикует request с `reply_to` + `correlation_id`.
- Server отвечает в reply queue; client матчит по correlation_id.
- Минусы vs gRPC: сложнее таймауты, tracing, контракты; плюс — буферизация при недоступности worker'а.
- В EH RPC между сервисами — **gRPC**, не RabbitMQ.

## Когда выбирать RabbitMQ

- Нужна **сложная маршрутизация** (topic/fanout/headers) без написания своего роутера.
- Work queues с приоритетами, delayed jobs, classic enterprise integration.
- Команда уже умеет AMQP; есть готовые библиотеки и UI для ops.
- Не лучший выбор, если нужен долгий event log / replay / высокий throughput stream (смотри Kafka).
- Не лучший выбор для ультра-лёгкой шины Go-микросервисов с subjects (смотри NATS JetStream — путь EH).

## Ops и типичные боли

- Кластер и сетевые партиции: split-brain, «нода думает что она мастер».
- Memory/disk alarms: publisher confirm + блокировка публикации при нехватке ресурсов.
- Unacked messages раздувают память — следи за prefetch и скоростью consumer'ов.
- Мониторинг: queue depth, consumer count, ack rate, DLQ growth.

## Связь с Event Horizon

- EH не использует RabbitMQ как основную шину: доменные события идут через **NATS JetStream**, purchase-flow Week 5 — через **Kafka** (`purchase.paid`).
- Имеет смысл сравнить на собесе: RMQ bindings ≈ NATS subjects/wildcards; RMQ ack ≈ JetStream ManualAck; RMQ не даёт Kafka-style replay.

## Типичные вопросы на собесе

- Чем direct exchange отличается от topic и fanout?
- Зачем prefetch и как его подобрать?
- Как устроены DLX/DLQ и retry с backoff?
- Почему durable queue без persistent message всё ещё может потерять данные?
- Как реализовать RPC на RabbitMQ и какие у этого минусы?
- At-least-once vs at-most-once в терминах ack?
- Когда выберете RabbitMQ, а когда Kafka или NATS?
