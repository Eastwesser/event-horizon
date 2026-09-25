# Economy check — shop buyers after reference_id incident

**Cluster DB snapshot:** 2026-09-20

## Buyers (3 users, 23 completed purchases)

| User | Email | Role | Purchases | Ticket value | Spends recorded | Tickets now |
|------|-------|------|-----------|--------------|-----------------|-------------|
| `1502a3fa-…` | admin@eventhorizon.local | **admin** | 13 | 23 799 | 3 (300) | 999 708 (god seed) |
| `7fc8a659-…` | tuzer@example.com | user | 9 | 1 150 | **0** | 10 028 |
| `a11d3767-…` | test@example.com | user | 1 | 100 | 1 (100) | 900 |

## Interpretation

- **Admin** — seed god account (1M tickets). Free merch/skins here are **test inflation**, not real economy damage.
- **tuzer@example.com** — likely a **dev/test** user (example.com). Got ~1 150 tickets of free items; still holds 10 028 tickets. Not a child production player on this cluster.
- **test@example.com** — early smoke user; only purchase was correctly charged (July).

## Verdict for this cluster

All three buyers look **test/dev**, not production children. No clawback needed here.  
If the same shop image ran on a public env, re-run this query there before deciding clawback.

## Uncharged ticket value (approx.)

~24 649 of ~24 949 purchase value had no matching `shop_purchase` spend (admin + tuzer freebies).
