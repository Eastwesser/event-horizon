[denismatveev@c0der event_horizon]$ make test-k6  
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

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m0s max duration (incl. graceful stop):
              * default: 20 looping VUs for 30s (gracefulStop: 30s)



  █ THRESHOLDS 

    errors
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✓ 'p(95)<800' p(95)=539.16ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 2596    81.63476/s
    checks_succeeded...: 100.00% 2596 out of 2596
    checks_failed......: 0.00%   0 out of 2596

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 1298

    HTTP
    http_req_duration..............: avg=312.35ms min=22.05ms  med=215.45ms max=4.8s  p(90)=410.93ms p(95)=539.16ms
      { expected_response:true }...: avg=312.35ms min=22.05ms  med=215.45ms max=4.8s  p(90)=410.93ms p(95)=539.16ms
    http_req_failed................: 0.00%  0 out of 1299
    http_reqs......................: 1299   40.848826/s

    EXECUTION
    iteration_duration.............: avg=936.05ms min=452.65ms med=757.03ms max=5.82s p(90)=1.07s    p(95)=1.25s   
    iterations.....................: 649    20.40869/s
    vus............................: 20     min=0         max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 47 MB  1.5 MB/s
    data_sent......................: 632 kB 20 kB/s




running (0m31.8s), 00/20 VUs, 649 complete and 0 interrupted iterations
default ✓ [======================================] 20 VUs  30s
[denismatveev@c0der event_horizon]$ 