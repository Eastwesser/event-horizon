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

Download nats:

curl -sf https://binaries.nats.dev/nats-io/natscli/nats@latest | sh
sudo mv nats /usr/local/bin/

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

cd /home/denismatveev/event_horizon

# 1. Сборка бинарника NATS Hub
cd services/nats-hub
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub main.go
cd ../..

# 2. Docker образ
docker build -t eastwesser/nats-hub:latest -f Dockerfile.nats-hub.bin .

# 3. Пуш
docker push eastwesser/nats-hub:latest

# 4. Перезапуск только NATS Hub
docker-compose -f deployments/docker-compose.cluster.yml stop nats-hub
docker-compose -f deployments/docker-compose.cluster.yml rm -f nats-hub
docker-compose -f deployments/docker-compose.cluster.yml up -d nats-hub

# 5. Проверка
curl http://localhost:9097/health

🔍 Проверка NATS без CLI (через HTTP)
NATS имеет HTTP API для мониторинга:

bash
# 1. Статус NATS-1
curl -s http://localhost:8222/varz | jq .

# 2. Статистика Stream'ов
curl -s http://localhost:8222/streams | jq .

# 3. Информация о Stream EVENTS
curl -s http://localhost:8222/streams/EVENTS | jq .

# 4. Количество сообщений в Stream
curl -s http://localhost:8222/streams/EVENTS | jq '.state.messages'

# 5. Все consumer'ы
curl -s http://localhost:8222/streams/EVENTS/consumers | jq .
📊 Проверка через NATS Exporter
bash
# Метрики NATS
curl -s http://localhost:7777/metrics | grep -E "nats_(streams|messages|consumers)" | head -10
🚀 Быстрая проверка NATS через HTTP API
bash
echo "🔍 NATS Cluster Status:"
curl -s http://localhost:8222/varz | jq '{server_name: .server_name, version: .version, uptime: .uptime, connections: .connections, subscriptions: .subscriptions}'

echo ""
echo "🔍 Stream EVENTS:"
curl -s http://localhost:8222/streams/EVENTS | jq '{name: .config.name, messages: .state.messages, bytes: .state.bytes, subjects: .config.subjects}'

echo ""
echo "🔍 Consumers:"
curl -s http://localhost:8222/streams/EVENTS/consumers | jq '.[] | {name: .name, subject: .config.filter_subject, delivered: .delivered}'
📋 Итоговый чек-лист
Проверка	Команда	Ожидаемый результат
NATS статус	curl -s http://localhost:8222/varz | jq .server_name	nats-1
Stream EVENTS	curl -s http://localhost:8222/streams/EVENTS | jq .state.messages	Число > 0
NATS Hub логи	docker logs event-horizon-nats-hub --tail 5	✅ Stream EVENTS exists
Health check	curl http://localhost:9097/health	{"status":"ok"}