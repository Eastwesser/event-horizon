package worker

import (
	"context"
	"path/filepath"

	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_4/models"
	"github.com/wb_technoschool/level_3/task_4/processor"
	"github.com/wb_technoschool/level_3/task_4/queue"
	"github.com/wb_technoschool/level_3/task_4/storage"
)

type Worker struct {
	storage   storage.Storage
	fileStore *storage.FileStorage
	queue     queue.Queue
	processor *processor.Processor
}

func NewWorker(
	storage storage.Storage,
	fileStore *storage.FileStorage,
	queue queue.Queue,
	processor *processor.Processor,
) *Worker {
	return &Worker{
		storage:   storage,
		fileStore: fileStore,
		queue:     queue,
		processor: processor,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	zlog.Logger.Info().Msg("Worker started, waiting for tasks...")

	tasks, err := w.queue.Consume(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Worker stopped")
			return ctx.Err()
		case task, ok := <-tasks:
			if !ok {
				zlog.Logger.Info().Msg("Task channel closed")
				return nil
			}
			go w.processTask(ctx, task)
		}
	}
}

func (w *Worker) processTask(ctx context.Context, task *models.ProcessingTask) {
	zlog.Logger.Info().Str("image_id", task.ImageID).Msg("Processing image")

	image, err := w.storage.Get(task.ImageID)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("image_id", task.ImageID).Msg("Failed to get image")
		return
	}

	w.storage.UpdateStatus(task.ImageID, models.StatusProcessing, "")

	processedPath := filepath.Join("processed", task.ImageID+".jpg")
	thumbnailPath := filepath.Join("thumbnails", task.ImageID+".jpg")

	fullProcessedPath := filepath.Join(w.fileStore.BasePath(), processedPath)
	fullThumbnailPath := filepath.Join(w.fileStore.BasePath(), thumbnailPath)

	if err := w.processor.ProcessImage(
		image.OriginalPath,
		fullProcessedPath,
		800,
		600,
		200,
		"Watermark",
	); err != nil {
		zlog.Logger.Error().Err(err).Str("image_id", task.ImageID).Msg("Failed to process image")
		w.storage.UpdateStatus(task.ImageID, models.StatusFailed, err.Error())
		return
	}

	if err := w.processor.ProcessThumbnail(
		image.OriginalPath,
		fullThumbnailPath,
		200,
	); err != nil {
		zlog.Logger.Error().Err(err).Str("image_id", task.ImageID).Msg("Failed to process thumbnail")
		w.storage.UpdateStatus(task.ImageID, models.StatusFailed, err.Error())
		return
	}

	w.storage.UpdatePaths(task.ImageID, fullProcessedPath, fullThumbnailPath)
	w.storage.UpdateStatus(task.ImageID, models.StatusCompleted, "")

	zlog.Logger.Info().Str("image_id", task.ImageID).Msg("Image processed successfully")
}

