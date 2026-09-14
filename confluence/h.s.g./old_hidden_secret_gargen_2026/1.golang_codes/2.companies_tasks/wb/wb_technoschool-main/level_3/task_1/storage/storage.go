package storage

import (
	"errors"
	"sync"

	"github.com/wb_technoschool/level_3/task_1/models"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type Storage interface {
	Create(notification *models.Notification) error
	Get(id string) (*models.Notification, error)
	Update(notification *models.Notification) error
	Delete(id string) error
	GetAll() ([]*models.Notification, error)
}

type InMemoryStorage struct {
	mu            sync.RWMutex
	notifications map[string]*models.Notification
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		notifications: make(map[string]*models.Notification),
	}
}

func (s *InMemoryStorage) Create(notification *models.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[notification.ID] = notification
	return nil
}

func (s *InMemoryStorage) Get(id string) (*models.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	notification, exists := s.notifications[id]
	if !exists {
		return nil, ErrNotificationNotFound
	}
	notificationCopy := *notification
	return &notificationCopy, nil
}

func (s *InMemoryStorage) Update(notification *models.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.notifications[notification.ID]; !exists {
		return ErrNotificationNotFound
	}
	s.notifications[notification.ID] = notification
	return nil
}

func (s *InMemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.notifications[id]; !exists {
		return ErrNotificationNotFound
	}
	delete(s.notifications, id)
	return nil
}

func (s *InMemoryStorage) GetAll() ([]*models.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	notifications := make([]*models.Notification, 0, len(s.notifications))
	for _, notification := range s.notifications {
		notificationCopy := *notification
		notifications = append(notifications, &notificationCopy)
	}
	return notifications, nil
}

