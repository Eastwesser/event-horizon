# 1. Сгенерировать новые proto файлы
cd services/game
make proto  # или protoc --go_out=. --go-grpc_out=. proto/*.proto

cd ../leaderboard
make proto

# 2. Пересобрать сервисы
cd services/game && go build -o game-service ./cmd/main.go
cd services/leaderboard && go build -o leaderboard-service ./cmd/main.go
cd services/gateway && go build -o gateway ./cmd/main.go

# 3. Перезапустить
pkill -f "game-service|leaderboard-service|gateway"
cd services/game && ./game-service > /tmp/game.log 2>&1 &
cd services/leaderboard && ./leaderboard-service > /tmp/leaderboard.log 2>&1 &
cd services/gateway && ./gateway > /tmp/gateway.log 2>&1 &