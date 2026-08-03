# 📦 Inventory Service

Сервис управления каталогом товаров авторов (художников). Предоставляет gRPC API для CRUD-операций с товарами, поддерживает динамические атрибуты и интеграцию с NATS через Outbox-паттерн.

---

## 🏗️ Архитектура

Inventory Service — это **отдельный микросервис** со своей собственной логикой и хранилищем.
## Компоненты Inventory Service

┌─────────────────────────────────────────────────────────────────────────┐
│                           INVENTORY SERVICE                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────────┐  │
│  │  gRPC API   │    │  Сервисный  │    │        Repository           │  │
│  │   :50059    │◄───│    слой     │───►│  (PostgreSQL / MongoDB)     │  │
│  └─────────────┘    └─────────────┘    └─────────────────────────────┘  │
│         │                  │                        │                   │
│         │                  ▼                        ▼                   │
│         │         ┌─────────────┐    ┌─────────────────────────────┐    │
│         │         │    Redis    │    │           Outbox            │    │
│         │         │   (кеш)     │    │     (таблица в PG)          │    │
│         │         └─────────────┘    └─────────────────────────────┘    │
│         │                                    │                          │
│         │                                    ▼                          │
│         │                            ┌─────────────┐                    │
│         └────────────────────────────│    NATS     │                    │
│                                      │ (JetStream) │                    │
│                                      └─────────────┘                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘


### Компоненты

| Компонент         | Назначение                                | Технология                     |
| :---------------- | :---------------------------------------- | :----------------------------- |
| **gRPC API**      | Приём запросов от Gateway                 | gRPC, порт 50059               |
| **Сервисный слой**| Бизнес-логика, валидация                  | Go                             |
| **Repository**    | Абстракция доступа к данным               | Интерфейс + адаптеры           |
| **PostgreSQL**    | Основное хранилище товаров                | Таблицы: `inventory_items`, `outbox` |
| **MongoDB**       | Альтернативное хранилище (опционально)    | Коллекция: `inventory_items`   |
| **Redis**         | Кеширование товаров                       | TTL 5 минут                    |
| **Outbox**        | Надёжная доставка событий                 | Таблица в PostgreSQL           |
| **NATS JetStream**| Асинхронные события                       | Публикация: `inventory.item.created` |

---

## 🗄️ Базы данных

### PostgreSQL (основная)

Inventory Service использует **отдельную базу данных** `eventhorizon_shop` (разделяет с Shop Service, но имеет свои таблицы).

**Таблицы:**
- `inventory_items` — основная таблица товаров с JSONB-атрибутами
- `outbox` — таблица для Outbox-паттерна
- `goose_db_version` — служебная таблица миграций

**Индексы:**
- GIN-индекс на `attributes` (JSONB) для быстрого поиска по динамическим полям
- Индекс на `author_id` и `type` для фильтрации
- Full-text search по `name` (русский язык)

### MongoDB (альтернативная, тренировочная)

Реализован **адаптер** для MongoDB (`mongo_repo.go`), который позволяет переключиться через переменную окружения `INVENTORY_DRIVER=mongo`.

**Для чего:**
1. Тренировка работы с MongoDB в Go
2. Гибкий выбор БД под разные сценарии
3. Демонстрация паттерна "Адаптер"

> **Важно:** В продакшене используется PostgreSQL. MongoDB — опциональная альтернатива.

### Redis (кеш)

- **Ключи:** `inventory:item:{id}`
- **TTL:** 5 минут

---

## 📤 Outbox + NATS (надёжная доставка событий)

### Проблема

При создании товара нужно:
1. Сохранить товар в БД ✅
2. Оповестить другие сервисы (Shop) через NATS ❓

Если NATS недоступен — событие потеряется.

### Решение: Transactional Outbox

┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Сервис     │───►│   Outbox     │───►│    NATS      │
│ (Inventory)  │    │    (PG)      │    │ (JetStream)  │
└──────────────┘    └──────────────┘    └──────────────┘
       │                                      │
       │                                      ▼
       │                             ┌──────────────┐
       └────────────────────────────►│  Подписчик   │
                                     │    (Shop)    │
                                     └──────────────┘

### Как работает

1. Сервис сохраняет товар и событие в **одной транзакции** в таблицу `outbox`
2. **Outbox Worker** (отдельная горутина) читает необработанные события
3. Если NATS упал — событие остаётся в `outbox`, воркер повторяет попытки
4. После успешной публикации — помечает событие как `processed`
5. **Shop Service** подписан на `inventory.item.created` и создаёт товар у себя

### Преимущество

Ты никогда не потеряешь событие, даже если NATS временно недоступен.

### Схема outbox

```sql
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,   -- 'inventory.item.created'
    payload JSONB NOT NULL,              -- данные события
    created_at TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP
);
```

🔌 Интеграции
```text
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Gateway    │───►│  Inventory   │───►│    NATS      │
│              │    │   (gRPC)     │    │   (Outbox)   │
└──────────────┘    └──────────────┘    └──────────────┘
                                                  │
                                                  ▼
                                         ┌──────────────┐
                                         │     Shop     │
                                         │  (Consumer)  │
                                         └──────────────┘
Интеграции (таблица):

Компонент	Назначение	Протокол
PostgreSQL	Основное хранилище	SQL
MongoDB	Альтернативное хранилище (опционально)	BSON
Redis	Кеширование товаров	Redis Protocol
NATS JetStream	Асинхронные события	NATS Protocol
Gateway	Входная точка для клиентов	gRPC


gRPC API (таблица):

Метод	Описание
CreateItem	Создание товара (с сохранением в outbox)
GetItem	Получение товара по ID
UpdateItem	Обновление товара
DeleteItem	Удаление товара
SearchItems	Поиск с фильтрами (JSONB-атрибуты)
GetByAuthor	Получить все товары автора
GetByType	Получить товары по типу
```

📊 Метрики и мониторинг
Health check:

bash
curl http://localhost:9096/health
Ответ:

json
{
  "status": "ok",
  "service": "inventory",
  "driver": "postgres"
}
Prometheus метрики:

bash
curl http://localhost:9096/metrics
🔧 Конфигурация
Переменная	Описание	По умолчанию
INVENTORY_GRPC_PORT	Порт gRPC сервера	50059
INVENTORY_METRICS_PORT	Порт для метрик и health check	9096
INVENTORY_DRIVER	Драйвер: postgres или mongo	postgres
INVENTORY_PG_HOST	Хост PostgreSQL	localhost
INVENTORY_PG_PORT	Порт PostgreSQL	5465
INVENTORY_PG_USER	Пользователь PostgreSQL	eventhorizon
INVENTORY_PG_PASSWORD	Пароль PostgreSQL	eventhorizon
INVENTORY_PG_DB	База данных	eventhorizon_shop
INVENTORY_REDIS_ADDR	Адрес Redis	localhost:6379
INVENTORY_MONGO_URI	URI MongoDB	mongodb://localhost:27017
INVENTORY_MONGO_DB	Имя БД в MongoDB	inventory
NATS_URL	URL NATS кластера	nats://localhost:4222

🚀 Быстрый старт
bash
# Установить зависимости
go mod download

# Сгенерировать protobuf
make proto

# Запустить с PostgreSQL (рекомендуется)
export INVENTORY_DRIVER=postgres
go run cmd/main.go

# Запустить с MongoDB (тренировочный режим)
export INVENTORY_DRIVER=mongo
go run cmd/main.go

🐳 Docker
```bash
docker build -t eastwesser/inventory:latest -f Dockerfile.inventory.bin .
docker run -p 50059:50059 -p 9096:9099 eastwesser/inventory:latest
```

📌 Следующие шаги
☑ Базовая CRUD-функциональность
☑ Поддержка PostgreSQL и MongoDB
☑ Outbox + NATS интеграция
☑ Redis кеширование
☑ Метрики и health check
☑ Миграции через Goose
□ Интеграционные тесты
□ Документация OpenAPI
🔗 Связанные сервисы
Сервис	Роль	Интеграция
Shop	Потребитель событий (создаёт товары)	NATS (подписка)
Billing	Баланс авторов (не используется напрямую)	—
Gateway	Входная точка для клиентов	gRPC