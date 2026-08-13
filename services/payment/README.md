# Payment service

Real-money **subscriptions** (Boosty redirect). Does **not** touch in-game tickets — that stays in Billing.

## Ports

| | Host | Container |
|---|---|---|
| gRPC | 50058 | 50058 |
| Metrics `/health` `/ready` | 9103 | 9103 |
| Postgres | 5467 | 5432 |
| Redis | 6386 | 6379 |

## Plans

| Plan | RUB/month | Theme (docs) |
|---|---|---|
| `present` | 200 | Present / Zenless |
| `future` | 300 | Future / Sci-fi |

Active subscription unlocks merch redemption (`CanPurchaseMerch`).

## HTTP (via Gateway)

- `POST /api/payment/checkout` `{ "plan": "present" }` → Boosty URL + `payment_id`
- `GET /api/payment/subscription`
- `GET /api/payment/can-purchase-merch`
- `POST /api/payment/webhook` `{ "payment_id", "provider_ref", "webhook_secret?" }` → activates sub

## Events (NATS outbox)

- `payment.completed`

## Rebuild

```bash
cd ~/event_horizon/services/payment
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o payment-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.payment.bin -t eastwesser/payment:latest .
docker push eastwesser/payment:latest
```

Also rebuild **gateway** (new routes + `PAYMENT_ADDR`).
