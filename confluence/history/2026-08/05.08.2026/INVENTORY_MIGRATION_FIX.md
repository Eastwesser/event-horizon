✅ Проверка после деплоя:
bash
# 1. Проверить Inventory
curl http://localhost:9096/health | jq .

# 2. Проверить метрики
curl http://localhost:9096/metrics | head -5

# 3. Проверить Prometheus (должен быть UP)
# Открыть http://localhost:9090/targets

# 4. Проверить логи Inventory
docker logs deployments-inventory-1 --tail=20

# 5. Создать товар
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

curl -X POST http://localhost:8079/api/inventory/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"брелок","name":"Брелок с медведем","price":150}' | jq '.'
