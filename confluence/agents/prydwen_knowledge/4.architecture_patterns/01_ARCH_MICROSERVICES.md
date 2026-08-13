# Микросервисная архитектура (senior / red_mad_robot)

Микросервисы — способ нарезать систему по **бизнес-границам** с независимым деплоем. Цена: сеть, распределённые транзакции, наблюдаемость. Event Horizon — учебный/боевой полигон: gateway + доменные gRPC-сервисы + NATS/Kafka + отдельные БД.

## Границы сервисов (bounded context)

- Режь по домену (Auth, Billing, Shop, Inventory), не по слоям («весь persistence в один сервис»).
- У сервиса своя модель и ideally своё хранилище; чужие таблицы не читать напрямую.
- Контракт наружу: gRPC proto / события (NATS subjects, Kafka topics) — не shared DB.
- Признак плохой границы: постоянные синхронные цепочки A→B→C на каждый user request и частые «распределённые join'ы».
- Признак хорошей: сервис можно понять, протестировать и задеплоить относительно независимо.

## Монолит vs микросервисы vs «распределённый монолит»

- **Монолит**: проще транзакции и рефакторинг; хуже независимый scale/deploy команд.
- **Микросервисы**: независимые релизы, scale по горячим точкам; сложнее ops и consistency.
- **Распределённый монолит**: много процессов, но жёсткая связность (общая БД, синхронные вызовы везде) — минусы обоих миров.
- На старте продукта часто честный модульный монолит выгоднее; EH сознательно тренирует микросервисный стиль курса Kozirev.

## Sync vs async взаимодействие

- **Sync (gRPC/HTTP)**: нужен ответ здесь и сейчас (login, get balance, purchase orchestration на критическом пути).
- **Async (NATS/Kafka)**: уведомления, проекции, fan-out, то что можно eventual (History, Analytics, Fulfillment).
- Правило: синхронно — минимум зависимостей на запросе пользователя; побочные эффекты — в события.
- Таймауты, retries, circuit breaker обязательны на sync-пути (в EH — gateway circuit).

## Карта сервисов Event Horizon (упрощённо)

| Сервис | Роль | Связи |
|--------|------|-------|
| **gateway** | Gin HTTP/WS → gRPC, JWT, OpenAPI `/docs` | все публичные API |
| **auth** | регистрация/логин, JWT, сессии Redis | Profile/events `user.registered` |
| **profile** | профиль пользователя | NATS score/user |
| **billing** | баланс билетов, outbox `balance.updated` | Shop (Spend/Add) |
| **shop** | витрина, покупка, Saga/compensation | Billing, Inventory sync, NATS/Kafka |
| **inventory** | товары (PG+Mongo), outbox `inventory.item.created` | Shop |
| **payment** | платежи, `payment.completed` | History/Analytics |
| **fulfillment** | выдача после оплаты | Kafka `purchase.paid` |
| **notification** | уведомления | Kafka/NATS |
| **game / leaderboard** | игровой контур, scores | `score.updated`, Redis top |
| **history** | audit trail PG | durable NATS consumers |
| **analytics** | ClickHouse агрегаты | durable NATS consumers |
| **authors** | авторы, `author.upserted` | History/Analytics |
| **nats-hub** | инициализация streams | NATS cluster |

## Данные и consistency

- Каждый сервис — свой schema/DB где возможно (billing PG, inventory PG/Mongo, analytics CH).
- Кросс-сервисные инварианты — через **Saga/компенсацию** и **Outbox**, не через 2PC.
- Чтение чужих данных: API или проекция из событий (CQRS lite), не shared table.
- Идемпотентность на границах: повтор gRPC и повтор события не должны ломать деньги/сток.

## Организация команд и деплой

- Ideal: команда владеет сервисом end-to-end (code, proto, метрики, алерты).
- Контракты версионируй (proto backward compatible); breaking changes — осознанно.
- Независимый деплой: обратная совместимость API/событий на время раскатки.
- В EH образы `Dockerfile.*.bin`, compose/k3s; секреты не в git.

## Антипаттерны

- Chatty sync: 10 gRPC hop'ов на один HTTP-запрос.
- Shared database между «сервисами».
- События без схемы и без версии payload.
- Отсутствие correlation/trace id на всём пути.
- Бизнес-логика в gateway (gateway — BFF/proxy, не God-service).

## Типичные вопросы на собесе

- Как нарезать bounded context и как понять, что граница плохая?
- Когда оставить монолит, а когда резать на сервисы?
- Sync vs async: критерии выбора на критическом пути покупки?
- Что такое распределённый монолит?
- Как в EH устроена карта сервисов и кто пишет `shop.purchased` / `purchase.paid`?
- Как обеспечить независимость деплоя при изменении proto?
- Где хранить «источник истины» для баланса и для DAU?
