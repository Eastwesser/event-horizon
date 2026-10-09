# Load posture note (closes ticklist §14 capture)

| Source | Meaning |
|--------|---------|
| `LOAD_TESTS/METRICS.md` + Miro panel | **Targets** (10k DAU model) |
| `LOAD_RESULTS/*` 08.10 | **Measured** local CORE |

| Metric | Target (prod model) | Local CORE (08.10) |
|--------|---------------------|--------------------|
| RPS | ~17–50 avg, peak ~35–100 | browse ~35 HTTP req/s @ ~20 VU |
| p95 latency | &lt;200 ms | ~705 ms (laptop+docker) |
| Errors | low | 0% on CORE browse/purchase |
| Concurrent HTTP/WS | 5k / 2.5k model | not load-tested to ceiling |
| Auth blast 500 VU | N/A | ~99% fail — bcrypt12 + login 5/s (by design) |

Grafana 03:00 watchlist (ops): Auth/GW p99, 5xx, free RAM, JetStream lag ≫1000.  
Selectel ~$430 / 5 VM — still the planning figure.  
Risk write-up: auth write path is the bottleneck; read path (shop/LB/submit) stayed green under CORE.
