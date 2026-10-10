# Dream DevOps ideas (not v1.1 leftovers)

Parking lot for a **real multi-year / multi-host** SRE stack.  
These are **not** blockers for closing Event Horizon v1.1.0. Local + interview demos already have Prometheus + Grafana + Alertmanager + Jaeger + NATS/Redis/PG exporters.

See also: `DAILY_OPS.md` §3 gap table (source of the first rows).

---

## Observability “three pillars+”

| Idea | Why someday | Rough effort |
|------|-------------|--------------|
| **Loki** (or ELK) | Central logs instead of `docker logs` across hosts | M |
| **Tempo** (or Jaeger→Grafana) | Traces next to metrics in Explore | S–M |
| **node-exporter + cAdvisor** | Host CPU/RAM/disk + container RSS | S |
| **blackbox_exporter** | Probe `:8079/ready` + `/api/v1/...` from Prom | S |
| **Multi-DSN postgres_exporter** | One exporter (or many) per service DB | M |
| **Multi-Redis exporters** | shop/billing/LB Redis, not only auth | S |
| **Recording rules** | Pre-agg heavy PromQL for dashboards | S |
| **circuit_breaker gauge** | Alert when GW CB opens | S (app metric) |
| **SLO / multi-window burn** | Error-budget on-call, not raw thresholds | M |
| **Gateway metrics on host** | Avoid confusion with shop `:9095` | S (compose ports) |

---

## Delivery / platform

| Idea | Why someday |
|------|-------------|
| Hardened **first-boot** story on bare VMs (already improved; keep testing after reboot) | Ops pain |
| **k3s / Helm** as default path (chart exists; day-to-day is still compose) | Interview + “real” deploy |
| Image provenance: **tag + digest** pin, not only `:latest` | Drift (this was the v1 incident class) |
| CI gate: smoke `GET /api/v1/leaderboard` against rebuilt gateway | Prevent FE/GW skew |
| Alertmanager → Telegram always-on in staging | Human signal |
| Backup/restore runbook for PG volumes | Power-cut resilience |
| Chaos: kill NATS / one gateway, watch CB + balancer | Resilience demos |

---

## Explicitly out of scope for “0 leftovers” on v1.1

- Building Loki/Tempo/ELK now  
- Full prod on-call SLO program  
- Multi-region / Federation Prometheus  

If you open a future epic, copy a row from this file into a dated `history/…/TODO` and size it.
