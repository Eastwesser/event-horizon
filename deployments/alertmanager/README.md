# Alertmanager (Week 8 stub)

Optional Telegram alerts for high shop order rate.

## Enable

1. Set env on the alertmanager container:
   - `TELEGRAM_BOT_TOKEN`
   - `TELEGRAM_CHAT_ID`

2. Add to `prometheus/prometheus.yml`:

```yaml
rule_files:
  - /etc/prometheus/alerts.yml
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

3. Example rule (`deployments/prometheus/alerts.yml`):

```yaml
groups:
  - name: shop
    rules:
      - alert: HighOrderRate
        expr: rate(orders_total[1m]) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Shop order rate above 10/min"
```

ELK stack is **not** required for Event Horizon — structured `slog` JSON logs (`LOG_FORMAT=json`) are enough for local dev.
