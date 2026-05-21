cd ~/event_horizon

# Останавливаем
make stop-services

# Пересобираем leaderboard (с правильным интерфейсом)
cd services/leaderboard
go build -o leaderboard-service ./cmd/main.go

# Пересобираем gateway
cd ../gateway
go build -o gateway ./cmd/main.go

# Пересобираем billing
cd ../billing
go build -o billing-service ./cmd/main.go

# Пересобираем game (на всякий случай)
cd ../game
go build -o game-service ./cmd/main.go

# Пересобираем auth
cd ../auth
go build -o auth-service ./cmd/main.go

# Запускаем всё
cd ~/event_horizon

make down

make up

make all

# Проверяем статус
make status