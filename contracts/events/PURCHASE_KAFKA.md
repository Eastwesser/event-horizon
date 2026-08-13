# Kafka purchase-flow contracts (Week 5 / Kozirev Order+Assembly adapted to Event Horizon)

## Topics

| Topic | Producer | Consumer |
|-------|----------|----------|
| `purchase.paid` | shop | fulfillment, notification |
| `purchase.fulfilled` | fulfillment | shop (optional status), notification |

## PurchasePaid

```json
{
  "event_uuid": "uuid",
  "event_type": "PurchasePaid",
  "purchase_uuid": "uuid",
  "user_uuid": "uuid",
  "item_uuid": "uuid",
  "price": 100
}
```

## PurchaseFulfilled

```json
{
  "event_uuid": "uuid",
  "event_type": "PurchaseFulfilled",
  "purchase_uuid": "uuid",
  "user_uuid": "uuid",
  "item_uuid": "uuid"
}
```

Go types: `contracts/events/purchase.go`

Env: `KAFKA_BROKERS=kafka:9092`

Real Sarama client: `go get github.com/IBM/sarama@v1.45.2` then build with `-tags kafka`.
