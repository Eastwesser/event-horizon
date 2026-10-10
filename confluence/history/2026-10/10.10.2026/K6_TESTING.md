[denismatveev@c0der event_horizon]$ K6_VUS=20 K6_DURATION=60s make test-k6
CORE k6 browse.js → balancer :8079 (requires stack + k6)
BASE_URL=http://localhost:8079  user=admin@eventhorizon.local

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/browse.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m30s max duration (incl. graceful stop):
              * default: 20 looping VUs for 1m0s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✓ 'p(95)<800' p(95)=429.18ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 6232    100.857944/s
    checks_succeeded...: 100.00% 6232 out of 6232
    checks_failed......: 0.00%   0 out of 6232

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 3116

    HTTP
    http_req_duration..............: avg=229.69ms min=35.54ms  med=208.7ms  max=1.39s p(90)=354.43ms p(95)=429.18ms
      { expected_response:true }...: avg=229.69ms min=35.54ms  med=208.7ms  max=1.39s p(90)=354.43ms p(95)=429.18ms
    http_req_failed................: 0.00%  0 out of 3117
    http_reqs......................: 3117   50.445156/s

    EXECUTION
    iteration_duration.............: avg=775.09ms min=447.55ms med=741.75ms max=2.35s p(90)=963.72ms p(95)=1.07s
    iterations.....................: 1558   25.214486/s
    vus............................: 20     min=0         max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 113 MB 1.8 MB/s
    data_sent......................: 1.5 MB 25 kB/s




running (1m01.8s), 00/20 VUs, 1558 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  1m0s
[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ K6_VUS=50 K6_DURATION=60s make test-k6
CORE k6 browse.js → balancer :8079 (requires stack + k6)
BASE_URL=http://localhost:8079  user=admin@eventhorizon.local

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: deployments/k6/browse.js
        output: -

     scenarios: (100.00%) 1 scenario, 50 max VUs, 1m30s max duration (incl. graceful stop):
              * default: 50 looping VUs for 1m0s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✗ 'p(95)<800' p(95)=958.98ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 9560    154.542149/s
    checks_succeeded...: 100.00% 9560 out of 9560
    checks_failed......: 0.00%   0 out of 9560

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 4780

    HTTP
    http_req_duration..............: avg=481.76ms min=82.33ms  med=429.26ms max=2.12s p(90)=759.75ms p(95)=958.98ms
      { expected_response:true }...: avg=481.76ms min=82.33ms  med=429.26ms max=2.12s p(90)=759.75ms p(95)=958.98ms
    http_req_failed................: 0.00%  0 out of 4781
    http_reqs......................: 4781   77.28724/s

    EXECUTION
    iteration_duration.............: avg=1.27s    min=599.56ms med=1.2s     max=3.27s p(90)=1.72s    p(95)=1.99s
    iterations.....................: 2390   38.635537/s
    vus............................: 40     min=40        max=50
    vus_max........................: 50     min=50        max=50

    NETWORK
    data_received..................: 173 MB 2.8 MB/s
    data_sent......................: 2.3 MB 38 kB/s




running (1m01.9s), 00/50 VUs, 2390 complete and 0 interrupted iterations
default ✓ [======================================] 50 VUs  1m0s
ERRO[0062] thresholds on metrics 'http_req_duration' have been crossed 
make: *** [Makefile:118: test-k6] Ошибка 99
[denismatveev@c0der event_horizon]$ 



[denismatveev@c0der event_horizon]$ K6_VUS=5 make test-k6-purchase

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
INFO[0000] purchase.js: assigned 4 item(s) to VUs: c586a47f-b6d3-49d3-a627-26903aa5e26e, 82be50db-670b-48c6-beb9-7e00d584f6de, b2222222-2222-4222-8222-222222222201, b2222222-2222-4222-8222-222222222202  source=console


  █ THRESHOLDS 

    errors
    ✓ 'rate<0.15' rate=0.00%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=196.06ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 900     28.720633/s
    checks_succeeded...: 100.00% 900 out of 900
    checks_failed......: 0.00%   0 out of 900

    ✓ purchase 200
    ✓ purchase success
    ✓ cancel1 200
    ✓ cancel1 success
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 0.00%  0 out of 450

    HTTP
    http_req_duration..............: avg=102.88ms min=18.6ms   med=89.34ms  max=585.66ms p(90)=164.2ms p(95)=196.06ms
      { expected_response:true }...: avg=102.88ms min=18.6ms   med=89.34ms  max=585.66ms p(90)=164.2ms p(95)=196.06ms
    http_req_failed................: 0.00%  0 out of 452
    http_reqs......................: 452    14.42414/s

    EXECUTION
    iteration_duration.............: avg=843.52ms min=625.88ms med=798.35ms max=1.36s    p(90)=1s      p(95)=1.01s
    iterations.....................: 180    5.744127/s
    vus............................: 3      min=3        max=5
    vus_max........................: 5      min=5        max=5

    NETWORK
    data_received..................: 150 kB 4.8 kB/s
    data_sent......................: 250 kB 8.0 kB/s




running (0m31.3s), 0/5 VUs, 180 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
[denismatveev@c0der event_horizon]$ 