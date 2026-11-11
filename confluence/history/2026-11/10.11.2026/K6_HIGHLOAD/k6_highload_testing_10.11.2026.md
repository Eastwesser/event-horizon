# K6 Highload — 10.11.2026

Цель: измерить потолок текущей конфигурации EH (8 GiB all-in-one).
Лестница: P=10 → P=50 → P=200 → P=1000 → (опц.) P=10000.

## Сводка

| Профиль | VU | Длит. | RPS | p95 (ms) | p99 (ms) | Failed % | Где упало |
|---------|----|-------|-----|----------|----------|----------|-----------|
| P=10    |    |       |     |          |          |          |           |
| P=50    |    |       |     |          |          |          |           |
| P=200   |    |       |     |          |          |          |           |
| P=1000  |    |       |     |          |          |          |           |

## По профилям

### P=10 (санити)
- Команда:
- Вывод: (см. terminal_outputs_10.11.2026.md#p10)
- Наблюдения:

### P=50
...

## Выводы

- Реалистичный потолок на 8 GiB: **N VU**
- Первый сервис, который деградирует: **...**
- Что держится дольше всего: **...**
- Рекомендация на v1.1.3.7:


Как это запускать (поэтапно)
Сначала — окружение (один раз в шелле, из корня репо):

bash
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD='changeme-dev-admin'   # или твой SEED_ADMIN_PASSWORD
Этап 1 — «душ» (P=50, ~2 мин):

bash
PROFILE=50 K6_DURATION=60s K6_RAMP=30s \
  k6 run deployments/k6/highload/10k-probe.js
Смотрим: p95 держится < 800ms, failed < 2%. Если да — идём дальше.

Этап 2 — «нагрузка» (P=200, ~3 мин):

bash
PROFILE=200 K6_DURATION=90s K6_RAMP=30s \
  k6 run deployments/k6/highload/10k-probe.js
Смотрим: где начинает расти p95, что первым краснеет (browse / submit / purchase). В этот момент обязательно второй терминал смотрит:

bash
docker stats --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'
и Grafana (http://localhost:3000, job=gateway, billing, shop).

Этап 3 — «стресс» (P=1000):

bash
PROFILE=1000 K6_DURATION=60s K6_RAMP=30s \
  k6 run deployments/k6/highload/10k-probe.js
Тут уже ждём деградацию. Фиксируем: первый сервис, который упал (по gateway_requests_total{status=~"5.."} и docker stats). Это и есть «где предел».

Этап 4 — «10k» (опционально, с готовностью ронять):

bash
PROFILE=10000 K6_DURATION=30s K6_RAMP=60s \
  k6 run deployments/k6/highload/10k-probe.js
Честно предупреждаю: на 8 GiB это уронит стек. Но это и есть гипотеза — «где именно». Если решишь — снимаем дампы docker stats и nats stream info каждые 10 секунд в фоне, чтобы был постмортем.



==========================

### P=50 — 10.11.2026, 2m05s

- Команда: `env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s k6 run deployments/k6/highload/10k-probe.js`
- Requests: 8366, Failed: 10.40% (threshold 10% crossed)
- p50/p95/p99 = 0.0 ms — **баг handleSummary** (ключи `p(95)`, а не `p95`), чинится в v1.1.3.7
- Реальные:
  - browse   p95 = 2999.0 ms
  - submit   p95 = 1775.0 ms
  - purchase p95 = 4150.7 ms
  - max             = 8018.2 ms
- Наблюдения:
  - Все VU ходят одним admin-токеном → **429 от rate limiter** (общий ~100/s) — архитектурный шум, не потолок системы.
  - purchase 4.1с — сага + Billing + NATS; ожидаемо, но хвост длинный.
  - max 8с — вероятно, один purchase-цикл упёрся в retry/таймаут.
- Гипотеза: при P=50 упираемся в **rate limiter на одном токене**, а не в БД/Redis. Нужен пул токенов, чтобы измерить систему.


### P=50 — v1 (старый handleSummary, 10.11.2026)

- Команда: `env -u K6_VUS ... EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s k6 run .../10k-probe.js`
- Requests: 5453, Failed: 12.98% (thresholds p95 + failed crossed)
- p50/p99 = 0.0 ms — **баг summary** (ключи p(50)/p(99) в k6 2.x), чинится
- p95 = 1738.2 ms, max = 35745.3 ms (**35 секунд!**)
- Сценарии:
  - browse   p95 = 5896.0 ms, errors = 8.99%  → 6 секунд на каталог, Nginx/Gateway отдаёт non-JSON (invalid char 'N' / 'B')
  - submit   p95 = 5328.7 ms, errors = 3.58%  → 5 секунд на submit
  - purchase p95 = 6408.8 ms, errors = **100.00%** → полный отказ, при 0% 409/429
- status_409 = 0%, status_429 = 0% → **не лимитер**, инфраструктура
- Наблюдения:
  - `invalid character 'N'` / `'B'` = Nginx/Gateway отдаёт **текстовый 502/503** вместо JSON.
  - purchase 100% fail при 0% 409/429 = сага/Shop/Billing **не отвечает**.
  - max 35s = таймаут, не «медленно».
- Гипотеза: под P=50 **первым падает entrypoint (Nginx/Gateway)**, а не БД. Нужен `docker stats` + логи Nginx/Gateway.

### P=50 — v2 (safeJSON, 10.11.2026)

- Фикс: `safeJSON` на всех ручках, `status_5xx` отдельно, `bodyPreview` в WARN.
- Команда: (та же)
- Вывод: (см. terminal_outputs_10.11.2026.md#p50-v2)
- Наблюдения:
  - WARN `shop/items non-200: status=... body=...` → **кто** отдаёт non-JSON.
  - `status_429` — если высокий, это rate limiter (нужен пул токенов).
  - `status_5xx` — если высокий, это Nginx/Gateway/сервис.

## Инфраструктура теста

- `seed-users.js` — регистрация N юзеров `k6_u_*@event-horizon.test`, пароль `k6_secret_2026`.
- `scripts/k6-seed-and-run.sh` — обёртка: seed → balance (SQL, 100к tickets+lamps) → json → run.
- `10k-probe.js` — с пулом токенов (`EH_K6_USERS_FILE`) или single-token (`EH_K6_EMAIL`).
- Баланс начисляется **SQL**, не через игру — для purchase-сценария нужен стартовый капитал.
- `raisa.matveeva` не затрагивается. Чистка: `bash scripts/k6-seed-and-run.sh clean`.


Как запускать (теперь — через env-строку, без путей)
bash
EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  k6 run deployments/k6/highload/10k-probe.js
Никаких open, никаких ../../, никаких $(pwd). cat читает файл, env передаёт строку, k6 парсит в SharedArray.

Что поменялось в файле
usersPool — теперь три ветки: EH_K6_USERS_JSON (env-строка, рекомендуется) → EH_K6_USERS_FILE (legacy, open) → null (single-token). Env-строка не зависит от cwd.

sanitizeUsers — вынесен в функцию, применяется в обеих ветках. Убирает \r\n, обрезает пробелы.

setup() — в ошибке теперь пишет Set EH_K6_USERS_JSON вместо EH_K6_USERS_FILE.

handleSummary — без изменений, всё те же p(50)/p(95)/p(99).

### P=50 — v3 (пул 10 юзеров, env-JSON, 10.11.2026) ✅ ВАЛИДНЫЙ ЗАМЕР

- Команда:
  EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  k6 run deployments/k6/highload/10k-probe.js
- Pool: 10/10 логинов, 10 unowned items
- Requests: 8535, Failed: **0.36%** (было 76.43% в single-token)
- p95: 824.7 ms, max: 4183.4 ms (p50/p99 = 0.0 — баг summary, чинится)
- Сценарии:
  - browse   p95=2481.1 ms, errors=0.05%  ✅
  - submit   p95=859.0 ms,  errors=0.00%  ✅
  - purchase p95=3034.0 ms, errors=16.04% ⚠️ (баг скрипта, не системы)
- status_409=0.35%, status_429=**0.00%** ✅, status_5xx=0.01%
- Наблюдения:
  - **Rate limiter побеждён пулом** (10 юзеров вместо 1 admin). 429 = 0%.
  - **browse p95=2.5s** — честный read path, надо разложить по ручкам.
  - **purchase 16%** — баг скрипта: (а) cancel2 check слишком строгий (`already_refunded=false` — валидный ответ), (б) items кончаются → 409 `item already owned`.
  - **inventory CB открылся 1 раз** (`503 circuit=inventory`) — первый сигнал деградации. Скорее всего, `inventory/items` с фильтрами тяжёлый.
- Вывод: **P=50 — потолок CORE на 8 GiB.** Дальше P=200 с пулом 100–200 юзеров.

### Инфраструктура: грабли (10.11.2026)

1. `psql -t -A` добавляет `\r` → `jq` сохраняет → k6 `JSON.parse` падает.
   Фикс: `tr -d '\r'`, `jq -Rnc`, валидация `jq -e .`.
2. Кросс-БД `eventhorizon.users` из billing не работает.
   Фикс: `user_id` из auth → INSERT в billing через SQL-файл.
3. `docker exec -it` в `while read` → `input device is not a TTY`.
   Фикс: `docker exec -i` + SQL-файл через `<`.
4. `all` без `clean` → накопление юзеров.
   Фикс: `clean` перед `seed`.
5. `open(path)` в k6 резолвит путь **относительно cwd**, не скрипта.
   Абсолютный → `[object Object]` → `invalid character 'o'`.
   Относительный → `stat ... no such file`.
   Фикс: **env-JSON** — `EH_K6_USERS_JSON="$(cat ...)"`.

### P=50 — v4 (чистый замер, 10.11.2026) ✅ BASELINE

- Команда:
  EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  k6 run deployments/k6/highload/10k-probe.js
- Pool: 10/10, unowned_items=10
- Requests: 9719, Failed: **0.67%** (только purchase_owned 409, реальных ошибок 0)
- p50: 314.6 ms, p95: 655.6 ms, p99: 1635.4 ms, max: 1635.4 ms
- Сценарии:
  - browse   p95=1894.0 ms, errors=0.00%  ✅
  - submit   p95=693.7 ms,  errors=0.00%  ✅
  - purchase p95=2264.2 ms, errors=0.00%  ✅
- browse breakdown (p95):
  - shop/items      : 690.0 ms
  - inventory/items : 632.8 ms
  - leaderboard     : 554.3 ms
  - profile         : 460.3 ms
- status_409=0.67%, status_429=**0.00%** ✅, status_5xx=**0.00%** ✅
- purchase_owned=27.43% (409 item already owned — не ошибка, но шум)
- Наблюдения:
  - **P=50 полностью зелёный.** Реальных ошибок 0.
  - **browse 1.9s — равномерная деградация**, нет одной «виноватой» ручки. Все 4 по 500–700 ms.
  - **purchase_owned 27%** — 10 items на 5 VU не хватает. Для P=200 нужно 50–100 items.
  - **CB Inventory не открывался** (было 1 раз в v3).
- Вывод: **P=50 — стабильный baseline.** Дальше P=200 с пулом 100 юзеров и 50+ items.

### Сводка по профилям

| Профиль | VU | RPS | p95 (ms) | p99 (ms) | Failed % | 429 % | 5xx % | Где упало |
|---------|----|-----|----------|----------|----------|-------|-------|-----------|
| P=50 v1 (single admin) | 50 | ~45 | 1738 | — | 12.98 | 0 | 1.74 | rate limiter + admin no balance |
| P=50 v2 (safeJSON) | 50 | ~160 | 393 | — | 76.43 | 73.62 | 1.74 | rate limiter |
| P=50 v3 (pool 10) | 50 | ~70 | 824 | — | 0.36 | 0 | 0.01 | purchase check |
| **P=50 v4 (чистый)** | 50 | ~78 | **655** | **1635** | **0.67** | **0** | **0** | **— (зелёный)** |
| P=200 | ? | ? | ? | ? | ? | ? | ? | ? |
| P=1000 | ? | ? | ? | ? | ? | ? | ? | ? |   


3. Вывод: P=200 — это за потолком 8 GiB
Три независимых сигнала:

Bulkhead full в Shop/Inventory — 140 browse-VU переполняют пулы.

Deadlock в Shop — 20 purchase-VU на 10 items.

502 Nginx — Gateway-реплики упали.

P=50 — зелёный. P=200 — красный. Значит, потолок 8 GiB all-in-one — где-то между 50 и 200. Скорее всего, ~80–100 VU.

4. Что это значит для «избытка коллов»
Ты хотел понять, где избыток коллов. P=200 дал ответ:

inventory/items — 6.9s (было 633ms). Это самая тяжёлая ручка. toPostgresFilter + Seq Scan + Cache-Aside. Под 140 VU кэш не спасает (Cache-Aside промахи), PG захлёбывается.

shop/items — 5.8s (было 690ms). Cache-Aside + PG. Та же история.

leaderboard — 3.6s (было 554ms). Redis Sorted Set должен быть быстрым, но Gateway → Leaderboard gRPC + ValidateToken добавляет. Под 140 VU — gRPC-хопы копятся.

profile — 3.6s (было 460ms). Агрегатор (Auth + Game + Billing). Под 140 VU — 3 gRPC-хопа × 140 = 420 одновременных вызовов.

Избыток коллов №1: ValidateToken на каждый запрос. 140 browse-VU × 4 ручки = 560 ValidateToken/итерация. Auth не тянет.
Избыток коллов №2: Profile агрегатор — 3 gRPC-хопа на запрос.
Избыток коллов №3: inventory/items — Seq Scan без индексов.

### P=200 — 10.11.2026 (ЗА ПОТОЛКОМ) 🔴

- Команда: EH_PROFILE=200, 10 токенов, 10 items
- Requests: 10923, Failed: 3.42%, p99: **35.9s**
- browse   p95=**17309 ms**  (P=50: 1894 ms)  ×9
- submit   p95=6861 ms       (P=50: 694 ms)   ×10
- purchase p95=**22025 ms**  (P=50: 2264 ms)  ×10
- status_429=0%, status_5xx=1.74%, purchase_owned=59.37%

**Три класса отказов:**
1. **Bulkhead full (503)** — Shop, Inventory. 140 browse-VU переполняют пулы.
2. **Deadlock (500)** — Shop `purchase`/`cancel`. 20 VU на 10 items + FOR UPDATE.
3. **502 Nginx** — Gateway-реплики упали под нагрузкой.

**Вывод:** P=200 — **за потолком 8 GiB all-in-one**. Потолок между 50 и 200, вероятно ~80–100 VU.

**Избыток коллов (по breakdown):**
- `inventory/items` 6.9s — Seq Scan + фильтры, нужны индексы.
- `shop/items` 5.8s — Cache-Aside промахи под нагрузкой.
- `leaderboard` 3.6s — gRPC-хоп + ValidateToken.
- `profile` 3.6s — 3 gRPC-хопа (Auth + Game + Billing).

**Следующий шаг:** P=100 (между 50 и 200) — найти точку перегиба.

