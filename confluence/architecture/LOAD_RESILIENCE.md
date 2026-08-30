# Load resilience — lessons for Event Horizon

> Derived from a real peak-load incident (`confluence/history/2026-08/22.08.2026/news/SCAM WARNING.txt`).
> Goal: avoid the same failure modes under shop/payment/game spikes.

---

## Incident pattern (external case)

| Signal | Meaning |
|--------|---------|
| Client `TimeoutError` on hot user API | Client waited longer than timeout |
| Viral campaign spike | Traffic ≫ average; pools saturate |
| Same GET on every screen | No cache → amplified DB load |

---

## EH risk map

| Hot path | Today | Risk |
|----------|--------|------|
| whoami / user / profile | Auth Redis + Gateway auth | High if SPA polls hard |
| shop purchase / payment | CB + idempotent confirm | Medium |
| game submit | **Outbox for `score.updated`** | Lower after outbox |
| leaderboard GET | Redis-backed | Medium |
| analytics | ClickHouse, admin-only | Low for players |

---

## What EH already has

- Gateway **rate limiting** + **circuit breakers**
- Auth sessions in **Redis**
- Inventory / billing / shop / payment / **game** Outbox
- `/health` + `/ready`; Prometheus + Grafana
- Payment webhook idempotency
- Merch gate: probe 200 vs purchase 403
- Frontend axios **timeout 15s** (checkout **30s**)

---

## Checklist status (23.08.2026)

### Must for campaign / launch day

| # | Item | Status |
|---|------|--------|
| 1 | Cache read-heavy whoami/profile | **Done** — Gateway 5s response cache (`authReadCache`) |
| 2 | Keep rate limits ON | **Done** — Gateway `RateLimitMiddleware` |
| 3 | Client timeouts | **Done** — axios 15s / checkout 30s |
| 4 | Graceful degradation (`_partial`) | **Done** — profile falls back to cached whoami + `_partial: true` |
| 5 | Load test 5× peak | **Ready** — `make test-k6` (run when stack is up) |

### Should

| # | Item | Status |
|---|------|--------|
| 6 | Game Outbox `score.updated` | **Done** (code + migration; rebuild game image) |
| 7 | Alert p95 / circuit-open | **Done** — `deployments/grafana/provisioning/alerting/event-horizon.yml` |
| 8 | Scale Gateway replicas | **Done** in compose (gateway ×3 + balancer) |

### Later

| # | Item | Status |
|---|------|--------|
| 9 | Patroni Auth pilot | **Stubs ready** — `deployments/patroni/auth/` |
| 10 | Redis Cluster | Later |
| 11 | Read replicas | Later |

---

## Anti-patterns (do not)

- Raising client timeout to 60s as the only fix
- Disabling rate limits “for the event”
- Polling `/api/auth/user` every second
- Direct NATS publish after commit for money/rewards (use Outbox)
- Single shared Postgres for all domains at high QPS

---

## Campaign playbook

| When | Action |
|------|--------|
| T−14d | Load test top endpoints |
| T−7d | Verify Redis caches + rate limits |
| T−3d | Grafana alerts |
| T−1d | Image rollback dry-run |
| During | Watch CB, PG conns, NATS lag |
| After | Postmortem |

---

## Related

- [`API_ROUTES.md`](API_ROUTES.md)
- [`STATUS_CODES.md`](STATUS_CODES.md)
- [`PATRONI.md`](PATRONI.md)
- [`GAME_OUTBOX.md`](GAME_OUTBOX.md)
- [`BRANDING.md`](BRANDING.md)
