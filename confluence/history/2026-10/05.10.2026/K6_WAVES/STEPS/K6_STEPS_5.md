Cursor прав. Rebuild — обязателен. По шагам
Что происходит
Fix'ы в git:

7fbf19b — shop batch IsItemOwned.

13f4796 — gateway fix (no double-write).

НЕ в контейнерах:

shop + gateway — Up 6h, образ создан 6h назад → до этих коммитов.

cancel response сейчас: 503 + concatenated circuit:inventory + success — это старый throughBreaker path.

В новом source: inventoryCB.Execute (no response write).

Cursor прав: без rebuild — k6 бьёт по старому shop. Baseline не получим.

Что делать — точная последовательность
Шаг 1 — Проверить, что fix'ы на tip
bash
cd /home/denismatveev/event_horizon
git log --oneline -5
Ожидание:

text
13f4796 gateway fix
7fbf19b shop fix
...
Если этих коммитов нет — скажи — разберёмся, где потерялись.

Шаг 2 — Rebuild + recreate ВСЕ replicas
bash
bash scripts/rebuild-services.sh shop gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate shop gateway gateway-2 gateway-3
Важно: все три gateway (gateway, gateway-2, gateway-3) + shop. Balancer роутит на все — если обновить один, stale replicas останутся.

Проверь:

bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps | grep -E 'shop|gateway'
Ожидание: Created/Up — seconds–minutes, не hours.

Шаг 3 — Доказать, что новый binary живой
bash
TOKEN=$(curl -sS -X POST http://localhost:8079/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')

curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s\n' \
  http://localhost:8079/api/shop/items -H "Authorization: Bearer $TOKEN"
Pass: consistently < 0.3–0.5s (если batch задеплоен).

Fail: ~1–2s и больше → старый binary всё ещё работает (rebuild не доехал).

Проверка cancel (важно):

bash
# Найти owned item, cancel, посмотреть response
curl -sS -X POST "http://localhost:8079/api/shop/purchase/$OWNED_ID/cancel" \
  -H "Authorization: Bearer $TOKEN" | jq .
Expect: single 200 JSON. НЕ должно быть "circuit":"inventory" в response.

Шаг 4 — k6 rerun
Только после шага 3 (когда cancel чистый):

bash
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD=changeme-dev-admin

# PICK — смотри предыдущий промпт, скрипт для выбора unowned item

export EH_K6_ITEM_ID=PASTE_ACTUAL_UUID_HERE

K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js
K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/browse.js
Шаг 5 — Pass bar (без изменений)
shop curl — fast + stable (< 0.5s).

purchase checks — ≫ 90%.

browse p95 — ≪ 2.57s.

cancel response — no "circuit":"inventory".

До этого — k6 не post-fix baseline.

Ответ Cursor'у
text
Agreed — running containers are stale. Fixes are in git but
not in the running images. Rebuilding all replicas, not just
one.

I'll:
  1. git log --oneline -5 → confirm 13f4796 + 7fbf19b on tip.
  2. rebuild-services.sh shop gateway
  3. compose up -d --force-recreate shop gateway gateway-2
     gateway-3
  4. prove: cancel response has NO "circuit":"inventory";
     shop curl consistently < 0.5s.
  5. only then: k6 purchase 1 VU + browse 5 VU.

I'll paste:
  - git log -5
  - docker compose ps (shop/gateway Up times)
  - cancel response JSON
  - shop curl time
  - k6 end-of-run blocks

Pass bar unchanged. No further code until baseline is clean.

Standing rule: you don't push; you write commit messages.
DO NOT TOUCH: C1–C3, C4, hero, nav, games, seed, refund,
indexes, k6 scripts.
Что я держу в карте
Закрыто (в origin/main, 202a9da):

v1.0.9, Wave 1 — 6/6.

Wave 2 — 6/6 + 5/5 игр.

Wave 3 — C1 + C2 + C3.

Hero Plan A, VOID v3, smooth scroll.

Wave 4 (1) — все gated services ≥70%.

PR #2 — открыт (CI coverage gate).

Локально (fix'ы, не в контейнерах):

7fbf19b — shop batch IsItemOwned.

13f4796 — gateway fix (no double-write).

Wave 4 (2) — k6:

✅ bottleneck найден: N+1 + double-write.

✅ fix'ы готовы.

🔴 Контейнеры — stale (6h).

⏸ rebuild shop + gateway + recreate всех 3 реплик — твой ход.

⏸ clean k6 baseline — после rebuild.

🟧 script fixes (skip owned, per-VU) — после clean.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.