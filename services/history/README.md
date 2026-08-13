# History service

Append-only platform event trail with retention purge (~30 days).

## Ports

| | Host | Container |
|---|---|---|
| gRPC | 50062 | 50062 |
| Metrics `/health` `/ready` | 9105 | 9105 |
| Postgres | 5469 | 5432 |

## Ingest (NATS JetStream)

`payment.completed`, `author.upserted`, `shop.purchased`, `score.updated`, `user.registered`

## API (via gateway)

- `GET /api/history?event_type=&limit=&offset=` — own events (auth)

## Env

- `RETENTION_DAYS=30`
