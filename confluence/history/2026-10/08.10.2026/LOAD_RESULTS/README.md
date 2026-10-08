# Load results (v1.1.0)

## CORE (use these)

```bash
# Credentials from scripts/.env.seed.admin (SEED_ADMIN_*) or export EH_K6_*
make test-k6              # deployments/k6/browse.js @ :8079 (~20 VU)
make test-k6-purchase     # purchase.js (default 1 VU)
```

Manual:

```bash
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD='…'   # from seed-admin / .env.seed.admin
k6 run deployments/k6/browse.js
K6_VUS=1 k6 run deployments/k6/purchase.js
```

Expect (local): checks >90%, p95 &lt;800 ms on browse thresholds.

## Legacy (do not use for v1.1.0 reports)

| Script | Notes |
|--------|--------|
| `deployments/k6/legacy/loadtest.js` | ~500 VU blast — auth collapses (bcrypt + rate limit); reads may stay green |
| `scripts/loadtest/loadtest_balancer.js` | Retargeted to `:8079`; still legacy submit blast |

## Seeds (Docker, not host `psql` socket)

```bash
make seed-v110    # shop cleanup + themes/skins + history demo + migrate-profile
# or piecemeal:
make seed-shop
make seed-themes
make seed-history
make migrate-profile
```

## Findings 08.10.2026 (Denis run)

- `make test-k6` previously failed: missing `EH_K6_EMAIL`/`PASSWORD` — **fixed in Makefile**.
- 500 VU `loadtest.js`: ~99% login fails — capacity, not regression; shop/LB/submit green.
- `loadtest_balancer.js` was on `:8080` — retargeted to `:8079`.
- Host `psql "$SHOP_DSN"` failed (empty DSN → unix socket) — use `make seed-*` / `docker exec`.

Drop new CORE summaries as `browse-YYYYMMDD.txt` / `purchase-YYYYMMDD.txt` in this folder.
