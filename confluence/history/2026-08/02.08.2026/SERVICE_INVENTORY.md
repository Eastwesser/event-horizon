🏗️ Inventory Service — полный проект
Я напишу тебе каркас с реальным кодом, который ты сможешь сразу запустить. Поехали!

1. Структура каталога
Создай в services/:

text
services/inventory/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── grpc_handler.go
│   ├── repository/
│   │   ├── repository.go          // интерфейс
│   │   ├── postgres_repo.go       // PG реализация
│   │   └── mongo_repo.go          // Mongo реализация
│   ├── service/
│   │   └── inventory_service.go
│   └── model/
│       └── item.go
├── proto/
│   ├── inventory.proto
│   ├── inventory.pb.go
│   └── inventory_grpc.pb.go
├── migrations/
│   └── 20260801000000_init_inventory.sql
├── Dockerfile
├── go.mod
└── README.md
2. Модель (model/item.go)
go
package model

import (
    "time"
)

// Item — универсальная модель товара
type Item struct {
    ID          string                 `json:"id"`
    AuthorID    string                 `json:"author_id"`
    Type        string                 `json:"type"`        // "брелок", "картина", "фенечка"
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Price       float64                `json:"price"`
    Stock       int                    `json:"stock"`
    Attributes  map[string]interface{} `json:"attributes"`  // динамические поля
    Images      []string               `json:"images"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
3. Интерфейс репозитория (repository/repository.go)
go
package repository

import (
    "context"
    "event-horizon/services/inventory/internal/model"
)

type InventoryRepository interface {
    CreateItem(ctx context.Context, item *model.Item) error
    GetItem(ctx context.Context, id string) (*model.Item, error)
    UpdateItem(ctx context.Context, item *model.Item) error
    DeleteItem(ctx context.Context, id string) error
    SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error)
    GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, error)
    GetByType(ctx context.Context, itemType string) ([]*model.Item, error)
}
4. PostgreSQL реализация (repository/postgres_repo.go)
Схема БД (миграция):

sql
-- migrations/20260801000000_init_inventory.sql
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    attributes JSONB NOT NULL DEFAULT '{}',
    images TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_inventory_author ON inventory_items(author_id);
CREATE INDEX idx_inventory_type ON inventory_items(type);
CREATE INDEX idx_inventory_attributes ON inventory_items USING GIN(attributes);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_inventory_updated_at BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
Код репозитория:

go
package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "event-horizon/services/inventory/internal/model"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/lib/pq"
)

type PostgresRepo struct {
    db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
    return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateItem(ctx context.Context, item *model.Item) error {
    if item.ID == "" {
        item.ID = uuid.New().String()
    }
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()

    attrsJSON, err := json.Marshal(item.Attributes)
    if err != nil {
        return fmt.Errorf("marshal attributes: %w", err)
    }

    query := `
        INSERT INTO inventory_items (
            id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `
    _, err = r.db.ExecContext(ctx, query,
        item.ID, item.AuthorID, item.Type, item.Name, item.Description,
        item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
        item.CreatedAt, item.UpdatedAt,
    )
    return err
}

func (r *PostgresRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
    query := `
        SELECT id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        FROM inventory_items WHERE id = $1
    `
    var item model.Item
    var attrsJSON []byte
    var images []string

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
        &item.Price, &item.Stock, &attrsJSON, pq.Array(&images),
        &item.CreatedAt, &item.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }

    if len(attrsJSON) > 0 {
        if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
            return nil, fmt.Errorf("unmarshal attributes: %w", err)
        }
    }
    item.Images = images
    return &item, nil
}

func (r *PostgresRepo) UpdateItem(ctx context.Context, item *model.Item) error {
    item.UpdatedAt = time.Now()

    attrsJSON, err := json.Marshal(item.Attributes)
    if err != nil {
        return fmt.Errorf("marshal attributes: %w", err)
    }

    query := `
        UPDATE inventory_items SET
            author_id = $1, type = $2, name = $3, description = $4,
            price = $5, stock = $6, attributes = $7, images = $8, updated_at = $9
        WHERE id = $10
    `
    _, err = r.db.ExecContext(ctx, query,
        item.AuthorID, item.Type, item.Name, item.Description,
        item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
        item.UpdatedAt, item.ID,
    )
    return err
}

func (r *PostgresRepo) DeleteItem(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM inventory_items WHERE id = $1", id)
    return err
}

func (r *PostgresRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
    var conditions []string
    var args []interface{}
    argCount := 1

    // Фильтры: author_id, type, price_min, price_max, attributes (JSONB)
    if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
        conditions = append(conditions, fmt.Sprintf("author_id = $%d", argCount))
        args = append(args, authorID)
        argCount++
    }
    if itemType, ok := filters["type"].(string); ok && itemType != "" {
        conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
        args = append(args, itemType)
        argCount++
    }
    if priceMin, ok := filters["price_min"].(float64); ok && priceMin > 0 {
        conditions = append(conditions, fmt.Sprintf("price >= $%d", argCount))
        args = append(args, priceMin)
        argCount++
    }
    if priceMax, ok := filters["price_max"].(float64); ok && priceMax > 0 {
        conditions = append(conditions, fmt.Sprintf("price <= $%d", argCount))
        args = append(args, priceMax)
        argCount++
    }

    // Поиск по атрибутам (JSONB)
    // Пример: {"material": "серебро"} -> attributes @> '{"material": "серебро"}'
    if attrs, ok := filters["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
        attrsJSON, err := json.Marshal(attrs)
        if err == nil {
            conditions = append(conditions, fmt.Sprintf("attributes @> $%d", argCount))
            args = append(args, string(attrsJSON))
            argCount++
        }
    }

    // Поиск по тексту (name, description)
    if query, ok := filters["query"].(string); ok && query != "" {
        conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
        searchTerm := "%" + query + "%"
        args = append(args, searchTerm)
        argCount++
    }

    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }

    // Считаем общее количество
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_items %s", whereClause)
    var total int64
    if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    // Выборка с пагинацией
    query := fmt.Sprintf(`
        SELECT id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
        FROM inventory_items %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereClause, argCount, argCount+1)
    args = append(args, limit, offset)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var items []*model.Item
    for rows.Next() {
        var item model.Item
        var attrsJSON []byte
        var images []string

        err := rows.Scan(
            &item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
            &item.Price, &item.Stock, &attrsJSON, pq.Array(&images),
            &item.CreatedAt, &item.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }

        if len(attrsJSON) > 0 {
            if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
                return nil, 0, err
            }
        }
        item.Images = images
        items = append(items, &item)
    }

    return items, total, nil
}

func (r *PostgresRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, error) {
    items, _, err := r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
    return items, err
}

func (r *PostgresRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, error) {
    items, _, err := r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
    return items, err
}
5. MongoDB реализация (repository/mongo_repo.go)
go
package repository

import (
    "context"
    "event-horizon/services/inventory/internal/model"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
    collection *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) *MongoRepo {
    return &MongoRepo{
        collection: db.Collection("inventory_items"),
    }
}

// Вспомогательная функция для преобразования
func toMongoFilter(filters map[string]interface{}) bson.M {
    m := bson.M{}
    if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
        m["author_id"] = authorID
    }
    if itemType, ok := filters["type"].(string); ok && itemType != "" {
        m["type"] = itemType
    }
    if priceMin, ok := filters["price_min"].(float64); ok && priceMin > 0 {
        m["price"] = bson.M{"$gte": priceMin}
    }
    if priceMax, ok := filters["price_max"].(float64); ok && priceMax > 0 {
        if _, ok := m["price"]; ok {
            m["price"].(bson.M)["$lte"] = priceMax
        } else {
            m["price"] = bson.M{"$lte": priceMax}
        }
    }
    // Атрибуты в Mongo хранятся как обычные поля
    if attrs, ok := filters["attributes"].(map[string]interface{}); ok {
        for k, v := range attrs {
            m["attributes."+k] = v
        }
    }
    if query, ok := filters["query"].(string); ok && query != "" {
        m["$or"] = []bson.M{
            {"name": bson.M{"$regex": query, "$options": "i"}},
            {"description": bson.M{"$regex": query, "$options": "i"}},
        }
    }
    return m
}

func (r *MongoRepo) CreateItem(ctx context.Context, item *model.Item) error {
    if item.ID == "" {
        item.ID = primitive.NewObjectID().Hex()
    }
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()

    // Mongo использует _id как ObjectID, но мы сохраним наш ID в поле id
    doc := bson.M{
        "id":          item.ID,
        "author_id":   item.AuthorID,
        "type":        item.Type,
        "name":        item.Name,
        "description": item.Description,
        "price":       item.Price,
        "stock":       item.Stock,
        "attributes":  item.Attributes,
        "images":      item.Images,
        "created_at":  item.CreatedAt,
        "updated_at":  item.UpdatedAt,
    }
    _, err := r.collection.InsertOne(ctx, doc)
    return err
}

func (r *MongoRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
    var result bson.M
    err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&result)
    if err != nil {
        return nil, err
    }
    return mapToItem(result), nil
}

func (r *MongoRepo) UpdateItem(ctx context.Context, item *model.Item) error {
    item.UpdatedAt = time.Now()
    update := bson.M{
        "$set": bson.M{
            "author_id":   item.AuthorID,
            "type":        item.Type,
            "name":        item.Name,
            "description": item.Description,
            "price":       item.Price,
            "stock":       item.Stock,
            "attributes":  item.Attributes,
            "images":      item.Images,
            "updated_at":  item.UpdatedAt,
        },
    }
    _, err := r.collection.UpdateOne(ctx, bson.M{"id": item.ID}, update)
    return err
}

func (r *MongoRepo) DeleteItem(ctx context.Context, id string) error {
    _, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
    return err
}

func (r *MongoRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
    mongoFilter := toMongoFilter(filters)
    
    // Общее количество
    total, err := r.collection.CountDocuments(ctx, mongoFilter)
    if err != nil {
        return nil, 0, err
    }

    opts := options.Find().
        SetLimit(int64(limit)).
        SetSkip(int64(offset)).
        SetSort(bson.D{{Key: "created_at", Value: -1}})

    cursor, err := r.collection.Find(ctx, mongoFilter, opts)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)

    var items []*model.Item
    for cursor.Next(ctx) {
        var result bson.M
        if err := cursor.Decode(&result); err != nil {
            return nil, 0, err
        }
        items = append(items, mapToItem(result))
    }
    return items, total, nil
}

func (r *MongoRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, error) {
    items, _, err := r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
    return items, err
}

func (r *MongoRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, error) {
    items, _, err := r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
    return items, err
}

// Вспомогательная функция
func mapToItem(m bson.M) *model.Item {
    item := &model.Item{
        ID:          getString(m, "id"),
        AuthorID:    getString(m, "author_id"),
        Type:        getString(m, "type"),
        Name:        getString(m, "name"),
        Description: getString(m, "description"),
        Price:       getFloat64(m, "price"),
        Stock:       getInt(m, "stock"),
        Images:      getStringSlice(m, "images"),
        CreatedAt:   getTime(m, "created_at"),
        UpdatedAt:   getTime(m, "updated_at"),
    }
    if attrs, ok := m["attributes"].(bson.M); ok {
        item.Attributes = make(map[string]interface{})
        for k, v := range attrs {
            item.Attributes[k] = v
        }
    }
    return item
}

// Хелперы для безопасного извлечения
func getString(m bson.M, key string) string {
    if v, ok := m[key].(string); ok {
        return v
    }
    return ""
}

func getFloat64(m bson.M, key string) float64 {
    switch v := m[key].(type) {
    case float64:
        return v
    case int:
        return float64(v)
    case int64:
        return float64(v)
    default:
        return 0
    }
}

func getInt(m bson.M, key string) int {
    switch v := m[key].(type) {
    case int:
        return v
    case int32:
        return int(v)
    case int64:
        return int(v)
    default:
        return 0
    }
}

func getStringSlice(m bson.M, key string) []string {
    if v, ok := m[key].(primitive.A); ok {
        result := make([]string, len(v))
        for i, val := range v {
            if s, ok := val.(string); ok {
                result[i] = s
            }
        }
        return result
    }
    return []string{}
}

func getTime(m bson.M, key string) time.Time {
    if v, ok := m[key].(primitive.DateTime); ok {
        return v.Time()
    }
    return time.Time{}
}
6. gRPC хендлер (handler/grpc_handler.go)
go
package handler

import (
    "context"
    "event-horizon/services/inventory/internal/model"
    "event-horizon/services/inventory/internal/repository"
    "event-horizon/services/inventory/internal/service"
    pb "event-horizon/services/inventory/proto"
    "fmt"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type GRPCHandler struct {
    pb.UnimplementedInventoryServiceServer
    service *service.InventoryService
}

func NewGRPCHandler(repo repository.InventoryRepository) *GRPCHandler {
    return &GRPCHandler{
        service: service.NewInventoryService(repo),
    }
}

func (h *GRPCHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.ItemResponse, error) {
    item := &model.Item{
        AuthorID:    req.AuthorId,
        Type:        req.Type,
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Stock:       int(req.Stock),
        Attributes:  req.Attributes.AsMap(),
        Images:      req.Images,
    }
    
    if err := h.service.CreateItem(ctx, item); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create item: %v", err)
    }

    return &pb.ItemResponse{
        Item: &pb.Item{
            Id:          item.ID,
            AuthorId:    item.AuthorID,
            Type:        item.Type,
            Name:        item.Name,
            Description: item.Description,
            Price:       item.Price,
            Stock:       int32(item.Stock),
            Attributes:  structpb.NewStruct(item.Attributes),
            Images:      item.Images,
            CreatedAt:   item.CreatedAt.String(),
            UpdatedAt:   item.UpdatedAt.String(),
        },
    }, nil
}

// Реализуй остальные методы (GetItem, UpdateItem, DeleteItem, SearchItems) аналогично
7. Dockerfile
dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o inventory-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/inventory-service .
EXPOSE 50055
CMD ["./inventory-service"]
8. go.mod
go
module event-horizon/services/inventory

go 1.21

require (
    github.com/google/uuid v1.6.0
    github.com/lib/pq v1.10.9
    go.mongodb.org/mongo-driver v1.15.0
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
)