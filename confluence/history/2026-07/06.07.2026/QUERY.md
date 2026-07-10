🔥 Приоритет 1: Backend (Shop Service)
Это основа, без неё фронтенд не взлетит.

Создать структуру services/shop

Написать shop.proto:

GetItems (список товаров)

PurchaseItem (покупка)

GetInventory (инвентарь)

База данных:

Таблицы: items, inventory, purchases

Миграция через goose

Репозитории:

postgres_repo.go — CRUD для товаров, инвентаря, покупок

redis_repo.go — кеш товаров (TTL 5 минут)

Сервис:

GetItems — сначала Redis, потом PG

PurchaseItem:

Проверить наличие товара
Проверить, не куплен ли уже (для merch — 1 раз в месяц)
Связаться с Billing (gRPC) для проверки баланса и списания
Записать покупку в БД
Опубликовать событие shop.purchased в NATS
GetInventory — список купленных товаров пользователя

gRPC хендлер + cmd/main.go:

Подключение к PostgreSQL, Redis, NATS

gRPC-клиент для Billing

Интеграция с Gateway:

Добавить эндпоинты:

GET /api/shop/items

POST /api/shop/purchase

GET /api/shop/inventory

Авторизация через JWT

1. Магазин (Frontend) — приоритет 🔥🔥🔥
1.1. Создать страницу магазина
Создать frontend/src/pages/Shop.tsx (или frontend/src/components/Shop/Shop.tsx).

Добавить роут в App.tsx (/shop).

Добавить ссылку в меню (Header/Navigation).

1.2. Компоненты магазина
Карточка товара (ShopItemCard.tsx):

Иконка (изображение)

Название

Описание

Цена в билетиках

Кнопка "Купить" (с подтверждением)

Статус ("Доступно" / "Куплено")

Список товаров (с фильтром по категориям):

Вкладки: "Все", "Кастомизация", "Мерч"

Инвентарь пользователя (отдельный раздел на странице):

Список купленных товаров

Возможность применить кастомизацию (если применимо)

1.3. Интеграция с API (бэкенд)
Создать эндпоинты в frontend/src/services/api.ts:

getShopItems(category?)

purchaseItem(itemId)

getInventory()

Добавить JWT-авторизацию для всех запросов (уже есть в api.ts).

1.4. Стилизация
Адаптировать CSS из существующих компонентов (тёмная тема, акценты).

Сетка товаров (grid, адаптивность).

Кнопки и карточки — единый стиль с проектом.

🔥 Приоритет 1: Магазин (Backend)
1.1. Базовый сервис Shop
Создать структуру services/shop/ (cmd, internal, proto, migrations)

Написать shop.proto с методами:

GetItems — список товаров (с фильтром по категории/игре)

PurchaseItem — покупка товара

GetInventory — инвентарь пользователя

Сгенерировать protobuf (protoc --go_out...)

Создать go.mod и подтянуть зависимости

1.2. База данных
Создать таблицы в PostgreSQL (items, inventory, purchases)

Написать миграцию и накатить (goose up)

1.3. Репозитории
postgres_repo.go:

GetItems(ctx, gameID, category)

GetItemByID(ctx, itemID)

PurchaseItem(ctx, userID, itemID, price)

GetUserInventory(ctx, userID)

CheckUserBalance(ctx, userID, amount) — через gRPC к Billing

redis_repo.go: кеш товаров (TTL 5 минут)

1.4. Бизнес-логика
shop_service.go:

GetItems — сначала Redis, потом PostgreSQL

PurchaseItem:

Проверить доступность товара
Проверить, не куплен ли уже
Проверить баланс через Billing
Списать билетики
Записать покупку
Добавить в инвентарь
Опубликовать NATS shop.purchased
GetInventory

1.5. gRPC handler + main.go
grpc_handler.go — вызовы сервиса

cmd/main.go:

Подключение к PostgreSQL, Redis, NATS, Billing (gRPC)

Запуск gRPC сервера

Graceful shutdown

Метрики (Prometheus)

1.6. Интеграция с Gateway
Добавить эндпоинты в gateway/cmd/main.go:

GET /api/shop/items?game_id=flappy&category=game_skin

POST /api/shop/purchase

GET /api/shop/inventory

Создать gRPC-клиент для Shop в Gateway

Добавить JWT-авторизацию

1.7. NATS
Добавить shop.purchased в nats-hub

Подписка на user.registered — выдать стартовый бонус (например, 100 билетиков)


Давай ограничимся сейчас написанием сервиса на бэкэнде для магазина.