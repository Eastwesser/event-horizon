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