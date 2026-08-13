# Payment service status — Event Horizon (13.08.2026)

## Done (Stage 1 / voice-message)

- [x] Clean Architecture: config / handler / service / repository / model / worker
- [x] Thin `cmd/main.go` + `internal/app` DI
- [x] Proto gRPC: CreateCheckout, ConfirmPayment, GetSubscription, CanPurchaseMerch
- [x] Postgres: `payments`, `subscriptions`, `outbox` + migrator
- [x] Redis cache for active subscription
- [x] Outbox → NATS `payment.completed`
- [x] Boosty redirect URL via `BOOSTY_CHECKOUT_URL` (query: payment_id, plan, amount)
- [x] Webhook secret gate (`PAYMENT_WEBHOOK_SECRET`)
- [x] Gateway routes + compose (postgres/redis/payment) + Prometheus scrape `:9103`
- [x] Plans: present=200₽, future=300₽ (from ECONOMICS / ANTIFRAUD docs)

## Not in this pass

- Live Boosty API signature verification (manual/webhook stub)
- Shop auto-block of merch without subscription (expose `can-purchase-merch` for FE first)

Authors / History / Analytics: see `AUTHORS_HISTORY_ANALYTICS_STATUS.md` (done in same Stage 1 pass).

## Rebuild & push

Payment binary verified locally (`services/payment/payment-service`). Gateway rebuilt with payment routes (`services/gateway/gateway-service`).

```bash
# if you need to rebuild:
(cd services/payment && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o payment-service ./cmd/main.go)
docker build -f Dockerfile.payment.bin -t eastwesser/payment:latest .
(cd services/gateway && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go)
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

docker push eastwesser/payment:latest 2>&1 | tee confluence/history/2026-08/13.08.2026/push-payment.log
docker push eastwesser/gateway:latest 2>&1 | tee confluence/history/2026-08/13.08.2026/push-gateway.log
```

Then: `docker compose -f deployments/docker-compose.cluster.yml up -d postgres-payment redis-payment payment gateway gateway-2 gateway-3`