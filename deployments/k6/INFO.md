# To run:

cd deployments/k6

k6 run e2e-test.js

[denismatveev@c0der event_horizon]$ cd deployments/k6d deployments/k6
k6 run e2e-test.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: e2e-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 1 max VUs, 40s max duration (incl. graceful stop):
              * default: 1 looping VUs for 10s (gracefulStop: 30s)



  █ TOTAL RESULTS 

    checks_total.......: 20      1.716161/s
    checks_succeeded...: 100.00% 20 out of 20
    checks_failed......: 0.00%   0 out of 20

    ✓ ✅ регистрация успешна
    ✓ ✅ логин успешен
    ✓ ✅ рекорд отправлен
    ✓ ✅ лидерборд доступен

    HTTP
    http_req_duration..............: avg=205.47ms min=11.37ms med=72.11ms max=704.71ms p(90)=618.55ms p(95)=662.42ms
      { expected_response:true }...: avg=205.47ms min=11.37ms med=72.11ms max=704.71ms p(90)=618.55ms p(95)=662.42ms
    http_req_failed................: 0.00%  0 out of 20
    http_reqs......................: 20     1.716161/s

    EXECUTION
    iteration_duration.............: avg=2.33s    min=2.28s   med=2.34s   max=2.36s    p(90)=2.36s    p(95)=2.36s   
    iterations.....................: 5      0.42904/s
    vus............................: 1      min=1       max=1
    vus_max........................: 1      min=1       max=1

    NETWORK
    data_received..................: 5.9 kB 505 B/s
    data_sent......................: 5.2 kB 448 B/s




running (11.7s), 0/1 VUs, 5 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  10s