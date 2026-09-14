package service

import (
	"context"

	"github.com/wb_technoschool/level_3/task_4/models"
	"github.com/wb_technoschool/level_3/task_4/queue"
	"github.com/wb_technoschool/level_3/task_4/storage"
)

type ImageService struct {
	storage storage.Storage
	queue   queue.Queue
}

func NewImageService(storage storage.Storage, queue queue.Queue) *ImageService {
	return &ImageService{
		storage: storage,
		queue:   queue,
	}
}

func (s *ImageService) UploadImage(ctx context.Context, image *models.Image) error {
	if err := s.storage.Save(image); err != nil {
		return err
	}

	task := &models.ProcessingTask{
		ImageID: image.ID,
		Action:  "process",
	}

	return s.queue.Publish(ctx, task)
}

