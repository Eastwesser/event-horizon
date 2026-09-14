package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wb_technoschool/level_3/task_1/cache"
	"github.com/wb_technoschool/level_3/task_1/models"
	"github.com/wb_technoschool/level_3/task_1/queue"
	"github.com/wb_technoschool/level_3/task_1/storage"
)

type NotificationService struct {
	storage storage.Storage
	cache   cache.Cache
	queue   queue.Queue
}

func NewNotificationService(storage storage.Storage, cache cache.Cache, queue queue.Queue) *NotificationService {
	return &NotificationService{
		storage: storage,
		cache:   cache,
		queue:   queue,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	if req.ScheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled_at must be in the future")
	}

	notification := &models.Notification{
		ID:          uuid.New().String(),
		Channel:     req.Channel,
		Recipient:   req.Recipient,
		Subject:     req.Subject,
		Message:     req.Message,
		ScheduledAt: req.ScheduledAt,
		Status:      models.StatusScheduled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		RetryCount:  0,
	}

	if err := s.storage.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.cache.Set(ctx, notification)

	if err := s.queue.Publish(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to publish notification: %w", err)
	}

	return notification, nil
}

func (s *NotificationService) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	notification, err := s.cache.Get(ctx, id)
	if err == nil && notification != nil {
		return notification, nil
	}

	notification, err = s.storage.Get(id)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, notification)

	return notification, nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, id string) error {
	notification, err := s.storage.Get(id)
	if err != nil {
		return err
	}

	if notification.Status != models.StatusScheduled && notification.Status != models.StatusPending {
		return fmt.Errorf("cannot cancel notification with status %s", notification.Status)
	}

	notification.Status = models.StatusCancelled
	notification.UpdatedAt = time.Now()

	if err := s.storage.Update(notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	s.cache.Delete(ctx, id)

	return nil
}

func (s *NotificationService) GetAllNotifications(ctx context.Context) ([]*models.Notification, error) {
	return s.storage.GetAll()
}

