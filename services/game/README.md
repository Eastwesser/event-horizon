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