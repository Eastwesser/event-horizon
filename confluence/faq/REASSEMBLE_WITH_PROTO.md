## Шаг 1: Перегенерировать proto для всех сервисов

```bash
# Для leaderboard:
cd /home/denismatveev/event_horizon/services/leaderboard
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

# Для game:
cd /home/denismatveev/event_horizon/services/game
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

# Для auth:
cd /home/denismatveev/event_horizon/services/auth
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

# Для billing:
cd /home/denismatveev/event_horizon/services/billing
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

# Для inventory:
cd /home/denismatveev/event_horizon/services/inventory
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

# Или одна команда для всех сразу:

cd /home/denismatveev/event_horizon
for svc in auth game billing leaderboard; do
  cd services/$svc
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
  echo "✅ $svc proto regenerated"
  cd ../..
done

cd ~/event_horizon
echo "🎉 All protos regenerated!"
```

## Шаг 2: Пересобрать все сервисы

```bash
cd /home/denismatveev/event_horizon

# Game
cd services/game
go build -o game-service ./cmd/main.go
echo "✅ Game service built"

# Leaderboard
cd ../leaderboard
go build -o leaderboard-service ./cmd/main.go
echo "✅ Leaderboard service built"

# Auth
cd ../auth
go build -o auth-service ./cmd/main.go
echo "✅ Auth service built"

# Billing
cd ../billing
go build -o billing-service ./cmd/main.go
echo "✅ Billing service built"

# Gateway
cd ../gateway
go build -o gateway ./cmd/main.go
echo "✅ Gateway service built"

# Inventory
cd ../inventory
go build -o inventory-service ./cmd/main.go
echo "✅ Inventory service built"

cd ~/event_horizon
echo "🎉 All services built!"
```

## Шаг 3: Полный перезапуск

```bash
cd /home/denismatveev/event_horizon

# Остановить всё
make stop-services

# Остановить и очистить Docker контейнеры
docker-compose -f deployments/docker-compose.cluster.yml down -v

# Поднять Docker контейнеры заново
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Запустить все Go сервисы
make all

# Проверить статус
make status

# Проверить Gateway
curl http://localhost:8080/health

echo "🎉 Full restart completed!"
```

## Или одной командой через твой скрипт:

```bash
cd /home/denismatveev/event_horizon
./restart.sh
```

## Если всё прошло успешно — проверить nickname в лидерборде:

```bash
# 1. Создать тестового пользователя (если нет)
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"nicktest@example.com","password":"123456"}'

# 2. Залогиниться (получить токен и user_id)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"nicktest@example.com","password":"123456"}'

# 3. Отправить рекорд с nickname (замени токен и user_id)
curl -X POST http://localhost:8080/api/game/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "user_id": "<YOUR_USER_ID>",
    "game_id": "hexagon",
    "level": 1,
    "score": 777,
    "user_email": "nicktest@example.com",
    "nickname": "ТестовыйНик",
    "seed": "test_seed",
    "moves": []
  }'

# 4. Проверить лидерборд

curl "http://localhost:8080/api/leaderboard?game_id=hexagon&limit=5" | jq .

# Должен увидеть "nickname": "ТестовыйНик"
```
