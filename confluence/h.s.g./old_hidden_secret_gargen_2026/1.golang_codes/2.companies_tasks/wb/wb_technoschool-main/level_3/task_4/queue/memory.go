package queue

import (
	"context"
	"sync"

	"github.com/wb_technoschool/level_3/task_4/models"
)

type InMemoryQueue struct {
	mu    sync.Mutex
	tasks []*models.ProcessingTask
	ch    chan *models.ProcessingTask
}

func NewInMemoryQueue() *InMemoryQueue {
	return &InMemoryQueue{
		ch: make(chan *models.ProcessingTask, 100),
	}
}

func (q *InMemoryQueue) Publish(ctx context.Context, task *models.ProcessingTask) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case q.ch <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *InMemoryQueue) Consume(ctx context.Context) (<-chan *models.ProcessingTask, error) {
	return q.ch, nil
}

func (q *InMemoryQueue) Close() error {
	close(q.ch)
	return nil
}

