## UPD INFO 3rd of July, 2026:

```bash
cd ~/event_horizon/services/balancer
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o balancer-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.balancer.bin -t eastwesser/balancer:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d balancer
```
