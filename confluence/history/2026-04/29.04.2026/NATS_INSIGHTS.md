# NATS

NATS — это pure pub/sub, ближе к Kafka, чем к RabbitMQ, но с нюансами.

Характеристика	        RabbitMQ	                Kafka	                    NATS JetStream

Модель	                Очереди (push)	            Логи (pull)	                Потоки (pull/push)
Outbox паттерн	        ✅ Через транзакции	        ✅ Через transactions	    ✅ Есть через JetStream
Хранение	            RAM/диск (опционально)	    Диск (по умолчанию)	        Диск (в JetStream)
Сохранность сообщений	Да (persistent)	            Да	                        Да (JetStream)
Скорость	            Средняя	                    Высокая	                    Очень высокая
Сложность	            Средняя	                    Высокая	                    Низкая/Средняя

# Outbox в NATS

Через JetStream и транзакции в БД:

```go
// Паттерн outbox в Go + NATS
func CreateUser(ctx context.Context, db *pgxpool.Pool, js nats.JetStream) error {
    tx, _ := db.Begin(ctx)
    
    // 1. Сохраняем пользователя в БД
    tx.Exec(ctx, "INSERT INTO users...")
    
    // 2. Сохраняем событие в outbox таблицу
    tx.Exec(ctx, "INSERT INTO outbox (event_type, payload) VALUES ('user.created', ...)")
    
    // 3. Коммитим транзакцию
    tx.Commit(ctx)
    
    // 4. Отправляем в NATS (уже после коммита)
    js.Publish("user.created", payload)
    
    // ИЛИ ещё надежнее: отдельный воркер, который читает outbox и отправляет в NATS
}
```

# Для leaderboard в проекте:

```go
// Game сервис публикует рекорд
js.Publish("score.updated", scoreEvent)

// Leaderboard сервис подписан
js.Subscribe("score.updated", func(msg *nats.Msg) {
    // Обновляем Redis sorted set
    redis.ZAdd(ctx, "leaderboard:flappy", &redis.Z{Score: score, Member: userID})
    
    // Рассылаем новый топ-10 всем WebSocket клиентам
    broadcastTop10()
})
```

# NATS даёт:

Exactly-once доставку (через JetStream + idempotency keys)
Историю сообщений (как Kafka) — можешь переиграть события
Очень высокую скорость (микросекунды)


# Архитектура с NATS (как я вижу)

```text
Game Service ──┐
Billing ───────┼──► NATS JetStream ──┬──► Leaderboard (обновление топа)
Auth ──────────┘                      ├──► Analytics (логирование событий)
                                      └──► Notification (алерты о рекордах)
```

Leaderboard ──► WebSocket ──► Клиенты (увидят топ в реальном времени)

Это pub/sub. Game не знает о Leaderboard. 
Leaderboard просто слушает топик score.updated.