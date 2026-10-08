# k6 load scripts (Wave 4)

## Auth

```bash
# Or: make test-k6  (loads scripts/.env.seed.admin)
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD=...
```

## CORE (reports)

| Script | Purpose | Make |
|--------|---------|------|
| `browse.js` | shop + inventory browse (~20 VU) | `make test-k6` |
| `purchase.js` | purchase → cancel path | `make test-k6-purchase` |

## Legacy (not for CORE reports)

| Script | Purpose |
|--------|---------|
| `legacy/loadtest.js` | ~500 VU blast — auth will fail (bcrypt + rate limit) |
| `../../scripts/loadtest/loadtest_balancer.js` | submit blast @ `:8079` |
| `e2e-test.js` | 1 VU smoke |

## Run

```bash
make test-k6
make test-k6-purchase
# or:
k6 run deployments/k6/browse.js
```

Optional: `K6_VUS`, `K6_DURATION`, `EH_K6_ITEM_ID`.

## Indexes

Do **not** add indexes until after a measured run + `EXPLAIN`.
