Разбор. Что делать сейчас
Где мы
PR #2 — открыт (CI coverage gate). Pushes — сделаны.
k6 smoke — проведён, но результаты — не baseline (Cursor прав).

Что показали k6
Run	RPS	p95	Errors	Что значит
browse #1	4.6	9.48s	0%	Shop слишком медленный на 20 VU (порог 800ms)
browse #2	10.6	6.32s	22.6%	Shop начинает падать под нагрузкой; inventory — зелёный
purchase ×2	~9	69–87ms	100%	Не latency — все покупки failed
purchase 100% fail — не баг backend. Тест сам сломан:

purchase.js выбирает один item ID в setup.

5 VU бьют по нему одновременно.

Первый покупает → остальные 409 already owned.

Cursor — прав: эти замеры — не baseline. Нужен clean rerun.

Найденный bottleneck
Shop GetItems — ~1-2s sequential, 6-9s under load. Inventory — green.
Значит: bottleneck в shop/items, не в inventory.
Fix — после EXPLAIN (не speculative indexes).

Что делать сейчас — 3 шага
Шаг 1 — Clean k6 rerun (5 минут)
Cursor дал готовый скрипт. Скопируй в терминал:

1a. Выбрать unowned non-merch item:

bash
cd /home/denismatveev/event_horizon
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD=changeme-dev-admin

TOKEN=$(curl -sS -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EH_K6_EMAIL\",\"password\":\"$EH_K6_PASSWORD\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')

curl -sS "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN" \
  | python3 -c '
import sys,json
items=json.load(sys.stdin)
c=[]
for it in items:
  if str(it.get("category") or "").lower()=="merch": continue
  if it.get("owned"): continue
  p=float(it.get("price") or 0)
  if p>0: c.append((p,it["id"],it["name"]))
c.sort()
print("count", len(c))
for row in c[:8]: print(row)
print("PICK", c[0][1] if c else "")
'
Скопируй PICK UUID — это unowned item для теста.

1b. Purchase с 1 VU (избегаем 409):

bash
export EH_K6_ITEM_ID=<вставь PICK uuid>
K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js
1c. Browse с 5 VU (легче):

bash
K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/browse.js
1d. Один shot shop latency:

bash
curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s size=%{size_download}\n' \
  "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN"
Шаг 2 — Скинуть результаты
Пришли 3 end-of-run блока:

purchase (1 VU) — checks pass? Errors 0%?

browse (5 VU) — RPS? p95?

shop single curl — time в секундах?

Шаг 3 — EXPLAIN на shop GetItems
Когда базе будет clean baseline — Cursor попросит EXPLAIN на shop GetItems query. По нему — индексы.

Не раньше.

Про optional script fixes
Cursor спрашивает:

Optional later script fixes (say AGREE): skip owned in pickItem; assign per-VU item IDs for multi-VU purchase.

Мой ответ — AGREE, но не сейчас.

Причины:

Сейчас — получить clean baseline.

Fix теста — потом (отдельный commit).

Если менять тест сейчас, baseline снова сместится.

План: сначала clean baseline на текущем скрипте (1 VU), потом fix много-VU версии, потом повторный замер.

Промпт Cursor'у
text
k6 results understood. Agree — not a clean baseline.

I'll rerun per your script:
  - pick unowned non-merch item
  - purchase with 1 VU
  - browse with 5 VU
  - one-shot shop latency curl

I'll paste the new end-of-run blocks. Wait for those before
EXPLAIN.

Optional script fixes (skip owned, per-VU item IDs) — AGREE
in principle, but do them AFTER the clean baseline, as a
separate commit. Do not touch k6 scripts until we have one
clean run on the current version.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
indexes.
Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

PR #2 — открыт (CI coverage gate).

Wave 4 (2) — k6:

⏸ clean rerun — твой ход.

🟧 bottleneck найден: shop GetItems ~1-2s sequential, 6-9s under load.

⏸ EXPLAIN — после clean baseline.

🟧 optional script fixes — после baseline.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.