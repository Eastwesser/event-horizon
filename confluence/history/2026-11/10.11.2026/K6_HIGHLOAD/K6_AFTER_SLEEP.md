[denismatveev@c0der event_horizon]$ EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
  env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  k6 run deployments/k6/highload/10k-probe.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/highload/10k-probe.js
        output: -

     scenarios: (100.00%) 3 scenarios, 50 max VUs, 2m10s max duration (incl. graceful stop):
              * browse: Up to 35 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: browseScenario, gracefulStop: 30s)
              * purchase: Up to 5 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: purchaseScenario, gracefulStop: 30s)
              * submit: Up to 10 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: submitScenario, gracefulStop: 30s)

INFO[0005] 10k-probe: PROFILE=50 pool=10/10 (from 10)    source=console
INFO[0005] 10k-probe: PROFILE=50 DURATION=60s RAMP=30s unowned_items=10 pool=10  source=console

==================================================
 10k-probe — PROFILE=50  DURATION=60s  RAMP=30s
==================================================
  Requests total : 9719
  Failed         : 0.67 %
  p50            : 314.6 ms   (avg fallback)
  p95            : 655.6 ms
  p99            : 1635.4 ms   (max fallback)
  max            : 1635.4 ms

  browse   p95   : 1894.0 ms   errors: 0.00%
  submit   p95   : 693.7 ms   errors: 0.00%
  purchase p95   : 2264.2 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 690.0 ms
    inventory/items : 632.8 ms
    leaderboard     : 554.3 ms
    profile         : 460.3 ms

  status_409     : 0.67 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 27.43 %   (409 item already owned — не ошибка)
==================================================

running (2m05.8s), 00/50 VUs, 3016 complete and 0 interrupted iterations
browse   ✓ [======================================] 00/35 VUs  2m0s
purchase ✓ [======================================] 0/5 VUs    2m0s
submit   ✓ [======================================] 00/10 VUs  2m0s
[denismatveev@c0der event_horizon]$ 

1. Что показал P=50 v4 (чистый)
text
Requests total : 9719
Failed         : 0.67 %
p50            : 314.6 ms   ← больше не 0.0
p95            : 655.6 ms
p99            : 1635.4 ms  ← больше не 0.0
max            : 1635.4 ms  ← совпадает с p99, значит хвост короткий

browse   p95   : 1894.0 ms   errors: 0.00%   ✅
submit   p95   : 693.7 ms    errors: 0.00%   ✅
purchase p95   : 2264.2 ms   errors: 0.00%   ✅ (было 16%)

browse breakdown (p95):
  shop/items      : 690.0 ms
  inventory/items : 632.8 ms
  leaderboard     : 554.3 ms
  profile         : 460.3 ms

status_409     : 0.67 %       ← это purchase_owned, не ошибки
status_429     : 0.00 %       ✅
status_5xx     : 0.00 %       ✅ (inventory CB больше не открывался)
purchase_owned : 27.43 %      ← 409 item already owned
2. Что это значит
P=50 — полностью зелёный. 0.67% failed, но это только purchase_owned (409). Реальных ошибок — 0. Rate limiter не мешает (0% 429). 5xx — 0. CB Inventory не открывался.

Browse breakdown — вот это самое интересное. browse p95 = 1894 ms — это сумма 4 запросов. По отдельности:

shop/items — 690 ms

inventory/items — 632 ms

leaderboard — 554 ms

profile — 460 ms

Сумма = 2336 ms, но browse_duration = 1894 ms (потому что p95 суммы ≠ сумме p95, но близко). Все 4 ручки примерно равны, ~500–700 ms каждая. Нет одной «виноватой» ручки — это равномерная деградация read path под P=50.

Что это значит: каждый запрос проходит через Gateway → Auth ValidateToken (gRPC) → сервис → PG/Redis. И каждый добавляет ~500–700 ms. Это не один медленный сервис, это общая задержка стека. Скорее всего:

Auth ValidateToken на каждый запрос (Redis lookup + gRPC хоп) — ~50–100 ms;

Gateway → сервис gRPC — ~50–100 ms;

Сервис → PG/Redis — ~100–300 ms;

Обратно через Gateway — ~50–100 ms.

Под 35 browse-VU × 4 ручки = 140 rps на read path. На 8 GiB all-in-one это уже близко к потолку, но не за ним.

Purchase_owned = 27.43% — это не ошибка, но это значит: 10 items на 5 VU не хватает для 60s-прогона. VU делает buy → cancel → buy, но cancel не всегда успевает (или item возвращается не мгновенно). 27% итераций получают 409. Это шум, но он говорит: для P=200 нужно больше items (минимум 50–100 unowned non-merch).

Purchase p95 = 2264 ms — сага (Billing SpendCurrency + Shop stock + outbox + NATS) под 5 VU. Это медленнее, чем browse, но ожидаемо (сага + синхронный Billing).