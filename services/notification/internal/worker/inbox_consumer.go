package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
)

const (
	SubjectApplicationSubmitted = "author.application.submitted"
	SubjectApplicationApproved  = "author.application.approved"
	SubjectRecordBeaten         = "event.leaderboard.record_beaten"
)

type submittedEvent struct {
	Event          string   `json:"event"`
	ApplicationID  string   `json:"application_id"`
	UserID         string   `json:"user_id"`
	DisplayName    string   `json:"display_name"`
	AdminIDs       []string `json:"admin_ids"`
}

type approvedEvent struct {
	Event         string `json:"event"`
	ApplicationID string `json:"application_id"`
	UserID        string `json:"user_id"`
}

type recordBeatenEvent struct {
	Event          string `json:"event"`
	BeatenUserID   string `json:"beaten_user_id"`
	NewNickname    string `json:"new_nickname"`
	GameID         string `json:"game_id"`
	Level          int    `json:"level"`
	NewScore       int    `json:"new_score"`
}

type InboxConsumer struct {
	js    nats.JetStreamContext
	inbox *service.InboxService
	log   *slog.Logger
}

func NewInboxConsumer(js nats.JetStreamContext, inbox *service.InboxService, log *slog.Logger) *InboxConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &InboxConsumer{js: js, inbox: inbox, log: log}
}

func (c *InboxConsumer) Start(ctx context.Context) {
	subjects := []string{SubjectApplicationSubmitted, SubjectApplicationApproved, SubjectRecordBeaten}
	for _, subj := range subjects {
		s := subj
		durable := "notification-inbox-" + strings.ReplaceAll(s, ".", "-")
		var err error
		for attempt := 0; attempt < 30; attempt++ {
			if ctx.Err() != nil {
				return
			}
			_, err = c.js.Subscribe(s, func(msg *nats.Msg) {
				if err := c.handle(ctx, msg.Subject, msg.Data); err != nil {
					c.log.Error("inbox handle", "subject", msg.Subject, "err", err)
					_ = msg.Nak()
					return
				}
				_ = msg.Ack()
			}, nats.Durable(durable), nats.ManualAck())
			if err == nil {
				c.log.Info("inbox consuming NATS", "subject", s, "durable", durable)
				break
			}
			c.log.Warn("inbox subscribe retry", "subject", s, "attempt", attempt+1, "err", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
		if err != nil {
			c.log.Error("inbox subscribe failed", "subject", s, "err", err)
		}
	}
	<-ctx.Done()
}

func (c *InboxConsumer) handle(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case SubjectApplicationSubmitted:
		var e submittedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		return c.inbox.HandleSubmitted(ctx, e.ApplicationID, e.UserID, e.DisplayName, e.AdminIDs)
	case SubjectApplicationApproved:
		var e approvedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		return c.inbox.HandleApproved(ctx, e.ApplicationID, e.UserID)
	case SubjectRecordBeaten:
		var e recordBeatenEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return err
		}
		level := e.Level
		if level < 1 {
			level = 1
		}
		return c.inbox.HandleRecordBeaten(ctx, e.BeatenUserID, e.NewNickname, e.GameID, level, e.NewScore)
	default:
		c.log.Info("inbox ignored subject", "subject", subject)
		return nil
	}
}
