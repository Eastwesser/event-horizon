package notifiers

import (
	"context"
	"fmt"

	"github.com/wb_technoschool/level_3/task_1/models"
)

// Notifier представляет интерфейс для отправки уведомлений
type Notifier interface {
	Send(ctx context.Context, notification *models.Notification) error
	ChannelType() models.NotificationChannel
}

// NotifierFactory создает notifier на основе канала
type NotifierFactory struct {
	notifiers map[models.NotificationChannel]Notifier
}

// NewNotifierFactory создает новую фабрику notifier'ов
func NewNotifierFactory() *NotifierFactory {
	return &NotifierFactory{
		notifiers: make(map[models.NotificationChannel]Notifier),
	}
}

// Register регистрирует notifier для канала
func (f *NotifierFactory) Register(notifier Notifier) {
	f.notifiers[notifier.ChannelType()] = notifier
}

// GetNotifier возвращает notifier для указанного канала
func (f *NotifierFactory) GetNotifier(channel models.NotificationChannel) (Notifier, error) {
	notifier, exists := f.notifiers[channel]
	if !exists {
		return nil, fmt.Errorf("notifier for channel %s not found", channel)
	}
	return notifier, nil
}

