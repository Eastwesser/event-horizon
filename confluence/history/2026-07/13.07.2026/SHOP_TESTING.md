5. Проверяем интеграцию Shop с Billing
Проверяем, что Shop видит Billing:

bash
# Проверяем, что биллинг доступен из Shop контейнера
docker exec deployments-shop-1 sh -c "nc -zv billing 50053"
6. Тестируем Shop API через curl
bash
# 1. Логинимся и получаем токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"

# 2. Получаем список товаров
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 3. Проверяем баланс
curl -s -X GET "http://localhost:8079/api/billing/balance/all?user_id=test-user" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

# 4. Покупаем товар (замените item_id на реальный)
curl -s -X POST http://localhost:8079/api/shop/buy \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"item_id":"<item_id>"}' \
  | jq '.'

# 5. Проверяем инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
7. Если Shop не работает - проверяем его логи
bash
docker logs deployments-shop-1

Проверяем миграции Shop
bash
# Проверяем, что таблицы созданы
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "\dt"

# Проверяем, есть ли товары
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "SELECT * FROM items;"

10. Если товаров нет - добавляем тестовые
bash
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "
INSERT INTO items (name, description, price, category, image_url, available) VALUES
('Радужные трубы', 'Сделайте трубы в Flappy радужными!', 100, 'game_skin', '/images/rainbow_pipes.png', true),
('Золотая птичка', 'Птичка становится золотой!', 200, 'game_skin', '/images/golden_bird.png', true),
('Космический фон', 'Космический фон для любой игры', 150, 'game_skin', '/images/cosmic_bg.png', true),
('Блинный мерч', 'Футболка с блином', 50, 'merch', '/images/pancake_tshirt.png', true);
"

