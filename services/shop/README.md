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

┌─────────────────────────────────────────────────────────────┐
│                     SHOP ECOSYSTEM                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Shop API   │    │   Catalog    │    │  Inventory   │  │
│  │   Gateway    │◄───│   Service    │    │   Service    │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                   │                   │           │
│         ▼                   ▼                   ▼           │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Purchase   │    │   Supplier   │    │  Analytics   │  │
│  │   Service    │    │   Service    │    │   Service    │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

🔗 Интеграции
Сервис	Протокол	Назначение
Gateway	gRPC	Приём запросов от фронтенда
Billing	gRPC	Проверка баланса и списание билетиков
NATS	Async	Публикация события shop.purchased
📚 gRPC методы
Метод	Описание
GetItems	Список товаров (с фильтром по категории/игре)
PurchaseItem	Покупка товара (списание билетиков)
GetInventory	Инвентарь пользователя
🗄️ База данных
Таблицы
sql
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
Начальные товары
sql
INSERT INTO items (name, description, price, category, game_id, image_url) VALUES
('Космические блины', 'Блины в стиле космос!', 100, 'game_skin', 'hexagon', '/images/space_pancakes.png'),
('Радужные блоки', 'Разноцветные блоки для башни', 100, 'game_skin', 'towers', '/images/rainbow_blocks.png'),
('Карточки со зверями', 'Животные вместо фруктов в Меморине', 150, 'game_skin', 'memory', '/images/animal_cards.png'),
('Радужные трубы', 'Сделайте трубы в Flappy радужными!', 100, 'game_skin', 'flappy', '/images/rainbow_pipes.png'),
('Золотая птичка', 'Птичка становится золотой!', 200, 'game_skin', 'flappy', '/images/golden_bird.png'),
('Блинный мерч', 'Футболка с блином', 50, 'merch', NULL, '/images/pancake_tshirt.png');
🧪 Тестирование
1. Настройка тестового пользователя
bash
# Получаем ID пользователя
docker exec -it event-horizon-postgres psql -U eventhorizon -d eventhorizon -c "SELECT id, email FROM users WHERE email = 'tuzer@example.com';"

# Добавляем баланс (10000 билетиков и 10000 лампочек)
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "
INSERT INTO user_currencies (user_id, currency_type, balance) 
VALUES ('7fc8a659-1bb2-4d7c-b60e-c140239d5c62', 'tickets', 10000)
ON CONFLICT (user_id, currency_type) 
DO UPDATE SET balance = 10000;

INSERT INTO user_currencies (user_id, currency_type, balance) 
VALUES ('7fc8a659-1bb2-4d7c-b60e-c140239d5c62', 'lamps', 10000)
ON CONFLICT (user_id, currency_type) 
DO UPDATE SET balance = 10000;
"

# Очищаем кеш Redis
docker exec -it event-horizon-redis-billing redis-cli FLUSHALL
2. Проверка API
bash
# Получаем токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"

# Проверяем баланс (должно быть 10000)
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Получаем список товаров
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Покупаем товар (Радужные трубы за 100 билетиков)
ITEM_ID="6a1de8dd-9457-4aa4-99a7-78267aee731d"
curl -s -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"item_id\":\"$ITEM_ID\"}" | jq '.'

# Проверяем инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Проверяем баланс после покупки (должно быть 9900)
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
3. Полный цикл тестирования
bash
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login -H "Content-Type: application/json" -d '{"email":"tuzer@example.com","password":"tuzer1"}' | jq -r '.access_token') && \
echo "=== 1. Баланс ДО покупки ===" && \
curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.' && \
echo "=== 2. Список товаров ===" && \
curl -s -X GET http://localhost:8079/api/shop/items -H "Authorization: Bearer $TOKEN" | jq '.[] | {id, name, price}' && \
echo "=== 3. Покупка ===" && \
curl -s -X POST http://localhost:8079/api/shop/purchase -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.' && \
echo "=== 4. Баланс ПОСЛЕ покупки ===" && \
curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.' && \
echo "=== 5. Инвентарь ===" && \
curl -s -X GET http://localhost:8079/api/shop/inventory -H "Authorization: Bearer $TOKEN" | jq '.'
🚀 Сборка и запуск
bash
cd /home/denismatveev/event_horizon/services/shop

# Сборка бинарника
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go

# Сборка Docker образа
cd /home/denismatveev/event_horizon
docker build -t eastwesser/shop:latest -f Dockerfile.shop.bin .

# Пуш в Docker Hub
docker push eastwesser/shop:latest

# Перезапуск
make deploy
🔧 Управление кешем
Очистка кеша товаров
bash
docker exec -it event-horizon-redis-shop redis-cli FLUSHALL
Очистка кеша баланса Billing
bash
docker exec -it event-horizon-redis-billing redis-cli FLUSHALL
🔥 Актуальные товары
Игра	Товар	Цена	Описание
Flappy Bird	Радужные трубы	100	Разноцветные трубы вместо зелёных
Flappy Bird	Золотая птичка	200	Птичка становится золотой
Hexagon	Космические блины	100	Блины в космическом стиле
Towers	Радужные блоки	100	Разноцветные блоки для башни
Memory	Карточки со зверями	150	Животные вместо фруктов
Общий	Блинный мерч	50	Футболка с блином
📌 Следующие шаги
Базовый функционал магазина

Интеграция с Billing

Инвентарь пользователя

Кеширование товаров

Инвалидация кеша баланса

Применение кастомизации в играх

Полка чудес (мерч от художников)

Дата: 13.07.2026
Версия: v1.0.6

text

## 4. Пересобираем и деплоим

```bash
cd /home/denismatveev/event_horizon/services/shop

# Обновляем зависимости
go mod tidy

# Собираем
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go

# Собираем Docker образ
cd /home/denismatveev/event_horizon
docker build -t eastwesser/shop:latest -f Dockerfile.shop.bin .

# Пушим
docker push eastwesser/shop:latest

# Перезапускаем
make deploy

# Ждем 10 секунд
sleep 10

# Тестируем
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login -H "Content-Type: application/json" -d '{"email":"tuzer@example.com","password":"tuzer1"}' | jq -r '.access_token') && \
echo "=== Баланс ДО ===" && curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.' && \
echo "=== Покупка ===" && curl -s -X POST http://localhost:8079/api/shop/purchase -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"item_id":"6a1de8dd-9457-4aa4-99a7-78267aee731d"}' | jq '.' && \
echo "=== Баланс ПОСЛЕ ===" && curl -s -X GET "http://localhost:8079/api/billing/balance/all" -H "Authorization: Bearer $TOKEN" | jq '.'
Теперь баланс должен обновляться сразу после покупки! 🚀

---
ОБНОВЛЕНО

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
🔗 Интеграции
Сервис	Протокол	Назначение
Gateway	gRPC	Приём запросов от фронтенда
Billing	gRPC	Проверка баланса и списание билетиков
NATS	Async	Публикация события shop.purchased
📚 gRPC методы
Метод	Описание
GetItems	Список товаров (с фильтром по категории/игре)
PurchaseItem	Покупка товара (списание билетиков)
GetInventory	Инвентарь пользователя
🗄️ База данных
Таблицы
sql
-- Товары
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price INT NOT NULL,
    category TEXT NOT NULL,
    game_id TEXT,
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
Текущие товары
ID	Название	Цена	Категория	Игра
6a1de8dd...	Радужные трубы	100	game_skin	flappy
82be50db...	Золотая птичка	200	game_skin	flappy
6c827e89...	Радужные трубы	100	game_skin	-
49b5b790...	Золотая птичка	200	game_skin	-
52673df7...	Космический фон	150	game_skin	-
b5b332fd...	Блинный мерч	50	merch	-
🧪 Тестирование
1. Настройка тестового пользователя
bash
# Добавляем баланс
docker exec -it event-horizon-postgres-billing psql -U eventhorizon -d eventhorizon_billing -c "
INSERT INTO user_currencies (user_id, currency_type, balance) 
VALUES ('7fc8a659-1bb2-4d7c-b60e-c140239d5c62', 'tickets', 10000)
ON CONFLICT (user_id, currency_type) 
DO UPDATE SET balance = 10000;
"

# Очищаем кеш Redis
docker exec -it event-horizon-redis-billing redis-cli FLUSHALL
2. Проверка API
bash
# Получаем токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

# Получаем список товаров
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Покупаем товар
ITEM_ID="6a1de8dd-9457-4aa4-99a7-78267aee731d"
curl -s -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"item_id\":\"$ITEM_ID\"}" | jq '.'

# Проверяем инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'
🚀 Сборка и запуск
bash
cd /home/denismatveev/event_horizon/services/shop

# Сборка бинарника
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go

# Сборка Docker образа
cd /home/denismatveev/event_horizon
docker build -t eastwesser/shop:latest -f Dockerfile.shop.bin .

# Пуш в Docker Hub
docker push eastwesser/shop:latest

# Перезапуск
make deploy
🔧 Управление кешем
bash
# Очистка кеша товаров
docker exec -it event-horizon-redis-shop redis-cli FLUSHALL

# Очистка кеша баланса
docker exec -it event-horizon-redis-billing redis-cli FLUSHALL
🎮 Интеграция с играми
Flappy Bird
Радужные трубы - меняет цвет труб на радужный

Золотая птичка - меняет цвет птички на золотой

Hexagon
Космические блины - меняет текстуру блинов

Towers
Радужные блоки - меняет цвет блоков

Memory
Карточки со зверями - меняет картинки на карточках

📌 Следующие шаги

Базовый функционал магазина
Интеграция с Billing
Инвентарь пользователя
Кеширование товаров
Инвалидация кеша баланса
Фронтенд магазина
Применение скинов в играх
Добавление реальных картинок
Мерч для профиля

Дата: 13.07.2026
Версия: v1.0.6
Статус: 🟢 Работает

--

cd /home/denismatveev/event_horizon/services/shop

# 1. Генерируем protobuf
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/shop.proto

# 2. Проверяем что сгенерировалось
ls -la proto/*.pb.go

# 3. Скачиваем зависимости
go mod tidy

# 4. Собираем бинарник
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go

# 5. Проверяем бинарник
ls -la shop-service

# 6. Собираем Docker образ
cd /home/denismatveev/event_horizon
docker build -t eastwesser/shop:latest -f Dockerfile.shop.bin .

# 7. Пушим в Docker Hub
docker push eastwesser/shop:latest

# 8. Деплоим
make deploy

# 9. Проверяем логи
docker logs deployments-shop-1 --tail=20

# 10. Проверяем инвентарь с датой
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login -H "Content-Type: application/json" -d '{"email":"tuzer@example.com","password":"tuzer1"}' | jq -r '.access_token')
curl -s -X GET http://localhost:8079/api/shop/inventory -H "Authorization: Bearer $TOKEN" | jq '.'

Чек 2 августа 2026:

cd /home/denismatveev/event_horizon

# 1. Перегенерировать proto (shop)
cd services/shop
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/*.proto
echo "✅ Shop proto regenerated"

# 2. Собрать бинарник
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shop-service ./cmd/main.go
echo "✅ Shop binary built"

# 3. Собрать Docker образ
cd /home/denismatveev/event_horizon
docker build -t eastwesser/shop:latest -f Dockerfile.shop.bin .
echo "✅ Shop Docker image built"

# 4. Запустить
make deploy
echo "✅ Shop deployed"

# 5. Проверить логи
docker logs deployments-shop-1 --tail=20
🧪 Проверка Shop
bash
# 1. Получить токен
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tuzer@example.com","password":"tuzer1"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"

# 2. Проверить баланс ДО покупки
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. Список товаров
curl -s -X GET http://localhost:8079/api/shop/items \
  -H "Authorization: Bearer $TOKEN" | jq '.[] | {id, name, price}'

# 4. Купить товар (используй ID из списка)
ITEM_ID="6a1de8dd-9457-4aa4-99a7-78267aee731d"
curl -s -X POST http://localhost:8079/api/shop/purchase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"item_id\":\"$ITEM_ID\"}" | jq '.'

# 5. Проверить баланс ПОСЛЕ покупки
curl -s -X GET "http://localhost:8079/api/billing/balance/all" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 6. Проверить инвентарь
curl -s -X GET http://localhost:8079/api/shop/inventory \
  -H "Authorization: Bearer $TOKEN" | jq '.'
Ожидаемый результат:

Баланс уменьшился на цену товара.

Товар появился в инвентаре.

В логах Shop видно ✅ Shop item created from inventory (если покупаешь мерч).