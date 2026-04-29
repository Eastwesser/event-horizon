## План на 30.04.2026

1. Добавить PostgreSQL/Redis для Game, Billing, Leaderboard в docker-compose
2. Gateway: базовая структура, HTTP → Auth gRPC прокси
3. NATS тест: отправить событие из Gateway, поймать в отдельном воркере
4. Leaderboard: Redis Sorted Set + подписка на NATS
Вопрос к тебе перед отдыхом

"Завтра поговорим про NATS и связь в нашем проекте."
Какую часть NATS хочешь разобрать подробнее?

А) JetStream vs Core NATS — когда что использовать?
Б) Кластер из 3 нод — как настроить и протестировать отказоустойчивость
В) Exactly-once доставка — idempotency keys + outbox
Г) Паттерн "request-reply" поверх NATS — синхронные вызовы через асинхронную шину
Д) Сравнение с Kafka/RabbitMQ — для собеседований


План на завтра (30.04)

Утро

Docker Compose — добавить БД для Game, Billing, Leaderboard
NATS кластер — поднять 3 ноды или оставить 1 (решай по настроению)
День

Gateway — базовый HTTP → gRPC прокси в Auth
NATS тест — отправить test.event из Gateway, поймать в отдельном воркере
Вечер

Leaderboard — Redis Sorted Set + подписка на score.updated
Game — заглушка, которая публикует рекорды в NATS
Код, который тебя ждёт утром

Gateway (скелет, заполним завтра):

go
// services/gateway/cmd/main.go
package main

func main() {
    // 1. Подключиться к Auth по gRPC
    // 2. Создать HTTP сервер (Gin)
    // 3. POST /api/auth/register → authClient.Register
    // 4. POST /api/auth/login → authClient.Login
}
Game (заглушка):

go
// services/game/cmd/main.go
func main() {
    // 1. Подключиться к NATS
    // 2. gRPC метод SubmitScore
    // 3. Публикация в NATS: "score.updated"
}