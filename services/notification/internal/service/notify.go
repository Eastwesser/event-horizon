package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
)

type Notifier struct {
	log      *slog.Logger
	token    string
	chatID   string
	client   *http.Client
}

func NewNotifier(log *slog.Logger, token, chatID string) *Notifier {
	return &Notifier{
		log:    log,
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case kafka.TopicPurchasePaid:
		e, err := events.UnmarshalPurchasePaid(msg.Value)
		if err != nil {
			return err
		}
		return n.notify(ctx, fmt.Sprintf("PurchasePaid: user=%s item=%s price=%d purchase=%s",
			e.UserUUID, e.ItemUUID, e.Price, e.PurchaseUUID))
	case kafka.TopicPurchaseFulfilled:
		e, err := events.UnmarshalPurchaseFulfilled(msg.Value)
		if err != nil {
			return err
		}
		return n.notify(ctx, fmt.Sprintf("PurchaseFulfilled: user=%s item=%s purchase=%s",
			e.UserUUID, e.ItemUUID, e.PurchaseUUID))
	default:
		n.log.Info("ignored topic", "topic", msg.Topic)
		return nil
	}
}

func (n *Notifier) notify(ctx context.Context, text string) error {
	n.log.Info("notification", "text", text)
	if n.token == "" || n.chatID == "" {
		return nil
	}
	body, _ := json.Marshal(map[string]string{
		"chat_id": n.chatID,
		"text":    text,
	})
	url := "https://api.telegram.org/bot" + n.token + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d", resp.StatusCode)
	}
	return nil
}
