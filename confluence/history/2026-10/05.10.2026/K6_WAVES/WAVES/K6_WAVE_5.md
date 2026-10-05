denismatveev@c0der event_horizon]$ TOKEN=$(curl -sS -X POST http://localhost:8079/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
[denismatveev@c0der event_horizon]$ curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s\n' \
  http://localhost:8079/api/shop/items -H "Authorization: Bearer $TOKEN"
shop 200 1.083702s
[denismatveev@c0der event_horizon]$ export BASE_URL=http://localhost:8079
[denismatveev@c0der event_horizon]$ export EH_K6_EMAIL=admin@eventhorizon.local
[denismatveev@c0der event_horizon]$ export EH_K6_PASSWORD=changeme-dev-admin
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=PASTE_ACTUAL_UUID_HERE
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=c586a47f-b6d3-49d3-a627-26903aa5e26e
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



  █ THRESHOLDS 

    errors
    ✗ 'rate<0.15' rate=100.00%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=149.58ms

    http_req_failed
    ✗ 'rate<0.05' rate=96.42%


  █ TOTAL RESULTS 

    checks_total.......: 108     3.479129/s
    checks_succeeded...: 0.00%   0 out of 108
    checks_failed......: 100.00% 108 out of 108

    ✗ purchase 200
      ↳  0% — ✓ 0 / ✗ 54
    ✗ purchase success
      ↳  0% — ✓ 0 / ✗ 54

    CUSTOM
    errors.........................: 100.00% 54 out of 54

    HTTP
    http_req_duration..............: avg=58.47ms  min=8.08ms   med=34.44ms  max=620.78ms p(90)=83.03ms  p(95)=149.58ms
      { expected_response:true }...: avg=322.71ms min=24.64ms  med=322.71ms max=620.78ms p(90)=561.17ms p(95)=590.98ms
    http_req_failed................: 96.42%  54 out of 56
    http_reqs......................: 56      1.803993/s

    EXECUTION
    iteration_duration.............: avg=558.73ms min=509.45ms med=541.84ms max=879.61ms p(90)=609.8ms  p(95)=664.28ms
    iterations.....................: 54      1.739565/s
    vus............................: 1       min=1        max=1
    vus_max........................: 1       min=1        max=1

    NETWORK
    data_received..................: 64 kB   2.1 kB/s
    data_sent......................: 31 kB   988 B/s




running (0m31.0s), 0/1 VUs, 54 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
ERRO[0032] thresholds on metrics 'errors, http_req_failed' have been crossed 
[denismatveev@c0der event_horizon]$ K6_VUS=5 K6_DURATION=30s k6 run deployments/k6/browse.js

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
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✓ 'p(95)<800' p(95)=354.59ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 944     29.841239/s
    checks_succeeded...: 100.00% 944 out of 944
    checks_failed......: 0.00%   0 out of 944

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 472

    HTTP
    http_req_duration..............: avg=161.87ms min=20.37ms  med=132.15ms max=989.75ms p(90)=267.03ms p(95)=354.59ms
      { expected_response:true }...: avg=161.87ms min=20.37ms  med=132.15ms max=989.75ms p(90)=267.03ms p(95)=354.59ms
    http_req_failed................: 0.00%  0 out of 473
    http_reqs......................: 473    14.952231/s

    EXECUTION
    iteration_duration.............: avg=645ms    min=355.09ms med=596.24ms max=1.73s    p(90)=877.42ms p(95)=1.02s   
    iterations.....................: 236    7.46031/s
    vus............................: 5      min=0        max=5
    vus_max........................: 5      min=5        max=5

    NETWORK
    data_received..................: 18 MB  569 kB/s
    data_sent......................: 228 kB 7.2 kB/s




running (0m31.6s), 0/5 VUs, 236 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
[denismatveev@c0der event_horizon]$ 


And this

running (0m31.6s), 0/5 VUs, 236 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
[denismatveev@c0der event_horizon]$ curl -sS http://localhost:8079/api/shop/items -H "Authorization: Bearer $TOKEN" \
  | python3 -c '
import sys,json
for it in json.load(sys.stdin):
  if it["id"]=="c586a47f-b6d3-49d3-a627-26903aa5e26e":
    print(it.get("name"), "owned=", it.get("owned"), "price=", it.get("price"))
'
Карточки со зверями owned= True price= 150
[denismatveev@c0der event_horizon]$ curl -sS -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"c586a47f-b6d3-49d3-a627-26903aa5e26e"}' | jq .
{
  "error": "item already owned",
  "grpc_code": "AlreadyExists"
}
[denismatveev@c0der event_horizon]$ curl -sS -X POST http://localhost:8079/api/shop/purchase/c586a47f-b6d3-49d3-a627-26903aa5e26e/cancel \
  -H "Authorization: Bearer $TOKEN" | jq .
{
  "already_refunded": false,
  "message": "Purchase cancelled",
  "new_balance": 975244,
  "refunded_amount": 150,
  "success": true
}
[denismatveev@c0der event_horizon]$ 
