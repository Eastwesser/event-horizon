🚀 Сборка и деплой Gateway
bash
cd /home/denismatveev/event_horizon

# 1. Обновить зависимости Gateway
cd services/gateway
go mod tidy

# 2. Собрать бинарник
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gateway-service ./cmd/main.go

# 3. Собрать Docker образ
cd /home/denismatveev/event_horizon
docker build -t eastwesser/gateway:latest -f Dockerfile.gateway.bin .

# 4. Запушить в Docker Hub
docker push eastwesser/gateway:latest

# 5. Перезапустить
make deploy

# 6. Проверить логи
docker logs deployments-gateway-1 --tail=30
✅ Проверка новых эндпоинтов Inventory
bash
# Получить токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# 1. Создать товар
curl -X POST http://localhost:8079/api/inventory/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "брелок",
    "name": "Тестовый брелок",
    "description": "Красивый брелок из серебра",
    "price": 100.50,
    "stock": 10,
    "attributes": {"material": "серебро", "weight": 15},
    "images": ["/images/keychain1.png"]
  }' | jq '.'

# 2. Получить список товаров
curl -X GET "http://localhost:8079/api/inventory/items?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. Получить товар по ID
ITEM_ID="..." # из предыдущего ответа
curl -X GET "http://localhost:8079/api/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 4. Обновить товар
curl -X PUT "http://localhost:8079/api/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Обновлённый брелок",
    "price": 150.00
  }' | jq '.'

# 5. Удалить товар
curl -X DELETE "http://localhost:8079/api/inventory/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'