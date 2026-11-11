# Ждём, пока Inventory поднимется
until curl -fsS http://localhost:9096/ready; do sleep 1; done

# Warmup: 10 запросов SearchItems
TOKEN=$(curl -s -X POST http://localhost:8079/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@eventhorizon.local","password":"changeme-dev-admin"}' | jq -r '.access_token')

for i in $(seq 1 10); do
  curl -s -o /dev/null -H "Authorization: Bearer $TOKEN" \
    http://localhost:8079/api/v1/inventory/items
  curl -s -o /dev/null -H "Authorization: Bearer $TOKEN" \
    http://localhost:8079/api/v1/shop/items
done