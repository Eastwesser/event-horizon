k6 run -e TARGET=10 loadtest.js

[denismatveev@c0der event_horizon]$ k6 run -e TARGET=10 loadtest.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: loadtest.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 10 looping VUs for 2m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)

INFO[0000] ✅ Token obtained                              source=console


  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<5000' p(95)=66.81ms


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=23.49ms  min=887.92µs med=15.54ms  max=488.01ms p(90)=46.73ms  p(95)=66.81ms
      { expected_response:true }...: avg=23.49ms  min=887.92µs med=15.54ms  max=488.01ms p(90)=46.73ms  p(95)=66.81ms
    http_req_failed................: 0.00%  0 out of 6081
    http_reqs......................: 6081   50.465985/s

    EXECUTION
    iteration_duration.............: avg=604.51ms min=521ms    med=585.55ms max=1.41s    p(90)=680.78ms p(95)=721.7ms
    iterations.....................: 1520   12.614422/s
    vus............................: 1      min=1         max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 1.4 MB 12 kB/s
    data_sent......................: 3.4 MB 28 kB/s




running (2m00.5s), 00/10 VUs, 1520 complete and 0 interrupted iterations
default ✓ [======================================] 00/10 VUs  2m0s
[denismatveev@c0der event_horizon]$ 

-----------------
k6 run -e TARGET=100 loadtest.js

[denismatveev@c0der event_horizon]$ k6 run -e TARGET=100 loadtest.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: loadtest.js
        output: -

     scenarios: (100.00%) 1 scenario, 100 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 100 looping VUs for 2m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)

INFO[0000] ✅ Token obtained                              source=console


  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<5000' p(95)=258.08ms


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=91.89ms  min=1.41ms   med=65.58ms  max=917.13ms p(90)=200.14ms p(95)=258.08ms
      { expected_response:true }...: avg=91.89ms  min=1.41ms   med=65.58ms  max=917.13ms p(90)=200.14ms p(95)=258.08ms
    http_req_failed................: 0.00%  0 out of 40953
    http_reqs......................: 40953  339.674459/s

    EXECUTION
    iteration_duration.............: avg=884.07ms min=516.05ms med=826.17ms max=2.15s    p(90)=1.21s    p(95)=1.34s   
    iterations.....................: 10238  84.916541/s
    vus............................: 3      min=3          max=100
    vus_max........................: 100    min=100        max=100

    NETWORK
    data_received..................: 9.7 MB 80 kB/s
    data_sent......................: 23 MB  189 kB/s




running (2m00.6s), 000/100 VUs, 10238 complete and 0 interrupted iterations
default ✓ [======================================] 000/100 VUs  2m0s
[denismatveev@c0der event_horizon]$ 

-----------------
k6 run -e TARGET=1000 loadtest.js

WARN[0098] Request Failed                                error="Post \"http://localhost:8080/api/game/submit\": request timeout"
WARN[0098] Request Failed                                error="Post \"http://localhost:8080/api/game/submit\": request timeout"


  █ THRESHOLDS 

    http_req_duration
    ✗ 'p(95)<5000' p(95)=5s


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=1.28s min=2.03ms   med=698.63ms max=8.07s p(90)=3.76s  p(95)=5s    
      { expected_response:true }...: avg=1.04s min=2.03ms   med=643.51ms max=6.48s p(90)=2.6s   p(95)=3.42s 
    http_req_failed................: 5.90% 2788 out of 47209
    http_reqs......................: 47209 390.382781/s

    EXECUTION
    iteration_duration.............: avg=7.65s min=535.67ms med=5.09s    max=55.1s p(90)=19.32s p(95)=26.99s
    iterations.....................: 11802 97.593628/s
    vus............................: 14    min=0             max=1000
    vus_max........................: 1000  min=628           max=1000

    NETWORK
    data_received..................: 11 MB 87 kB/s
    data_sent......................: 26 MB 217 kB/s




running (2m00.9s), 0000/1000 VUs, 11802 complete and 0 interrupted iterations
default ✓ [======================================] 0000/1000 VUs  2m0s
ERRO[0123] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 

-----------------
