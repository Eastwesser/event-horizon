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

INFO[0008] 10k-probe: PROFILE=50 pool=10/10 (from 10)    source=console
INFO[0008] 10k-probe: PROFILE=50 DURATION=60s RAMP=30s unowned_items=10 pool=10  source=console

==================================================
 10k-probe — PROFILE=50  DURATION=60s  RAMP=30s
==================================================
  Requests total : 5959
  Failed         : 1.11 %
  p50            : 586.6 ms   (avg fallback)
  p95            : 1337.5 ms
  p99            : 5149.1 ms   (max fallback)
  max            : 5149.1 ms

  browse   p95   : 4222.6 ms   errors: 0.00%
  submit   p95   : 1750.3 ms   errors: 0.00%
  purchase p95   : 4397.3 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 1313.6 ms
    inventory/items : 1273.8 ms
    leaderboard     : 1221.0 ms
    profile         : 1084.0 ms

  status_409     : 1.11 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 39.76 %   (409 item already owned — не ошибка)
==================================================

running (2m08.3s), 00/50 VUs, 1953 complete and 0 interrupted iterations
browse   ✓ [======================================] 00/35 VUs  2m0s
purchase ✓ [======================================] 0/5 VUs    2m0s
submit   ✓ [======================================] 00/10 VUs  2m0s
[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \
  EH_PROFILE=100 EH_DURATION=60s EH_RAMP=30s \
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

     scenarios: (100.00%) 3 scenarios, 100 max VUs, 2m10s max duration (incl. graceful stop):
              * browse: Up to 70 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: browseScenario, gracefulStop: 30s)
              * purchase: Up to 10 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: purchaseScenario, gracefulStop: 30s)
              * submit: Up to 20 looping VUs for 2m0s over 3 stages (gracefulRampDown: 10s, exec: submitScenario, gracefulStop: 30s)

INFO[0006] 10k-probe: PROFILE=100 pool=10/10 (from 10)   source=console
INFO[0006] 10k-probe: PROFILE=100 DURATION=60s RAMP=30s unowned_items=10 pool=10  source=console

==================================================
 10k-probe — PROFILE=100  DURATION=60s  RAMP=30s
==================================================
  Requests total : 8904
  Failed         : 0.69 %
  p50            : 834.8 ms   (avg fallback)
  p95            : 1827.8 ms
  p99            : 5912.8 ms   (max fallback)
  max            : 5912.8 ms

  browse   p95   : 5057.0 ms   errors: 0.00%
  submit   p95   : 1600.0 ms   errors: 0.00%
  purchase p95   : 7975.4 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 2242.9 ms
    inventory/items : 1715.0 ms
    leaderboard     : 1329.6 ms
    profile         : 1205.2 ms

  status_409     : 0.69 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 29.19 %   (409 item already owned — не ошибка)
==================================================

running (2m06.5s), 000/100 VUs, 3032 complete and 0 interrupted iterations
browse   ✓ [======================================] 00/70 VUs  2m0s
purchase ✓ [======================================] 00/10 VUs  2m0s
submit   ✓ [======================================] 00/20 VUs  2m0s
[denismatveev@c0der event_horizon]$ 










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

INFO[0012] 10k-probe: PROFILE=200 pool=10/10 (from 10)   source=console
INFO[0012] 10k-probe: PROFILE=200 DURATION=90s RAMP=30s unowned_items=10 pool=10  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0041] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0042] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0042] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0042] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0059] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0059] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0065] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] cancel1 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0066] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] cancel2 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0067] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0092] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0093] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0097] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0098] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0101] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0102] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0103] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0107] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0108] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0108] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0108] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0108] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0109] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0109] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0109] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0120] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] cancel1 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0120] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] cancel2 non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0121] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0122] shop/items non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console
WARN[0122] cancel1 non-200: status=500 body={"error":"lock item: pq: canceling statement due to lock timeout (55P03)","grpc_code":"Internal"}  source=console
WARN[0128] purchase non-200: status=503 body={"bulkhead":"shop","error":"service overloaded"}{"error":"bulkhead full"}  source=console

==================================================
 10k-probe — PROFILE=200  DURATION=90s  RAMP=30s
==================================================
  Requests total : 14187
  Failed         : 2.87 %
  p50            : 1511.4 ms   (avg fallback)
  p95            : 3977.3 ms
  p99            : 12231.2 ms   (max fallback)
  max            : 12231.2 ms

  browse   p95   : 9477.5 ms   errors: 7.15%
  submit   p95   : 4131.8 ms   errors: 0.00%
  purchase p95   : 14686.1 ms   errors: 5.39%

  browse breakdown (p95):
    shop/items      : 4687.3 ms
    inventory/items : 3687.5 ms
    leaderboard     : 1877.0 ms
    profile         : 1733.0 ms

  status_409     : 1.29 %
  status_429     : 0.00 %
  status_5xx     : 1.60 %
  purchase_owned : 52.14 %   (409 item already owned — не ошибка)
==================================================

running (2m40.2s), 000/200 VUs, 5032 complete and 1 interrupted iterations
browse   ✓ [======================================] 000/140 VUs  2m30s
purchase ✓ [======================================] 00/20 VUs    2m30s
submit   ✓ [======================================] 00/40 VUs    2m30s
ERRO[0163] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 
