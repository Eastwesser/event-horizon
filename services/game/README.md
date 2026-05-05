# игровая логика

```text
services/game/
├── cmd/main.go
├── games/
│   └── hexagons/
│       ├── board.go          # логика доски
│       ├── validator.go      # валидация ходов
│       └── generator.go      # генерация начального состояния
├── internal/
│   ├── config/config.go
│   ├── repository/
│   │   └── game_repo.go      # PostgreSQL для рекордов игроков
│   ├── validator/
│   │   └── move_validator.go
│   ├── service/
│   │   └── game_service.go
│   └── handler/
│       └── grpc_handler.go
├── proto/
│   ├── game.proto
│   ├── game.pb.go
│   └── game_grpc.pb.go
└── go.mod
```

# 1. Проверить, что сервис доступен
grpcurl -plaintext localhost:50052 list
# Должно быть: game.GameService

# 2. Проверить методы
grpcurl -plaintext localhost:50052 list game.GameService
# Должно быть: GetGameInfo, SubmitScore

# 3. Получить информацию об игре
grpcurl -plaintext -d '{"game_id":"hexagon"}' \
  localhost:50052 game.GameService/GetGameInfo

# 4. Отправить рекорд (валидация пока заглушка)
grpcurl -plaintext -d '{
  "user_id": "user-001",
  "game_id": "hexagon",
  "level": 3,
  "score": 250,
  "seed": "test_seed_123",
  "moves": []
}' localhost:50052 game.GameService/SubmitScore

# 5. Success
grpcurl -plaintext -d '{
  "user_id": "user-001",
  "game_id": "hexagon",
  "level": 3,
  "score": 250,
  "seed": "test_seed_123",
  "moves": [
    {
      "fromX": 0, "fromY": 0,
      "toX": 1, "toY": 1,
      "timestamp": 1000
    }
  ]
}' localhost:50052 game.GameService/SubmitScore
{
  "success": true,
  "newHighscore": 250,
  "message": "score submitted successfully",
  "lampsEarned": 10,
  "ticketsEarned": 2
}
