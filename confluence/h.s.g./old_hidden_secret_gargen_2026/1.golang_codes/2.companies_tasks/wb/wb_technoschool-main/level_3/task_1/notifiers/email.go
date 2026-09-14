package notifiers

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/wb_technoschool/level_3/task_1/models"
)

// EmailNotifier отправляет уведомления по email
type EmailNotifier struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewEmailNotifier(host string, port int, user, password, from string) *EmailNotifier {
	return &EmailNotifier{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (n *EmailNotifier) Send(ctx context.Context, notification *models.Notification) error {
	// Если не настроен SMTP, возвращаем ошибку
	if n.user == "" || n.password == "" {
		return fmt.Errorf("SMTP not configured, cannot send email")
	}

	subject := notification.Subject
	if subject == "" {
		subject = "Notification"
	}

	// Формируем email
	msg := []string{
		fmt.Sprintf("From: %s", n.from),
		fmt.Sprintf("To: %s", notification.Recipient),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		fmt.Sprintf("<html><body><h2>%s</h2><p>%s</p></body></html>", subject, notification.Message),
	}

	// Аутентификация
	auth := smtp.PlainAuth("", n.user, n.password, n.host)

	// Отправка
	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	err := smtp.SendMail(
		addr,
		auth,
		n.from,
		[]string{notification.Recipient},
		[]byte(strings.Join(msg, "\r\n")),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (n *EmailNotifier) ChannelType() models.NotificationChannel {
	return models.ChannelEmail
}

