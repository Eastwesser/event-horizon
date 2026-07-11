🧪 Курлы после деплоя

1. Получить токен (для авторизации)
```bash
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"
```

2. Получить список товаров
```bash
curl -X GET "http://localhost:8079/api/shop/items?category=all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

3. Купить товар (например, радужные трубы)
```bash
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id": "UUID_ТОВАРА"}' | jq '.'
```

Примечание: item_id возьми из ответа на предыдущий запрос (/api/shop/items).

4. Посмотреть инвентарь
```bash
curl -X GET "http://localhost:8079/api/shop/inventory" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

✅ Ожидаемые ответы
Эндпоинт	Ожидаемый статус
/api/shop/items	200 OK + список товаров
/api/shop/purchase	200 OK + {"success": true, "new_balance": 99}
/api/shop/inventory	200 OK + список купленных товаров


---

# 1. Получить токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# 2. Получить список товаров
curl -X GET "http://localhost:8079/api/shop/items?category=all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. Купить товар (если есть)
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id": "UUID_ТОВАРА"}' | jq '.'

# 4. Проверить инвентарь
curl -X GET "http://localhost:8079/api/shop/inventory" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# ---------------------------------------------------------------

# 5. 💰 Дать тестовому пользователю 1000 билетиков

docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "
INSERT INTO user_currencies (user_id, currency_type, balance, updated_at)
VALUES ('a11d3767-a218-4281-b0c7-b9ce4c811080', 'tickets', 1000, NOW())
ON CONFLICT (user_id, currency_type) DO UPDATE
SET balance = 1000, updated_at = NOW();
"

---

🧪 Тестовый цикл (копируй и вставляй по порядку)
bash
# 1. Получить свежий токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"

# 2. Проверить баланс (должен быть 1000)
curl -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. Купить радужные трубы
ITEM_ID="6a1de8dd-9457-4aa4-99a7-78267aee731d"
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"item_id\": \"$ITEM_ID\"}" | jq '.'

# 4. Проверить инвентарь
curl -X GET "http://localhost:8079/api/shop/inventory" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 5. Проверить баланс после покупки
curl -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
