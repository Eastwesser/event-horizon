package worker

import (
	"context"
	"math"
	"time"

	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_1/cache"
	"github.com/wb_technoschool/level_3/task_1/models"
	"github.com/wb_technoschool/level_3/task_1/notifiers"
	"github.com/wb_technoschool/level_3/task_1/queue"
	"github.com/wb_technoschool/level_3/task_1/storage"
)

const (
	MaxRetries        = 5
	InitialRetryDelay = 1 * time.Second
	MaxRetryDelay     = 5 * time.Minute
)

type Worker struct {
	storage         storage.Storage
	cache           cache.Cache
	queue           queue.Queue
	notifierFactory *notifiers.NotifierFactory
}

func NewWorker(
	storage storage.Storage,
	cache cache.Cache,
	queue queue.Queue,
	notifierFactory *notifiers.NotifierFactory,
) *Worker {
	return &Worker{
		storage:         storage,
		cache:           cache,
		queue:           queue,
		notifierFactory: notifierFactory,
	}
}

// Start запускает worker
func (w *Worker) Start(ctx context.Context) error {
	zlog.Logger.Info().Msg("Worker started, waiting for notifications...")

	notifications, err := w.queue.Consume(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Worker stopped")
			return ctx.Err()
		case notification, ok := <-notifications:
			if !ok {
				zlog.Logger.Info().Msg("Notification channel closed")
				return nil
			}
			w.processNotification(ctx, notification)
		}
	}
}

func (w *Worker) processNotification(ctx context.Context, notification *models.Notification) {
	zlog.Logger.Info().
		Str("id", notification.ID).
		Str("channel", string(notification.Channel)).
		Str("recipient", notification.Recipient).
		Msg("Processing notification")

	notification.Status = models.StatusSending
	notification.UpdatedAt = time.Now()
	if err := w.storage.Update(notification); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to update notification status")
	}
	w.cache.Set(ctx, notification)

	notifier, err := w.notifierFactory.GetNotifier(notification.Channel)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("channel", string(notification.Channel)).Msg("Failed to get notifier")
		w.handleFailure(ctx, notification, err)
		return
	}

	err = notifier.Send(ctx, notification)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("id", notification.ID).Msg("Failed to send notification")
		w.handleFailure(ctx, notification, err)
		return
	}

	zlog.Logger.Info().Str("id", notification.ID).Msg("Notification sent successfully")
	now := time.Now()
	notification.Status = models.StatusSent
	notification.SentAt = &now
	notification.UpdatedAt = now
	notification.LastError = ""

	if err := w.storage.Update(notification); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to update notification")
	}
	w.cache.Set(ctx, notification)
}

func (w *Worker) handleFailure(ctx context.Context, notification *models.Notification, err error) {
	notification.RetryCount++
	notification.LastError = err.Error()
	notification.UpdatedAt = time.Now()

	if notification.RetryCount >= MaxRetries {
		zlog.Logger.Error().
			Str("id", notification.ID).
			Int("retries", MaxRetries).
			Msg("Notification failed after max retries")
		notification.Status = models.StatusFailed
		if err := w.storage.Update(notification); err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to update notification")
		}
		w.cache.Set(ctx, notification)
		return
	}

	delay := calculateRetryDelay(notification.RetryCount)
	zlog.Logger.Warn().
		Str("id", notification.ID).
		Dur("delay", delay).
		Int("attempt", notification.RetryCount).
		Int("max", MaxRetries).
		Msg("Notification will be retried")

	notification.Status = models.StatusPending
	if err := w.storage.Update(notification); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to update notification")
	}
	w.cache.Set(ctx, notification)

	go func() {
		time.Sleep(delay)
		if err := w.queue.Publish(context.Background(), notification); err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to republish notification")
		}
	}()
}

func calculateRetryDelay(retryCount int) time.Duration {
	delay := InitialRetryDelay * time.Duration(math.Pow(2, float64(retryCount-1)))
	if delay > MaxRetryDelay {
		delay = MaxRetryDelay
	}
	return delay
}
