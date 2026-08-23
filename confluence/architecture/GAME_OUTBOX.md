# Game service — Outbox

## Status (23.08.2026): **implemented**

| Piece | Location |
|-------|----------|
| Migration | `services/game/migrations/20260823120000_add_outbox.sql` |
| Repo | `EnqueueOutbox`, `SaveHighscoreAndEnqueueOutbox` (+ kept `SaveHighscore`) |
| Worker | `services/game/internal/worker/outbox_worker.go` |
| DI | `initOutbox` starts worker |
| Service | Prefer outbox; **fallback** to legacy `js.Publish` if outbox insert fails |

Subject stays **`score.updated`** (Leaderboard / Billing unchanged).

---

## Flow

1. Validate score / rewards  
2. If new record: **one TX** → highscore + outbox row  
   Else: outbox row only  
3. Worker publishes to JetStream and marks `processed`  
4. If outbox fails → legacy `SaveHighscore` + direct Publish (logged)

---

## Deploy

Rebuild/push **game** image and recreate container (migration runs on start):

```bash
cd ~/event_horizon
(cd services/game && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o game-service ./cmd/main.go)
docker build -f Dockerfile.game.bin -t eastwesser/game:latest .
docker push eastwesser/game:latest
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d game
```

---

## Why it mattered

Direct publish after DB write could lose leaderboard/reward events on NATS blips. Outbox matches Shop/Billing/Inventory.
