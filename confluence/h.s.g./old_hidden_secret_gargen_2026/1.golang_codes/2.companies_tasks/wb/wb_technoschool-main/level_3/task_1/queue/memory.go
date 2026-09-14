package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/wb_technoschool/level_3/task_1/models"
)

// InMemoryQueue реализует простую in-memory очередь для тестирования без RabbitMQ
type InMemoryQueue struct {
	mu            sync.Mutex
	notifications chan *models.Notification
}

// NewInMemoryQueue создает новую in-memory очередь
func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		notifications: make(chan *models.Notification, 100),
	}
}

func (q *InMemoryQueue) Publish(ctx context.Context, notification *models.Notification) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case q.notifications <- notification:
		log.Printf("Published notification %s to in-memory queue", notification.ID)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		log.Printf("Warning: Queue is full, notification %s may be delayed", notification.ID)
		// Пытаемся добавить с таймаутом
		select {
		case q.notifications <- notification:
			return nil
		case <-time.After(5 * time.Second):
			return ErrQueueFull
		}
	}
}

func (q *InMemoryQueue) Consume(ctx context.Context) (<-chan *models.Notification, error) {
	notifications := make(chan *models.Notification)

	go func() {
		defer close(notifications)
		for {
			select {
			case <-ctx.Done():
				return
			case notification, ok := <-q.notifications:
				if !ok {
					return
				}

				// Проверяем, не пора ли отправлять уведомление
				delay := time.Until(notification.ScheduledAt)
				if delay > 0 {
					log.Printf("Notification %s scheduled for %s, waiting %v",
						notification.ID, notification.ScheduledAt.Format(time.RFC3339), delay)
					
					// Ждем до запланированного времени
					select {
					case <-time.After(delay):
						// Время пришло
					case <-ctx.Done():
						return
					}
				}

				// Отправляем уведомление
				select {
				case notifications <- notification:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return notifications, nil
}

func (q *InMemoryQueue) Close() error {
	close(q.notifications)
	return nil
}

var ErrQueueFull = &QueueError{Message: "queue is full"}

type QueueError struct {
	Message string
}

func (e *QueueError) Error() string {
	return e.Message
}

