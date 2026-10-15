[denismatveev@c0der event_horizon]$ find services/notification -type f -name '*.go' | sort
echo "===== main ====="
cat services/notification/cmd/main.go
echo "===== app ====="
for f in services/notification/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/notification/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/notification/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/notification/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/notification/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/notification/cmd/main.go
services/notification/internal/app/app.go
services/notification/internal/app/di.go
services/notification/internal/config/config.go
services/notification/internal/config/config_test.go
services/notification/internal/handler/grpc_handler.go
services/notification/internal/interceptor/interceptors.go
services/notification/internal/model/notification.go
services/notification/internal/repository/repo.go
services/notification/internal/service/inbox.go
services/notification/internal/service/inbox_test.go
services/notification/internal/service/notify.go
services/notification/internal/service/notify_test.go
services/notification/internal/worker/inbox_consumer.go
services/notification/migrations/embed.go
services/notification/proto/notification_grpc.pb.go
services/notification/proto/notification.pb.go
services/notification/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/notification/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init notification app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("notification stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/notification/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/notification/internal/config"
)

type App struct {
	cfg           *config.Config
	log           *slog.Logger
	closer        *closer.Closer
	di            *diContainer
	grpcServer    *grpc.Server
	metricsServer *http.Server
	listener      net.Listener
}

func New(ctx context.Context) (*App, error) {
	cfg := config.Load()
	a := &App{cfg: cfg, di: newDiContainer(cfg)}
	if err := a.init(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("notification gRPC listening", "port", a.cfg.GRPCPort)
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.log.Error("grpc serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func (a *App) Close(ctx context.Context) error { return a.closer.CloseAll(ctx) }
--- services/notification/internal/app/di.go ---
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/services/notification/internal/config"
	"github.com/Eastwesser/event-horizon/services/notification/internal/handler"
	"github.com/Eastwesser/event-horizon/services/notification/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/notification/internal/repository"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
	"github.com/Eastwesser/event-horizon/services/notification/internal/worker"
	"github.com/Eastwesser/event-horizon/services/notification/migrations"
	pb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

type diContainer struct {
	cfg      *config.Config
	dbpool   *pgxpool.Pool
	js       nats.JetStreamContext
	nc       *nats.Conn
	repo     *repository.PostgresRepo
	inbox    *service.InboxService
	notifier *service.Notifier
	api      pb.NotificationServiceServer
	consumer kafka.Consumer
}

func newDiContainer(cfg *config.Config) *diContainer { return &diContainer{cfg: cfg} }

func (a *App) init(ctx context.Context) error {
	for _, step := range []func(context.Context) error{
		a.initLogger, a.initCloser, a.initPostgres, a.initNATS, a.initDomain, a.initWorkers, a.initGRPC, a.initMetricsHTTP,
	} {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	level := slog.LevelInfo
	switch strings.ToLower(a.cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	var h slog.Handler
	if strings.EqualFold(a.cfg.LogFormat, "json") {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	a.log = slog.New(h)
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(a.cfg.DSN())
	if err != nil {
		return err
	}
	poolCfg.MaxConns, poolCfg.MinConns, poolCfg.MaxConnLifetime = 25, 10, 5*time.Minute
	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return err
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return err
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(db), migrations.FS); err != nil {
		db.Close()
		return fmt.Errorf("migrations: %w", err)
	}
	a.di.dbpool = db
	a.closer.AddNamed("postgres", func(context.Context) error { db.Close(); return nil })
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	var nc *nats.Conn
	var err error
	for i := 0; i < 20; i++ {
		nc, err = nats.Connect(a.cfg.NATSURL)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return err
	}
	a.di.nc, a.di.js = nc, js
	a.closer.AddNamed("nats", func(context.Context) error { return nc.Drain() })
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.repo = repository.NewPostgresRepo(a.di.dbpool)
	a.di.inbox = service.NewInbox(a.di.repo, a.log)
	a.di.notifier = service.NewNotifier(a.log, a.cfg.TelegramToken, a.cfg.TelegramChatID)
	a.di.api = handler.NewGRPCHandler(a.di.inbox)

	kcfg := kafka.LoadConfig()
	if a.cfg.KafkaBrokers != "" {
		parts := strings.Split(a.cfg.KafkaBrokers, ",")
		var brokers []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				brokers = append(brokers, p)
			}
		}
		kcfg = kafka.Config{Brokers: brokers, Enabled: len(brokers) > 0}
	}
	a.di.consumer = kafka.NewConsumer(kcfg, kafka.ConsumerGroupNotify,
		[]string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}, a.log)
	a.closer.AddNamed("kafka consumer", func(context.Context) error { return a.di.consumer.Close() })
	return nil
}

func (a *App) initWorkers(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	go worker.NewInboxConsumer(a.di.js, a.di.inbox, a.log).Start(runCtx)
	if a.di.js != nil {
		go a.consumePurchaseNATS(runCtx)
	}
	if a.di.js == nil && a.cfg.KafkaBrokers != "" {
		go func() {
			a.log.Info("notification consuming Kafka purchase topics")
			if err := a.di.consumer.Consume(runCtx, a.di.notifier.Handle); err != nil && runCtx.Err() == nil {
				a.log.Error("kafka consume stopped", "err", err)
			}
		}()
	}
	a.closer.AddNamed("workers", func(context.Context) error { cancel(); return nil })
	_ = ctx
	return nil
}

func (a *App) consumePurchaseNATS(ctx context.Context) {
	subjects := []string{kafka.TopicPurchasePaid, kafka.TopicPurchaseFulfilled}
	for _, subj := range subjects {
		s := subj
		durable := "notification-" + strings.ReplaceAll(s, ".", "-")
		_, err := a.di.js.Subscribe(s, func(msg *nats.Msg) {
			err := a.di.notifier.Handle(ctx, kafka.Message{
				Topic: msg.Subject,
				Value: msg.Data,
			})
			if err != nil {
				a.log.Error("notify handle", "subject", msg.Subject, "err", err)
				_ = msg.Nak()
				return
			}
			_ = msg.Ack()
		}, nats.Durable(durable), nats.ManualAck())
		if err != nil {
			a.log.Error("nats subscribe failed", "subject", s, "err", err)
		} else {
			a.log.Info("notification consuming NATS", "subject", s, "durable", durable)
		}
	}
	<-ctx.Done()
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return err
	}
	a.listener = lis
	a.closer.AddNamed("tcp", func(context.Context) error { return lis.Close() })
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.Recovery(), interceptor.Logger(), interceptor.Validate()))
	pb.RegisterNotificationServiceServer(s, a.di.api)
	reflection.Register(s)
	a.grpcServer = s
	a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "notification"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.di.dbpool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "notification"})
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to Event Horizon notifications bot",
		})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() { _ = srv.ListenAndServe() }()
	a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
===== handler =====
--- services/notification/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

type GRPCHandler struct {
	pb.UnimplementedNotificationServiceServer
	inbox *service.InboxService
}

func NewGRPCHandler(inbox *service.InboxService) *GRPCHandler {
	return &GRPCHandler{inbox: inbox}
}

func (h *GRPCHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	list, total, unread, err := h.inbox.List(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()), req.GetUnreadOnly())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Notification, 0, len(list))
	for _, n := range list {
		var readUnix int64
		if n.ReadAt != nil {
			readUnix = n.ReadAt.Unix()
		}
		out = append(out, &pb.Notification{
			Id:             n.ID,
			UserId:         n.UserID,
			Title:          n.Title,
			Body:           n.Body,
			Link:           n.Link,
			CreatedAtUnix:  n.CreatedAt.Unix(),
			ReadAtUnix:     readUnix,
		})
	}
	return &pb.ListNotificationsResponse{
		Notifications: out,
		Total:         total,
		UnreadCount:   unread,
	}, nil
}

func (h *GRPCHandler) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	n, err := h.inbox.MarkRead(ctx, req.GetUserId(), req.GetNotificationId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.MarkReadResponse{Marked: n}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
===== service =====
--- services/notification/internal/service/inbox.go ---
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
--- services/notification/internal/service/inbox_test.go ---
package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
)

type memStore struct {
	items []*model.Notification
}

func (m *memStore) Insert(_ context.Context, n *model.Notification) (string, error) {
	for _, x := range m.items {
		if x.UserID == n.UserID && x.Source == n.Source && n.Source != "" {
			return x.ID, nil
		}
	}
	if n.ID == "" {
		n.ID = "id-" + n.Source
	}
	cp := *n
	m.items = append(m.items, &cp)
	return n.ID, nil
}

func (m *memStore) List(_ context.Context, userID string, _, _ int, unreadOnly bool) ([]*model.Notification, int64, int64, error) {
	var out []*model.Notification
	var unread int64
	for _, n := range m.items {
		if n.UserID != userID {
			continue
		}
		if n.ReadAt == nil {
			unread++
		}
		if unreadOnly && n.ReadAt != nil {
			continue
		}
		out = append(out, n)
	}
	return out, int64(len(out)), unread, nil
}

func (m *memStore) MarkRead(_ context.Context, userID, id string) (int64, error) {
	var n int64
	for _, x := range m.items {
		if x.UserID != userID {
			continue
		}
		if id != "" && x.ID != id {
			continue
		}
		if x.ReadAt == nil {
			t := x.CreatedAt
			x.ReadAt = &t
			n++
		}
	}
	return n, nil
}

func TestHandleSubmitted_CreatesPerAdmin(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleSubmitted(context.Background(), "app1", "u1", "Иван", []string{"a1", "a2", ""}); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 2 {
		t.Fatalf("want 2 rows, got %d", len(st.items))
	}
}

func TestHandleApproved(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleApproved(context.Background(), "app1", "u1"); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 1 || st.items[0].Link != "/author/dashboard" {
		t.Fatalf("unexpected: %+v", st.items)
	}
}

func TestHandleSubmitted_NoAdmins(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleSubmitted(context.Background(), "app1", "u1", "X", nil); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 0 {
		t.Fatalf("expected empty")
	}
}
--- services/notification/internal/service/notify.go ---
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
--- services/notification/internal/service/notify_test.go ---
package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
)

func TestHandle_IgnoredTopic(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	if err := n.Handle(context.Background(), kafka.Message{Topic: "unrelated"}); err != nil {
		t.Fatalf("ignored topic: %v", err)
	}
}

func TestHandle_PurchasePaid_NoTelegram(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	body, err := events.PurchasePaid{
		EventUUID:    "e1",
		PurchaseUUID: "p1",
		UserUUID:     "u1",
		ItemUUID:     "i1",
		Price:        10,
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := n.Handle(context.Background(), kafka.Message{
		Topic: kafka.TopicPurchasePaid,
		Value: body,
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}
}

func TestHandle_PurchasePaid_BadJSON(t *testing.T) {
	n := NewNotifier(slog.Default(), "", "")
	err := n.Handle(context.Background(), kafka.Message{
		Topic: kafka.TopicPurchasePaid,
		Value: []byte("not-json"),
	})
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}
===== repository =====
--- services/notification/internal/repository/repo.go ---
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Insert(ctx context.Context, n *model.Notification) (string, error) {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (id, user_id, title, body, link, source_event, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (user_id, source_event) WHERE source_event <> '' DO NOTHING`,
		n.ID, n.UserID, n.Title, n.Body, n.Link, n.Source, n.CreatedAt)
	return n.ID, err
}

func (r *PostgresRepo) List(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*model.Notification, int64, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var unread int64
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&unread); err != nil {
		return nil, 0, 0, err
	}
	var total int64
	countQ := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	if unreadOnly {
		countQ += ` AND read_at IS NULL`
	}
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	q := `
		SELECT id, user_id, title, body, link, source_event, created_at, read_at
		FROM notifications WHERE user_id = $1`
	if unreadOnly {
		q += ` AND read_at IS NULL`
	}
	q += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	var out []*model.Notification
	for rows.Next() {
		var n model.Notification
		var readAt *time.Time
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Link, &n.Source, &n.CreatedAt, &readAt); err != nil {
			return nil, 0, 0, err
		}
		n.ReadAt = readAt
		out = append(out, &n)
	}
	return out, total, unread, rows.Err()
}

func (r *PostgresRepo) MarkRead(ctx context.Context, userID, notificationID string) (int64, error) {
	now := time.Now().UTC()
	if notificationID == "" {
		tag, err := r.db.Exec(ctx, `
			UPDATE notifications SET read_at = $1
			WHERE user_id = $2 AND read_at IS NULL`, now, userID)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE notifications SET read_at = $1
		WHERE id = $2 AND user_id = $3 AND read_at IS NULL`, now, notificationID, userID)
	if err != nil {
		return 0, err
	}
	n := tag.RowsAffected()
	if n == 0 {
		var exists bool
		if err := r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE id=$1 AND user_id=$2)`,
			notificationID, userID).Scan(&exists); err != nil {
			return 0, err
		}
		if !exists {
			return 0, model.ErrNotFound
		}
	}
	return n, nil
}
===== worker =====
--- services/notification/internal/worker/inbox_consumer.go ---
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
[denismatveev@c0der event_horizon]$ 
