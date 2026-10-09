package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
)

type InboxStore interface {
	Insert(ctx context.Context, n *model.Notification) (string, error)
	List(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*model.Notification, int64, int64, error)
	MarkRead(ctx context.Context, userID, notificationID string) (int64, error)
}

type InboxService struct {
	repo InboxStore
	log  *slog.Logger
}

func NewInbox(repo InboxStore, log *slog.Logger) *InboxService {
	if log == nil {
		log = slog.Default()
	}
	return &InboxService{repo: repo, log: log}
}

func (s *InboxService) List(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*model.Notification, int64, int64, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, 0, 0, model.ErrInvalidInput
	}
	return s.repo.List(ctx, userID, limit, offset, unreadOnly)
}

func (s *InboxService) MarkRead(ctx context.Context, userID, notificationID string) (int64, error) {
	if strings.TrimSpace(userID) == "" {
		return 0, model.ErrInvalidInput
	}
	return s.repo.MarkRead(ctx, userID, notificationID)
}

// HandleSubmitted notifies each admin about a new author application.
func (s *InboxService) HandleSubmitted(ctx context.Context, applicationID, applicantID, displayName string, adminIDs []string) error {
	if len(adminIDs) == 0 {
		s.log.Warn("author.application.submitted: no admin_ids", "application_id", applicationID)
		return nil
	}
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "пользователь"
	}
	title := "Новая заявка автора"
	body := fmt.Sprintf("%s подал(а) заявку на роль автора.", name)
	link := "/admin"
	for _, adminID := range adminIDs {
		adminID = strings.TrimSpace(adminID)
		if adminID == "" {
			continue
		}
		src := fmt.Sprintf("author.application.submitted:%s:%s", applicationID, adminID)
		if _, err := s.repo.Insert(ctx, &model.Notification{
			UserID: adminID,
			Title:  title,
			Body:   body,
			Link:   link,
			Source: src,
		}); err != nil {
			return err
		}
	}
	return nil
}

// HandleApproved notifies the applicant that their application was approved.
func (s *InboxService) HandleApproved(ctx context.Context, applicationID, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.ErrInvalidInput
	}
	src := fmt.Sprintf("author.application.approved:%s", applicationID)
	_, err := s.repo.Insert(ctx, &model.Notification{
		UserID: userID,
		Title:  "Заявка одобрена",
		Body:   "Ваша заявка на роль автора одобрена. Добро пожаловать!",
		Link:   "/author/dashboard",
		Source: src,
	})
	return err
}

// HandleRecordBeaten notifies the previous #1 that someone took their place.
func (s *InboxService) HandleRecordBeaten(ctx context.Context, beatenUserID, newNickname, gameID string, level, newScore int) error {
	beatenUserID = strings.TrimSpace(beatenUserID)
	if beatenUserID == "" || strings.TrimSpace(gameID) == "" {
		return model.ErrInvalidInput
	}
	nick := strings.TrimSpace(newNickname)
	if nick == "" {
		nick = "игрок"
	}
	link := fmt.Sprintf("/leaderboard?game=%s&level=%d&highlight=%s", gameID, level, beatenUserID)
	src := fmt.Sprintf("leaderboard.record_beaten:%s:%d:%s:%d", gameID, level, beatenUserID, newScore)
	_, err := s.repo.Insert(ctx, &model.Notification{
		UserID: beatenUserID,
		Title:  "Рекорд побит",
		Body:   fmt.Sprintf("%s обошёл(ла) тебя в %s (счёт %d). Открыть лидерборд.", nick, gameID, newScore),
		Link:   link,
		Source: src,
	})
	return err
}
