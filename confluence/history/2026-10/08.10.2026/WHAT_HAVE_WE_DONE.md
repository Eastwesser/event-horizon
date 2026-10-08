# What have we done (v1.1.0 polish, 08.10.2026)

HEAD tip: `03adfbe` (and later) on `main`. Full ticklist: `TICKLIST_LAST_TODO_1.md`.

## Shipped in this wave

1. **Repo cleanup** — scripts → `scripts/{ops,metrics,loadtest}/`; Dockerfiles stay at root; FE game/shop stores co-located.
2. **Site/shop** — About, inventory ×qty, merch catalog, tickets Top-N, themes/skins SQL, history NATS dual-publish, mobile chrome.
3. **Games** — cosmic skins, Memonia map fix, 2048 drag, Balance everywhere, boost/GO polish, Tamagotchi/Gears names.
4. **Achievements** — amateur/pro/hero tiers × 8 games (`20261008120000_game_tier_achievements.sql`).
5. **Docs** — interview ports map, load runbook, security baseline, site/games feedback files.

## Denis runtime next

```bash
make deploy   # smoke
psql "$SHOP_DSN" -f scripts/cleanup-shop-content-v2.sql
psql "$SHOP_DSN" -f scripts/seed-shop-themes-skins.sql
psql "$HISTORY_DSN" -f scripts/seed-history-demo.sql
# profile migrations include new achievements
make test-k6  # after stack up → drop summary in LOAD_RESULTS/
```

## Still open (not code-blocked)

- Miro export / board sync
- Full highload numbers vs METRICS.md
- Avatar upload, 108 authors seed, C4 monetization
- MCP / Tetiva — last
