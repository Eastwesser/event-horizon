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
items=json.load(sys.stdin)
c=[]
for it in items:
  if str(it.get("category") or "").lower()=="merch": continue
  if it.get("owned"): continue
  p=float(it.get("price") or 0)
  if p>0: c.append((p,it["id"],it["name"]))
c.sort()
print("count", len(c))
for row in c[:8]: print(row)
print("PICK", c[0][1] if c else "")
'
count 5
(100.0, 'a61fb41c-2448-43fd-b088-518690a5ab1b', 'Радужные трубы')
(150.0, 'c586a47f-b6d3-49d3-a627-26903aa5e26e', 'Карточки со зверями')
(200.0, '5a3b1a32-dcae-4025-a8f5-da01748dab92', 'Золотая птичка')
(200.0, '80fcd762-aefc-4e1a-ac1e-b8884f2a00a1', 'Золотая птичка')
(200.0, '82be50db-670b-48c6-beb9-7e00d584f6de', 'Золотая птичка')
PICK a61fb41c-2448-43fd-b088-518690a5ab1b
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=<a61fb41c-2448-43fd-b088-518690a5ab1b>
bash: синтаксическая ошибка рядом с неожиданным маркером «newline»
[denismatveev@c0der event_horizon]$ export EH_K6_ITEM_ID=a61fb41c-2448-43fd-b088-518690a5ab1b
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
    ✗ 'rate<0.15' rate=96.15%

    http_req_duration
    ✓ 'p(95)<1200' p(95)=417.76ms

    http_req_failed
    ✗ 'rate<0.05' rate=92.59%


  █ TOTAL RESULTS 

    checks_total.......: 104    3.247319/s
    checks_succeeded...: 3.84%  4 out of 104
    checks_failed......: 96.15% 100 out of 104

    ✗ purchase 200
      ↳  2% — ✓ 1 / ✗ 49
    ✗ purchase success
      ↳  2% — ✓ 1 / ✗ 49
    ✗ cancel1 200
      ↳  0% — ✓ 0 / ✗ 1
    ✗ cancel1 success
      ↳  0% — ✓ 0 / ✗ 1
    ✓ cancel2 200
    ✓ cancel2 idempotent

    CUSTOM
    errors.........................: 96.15% 50 out of 52

    HTTP
    http_req_duration..............: avg=124.93ms min=5.71ms   med=22.92ms  max=2.2s  p(90)=243.31ms p(95)=417.76ms
      { expected_response:true }...: avg=575.65ms min=219.42ms med=448.34ms max=1.18s p(90)=1.01s    p(95)=1.1s    
    http_req_failed................: 92.59% 50 out of 54
    http_reqs......................: 54     1.686108/s

    EXECUTION
    iteration_duration.............: avg=603.74ms min=507.98ms med=524.64ms max=2.71s p(90)=624.02ms p(95)=755.77ms
    iterations.....................: 50     1.561211/s
    vus............................: 1      min=0        max=1
    vus_max........................: 1      min=1        max=1

    NETWORK
    data_received..................: 64 kB  2.0 kB/s
    data_sent......................: 30 kB  922 B/s




running (0m32.0s), 0/1 VUs, 50 complete and 0 interrupted iterations
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
    ✗ 'p(95)<800' p(95)=2.57s

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 264     8.237368/s
    checks_succeeded...: 100.00% 264 out of 264
    checks_failed......: 0.00%   0 out of 264

    ✓ shop items 200
    ✓ shop items is array
    ✓ inventory items 200
    ✓ inventory items has items

    CUSTOM
    errors.........................: 0.00%  0 out of 132

    HTTP
    http_req_duration..............: avg=1s    min=29.27ms med=682.05ms max=2.94s p(90)=2.32s p(95)=2.57s
      { expected_response:true }...: avg=1s    min=29.27ms med=682.05ms max=2.94s p(90)=2.32s p(95)=2.57s
    http_req_failed................: 0.00%  0 out of 133
    http_reqs......................: 133    4.149886/s

    EXECUTION
    iteration_duration.............: avg=2.33s min=1.32s   med=2.38s    max=3.38s p(90)=3.01s p(95)=3.22s
    iterations.....................: 66     2.059342/s
    vus............................: 1      min=1        max=5
    vus_max........................: 5      min=5        max=5

    NETWORK
    data_received..................: 5.0 MB 157 kB/s
    data_sent......................: 64 kB  2.0 kB/s




running (0m32.0s), 0/5 VUs, 66 complete and 0 interrupted iterations
default ✓ [======================================] 5 VUs  30s
ERRO[0032] thresholds on metrics 'http_req_duration' have been crossed 
[denismatveev@c0der event_horizon]$ curl -sS -o /dev/null -w 'shop %{http_code} %{time_total}s size=%{size_download}\n' \
  "$BASE_URL/api/shop/items" -H "Authorization: Bearer $TOKEN"
Paste the new end-of-run blocks. After a clean purchase (checks
shop 200 0.667890s size=52148
bash: синтаксическая ошибка рядом с неожиданным маркером «(»
[denismatveev@c0der event_horizon]$ 