# Inventory Service

Сервис управления каталогом товаров авторов (художников).

## 🚀 Быстрый старт

```bash
# Установить зависимости
go mod download

# Сгенерировать protobuf (если меняли .proto)
make proto

# Запустить с PostgreSQL
export INVENTORY_DRIVER=postgres
go run cmd/main.go

# Запустить с MongoDB
export INVENTORY_DRIVER=mongo
go run cmd/main.go
🐳 Docker
bash
docker build -t inventory-service .
docker run -p 50055:50055 inventory-service
📦 Конфигурация
Переменная	Описание	По умолчанию
INVENTORY_GRPC_PORT	Порт gRPC сервера	50055
INVENTORY_PG_HOST	Хост PostgreSQL	localhost
INVENTORY_PG_PORT	Порт PostgreSQL	5465
INVENTORY_PG_USER	Пользователь PostgreSQL	postgres
INVENTORY_PG_PASSWORD	Пароль PostgreSQL	postgres
INVENTORY_PG_DB	База данных	inventory
INVENTORY_MONGO_URI	URI MongoDB	mongodb://localhost:27017
INVENTORY_MONGO_DB	Имя БД в MongoDB	inventory
INVENTORY_DRIVER	Драйвер: postgres или mongo	postgres
📚 API
Сервис предоставляет gRPC API (см. proto/inventory.proto):

CreateItem — создание товара

GetItem — получение товара по ID

UpdateItem — обновление товара

DeleteItem — удаление товара

SearchItems — поиск с фильтрами

GetByAuthor — товары автора

GetByType — товары по типу

--

Как работает Outbox + NATS?
Ты правильно понял! Схема такая:

text
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Сервис    │────►│   Outbox    │────►│    NATS     │
│  (Inventory)│     │   (PG)      │     │  (JetStream)│
└─────────────┘     └─────────────┘     └─────────────┘
                           │                    │
                           │                    ▼
                           │              ┌─────────────┐
                           └──────────────│  Подписчик  │
                                          │   (Shop)    │
                                          └─────────────┘
Как это работает:

Сервис (Inventory) сохраняет данные и событие в одной транзакции в таблицу outbox.

Отдельный воркер (в том же сервисе) читает из outbox и публикует в NATS.

Если NATS упал — событие остаётся в outbox. Воркер будет повторять попытки.

После успешной публикации — помечает событие как processed.

Преимущество: Ты никогда не потеряешь событие, даже если NATS временно недоступен.


----
