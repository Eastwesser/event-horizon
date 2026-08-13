# Analytics service

OLAP metrics from **ClickHouse** (HTTP client, no native driver).

## Ports

| | Host | Container |
|---|---|---|
| gRPC | 50057 | 50057 |
| Metrics `/health` `/ready` | 9106 | 9106 |
| ClickHouse HTTP | 8123 | 8123 |
| ClickHouse native | 9000 | 9000 |

## RPCs

`RecordEvent`, `GetDAU`, `GetMAU`, `GetRetention`

## API (via gateway)

- `GET /api/analytics/dau?days=30`
- `GET /api/analytics/mau?days=30`
- `GET /api/analytics/retention?cohort_days_ago=7&window_days=7`

## Ingest

Same NATS subjects as History → `eventhorizon.analytics_events`

## Rebuild

```bash
cd ~/event_horizon/services/analytics
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o analytics-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker compose -f deployments/docker-compose.cluster.yml up -d clickhouse analytics
```
