# Authors service

Public author profiles (linked to auth `user_id`). Inventory stocking stays on Inventory gRPC.

## Ports

| | Host | Container |
|---|---|---|
| gRPC | 50061 | 50061 |
| Metrics `/health` `/ready` | 9104 | 9104 |
| Postgres | 5468 | 5432 |
| Redis | 6387 | 6379 |

## API (via gateway)

- `PUT /api/authors/me` — upsert own profile (auth)
- `GET /api/authors/:user_id`
- `GET /api/authors?limit=&offset=`

## Events

Outbox → NATS `author.upserted`
