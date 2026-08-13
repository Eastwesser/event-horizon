# Authors → History → Analytics status — Event Horizon (13.08.2026)

## Done (Stage 1 / voice-message)

### Authors (`:50061` / metrics `:9104`)
- [x] Clean Arch + DI + migrator (`authors`, `outbox`)
- [x] Redis cache, outbox → NATS `author.upserted`
- [x] Gateway: `PUT /api/authors/me`, `GET /api/authors`, `GET /api/authors/:user_id`
- [x] Compose: postgres-authors (5468), redis-authors (6387)

### History (`:50062` / metrics `:9105`)
- [x] Append-only `events` + retention worker (`RETENTION_DAYS=30`)
- [x] NATS ingest: payment / author / shop / score / user.registered
- [x] Gateway: `GET /api/history`
- [x] Compose: postgres-history (5469)

### Analytics (`:50057` / metrics `:9106`)
- [x] ClickHouse HTTP client (stdlib, no CH Go module)
- [x] Table `eventhorizon.analytics_events` + DAU / MAU / Retention RPCs
- [x] NATS ingest (same subjects as History)
- [x] Gateway: `/api/analytics/{dau,mau,retention}`
- [x] Compose: clickhouse (8123/9000) + analytics

## Next (after this push)

- RAG / MCP server (Stage 1 finish)
- OpenAPI paths for the new routes
- Optional: admin-only gate on analytics endpoints

## Rebuild & push

```bash
# binaries
(cd services/authors && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o authors-service ./cmd/main.go)
(cd services/history && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o history-service ./cmd/main.go)
(cd services/analytics && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o analytics-service ./cmd/main.go)
(cd services/gateway && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go)

# images
docker build -f Dockerfile.authors.bin -t eastwesser/authors:latest .
docker build -f Dockerfile.history.bin -t eastwesser/history:latest .
docker build -f Dockerfile.analytics.bin -t eastwesser/analytics:latest .
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

docker push eastwesser/authors:latest
docker push eastwesser/history:latest
docker push eastwesser/analytics:latest
docker push eastwesser/gateway:latest

docker compose -f deployments/docker-compose.cluster.yml up -d \
  postgres-authors redis-authors authors \
  postgres-history history \
  clickhouse analytics \
  gateway gateway-2 gateway-3
```
