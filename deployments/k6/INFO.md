# k6 load scripts (Wave 4)

## Auth

All scripts except pure public probes need:

```bash
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=...
export EH_K6_PASSWORD=...
```

## Scripts

| Script | Purpose |
|--------|---------|
| `browse.js` | `GET /api/shop/items` + `GET /api/inventory/items` |
| `purchase.js` | purchase → cancel → cancel (per-VU unowned item; skips `owned`) |
| `leaderboard.js` | *(queued)* hot leaderboard read |
| `auth_burst.js` | *(queued)* login burst |
| `e2e-test.js` | legacy 1 VU smoke (register/login/submit/lb) |
| `loadtest.js` | **legacy / outdated paths** — prefer split scripts; do not use for Wave 4 reports |

## Run

```bash
cd deployments/k6
k6 run browse.js
k6 run purchase.js
```

Optional: `K6_VUS`, `K6_DURATION`, `EH_K6_ITEM_ID` (pin VU1 item only if unowned).
`purchase.js` auto-picks unowned non-merch ids (one per VU); extra VUs idle if the pool is short.

## Indexes

Do **not** add indexes until after a measured run + `EXPLAIN`. Report candidates first.
