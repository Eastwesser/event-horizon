# Alertmanager → Telegram

Prometheus evaluates `deployments/prometheus/alerts.yml` and sends firing alerts to Alertmanager, which posts to Telegram when credentials are set.

## Enable

1. Put secrets in your env / `.env` (never commit them):

```bash
TELEGRAM_BOT_TOKEN=123456:ABC...
TELEGRAM_CHAT_ID=-1001234567890   # group/channel id (integer)
```

2. Start the observability stack (includes `alertmanager` in `docker-compose.cluster.yml`):

```bash
docker compose -f deployments/docker-compose.cluster.yml up -d prometheus alertmanager grafana
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

## Rules

| Alert | When |
|-------|------|
| GatewayDown / AuthDown / GameDown / BillingDown / LeaderboardDown | `up{job=…} == 0` for 1m |
| InventoryDown | inventory scrape down 2m |
| HighOrderRate | `rate(orders_total[1m]) > 10` for 1m |

ELK is **not** required — structured `slog` JSON (`LOG_FORMAT=json`) is enough for local dev.
