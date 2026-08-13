# Week 8 status — Event Horizon (13.08.2026, shortened)

## Done

- [x] Business metrics in `platform/pkg/metrics/business.go`:
  - `orders_total`, `orders_revenue_total` — shop `PurchaseItem`
  - `assembly_duration_seconds` — fulfillment Kafka handler
- [x] Fulfillment `/metrics` endpoint (Prometheus scrape on `:9101`)
- [x] Prometheus alert rule stub: `deployments/prometheus/alerts.yml` (`HighOrderRate`)
- [x] Alertmanager config stub: `deployments/alertmanager/` + README

## Skipped (by design — shortened W8)

- Full ELK stack (use `LOG_FORMAT=json` + slog instead of Zap migration)
- Telegram webhook wiring in Alertmanager (documented only; needs tokens)
- Zap replacement across all services

## Verify

```bash
task w8-check
curl -s localhost:9095/metrics | grep orders_total   # shop
curl -s localhost:9101/metrics | grep assembly_duration_seconds
```

## Rebuild

```bash
(cd services/shop && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go)
docker build -f Dockerfile.shop.bin -t eastwesser/shop:latest .
(cd services/fulfillment && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags kafka -ldflags="-s -w" -o fulfillment-service ./cmd/main.go)
docker build -f Dockerfile.fulfillment.bin -t eastwesser/fulfillment:latest .
```

After a purchase flow, `orders_total` should increment; fulfillment delay should appear in `assembly_duration_seconds` histogram.
