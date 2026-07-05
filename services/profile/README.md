# 👤 Profile Service

**Агрегатор пользовательских данных.**  
Собирает информацию из разных микросервисов в одном месте, чтобы фронтенд мог получить полный профиль за один запрос.

---

## 🧠 Зачем это нужно

В Event Horizon **данные о пользователе разбросаны по разным сервисам**:

| Что | Где лежит |
|-----|-----------|
| Email, никнейм | Auth (PostgreSQL) |
| Рекорды по играм | Game (PostgreSQL) |
| Лампочки, билетики | Billing (PostgreSQL) |
| Топ-10 | Leaderboard (Redis) |

Вместо того чтобы фронтенд делал 4 запроса и собирал всё сам, **Profile Service** делает это за него.

---

## 🏗️ Архитектура (CQRS + Read Model)

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                          WRITE SIDE (События)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [Game] ──► score.updated ──► [Profile Service] ──► [Profile DB]          │
│  (рекорды, лампочки, билетики)                                              │
│                                                                             │
│  [Auth] ──► user.registered ──► [Profile Service] ──► [Profile DB]        │
│  (email, никнейм)                                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           READ SIDE (Запросы)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [Gateway] ──► gRPC ──► [Profile Service] ──► [Profile DB]                │
│  GET /api/profile                                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
Почему это CQRS?

Запись (Write) — через события (NATS).

Чтение (Read) — через gRPC.

Разные модели — данные для чтения агрегированы и денормализованы.

📊 Хранилище
Одна таблица — весь профиль.

sql
CREATE TABLE user_profiles (
    user_id     UUID PRIMARY KEY,
    email       TEXT NOT NULL,
    nickname    TEXT,
    total_score INT DEFAULT 0,
    best_scores JSONB DEFAULT '{}',  -- {"hexagon": 6349, "flappy": 200, ...}
    lamps       INT DEFAULT 0,
    tickets     INT DEFAULT 0,
    updated_at  TIMESTAMP DEFAULT NOW()
);

🛠️ gRPC методы
Метод	Описание	Кто вызывает
GetProfile	Получить профиль пользователя	Gateway (/api/profile)
UpdateProfile	Обновить профиль (внутренний)	Profile Service (сам себя)
🔌 NATS подписки
Subject	Откуда	Что обновляем
score.updated	Game	total_score, best_scores, lamps, tickets
user.registered	Auth	email, nickname
🚀 Сборка
bash
cd ~/event_horizon/services/profile

# 1. Скачать зависимости
go mod tidy

# 2. Сгенерировать proto
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/profile.proto

# 3. Собрать бинарник
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o profile-service ./cmd/main.go

# 4. Собрать Docker-образ
cd ~/event_horizon
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .

✅ Проверка
bash
# 1. Проверить, что Profile слушает
curl -s http://localhost:9099/metrics | head -5

# 2. Получить профиль через Gateway
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

curl -X GET http://localhost:8079/api/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'
Пример ответа:

json
{
  "user_id": "7f63d6fc-b000-4667-b8ce-4fa24da9920a",
  "email": "s1ntezc0der@gmail.com",
  "nickname": "Kotislaw",
  "total_score": 6349,
  "best_scores": {
    "hexagon": 6349,
    "flappy": 200,
    "memory": 720,
    "towers": 1540
  },
  "lamps": 123,
  "tickets": 456
}

🧠 Почему именно так
Аспект	Что даёт
Один запрос	Фронтенд не собирает данные из 4 сервисов
Асинхронное обновление	Profile не тормозит Game/Auth
Денормализация	Данные уже готовы к выдаче
Масштабируемость	Profile можно реплицировать независимо

🔮 Дальше
При добавлении нового поля в профиль:

Добавить поле в user_profiles таблицу.

Добавить поле в proto/profile.proto.

Обновить обработчики NATS событий.

Пользователь получает обновлённый профиль без изменения кода на фронтенде. 🚀


# CQRS pattern

┌─────────────────────────────────────────────────────────────────────────────┐
│                          WRITE SIDE (Events)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [Game] ──► NATS ──► [Profile Updater] ──► [PostgreSQL (profile_db)]        │
│  (score.updated)                                                            │
│                                                                             │
│  [Auth] ──► NATS ──► [Profile Updater] ──► [PostgreSQL (profile_db)]        │
│  (user.registered)                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           READ SIDE (Queries)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [Gateway] ──► gRPC ──► [Profile Service] ──► [PostgreSQL (profile_db)]     │
│  GET /api/profile                                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

cd ~/event_horizon/services/profile

# 1. Скачать зависимости
go mod init github.com/Eastwesser/event-horizon/services/profile
go get github.com/jackc/pgx/v5
go get github.com/nats-io/nats.go
go get google.golang.org/grpc
go get google.golang.org/protobuf

# 2. Сгенерировать proto
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/profile.proto

# 3. Собрать
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o profile-service ./cmd/main.go

# 4. Собрать Docker-образ
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .
docker push eastwesser/profile:latest

# 5. Пересобрать
cd ~/event_horizon

# Собрать образ
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .

# Запушить (если нужно)
docker push eastwesser/profile:latest

# Перезапустить всё
docker-compose -f deployments/docker-compose.cluster.yml up -d

===============
4. Пересобираем и перезапускаем
bash
cd ~/event_horizon

# 1. Пересобрать образ Profile
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .

# 2. Пересобрать Gateway
cd services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd ../..
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

# 3. Перезапустить всё
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

===============

# Логи Profile
docker logs deployments-profile-1 --tail=20

# Проверить, что Profile слушает
curl -s http://localhost:9099/metrics | head -5


🎯 Итог

Profile Service:

Собирает данные из Auth (через NATS user.registered)

Собирает данные из Game (через NATS score.updated)

Хранит агрегированный профиль в отдельной БД

Отдаёт единый профиль через gRPC

Теперь Gateway может просто обращаться к Profile Service, а не собирать данные из разных мест. Это и есть CQRS + Read Model в чистом виде! 🚀


# Profile Service

Сервис для агрегации данных пользователя (CQRS + Read Model).

## Архитектура

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                          WRITE SIDE (Events)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Game] ──► NATS ──► [Profile Updater] ──► [PostgreSQL (profile_db)]       │
│  (score.updated)                                                            │
│  [Auth] ──► NATS ──► [Profile Updater] ──► [PostgreSQL (profile_db)]       │
│  (user.registered)                                                          │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           READ SIDE (Queries)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Gateway] ──► gRPC ──► [Profile Service] ──► [PostgreSQL (profile_db)]    │
│  GET /api/profile                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
Хранилище
sql
CREATE TABLE user_profiles (
    user_id     UUID PRIMARY KEY,
    email       TEXT NOT NULL,
    nickname    TEXT,
    total_score INT DEFAULT 0,
    best_scores JSONB DEFAULT '{}',
    lamps       INT DEFAULT 0,
    tickets     INT DEFAULT 0,
    updated_at  TIMESTAMP DEFAULT NOW()
);
gRPC методы
Метод	Описание
GetProfile	Получить профиль пользователя
UpdateProfile	Обновить профиль (используется внутри)
Сборка
bash
cd ~/event_horizon/services/profile
go mod tidy
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/profile.proto
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o profile-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .
Проверка
bash
# Проверить, что Profile слушает
curl -s http://localhost:9099/metrics | head -5

# Получить профиль через Gateway
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

curl -X GET http://localhost:8079/api/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'