# Alertmanager → Telegram

Prometheus evaluates `deployments/prometheus/alerts.yml` and sends firing alerts to Alertmanager, which posts to Telegram when credentials are set.

## Enable

1. Put secrets in your env / `.env` (never commit them):

```bash
TELEGRAM_BOT_TOKEN=123456:ABC...
TELEGRAM_CHAT_ID=-1001234567890   # group/channel id (integer)
```

2. Start the observability stack (included in thin `make deploy`):

```bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d prometheus alertmanager grafana
```

3. Without `TELEGRAM_*`, Alertmanager still runs with a **null** receiver (alerts are dropped locally — no crash loop).

## Smoke test

```bash
# Fire a test alert via Alertmanager API
curl -XPOST http://localhost:9193/api/v2/alerts -H 'Content-Type: application/json' -d '[
  {
    "labels": {"alertname":"TestTelegram","severity":"warning","service":"ops"},
    "annotations": {"summary":"EH Telegram smoke","description":"If you see this, Telegram works."}
  }
]'
```

UI: http://localhost:9193 (host port; container still listens on `:9093` for Prometheus).

Daily ops: `make obs-check` · full kit: `confluence/history/2026-10/09.10.2026/DAILY_OPS.md`.

## Rules (`deployments/prometheus/alerts.yml`)

| Alert | When |
|-------|------|
| GatewayDown / AuthDown / GameDown / BillingDown / LeaderboardDown / ShopDown / BalancerDown | `up{job=…} == 0` for 1m |
| InventoryDown / PaymentDown | scrape down 2m |
| NotificationDown / NATSExporterDown / PostgresExporterDown / RedisExporterDown | scrape down 2m (warning) |
| HighGatewayErrorRate | gateway 5xx ratio > 5% for 5m |
| HighGatewayLatencyP95 | gateway p95 > 2s for 5m |
| HighOrderRate | `rate(orders_total[1m]) > 10` for 1m |
| HighRedisMemory | redis used > 500MiB for 10m |

ELK is **not** required — structured `slog` JSON (`LOG_FORMAT=json`) is enough for local dev.
