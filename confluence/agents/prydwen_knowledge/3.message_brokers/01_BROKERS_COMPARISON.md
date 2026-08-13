# Сравнение брокеров сообщений: Kafka vs NATS vs RabbitMQ

Контекст для senior-собеса (red_mad_robot / backend Go): когда нужен брокер, чем отличаются модели доставки, порядок, replay и операционная нагрузка. В Event Horizon основная шина — **NATS JetStream**; Kafka используется точечно для purchase-flow (`purchase.paid`).

## Когда нужен брокер

- Синхронный RPC (gRPC/HTTP) — запрос-ответ, жёсткая связность, нужен немедленный результат.
- Асинхронный брокер — развязка producer/consumer, пики нагрузки, fan-out, eventual consistency.
- Выбирай брокер, если: несколько потребителей одного события, нужна буферизация при даунтайме consumer'а, аудит/replay, саги между сервисами.
- Не тащи брокер «на всякий случай»: лишняя infra, сложнее отладка, at-least-once → обязательно идемпотентность.

## Throughput (пропускная способность)

- **Kafka**: ориентирован на высокий throughput (сотни тысяч–миллионы msg/s на кластер). Лог на диск, batch + zero-copy; масштаб через партиции.
- **NATS / JetStream**: очень низкая latency, высокий throughput на лёгких сообщениях; JetStream чуть медленнее core NATS из‑за persistence, но для доменных событий EH более чем достаточно.
- **RabbitMQ**: throughput ниже Kafka при сопоставимом железе; сильнее в сложной маршрутизации, чем в «firehose»-стримах.
- Практический вывод: если нужен event log / analytics pipeline — Kafka; если лёгкая шина микросервисов — NATS; если сложные routing rules / RPC over MQ — RabbitMQ.

## Ordering (порядок сообщений)

- **Kafka**: порядок **гарантируется внутри партиции** (ключ → одна партиция). Между партициями порядка нет. Ключ = `user_id` / `purchase_uuid`.
- **NATS JetStream**: порядок в рамках stream/consumer при sequential processing; при конкурентных consumer'ах порядок ослабевает — как и везде.
- **RabbitMQ**: порядок в очереди при одном consumer; при competing consumers и prefetch > 1 порядок не гарантирован.
- На собесе: «глобальный порядок во всей системе» — почти всегда ложь; уточняй границы (partition / queue / subject).

## Replay (повторное чтение)

- **Kafka**: сильная сторона — offset log, retention (время/размер), consumer может «перемотать» offset и переиграть историю. Идеально для reprocessing, CDC, analytics.
- **NATS JetStream**: есть replay по sequence/time для durable consumers; retention настраивается на stream. Удобно, но экосистема replay/tooling слабее Kafka.
- **RabbitMQ**: классически **нет** долгого replay после ack (сообщение удалено). Нужны отдельные audit-очереди / DLQ / внешнее хранилище событий.
- Redis Pub/Sub: fire-and-forget, без persistence и без replay — не замена брокеру доменных событий.

## Модель доставки и семантика

- **At-most-once**: потеря допустима (метрики «примерно»).
- **At-least-once**: стандарт для Kafka/NATS/RMQ при ack после обработки → дубли → **идемпотентный consumer**.
- **Exactly-once**: маркетинг; на практике — идемпотентность + дедуп по `event_uuid` / transactional outbox.
- Kafka: consumer group + commit offset; NATS JetStream: ManualAck + durable; RabbitMQ: ack/nack + prefetch.

## Ops и сложность эксплуатации

- **Kafka**: тяжёлая ops-нагрузка (брокеры, контроллеры/KRaft или ZooKeeper legacy, disk, rebalance, ISR). Нужен зрелый мониторинг lag, under-replicated partitions.
- **RabbitMQ**: проще старт, Erlang VM, mirrored/quorum queues, management UI; кластер и сеть — частые боли на проде.
- **NATS**: минимальный footprint, простой кластер; JetStream добавляет disk/replication. В EH: 3 ноды в `docker-compose.cluster.yml`.
- Стоимость команды: Kafka часто требует выделенного владельца платформы; NATS — «встроили и поехали» для mid-size систем.

## Матрица выбора (шпаргалка)

- Event sourcing / audit log / долгий retention → **Kafka**.
- Микросервисная шина, low latency, subjects, Go-friendly → **NATS JetStream** (выбор EH).
- Сложная маршрутизация (topic/fanout/headers), приоритеты, RPC-паттерн → **RabbitMQ**.
- Кэш-инвалидация / эфемерные уведомления → Redis Pub/Sub или NATS core (без JetStream).
- В EH: NATS для `shop.purchased`, `inventory.item.created`, …; Kafka для `purchase.paid` → Fulfillment (Week 5).

## Типичные ошибки на собесе

- Путать queue (point-to-point) и pub/sub (fan-out).
- Обещать exactly-once без outbox + идемпотентности.
- Игнорировать consumer lag и backpressure.
- Класть большие payload в брокер вместо ссылки на object storage / id сущности.

## Типичные вопросы на собесе

- Чем Kafka отличается от RabbitMQ по модели хранения и replay?
- Как обеспечить порядок событий по одному пользователю?
- Почему at-least-once почти всегда и что делать с дублями?
- Когда NATS предпочтительнее Kafka в микросервисах?
- Как измерить и что делать с consumer lag?
- Почему Redis Pub/Sub не подходит для критичных доменных событий?
- Как в вашей системе разделены NATS и Kafka (пример EH)?
