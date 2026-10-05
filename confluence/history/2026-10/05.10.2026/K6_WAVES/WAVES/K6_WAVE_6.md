[denismatveev@c0der event_horizon]$ export BASE_URL=http://localhost:8079
[denismatveev@c0der event_horizon]$ export EH_K6_EMAIL=admin@eventhorizon.local
[denismatveev@c0der event_horizon]$ export EH_K6_PASSWORD=changeme-dev-admin
[denismatveev@c0der event_horizon]$ unset EH_K6_ITEM_ID
[denismatveev@c0der event_horizon]$ K6_VUS=1 K6_DURATION=30s k6 run deployments/k6/purchase.js

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

INFO[0000] purchase.js: assigned 1 item(s) to VUs: c586a47f-b6d3-49d3-a627-26903aa5e26e  source=console


  █ THRESHOLDS 

    errors
    ✓ 'rate<0.15' rate=0.00%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=300.31ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 222     7.115554/s
    checks_succeeded...: 100.00% 222 out of 222
    checks_failed......: 0.00%   0 out of 222

    ✓ purchase 200
    ✓ purchase success
    ✓ cancel1 200
    ✓ cancel1 success
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 0.00% 0 out of 111

    HTTP
    http_req_duration..............: avg=109.36ms min=12.45ms  med=83.09ms  max=452.65ms p(90)=232.84ms p(95)=300.31ms
      { expected_response:true }...: avg=109.36ms min=12.45ms  med=83.09ms  max=452.65ms p(90)=232.84ms p(95)=300.31ms
    http_req_failed................: 0.00% 0 out of 113
    http_reqs......................: 113   3.621881/s

    EXECUTION
    iteration_duration.............: avg=827.35ms min=601.51ms med=772.19ms max=1.16s    p(90)=1.07s    p(95)=1.14s   
    iterations.....................: 37    1.185926/s
    vus............................: 1     min=1        max=1
    vus_max........................: 1     min=1        max=1

    NETWORK
    data_received..................: 78 kB 2.5 kB/s
    data_sent......................: 62 kB 2.0 kB/s




running (0m31.2s), 0/1 VUs, 37 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
[denismatveev@c0der event_horizon]$ K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/purchase.js

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

WARN[0000] only 4 unowned item(s) for 5 VUs — VUs > 4 will idle  source=console
INFO[0000] purchase.js: assigned 4 item(s) to VUs: c586a47f-b6d3-49d3-a627-26903aa5e26e, 80fcd762-aefc-4e1a-ac1e-b8884f2a00a1, 5a3b1a32-dcae-4025-a8f5-da01748dab92, 82be50db-670b-48c6-beb9-7e00d584f6de  source=console


  █ THRESHOLDS 

    errors
    ✓ 'rate<0.15' rate=0.00%

    http_req_duration
    ✗ 'p(95)<1200' p(95)=2.86s

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 264     8.378466/s
    checks_succeeded...: 100.00% 264 out of 264
    checks_failed......: 0.00%   0 out of 264

    ✓ purchase 200
    ✓ purchase success
    ✓ cancel1 200
    ✓ cancel1 success
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 0.00% 0 out of 132

    HTTP
    http_req_duration..............: avg=737.83ms min=21.86ms  med=309.13ms max=5.95s p(90)=1.98s p(95)=2.86s
      { expected_response:true }...: avg=737.83ms min=21.86ms  med=309.13ms max=5.95s p(90)=1.98s p(95)=2.86s
    http_req_failed................: 0.00% 0 out of 134
    http_reqs......................: 134   4.252706/s

    EXECUTION
    iteration_duration.............: avg=2.07s    min=800.83ms med=1.09s    max=8.73s p(90)=5.1s  p(95)=6.49s
    iterations.....................: 74    2.34851/s
    vus............................: 4     min=4        max=5
    vus_max........................: 5     min=5        max=5

    NETWORK
    data_received..................: 83 kB 2.6 kB/s
    data_sent......................: 73 kB 2.3 kB/s




running (0m31.5s), 0/5 VUs, 74 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0032] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 