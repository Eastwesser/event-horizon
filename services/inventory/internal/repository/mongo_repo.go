package repository

import (
    "context"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
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

func (r *MongoRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
    items, total, err := r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
    return items, total, err
}

func (r *MongoRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
    items, total, err := r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
    return items, total, err
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
