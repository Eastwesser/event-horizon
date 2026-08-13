# Week 5 status — Event Horizon (13.08.2026) — CLOSED

## Verified (Emma offline download + builds)

From `offline_treminal_results.md`:
- [x] `go mod download` for all W5-related modules
- [x] `go get github.com/IBM/sarama@v1.45.2` in platform
- [x] `go build -tags kafka` for fulfillment, notification, shop
- [x] Docker images: kafka 3.8.1, alpine, postgres, redis, mongo
- [x] Default builds: fulfillment, notification, nats-hub, balancer, shop, gateway, auth

## Delivered earlier

- [x] DI on all services (`task di-check`)
- [x] Kafka KRaft in compose + cluster
- [x] `platform/pkg/kafka` (noop default / Sarama with `-tags kafka`)
- [x] `contracts/events` PurchasePaid / PurchaseFulfilled
- [x] fulfillment + notification services
- [x] Shop publishes `purchase.paid`
- [x] `task w5-check`

## Rebuild reminder (your machine)

```bash
for svc in shop fulfillment notification; do
  (cd services/$svc && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags kafka -ldflags="-s -w" -o ${svc}-service ./cmd/main.go)
  docker build -f Dockerfile.${svc}.bin -t eastwesser/${svc}:latest .
done
```
