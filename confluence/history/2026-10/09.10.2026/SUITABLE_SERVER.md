# Suitable server for Event Horizon (v1.1 local → small prod)

Honest capacity picture for **Denis’s ~8 GiB RAM VM** vs what the **product model** assumes, plus a right-sized Selectel/Hetzner-class box.

Sources: `confluence/history/2026-10/07.10.2026/LOAD_TESTS/METRICS.md`, `08.10.2026/LOAD_RESULTS/README.md`, `deployments/k6/*`, `scripts/loadtest/*`.

---

## 1. What the scripts actually measure

| Script | Kind | Default load | Path mix |
|--------|------|--------------|----------|
| `deployments/k6/browse.js` (**CORE**) | **Read-heavy** | ~20 VU × 30s | `GET /api/v1/shop/items` + `GET /api/v1/inventory/items` (+ login once) |
| `deployments/k6/purchase.js` (**CORE**) | **Write** | 1–5 VU | `POST purchase` + cancel (idempotency) |
| `deployments/k6/e2e-test.js` | Smoke | 1 VU | register/login/submit/LB |
| `scripts/loadtest/loadtest_balancer.js` | Legacy **write blast** | stages → 100 VU | repeated `POST /api/game/submit` (still legacy `/api` in file — prefer CORE) |
| `deployments/k6/legacy/loadtest.js` | Legacy blast | ~500 VU | auth collapses (bcrypt + rate limit) — **not** a fair “app capacity” number |

**Read/write today (realistic product mix):** planning target ≈ **2:1–3:1 reads:writes** (METRICS.md). CORE browse is almost pure read; purchase/submit are the write path. A fair combined run is ~70% browse + 30% submit/purchase VUs, not 500 logins.

---

## 2. What you can *really* expect on ~8 GiB RAM VM

Whole thin stack (NATS×3, many PG, Redis, gateways×3, apps, Prom/Grafana/Jaeger) already eats most of 8 GiB.

| Signal | Realistic on this VM | Notes |
|--------|----------------------|--------|
| Steady HTTP RPS (browse, authenticated) | **~20–80 RPS** green | Matches `browse.js` thresholds (p95 &lt;800 ms @ 20 VU) when not oversubscribed |
| Concurrent **authenticated** HTTP VUs | **~20–50** comfortable; **~100** flaky | Legacy 100 VU submit stages stress CPU + DB connections |
| Concurrent **logins** | **≪ 50** meaningful | bcrypt cost 12 + rate limit — 500 VU login = designed fail |
| WebSocket (LB) | hundreds, not thousands | One balancer process + Redis; not load-tested at 2.5k yet |
| Peak vs planning model (10k DAU → ~30–50 RPS) | **Model fits a bigger box**; VM proves *correctness*, not 10k DAU | |

So lacklustre legacy “500 VU” numbers are **expected** on this hardware — they measure auth ceiling + RAM, not “games submit is broken.”

---

## 3. Planning model (where you *want* to be)

From METRICS (10k DAU sketch):

- Average API ≈ **15–35 RPS**, peak ≈ **×2**  
- Read:write ≈ **2:1–3:1**  
- HTTP concurrent ≈ **thousands** only at full product scale; local demos never need that  
- WS planning: up to ~2.5k (product wish) — needs dedicated headroom  

---

## 4. Recommended servers

### A. Dev / CI / demos (what you have + slight upgrade)

| | Spec |
|---|------|
| **vCPU** | 4 |
| **RAM** | **16 GiB** (8 is tight with full compose) |
| **Disk** | 80–160 GiB SSD |
| **Net** | 100–200 Mbit |
| **Role** | `make deploy` + CORE k6 (20–50 VU browse) |
| **Ballpark** | ~€10–20/mo (Hetzner CX32 / Selectel similar) |

### B. Suitable “small prod / interview load” (recommended next buy)

| | Spec |
|---|------|
| **vCPU** | **8** |
| **RAM** | **32 GiB** |
| **Disk** | **200–320 GiB** NVMe |
| **Net** | 1 Gbit |
| **Role** | Full stack + Prometheus/Grafana + CORE k6 **50–150 VU** browse + modest purchase/submit; room for NATS + PG without OOM |
| **Ballpark** | ~€30–50/mo (Hetzner CPX41 / Selectel 8c/32G class) |

This is the **sweet spot** for Event Horizon as a highload-*teaching* project: you can show real RPS/p95 without pretending to be 100k MAU on one laptop.

### C. Closer to METRICS “5 servers / ~$430” fantasy

Split later (not day-one):

1. Edge: balancer + gateway×N + static FE  
2. Data: PG (or managed) + Redis  
3. Bus: NATS cluster  
4. Workers: fulfillment/notification/analytics + ClickHouse  
5. Observability: Prom/Grafana/Jaeger  

Until then, **one 8c/32G** is enough to grow into.

---

## 5. How to re-measure tomorrow (tired-proof)

```bash
# Read capacity
K6_VUS=20 K6_DURATION=60s make test-k6

# Stronger read (only on ≥16 GiB)
K6_VUS=50 K6_DURATION=60s make test-k6

# Write / purchase (keep VUs low — inventory/items)
K6_VUS=5 make test-k6-purchase

# Optional write mix (fix /api/v1 in script first if still legacy):
# BASE_URL=http://localhost:8079 k6 run scripts/loadtest/loadtest_balancer.js
```

Drop summaries under `confluence/history/2026-10/08.10.2026/LOAD_RESULTS/` as `browse-YYYYMMDD.txt`.

---

## 6. Bottom line

- **Now (8 GiB):** prove the system; CORE ~20 VU browse is the honest number.  
- **Suitable next server:** **8 vCPU / 32 GiB / NVMe** — matches a highload portfolio project without enterprise spend.  
- **Dream SRE** (Loki, multi-exporter, SLOs) → `DREAM_DEVOPS_IDEAS.md`, not this shopping list.
