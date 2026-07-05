🚀 Шаг 1: Перегенерировать protobuf
bash
cd ~/event_horizon/services/auth
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/auth.proto

🚀 Шаг 2: Пересобрать Auth и Gateway
bash
# Auth
cd ~/event_horizon/services/auth
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.auth.bin -t eastwesser/auth:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d auth

# Gateway
cd ~/event_horizon/services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3

🚀 Шаг 3: Проверить
bash
# 1. Получить токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# 2. Обновить никнейм
curl -X POST http://localhost:8079/api/auth/update-nickname \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"nickname":"NewNick"}' | jq '.'

# 3. Проверить баланс
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 1. Получить токен для первого пользователя
TOKEN1=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# 2. Сменить ник
curl -X POST http://localhost:8079/api/auth/update-nickname \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN1" \
  -d '{"nickname":"NewNick"}' | jq '.'
