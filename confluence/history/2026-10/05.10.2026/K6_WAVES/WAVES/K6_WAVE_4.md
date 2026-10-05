[denismatveev@c0der event_horizon]$ export BASE_URL=http://localhost:8079
[denismatveev@c0der event_horizon]$ export EH_K6_EMAIL=admin@eventhorizon.local
[denismatveev@c0der event_horizon]$ export EH_K6_PASSWORD=changeme-dev-admin
[denismatveev@c0der event_horizon]$ 
[denismatveev@c0der event_horizon]$ TOKEN=$(curl -sS -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EH_K6_EMAIL\",\"password\":\"$EH_K6_PASSWORD\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
curl -sS "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN" \
  | python3 -c '
import sys,json
c=[]
for it in json.load(sys.stdin):
  if str(it.get("category") or "").lower()=="merch" or it.get("owned"): continue
  p=float(it.get("price") or 0)
  if p>0: c.append((p,it["id"],it["name"]))
c.sort()
print("PICK", c[0][1] if c else "")
for row in c[:5]: print(row)
'
[denismatveev@c0der event_horizon]$ TOKEN=$(curl -sS -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EH_K6_EMAIL\",\"password\":\"$EH_K6_PASSWORD\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
curl -sS "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN" \
  | python3 -c '
import sys,json
c=[]
for it in json.load(sys.stdin):
  if str(it.get("category") or "").lower()=="merch" or it.get("owned"): continue
  p=float(it.get("price") or 0)
  if p>0: c.append((p,it["id"],it["name"]))
c.sort()
print("PICK", c[0][1] if c else "")
for row in c[:5]: print(row)
'
PICK c586a47f-b6d3-49d3-a627-26903aa5e26e
(150.0, 'c586a47f-b6d3-49d3-a627-26903aa5e26e', 'Карточки со зверями')
(200.0, '5a3b1a32-dcae-4025-a8f5-da01748dab92', 'Золотая птичка')
(200.0, '80fcd762-aefc-4e1a-ac1e-b8884f2a00a1', 'Золотая птичка')
(200.0, '82be50db-670b-48c6-beb9-7e00d584f6de', 'Золотая птичка')
[denismatveev@c0der event_horizon]$ ^C
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=c586a47f-b6d3-49d3-a627-26903aa5e26e
[denismatveev@c0der event_horizon]$ curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s\n' \
  "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN"
shop 200 0.637410s
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
    ✗ 'rate<0.15' rate=95.91%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=489.66ms

    http_req_failed
    ✗ 'rate<0.05' rate=92.15%


  █ TOTAL RESULTS 

    checks_total.......: 98     3.028963/s
    checks_succeeded...: 4.08%  4 out of 98
    checks_failed......: 95.91% 94 out of 98

    ✗ purchase 200
      ↳  2% — ✓ 1 / ✗ 46
    ✗ purchase success
      ↳  2% — ✓ 1 / ✗ 46
    ✗ cancel1 200
      ↳  0% — ✓ 0 / ✗ 1
    ✗ cancel1 success
      ↳  0% — ✓ 0 / ✗ 1
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 95.91% 47 out of 49

    HTTP
    http_req_duration..............: avg=154.85ms min=7.94ms   med=49.25ms  max=1.84s p(90)=370.3ms  p(95)=489.66ms
      { expected_response:true }...: avg=947.01ms min=109.49ms med=917.92ms max=1.84s p(90)=1.66s    p(95)=1.75s   
    http_req_failed................: 92.15% 47 out of 51
    http_reqs......................: 51     1.576297/s

    EXECUTION
    iteration_duration.............: avg=639.55ms min=509.15ms med=556.66ms max=2.83s p(90)=779.76ms p(95)=871.2ms 
    iterations.....................: 47     1.452666/s
    vus............................: 1      min=0        max=1
    vus_max........................: 1      min=1        max=1

    NETWORK
    data_received..................: 63 kB  2.0 kB/s
    data_sent......................: 28 kB  861 B/s




running (0m32.4s), 0/1 VUs, 47 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  30s
ERRO[0033] thresholds on metrics 'errors, http_req_failed' have been crossed 
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
    ✗ 'p(95)<800' p(95)=3.13s

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 244     7.322886/s
    checks_succeeded...: 100.00% 244 out of 244
    checks_failed......: 0.00%   0 out of 244

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 122

    HTTP
    http_req_duration..............: avg=1.14s min=24.88ms  med=517.12ms max=3.46s p(90)=2.85s p(95)=3.13s
      { expected_response:true }...: avg=1.14s min=24.88ms  med=517.12ms max=3.46s p(90)=2.85s p(95)=3.13s
    http_req_failed................: 0.00%  0 out of 123
    http_reqs......................: 123    3.691455/s

    EXECUTION
    iteration_duration.............: avg=2.62s min=726.15ms med=2.73s    max=3.96s p(90)=3.66s p(95)=3.81s
    iterations.....................: 61     1.830722/s
    vus............................: 2      min=2        max=5
    vus_max........................: 5      min=5        max=5

    NETWORK
    data_received..................: 4.7 MB 140 kB/s
    data_sent......................: 59 kB  1.8 kB/s




running (0m33.3s), 0/5 VUs, 61 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0034] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ 