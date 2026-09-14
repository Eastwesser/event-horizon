package service

import (
	"context"
	"time"

	"github.com/wb-go/wb_technoschool/level_3/task_6/models"
	"github.com/wb-go/wb_technoschool/level_3/task_6/storage"
)

type Service struct {
	storage *storage.Storage
}

func New(storage *storage.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) CreateTransaction(ctx context.Context, tx *models.Transaction) (int64, error) {
	if err := tx.Validate(); err != nil {
		return 0, err
	}

	return s.storage.CreateTransaction(ctx, tx)
}

func (s *Service) GetTransaction(ctx context.Context, id int64) (*models.Transaction, error) {
	return s.storage.GetTransaction(ctx, id)
}

func (s *Service) ListTransactions(ctx context.Context, filters storage.ListFilters) ([]models.Transaction, error) {
	return s.storage.ListTransactions(ctx, filters)
}

func (s *Service) UpdateTransaction(ctx context.Context, id int64, tx *models.Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}

	return s.storage.UpdateTransaction(ctx, id, tx)
}

func (s *Service) DeleteTransaction(ctx context.Context, id int64) error {
	return s.storage.DeleteTransaction(ctx, id)
}

func (s *Service) GetAnalytics(ctx context.Context, fromDate, toDate time.Time) (*models.Analytics, error) {
	return s.storage.GetAnalytics(ctx, fromDate, toDate)
}

func (s *Service) GetCategoryAnalytics(ctx context.Context, fromDate, toDate time.Time) ([]models.CategoryAnalytics, error) {
	return s.storage.GetCategoryAnalytics(ctx, fromDate, toDate)
}

func (s *Service) ExportToCSV(ctx context.Context, fromDate, toDate time.Time) (string, error) {
	return s.storage.ExportToCSV(ctx, fromDate, toDate)
}

