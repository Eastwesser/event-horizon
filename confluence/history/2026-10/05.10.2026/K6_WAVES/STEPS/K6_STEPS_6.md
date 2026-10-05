Что сделано
k6 purchase.js fix:

Фильтрует owned === false, non-merch, price > 0.

Назначает по одному item на VU через __VU.

Игнорирует owned/missing EH_K6_ITEM_ID с warning.

Idle extra VU, если pool короткий.

Buy → cancel → cancel — idempotency проверка.

Локальный commit — не запушен. Commit message готов.

Запускай rerun
Копируй как есть (без < >):

bash
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD=changeme-dev-admin
unset EH_K6_ITEM_ID

K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js
K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/purchase.js
Важно: unset EH_K6_ITEM_ID — снять пин из прошлого теста. Иначе тот же owned item → 409.

Ожидания
Run	Pass bar
purchase 1 VU	checks ≫ 90%, errors ~0%
purchase 5 VU	checks ≫ 90% (нужно ≥5 unowned skins)
Если unowned skins < 5 — script idle extra VUs, warning в логе. Это ок, baseline всё равно получим (по фактическому количеству VUs).

Что скинуть
Два end-of-run блока (1 VU, 5 VU):

checks pass/fail.

http_req_duration p95.

http_req_failed rate.

thresholds ✓/✗.

Warning'и про pool short (если есть).

Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

PR #2 — открыт (CI coverage gate).

Wave 4 (2) — k6:

✅ browse 5 VU — p95 354ms, 0% errors (-86% от Wave 2).

✅ gateway fix — cancel чистый 200.

✅ purchase backend — правильный (409 для owned).

🟧 k6 purchase.js — fixed (local, не запушен).

⏸ rerun — твой ход.

⏸ commit k6 fix — после verify.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.