package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepo — адаптер для MongoDB, реализующий интерфейс InventoryRepository.
// Использует гибкую схему с attributes map[string]interface{} для хранения
// разнородных товаров (брелоки, картины, фенечки).
type MongoRepo struct {
	collection *mongo.Collection
	client     *mongo.Client // Сохраняем клиент для транзакций и health check
}

// NewMongoRepo создает новый репозиторий и настраивает индексы.
// Индексы критичны для производительности: author_id, type, полнотекстовый поиск по attributes.
func NewMongoRepo(db *mongo.Database) *MongoRepo {
	collection := db.Collection("inventory_items")
	client := db.Client()

	// Создаем индексы для частых запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		// Составной индекс для получения товаров автора с сортировкой по дате
		{
			Keys: bson.D{
				{Key: "author_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		// Индекс для фильтрации по типу товара
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		// Полнотекстовый индекс по атрибутам (веса для релевантности)
		{
			Keys: bson.D{
				{Key: "attributes", Value: "text"},
			},
			Options: options.Index().
				SetDefaultLanguage("russian").
				SetWeights(bson.M{
					"attributes.name":        10,
					"attributes.description": 5,
					"attributes.material":    3,
				}),
		},
		// Составной индекс для фильтрации по типу + цене (сортировка)
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "price", Value: 1},
			},
		},
		// Индекс для поиска по полю id (уникальный)
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Printf("⚠️ Failed to create MongoDB indexes: %v", err)
	}

	return &MongoRepo{
		collection: collection,
		client:     client,
	}
}

// ---------------------- Вспомогательные функции ----------------------

// toMongoFilter преобразует map фильтров в bson.M для MongoDB.
// Поддерживает фильтрацию по author_id, type, price_min/max, query (текст),
// а также вложенные атрибуты с операторами $in, $gt, $lt.
func toMongoFilter(filters map[string]interface{}) bson.M {
	m := bson.M{}

	// Фильтр по автору
	if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
		m["author_id"] = authorID
	}

	// Фильтр по одному типу
	if itemType, ok := filters["type"].(string); ok && itemType != "" {
		m["type"] = itemType
	}

	// Фильтр по нескольким типам (массив)
	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		m["type"] = bson.M{"$in": types}
	}

	// Диапазон цен
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

	// Фильтр по вложенным атрибутам (attributes.color, attributes.size и т.д.)
	if attrs, ok := filters["attributes"].(map[string]interface{}); ok {
		for k, v := range attrs {
			// Поддержка операторов: {"color": {"$in": ["red", "blue"]}}
			if conditions, ok := v.(map[string]interface{}); ok {
				for op, val := range conditions {
					if op == "$in" || op == "$nin" {
						m["attributes."+k] = bson.M{op: val}
					} else if op == "$gt" || op == "$gte" || op == "$lt" || op == "$lte" {
						m["attributes."+k] = bson.M{op: val}
					} else {
						m["attributes."+k] = v // fallback
					}
				}
			} else {
				m["attributes."+k] = v
			}
		}
	}

	// Полнотекстовый поиск по name и description (через $or)
	if query, ok := filters["query"].(string); ok && query != "" {
		m["$or"] = []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	// Не показываем мягко удаленные записи
	m["deleted_at"] = nil

	return m
}

// mapToItem преобразует bson.M (документ MongoDB) в модель Item.
// Использует хелперы для безопасного извлечения значений разных типов.
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

	// Извлекаем attributes (хранятся как вложенный bson.M)
	if attrs, ok := m["attributes"].(bson.M); ok {
		item.Attributes = make(map[string]interface{})
		for k, v := range attrs {
			item.Attributes[k] = v
		}
	} else if attrs, ok := m["attributes"].(map[string]interface{}); ok {
		item.Attributes = attrs
	}

	return item
}

// ---------------------- Хелперы для безопасного извлечения данных ----------------------

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

// ---------------------- CRUD-методы (реализация интерфейса) ----------------------

// CreateItem создает новый товар в коллекции.
// Генерирует ID в формате hex, устанавливает временные метки.
func (r *MongoRepo) CreateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		item.ID = primitive.NewObjectID().Hex()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

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
		"deleted_at":  nil, // явно указываем, что не удален
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

// GetItem возвращает товар по ID (игнорирует удаленные).
func (r *MongoRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
	var result bson.M
	err := r.collection.FindOne(ctx, bson.M{"id": id, "deleted_at": nil}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return mapToItem(result), nil
}

// UpdateItem обновляет все поля товара (стратегия PUT).
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
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": item.ID, "deleted_at": nil}, update)
	return err
}

// DeleteItem выполняет жесткое удаление товара.
func (r *MongoRepo) DeleteItem(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// SoftDeleteItem выполняет мягкое удаление — устанавливает deleted_at и status.
func (r *MongoRepo) SoftDeleteItem(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"deleted_at": time.Now(),
				"status":     "deleted",
			},
		},
	)
	return err
}

// SearchItems выполняет поиск товаров с фильтрами, пагинацией и сортировкой.
// Автоматически исключает мягко удаленные записи (deleted_at: nil).
func (r *MongoRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	mongoFilter := toMongoFilter(filters)

	// Общее количество (без учета пагинации)
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

// GetByAuthor возвращает все товары автора (с пагинацией по умолчанию 1000).
func (r *MongoRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}

// GetByType возвращает все товары указанного типа (с пагинацией по умолчанию 1000).
func (r *MongoRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}

// ---------------------- Дополнительные методы для MongoDB ----------------------

// ReserveItem — пример использования транзакций в MongoDB.
// Бронирует товар: проверяет остаток, уменьшает stock и создает запись в истории.
// Возвращает новый остаток.
func (r *MongoRepo) ReserveItem(ctx context.Context, itemID string, quantity int) (int, error) {
	session, err := r.client.StartSession()
	if err != nil {
		return 0, err
	}
	defer session.EndSession(ctx)

	var remaining int
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Проверяем наличие и текущий остаток
		var result bson.M
		err := r.collection.FindOne(sessCtx, bson.M{"id": itemID, "deleted_at": nil}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, model.ErrItemNotFound
			}
			return nil, err
		}

		stock := getInt(result, "stock")
		if stock < quantity {
			return nil, model.ErrNotEnoughStock
		}

		// 2. Уменьшаем остаток
		_, err = r.collection.UpdateOne(
			sessCtx,
			bson.M{"id": itemID},
			bson.M{"$inc": bson.M{"stock": -quantity}},
		)
		if err != nil {
			return nil, err
		}

		// 3. Получаем новый остаток
		var updatedResult bson.M
		err = r.collection.FindOne(sessCtx, bson.M{"id": itemID}).Decode(&updatedResult)
		if err != nil {
			return nil, err
		}
		remaining = getInt(updatedResult, "stock")

		// 4. Создаём запись в истории (другая коллекция)
		historyColl := r.collection.Database().Collection("reservation_history")
		_, err = historyColl.InsertOne(sessCtx, bson.M{
			"item_id":     itemID,
			"quantity":    quantity,
			"remaining":   remaining,
			"reserved_at": time.Now(),
		})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return 0, err
	}

	return remaining, nil
}

// SearchByText — полнотекстовый поиск по индексу attributes.
// Использует $text для поиска по полям, указанным в индексе (name, description, material).
func (r *MongoRepo) SearchByText(ctx context.Context, query string, limit, offset int) ([]*model.Item, int64, error) {
	filter := bson.M{
		"$text":      bson.M{"$search": query},
		"deleted_at": nil,
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})

	cursor, err := r.collection.Find(ctx, filter, opts)
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

// BulkCreateItems — массовая вставка товаров для начальной загрузки или импорта.
// Использует BulkWrite для производительности.
func (r *MongoRepo) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	var operations []mongo.WriteModel

	for _, item := range items {
		if item.ID == "" {
			item.ID = primitive.NewObjectID().Hex()
		}
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

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
			"deleted_at":  nil,
		}

		operations = append(operations, mongo.NewInsertOneModel().SetDocument(doc))
	}

	_, err := r.collection.BulkWrite(ctx, operations)
	return err
}

// WatchChanges — альтернатива Outbox для MongoDB.
// Слушает изменения коллекции (insert/update) и публикует события в NATS.
// Это позволяет реализовать событийную архитектуру без дополнительной таблицы outbox.
func (r *MongoRepo) WatchChanges(ctx context.Context, js nats.JetStreamContext) error {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: bson.D{{Key: "$in", Value: []string{"insert", "update"}}}},
		}}},
	}

	stream, err := r.collection.Watch(ctx, pipeline)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var event bson.M
		if err := stream.Decode(&event); err != nil {
			log.Printf("Failed to decode change event: %v", err)
			continue
		}

		payload, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue
		}

		_, err = js.Publish("inventory.item.changed", payload)
		if err != nil {
			log.Printf("Failed to publish change event: %v", err)
		} else {
			log.Printf("📡 Published change event for item: %s", event)
		}
	}

	return stream.Err()
}

// GetItemWithAuthor — пример агрегации с $lookup.
// Объединяет товар с информацией об авторе из коллекции authors.
func (r *MongoRepo) GetItemWithAuthor(ctx context.Context, itemID string) (*model.Item, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"id": itemID, "deleted_at": nil}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "authors"},
			{Key: "localField", Value: "author_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "as", Value: "author"},
		}}},
		{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$author"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "price", Value: 1},
			{Key: "author_name", Value: "$author.name"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, model.ErrItemNotFound
	}

	return mapToItem(results[0]), nil
}

// GetStats — возвращает статистику по товарам в MongoDB.
func (r *MongoRepo) GetStats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{
		ByType:   make(map[string]int64),
		ByAuthor: make(map[string]int64),
	}

	// 1. Общее количество
	total, err := r.collection.CountDocuments(ctx, bson.M{"deleted_at": nil})
	if err != nil {
		return nil, err
	}
	stats.TotalItems = total

	// 2. Группировка по типу
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, r := range results {
		stats.ByType[r.ID] = r.Count
	}

	// 3. Группировка по автору (топ 10)
	pipeline = mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$author_id"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err = r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var authorResults []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &authorResults); err != nil {
		return nil, err
	}

	for _, r := range authorResults {
		stats.ByAuthor[r.ID] = r.Count
	}

	return stats, nil
}

// RestoreItem — восстанавливает мягко удаленный товар.
func (r *MongoRepo) RestoreItem(ctx context.Context, id string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": id, "deleted_at": bson.M{"$ne": nil}},
		bson.M{
			"$set": bson.M{
				"deleted_at": nil,
				"status":     "active",
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return model.ErrItemNotFound
	}
	return nil
}