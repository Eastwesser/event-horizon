# Урок 06 — брокеры (unknown date)

Sync vs async, RabbitMQ vs Kafka vs Redis PubSub vs NATS, DLQ. Теория в `06_full_brokers.md` / `06_brokers_theme.md`.

## Задачи / что говорить

### 1. Sync vs async broker
Sync: логин, платёж, поиск. Async: аналитика, email, сглаживание пиков, decoupling. Фраза: «клиенту — accepted, обработка позже; брокер переживает падение консьюмера».

### 2. DLQ + retry
Сообщения с исчерпанными попытками → Dead Letter Queue; компенсирующие транзакции / ручной разбор. Kafka: отдельный топик; Rabbit: DLX.

### Сравнение (шпаргалка)
| | Kafka | RabbitMQ |
|--|--|--|
| Модель | log, dumb broker | smart broker, push |
| Throughput | очень высокий | ниже |
| Replay | да (offset) | обычно нет |

## Как запускать

```bash
export GOWORK=off
cd code/01_sync_vs_async_broker && go run main.go
```
