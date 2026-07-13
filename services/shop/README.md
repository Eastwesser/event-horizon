# 🛒 Shop Service — Магазин за билетики

Микросервис для покупки кастомизации и мерча за внутриигровую валюту (билетики).

---

## 📦 Архитектура

```text
[Shop :50055] — gRPC
    │
    ├── PostgreSQL (товары, инвентарь, история покупок)
    ├── Redis (кеш товаров, TTL 5 минут)
    └── NATS (публикация shop.purchased)
```

## 🔗 Интеграции
Сервис	Протокол	Назначение
Gateway	gRPC	Приём запросов от фронтенда
Billing	gRPC	Проверка баланса и списание билетиков
NATS	Async	Публикация события shop.purchased

## 📚 gRPC методы
Метод	Описание
GetItems	Список товаров (с фильтром по категории/игре)
PurchaseItem	Покупка товара (списание билетиков)
GetInventory	Инвентарь пользователя

## 🗄️ База данных
```sql
-- Товары
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price INT NOT NULL,              -- стоимость в билетиках
    category TEXT NOT NULL,          -- 'game_skin', 'merch', 'profile_theme'
    game_id TEXT,                    -- 'flappy', 'hexagon', NULL для мерча
    image_url TEXT,
    available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Инвентарь пользователя
CREATE TABLE inventory (
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    purchased_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, item_id)
);

-- История покупок
CREATE TABLE purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    price INT NOT NULL,
    currency_type TEXT DEFAULT 'tickets',
    purchased_at TIMESTAMP DEFAULT NOW()
);
```

## 🧪 Тестирование
Получить список товаров
```bash
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' \
  | jq -r '.access_token')

curl -X GET "http://localhost:8079/api/shop/items?category=all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

Купить товар
```bash
ITEM_ID="6a1de8dd-9457-4aa4-99a7-78267aee731d"
curl -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"item_id\": \"$ITEM_ID\"}" | jq '.'
```

Проверить инвентарь
```bash
curl -X GET "http://localhost:8079/api/shop/inventory" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

🔥 Товары (актуальные)
Игра	Товар	Цена	Описание
Flappy Bird	Радужные трубы	100	Разноцветные трубы вместо зелёных
Flappy Bird	Золотая птичка	200	Птичка становится золотой
Hexagon	Космические блины	100	Блины в космическом стиле
Towers	Радужные блоки	100	Разноцветные блоки для башни
Memory	Карточки со зверями	150	Животные вместо фруктов

🚀 Сборка и запуск
```bash
cd ~/event_horizon/services/shop
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go
cd ~/event_horizon
docker build -f Dockerfile.shop.bin -t eastwesser/shop:latest .
docker-compose -f deployments/docker-compose.cluster.yml up -d shop
```

## 🎮 **Добавим товары для всех игр**

```bash
docker exec -it event-horizon-postgres-shop psql -U eventhorizon -d eventhorizon_shop -c "
INSERT INTO items (name, description, price, category, game_id, image_url) VALUES
('Космические блины', 'Блины в стиле космос!', 100, 'game_skin', 'hexagon', '/images/space_pancakes.png'),
('Радужные блоки', 'Разноцветные блоки для башни', 100, 'game_skin', 'towers', '/images/rainbow_blocks.png'),
('Карточки со зверями', 'Животные вместо фруктов в Меморине', 150, 'game_skin', 'memory', '/images/animal_cards.png'),
('Радужные трубы', 'Сделайте трубы в Flappy радужными!', 100, 'game_skin', 'flappy', '/images/rainbow_pipes.png'),
('Золотая птичка', 'Птичка становится золотой!', 200, 'game_skin', 'flappy', '/images/golden_bird.png');
"
```


📌 Следующие шаги

Добавить товары для всех игр (Hexagon, Towers, Memory).

Фронтенд — страница магазина и инвентаря.

Применение кастомизации — сохранять выбор пользователя в Redis/БД.

Полка чудес — мерч от художников (этап 2).

Дата: 11.07.2026
Версия: v1.0.5


git add .
git commit -m "feat: add shop items for all games & update README"
git push origin main