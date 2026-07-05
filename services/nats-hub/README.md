# 🧩 NATS Hub

**Один файл, одна задача** — создать Stream'ы в NATS, чтобы другие сервисы могли публиковать и подписываться на события.

---

## 📌 Зачем это нужно

В Event Horizon используется **два канала связи**:

| Канал | Протокол | Используется для |
|-------|----------|------------------|
| **gRPC** | Синхронный | Запросы-ответы (Auth, Game, Billing, Leaderboard) |
| **NATS** | Асинхронный | События (score.updated, user.registered) |

**NATS Hub** — это инфраструктурный сервис, который:
- Создаёт Stream `EVENTS` при старте.
- Содержит список всех subject'ов, которые используются в системе.
- Ничего больше не делает — просто висит и ждёт.

---

## 🔄 Как это работает

```text
[nats-hub] ── создаёт Stream ──► [NATS]
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
         ▼                          ▼                          ▼
    [Game]                     [Leaderboard]               [Billing]
(публикует score.updated)   (подписан на score.updated)  (подписан на score.updated)
                                    │
                                    ▼
                               [Profile]
                         (подписан на score.updated)
                                    │
                                    ▼
                              [Gateway]
                        (подписан → WebSocket → клиент)
```

## 📋 Stream EVENTS

| Subject | Кто публикует | Кто подписан |
|---------|---------------|--------------|
| `score.updated` | Game | Leaderboard, Billing, Profile, Gateway |
| `user.registered` | Auth | Profile |
| `shop.purchased` | Shop (в планах) | Analytics (в планах) |
| `payment.completed` | Payment (в планах) | Billing, Analytics (в планах) |

---

## 🚀 Как запустить

### 1. Собрать бинарник

```bash
cd ~/event_horizon/services/nats-hub
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub main.go
```

2. Собрать Docker-образ
```bash
docker build -f Dockerfile.nats-hub.bin -t eastwesser/nats-hub:latest .
```

3. Запустить в Docker Compose
```yaml
nats-hub:
  image: eastwesser/nats-hub:latest
  environment:
    - NATS_URL=nats://event-horizon-nats:4222
  depends_on:
    - nats
  networks:
    - event-horizon-net
  restart: on-failure
```

✅ Проверка

```bash
# Проверить, что Hub создал Stream
docker logs deployments-nats-hub-1 --tail=10

# Посмотреть Stream в NATS
nats stream info EVENTS -s nats://localhost:4222
```

Если в логах есть:

```text
✅ Stream EVENTS created
```
— значит, всё работает.

🧠 Почему именно так
Аспект	Что даёт
Надёжность	Stream создаётся один раз, независимо от остальных сервисов
Простота	Каждый сервис знает только о своих subject'ах
Масштабируемость	Добавляешь новый subject → обновляешь nats-hub → перезапускаешь
Тестируемость	Можно запустить nats-hub отдельно и проверить Stream'ы

🔮 Дальше
При добавлении нового сервиса (Shop, Analytics, Payment) — просто добавляешь его subject в список в main.go и обновляешь образ:

```go
Subjects: []string{
    "event.>",
    "score.updated",
    "user.registered",
    "shop.purchased",     // новый
    "payment.completed",  // новый
},
```

Всё остальное делает NATS Hub. 🚀