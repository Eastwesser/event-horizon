# Load results (v1.1.0)

## How to run (after `make deploy` / stack up)

```bash
# CORE suite (gateway :8080)
make test-k6
# or:
k6 run deployments/k6/loadtest.js
# lighter / alternate:
k6 run scripts/loadtest/loadtest.js

# Balancer-focused:
k6 run scripts/loadtest/loadtest_balancer.js
```

Env knobs (see script headers): `TARGET`, `BASE_URL`.

## Status 08.10.2026

- **Blocked on runtime:** gateway `/health` not up on this machine during polish wave — no numbers yet.
- After Denis smoke deploy, drop k6 summary JSON / screenshots here and tick §14 in `TICKLIST_LAST_TODO_1.md`.

## Targets (from `LOAD_TESTS/METRICS.md`)

| Metric | Target |
|--------|--------|
| p50 / p90 / p95 / p99 | &lt;50 / &lt;150 / &lt;200 / &lt;300 ms |
| Error / 5xx | near zero under CORE |
| Peak guidance | CPU/RAM/IO/net &lt;70% |
