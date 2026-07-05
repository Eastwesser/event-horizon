# 🛒 TODO — Shop Service (Магазин)

**Дата:** 06.07.2026  
**Цель:** Реализовать магазин за билетики (внутриигровая валюта)  
**Версия:** v1.0.5 (план)

---

## 📌 Бизнес-требования

- Пользователь может тратить билетики на:
  - **Кастомизацию игр** (радужные трубы в Flappy, скины для блинов, темы для профиля).
  - **Реальный мерч** (фенечки от художников) — за 20,000 билетиков (этап 2).
- Товары ротируются раз в месяц (без FOMO).
- Покупка оформляется через Billing (списание билетиков).
- После покупки — событие в NATS (`shop.purchased`).

---

## 🏗️ Архитектура (дизайн)

```text
┌─────────────────────────────────────────────────────────────┐
│                      SHOP SERVICE                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [Gateway] ──► gRPC ──► [Shop Service]                     │
│  GET /api/shop/items                                        │
│  POST /api/shop/purchase                                    │
│                                                             │
│  [Shop Service] ──► [Billing] (списание билетиков)         │
│  [Shop Service] ──► NATS (shop.purchased)                  │
│                                                             │
│  [Shop Service] ──► PostgreSQL (items, inventory, history) │
│  [Shop Service] ──► Redis (кеш товаров)                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
📂 Структура сервиса
text
services/shop/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── grpc_handler.go
│   ├── repository/
│   │   ├── postgres_repo.go
│   │   └── redis_repo.go
│   └── service/
│       └── shop_service.go
├── proto/
│   └── shop.proto
├── migrations/
│   └── 20260706000000_init_shop_schema.sql
├── Dockerfile
└── README.md
🔧 Задачи (по этапам)
Этап 1: Базовая структура
Создать папку services/shop

Написать shop.proto — gRPC методы:

GetItems — список товаров

PurchaseItem — покупка товара

GetInventory — инвентарь пользователя

Сгенерировать protobuf (protoc --go_out...)

Создать go.mod и подтянуть зависимости (go mod tidy)

Этап 2: База данных
Создать таблицы в PostgreSQL:

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
Написать миграцию для Shop БД

Накатить миграцию (goose up)

Этап 3: Репозитории
Реализовать postgres_repo.go:

GetItems(ctx, gameID, category) ([]Item, error)

GetItemByID(ctx, itemID) (*Item, error)

PurchaseItem(ctx, userID, itemID, price) error

GetUserInventory(ctx, userID) ([]Item, error)

CheckUserBalance(ctx, userID, amount) error (через Billing gRPC)

Реализовать redis_repo.go:

Кеш товаров (SetItems, GetItems) — TTL 5 минут

Этап 4: Сервис (бизнес-логика)
Реализовать shop_service.go:

GetItems(ctx, gameID, category) ([]Item, error) — сначала Redis, потом PostgreSQL

PurchaseItem(ctx, userID, itemID) error:

Проверить, есть ли товар в наличии (available=true)
Проверить, не купил ли пользователь его раньше
Проверить баланс через Billing (gRPC)
Списать билетики через Billing
Записать покупку в purchases
Добавить в инвентарь
Опубликовать событие в NATS (shop.purchased)
GetInventory(ctx, userID) ([]Item, error)

Этап 5: gRPC хендлер
Реализовать grpc_handler.go:

GetItems — вызывает shopService.GetItems

PurchaseItem — вызывает shopService.PurchaseItem

GetInventory — вызывает shopService.GetInventory

Этап 6: Основной файл (cmd/main.go)
Подключение к PostgreSQL

Подключение к Redis

Подключение к NATS (кластер)

Подключение к Billing (gRPC-клиент)

Инициализация репозиториев и сервиса

Запуск gRPC сервера

Graceful shutdown

Метрики (Prometheus)

Этап 7: Интеграция с Gateway
Добавить эндпоинты в gateway/cmd/main.go:

GET /api/shop/items?game_id=flappy&category=game_skin

POST /api/shop/purchase (JSON: item_id)

GET /api/shop/inventory

Создать gRPC-клиент для Shop в Gateway

Добавить авторизацию (JWT) для всех эндпоинтов

Этап 8: NATS события
Подписка на user.registered — выдать стартовый набор товаров (например, 1 бесплатная кастомизация)

Публикация shop.purchased — уведомить Analytics, Notification

Добавить subject в nats-hub:

go
Subjects: []string{
    "event.>",
    "score.updated",
    "user.registered",
    "shop.purchased",
}
Этап 9: Логика начисления билетиков (бонус)
При регистрации — начислить 100 билетиков (через Billing)

Ежедневный бонус — 10 билетиков за вход (опционально)

Этап 10: Антифрод (защита)
Ограничение — 1 фенечка в месяц на аккаунт

Привязка к бусти / email — для выкупа реального мерча

Мониторинг аномалий — > 1 покупка за месяц

Device ID + fingerprint — для предотвращения мультиаккаунтинга

Капча при достижении 10k билетиков

Лимит на количество аккаунтов с одного IP (max 3)

Этап 11: Frontend (React)
Страница магазина (список товаров)

Кнопка "Купить" с подтверждением

Отображение баланса (билетики)

Страница инвентаря (купленные товары)

Этап 12: Тестирование
Юнит-тесты для Shop Service

Интеграционные тесты (purchase flow)

Нагрузочное тестирование (k6)

E2E-тест (регистрация → игра → покупка кастомизации)

Этап 13: Документация
README.md для Shop Service

OpenAPI для эндпоинтов

Обновление главного README (добавить Shop в список сервисов)

📅 Приоритеты
Приоритет	Задача	Время
🔥 1	Структура + proto	1 час
🔥 2	БД + миграции	1 час
🔥 3	Репозитории	2 часа
🔥 4	Сервис (бизнес-логика)	3 часа
🔥 5	gRPC хендлер + main.go	2 часа
🔥 6	Интеграция с Gateway	2 часа
🔥 7	NATS события	1 час
🔥 8	Frontend (React)	3 часа
🔥 9	Тестирование	2 часа
Итого: ~17 часов (2–3 дня)

🚀 Следующие шаги после магазина
Notification Service — уведомления о покупках.

Analytics Service — DAU, MAU, популярность товаров.

Payment Service — реальные платежи (Boosty/Stripe).

📌 Заметки
Кастомизация: покупка скина для игры → нужно связать с фронтендом (например, сохранять выбор в Redis/БД).

Фенечки (мерч): реальный товар, требуется интеграция с художниками и логистика (этап 2).

Этика: без FOMO, без скрытого гринда.

Дата: 06.07.2026
Автор: Денис Матвеев (Eastwesser)
Статус: План утверждён