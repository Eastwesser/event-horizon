# Пересборка сервисов с pprof

## Один сервис (пример для Gateway)
cd ~/event_horizon/services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d gateway gateway-2 gateway-3

## Все сервисы
for service in auth billing game gateway leaderboard balancer; do
    cd services/$service
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${service}-service ./cmd/main.go
    cd ../..
    docker build -f Dockerfile.${service}.bin -t eastwesser/${service}:latest .
done
docker-compose -f deployments/docker-compose.cluster.yml up -d

# PPROF TOOLS
go tool pprof -http=:8084 /tmp/cpu.prof


```bash
# Текстовый режим (самый полезный)
go tool pprof -top /tmp/cpu.prof
go tool pprof -text /tmp/cpu.prof | head -50

# Подробно по функции
go tool pprof -list="main" /tmp/cpu.prof
go tool pprof -list="SubmitScore" /tmp/cpu.prof

# Установить Graphviz для визуализации
sudo pacman -S graphviz
go tool pprof -png /tmp/cpu.prof > /tmp/cpu.png

# Или скопировать на локальную машину
scp denismatveev@c0der:/tmp/cpu.prof ~/Downloads/
go tool pprof -http=:8080 ~/Downloads/cpu.prof
```