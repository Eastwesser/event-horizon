[denismatveev@c0der event_horizon]$ export BASE_URL=http://localhost:8079
[denismatveev@c0der event_horizon]$ export EH_K6_EMAIL=admin@eventhorizon.local
[denismatveev@c0der event_horizon]$ export EH_K6_PASSWORD=changeme-dev-admin
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=<unowned-non-merch-uuid>
K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js
K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/browse.js
curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s\n' \
  "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN"
bash: синтаксическая ошибка рядом с неожиданным маркером «newline»

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/purchase.js
        output: -

     scenarios: (100.00%) 1 scenario, 1 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 1 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✗ 'rate<0.15' rate=96.07%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=651.32ms

    http_req_failed
    ✗ 'rate<0.05' rate=92.45%


  █ TOTAL RESULTS 

    checks_total.......: 102    3.064607/s
    checks_succeeded...: 3.92%  4 out of 102
    checks_failed......: 96.07% 98 out of 102

    ✗ purchase 200
      ↳  2% — ✓ 1 / ✗ 48
    ✗ purchase success
      ↳  2% — ✓ 1 / ✗ 48
    ✗ cancel1 200
      ↳  0% — ✓ 0 / ✗ 1
    ✗ cancel1 success
      ↳  0% — ✓ 0 / ✗ 1
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 96.07% 49 out of 51

    HTTP
    http_req_duration..............: avg=153.59ms min=6.15ms   med=24.01ms  max=1.67s p(90)=406.06ms p(95)=651.32ms
      { expected_response:true }...: avg=852.81ms min=50.18ms  med=844.81ms max=1.67s p(90)=1.48s    p(95)=1.57s   
    http_req_failed................: 92.45% 49 out of 53
    http_reqs......................: 53     1.592394/s

    EXECUTION
    iteration_duration.............: avg=620.67ms min=506.77ms med=526.91ms max=1.32s p(90)=889.79ms p(95)=1.06s   
    iterations.....................: 49     1.472213/s
    vus............................: 1      min=0        max=1
    vus_max........................: 1      min=1        max=1

    NETWORK
    data_received..................: 64 kB  1.9 kB/s
    data_sent......................: 29 kB  871 B/s




running (0m33.3s), 0/1 VUs, 49 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
ERRO[0034] thresholds on metrics 'errors, http_req_failed' have been crossed 

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/browse.js
        output: -

     scenarios: (100.00%) 1 scenario, 5 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 5 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✓ 'rate<0.1' rate=5.10%

    http_req_duration
    ✗ 'p(95)<800' p(95)=1.12s

    http_req_failed
    ✗ 'rate<0.05' rate=5.08%


  █ TOTAL RESULTS 

    checks_total.......: 588    18.983389/s
    checks_succeeded...: 94.89% 558 out of 588
    checks_failed......: 5.10%  30 out of 588

    ✗ shop items 200
      ↳  89% — ✓ 132 / ✗ 15
    ✗ shop items is array
      ↳  89% — ✓ 132 / ✗ 15
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 5.10%  15 out of 294

    HTTP
    http_req_duration..............: avg=356.91ms min=14.34ms  med=87.95ms  max=1.96s p(90)=919.19ms p(95)=1.12s
      { expected_response:true }...: avg=373.74ms min=14.34ms  med=130.45ms max=1.96s p(90)=939.5ms  p(95)=1.12s
    http_req_failed................: 5.08%  15 out of 295
    http_reqs......................: 295    9.523979/s

    EXECUTION
    iteration_duration.............: avg=1.03s    min=353.85ms med=979.64ms max=2.44s p(90)=1.46s    p(95)=1.61s
    iterations.....................: 147    4.745847/s
    vus............................: 5      min=5         max=5
    vus_max........................: 5      min=5         max=5

    NETWORK
    data_received..................: 10 MB  337 kB/s
    data_sent......................: 142 kB 4.6 kB/s




running (0m31.0s), 0/5 VUs, 147 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0032] thresholds on metrics 'http_req_duration, http_req_failed' have been crossed 
shop 401 0.057222s
[denismatveev@c0der event_horizon]$ 