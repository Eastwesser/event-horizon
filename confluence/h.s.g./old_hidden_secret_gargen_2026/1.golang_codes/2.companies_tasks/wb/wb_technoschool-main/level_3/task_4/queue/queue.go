package queue

import (
	"context"

	"github.com/wb_technoschool/level_3/task_4/models"
)

type Queue interface {
	Publish(ctx context.Context, task *models.ProcessingTask) error
	Consume(ctx context.Context) (<-chan *models.ProcessingTask, error)
	Close() error
}

