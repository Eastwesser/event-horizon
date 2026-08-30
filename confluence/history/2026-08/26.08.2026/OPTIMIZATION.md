# Memory / stack optimization (26–30.08.2026)

**Thin default (`make deploy`):** NATS JetStream + Docker Compose apps + ClickHouse/analytics + Prometheus/Grafana/Jaeger.  
**Heavy later (`make deploy-heavy`):** adds **Kafka broker** only (k3s stays `make deploy-k3s`). Kafka/k8s files are **not deleted**.

Naming: **`deploy-heavy`** fits better than `strong-deploy` (matches `deploy` / `stop-heavy`). `deploy-full` is an alias.

---

## Event bus

| Path | Bus |
|------|-----|
| Shop → fulfillment → notification | **NATS** subjects `purchase.paid` / `purchase.fulfilled` |
| Shop inventory / merch trail | NATS `shop.purchased` (outbox) |
| Payment | NATS `payment.completed` |
| Analytics ingest | NATS (same subjects as before) |
| Kafka | Optional dual-write when `KAFKA_BROKERS` set + `make deploy-heavy` |

---

## Make targets

```bash
make deploy          # thin: everything you need day-to-day (incl. fulfillment, notification, analytics)
make deploy-heavy    # + Kafka broker (stronger machine)
make deploy-full     # alias → deploy-heavy
make deploy-kafka    # start Kafka profile only
make stop-heavy      # stop Kafka/qdrant only (does NOT stop the 3 apps)
make deploy-k3s      # untouched; use on a cluster box
```

---

## Report: the 3 services (fixed for thin)

### Before

| Service | Why “failed” |
|---------|----------------|
| notification / fulfillment | Behind `profiles: [kafka]` + `stop-heavy` → Exited |
| analytics | Behind `profiles: [analytics]` + ClickHouse race / NATS durable dots |

### After (this change)

1. **Fulfillment & notification** consume/publish on **NATS**; Kafka code kept, used only if `KAFKA_BROKERS` non-empty.
2. **Shop** publishes `purchase.paid` to NATS (and Kafka when configured).
3. **nats-hub** EVENTS stream includes `purchase.paid`, `purchase.fulfilled`, `author.upserted`, …
4. **ClickHouse + analytics** back on thin deploy; analytics waits for healthy ClickHouse; durable names use dashes.
5. Kafka compose service still under `profiles: [kafka]` — leave it be.

### Rebuild / restart

```bash
bash scripts/rebuild-services.sh nats-hub shop fulfillment notification analytics
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d \
  nats-hub shop fulfillment notification clickhouse analytics
```

---

## Always-on vs heavy

**Thin (default)**

- Postgres · Redis · NATS · microservices · Gateway  
- fulfillment · notification · analytics · ClickHouse  
- Prometheus · Grafana · Jaeger  

**Heavy / later**

- Kafka broker (`make deploy-heavy`)  
- k3s (`make deploy-k3s`)  
- Qdrant (`--profile qdrant`)  
