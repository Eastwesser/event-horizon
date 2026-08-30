# DB / login / Prometheus — diagnosis (30.08.2026)

## What was actually broken

### 1) Login / register (frontend 500)

Auth itself is **up**. Direct API works with a valid password (≥8 chars):

```bash
curl -X POST http://localhost:8079/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123","nickname":"You"}'
```

The SPA form said **«мин. 6 символов»** while Auth validates **8–128**. Short password → gRPC `InvalidArgument` → Gateway returned **500** (otel-wrapped error, `status.FromError` failed) → frontend showed «Ошибка соединения».

**Fix:** Register UI `minLength={8}` + show validation text; Gateway maps wrapped InvalidArgument → **400**.

### 2) Empty `POSTGRES_*` on `make deploy`

`.env` had JWT/Docker/Ansible but **no** `POSTGRES_USER` / `POSTGRES_PASSWORD`. After we switched compose to `${POSTGRES_USER}`, compose injected **blank** creds (`DB_USER=`).

Existing Postgres **volumes** were already initialized as `eventhorizon`, so most DBs stayed healthy. Go services fall back to `eventhorizon` when env is empty — Auth kept working.

**Fix:** appended `POSTGRES_*` / `GRAFANA_*` to local `.env`; compose now uses `${POSTGRES_USER:-eventhorizon}` so a missing `.env` key does not blank the password.

Do **not** change the password without `make clean` (volumes keep the original superuser).

### 3) Prometheus: game **down** (`lookup game: no such host`)

`deployments-game-1` had **Exited (1)** — started before Postgres was ready:

- `lookup event-horizon-postgres-game: no such host`
- `database system is starting up`
- `connection refused`

Restart after Postgres healthy: game is up (outbox worker + `:9092`). Compose now `depends_on: postgres-game (healthy)`.

### 4) Prometheus: notification **down** (HTTP 404 on `/metrics`)

Notification **was running**. Prometheus scrapes `/metrics`; the service only had `/health` and `/ready` → **404**. Kafka lookup `kafka` failed at boot (broker not ready) → noop consumer, still HTTP up.

**Fix:** `/metrics` via `promhttp`; recreate notification after Kafka is healthy.

---

## Recreate after image rebuild

```bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d \
  game gateway gateway-2 gateway-3 notification auth
```
