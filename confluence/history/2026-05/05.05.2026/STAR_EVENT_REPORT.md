S.T.A.R. Report: EventHorizon Platform

Project: EventHorizon — игровая платформа с real-time leaderboard
Role: System Architect / Backend Engineer
Tech Stack: Go, gRPC, NATS JetStream, PostgreSQL, Redis, Docker Compose

S — Situation (Контекст)

Требовалось спроектировать и реализовать игровую платформу, способную выдерживать 10k RPS, с двумя ключевыми требованиями:

Real-time leaderboard — игроки должны видеть изменения в топ-10 мгновенно после установки рекорда
Микросервисная архитектура — пять независимых сервисов (Auth, Gateway, Game, Billing, Leaderboard) с возможностью горизонтального масштабирования
Дополнительное ограничение: Вся система должна подниматься из одной папки на Linux-виртуалке (docker-compose up).

T — Task (Задача)

Основная техническая задача: Выбрать и внедрить событийную шину, которая обеспечит:

Асинхронную передачу событий между микросервисами
Гарантированную доставку (результат игры не должен теряться)
Поддержку high-load (1000+ событий/сек)
Простоту разработки и отладки
Конкретный вызов: Game сервис публикует рекорды → Leaderboard получает их и обновляет топ-10 в Redis. Нужно было сделать это надёжно, без потери сообщений, с возможностью replay событий при необходимости.

A — Action (Действия)

Исследование

Рассмотрел три варианта:

RabbitMQ — знакомая технология, но избыточна для pub/sub
Apache Kafka — мощно, но тяжело поднимать и настраивать для pet-проекта
NATS JetStream — лёгкий, быстрый, с дисковым хранилищем и exactly-once доставкой
Выбрал NATS JetStream по причинам:

Один бинарник, запускается через docker-compose за минуту
Встроенное хранение сообщений (JetStream) — не боюсь потерять рекорд
Поддержка durable subscriptions — consumer может перезапуститься без потери событий
Нативная интеграция с Go (клиент nats.go)
Реализация

1. Поднял NATS в docker-compose.cluster.yml:

yaml
nats:
  image: nats:2.10-alpine
  command: ["-js"]  # JetStream mode
  ports:
    - "4222:4222"
    - "8222:8222"
2. В Gateway добавил публикацию событий:

go
js.Publish("score.updated", eventData)
3. В Leaderboard реализовал durable subscriber:

go
js.Subscribe("score.updated", func(msg *nats.Msg) {
    var event ScoreEvent
    json.Unmarshal(msg.Data, &event)
    leaderboardService.UpdateScore(...)
    msg.Ack()
}, nats.Durable("leaderboard-durable"), nats.ManualAck())
4. Добавил проверку через JetStream:

Создал stream SCORES с хранением на диске
Настроил MaxAge: 24h для автоматической очистки
Проверил, что при падении consumer'а сообщения не теряются
Интеграционное тестирование

bash
# Публикация события через nats CLI
nats pub "score.updated" '{"user_id":"user-001","score":250}' --server localhost:4222

# В логах leaderboard:
📡 Received score update via NATS: user=user-001 score=250
✅ Score updated for user-001, new rank: 4
R — Result (Результат)

Ключевые достижения:

✅ Real-time leaderboard работает — задержка от публикации до обновления Redis ~10-30 мс
✅ События не теряются — JetStream сохраняет сообщения на диск, durable subscription переживает рестарт
✅ Архитектура готова к scale — добавление второго инстанса Leaderboard не требует изменений (NATS сам разносит события)
✅ Простота разработки — вся система поднимается одной командой docker-compose up

Метрики (на локальной машине):

Публикация события: ~0.5-1 мс
Обработка в Leaderboard + обновление Redis: ~5-10 мс
Полная цепочка (Gateway → Game → NATS → Leaderboard → Redis): ~36 мс (по данным Gin логов)
Бонус: NATS CLI оказался удобным инструментом для отладки:

bash
nats sub "score.>" --server localhost:4222  # подписка на все события
nats stream info SCORES                      # статистика очереди
What I Learned / Next Steps

Что дал NATS по сравнению с Kafka/RabbitMQ:

Невероятно прост в настройке (один контейнер, одна команда)
Достаточно быстро для 10k RPS (микросекунды на публикацию)
JetStream даёт persistence без головной боли с Zookeeper
Что добавить дальше:

NATS кластер из 3 нод для отказоустойчивости
Prometheus метрики (NATS уже отдаёт их на порту 8222)
Outbox паттерн для exactly-once в Billing сервисе
Личный вывод: NATS — идеальный выбор для микросервисных pet-проектов и стартапов. Сочетает простоту RabbitMQ и надёжность Kafka, но без их сложности.

P.S. Ментору привет от Хохо. Скажи, что мы тут построили систему, где рекорды летают быстрее, чем блины пекутся. 🥞🚀