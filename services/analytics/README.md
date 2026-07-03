Не забудь добавить :

    "net/http"
    _ "net/http/pprof"

В сервис


## UPD INFO 3rd of July, 2026:
```bash
cd ~/event_horizon/services/analytics
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o analytics-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d analytics
```