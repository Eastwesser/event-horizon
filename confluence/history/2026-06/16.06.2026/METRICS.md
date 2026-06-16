# Auth (9091)
curl -s http://localhost:9091/metrics | head -10

# Game (9092)
curl -s http://localhost:9092/metrics | head -10

# Billing (9093)
curl -s http://localhost:9093/metrics | head -10

# Leaderboard (9094)
curl -s http://localhost:9094/metrics | head -10

# Gateway (9095)
curl -s http://localhost:9095/metrics | head -10


{
  "title": "EventHorizon Business Metrics",
  "tags": ["eventhorizon", "business"],
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "panels": [
    {
      "title": "RPS by endpoint",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(gateway_requests_total[$__rate_interval])) by (path)",
          "legendFormat": "{{path}}"
        }
      ]
    },
    {
      "title": "Game submissions",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(game_submits_total[$__rate_interval])) by (game_id, status)",
          "legendFormat": "{{game_id}} - {{status}}"
        }
      ]
    },
    {
      "title": "Score distribution (P95)",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(game_score_histogram_bucket[$__rate_interval])) by (le, game_id))",
          "legendFormat": "{{game_id}}"
        }
      ]
    },
    {
      "title": "P95 latency",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum(rate(gateway_request_duration_seconds_bucket[$__rate_interval])) by (le, method, path))",
          "legendFormat": "{{method}} {{path}}"
        }
      ]
    }
  ]
}