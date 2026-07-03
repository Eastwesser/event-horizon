for service in auth billing game leaderboard; do
    protoc \
        --proto_path=services/$service/proto \
        --go_out=services/$service/proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=services/$service/proto \
        --go-grpc_opt=paths=source_relative \
        services/$service/proto/$service.proto
done


2. Если нет коммита — пересобери бинарники в правильном порядке.

Проблема в том, что make start-services пытается собрать всё с нуля, а auth и gateway падают.

Собери их по одному, используя правильную версию protobuf:

bash
cd /home/denismatveev/event_horizon

# 1. Перегенерируй proto с правильными опциями (как вчера)
rm -f services/auth/proto/auth.pb.go services/billing/proto/billing.pb.go

for service in auth billing game leaderboard; do
    protoc \
        --proto_path=services/$service/proto \
        --go_out=services/$service/proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=services/$service/proto \
        --go-grpc_opt=paths=source_relative \
        services/$service/proto/$service.proto
done

# 2. Собери бинарники в правильном порядке
cd services/auth
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-service ./cmd/main.go
cd ../..

cd services/gateway
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway-service ./cmd/main.go
cd ../..

# 3. Запусти все сервисы по очереди
cd services/auth && ./auth-service &
cd services/billing && ./billing-service &
cd services/game && ./game-service &
cd services/leaderboard && ./leaderboard-service &
cd services/gateway && ./gateway-service &

# 4. Проверь
curl http://localhost:8080/health
Если не поможет — смотри, какие файлы лежат в services/auth/proto/ и services/billing/proto/:

bash
ls -la services/auth/proto/*.pb.go
ls -la services/billing/proto/*.pb.go