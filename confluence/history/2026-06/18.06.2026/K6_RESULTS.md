[denismatveev@c0der event_horizon]$ k6 run loadtest.js

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: loadtest.js
        output: -

     scenarios: (100.00%) 1 scenario, 100 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 100 looping VUs for 2m0s over 4 stages (gracefulRampDown: 30s, gracefulStop: 30s)

INFO[0000] ✅ Token obtained for k6load@test.com          source=console


  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<5000' p(95)=25.84ms

    http_req_failed
    ✗ 'rate<0.5' rate=99.39%


  █ TOTAL RESULTS 

    HTTP
    http_req_duration..............: avg=8.02ms  min=237.16µs med=4.82ms  max=144.43ms p(90)=17.27ms p(95)=25.84ms
      { expected_response:true }...: avg=17.93ms min=4.26ms   med=13.94ms max=78.68ms  p(90)=33.44ms p(95)=43.27ms
    http_req_failed................: 99.39% 18481 out of 18593
    http_reqs......................: 18593  153.619502/s

    EXECUTION
    iteration_duration.............: avg=1.03s   min=1s       med=1.03s   max=1.46s    p(90)=1.07s   p(95)=1.09s  
    iterations.....................: 4648   38.40281/s
    vus............................: 1      min=1              max=99 
    vus_max........................: 100    min=100            max=100

    NETWORK
    data_received..................: 3.9 MB 32 kB/s
    data_sent......................: 10 MB  85 kB/s




running (2m01.0s), 000/100 VUs, 4648 complete and 0 interrupted iterations
default ✓ [======================================] 000/100 VUs  2m0s
ERRO[0121] thresholds on metrics 'http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ 


[denismatveev@c0der event_horizon]$ k6 run loadtest.js 2>&1 | head -100

         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: loadtest.js
        output: -

     scenarios: (100.00%) 1 scenario, 100 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 100 looping VUs for 2m0s over 4 stages (gracefulRampDown: 30s, gracefulStop: 30s)

time="2026-06-18T20:07:09+03:00" level=info msg="✅ Token obtained" source=console

running (0m01.0s), 001/100 VUs, 0 complete and 0 interrupted iterations
default   [   1% ] 001/100 VUs  0m00.9s/2m00.0s

running (0m02.0s), 001/100 VUs, 1 complete and 0 interrupted iterations
default   [   2% ] 001/100 VUs  0m01.9s/2m00.0s

running (0m03.0s), 001/100 VUs, 2 complete and 0 interrupted iterations
default   [   2% ] 001/100 VUs  0m02.9s/2m00.0s

running (0m04.0s), 002/100 VUs, 3 complete and 0 interrupted iterations
default   [   3% ] 002/100 VUs  0m03.9s/2m00.0s

running (0m05.1s), 002/100 VUs, 5 complete and 0 interrupted iterations
default   [   4% ] 002/100 VUs  0m05.0s/2m00.0s

running (0m06.0s), 002/100 VUs, 7 complete and 0 interrupted iterations
default   [   5% ] 002/100 VUs  0m05.9s/2m00.0s
time="2026-06-18T20:07:16+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:16+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console

running (0m07.0s), 003/100 VUs, 9 complete and 0 interrupted iterations
default   [   6% ] 003/100 VUs  0m06.9s/2m00.0s
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:17+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console

running (0m08.0s), 003/100 VUs, 12 complete and 0 interrupted iterations
default   [   7% ] 003/100 VUs  0m07.9s/2m00.0s
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:18+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console

running (0m09.0s), 003/100 VUs, 15 complete and 0 interrupted iterations
default   [   7% ] 003/100 VUs  0m08.9s/2m00.0s
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console

running (0m10.0s), 003/100 VUs, 18 complete and 0 interrupted iterations
default   [   8% ] 003/100 VUs  0m09.9s/2m00.0s
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:19+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console

running (0m11.0s), 004/100 VUs, 21 complete and 0 interrupted iterations
default   [   9% ] 004/100 VUs  0m10.9s/2m00.0s
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:20+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Towers: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Hexagon: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Memory: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
time="2026-06-18T20:07:21+03:00" level=info msg="❌ Flappy: 429 - {\"error\":\"Too many requests. Please try again later.\",\"retry_after\":1}" source=console
[denismatveev@c0der event_horizon]$ 