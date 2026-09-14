package notifiers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wb_technoschool/level_3/task_1/models"
)

// TelegramNotifier отправляет уведомления через Telegram
type TelegramNotifier struct {
	botToken string
}

func NewTelegramNotifier(botToken string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
	}
}

func (n *TelegramNotifier) Send(ctx context.Context, notification *models.Notification) error {
	if n.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	message := notification.Message
	if notification.Subject != "" {
		message = fmt.Sprintf("<b>%s</b>\n\n%s", notification.Subject, notification.Message)
	}

	payload := map[string]interface{}{
		"chat_id":    notification.Recipient,
		"text":       message,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

func (n *TelegramNotifier) ChannelType() models.NotificationChannel {
	return models.ChannelTelegram
}

