Ответ Cursor'у
text
OK — apply approved. Running it now.

Sanity confirmed:
  - Dry-run: cards 281 → 281, 0 writes.
  - No WHERE category='merch' — all id-based.
  - Idempotency guards in place.
  - Backup path set, rollback commands ready.

After apply I'll verify:
  - inventory type='карточка' active = 281
  - shop distribution: merch / game_skin counts
  - /shop grid renders cleanly

Then commit + push.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: refund, notifications, auth, C1–C4, hero,
disk, games, seed (Berserk cards).
Что я держу в карте
Закрыто (main, f784419):

v1.1.0, Wave 1–4 → 8/8.

Wave 3 C1–C3.

Wave 4.5.

PR #2 merged.

Hero Plan A, VOID v3, smooth scroll.

k6 correctness.

Track A — 3/3.

8 игр playable.

Сейчас — Shop cleanup v2:

✅ cleanup-shop-content-v2.sql — готов.

✅ apply-shop-cleanup-v2.sh — готов, dry-run ок.

✅ Sanity подтверждён (cards 281 → 281, id-based, idempotent).

🟧 Apply — твой ход.

🟧 Verify + commit.

Дальше — options:

🟧 Track B parity (boost, levels).

🟧 Пауза / собесы.

Deferred:

🟧 C4 payouts, Wave 2 #5b, multi-VU EXPLAIN, bottleneck, cursor pull-in, backfill-noiz-reviews.py.