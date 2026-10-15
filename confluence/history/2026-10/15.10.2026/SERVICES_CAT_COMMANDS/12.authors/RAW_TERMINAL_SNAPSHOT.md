[denismatveev@c0der event_horizon]$ find services/authors -type f -name '*.go' | sort
echo "===== main ====="
cat services/authors/cmd/main.go
echo "===== app ====="
for f in services/authors/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/authors/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/authors/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/authors/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== validate ====="
for f in services/authors/internal/validate/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/authors/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/authors/cmd/main.go
services/authors/internal/app/app.go
services/authors/internal/app/di.go
services/authors/internal/config/config.go
services/authors/internal/handler/grpc_handler.go
services/authors/internal/interceptor/interceptors.go
services/authors/internal/model/author.go
services/authors/internal/repository/postgres_integration_test.go
services/authors/internal/repository/repo.go
services/authors/internal/service/authors_service.go
services/authors/internal/service/authors_service_test.go
services/authors/internal/validate/validate_test.go
services/authors/internal/worker/outbox_worker.go
services/authors/migrations/embed.go
services/authors/proto/authors_grpc.pb.go
services/authors/proto/authors.pb.go
services/authors/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/authors/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init authors app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("authors stopped: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/authors/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/authors/internal/config"
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
		a.log.Info("authors gRPC listening", "port", a.cfg.GRPCPort)
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
--- services/authors/internal/app/di.go ---
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
	"github.com/Eastwesser/event-horizon/services/authors/internal/config"
	"github.com/Eastwesser/event-horizon/services/authors/internal/handler"
	"github.com/Eastwesser/event-horizon/services/authors/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/authors/internal/repository"
	"github.com/Eastwesser/event-horizon/services/authors/internal/service"
	"github.com/Eastwesser/event-horizon/services/authors/internal/worker"
	"github.com/Eastwesser/event-horizon/services/authors/migrations"
	pb "github.com/Eastwesser/event-horizon/services/authors/proto"
)

type diContainer struct {
	cfg    *config.Config
	dbpool *pgxpool.Pool
	cache  *repository.RedisRepo
	js     nats.JetStreamContext
	nc     *nats.Conn
	repo   *repository.PostgresRepo
	svc    *service.AuthorsService
	api    pb.AuthorsServiceServer
}

func newDiContainer(cfg *config.Config) *diContainer { return &diContainer{cfg: cfg} }

func (a *App) init(ctx context.Context) error {
	for _, step := range []func(context.Context) error{
		a.initLogger, a.initCloser, a.initPostgres, a.initRedis, a.initNATS, a.initDomain, a.initOutbox, a.initGRPC, a.initMetricsHTTP,
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

func (a *App) initRedis(ctx context.Context) error {
	cache := repository.NewRedisRepo(a.cfg.RedisAddr, 30*time.Minute)
	if err := cache.Ping(ctx); err != nil {
		a.log.Warn("redis degraded", "err", err)
	}
	a.di.cache = cache
	a.closer.AddNamed("redis", func(context.Context) error { return cache.Close() })
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
	a.di.svc = service.New(a.di.repo, a.di.cache)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initOutbox(_ context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	go worker.NewOutboxWorker(a.di.dbpool, a.di.js).Start(ctx)
	a.closer.AddNamed("outbox", func(context.Context) error { cancel(); return nil })
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCPort)
	if err != nil {
		return err
	}
	a.listener = lis
	a.closer.AddNamed("tcp", func(context.Context) error { return lis.Close() })
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.Recovery(), interceptor.Logger(), interceptor.Validate()))
	pb.RegisterAuthorsServiceServer(s, a.di.api)
	reflection.Register(s)
	a.grpcServer = s
	a.closer.AddNamed("grpc", func(context.Context) error { s.GracefulStop(); return nil })
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "authors"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.di.dbpool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "authors"})
	})
	srv := &http.Server{Addr: ":" + a.cfg.MetricsPort, Handler: mux}
	a.metricsServer = srv
	go func() { _ = srv.ListenAndServe() }()
	a.closer.AddNamed("metrics", func(ctx context.Context) error { return srv.Shutdown(ctx) })
	return nil
}
===== handler =====
--- services/authors/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
	"github.com/Eastwesser/event-horizon/services/authors/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/authors/proto"
)

type GRPCHandler struct {
	pb.UnimplementedAuthorsServiceServer
	svc *service.AuthorsService
}

func NewGRPCHandler(svc *service.AuthorsService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func toPB(a *model.Author) *pb.Author {
	if a == nil {
		return nil
	}
	out := &pb.Author{
		Id: a.ID, UserId: a.UserID, DisplayName: a.DisplayName, Bio: a.Bio,
		AvatarUrl: a.AvatarURL, Portfolio: a.Portfolio, Active: a.Active,
		CreatedAtUnix: a.CreatedAt.Unix(), UpdatedAtUnix: a.UpdatedAt.Unix(),
	}
	if a.VerifiedAt != nil {
		out.VerifiedAtUnix = a.VerifiedAt.Unix()
	}
	return out
}

func applicationToPB(a *model.AuthorApplication) *pb.AuthorApplication {
	if a == nil {
		return nil
	}
	out := &pb.AuthorApplication{
		Id:            a.ID,
		UserId:        a.UserID,
		Status:        string(a.Status),
		DisplayName:   a.Payload.DisplayName,
		Portfolio:     a.Payload.Portfolio,
		Motivation:    a.Payload.Motivation,
		ContactEmail:  a.Payload.ContactEmail,
		CreatedAtUnix: a.CreatedAt.Unix(),
		ReviewedBy:    a.ReviewedBy,
		ReviewerNote:  a.ReviewerNote,
	}
	if a.ReviewedAt != nil {
		out.ReviewedAtUnix = a.ReviewedAt.Unix()
	}
	return out
}

func (h *GRPCHandler) UpsertProfile(ctx context.Context, req *pb.UpsertProfileRequest) (*pb.UpsertProfileResponse, error) {
	a, err := h.svc.UpsertProfile(ctx, req.GetUserId(), req.GetDisplayName(), req.GetBio(), req.GetAvatarUrl(), req.GetPortfolio())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.UpsertProfileResponse{Author: toPB(a)}, nil
}

func (h *GRPCHandler) GetAuthor(ctx context.Context, req *pb.GetAuthorRequest) (*pb.GetAuthorResponse, error) {
	a, err := h.svc.GetAuthor(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetAuthorResponse{Author: toPB(a)}, nil
}

func (h *GRPCHandler) ListAuthors(ctx context.Context, req *pb.ListAuthorsRequest) (*pb.ListAuthorsResponse, error) {
	list, total, err := h.svc.ListAuthors(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Author, 0, len(list))
	for _, a := range list {
		out = append(out, toPB(a))
	}
	return &pb.ListAuthorsResponse{Authors: out, Total: total}, nil
}

func (h *GRPCHandler) SubmitApplication(ctx context.Context, req *pb.SubmitApplicationRequest) (*pb.SubmitApplicationResponse, error) {
	app, err := h.svc.SubmitApplication(
		ctx,
		req.GetUserId(),
		req.GetCallerRole(),
		req.GetDisplayName(),
		req.GetPortfolio(),
		req.GetMotivation(),
		req.GetContactEmail(),
	)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.SubmitApplicationResponse{Application: applicationToPB(app)}, nil
}

func (h *GRPCHandler) GetMyApplication(ctx context.Context, req *pb.GetMyApplicationRequest) (*pb.GetMyApplicationResponse, error) {
	app, err := h.svc.GetMyApplication(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetMyApplicationResponse{Application: applicationToPB(app)}, nil
}

func (h *GRPCHandler) ListApplications(ctx context.Context, req *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	list, total, err := h.svc.ListApplications(ctx, req.GetStatus(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.AuthorApplication, 0, len(list))
	for _, a := range list {
		out = append(out, applicationToPB(a))
	}
	return &pb.ListApplicationsResponse{Applications: out, Total: total}, nil
}

func (h *GRPCHandler) ApproveApplication(ctx context.Context, req *pb.ApproveApplicationRequest) (*pb.ApproveApplicationResponse, error) {
	app, author, err := h.svc.ApproveApplication(ctx, req.GetApplicationId(), req.GetReviewerId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.ApproveApplicationResponse{
		Application: applicationToPB(app),
		Author:      toPB(author),
	}, nil
}

func (h *GRPCHandler) RejectApplication(ctx context.Context, req *pb.RejectApplicationRequest) (*pb.RejectApplicationResponse, error) {
	app, err := h.svc.RejectApplication(ctx, req.GetApplicationId(), req.GetReviewerId(), req.GetReviewerNote())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.RejectApplicationResponse{Application: applicationToPB(app)}, nil
}

func (h *GRPCHandler) RevertApplication(ctx context.Context, req *pb.RevertApplicationRequest) (*pb.RevertApplicationResponse, error) {
	app, err := h.svc.RevertApplication(ctx, req.GetApplicationId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.RevertApplicationResponse{Application: applicationToPB(app)}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound), errors.Is(err, model.ErrApplicationMissing):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidInput),
		errors.Is(err, model.ErrAlreadyPrivileged),
		errors.Is(err, model.ErrAlreadyReviewed):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
===== service =====
--- services/authors/internal/service/authors_service.go ---
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
)

// AuthorStore is the persistence port used by AuthorsService (Postgres in prod).
type AuthorStore interface {
	Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error
	GetByUserID(ctx context.Context, userID string) (*model.Author, error)
	List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error)
	GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error)
	GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error)
	InsertApplication(ctx context.Context, a *model.AuthorApplication) error
	ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error)
	ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error)
	RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error)
	RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error)
}

// AuthorCache is the optional Redis port for author profiles.
type AuthorCache interface {
	Get(ctx context.Context, userID string) (*model.Author, error)
	Set(ctx context.Context, a *model.Author) error
	Delete(ctx context.Context, userID string) error
}

type AuthorsService struct {
	repo  AuthorStore
	cache AuthorCache
}

func New(repo AuthorStore, cache AuthorCache) *AuthorsService {
	return &AuthorsService{repo: repo, cache: cache}
}

func (s *AuthorsService) UpsertProfile(ctx context.Context, userID, displayName, bio, avatar, portfolio string) (*model.Author, error) {
	if userID == "" || displayName == "" {
		return nil, model.ErrInvalidInput
	}
	a := &model.Author{
		UserID:      userID,
		DisplayName: displayName,
		Bio:         bio,
		AvatarURL:   avatar,
		Portfolio:   portfolio,
		Active:      true,
	}
	event := map[string]any{
		"event":        "author.upserted",
		"user_id":      userID,
		"display_name": displayName,
		"timestamp":    time.Now().Unix(),
	}
	if err := s.repo.Upsert(ctx, a, "author.upserted", event); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, a)
	}
	return a, nil
}

func (s *AuthorsService) GetAuthor(ctx context.Context, userID string) (*model.Author, error) {
	if s.cache != nil {
		if a, err := s.cache.Get(ctx, userID); err == nil {
			return a, nil
		}
	}
	a, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, a)
	}
	return a, nil
}

func (s *AuthorsService) ListAuthors(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *AuthorsService) SubmitApplication(
	ctx context.Context,
	userID, callerRole, displayName, portfolio, motivation, contactEmail string,
) (*model.AuthorApplication, error) {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	portfolio = strings.TrimSpace(portfolio)
	motivation = strings.TrimSpace(motivation)
	contactEmail = strings.TrimSpace(contactEmail)
	role := strings.ToLower(strings.TrimSpace(callerRole))

	if userID == "" || displayName == "" || motivation == "" || contactEmail == "" {
		return nil, model.ErrInvalidInput
	}
	if role == "author" || role == "admin" {
		return nil, model.ErrAlreadyPrivileged
	}

	if pending, err := s.repo.GetPendingApplicationByUserID(ctx, userID); err == nil {
		return pending, nil
	} else if !errors.Is(err, model.ErrApplicationMissing) {
		return nil, err
	}

	app := &model.AuthorApplication{
		ID:     uuid.NewString(),
		UserID: userID,
		Status: model.ApplicationPending,
		Payload: model.ApplicationPayload{
			DisplayName:  displayName,
			Portfolio:    portfolio,
			Motivation:   motivation,
			ContactEmail: contactEmail,
		},
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.InsertApplication(ctx, app); err != nil {
		// Race: another pending insert — return existing.
		if pending, getErr := s.repo.GetPendingApplicationByUserID(ctx, userID); getErr == nil {
			return pending, nil
		}
		return nil, err
	}
	return app, nil
}

func (s *AuthorsService) GetMyApplication(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, model.ErrInvalidInput
	}
	return s.repo.GetLatestApplicationByUserID(ctx, userID)
}

func (s *AuthorsService) ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", "pending", "approved", "rejected":
	default:
		return nil, 0, model.ErrInvalidInput
	}
	return s.repo.ListApplications(ctx, status, limit, offset)
}

func (s *AuthorsService) ApproveApplication(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
	applicationID = strings.TrimSpace(applicationID)
	reviewerID = strings.TrimSpace(reviewerID)
	if applicationID == "" || reviewerID == "" {
		return nil, nil, model.ErrInvalidInput
	}
	app, author, err := s.repo.ApproveApplicationInTx(ctx, applicationID, reviewerID)
	if err != nil {
		return nil, nil, err
	}
	if s.cache != nil && author != nil {
		_ = s.cache.Set(ctx, author)
	}
	return app, author, nil
}

func (s *AuthorsService) RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
	applicationID = strings.TrimSpace(applicationID)
	reviewerID = strings.TrimSpace(reviewerID)
	note = strings.TrimSpace(note)
	if applicationID == "" || reviewerID == "" {
		return nil, model.ErrInvalidInput
	}
	return s.repo.RejectApplication(ctx, applicationID, reviewerID, note)
}

func (s *AuthorsService) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
	applicationID = strings.TrimSpace(applicationID)
	if applicationID == "" {
		return nil, model.ErrInvalidInput
	}
	app, err := s.repo.RevertApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, app.UserID)
	}
	return app, nil
}
--- services/authors/internal/service/authors_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
)

type mockStore struct {
	upsertFn               func(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error
	getByUserIDFn          func(ctx context.Context, userID string) (*model.Author, error)
	listFn                 func(ctx context.Context, limit, offset int) ([]*model.Author, int64, error)
	getPendingFn           func(ctx context.Context, userID string) (*model.AuthorApplication, error)
	getLatestFn            func(ctx context.Context, userID string) (*model.AuthorApplication, error)
	insertAppFn            func(ctx context.Context, a *model.AuthorApplication) error
	listAppsFn             func(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error)
	approveFn              func(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error)
	rejectFn               func(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error)
	revertFn               func(ctx context.Context, applicationID string) (*model.AuthorApplication, error)
	insertCalls            int
	approveCalls           int
	rejectCalls            int
	revertCalls            int
	upsertCalls            int
}

func (m *mockStore) Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error {
	m.upsertCalls++
	if m.upsertFn != nil {
		return m.upsertFn(ctx, a, eventType, eventPayload)
	}
	return nil
}
func (m *mockStore) GetByUserID(ctx context.Context, userID string) (*model.Author, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, userID)
	}
	return nil, model.ErrNotFound
}
func (m *mockStore) List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockStore) GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	if m.getPendingFn != nil {
		return m.getPendingFn(ctx, userID)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	if m.getLatestFn != nil {
		return m.getLatestFn(ctx, userID)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) InsertApplication(ctx context.Context, a *model.AuthorApplication) error {
	m.insertCalls++
	if m.insertAppFn != nil {
		return m.insertAppFn(ctx, a)
	}
	return nil
}
func (m *mockStore) ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error) {
	if m.listAppsFn != nil {
		return m.listAppsFn(ctx, status, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockStore) ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
	m.approveCalls++
	if m.approveFn != nil {
		return m.approveFn(ctx, applicationID, reviewerID)
	}
	return nil, nil, model.ErrApplicationMissing
}
func (m *mockStore) RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
	m.rejectCalls++
	if m.rejectFn != nil {
		return m.rejectFn(ctx, applicationID, reviewerID, note)
	}
	return nil, model.ErrApplicationMissing
}
func (m *mockStore) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
	m.revertCalls++
	if m.revertFn != nil {
		return m.revertFn(ctx, applicationID)
	}
	return nil, model.ErrApplicationMissing
}

type mockCache struct {
	store  map[string]*model.Author
	gets   int
	sets   int
	dels   int
	getErr error
}

func (c *mockCache) Get(_ context.Context, userID string) (*model.Author, error) {
	c.gets++
	if c.getErr != nil {
		return nil, c.getErr
	}
	if c.store == nil {
		return nil, errors.New("miss")
	}
	a, ok := c.store[userID]
	if !ok {
		return nil, errors.New("miss")
	}
	cp := *a
	return &cp, nil
}
func (c *mockCache) Set(_ context.Context, a *model.Author) error {
	c.sets++
	if c.store == nil {
		c.store = map[string]*model.Author{}
	}
	cp := *a
	c.store[a.UserID] = &cp
	return nil
}
func (c *mockCache) Delete(_ context.Context, userID string) error {
	c.dels++
	if c.store != nil {
		delete(c.store, userID)
	}
	return nil
}

func TestUpsertProfile_Validation(t *testing.T) {
	svc := New(&mockStore{}, nil)
	if _, err := svc.UpsertProfile(context.Background(), "", "Name", "", "", ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty user: %v", err)
	}
	if _, err := svc.UpsertProfile(context.Background(), "u1", "", "", "", ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty name: %v", err)
	}
}

func TestUpsertProfile_SuccessWithPortfolioAndCache(t *testing.T) {
	store := &mockStore{
		upsertFn: func(_ context.Context, a *model.Author, eventType string, _ map[string]any) error {
			if a.UserID != "u1" || a.DisplayName != "Den" || a.Portfolio != "https://gh.com/x" || !a.Active {
				t.Fatalf("bad author: %+v", a)
			}
			if eventType != "author.upserted" {
				t.Fatalf("event=%s", eventType)
			}
			a.ID = "aid-1"
			return nil
		},
	}
	cache := &mockCache{}
	svc := New(store, cache)
	got, err := svc.UpsertProfile(context.Background(), "u1", "Den", "bio", "av", "https://gh.com/x")
	if err != nil {
		t.Fatal(err)
	}
	if got.Portfolio != "https://gh.com/x" || store.upsertCalls != 1 || cache.sets != 1 {
		t.Fatalf("got=%+v upsert=%d sets=%d", got, store.upsertCalls, cache.sets)
	}
}

func TestGetAuthor_CacheHit(t *testing.T) {
	cache := &mockCache{store: map[string]*model.Author{
		"u1": {UserID: "u1", DisplayName: "cached"},
	}}
	store := &mockStore{
		getByUserIDFn: func(context.Context, string) (*model.Author, error) {
			t.Fatal("repo should not be called on cache hit")
			return nil, nil
		},
	}
	got, err := New(store, cache).GetAuthor(context.Background(), "u1")
	if err != nil || got.DisplayName != "cached" || cache.gets != 1 {
		t.Fatalf("got=%+v err=%v gets=%d", got, err, cache.gets)
	}
}

func TestGetAuthor_CacheMissLoadsRepo(t *testing.T) {
	cache := &mockCache{getErr: errors.New("miss")}
	store := &mockStore{
		getByUserIDFn: func(_ context.Context, userID string) (*model.Author, error) {
			return &model.Author{UserID: userID, DisplayName: "from-db"}, nil
		},
	}
	got, err := New(store, cache).GetAuthor(context.Background(), "u1")
	if err != nil || got.DisplayName != "from-db" || cache.sets != 1 {
		t.Fatalf("got=%+v err=%v sets=%d", got, err, cache.sets)
	}
}

func TestListAuthors_Delegates(t *testing.T) {
	store := &mockStore{
		listFn: func(_ context.Context, limit, offset int) ([]*model.Author, int64, error) {
			if limit != 10 || offset != 5 {
				t.Fatalf("limit=%d offset=%d", limit, offset)
			}
			return []*model.Author{{UserID: "u1"}}, 1, nil
		},
	}
	list, total, err := New(store, nil).ListAuthors(context.Background(), 10, 5)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestSubmitApplication_ValidationAndPrivileged(t *testing.T) {
	svc := New(&mockStore{}, nil)
	_, err := svc.SubmitApplication(context.Background(), "", "user", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("empty user: %v", err)
	}
	_, err = svc.SubmitApplication(context.Background(), "u1", "author", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrAlreadyPrivileged) {
		t.Fatalf("author: %v", err)
	}
	_, err = svc.SubmitApplication(context.Background(), "u1", "ADMIN", "n", "", "why", "a@b.c")
	if !errors.Is(err, model.ErrAlreadyPrivileged) {
		t.Fatalf("admin: %v", err)
	}
}

func TestSubmitApplication_ReturnsExistingPending(t *testing.T) {
	existing := &model.AuthorApplication{ID: "app-1", UserID: "u1", Status: model.ApplicationPending}
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			return existing, nil
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "u1", "user", "Name", "p", "motivation", "a@b.c")
	if err != nil || got.ID != "app-1" || store.insertCalls != 0 {
		t.Fatalf("got=%+v err=%v inserts=%d", got, err, store.insertCalls)
	}
}

func TestSubmitApplication_CreatesNew(t *testing.T) {
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			return nil, model.ErrApplicationMissing
		},
		insertAppFn: func(_ context.Context, a *model.AuthorApplication) error {
			if a.UserID != "u1" || a.Status != model.ApplicationPending {
				t.Fatalf("bad app: %+v", a)
			}
			if a.Payload.DisplayName != "Name" || a.Payload.Motivation != "why me" || a.Payload.ContactEmail != "a@b.c" {
				t.Fatalf("payload: %+v", a.Payload)
			}
			if a.ID == "" {
				t.Fatal("id empty")
			}
			return nil
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "  u1  ", "user", "  Name  ", " pf ", "  why me  ", "  a@b.c  ")
	if err != nil || got.Status != model.ApplicationPending || store.insertCalls != 1 {
		t.Fatalf("got=%+v err=%v inserts=%d", got, err, store.insertCalls)
	}
}

func TestSubmitApplication_InsertRaceReturnsPending(t *testing.T) {
	pending := &model.AuthorApplication{ID: "race-1", UserID: "u1", Status: model.ApplicationPending}
	calls := 0
	store := &mockStore{
		getPendingFn: func(context.Context, string) (*model.AuthorApplication, error) {
			calls++
			if calls == 1 {
				return nil, model.ErrApplicationMissing
			}
			return pending, nil
		},
		insertAppFn: func(context.Context, *model.AuthorApplication) error {
			return errors.New("unique violation")
		},
	}
	got, err := New(store, nil).SubmitApplication(context.Background(), "u1", "user", "Name", "", "motivation", "a@b.c")
	if err != nil || got.ID != "race-1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGetMyApplication_Validation(t *testing.T) {
	_, err := New(&mockStore{}, nil).GetMyApplication(context.Background(), "  ")
	if !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
}

func TestGetMyApplication_Delegates(t *testing.T) {
	want := &model.AuthorApplication{ID: "a1", UserID: "u1"}
	store := &mockStore{
		getLatestFn: func(_ context.Context, userID string) (*model.AuthorApplication, error) {
			if userID != "u1" {
				t.Fatalf("user=%s", userID)
			}
			return want, nil
		},
	}
	got, err := New(store, nil).GetMyApplication(context.Background(), " u1 ")
	if err != nil || got.ID != "a1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestListApplications_StatusFilter(t *testing.T) {
	svc := New(&mockStore{
		listAppsFn: func(_ context.Context, status string, _, _ int) ([]*model.AuthorApplication, int64, error) {
			return []*model.AuthorApplication{{ID: "1", Status: model.ApplicationStatus(status)}}, 1, nil
		},
	}, nil)
	if _, _, err := svc.ListApplications(context.Background(), "weird", 10, 0); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("want invalid, got %v", err)
	}
	list, total, err := svc.ListApplications(context.Background(), "PENDING", 10, 0)
	if err != nil || total != 1 || list[0].Status != "pending" {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestApproveApplication_ValidationAndSuccess(t *testing.T) {
	svc := New(&mockStore{}, nil)
	if _, _, err := svc.ApproveApplication(context.Background(), "", "r1"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	now := time.Now().UTC()
	author := &model.Author{UserID: "u1", DisplayName: "Den", Active: true}
	app := &model.AuthorApplication{ID: "app-1", UserID: "u1", Status: model.ApplicationApproved, ReviewedAt: &now}
	store := &mockStore{
		approveFn: func(_ context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
			if applicationID != "app-1" || reviewerID != "admin-1" {
				t.Fatalf("%s %s", applicationID, reviewerID)
			}
			return app, author, nil
		},
	}
	cache := &mockCache{}
	gotApp, gotAuthor, err := New(store, cache).ApproveApplication(context.Background(), " app-1 ", " admin-1 ")
	if err != nil || gotApp.Status != model.ApplicationApproved || gotAuthor.UserID != "u1" || cache.sets != 1 {
		t.Fatalf("app=%+v author=%+v err=%v sets=%d", gotApp, gotAuthor, err, cache.sets)
	}
}

func TestApproveApplication_AlreadyReviewed(t *testing.T) {
	store := &mockStore{
		approveFn: func(context.Context, string, string) (*model.AuthorApplication, *model.Author, error) {
			return nil, nil, model.ErrAlreadyReviewed
		},
	}
	_, _, err := New(store, nil).ApproveApplication(context.Background(), "app-1", "admin")
	if !errors.Is(err, model.ErrAlreadyReviewed) {
		t.Fatalf("%v", err)
	}
}

func TestRejectApplication_ValidationAndSuccess(t *testing.T) {
	if _, err := New(&mockStore{}, nil).RejectApplication(context.Background(), "a", "", "note"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatal("expected invalid")
	}
	store := &mockStore{
		rejectFn: func(_ context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
			if applicationID != "app-1" || reviewerID != "admin" || note != "nope" {
				t.Fatalf("%s %s %s", applicationID, reviewerID, note)
			}
			return &model.AuthorApplication{ID: applicationID, Status: model.ApplicationRejected, ReviewerNote: note}, nil
		},
	}
	got, err := New(store, nil).RejectApplication(context.Background(), " app-1 ", " admin ", "  nope  ")
	if err != nil || got.Status != model.ApplicationRejected || store.rejectCalls != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestRevertApplication_ClearsCache(t *testing.T) {
	if _, err := New(&mockStore{}, nil).RevertApplication(context.Background(), ""); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatal("expected invalid")
	}
	store := &mockStore{
		revertFn: func(_ context.Context, id string) (*model.AuthorApplication, error) {
			return &model.AuthorApplication{ID: id, UserID: "u1", Status: model.ApplicationPending}, nil
		},
	}
	cache := &mockCache{store: map[string]*model.Author{"u1": {UserID: "u1"}}}
	got, err := New(store, cache).RevertApplication(context.Background(), "app-1")
	if err != nil || got.Status != model.ApplicationPending || cache.dels != 1 {
		t.Fatalf("got=%+v err=%v dels=%d", got, err, cache.dels)
	}
}
===== repository =====
--- services/authors/internal/repository/postgres_integration_test.go ---
//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
	"github.com/Eastwesser/event-horizon/services/authors/internal/repository"
	"github.com/Eastwesser/event-horizon/services/authors/migrations"
)

func authorsPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("AUTHORS_TEST_DATABASE_URL"); dsn != "" {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			t.Fatalf("connect env dsn: %v", err)
		}
		if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
			pool.Close()
			t.Fatalf("migrate: %v", err)
		}
		return pool, func() { pool.Close() }
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "authors_test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Skipf("testcontainers postgres unavailable (Docker?): %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/authors_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("pool: %v", err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
		pool.Close()
		_ = c.Terminate(ctx)
		t.Fatalf("migrate: %v", err)
	}
	return pool, func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
}

func TestAuthors_UpsertAndList(t *testing.T) {
	pool, cleanup := authorsPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(pool)
	a := &model.Author{
		UserID: "user-1", DisplayName: "Nika", Bio: "paints", AvatarURL: "", Active: true,
	}
	if err := repo.Upsert(ctx, a, "author.upserted", map[string]any{"user_id": a.UserID}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if a.ID == "" {
		t.Fatal("expected id after insert")
	}

	got, err := repo.GetByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.DisplayName != "Nika" {
		t.Fatalf("name=%q", got.DisplayName)
	}

	list, total, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("total=%d len=%d", total, len(list))
	}
}
--- services/authors/internal/repository/repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UTC()
	a.UpdatedAt = now
	var existingID string
	err = tx.QueryRow(ctx, `SELECT id FROM authors WHERE user_id = $1`, a.UserID).Scan(&existingID)
	if errors.Is(err, pgx.ErrNoRows) {
		a.ID = uuid.NewString()
		a.CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO authors (id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			a.ID, a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Portfolio, a.Active, a.CreatedAt, a.UpdatedAt, a.VerifiedAt)
	} else if err != nil {
		return err
	} else {
		a.ID = existingID
		_, err = tx.Exec(ctx, `
			UPDATE authors SET display_name=$2, bio=$3, avatar_url=$4, portfolio=$5, active=$6, updated_at=$7, verified_at=$8
			WHERE user_id=$1`,
			a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Portfolio, a.Active, a.UpdatedAt, a.VerifiedAt)
	}
	if err != nil {
		return err
	}

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		uuid.NewString(), eventType, body)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepo) GetByUserID(ctx context.Context, userID string) (*model.Author, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at
		FROM authors WHERE user_id = $1`, userID)
	return scanAuthor(row)
}

func (r *PostgresRepo) GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications
		WHERE user_id = $1 AND status = 'pending'
		LIMIT 1`, userID)
	return scanApplication(row)
}

func (r *PostgresRepo) GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, userID)
	return scanApplication(row)
}

func (r *PostgresRepo) GetApplicationByID(ctx context.Context, id string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1`, id)
	return scanApplication(row)
}

func (r *PostgresRepo) InsertApplication(ctx context.Context, a *model.AuthorApplication) error {
	body, err := json.Marshal(a.Payload)
	if err != nil {
		return err
	}
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO author_applications
			(id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.UserID, string(a.Status), body, a.CreatedAt, a.ReviewedAt, a.ReviewedBy, a.ReviewerNote)
	return err
}

func (r *PostgresRepo) ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))

	var (
		total int64
		rows  pgx.Rows
		err   error
	)
	if status == "" {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM author_applications`).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
			FROM author_applications
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	} else {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM author_applications WHERE status = $1`, status).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
			FROM author_applications
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`, status, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*model.AuthorApplication, 0)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, app)
	}
	return out, total, rows.Err()
}

// ApproveApplicationInTx marks pending→approved and upserts authors profile in one transaction.
func (r *PostgresRepo) ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, nil, err
	}
	if app.Status != model.ApplicationPending {
		return nil, nil, model.ErrAlreadyReviewed
	}

	now := time.Now().UTC()
	app.Status = model.ApplicationApproved
	app.ReviewedAt = &now
	app.ReviewedBy = reviewerID
	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'approved', reviewed_at = $2, reviewed_by = $3, reviewer_note = ''
		WHERE id = $1`, applicationID, now, reviewerID)
	if err != nil {
		return nil, nil, err
	}

	author := &model.Author{
		UserID:      app.UserID,
		DisplayName: app.Payload.DisplayName,
		Bio:         app.Payload.Motivation,
		Portfolio:   app.Payload.Portfolio,
		Active:      true,
		VerifiedAt:  &now,
		UpdatedAt:   now,
	}

	var existingID string
	err = tx.QueryRow(ctx, `SELECT id FROM authors WHERE user_id = $1`, author.UserID).Scan(&existingID)
	if errors.Is(err, pgx.ErrNoRows) {
		author.ID = uuid.NewString()
		author.CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO authors (id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			author.ID, author.UserID, author.DisplayName, author.Bio, author.AvatarURL, author.Portfolio,
			author.Active, author.CreatedAt, author.UpdatedAt, author.VerifiedAt)
	} else if err != nil {
		return nil, nil, err
	} else {
		author.ID = existingID
		author.CreatedAt = now // filled from DB below if needed; keep updated
		_, err = tx.Exec(ctx, `
			UPDATE authors
			SET display_name=$2, bio=$3, portfolio=$4, active=true, updated_at=$5, verified_at=$6
			WHERE user_id=$1`,
			author.UserID, author.DisplayName, author.Bio, author.Portfolio, author.UpdatedAt, author.VerifiedAt)
		if err == nil {
			_ = tx.QueryRow(ctx, `SELECT created_at FROM authors WHERE user_id = $1`, author.UserID).Scan(&author.CreatedAt)
		}
	}
	if err != nil {
		return nil, nil, err
	}

	event := map[string]any{
		"event":        "author.upserted",
		"user_id":      author.UserID,
		"display_name": author.DisplayName,
		"source":       "application_approved",
		"application_id": applicationID,
		"timestamp":    now.Unix(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		return nil, nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		uuid.NewString(), "author.upserted", body)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return app, author, nil
}

func (r *PostgresRepo) RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, err
	}
	if app.Status != model.ApplicationPending {
		return nil, model.ErrAlreadyReviewed
	}

	now := time.Now().UTC()
	app.Status = model.ApplicationRejected
	app.ReviewedAt = &now
	app.ReviewedBy = reviewerID
	app.ReviewerNote = note
	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'rejected', reviewed_at = $2, reviewed_by = $3, reviewer_note = $4
		WHERE id = $1`, applicationID, now, reviewerID, note)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

// RevertApplication restores approved→pending and deletes the authors profile (Auth.UpdateRole compensation).
func (r *PostgresRepo) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, err
	}
	if app.Status != model.ApplicationApproved {
		return nil, model.ErrInvalidInput
	}

	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'pending', reviewed_at = NULL, reviewed_by = '', reviewer_note = ''
		WHERE id = $1`, applicationID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `DELETE FROM authors WHERE user_id = $1`, app.UserID)
	if err != nil {
		return nil, err
	}

	app.Status = model.ApplicationPending
	app.ReviewedAt = nil
	app.ReviewedBy = ""
	app.ReviewerNote = ""

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAuthor(row scannable) (*model.Author, error) {
	var a model.Author
	var verifiedAt *time.Time
	if err := row.Scan(
		&a.ID, &a.UserID, &a.DisplayName, &a.Bio, &a.AvatarURL, &a.Portfolio,
		&a.Active, &a.CreatedAt, &a.UpdatedAt, &verifiedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	a.VerifiedAt = verifiedAt
	return &a, nil
}

func scanApplication(row scannable) (*model.AuthorApplication, error) {
	var (
		a          model.AuthorApplication
		status     string
		payload    []byte
		reviewedAt *time.Time
	)
	if err := row.Scan(&a.ID, &a.UserID, &status, &payload, &a.CreatedAt, &reviewedAt, &a.ReviewedBy, &a.ReviewerNote); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrApplicationMissing
		}
		return nil, err
	}
	a.Status = model.ApplicationStatus(status)
	a.ReviewedAt = reviewedAt
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &a.Payload)
	}
	return &a, nil
}

func (r *PostgresRepo) List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM authors WHERE active = true`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at
		FROM authors WHERE active = true
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*model.Author
	for rows.Next() {
		a, err := scanAuthor(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, nil
}

type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepo(addr string, ttl time.Duration) *RedisRepo {
	return &RedisRepo{client: redis.NewClient(&redis.Options{Addr: addr, PoolSize: 10}), ttl: ttl}
}

func (r *RedisRepo) Ping(ctx context.Context) error { return r.client.Ping(ctx).Err() }
func (r *RedisRepo) Close() error                   { return r.client.Close() }

func authorKey(userID string) string { return fmt.Sprintf("authors:user:%s", userID) }

func (r *RedisRepo) Get(ctx context.Context, userID string) (*model.Author, error) {
	b, err := r.client.Get(ctx, authorKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var a model.Author
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *RedisRepo) Set(ctx context.Context, a *model.Author) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, authorKey(a.UserID), b, r.ttl).Err()
}

func (r *RedisRepo) Delete(ctx context.Context, userID string) error {
	return r.client.Del(ctx, authorKey(userID)).Err()
}
===== validate =====
--- services/authors/internal/validate/validate_test.go ---
package validate_test

import (
	"testing"

	authorspb "github.com/Eastwesser/event-horizon/services/authors/proto"
)

func TestUpsertProfileRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     *authorspb.UpsertProfileRequest
		wantErr bool
	}{
		{"valid", &authorspb.UpsertProfileRequest{UserId: "u1", DisplayName: "Alice"}, false},
		{"missing_user", &authorspb.UpsertProfileRequest{DisplayName: "Alice"}, true},
		{"missing_name", &authorspb.UpsertProfileRequest{UserId: "u1"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected: %v", err)
			}
		})
	}
}

func TestGetAuthorRequest_Validate(t *testing.T) {
	if err := (&authorspb.GetAuthorRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&authorspb.GetAuthorRequest{UserId: "u1"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestListAuthorsRequest_Validate(t *testing.T) {
	if err := (&authorspb.ListAuthorsRequest{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitApplicationRequest_Validate(t *testing.T) {
	valid := &authorspb.SubmitApplicationRequest{
		UserId: "u1", DisplayName: "A", Motivation: "why", ContactEmail: "a@b.c",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (&authorspb.SubmitApplicationRequest{UserId: "u1", DisplayName: "A", Motivation: "why"}).Validate(); err == nil {
		t.Fatal("expected contact_email error")
	}
}

func TestGetMyApplicationRequest_Validate(t *testing.T) {
	if err := (&authorspb.GetMyApplicationRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestListApplicationsRequest_Validate(t *testing.T) {
	if err := (&authorspb.ListApplicationsRequest{Status: "pending"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (&authorspb.ListApplicationsRequest{Status: "nope"}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestApproveApplicationRequest_Validate(t *testing.T) {
	if err := (&authorspb.ApproveApplicationRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&authorspb.ApproveApplicationRequest{ApplicationId: "a", ReviewerId: "r"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
===== worker =====
--- services/authors/internal/worker/outbox_worker.go ---
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
		SELECT id, event_type, payload FROM outbox
		WHERE processed = false ORDER BY created_at ASC LIMIT 50`)
	if err != nil {
		slog.Error("authors outbox: query", "err", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, eventType string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			continue
		}
		if w.js != nil {
			if _, err := w.js.Publish(eventType, payload); err != nil {
				slog.Error("authors outbox: publish", "event_type", eventType, "err", err)
				continue
			}
		}
		_, _ = w.db.Exec(ctx, `UPDATE outbox SET processed=true, processed_at=NOW() WHERE id=$1`, id)
	}
}
[denismatveev@c0der event_horizon]$ 