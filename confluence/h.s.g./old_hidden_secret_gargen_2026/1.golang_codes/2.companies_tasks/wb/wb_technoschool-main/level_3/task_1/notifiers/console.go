package notifiers

import (
	"context"

	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_1/models"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) Send(ctx context.Context, notification *models.Notification) error {
	zlog.Logger.Info().
		Str("to", notification.Recipient).
		Str("subject", notification.Subject).
		Str("message", notification.Message).
		Str("id", notification.ID).
		Msg("CONSOLE NOTIFICATION")
	return nil
}

func (n *ConsoleNotifier) ChannelType() models.NotificationChannel {
	return models.ChannelConsole
}
