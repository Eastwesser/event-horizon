# Site / shop feedback — v1.1.0 polish wave

## Done (code)

- About page from navbar brand; home blurbs (Builder / Gears / Tamagotchi); nick Modal (no `confirm`).
- Shop: card filters only on Cards; merch type chip; inventory grouped by `item_id` with ×qty; cancel copy «предмет»; themes icon ≠ skins (star vs palette).
- Create product: price in tickets; merch option; foil only for cards.
- Admin Top-N: tickets icon (not ₽); retention D0…D7 copy.
- History: dual NATS publish `user.registered` + `event.user.registered`; empty-state copy; `scripts/seed-history-demo.sql`.
- Cosmetics seed SQL: `scripts/seed-shop-themes-skins.sql` (cosmic rename + themes).
- Mobile: safe-area, shell insets, brand/logout collapse, game touch targets.
- F12: `/api/api/profile` fixed; LB `undefined` query spam fixed; shop inventory console spam removed.

## Apply on live DBs (Denis)

```bash
# shop
psql "$SHOP_DSN" -f scripts/cleanup-shop-content-v2.sql
psql "$SHOP_DSN" -f scripts/seed-shop-themes-skins.sql
# history (after rebuild gateway+history)
psql "$HISTORY_DSN" -f scripts/seed-history-demo.sql
```

## Open / product

- Avatar upload; LB seed nicknamed players; 108 authors seed.
- Author avatars from disk; revenue view (C4).
- Achievements ladder 3×/game (in progress).
- Full highload numbers after `make deploy` + `make test-k6`.

## F12 / bugs still watching

- Profile zeros until real ranked saves on deployed stack.
- Notifications «record beaten» deep-link not wired.
- Admin types count still reflects live DB until seed.
