🔁 Процесс заливки бинарников в Docker Hub
1. Собрать статический бинарник локально
bash
cd ~/event_horizon/services/gateway

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o gateway-service ./cmd/main.go
Что значат флаги:

CGO_ENABLED=0 — отключаем CGO (чистый Go)

GOOS=linux GOARCH=amd64 — собираем для Linux

-ldflags="-s -w" — уменьшаем размер бинарника

-extldflags=-static — статическая линковка

2. Проверить, что бинарник статический
bash
file gateway-service
# Должно быть: statically linked
3. Собрать Docker-образ из готового бинарника
bash
cd ~/event_horizon

# Создать Dockerfile (если нет)
cat > Dockerfile.gateway.bin << 'EOF'
FROM scratch
COPY services/gateway/gateway-service /gateway
EXPOSE 8080
CMD ["/gateway"]
EOF

# Собрать образ
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
Что важно: FROM scratch — пустой образ, в него кладётся только бинарник.

4. Запушить в Docker Hub
bash
# Залогиниться (один раз)
docker login -u eastwesser

# Запушить образ
docker push eastwesser/gateway:latest
5. Проверить, что образ есть
bash
docker images | grep gateway
# Или
docker pull eastwesser/gateway:latest
📦 Для всех сервисов (цикл)
bash
cd ~/event_horizon

for service in auth billing game leaderboard gateway; do
    echo "🔨 Building $service..."
    cd services/$service
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -extldflags=-static" \
        -o ${service}-service ./cmd/main.go
    cd ../..
    
    echo "🐳 Building image for $service..."
    docker build -f Dockerfile.${service}.bin -t eastwesser/${service}:latest .
    
    echo "📦 Pushing $service..."
    docker push eastwesser/${service}:latest
done

🔍 Проверить образы в Docker Hub

bash
# Список образов локально
docker images | grep eastwesser

# Логин и проверка
docker login -u eastwesser
docker pull eastwesser/auth:latest
docker pull eastwesser/billing:latest
docker pull eastwesser/game:latest
docker pull eastwesser/leaderboard:latest
docker pull eastwesser/gateway:latest

⚠️ Что НЕЛЬЗЯ делать

❌ Не пересобирать образ с генерацией protobuf внутри

❌ Не использовать FROM golang в Dockerfile (только FROM scratch)

❌ Не запускать go mod download или protoc внутри Docker-сборки

✅ Что работает

dockerfile

# Правильный Dockerfile
FROM scratch
COPY services/gateway/gateway-service /gateway
EXPOSE 8080
CMD ["/gateway"]