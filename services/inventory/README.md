# 📦 Inventory Service

Сервис управления каталогом товаров авторов (художников). Предоставляет gRPC API для CRUD-операций с товарами, поддерживает динамические атрибуты и интеграцию с NATS через Outbox-паттерн.

---

## 🏗️ Архитектура

Inventory Service — это **отдельный микросервис** со своей собственной логикой и хранилищем.

┌─────────────────────────────────────────────────────────────────────────┐
│                         INVENTORY SERVICE                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────────────┐ │
│  │  gRPC API   │   │  Сервисный  │   │        Repository           │ │
│  │  :50059     │◄──┤    слой     │───►│  (PostgreSQL / MongoDB)    │ │
│  └─────────────┘   └─────────────┘   └─────────────────────────────┘ │
│         │                  │                       │                   │
│         │                  ▼                       ▼                   │
│         │         ┌─────────────┐   ┌─────────────────────────────┐ │
│         │         │   Redis    │   │          Outbox             │ │
│         │         │   (кеш)    │   │   (таблица в PG)           │ │
│         │         └─────────────┘   └─────────────────────────────┘ │
│         │                  │                       │                   │
│         │                  ▼                       ▼                   │
│         │         ┌─────────────────────────────────────────────┐   │
│         └─────────│                NATS                        │   │
│                   │            (JetStream)                     │   │
│                   └─────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

---

## 🔄 Последние улучшения (2026-08-10)

### 1. Soft Delete (Мягкое удаление)

Добавлена поддержка мягкого удаления товаров через поле deleted_at.

**PostgreSQL:**

ALTER TABLE inventory_items ADD COLUMN deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX idx_inventory_deleted_at ON inventory_items(deleted_at);

**MongoDB:**

// Поле deleted_at добавлено в документы
// Все запросы автоматически фильтруют удаленные: deleted_at: nil

**API методы:**
- SoftDeleteItem(ctx, id) — помечает товар как удаленный
- GetItem, SearchItems — автоматически исключают удаленные записи

---

### 2. Индексы в MongoDB

Добавлены индексы для ускорения запросов:

// Составной индекс для товаров автора
{ author_id: 1, created_at: -1 }

// Полнотекстовый поиск по атрибутам
{ attributes: "text" } с весами:
  - attributes.name: 10
  - attributes.description: 5
  - attributes.material: 3

// Уникальный индекс по id
{ id: 1 } unique

---

### 3. Транзакции в MongoDB

Реализован метод ReserveItem с использованием session.WithTransaction:

func (r *MongoRepo) ReserveItem(ctx context.Context, itemID string, quantity int) error {
    session, _ := r.client.StartSession()
    return session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
        // 1. Проверка остатка
        // 2. Уменьшение stock
        // 3. Запись в историю
        return nil, nil
    })
}

---

### 4. Change Streams (альтернатива Outbox для MongoDB)

Реализован метод WatchChanges для подписки на изменения коллекции:

func (r *MongoRepo) WatchChanges(ctx context.Context, js nats.JetStreamContext) error {
    // Слушает insert/update операции
    // Публикует события в NATS
    // Не требует отдельной таблицы outbox
}

**Преимущества:**
- Нет отдельной таблицы outbox
- События доставляются в реальном времени
- Фильтрация только нужных операций

---

### 5. Batch операции в MongoDB

Добавлен метод BulkCreateItems для массовой вставки:

func (r *MongoRepo) BulkCreateItems(ctx context.Context, items []*model.Item) error {
    // Использует BulkWrite для производительности
    // Подходит для начальной загрузки и импорта
}

---

### 6. Полнотекстовый поиск в MongoDB

Реализован метод SearchByText через $text индекс:

func (r *MongoRepo) SearchByText(ctx context.Context, query string, limit, offset int) ([]*model.Item, int64, error) {
    filter := bson.M{
        "$text": bson.M{"$search": query},
        "deleted_at": nil,
    }
    // Сортировка по релевантности (textScore)
}

---

### 7. Улучшенное кеширование в Redis

**Было:** кеширование только отдельных товаров.

**Стало:** кеширование результатов поиска.

// Ключи:
inventory:item:{id}           // отдельный товар
inventory:search:{query}:{limit}:{offset}  // результаты поиска

// Инвалидация при изменениях
func (s *InventoryService) CreateItem(...) {
    // ...
    s.cache.InvalidateSearchCache(ctx)  // очистка всех поисковых кешей
}

---

### 8. Обработка ошибок

Добавлена централизованная ошибка ErrItemNotFound:

// internal/model/errors.go
var ErrItemNotFound = errors.New("item not found")

Используется во всех репозиториях для единообразной обработки "не найдено".

---

### 9. Единообразие интерфейсов

**Было:** сигнатуры GetByAuthor и GetByType различались в PostgreSQL и MongoDB.

**Стало:** все репозитории используют одинаковые сигнатуры:

type InventoryRepository interface {
    GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error)
    GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error)
}

Для пагинации используется SearchItems с limit и offset.

---

### 10. Новая миграция

Создана отдельная миграция 20260810120000_add_soft_delete_to_inventory.sql:

-- +goose Up
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_inventory_deleted_at ON inventory_items(deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_inventory_deleted_at;
ALTER TABLE inventory_items DROP COLUMN IF EXISTS deleted_at;

**Почему отдельная миграция:** не меняем существующие миграции, чтобы избежать проблем с goose версионированием.

---

## 🗄️ Базы данных

### PostgreSQL (основная)

**Таблицы:**
- inventory_items — основная таблица товаров с JSONB-атрибутами
- outbox — таблица для Outbox-паттерна
- goose_db_version — служебная таблица миграций

**Индексы:**
- GIN-индекс на attributes (JSONB) для быстрого поиска по динамическим полям
- Индекс на author_id и type для фильтрации
- Индекс на deleted_at для soft delete
- Full-text search по name

### MongoDB (альтернативная)

Реализован адаптер для MongoDB (mongo_repo.go), который позволяет переключиться через переменную окружения INVENTORY_DRIVER=mongo.

**Индексы в MongoDB:**

| Индекс | Тип | Назначение |
|--------|-----|------------|
| { author_id: 1, created_at: -1 } | Составной | Получение товаров автора |
| { type: 1 } | Простой | Фильтрация по типу |
| { attributes: "text" } | Полнотекстовый | Поиск по атрибутам с весами |
| { type: 1, price: 1 } | Составной | Сортировка по цене в категории |
| { id: 1 } | Уникальный | Поиск по ID |

### Redis (кеш)

**Ключи:**
- inventory:item:{id} — отдельный товар (TTL: 5 минут)
- inventory:search:{query}:{limit}:{offset} — результаты поиска (TTL: 5 минут)

**Инвалидация:**
- При создании, обновлении или удалении товара
- Автоматическая очистка всех поисковых кешей

---

## 🚀 Сборка и деплой

### Локальная сборка

cd /home/denismatveev/event_horizon/services/inventory

# Сборка бинарника
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o inventory-service ./cmd/main.go

### Docker образ

# Создаём Dockerfile
cat > Dockerfile.inventory.bin << 'EOF'
FROM scratch
COPY services/inventory/inventory-service /inventory-service
EXPOSE 50059 9096
CMD ["/inventory-service"]
EOF

# Сборка образа
docker build -t eastwesser/inventory:latest -f Dockerfile.inventory.bin .

# Пуш в Docker Hub
docker push eastwesser/inventory:latest

### Запуск

# Перезапуск сервиса
docker-compose -f deployments/docker-compose.cluster.yml down
docker-compose -f deployments/docker-compose.cluster.yml up -d

# Проверка
curl http://localhost:9096/health

---

## 📊 Метрики и мониторинг

### Health check

curl http://localhost:9096/health

**Ответ:**

{
  "status": "ok",
  "service": "inventory",
  "driver": "postgres"
}

### Prometheus метрики

curl http://localhost:9096/metrics

---

## 🔧 Конфигурация

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| INVENTORY_GRPC_PORT | Порт gRPC сервера | 50059 |
| INVENTORY_METRICS_PORT | Порт для метрик и health check | 9096 |
| INVENTORY_DRIVER | Драйвер: postgres или mongo | postgres |
| INVENTORY_PG_HOST | Хост PostgreSQL | localhost |
| INVENTORY_PG_PORT | Порт PostgreSQL | 5465 |
| INVENTORY_PG_USER | Пользователь PostgreSQL | eventhorizon |
| INVENTORY_PG_PASSWORD | Пароль PostgreSQL | eventhorizon |
| INVENTORY_PG_DB | База данных | eventhorizon_shop |
| INVENTORY_REDIS_ADDR | Адрес Redis | localhost:6379 |
| INVENTORY_MONGO_URI | URI MongoDB | mongodb://localhost:27017 |
| INVENTORY_MONGO_DB | Имя БД в MongoDB | inventory |
| NATS_URL | URL NATS кластера | nats://localhost:4222 |

---

## 🔗 Связанные сервисы

| Сервис | Роль | Интеграция |
|--------|------|------------|
| Shop | Потребитель событий (создаёт товары) | NATS (подписка) |
| Billing | Баланс авторов | — |
| Gateway | Входная точка для клиентов | gRPC |

---

## 📌 Следующие шаги

☑ Базовая CRUD-функциональность
☑ Поддержка PostgreSQL и MongoDB
☑ Outbox + NATS интеграция
☑ Redis кеширование
☑ Soft Delete (мягкое удаление)
☑ Индексы в MongoDB
☑ Транзакции в MongoDB (ReserveItem)
☑ Change Streams для MongoDB
☑ Batch операции (BulkCreateItems)
☑ Полнотекстовый поиск
☑ Единообразие интерфейсов
☑ Метрики и health check
☑ Миграции через Goose
□ Интеграционные тесты
□ Документация OpenAPI

---

## 📋 Резюме улучшений

| # | Улучшение | Статус |
|:---|:---|:---|
| 1 | Soft Delete (deleted_at) | ✅ |
| 2 | Индексы в MongoDB | ✅ |
| 3 | Транзакции в MongoDB (ReserveItem) | ✅ |
| 4 | Change Streams (альтернатива Outbox) | ✅ |
| 5 | Batch операции (BulkCreateItems) | ✅ |
| 6 | Полнотекстовый поиск (SearchByText) | ✅ |
| 7 | Кеширование результатов поиска в Redis | ✅ |
| 8 | Единообразие интерфейсов | ✅ |
| 9 | Новая миграция для deleted_at | ✅ |
| 10 | Обработка ошибок (ErrItemNotFound) | ✅ |

---

## 🐳 Улучшенный Dockerfile

FROM golang:1.25-alpine AS builder

# Устанавливаем необходимые утилиты
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY services/inventory/go.mod services/inventory/go.sum ./
RUN go mod download

# Копируем исходники
COPY services/inventory/ .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o inventory-service ./cmd/main.go

# Финальный образ (scratch для минимального размера)
FROM scratch

# Копируем CA сертификаты для HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем бинарник
COPY --from=builder /app/inventory-service /inventory-service

# Открываем порты
EXPOSE 50059 9096

# Точка входа
CMD ["/inventory-service"]

**Важно:** Проверь порты в docker-compose.yml (50059 и 9096 должны соответствовать).