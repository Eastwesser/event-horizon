LEADERBOARD РАБОТАЕТ!!!

text
✅ Redis подключён (127.0.0.1:6382)
✅ gRPC сервер слушает (:50054)
✅ GetTopScores возвращает топ-3:
   1. bob@example.com — 2300 очков 🥇
   2. carol@example.com — 1800 очков 🥈
   3. alice@example.com — 1500 очков 🥉
Что у тебя есть сейчас

Сервис	Статус	Порт	Назначение
Auth	✅	50051	JWT, регистрация, логин
Gateway	✅	8080	HTTP → gRPC прокси, NATS publisher
Leaderboard	✅	50054	Топ-10, Redis Sorted Set
NATS JetStream	✅	4222	Событийная шина
PostgreSQL (4 шт)	✅	5460-5463	Базы данных
Redis (4 шт)	✅	6379-6382	Кеши/сессии/топ
Следующий шаг: NATS подписка в Leaderboard

Сейчас leaderboard получает рекорды только через gRPC (UpdateScore). Надо добавить подписку на NATS, чтобы получать рекорды из Game сервиса автоматически.

Добавим NATS subscriber в leaderboard

1. Добавить зависимость в go.mod:

bash
cd ~/event_horizon/services/leaderboard
go get github.com/nats-io/nats.go

```text
┌──────────────┐    HTTP     ┌──────────────┐    gRPC     ┌──────────────┐
│    curl      │ ──────────► │   Gateway    │ ──────────► │     Auth     │
│              │             │   :8080      │             │    :50051    │
└──────────────┘             └──────────────┘             └──────────────┘
                                   │
                                   │ NATS publish
                                   ▼
                            ┌──────────────┐
                            │     NATS     │
                            │  JetStream   │
                            │   :4222      │
                            └──────────────┘
                                   │
                                   │ NATS subscribe (score.updated)
                                   ▼
┌──────────────┐    gRPC     ┌──────────────┐    Redis    ┌──────────────┐
│   Gateway    │ ◄────────── │ Leaderboard  │ ◄─────────► │ Redis Leader │
│   (WebSocket)│   (в плане) │   :50054     │             │   :6382      │
└──────────────┘             └──────────────┘             └──────────────┘
```