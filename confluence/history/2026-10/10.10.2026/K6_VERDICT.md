# k6 CORE verdict — 10.10.2026

Host: Denis VM (~8 GiB RAM, full thin compose).  
Raw paste: [`K6_TESTING.md`](./K6_TESTING.md).  
Sizing context: [`../09.10.2026/SUITABLE_SERVER.md`](../09.10.2026/SUITABLE_SERVER.md).

---

## Verdict (one line)

**On this box, 20 concurrent browse VUs is the green CORE ceiling; 50 VU still returns 100% OK but breaks the p95&lt;800 ms SLO; write/purchase path is healthy at low VU.**

---

## Scorecard

| Run | Result | Why |
|-----|--------|-----|
| **Browse 20 VU / 60s** | **PASS** | All thresholds green; 0% fail; p95 **429 ms** |
| **Browse 50 VU / 60s** | **SOFT FAIL (latency only)** | 0% errors, 100% checks — but p95 **959 ms** &gt; 800 ms |
| **Purchase 5 VU / 30s** | **PASS** | p95 **196 ms**; cancel idempotent OK (1 VU idle — only 4 free items) |

---

## Numbers that matter

### Read (browse.js → shop + inventory)

| Metric | 20 VU | 50 VU |
|--------|------:|------:|
| HTTP RPS | **~50.4** | **~77.3** |
| Iterations/s | ~25.2 | ~38.6 |
| p50 latency | 209 ms | 429 ms |
| p95 latency | **429 ms ✓** | **959 ms ✗** |
| p99-ish (max) | 1.39 s | 2.12 s |
| Error / fail rate | **0%** | **0%** |
| Throughput payload | ~1.8 MB/s in | ~2.8 MB/s in |

Interpretation: the stack **does not break** at 50 VU — it **slows**. Bottleneck is saturation on the 8 GiB all-in-one VM (gateways + many PG/Redis + NATS + apps), not a functional bug.

### Write (purchase.js)

| Metric | 5 VU (4 active) |
|--------|----------------:|
| HTTP RPS | ~14.4 |
| p95 | **196 ms ✓** |
| Failures | **0%** |
| Note | Seed/catalog limited unowned SKUs — scale VUs only after more items |

Purchase path is fine for demos; not yet a write-capacity ceiling test.

---

## Read / write posture

- These CORE scripts are **read-dominated** (browse) vs **narrow write** (purchase).  
- Product target remains ~**2:1–3:1 read:write** ([METRICS](../07.10.2026/LOAD_TESTS/METRICS.md)).  
- Honest combined story for interviews: *“~50 RPS authenticated catalog reads @ 20 VU with p95&lt;500 ms on a single 8 GiB VM; latency SLO slips before error rate does.”*

---

## Capacity claim (what you can say)

| Claim | Safe? |
|-------|-------|
| CORE browse **20 VU** meets repo thresholds | **Yes** |
| Steady ~**50 RPS** read on this host | **Yes** (measured) |
| ~**75–80 RPS** still correct but p95 ~1 s | **Yes** — soft capacity |
| 10k DAU / 30–50 RPS *product model* on this VM | **No** — model needs ~8c/32G ([SUITABLE_SERVER](../09.10.2026/SUITABLE_SERVER.md)) |
| 500 VU legacy login blast | **Irrelevant** — bcrypt + rate limit by design |

---

## Next steps (optional, not tonight)

1. On **16–32 GiB**: re-run `K6_VUS=50` — expect p95 back under 800 ms if RAM/CPU were the limiter.  
2. Seed more shop SKUs → `K6_VUS=10+` purchase.  
3. Drop dated summaries under `08.10.2026/LOAD_RESULTS/` (`browse-20261010.txt`).  
4. Do **not** quote 500 VU auth fails as “system RPS.”

---

## Bottom line

Event Horizon on your current VM is a **correct, demo-ready highload-teaching stack** with a clear knee: **~20 VU green, ~50 VU correct-but-slow**. Upgrade RAM/CPU when you want the 50 VU run to pass the same SLO — not because the microservices are broken.
