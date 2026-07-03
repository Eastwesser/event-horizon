Не забудь добавить :

    "net/http"
    _ "net/http/pprof"

В сервис


## UPD INFO 3rd of July, 2026:

cd ~/event_horizon/services/notification
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o notification-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.notification.bin -t eastwesser/notification:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d notification