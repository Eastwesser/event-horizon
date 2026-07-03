Не забудь добавить :

    "net/http"
    _ "net/http/pprof"

В сервис


## UPD INFO 3rd of July, 2026:

cd ~/event_horizon/services/payment
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o payment-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.payment.bin -t eastwesser/payment:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d payment