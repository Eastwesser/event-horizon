package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestMongo поднимает тестовую MongoDB (использует testcontainers или локальный экземпляр).
// Для CI/CD рекомендуется использовать testcontainers.
func setupTestMongo(t *testing.T) *MongoRepo {
	// Используем локальную MongoDB (для разработки)
	// В CI можно заменить на testcontainers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Создаем временную базу данных для тестов
	db := client.Database("test_inventory_" + time.Now().Format("20060102150405"))
	repo := NewMongoRepo(db)

	// Очищаем после тестов
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})

	return repo
}

// TestMongoRepo_CreateItem тестирует создание товара.
func TestMongoRepo_CreateItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	item := &model.Item{
		AuthorID:    "author-123",
		Type:        "брелок",
		Name:        "Ключница 'Дракон'",
		Description: "Металлическая ключница с гравировкой",
		Price:       1500.00,
		Stock:       10,
		Attributes: map[string]interface{}{
			"material": "металл",
			"weight":   "150g",
			"color":    "золотой",
		},
		Images: []string{"image1.jpg", "image2.jpg"},
	}

	err := repo.CreateItem(ctx, item)
	assert.NoError(t, err)
	assert.NotEmpty(t, item.ID)
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())

	// Проверяем, что товар сохранился
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.Name, saved.Name)
	assert.Equal(t, item.Price, saved.Price)
	assert.Equal(t, item.Attributes["material"], saved.Attributes["material"])
}

// TestMongoRepo_UpdateItem тестирует обновление товара.
func TestMongoRepo_UpdateItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "картина",
		Name:     "Закат над морем",
		Price:    5000.00,
		Stock:    5,
		Attributes: map[string]interface{}{
			"size":      "50x70cm",
			"technique": "масло",
		},
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Обновляем
	item.Name = "Закат над морем (обновлено)"
	item.Price = 5500.00
	item.Attributes["frame"] = "дерево"

	err = repo.UpdateItem(ctx, item)
	assert.NoError(t, err)

	// Проверяем
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Закат над морем (обновлено)", saved.Name)
	assert.Equal(t, 5500.00, saved.Price)
	assert.Equal(t, "дерево", saved.Attributes["frame"])
}

// TestMongoRepo_SoftDeleteItem тестирует мягкое удаление.
func TestMongoRepo_SoftDeleteItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "фенечка",
		Name:     "Фенечка с бусинами",
		Price:    300.00,
		Stock:    20,
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Мягко удаляем
	err = repo.SoftDeleteItem(ctx, item.ID)
	assert.NoError(t, err)

	// Прямой поиск по ID должен вернуть ошибку (deleted_at != nil)
	_, err = repo.GetItem(ctx, item.ID)
	assert.Error(t, err)

	// Поиск с фильтрами не должен возвращать удаленный товар
	items, total, err := repo.SearchItems(ctx, map[string]interface{}{"author_id": "author-123"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, items, 0)
}

// TestMongoRepo_SearchItems тестирует поиск с фильтрами.
func TestMongoRepo_SearchItems(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем несколько товаров
	items := []*model.Item{
		{
			AuthorID: "author-123",
			Type:     "брелок",
			Name:     "Ключница 'Дракон'",
			Price:    1500.00,
			Stock:    10,
			Attributes: map[string]interface{}{
				"material": "металл",
				"color":    "золотой",
			},
		},
		{
			AuthorID: "author-123",
			Type:     "брелок",
			Name:     "Ключница 'Сова'",
			Price:    1200.00,
			Stock:    8,
			Attributes: map[string]interface{}{
				"material": "дерево",
				"color":    "коричневый",
			},
		},
		{
			AuthorID: "author-456",
			Type:     "картина",
			Name:     "Закат над морем",
			Price:    5000.00,
			Stock:    5,
		},
	}

	for _, item := range items {
		err := repo.CreateItem(ctx, item)
		require.NoError(t, err)
	}

	// Тест 1: Поиск по author_id
	result, total, err := repo.SearchItems(ctx, map[string]interface{}{"author_id": "author-123"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)

	// Тест 2: Поиск по type
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{"type": "брелок"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Тест 3: Поиск по price_min
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{"price_min": 1300.0}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // 1500 и 5000

	// Тест 4: Поиск по атрибутам
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{
		"attributes": map[string]interface{}{
			"material": "металл",
		},
	}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Ключница 'Дракон'", result[0].Name)
}

// TestMongoRepo_ReserveItem тестирует транзакцию резервирования.
func TestMongoRepo_ReserveItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар с остатком 10
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "брелок",
		Name:     "Тестовый брелок",
		Price:    1000.00,
		Stock:    10,
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Резервируем 3 штуки
	err = repo.ReserveItem(ctx, item.ID, 3)
	assert.NoError(t, err)

	// Проверяем остаток
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, 7, saved.Stock)

	// Пытаемся зарезервировать больше, чем есть
	err = repo.ReserveItem(ctx, item.ID, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough stock")
}

// TestMongoRepo_BulkCreateItems тестирует массовую вставку.
func TestMongoRepo_BulkCreateItems(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	items := []*model.Item{
		{AuthorID: "author-1", Type: "тип1", Name: "Товар 1", Price: 100, Stock: 5},
		{AuthorID: "author-2", Type: "тип2", Name: "Товар 2", Price: 200, Stock: 10},
		{AuthorID: "author-3", Type: "тип3", Name: "Товар 3", Price: 300, Stock: 15},
	}

	err := repo.BulkCreateItems(ctx, items)
	assert.NoError(t, err)

	// Проверяем, что все создались
	for _, item := range items {
		saved, err := repo.GetItem(ctx, item.ID)
		assert.NoError(t, err)
		assert.Equal(t, item.Name, saved.Name)
	}
}