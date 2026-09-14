# Урок 13 — 29 апреля 2026

HTTP + JSON: долгий CreateOrder с ответом за ≤1с, клик-статистика (unique/day), first-success search по репликам.

## Задачи

### 1. CreateOrder async
Usecase может идти до 10с, HTTP-ответ ≤1с. Если не успели — `202`/`200` «in progress» + orderID; отдельная ручка статуса. Фразы: «не отменяй бизнес-горутину context'ом HTTP»; «идемпотентность по UUID / ON CONFLICT»; «poll status».

### 2. Click stats
Уникальная пара `(user, author)` за календарные сутки UTC. `RegisterClick` + `StatsForAuthors(yesterday)`. Структура: `date → author → set(user)`.

### 3. First success search
N реплик, вернуть первый успешный ответ, `WithCancel` + `defer cancel()`. Буфер канала ≥ N — иначе утечка горутин на send. Фраза: «cancel не убивает блокирующий Search сам — нужен отменяемый API».

## Как запускать

```bash
export GOWORK=off
cd code/01_create_order_async && go run main.go
```
