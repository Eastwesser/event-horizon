# Daily ops — Event Horizon (local compose)

One-page runbook for a programmer keeping the stack honest.  
Companion: `make obs-check`, `make daily`, Grafana `http://localhost:3000`, Prometheus `http://localhost:9090`.

---

## 0. URLs (memorize)

| What | URL / port |
|------|------------|
| Public API (balancer) | http://localhost:8079 |
| OpenAPI / Swagger | http://localhost:8079/docs · `/openapi.yaml` |
| Frontend preview | http://localhost:4173 (`make fe-preview`) |
| Prometheus | http://localhost:9090 → **Status → Targets** |
| Grafana | http://localhost:3000 (`GRAFANA_ADMIN_*`, default `admin`/`admin`) |
| Alertmanager | http://localhost:9193 |
| Jaeger | http://localhost:16686 |
| NATS exporter | http://localhost:7777/metrics |
| PG exporter | http://localhost:9187/metrics |
| Redis exporter | http://localhost:9121/metrics |

Dashboard folder after fix: **Dashboards → Event Horizon → Event Horizon — Ops Overview**.

---

## 1. Morning checklist (~5 minutes)

```bash
cd ~/event_horizon

# Containers up?
make status
# or: make ps

# Health / ready on metrics ports + Prom targets + Grafana
make daily
# (alias: make obs-check && make test-smoke)

# Any restart storms?
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps -a \
  | grep -E 'Restarting|Exit|unhealthy' || echo "no unhealthy/exited rows"
```

Open Prometheus → **Status → Targets**: every job should be **UP** (gateway has 3 instances).  
Open Grafana overview: Gateway RPS / 5xx% / p95 look sane; no red `up` dips.

---

## 2. Daily command kit

### Lifecycle

```bash
make deploy          # thin stack (apps + Prom/Grafana/Jaeger)
make deploy-full     # + kafka profile if you need it
make down            # stop
make clean           # stop + wipe volumes (destructive)
make restart         # down + deploy
make logs            # follow all
# one service:
docker compose --env-file .env -f deployments/docker-compose.cluster.yml logs -f gateway balancer auth
```

### Rebuild after code change

```bash
# pick services you touched
bash scripts/rebuild-services.sh gateway leaderboard notification
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d gateway gateway-2 gateway-3 leaderboard notification balancer
```

### Migrations / seeds

```bash
make migrate-all
make seed-admin          # needs scripts/.env.seed.admin
make seed-v110           # shop inventory + themes + history
make seed-lb-demo
make seed-card-artists
```

### Smoke / load

```bash
make test-smoke          # curl /health|/ready on host metrics ports
make test-k6             # CORE browse.js → :8079
make test-k6-purchase    # purchase path (1 VU default)
```

### Quick API sanity (manual)

```bash
# Balancer ready
curl -fsS http://localhost:8079/ready

# Login (use your seed admin)
curl -sS -X POST http://localhost:8079/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventhorizon.local","password":"YOUR_PASSWORD"}' | jq .

# Leaderboard read
curl -fsS 'http://localhost:8079/api/v1/leaderboard?game_id=hexagon&limit=5' | jq .
```

### Observability queries (Prom API)

```bash
# Targets down?
curl -sS 'http://localhost:9090/api/v1/query?query=up==0' | jq '.data.result'

# Gateway RPS
curl -sS 'http://localhost:9090/api/v1/query?query=sum(rate(gateway_requests_total[1m]))' \
  | jq -r '.data.result[0].value[1] // "n/a"'

# 5xx %
curl -sS 'http://localhost:9090/api/v1/query?query=100*sum(rate(gateway_requests_total{status=~"5.."}[5m]))/clamp_min(sum(rate(gateway_requests_total[5m])),1e-9)' \
  | jq -r '.data.result[0].value[1] // "0"'

# p95 latency (seconds)
curl -sS 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(gateway_request_duration_seconds_bucket[5m]))by(le))' \
  | jq -r '.data.result[0].value[1] // "n/a"'

# Active alerts
curl -sS http://localhost:9090/api/v1/alerts | jq '.data.alerts[]? | {alertname:.labels.alertname,state:.state,service:.labels.service}'
```

### Alertmanager Telegram smoke (only if `TELEGRAM_*` set in `.env`)

```bash
curl -XPOST http://localhost:9193/api/v2/alerts -H 'Content-Type: application/json' -d '[
  {
    "labels": {"alertname":"TestTelegram","severity":"warning","service":"ops"},
    "annotations": {"summary":"EH Telegram smoke"}
  }
]'
```

### Traces / logs when something smells

```bash
# Jaeger UI: search service=gateway, look at slow traces
xdg-open http://localhost:16686 2>/dev/null || true

# Structured logs from a hot path
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  logs --tail=200 gateway | jq -R 'fromjson? // .' 2>/dev/null || \
docker compose --env-file .env -f deployments/docker-compose.cluster.yml logs --tail=200 gateway
```

### Disk / Docker hygiene (weekly-ish)

```bash
docker system df
# careful: prune unused images/volumes only when you mean it
# docker system prune -f
```

---

## 3. Observability inventory — do we have enough?

### What you already have (thin deploy)

| Piece | Status | Notes |
|-------|--------|--------|
| Prometheus scrape of all app `/metrics` | ✅ | auth…analytics + gateway×3 + balancer |
| NATS / Redis / Postgres exporters | ✅ | scrapes in `prometheus.yml` |
| Alertmanager | ✅ | Telegram if `TELEGRAM_*`; else null receiver |
| Grafana datasource + dashboard provision | ✅ **fixed 09.10** | was mounting JSON as provider dir (broken) |
| Jaeger (OTLP) | ✅ | services get `JAEGER_ENDPOINT` |
| App `/health` + `/ready` | ✅ | compose healthchecks + `make test-smoke` |
| Business metrics | ✅ | `gateway_*`, `game_submits_total`, `orders_total`, `balancer_*` |
| SLI alerts (5xx, p95, downs) | ✅ **expanded 09.10** | see `deployments/prometheus/alerts.yml` |

### Gaps (honest — not blockers for local v1.1)

| Gap | Why it matters | When to care |
|-----|----------------|--------------|
| **No Loki / log aggregation** | Logs stay in `docker logs` | Multi-host / interview “ELK?” — structured slog is enough locally |
| **No Tempo** (Jaeger only) | Fine for demo; Tempo is Grafana-native | If you want traces inside Grafana Explore |
| **No node-exporter / cAdvisor** | No host CPU/disk/container RSS panels | Load tests on a real VM |
| **No blackbox_exporter** | External URL probe of `:8079/ready` from Prom | Prod-style uptime |
| **Postgres exporter → one DSN** | Only auth-ish `event-horizon-postgres`, not every service DB | Per-DB saturation / connection storms |
| **Redis exporter → one Redis** | Auth/session Redis; shop/billing Redis not scraped | Cache-specific incidents |
| **Gateway metrics not on host ports** | Prom scrapes Docker DNS; host `curl :9095` is **shop**, not gateway | Don’t confuse host maps with scrape targets |
| **nats-hub host `:9097`** | Collides with gateway-3 *internal* metrics port only on host map | Host curl `:9097` ≠ gateway-3 |
| **No recording rules** | Heavy PromQL re-eval on every panel | Large cardinality later |
| **No circuit_breaker Prometheus metric** | Old Grafana rule was dead; removed | Wire gauge if you want CB alerts |
| **No SLO/error-budget burn alerts** | Multi-window burn is prod on-call | After you pick SLOs |

**Verdict:** for local/dev + interview demo you have the **full metrics triangle** (Prometheus + Grafana + Alertmanager) plus **traces (Jaeger)** and **infra exporters**. You do **not** have a full “three pillars + host saturation + multi-DB” production SRE suite — that is the gap table above, not a missing Grafana install.

---

## 4. After pulling observability changes

```bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d prometheus alertmanager grafana
# if dashboard missing: recreate grafana (volume keeps admin prefs)
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate grafana
make obs-check
```

---

## 5. Failure → first look

| Symptom | Look at |
|---------|---------|
| Site 502 / balancer | `up{job="balancer"}`, gateway×3 `up`, `make logs` balancer/gateway |
| Login fails | auth `/ready`, Redis exporter, auth logs |
| Scores not on LB | game + leaderboard + NATS consumer pending panel |
| Shop purchase stuck | shop + billing + inventory + `orders_total` / fulfillment |
| Grafana empty panels | Prom Targets UP? metric name mismatch? traffic yet? |
| Alerts silent | Alertmanager UI `:9193`, `TELEGRAM_*` set?, Prom Alerts page |
