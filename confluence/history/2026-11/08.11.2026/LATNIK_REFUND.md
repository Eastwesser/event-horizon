Задача: проверить, что админ может отменить покупку Раисы.

Проблема: у Феи и Совы уже стоит fulfilled_at (Fulfillment их «собрал»), поэтому RefundPurchase вернёт 400 «Товар уже отправлен». Это ожидаемо (мы записали в TODO: «админ должен обходить fulfilled_at» — но это следующий патч).

Чтобы проверить AdminCancelPurchase сейчас, нужен товар без fulfilled_at. То есть:

Раиса покупает новый товар.

Сразу (до того, как Fulfillment через 10 секунд поставит fulfilled_at) админ отменяет покупку.

Шаги
1. Зайди под Раисой и купи любой товар (например, «Латник Дахора», b29bafcc-...).

Или через curl:

bash
RAISA_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"raisa.matveeva@event-horizon.com","password":"R@i$04ka"}' | jq -r '.access_token')

# Купить Латника
curl -i -s -X POST "http://localhost:8079/api/v1/shop/purchase" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $RAISA_TOKEN" \
  -d '{"item_id":"b29bafcc-3a7f-4766-8301-9b58bb19b498"}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
2. Сразу же (в течение 10 секунд) админ отменяет покупку:

bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')

curl -i -s -X POST "http://localhost:8079/api/v1/admin/shop/purchases/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"user_id":"d1bc5e60-227c-412e-80cf-f7e5de4c6b11","item_id":"b29bafcc-3a7f-4766-8301-9b58bb19b498"}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
Что ожидать:

200 OK — админ отменил покупку. ✅ AdminCancelPurchase работает.

400 «Товар уже отправлен» — Fulfillment успел поставить fulfilled_at (ты не успел за 10 секунд). Повтори быстрее или возьми товар, у которого fulfilled_at ещё не стоит.

404 — покупка не найдена (Раиса не купила этот товар).

403 — ты не админ (но ты админ).

3. Проверь БД:

bash
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c \
  "SELECT id, item_id, status FROM purchases WHERE user_id = 'd1bc5e60-227c-412e-80cf-f7e5de4c6b11' ORDER BY purchased_at DESC LIMIT 3;"
Статус Латника должен стать REFUNDED.

[denismatveev@c0der event_horizon]$ RAISA_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"raisa.matveeva@event-horizon.com","password":"R@i$04ka"}' | jq -r '.access_token')
[denismatveev@c0der event_horizon]$ curl -i -s -X POST "http://localhost:8079/api/v1/shop/purchase" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $RAISA_TOKEN" \
  -d '{"item_id":"b29bafcc-3a7f-4766-8301-9b58bb19b498"}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
HTTP/1.1 200 OK
Content-Length: 89
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'
Content-Type: application/json; charset=utf-8
Date: Sun, 08 Nov 2026 17:04:49 GMT
Permissions-Policy: geolocation=(), microphone=(), camera=()
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Permitted-Cross-Domain-Policies: none

{"message":"Purchase successful","new_balance":956282,"remaining_stock":0,"success":true}
---HTTP_CODE:200---
[denismatveev@c0der event_horizon]$ ADMIN_TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')
[denismatveev@c0der event_horizon]$ 
[denismatveev@c0der event_horizon]$ curl -i -s -X POST "http://localhost:8079/api/v1/admin/shop/purchases/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"user_id":"d1bc5e60-227c-412e-80cf-f7e5de4c6b11","item_id":"b29bafcc-3a7f-4766-8301-9b58bb19b498"}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'curl -i -s -X POST "http://localhost:8079/api/v1/admin/shop/purchases/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"user_id":"d1bc5e60-227c-412e-80cf-f7e5de4c6b11","item_id":"b29bafcc-3a7f-4766-8301-9b58bb19b498"}' \
  -w '\n---HTTP_CODE:%{http_code}---\n'
HTTP/1.1 200 OK
Content-Length: 126
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'
Content-Type: application/json; charset=utf-8
Date: Sun, 08 Nov 2026 17:05:17 GMT
Permissions-Policy: geolocation=(), microphone=(), camera=()
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Permitted-Cross-Domain-Policies: none

{"already_refunded":false,"message":"Purchase cancelled by admin","new_balance":966282,"refunded_amount":10000,"success":true}
---HTTP_CODE:200---
HTTP/1.1 200 OK
Content-Length: 110
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'
Content-Type: application/json; charset=utf-8
Date: Sun, 08 Nov 2026 17:05:20 GMT
Permissions-Policy: geolocation=(), microphone=(), camera=()
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-Permitted-Cross-Domain-Policies: none

{"already_refunded":true,"message":"Already refunded","new_balance":966282,"refunded_amount":0,"success":true}
---HTTP_CODE:200---
[denismatveev@c0der event_horizon]$ docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c \
  "SELECT id, item_id, status FROM purchases WHERE user_id = 'd1bc5e60-227c-412e-80cf-f7e5de4c6b11' ORDER BY purchased_at DESC LIMIT 3;"
                  id                  |               item_id                |  status  
--------------------------------------+--------------------------------------+----------
 ca1316b4-6e48-4dda-9cba-a6049ed48d55 | b29bafcc-3a7f-4766-8301-9b58bb19b498 | REFUNDED
 f9423ca3-6e8e-4857-ac26-16d858579e1d | 40c5c1c5-0973-41e4-9112-4bf0d43582a8 | REFUNDED
 1d0871c3-4ec3-40df-8e9d-38bfccda4065 | b29bafcc-3a7f-4766-8301-9b58bb19b498 | REFUNDED
(3 rows)

[denismatveev@c0der event_horizon]$ 