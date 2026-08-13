# NATS и JetStream в Event Horizon

NATS — лёгкий messaging system с subject-based адресацией. **Core NATS** — эфемерный pub/sub; **JetStream** — persistence, durable consumers, replay, ack. В Event Horizon JetStream — **основная шина доменных событий**. Кластер: 3 ноды (`nats-1..3`) в `docker-compose.cluster.yml`.

## Core NATS vs JetStream

- Core: publish/subscribe, fire-and-forget, нет гарантированного хранения после доставки online-подписчикам.
- JetStream: сообщения пишутся в **stream**, consumers читают с ack, переживают рестарт.
- Для критичных доменных событий (покупка, баланс, регистрация) — только JetStream (+ outbox на стороне producer).
- Core уместен для эфемерных сигналов (cache bust hints), где потеря допустима.

## Subjects: иерархия и wildcards

- Subject — строка с токенами через точку: `shop.purchased`, `inventory.item.created`.
- Wildcards: `*` — один токен; `>` — хвост (`event.>`).
- Конвенция EH: `<domain>.<action>` или `<domain>.<entity>.<action>`.
- Не смешивай команды и события в одном subject без дисциплины; событие = факт в прошлом (`purchased`, не `purchase`).

## Streams и consumers

- **Stream**: хранит сообщения по набору subjects, retention (limits/interest/workqueue), replicas.
- **Consumer**: push/pull, durable name, ack policy, deliver policy (all/last/new/by-start-sequence/time).
- **Durable**: именованный consumer; позиция сохраняется → рестарт сервиса продолжает с незакрытых сообщений.
- **ManualAck**: ack только после успешной записи в PG/ClickHouse; иначе Nak → retry.
- Competing consumers одного durable — масштабирование обработки (с оглядкой на порядок).

## Подключение в сервисах EH

- Env: `NATS_URL` (список `nats://host:4222,...`).
- После connect — `JetStream()`; outbox workers: `js.Publish("shop.purchased", data)`.
- `nats-hub` поднимает/синхронизирует streams при старте кластера — единая точка правды по subjects.
- Не храни `context.Context` в struct worker'а; передавай в `Start(ctx)`.

## Таблица subject'ов Event Horizon

| Subject | Producer | Consumer(s) |
|---------|----------|-------------|
| `inventory.item.created` | Inventory outbox | Shop (создаёт merch item) |
| `shop.purchased` | Shop | History, Analytics, … |
| `payment.completed` | Payment outbox | History, Analytics |
| `author.upserted` | Authors outbox | History, Analytics |
| `score.updated` | Game/Leaderboard path | Profile, History, Analytics |
| `user.registered` | Auth/Profile path | History, Analytics |
| `balance.updated` | Billing outbox | подписчики по необходимости |
| `shop.purchase.failed` | Shop compensation | ops/alerts |

Kafka отдельно: `purchase.paid` (Shop → Fulfillment), не заменяет NATS.

## Durable naming (практика EH)

- History: `history-<subject>`, ManualAck.
- Analytics: `analytics-<subject>`.
- Shop inventory sync: durable `shop-inventory-sync`.
- Имена стабильные: смена durable = новый consumer с нуля (или явная политика deliver).

## Delivery, retry, идемпотентность

- Ack после успешного side-effect (insert в History PG / Analytics CH).
- Nak / in-progress на transient ошибки (сеть, lock timeout).
- Дубли возможны → consumer идемпотентен по `event_uuid` / бизнес-ключу.
- Outbox на producer: событие в той же PG-транзакции, что и бизнес-изменение (Shop, Billing, Inventory, Payment, Authors).

## Ops и отладка

- `nats` CLI: stream ls/info, consumer info, lag.
- Exporter/метрики в compose; логи: `history ingest`, `analytics ingest`, outbox tick.
- Следи за: unacked, redelivery count, disk JetStream, недоступность нод кластера.
- Backpressure: медленный consumer → рост pending; масштабируй replicas сервиса или оптимизируй handler.

## Почему NATS выбран для EH

- Низкая latency, простой API для Go (`nats.go`).
- Достаточный persistence через JetStream без тяжести Kafka.
- Subject model хорошо ложится на доменные события микросервисов.
- Kafka оставлена для учебного/прод-контура fulfillment (Week 5), где важен классический log + consumer groups.

## Типичные вопросы на собесе

- Чем Core NATS отличается от JetStream?
- Как работают wildcards `*` и `>`?
- Что такое durable consumer и ManualAck?
- Как гарантировать, что событие не потеряется при падении сервиса сразу после commit БД? (outbox)
- Как именуете durable'ы и что будет при смене имени?
- Когда добавите Kafka рядом с NATS (пример EH `purchase.paid`)?
- Как отладить «сообщение опубликовали, consumer не видит»?
