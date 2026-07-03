# звёздный поток данных

# TREE

```text
services/leaderboard/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── repository/
│   │   └── redis_repo.go
│   ├── service/
│   │   └── leaderboard_service.go
│   └── handler/
│       └── grpc_handler.go
├── proto/
│   ├── leaderboard.proto
│   ├── leaderboard.pb.go
│   └── leaderboard_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```


## GRPC GENERATION
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/leaderboard.proto
```       

# COMPILE
```bash
cd ~/event_horizon/services/leaderboard
go mod tidy
go build -o leaderboard-service ./cmd/main.go
```

# Запустить Redis для leaderboard (уже должен быть)
docker ps | grep redis-leaderboard

# Запустить сервис
./leaderboard-service

# В другом терминале — проверить
grpcurl -plaintext localhost:50054 list
# Должен показать leaderboard.LeaderboardService

# Обновить счёт
grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "user_id": "user-123",
  "user_email": "alice@example.com",
  "score": 1500
}' localhost:50054 leaderboard.LeaderboardService/UpdateScore

# Получить топ-10
grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "limit": 5
}' localhost:50054 leaderboard.LeaderboardService/GetTopScores

--
# 1. Проверить, какие сервисы доступны
grpcurl -plaintext localhost:50054 list

# Должно показать:
# leaderboard.LeaderboardService
# grpc.reflection.v1alpha.ServerReflection

# 2. Проверить методы
grpcurl -plaintext localhost:50054 list leaderboard.LeaderboardService

# Должно показать:
# GetPlayerRank
# GetTopScores
# UpdateScore

# 3. Добавить несколько рекордов
grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "user_id": "user-001",
  "user_email": "alice@example.com",
  "score": 1500
}' localhost:50054 leaderboard.LeaderboardService/UpdateScore

grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "user_id": "user-002",
  "user_email": "bob@example.com",
  "score": 2300
}' localhost:50054 leaderboard.LeaderboardService/UpdateScore

grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "user_id": "user-003",
  "user_email": "carol@example.com",
  "score": 1800
}' localhost:50054 leaderboard.LeaderboardService/UpdateScore

# 4. Получить топ-5
grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "limit": 5
}' localhost:50054 leaderboard.LeaderboardService/GetTopScores

# Ожидаемый ответ:
# {
#   "entries": [
#     {"rank": 1, "userId": "user-002", "userEmail": "bob@example.com", "score": 2300},
#     {"rank": 2, "userId": "user-003", "userEmail": "carol@example.com", "score": 1800},
#     {"rank": 3, "userId": "user-001", "userEmail": "alice@example.com", "score": 1500}
#   ]
# }

# 5. Получить ранг конкретного игрока
grpcurl -plaintext -d '{
  "game_id": "hexagon",
  "user_id": "user-001"
}' localhost:50054 leaderboard.LeaderboardService/GetPlayerRank

# Ожидаемый ответ: rank: 3, score: 1500

# REDIS FOR LEADERBOARD:

```text
cd ~/event_horizon

# Проверить, какие контейнеры вообще запущены
docker ps

# Если redis-leaderboard нет в списке — поднять всё заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверить, что все Redis'ы запустились
docker ps | grep redis

# Должны быть:
# event-horizon-redis
# event-horizon-redis-game
# event-horizon-redis-billing
# event-horizon-redis-leaderboard  👈 этот нужен

# Проверить, что Redis leaderboard жив
docker exec event-horizon-redis-leaderboard redis-cli ping
# Должно быть: PONG

# Проверить порт
docker port event-horizon-redis-leaderboard
# Должно быть: 6382/tcp -> 0.0.0.0:6382

# Проверить подключение с хоста
redis-cli -h 127.0.0.1 -p 6382 ping
# Должно быть: PONG
```


## UPD INFO 3rd of July, 2026:

```bash
cd ~/event_horizon/services/leaderboard
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o leaderboard-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.leaderboard.bin -t eastwesser/leaderboard:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d leaderboard
```