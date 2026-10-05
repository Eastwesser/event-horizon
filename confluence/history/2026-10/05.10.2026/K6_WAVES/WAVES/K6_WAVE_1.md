cd /home/denismatveev/event_horizon
export BASE_URL=http://localhost:8079
export EH_K6_EMAIL=admin@eventhorizon.local
export EH_K6_PASSWORD=changeme-dev-admin
# browse (default 20 VUs / 30s, p95 < 800ms)
k6 run deployments/k6/browse.js
# purchase + cancel idempotency (default 5 VUs / 30s, p95 < 1200ms)
k6 run deployments/k6/purchase.js

[denismatveev@c0der event_horizon]$ export EH_K6_EMAIL=admin@eventhorizon.local
[denismatveev@c0der event_horizon]$ export EH_K6_PASSWORD=changeme-dev-admin
[denismatveev@c0der event_horizon]$ k6 run deployments/k6/browse.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/browse.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 20 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✗ 'p(95)<800' p(95)=9.48s

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 320     9.189426/s
    checks_succeeded...: 100.00% 320 out of 320
    checks_failed......: 0.00%   0 out of 320

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 160

    HTTP
    http_req_duration..............: avg=3.91s min=37.27ms med=4.49s max=10.08s p(90)=9.11s  p(95)=9.48s 
      { expected_response:true }...: avg=3.91s min=37.27ms med=4.49s max=10.08s p(90)=9.11s  p(95)=9.48s 
    http_req_failed................: 0.00%  0 out of 161
    http_reqs......................: 161    4.62343/s

    EXECUTION
    iteration_duration.............: avg=8.18s min=4.68s   med=6.85s max=14.05s p(90)=13.71s p(95)=13.83s
    iterations.....................: 80     2.297356/s
    vus............................: 11     min=0        max=20
    vus_max........................: 20     min=20       max=20

    NETWORK
    data_received..................: 6.1 MB 175 kB/s
    data_sent......................: 78 kB  2.2 kB/s




running (0m34.8s), 00/20 VUs, 80 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  30s
ERRO[0035] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 


k6 run deployments/k6/purchase.js


ERRO[0035] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ k6 run deployments/k6/purchase.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/purchase.js
        output: -

     scenarios: (100.00%) 1 scenario, 5 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 5 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✗ 'rate<0.15' rate=100.00%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=69.04ms

    http_req_failed
    ✗ 'rate<0.05' rate=99.30%


  █ TOTAL RESULTS 

    checks_total.......: 570     18.324125/s
    checks_succeeded...: 0.00%   0 out of 570
    checks_failed......: 100.00% 570 out of 570

    ✗ purchase 200
      ↳  0% — ✓ 0 / ✗ 285
    ✗ purchase success
      ↳  0% — ✓ 0 / ✗ 285

    CUSTOM
    errors.........................: 100.00% 285 out of 285

    HTTP
    http_req_duration..............: avg=27.4ms   min=5.53ms   med=17.07ms  max=526.76ms p(90)=43.18ms  p(95)=69.04ms 
      { expected_response:true }...: avg=436.99ms min=347.21ms med=436.99ms max=526.76ms p(90)=508.81ms p(95)=517.79ms
    http_req_failed................: 99.30%  285 out of 287
    http_reqs......................: 287     9.226358/s

    EXECUTION
    iteration_duration.............: avg=528.76ms min=506.88ms med=520.49ms max=712.64ms p(90)=547.74ms p(95)=568.76ms
    iterations.....................: 285     9.162063/s
    vus............................: 3       min=3          max=5
    vus_max........................: 5       min=5          max=5

    NETWORK
    data_received..................: 110 kB  3.5 kB/s
    data_sent......................: 159 kB  5.1 kB/s




running (0m31.1s), 0/5 VUs, 285 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0031] thresholds on metrics 'errors, http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ 


K6_VUS=20 K6_DURATION=30s k6 run deployments/k6/browse.js


[denismatveev@c0der event_horizon]$ K6_VUS=20 K6_DURATION=30s k6 run deployments/k6/browse.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/browse.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 20 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✗ 'rate<0.1' rate=22.65%

    http_req_duration
    ✗ 'p(95)<800' p(95)=6.32s

    http_req_failed
    ✗ 'rate<0.05' rate=22.58%


  █ TOTAL RESULTS 

    checks_total.......: 724    21.054017/s
    checks_succeeded...: 77.34% 560 out of 724
    checks_failed......: 22.65% 164 out of 724

    ✗ shop items 200
      ↳  54% — ✓ 99 / ✗ 82
    ✗ shop items is array
      ↳  54% — ✓ 99 / ✗ 82
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 22.65% 82 out of 362

    HTTP
    http_req_duration..............: avg=1.57s min=6.41ms   med=169.94ms max=7.7s  p(90)=5.76s p(95)=6.32s
      { expected_response:true }...: avg=2.01s min=27.88ms  med=238.06ms max=7.7s  p(90)=5.97s p(95)=6.53s
    http_req_failed................: 22.58% 82 out of 363
    http_reqs......................: 363    10.556089/s

    EXECUTION
    iteration_duration.............: avg=3.48s min=353.77ms med=4.66s    max=8.12s p(90)=6.78s p(95)=7.25s
    iterations.....................: 181    5.263504/s
    vus............................: 2      min=2         max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 9.5 MB 278 kB/s
    data_sent......................: 175 kB 5.1 kB/s




running (0m34.4s), 00/20 VUs, 181 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  30s
ERRO[0035] thresholds on metrics 'errors, http_req_duration, http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ 

[denismatveev@c0der event_horizon]$ K6_VUS=5  K6_DURATION=30s k6 run deployments/k6/purchase.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/purchase.js
        output: -

     scenarios: (100.00%) 1 scenario, 5 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 5 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✗ 'rate<0.15' rate=100.00%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=87.42ms

    http_req_failed
    ✗ 'rate<0.05' rate=99.29%


  █ TOTAL RESULTS 

    checks_total.......: 566     17.481375/s
    checks_succeeded...: 0.00%   0 out of 566
    checks_failed......: 100.00% 566 out of 566

    ✗ purchase 200
      ↳  0% — ✓ 0 / ✗ 283
    ✗ purchase success
      ↳  0% — ✓ 0 / ✗ 283

    CUSTOM
    errors.........................: 100.00% 283 out of 283

    HTTP
    http_req_duration..............: avg=37.07ms  min=6.31ms   med=20.16ms max=1.26s    p(90)=60.65ms  p(95)=87.42ms 
      { expected_response:true }...: avg=1s       min=746.24ms med=1s      max=1.26s    p(90)=1.21s    p(95)=1.23s   
    http_req_failed................: 99.29%  283 out of 285
    http_reqs......................: 285     8.802459/s

    EXECUTION
    iteration_duration.............: avg=533.87ms min=507ms    med=523.6ms max=820.97ms p(90)=561.02ms p(95)=586.45ms
    iterations.....................: 283     8.740687/s
    vus............................: 5       min=0          max=5
    vus_max........................: 5       min=5          max=5

    NETWORK
    data_received..................: 110 kB  3.4 kB/s
    data_sent......................: 158 kB  4.9 kB/s




running (0m32.4s), 0/5 VUs, 283 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0033] thresholds on metrics 'errors, http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ 
