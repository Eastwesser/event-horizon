# Запускаем всё вместе

# Терминал 1 — базы данных:

```bash
cd ~/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d
```

# Терминал 2 — Auth:

```bash
cd ~/event_horizon/services/auth
./auth-service
```

# Терминал 3 — Gateway:

```bash
cd ~/event_horizon/services/gateway
go run cmd/main.go
```

# Терминал 4 — NATS subscriber (для теста):

```bash
# Установка nats CLI (если ещё нет)
go install github.com/nats-io/natscli/nats@latest

# Подписка на события
nats sub "event.>" --server localhost:4222
```

# Тестируем

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"nats@example.com","password":"secret123"}'

# Логин
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"nats@example.com","password":"secret123"}'
```

В терминале с NATS подпиской ты должен увидеть:

```text
[#1] Received on "event.user.registered"
{"event":"user.registered","user_id":"...","email":"nats@example.com"}

[#2] Received on "event.user.logged_in"
{"event":"user.logged_in","email":"nats@example.com"}
```