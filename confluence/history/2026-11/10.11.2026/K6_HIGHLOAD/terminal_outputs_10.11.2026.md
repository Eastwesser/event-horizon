CLEANING

env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
  k6 run deployments/k6/highload/10k-probe.js

[denismatveev@c0der event_horizon]$ env -u K6_VUS -u K6_DURATION -u K6_RAMP -u K6_SCENARIOS \
  EH_PROFILE=50 EH_DURATION=60s EH_RAMP=30s \
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

INFO[0004] 10k-probe: PROFILE=50 DURATION=60s RAMP=30s unowned_items=10  source=console

==================================================
 10k-probe — PROFILE=50  DURATION=60s  RAMP=30s
==================================================
  Requests total : 8366
  Failed         : 10.40 %
  p50            : 0.0 ms
  p95            : 0.0 ms
  p99            : 0.0 ms
  max            : 8018.2 ms

  browse   p95   : 2999.0 ms
  submit   p95   : 1775.0 ms
  purchase p95   : 4150.7 ms
==================================================

running (2m05.3s), 00/50 VUs, 2543 complete and 0 interrupted iterations
browse   ✓ [======================================] 00/35 VUs  2m0s
purchase ✓ [======================================] 0/5 VUs    2m0s
submit   ✓ [======================================] 00/10 VUs  2m0s
ERRO[0126] thresholds on metrics 'http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ 

### P=50 — v2 (safeJSON, single-token, 10.11.2026)

- Requests: 19480, Failed: 76.43%
- p95: 393.4 ms, max: 5544.9 ms (p50/p99 = 0.0 — баг summary, чинится)
- browse   p95=1381.0 ms, errors=81.04%
- submit   p95=592.3 ms,  errors=0.00%
- purchase p95=1208.8 ms, errors=100.00%
- status_409=0%, status_429=**73.62%**, status_5xx=1.74%
- Наблюдения:
  - **73.62% — rate limiter на одном admin-токене.** Не потолок системы.
  - purchase 100% fail: `user not found or no balance` — у admin нет `user_currencies`.
  - `cancel2 503 circuit=shop` — Circuit Breaker Shop **открылся** (каскад после 500).
  - CB работает как задумано — это плюс.
- Вывод: **single-token mode невалиден для highload**. Нужен пул юзеров с балансом.
- Дальше: seed 10 юзеров → SQL 100к баланс → P=50 с пулом.


===============

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

=====================================================================================

[denismatveev@c0der event_horizon]$ EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=200 EH_DURATION=90s EH_RAMP=30s \
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

     scenarios: (100.00%) 3 scenarios, 200 max VUs, 2m40s max duration (incl. graceful stop):
              * browse: Up to 140 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: browseScenario, gracefulStop: 30s)
              * purchase: Up to 20 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: purchaseScenario, gracefulStop: 30s)
              * submit: Up to 40 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: submitScenario, gracefulStop: 30s)

INFO[0007] 10k-probe: PROFILE=200 pool=10/10 (from 10)   source=console
INFO[0008] 10k-probe: PROFILE=200 DURATION=90s RAMP=30s unowned_items=10 pool=10  source=console
WARN[0043] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0043] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0043] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0044] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0045] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0046] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0047] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0047] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0051] cancel2 non-200: status=500 body={"error":"pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0061] cancel2 non-200: status=500 body={"error":"pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0068] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0069] cancel1 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0069] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0069] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0069] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0069] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0070] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0070] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0070] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0070] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0070] cancel2 non-200: status=500 body={"error":"pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0076] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0076] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0076] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0084] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0084] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0084] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0084] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] cancel1 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0085] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] cancel1 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] cancel2 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0086] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0087] purchase non-200: status=500 body={"error":"failed to record purchase: pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0092] purchase non-200: status=500 body={"error":"failed to record purchase: pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0093] purchase non-200: status=500 body={"error":"failed to record purchase: pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0095] cancel1 non-200: status=500 body={"error":"pq: deadlock detected (40P01)","grpc_code":"Internal"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] cancel2 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] cancel2 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0099] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0100] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0109] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0110] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] cancel2 non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] inventory/items non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] leaderboard non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0122] profile non-200: status=502 body=Backend error  source=console
WARN[0122] shop/items non-200: status=502 body=Backend error  source=console
WARN[0129] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0129] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0129] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0130] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0131] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0131] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0131] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0131] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0132] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0132] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0132] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0132] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0133] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0133] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0133] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0133] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0138] inventory/items non-200: status=503 body={"bulkhead":"inventory","error":"service overloaded"}{"error":"bulkhead full"}  source=console

==================================================
 10k-probe — PROFILE=200  DURATION=90s  RAMP=30s
==================================================
  Requests total : 10923
  Failed         : 3.42 %
  p50            : 2032.6 ms   (avg fallback)
  p95            : 5922.3 ms
  p99            : 35880.9 ms   (max fallback)
  max            : 35880.9 ms

  browse   p95   : 17309.0 ms   errors: 7.32%
  submit   p95   : 6860.6 ms   errors: 0.00%
  purchase p95   : 22024.6 ms   errors: 13.49%

  browse breakdown (p95):
    shop/items      : 5814.5 ms
    inventory/items : 6889.0 ms
    leaderboard     : 3644.4 ms
    profile         : 3587.8 ms

  status_409     : 1.71 %
  status_429     : 0.00 %
  status_5xx     : 1.74 %
  purchase_owned : 59.37 %   (409 item already owned — не ошибка)
==================================================

running (2m47.3s), 000/200 VUs, 3910 complete and 63 interrupted iterations
browse   ✓ [======================================] 001/140 VUs  2m30s
purchase ✓ [======================================] 00/20 VUs    2m30s
submit   ✓ [======================================] 00/40 VUs    2m30s
ERRO[0169] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 


================




[denismatveev@c0der event_horizon]$ EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=150 EH_DURATION=90s EH_RAMP=30s \
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

     scenarios: (100.00%) 3 scenarios, 150 max VUs, 2m40s max duration (incl. graceful stop):
              * browse: Up to 105 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: browseScenario, gracefulStop: 30s)
              * purchase: Up to 15 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: purchaseScenario, gracefulStop: 30s)
              * submit: Up to 30 looping VUs for 2m30s over 3 stages (gracefulRampDown: 10s, exec: submitScenario, gracefulStop: 30s)

INFO[0012] 10k-probe: PROFILE=150 pool=10/10 (from 10)   source=console
INFO[0012] 10k-probe: PROFILE=150 DURATION=90s RAMP=30s unowned_items=9 pool=10  source=console

==================================================
 10k-probe — PROFILE=150  DURATION=90s  RAMP=30s
==================================================
  Requests total : 12448
  Failed         : 0.22 %
  p50            : 1267.6 ms   (avg fallback)
  p95            : 2740.7 ms
  p99            : 8454.2 ms   (max fallback)
  max            : 8454.2 ms

  browse   p95   : 7358.9 ms   errors: 0.00%
  submit   p95   : 2292.1 ms   errors: 0.00%
  purchase p95   : 13148.0 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 3388.6 ms
    inventory/items : 2325.0 ms
    leaderboard     : 1984.0 ms
    profile         : 1542.0 ms

  status_409     : 0.22 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 11.64 %   (409 item already owned — не ошибка)
==================================================

running (2m44.4s), 000/150 VUs, 4365 complete and 0 interrupted iterations
browse   ✓ [======================================] 000/105 VUs  2m30s
purchase ✓ [======================================] 00/15 VUs    2m30s
submit   ✓ [======================================] 00/30 VUs    2m30s
ERRO[0165] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 