Технический долг: NATS конфигурация

Проблема

Gateway использует hardcoded nats://event-horizon-nats:4222
Game, Billing, Leaderboard используют cfg.NATSUrl с дефолтом nats://localhost:4222

Auth вообще не использует NATS (возможно, так и задумано)

В Docker Compose нет переменной NATS_URL для сервисов

Что нужно сделать

Добавить NATS_URL=nats://event-horizon-nats:4222 во все сервисы в docker-compose.cluster.yml:
billing
game
leaderboard
gateway (хотя у него hardcode, но для единообразия)

В идеале - убрать hardcode из Gateway и использовать cfg.NATSUrl как все

Проверить, нужно ли NATS для Auth

Приоритет: 🟡 Средний

Не критично для работы, но важно для консистентности и будущего масштабирования.