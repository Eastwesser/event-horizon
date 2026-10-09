# Сервисы Event Horizon — по одной фразе (интервью)

| Сервис | Порт | Зачем |
|--------|------|--------|
| **Balancer** | HTTP 8079 | Единая точка входа: HTTP + WS, балансирует gateway |
| **Gateway** | 8081–8083 | JWT, REST→gRPC, OpenAPI `/docs`, rate limit |
| **Auth** | gRPC 50051 | Регистрация/логин, bcrypt 12, сессии в Redis |
| **Game** | 50052 | Submit счёта, boost validation, запись рекордов |
| **Billing** | 50053 | Лампочки (boost) и билетики (магазин) |
| **Leaderboard** | 50054 | Топы в Redis Sorted Set → WS клиенту |
| **Shop** | 50055 | Каталог/покупка/отмена, outbox `shop.purchased` |
| **Inventory** | 50059 | Витрина авторов (PG+Mongo), outbox создания товаров |
| **Profile** | 50060 | Ник, best scores, ачивки |
| **Payment** | 50058 | Подписка / CanPurchaseMerch для физ. мерча |
| **Authors** | 50061 | Кабинет автора, заяв, заяв |
| **History** | 50062 | Лента событий из NATS (register, score, shop, …) |
| **Analytics** | 50057 | DAU/MAU/retention в ClickHouse |
| **Fulfillment** | metrics | Consumer `purchase.paid` → выдача/логистика сигнала |
| **Notification** | metrics | In-app уведомления (колокольчик) |
| **NATS Hub** | — | Bootstrap JetStream stream `EVENTS` |
| **FE React** | 5173 dev | Игры + магазин; `baseURL: '/api/v1'` |

**Deploy:** `make deploy` = thin (NATS). `make deploy-heavy` = +Kafka (опционально).
