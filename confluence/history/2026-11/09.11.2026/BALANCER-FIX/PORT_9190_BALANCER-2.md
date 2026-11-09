# 1. Все порты, занятые на хосте
sudo ss -tlnp | grep -E 'LISTEN' | awk '{print $4}' | sort -u

# 2. Все порты, проброшенные контейнерами
docker ps -a --format '{{.Names}}\t{{.Ports}}' | grep -v '^$' | sort

# 3. Только занятые на хосте (0.0.0.0:XXXX)
docker ps -a --format '{{.Ports}}' | grep -oE '0\.0\.0\.0:[0-9]+' | cut -d: -f2 | sort -n | uniq

[denismatveev@c0der event_horizon]$ docker compose --env-file .env -f deployments/docker-compose.cluster.yml restart prometheus
[+] restart 0/1
 ⠹ Container event-horizon-prometheus Restarting                                                                                                                                     2.2s
[denismatveev@c0der event_horizon]$ sleep 30
[denismatveev@c0der event_horizon]$ curl -s 'http://localhost:9090/api/v1/targets?state=active' | jq '.data.activeTargets[] | select(.labels.job=="balancer") | {instance: .labels.instance, health: .health, lastError: .lastError}'
{
  "instance": "balancer:9098",
  "health": "up",
  "lastError": ""
}
{
  "instance": "balancer-2:9098",
  "health": "up",
  "lastError": ""
}
[denismatveev@c0der event_horizon]$ 

# Port 9190 for balancer-2 metrics

**Date:** 09.11.2026
**Context:** balancer-2 health-check + Prometheus scrape for the second replica.

## Why 9190

- `9098` — занят `balancer-1` (host metrics).
- `9099` — занят `profile` (host metrics). Первый кандидат, отброшен.
- `9097` — занят `nats-hub`.
- `9091–9106` — заняты сервисами (auth, game, billing, ...).
- `9190` — свободен, рядом с `9193` (Alertmanager) → логическая группа observability.

## What changed

### deployments/docker-compose.cluster.yml

`balancer-2`:

```yaml
    ports:
      - "8080:8079"   # HTTP
      - "9190:9098"   # metrics (9099 занят Profile, 9098 занят balancer-1)
Внутри docker-сети balancer-2 слушает :9098 (как balancer-1). На хост — 9190.

deployments/prometheus/prometheus.yml
yaml
  - job_name: 'balancer'
    static_configs:
      - targets:
          - 'balancer:9098'
          - 'balancer-2:9098'
    metrics_path: '/metrics'
Prometheus ходит по внутренним именам balancer:9098 / balancer-2:9098, не по хостовым :9190.

Verify
bash
# Host metrics (balancer-2)
curl -s http://localhost:9190/metrics | grep balancer_backend_healthy

# Prometheus sees both
curl -s 'http://localhost:9090/api/v1/targets?state=active' \
  | jq '.data.activeTargets[] | select(.labels.job=="balancer") | {instance: .labels.instance, health: .health}'
Expected: 3 healthy backends per balancer; both targets up.

Occupied host ports (09.11.2026)
Snapshot from docker ps -a --format '{{.Ports}}' | grep -oE '0\.0\.0\.0:[0-9]+' | cut -d: -f2 | sort -n | uniq:

text
80 3000 4222 4223 4224 4317 5460 5461 5462 5463 5464 5465 5466 5467 5468 5469 5470
6060 6222 6223 6224 6379 6380 6381 6382 6383 6384 6385 6386 6387 6389
7777 8079 8081 8082 8083 8123 8222 8223 8224 9000
9090 9091 9092 9093 9094 9095 9096 9097 9098 9099
9101 9102 9103 9104 9105 9106 9121 9187 9193 16686
50051 50052 50053 50054 50055 50057 50058 50059 50060 50061 50062 50063
Свободные для будущих сервисов: 9190 (занят balancer-2), 9191+, 9200-9999, 19092 (Kafka heavy), etc. Перед добавлением нового порта — прогнать sudo ss -tlnp | grep <port>.

text

**Плюс** — создать/дополнить `confluence/history/2026-11/09.11.2026/BALANCER-FIX/README.md` (индекс файлов):
