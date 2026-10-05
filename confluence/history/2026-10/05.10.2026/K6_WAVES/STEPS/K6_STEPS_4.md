Стоп. Сначала ребилд. Потом k6
Что Cursor говорит
Он прав. k6 Wave 3 не clean baseline, потому что:

Shop / gateway не пересобраны после fix'ов.

main на 2 коммита впереди origin.

Локальные образы — старые.

k6 бьёт по старому shop без batch fix'а.

EH_K6_ITEM_ID=<...> — bash error (угловые скобки).

Переменная не выставилась.

purchase снова по одному item → ~96% fail.

$TOKEN unset — shop curl вернул 401.

Не получили baseline.

Live probe сейчас:

buy → cancel — работает (значит, fix есть в коде).

shop list ~1–2s (значит, fix не задеплоен).

Диагноз: код fixed, образы старые. Нужен rebuild + recreate shop и gateway.

Что делать сейчас
Шаг 1 — Rebuild + recreate shop + gateway
Сначала проверь, что коммиты есть:

bash
git log --oneline -5
git log --oneline origin/main..HEAD
Ожидание: 2 коммита впереди (shop fix, gateway fix).

Потом rebuild:

bash
bash scripts/rebuild-services.sh shop gateway
docker compose --env-file .env -f deployments/docker-compose.cluster.yml \
  up -d --force-recreate shop gateway gateway-2 gateway-3 balancer
Проверь:

bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml ps | grep -E 'shop|gateway|balancer'
Все healthy?

Логи:

bash
docker compose --env-file .env -f deployments/docker-compose.cluster.yml logs --tail=20 shop | grep -i start
Момент старта — должен быть свежий (несколько секунд назад).

Шаг 2 — k6 rerun (правильная последовательность)
Копируй как есть, без < >:

bash
# 1. Выбор unowned item
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
c=[]
for it in json.load(sys.stdin):
  if str(it.get("category") or "").lower()=="merch" or it.get("owned"): continue
  p=float(it.get("price") or 0)
  if p>0: c.append((p,it["id"],it["name"]))
c.sort()
print("PICK", c[0][1] if c else "")
for row in c[:5]: print(row)
'

# 2. Если PICK пусто — отмени одну owned и повтори (см. Cursor'а).

# 3. Подставь UUID — БЕЗ угловых скобок:
export EH_K6_ITEM_ID=PASTE_ACTUAL_UUID_HERE

# 4. Один timed вызов — проверить, что batch fix задеплоен
curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s\n' \
  "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN"

# 5. k6 — purchase 1 VU
K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js

# 6. k6 — browse 5 VU
K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/browse.js
Pass bar (Cursor дал):

shop curl — ≪ 0.7s (если fix задеплоен).

purchase checks — ≫ 90% (если item unowned + 1 VU).

browse p95 — well under 2.57s, идеально <800ms.

Ответ Cursor'у
text
Understood — rebuild first, then k6.

I'll:
  1. Confirm local main has the shop + gateway fixes (2 ahead).
  2. rebuild-services.sh shop gateway + force-recreate.
  3. Run the corrected k6 script (no `< >`, $TOKEN set
     properly, unowned item, 1 VU purchase).
  4. Paste three blocks: shop curl time, purchase checks,
     browse p95.

No further code until those numbers are clean.

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

Локально (2 коммита):

shop fix — batch IsItemOwned.

gateway fix — no double-write response.

Wave 4 (2) — k6:

✅ bottleneck найден: N+1 + gateway double-write.

✅ fix'ы готовы.

⏸ rebuild shop + gateway — твой ход.

⏸ clean k6 baseline — после rebuild.

🟧 script fixes (skip owned, per-VU) — после clean.

Wave 4.5 (deferred):

🟧 Helm, Consul, OpenAPI.

Deferred:

🟧 C4, notifications, JWT role refresh, Wave 2 #5b, refund window, backfill-noiz-reviews.py, cursor pull-in.

