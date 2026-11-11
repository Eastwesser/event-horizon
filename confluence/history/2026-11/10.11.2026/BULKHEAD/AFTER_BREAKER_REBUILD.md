[denismatveev@c0der event_horizon]$ curl -fsS http://localhost/health
[denismatveev@c0der event_horizon]$ EH_K6_USERS_JSON="$(cat load_test_results/k6_users.json)" \load_test_results/k6_users.json)" \
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
  Requests total : 6986
  Failed         : 0.30 %
  p50            : 489.2 ms   (avg fallback)
  p95            : 1070.4 ms
  p99            : 4934.9 ms   (max fallback)
  max            : 4934.9 ms

  browse   p95   : 3165.8 ms   errors: 0.00%
  submit   p95   : 1145.6 ms   errors: 0.00%
  purchase p95   : 3462.5 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 1070.6 ms
    inventory/items : 1061.6 ms
    leaderboard     : 937.2 ms
    profile         : 814.2 ms

  status_409     : 0.30 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 12.65 %   (409 item already owned — не ошибка)
==================================================

running (2m05.7s), 00/50 VUs, 2218 complete and 0 interrupted iterations
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

INFO[0004] 10k-probe: PROFILE=100 pool=10/10 (from 10)   source=console
INFO[0004] 10k-probe: PROFILE=100 DURATION=60s RAMP=30s unowned_items=10 pool=10  source=console

==================================================
 10k-probe — PROFILE=100  DURATION=60s  RAMP=30s
==================================================
  Requests total : 8386
  Failed         : 1.72 %
  p50            : 892.6 ms   (avg fallback)
  p95            : 2209.1 ms
  p99            : 5095.7 ms   (max fallback)
  max            : 5095.7 ms

  browse   p95   : 5900.3 ms   errors: 0.00%
  submit   p95   : 2348.8 ms   errors: 0.00%
  purchase p95   : 7031.6 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 2489.3 ms
    inventory/items : 1975.0 ms
    leaderboard     : 1905.8 ms
    profile         : 1525.8 ms

  status_409     : 1.72 %
  status_429     : 0.00 %
  status_5xx     : 0.00 %
  purchase_owned : 57.14 %   (409 item already owned — не ошибка)
==================================================

running (2m04.9s), 000/100 VUs, 2891 complete and 0 interrupted iterations
browse   ✓ [======================================] 00/70 VUs  2m0s
purchase ✓ [======================================] 00/10 VUs  2m0s
submit   ✓ [======================================] 00/20 VUs  2m0s
ERRO[0125] thresholds on metrics 'http_req_duration' have been crossed 
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

INFO[0004] 10k-probe: PROFILE=200 pool=10/10 (from 10)   source=console
INFO[0004] 10k-probe: PROFILE=200 DURATION=90s RAMP=30s unowned_items=10 pool=10  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0076] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0077] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0078] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0079] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0080] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0081] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0081] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0082] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0082] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0082] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0082] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0083] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0083] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0083] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0084] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0085] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0086] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0087] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0088] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0089] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0090] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0090] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0091] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
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
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0094] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0095] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0096] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0097] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0098] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0099] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0100] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0101] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
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
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0104] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0105] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
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
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console
WARN[0106] inventory/items non-200: status=503 body={"circuit":"inventory","error":"service temporarily unavailable"}  source=console

==================================================
 10k-probe — PROFILE=200  DURATION=90s  RAMP=30s
==================================================
  Requests total : 12196
  Failed         : 3.21 %
  p50            : 1790.5 ms   (avg fallback)
  p95            : 4913.8 ms
  p99            : 18763.6 ms   (max fallback)
  max            : 18763.6 ms

  browse   p95   : 11514.6 ms   errors: 9.34%
  submit   p95   : 5569.4 ms   errors: 0.00%
  purchase p95   : 18623.5 ms   errors: 0.00%

  browse breakdown (p95):
    shop/items      : 5797.9 ms
    inventory/items : 4455.2 ms
    leaderboard     : 2223.7 ms
    profile         : 2040.0 ms

  status_409     : 1.27 %
  status_429     : 0.00 %
  status_5xx     : 1.96 %
  purchase_owned : 51.16 %   (409 item already owned — не ошибка)
==================================================

running (2m35.0s), 000/200 VUs, 4304 complete and 1 interrupted iterations
browse   ✓ [======================================] 000/140 VUs  2m30s
purchase ✓ [======================================] 00/20 VUs    2m30s
submit   ✓ [======================================] 00/40 VUs    2m30s
ERRO[0155] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 

