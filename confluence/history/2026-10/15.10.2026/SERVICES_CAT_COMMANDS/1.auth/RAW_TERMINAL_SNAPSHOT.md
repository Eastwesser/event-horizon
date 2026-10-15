[denismatveev@c0der event_horizon]$ find services/auth -type f -name '*.go' | sort
services/auth/cmd/main.go
services/auth/internal/app/app.go
services/auth/internal/app/di.go
services/auth/internal/config/config.go
services/auth/internal/config/config_test.go
services/auth/internal/config/interfaces.go
services/auth/internal/converter/user.go
services/auth/internal/converter/user_test.go
services/auth/internal/handler/grpc_handler.go
services/auth/internal/interceptor/auth.go
services/auth/internal/interceptor/logger.go
services/auth/internal/interceptor/recovery.go
services/auth/internal/interceptor/validate.go
services/auth/internal/jwt/manager.go
services/auth/internal/jwt/manager_test.go
services/auth/internal/model/errors.go
services/auth/internal/model/user.go
services/auth/internal/repository/mocks/user_repository_mock.go
services/auth/internal/repository/redis_repo.go
services/auth/internal/repository/refresh_store.go
services/auth/internal/repository/user_repo.go
services/auth/internal/service/auth_service.go
services/auth/internal/service/auth_service_test.go
services/auth/migrations/embed.go
services/auth/proto/auth_grpc.pb.go
services/auth/proto/auth.pb.go
services/auth/proto/validate.go
services/auth/tests/integration/e2e_auth_test.go
services/auth/tests/integration/user_repo_test.go
[denismatveev@c0der event_horizon]$ ^C
[denismatveev@c0der event_horizon]$ 

[denismatveev@c0der event_horizon]$ 
# Main и app
cat services/auth/cmd/main.go
echo "===== app ====="
cat services/auth/internal/app/*.go

# Handlers
echo "===== handler ====="
for f in services/auth/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done

# Service
echo "===== service ====="
for f in services/auth/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done

# Repository
echo "===== repository ====="
for f in services/auth/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done

# Converter
echo "===== converter ====="
for f in services/auth/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done

# JWT
echo "===== jwt ====="
for f in services/auth/internal/jwt/*.go; do echo "--- $f ---"; cat "$f"; done

# Interceptor
echo "===== interceptor ====="
for f in services/auth/internal/interceptor/*.go; do echo "--- $f ---"; cat "$f"; done

# Config
echo "===== config ====="
for f in services/auth/internal/config/*.go; do echo "--- $f ---"; cat "$f"; done
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/auth/internal/app"
	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "your-secret-key-change-me" {
		log.Println("WARNING: JWT_SECRET is using the insecure default — set JWT_SECRET before production")
	}

	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init auth app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("auth stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
)

// App is the runnable auth service process (gRPC + metrics HTTP).
type App struct {
	cfg           config.ConfigProvider
	log           *slog.Logger
	closer        *closer.Closer
	di            *diContainer
	grpcServer    *grpc.Server
	metricsServer *http.Server
	listener      net.Listener
}

// New wires dependencies and prepares servers (does not Serve yet).
func New(ctx context.Context) (*App, error) {
	cfg := config.Load()
	a := &App{
		cfg: cfg,
		di:  newDiContainer(cfg),
	}
	if err := a.init(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// RunUntilSignal serves gRPC until SIGINT/SIGTERM, then closes resources via closer.
func (a *App) RunUntilSignal(ctx context.Context) error {
	go func() {
		a.log.Info("auth gRPC listening", "port", a.cfg.GRPCConfig().Port())
		if err := a.grpcServer.Serve(a.listener); err != nil {
			a.log.Error("grpc serve stopped", "err", err)
		}
	}()
	if err := a.closer.WaitSignal(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// Close triggers graceful shutdown (idempotent via closer).
func (a *App) Close(ctx context.Context) error {
	return a.closer.CloseAll(ctx)
}

// GRPCAddr returns the bound address (useful for tests).
func (a *App) GRPCAddr() string {
	if a.listener == nil {
		return ""
	}
	return a.listener.Addr().String()
}
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
	"github.com/Eastwesser/event-horizon/services/auth/internal/handler"
	"github.com/Eastwesser/event-horizon/services/auth/internal/interceptor"
	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
	"github.com/Eastwesser/event-horizon/services/auth/migrations"
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

type diContainer struct {
	cfg config.ConfigProvider

	dbpool    *pgxpool.Pool
	redisRepo *repository.RedisAuthRepo
	userRepo  repository.UserRepository
	svc       service.AuthService
	api       pb.AuthServiceServer
}

func newDiContainer(cfg config.ConfigProvider) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initTracer,
		a.initPostgres,
		a.initRedis,
		a.initDomain,
		a.initGRPC,
		a.initMetricsHTTP,
	}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	lc := a.cfg.LoggerConfig()
	a.log = logger.NewFrom(lc.Level(), lc.Format())
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initTracer(ctx context.Context) error {
	endpoint := getenv("JAEGER_ENDPOINT", "localhost:4317")
	shutdown, err := tracing.Init(ctx, tracing.Config{
		ServiceName: "auth",
		Endpoint:    endpoint,
		Environment: getenv("ENVIRONMENT", "development"),
		Insecure:    true,
	})
	if err != nil {
		a.log.Warn("tracer init failed; continuing without OTLP", "err", err)
		return nil
	}
	a.closer.AddNamed("otel tracer", shutdown)
	a.log.Info("jaeger tracer ready", "endpoint", endpoint)
	return nil
}

func (a *App) initPostgres(ctx context.Context) error {
	pg := a.cfg.PostgresConfig()
	poolCfg, err := pgxpool.ParseConfig(pg.DSN())
	if err != nil {
		return fmt.Errorf("postgres config: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 10
	poolCfg.MaxConnLifetime = 5 * time.Minute

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(dbpool)
	if err := migrator.Up(sqlDB, migrations.FS); err != nil {
		dbpool.Close()
		return fmt.Errorf("migrations: %w", err)
	}

	a.di.dbpool = dbpool
	a.closer.AddNamed("postgres pool", func(context.Context) error {
		dbpool.Close()
		return nil
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	rc := a.cfg.RedisConfig()
	redisRepo := repository.NewRedisAuthRepo(rc.Addr(), time.Duration(rc.UserCacheTTLMinutes())*time.Minute)
	if err := redisRepo.Ping(ctx); err != nil {
		a.log.Warn("redis unavailable; sessions/cache degraded", "addr", rc.Addr(), "err", err)
	} else {
		a.log.Info("redis connected", "addr", rc.Addr())
	}
	a.di.redisRepo = redisRepo
	a.closer.AddNamed("redis", func(context.Context) error {
		return redisRepo.Close()
	})
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	jwtCfg := a.cfg.JWTConfig()
	accessTTL := time.Duration(jwtCfg.AccessMinutes()) * time.Minute
	refreshTTL := time.Duration(jwtCfg.RefreshDays()) * 24 * time.Hour
	tokens := jwtauth.NewManager(jwtCfg.Secret(), accessTTL, refreshTTL)

	a.di.userRepo = repository.NewPostgresUserRepo(a.di.dbpool)
	a.di.svc = service.NewAuthService(a.di.userRepo, a.di.redisRepo, tokens)
	a.di.api = handler.NewAuthHandler(a.di.svc)
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPCConfig().Port())
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error {
		return lis.Close()
	})

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
			metrics.UnaryServerInterceptor("auth"),
		),
	)
	pb.RegisterAuthServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("auth")
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	dbpool := a.di.dbpool
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "auth"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := dbpool.Ping(readyCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "auth"})
	})

	srv := &http.Server{Addr: ":" + a.cfg.MetricsConfig().Port(), Handler: mux}
	a.metricsServer = srv
	go func() {
		a.log.Info("metrics/health listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("metrics server error", "err", err)
		}
	}()
	a.closer.AddNamed("metrics http", func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
===== handler =====
--- services/auth/internal/handler/grpc_handler.go ---
package handler

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/auth/internal/converter"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateCredentials(req.Email, req.Password); err != nil {
		return nil, err
	}
	// Public registration always creates role=user (Wave 3: author via approval only).
	roleHint := strings.TrimSpace(req.Role)
	if roleHint != "" && roleHint != "user" {
		return nil, status.Error(codes.InvalidArgument, "public registration only allows role=user")
	}

	userID, role, err := h.authService.Register(ctx, req.Email, req.Password, "user")
	if err != nil {
		if errors.Is(err, model.ErrInvalidRole) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		UserId:  userID,
		Email:   req.Email,
		Success: true,
		Message: "user registered successfully",
		Role:    role,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateCredentials(req.Email, req.Password); err != nil {
		return nil, err
	}

	pair, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
		UserId:       pair.UserID,
		Email:        req.Email,
		Role:         pair.Role,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}
	pair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if errors.Is(err, model.ErrSessionStoreUnavailable) {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
		UserId:       pair.UserID,
		Role:         pair.Role,
	}, nil
}

func (h *AuthHandler) Whoami(ctx context.Context, req *pb.WhoamiRequest) (*pb.WhoamiResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token required")
	}
	user, err := h.authService.Whoami(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &pb.WhoamiResponse{
		UserId:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		Nickname: user.Nickname,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}

	userID, email, role, err := h.authService.ValidateToken(ctx, req.Token)
	if errors.Is(err, model.ErrSessionStoreUnavailable) {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:     true,
		UserId:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: 0, // можно добавить из claims
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}
	if err := h.authService.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	user, err := h.authService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	bestScores, totalScore, err := h.authService.GetUserScores(ctx, req.UserId)
	if err != nil {
		bestScores = make(map[string]int32)
		totalScore = 0
	}

	return converter.UserToProtoWithScores(user, bestScores, totalScore), nil
}

func (h *AuthHandler) UpdateNickname(ctx context.Context, req *pb.UpdateNicknameRequest) (*pb.UpdateNicknameResponse, error) {
	err := h.authService.UpdateNickname(ctx, req.UserId, req.Nickname)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateNicknameResponse{Success: true, Message: "nickname updated"}, nil
}

// UpdateRole меняет роль пользователя. Проверка "запрос делает admin" — на стороне Gateway.
func (h *AuthHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	if req.UserId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and role required")
	}
	if err := h.authService.UpdateRole(ctx, req.UserId, req.Role); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateRoleResponse{Success: true, Message: "role updated"}, nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	users, total, err := h.authService.ListUsers(ctx, req.GetQuery(), int(limit), int(offset))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*pb.ListUserEntry, 0, len(users))
	for _, u := range users {
		created := ""
		if !u.CreatedAt.IsZero() {
			created = u.CreatedAt.UTC().Format(time.RFC3339)
		}
		out = append(out, &pb.ListUserEntry{
			UserId:    u.ID,
			Email:     u.Email,
			Role:      u.Role,
			Nickname:  u.Nickname,
			CreatedAt: created,
		})
	}
	return &pb.ListUsersResponse{Users: out, Total: total}, nil
}

// validateCredentials mirrors intended protoc-gen-validate rules for Login/Register
// (email format, password min_len=8, max_len=128) until .pb.validate.go is generated.
func validateCredentials(email, password string) error {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return status.Error(codes.InvalidArgument, "email and password required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	n := utf8.RuneCountInString(password)
	if n < 8 || n > 128 {
		return status.Error(codes.InvalidArgument, "password must be 8-128 characters")
	}
	return nil
}
===== service =====
--- services/auth/internal/service/auth_service.go ---
// Package service — auth business logic (Week 6: access + refresh JWT).
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
)

const bcryptCost = 12

// Public registration may only create role=user. Author is granted via Wave 3 approval.
var validPublicRoles = map[string]bool{"user": true}
var validRoles = map[string]bool{"user": true, "author": true, "admin": true}

// TokenPair is returned by Login / RefreshToken.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       string
	Role         string
}

type authService struct {
	repo   repository.UserRepository
	cache  *repository.RedisAuthRepo
	tokens *jwtauth.Manager
}

type AuthService interface {
	Register(ctx context.Context, email, password, role string) (userID, resolvedRole string, err error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	ValidateToken(ctx context.Context, tokenString string) (userID, email, role string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	Whoami(ctx context.Context, accessToken string) (*model.User, error)
	Logout(ctx context.Context, tokenString string) error
	GetUser(ctx context.Context, userID string) (*model.User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	UpdateRole(ctx context.Context, userID, role string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
	ListUsers(ctx context.Context, query string, limit, offset int) ([]*model.User, int64, error)
}

func NewAuthService(repo repository.UserRepository, cache *repository.RedisAuthRepo, tokens *jwtauth.Manager) AuthService {
	return &authService{repo: repo, cache: cache, tokens: tokens}
}

// NewAuthServiceLegacy keeps older call sites working (tests / gradual migrate).
func NewAuthServiceLegacy(repo repository.UserRepository, cache *repository.RedisAuthRepo, jwtSecret string, expHours int) AuthService {
	access := 15 * time.Minute
	refresh := 7 * 24 * time.Hour
	if expHours > 0 && expHours < 24 {
		// treat small values as access hours for backward-compat tests
		access = time.Duration(expHours) * time.Hour
	}
	return NewAuthService(repo, cache, jwtauth.NewManager(jwtSecret, access, refresh))
}

func (s *authService) Register(ctx context.Context, email, password, role string) (string, string, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if existing != nil {
		return "", "", model.ErrUserAlreadyExists
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "user"
	}
	if !validPublicRoles[role] {
		return "", "", fmt.Errorf("%w: public registration only allows role=user", model.ErrInvalidRole)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", "", err
	}
	userID, err := s.repo.Create(ctx, email, string(hashedPassword), role)
	if err != nil {
		return "", "", err
	}
	return userID, role, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}
	role := user.Role
	if role == "" {
		role = "user"
	}
	return s.issuePair(ctx, user.ID, user.Email, role)
}

func (s *authService) issuePair(ctx context.Context, userID, email, role string) (*TokenPair, error) {
	access, accessJTI, accessExp, err := s.tokens.GenerateAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}
	refresh, refreshJTI, refreshExp, err := s.tokens.GenerateRefreshToken(userID, email, role)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if err := s.cache.CreateSession(ctx, accessJTI, userID, time.Until(accessExp)); err != nil {
			log.Printf("auth: store access session failed: %v", err)
		}
		if err := s.cache.SaveRefresh(ctx, refreshJTI, userID, time.Until(refreshExp)); err != nil {
			log.Printf("auth: store refresh failed: %v", err)
		}
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.tokens.AccessTTL().Seconds()),
		UserID:       userID,
		Role:         role,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (string, string, string, error) {
	claims, err := s.tokens.ValidateAccessToken(tokenString)
	if err != nil {
		// legacy tokens without type=access may still validate as MapClaims via fallback
		return "", "", "", model.ErrInvalidToken
	}
	if s.cache != nil && claims.JTI != "" {
		exists, cacheErr := s.cache.SessionExists(ctx, claims.JTI)
		if cacheErr != nil {
			log.Printf("auth: Redis session check failed: %v", cacheErr)
			return "", "", "", model.ErrSessionStoreUnavailable
		}
		if !exists {
			return "", "", "", model.ErrSessionRevoked
		}
	}
	return claims.UserID, claims.Email, claims.Role, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, model.ErrInvalidToken
	}
	if s.cache != nil && claims.JTI != "" {
		ok, cacheErr := s.cache.RefreshExists(ctx, claims.JTI)
		if cacheErr != nil {
			log.Printf("auth: Redis refresh check failed: %v", cacheErr)
			return nil, model.ErrSessionStoreUnavailable
		}
		if !ok {
			return nil, model.ErrSessionRevoked
		}
		_ = s.cache.DeleteRefresh(ctx, claims.JTI, claims.UserID)
	}
	// Re-read role from DB so approve/UpdateRole is picked up without re-login.
	role := claims.Role
	email := claims.Email
	if user, err := s.GetUser(ctx, claims.UserID); err == nil && user != nil {
		if user.Role != "" {
			role = user.Role
		}
		if user.Email != "" {
			email = user.Email
		}
	}
	return s.issuePair(ctx, claims.UserID, email, role)
}

func (s *authService) Whoami(ctx context.Context, accessToken string) (*model.User, error) {
	userID, _, _, err := s.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}

func (s *authService) Logout(ctx context.Context, tokenString string) error {
	if s.cache == nil {
		return nil
	}
	claims, err := jwtauth.ParseUnverified(tokenString)
	if err != nil {
		return err
	}
	if claims.JTI == "" {
		return nil
	}
	switch claims.Type {
	case jwtauth.TokenTypeRefresh:
		return s.cache.DeleteRefresh(ctx, claims.JTI, claims.UserID)
	default:
		// access or legacy — drop jti from user session set when known
		if claims.UserID != "" {
			return s.cache.DeleteSessionForUser(ctx, claims.JTI, claims.UserID)
		}
		return s.cache.DeleteSession(ctx, claims.JTI)
	}
}

func (s *authService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	if s.cache != nil {
		cached, err := s.cache.GetUserCache(ctx, userID)
		if err == nil {
			return cached, nil
		}
		if !errors.Is(err, repository.ErrCacheMiss) {
			log.Printf("auth: Redis GetUserCache failed: %v", err)
		}
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user != nil && s.cache != nil {
		if cacheErr := s.cache.SetUserCache(ctx, user); cacheErr != nil {
			log.Printf("auth: Redis SetUserCache failed: %v", cacheErr)
		}
	}
	return user, nil
}

func (s *authService) UpdateNickname(ctx context.Context, userID, nickname string) error {
	if err := s.repo.UpdateNickname(ctx, userID, nickname); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	return nil
}

func (s *authService) UpdateRole(ctx context.Context, userID, role string) error {
	if !validRoles[role] {
		return fmt.Errorf("%w: %q", model.ErrInvalidRole, role)
	}
	if err := s.repo.UpdateRole(ctx, userID, role); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	// Revoke access only — keep refresh so FE 401 → refresh picks up new role (B).
	if s.cache != nil {
		if err := s.cache.DeleteAllSessionsForUser(ctx, userID); err != nil {
			log.Printf("auth: DeleteAllSessionsForUser failed: %v", err)
		}
	}
	return nil
}

func (s *authService) invalidateUserCache(ctx context.Context, userID string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.InvalidateUserCache(ctx, userID); err != nil {
		log.Printf("auth: Redis InvalidateUserCache failed: %v", err)
	}
}

func (s *authService) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
	return s.repo.GetUserScores(ctx, userID)
}

func (s *authService) ListUsers(ctx context.Context, query string, limit, offset int) ([]*model.User, int64, error) {
	return s.repo.ListUsers(ctx, query, limit, offset)
}
--- services/auth/internal/service/auth_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"golang.org/x/crypto/bcrypt"

	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository/mocks"
)

const testJWTSecret = "test-secret-key-for-jwt-signing-tests"

func newTestService(repo *mocks.UserRepository) AuthService {
	return NewAuthServiceLegacy(repo, nil, testJWTSecret, 1)
}

func newTestServiceWithManager(repo *mocks.UserRepository) AuthService {
	mgr := jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)
	return NewAuthService(repo, nil, mgr)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, nil
		},
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			if role != "user" || email != "user@example.com" || passwordHash == "" {
				t.Fatalf("unexpected create args: %s %s %s", email, role, passwordHash)
			}
			return "uid-1", nil
		},
	}
	id, role, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if err != nil {
		t.Fatal(err)
	}
	if id != "uid-1" || role != "user" {
		t.Fatalf("got %s %s", id, role)
	}
}

func TestRegisterAlreadyExists(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "x", Email: email}, nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if !errors.Is(err, model.ErrUserAlreadyExists) {
		t.Fatalf("want ErrUserAlreadyExists, got %v", err)
	}
}

func TestRegisterInvalidRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "admin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "author"}, nil
		},
	}
	pair, err := newTestService(repo).Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.UserID != "uid-1" || pair.Role != "author" {
		t.Fatalf("unexpected login result: %+v", pair)
	}
}

func TestLoginBadPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	_, err = newTestService(repo).Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestUpdateRoleInvalid(t *testing.T) {
	err := newTestService(&mocks.UserRepository{}).UpdateRole(context.Background(), "uid-1", "superadmin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestUpdateRoleSuccess(t *testing.T) {
	called := false
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(ctx context.Context, userID, role string) error {
			called = true
			if userID != "uid-1" || role != "admin" {
				t.Fatalf("bad args")
			}
			return nil
		},
	}
	if err := newTestService(repo).UpdateRole(context.Background(), "uid-1", "admin"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("UpdateRole not called")
	}
}

func TestGetUserFromRepo(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "a@b.c", Role: "user"}, nil
		},
	}
	u, err := newTestService(repo).GetUser(context.Background(), "uid-1")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "uid-1" {
		t.Fatalf("got %s", u.ID)
	}
}

func TestRegisterDefaultRole(t *testing.T) {
	var gotRole string
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			gotRole = role
			return "uid-1", nil
		},
	}
	_, role, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if err != nil || role != "user" || gotRole != "user" {
		t.Fatalf("role=%q gotRole=%q err=%v", role, gotRole, err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
	}
	_, err := newTestService(repo).Login(context.Background(), "missing@example.com", "password123")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateTokenRoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	uid, email, role, err := svc.ValidateToken(context.Background(), pair.AccessToken)
	if err != nil || uid != "uid-1" || email != "user@example.com" || role != "user" {
		t.Fatalf("uid=%s email=%s role=%s err=%v", uid, email, role, err)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	_, _, _, err := newTestServiceWithManager(&mocks.UserRepository{}).ValidateToken(context.Background(), "not-a-jwt")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "author"}, nil
		},
		GetByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "user@example.com", Role: "author"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" || refreshed.UserID != "uid-1" || refreshed.Role != "author" {
		t.Fatalf("unexpected refresh: %+v", refreshed)
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	_, err := newTestServiceWithManager(&mocks.UserRepository{}).RefreshToken(context.Background(), "bad-token")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestWhoami(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "user@example.com", Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.Whoami(context.Background(), pair.AccessToken)
	if err != nil || u.ID != "uid-1" {
		t.Fatalf("user=%+v err=%v", u, err)
	}
}

func TestLogoutNilCache(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), pair.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateNickname(t *testing.T) {
	called := false
	repo := &mocks.UserRepository{
		UpdateNicknameFn: func(ctx context.Context, userID, nickname string) error {
			called = true
			if userID != "uid-1" || nickname != "neo" {
				t.Fatalf("bad args %s %s", userID, nickname)
			}
			return nil
		},
	}
	if err := newTestService(repo).UpdateNickname(context.Background(), "uid-1", "neo"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("UpdateNickname not called")
	}
}

func TestGetUserScores(t *testing.T) {
	repo := &mocks.UserRepository{
		GetUserScoresFn: func(ctx context.Context, userID string) (map[string]int32, int32, error) {
			return map[string]int32{"hanoi": 42}, 42, nil
		},
	}
	byGame, total, err := newTestService(repo).GetUserScores(context.Background(), "uid-1")
	if err != nil || total != 42 || byGame["hanoi"] != 42 {
		t.Fatalf("byGame=%v total=%d err=%v", byGame, total, err)
	}
}

func TestRegisterAuthorRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			t.Fatal("Create must not be called when role=author is rejected")
			return "", nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "author@example.com", "password123", "author")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestRegisterRepoErrors(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, want
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "user")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestLoginRepoError(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, want
		},
	}
	_, err := newTestService(repo).Login(context.Background(), "user@example.com", "password123")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestGetUserRepoError(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return nil, want
		},
	}
	_, err := newTestService(repo).GetUser(context.Background(), "uid-1")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterCreateError(t *testing.T) {
	want := errors.New("insert failed")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			return "", want
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "user")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateRoleRepoError(t *testing.T) {
	want := errors.New("update failed")
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(ctx context.Context, userID, role string) error { return want },
	}
	err := newTestService(repo).UpdateRole(context.Background(), "uid-1", "admin")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateNicknameRepoError(t *testing.T) {
	want := errors.New("nickname failed")
	repo := &mocks.UserRepository{
		UpdateNicknameFn: func(ctx context.Context, userID, nickname string) error { return want },
	}
	err := newTestService(repo).UpdateNickname(context.Background(), "uid-1", "neo")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestWhoamiInvalidToken(t *testing.T) {
	_, err := newTestServiceWithManager(&mocks.UserRepository{}).Whoami(context.Background(), "bad-token")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("got %v", err)
	}
}

func TestListUsers_Delegates(t *testing.T) {
	repo := &mocks.UserRepository{
		ListUsersFn: func(_ context.Context, query string, limit, offset int) ([]*model.User, int64, error) {
			if query != "den" || limit != 10 || offset != 5 {
				t.Fatalf("q=%s limit=%d offset=%d", query, limit, offset)
			}
			return []*model.User{{ID: "u1", Email: "den@x.com"}}, 1, nil
		},
	}
	list, total, err := newTestService(repo).ListUsers(context.Background(), "den", 10, 5)
	if err != nil || total != 1 || len(list) != 1 || list[0].ID != "u1" {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestRegisterRejectsAdminRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(context.Context, string) (*model.User, error) { return nil, nil },
		CreateFn: func(context.Context, string, string, string) (string, error) {
			t.Fatal("Create must not run for admin self-register")
			return "", nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "a@b.c", "password123", "admin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("got %v", err)
	}
}

func withMiniRedis(t *testing.T) (*repository.RedisAuthRepo, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	cache := repository.NewRedisAuthRepo(mr.Addr(), time.Minute)
	return cache, func() {
		_ = cache.Close()
		mr.Close()
	}
}

func TestLogout_WithCacheDeletesSession(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	mgr := jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)
	svc := NewAuthService(repo, cache, mgr)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := svc.ValidateToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("pre-logout validate: %v", err)
	}
	if err := svc.Logout(context.Background(), pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = svc.ValidateToken(context.Background(), pair.AccessToken)
	if !errors.Is(err, model.ErrSessionRevoked) {
		t.Fatalf("want ErrSessionRevoked after logout, got %v", err)
	}
}

func TestRefreshToken_RevokesOldRefresh(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
		GetByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "user@example.com", Role: "user"}, nil
		},
	}
	svc := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour))
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	next, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
	if err != nil || next.AccessToken == "" {
		t.Fatalf("refresh=%+v err=%v", next, err)
	}
	_, err = svc.RefreshToken(context.Background(), pair.RefreshToken)
	if !errors.Is(err, model.ErrSessionRevoked) {
		t.Fatalf("want revoked old refresh, got %v", err)
	}
}

func TestUpdateRole_RevokesAccessKeepsRefresh_DBRoleOnRefresh(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	role := "user"
	repo := &mocks.UserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: role}, nil
		},
		GetByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "user@example.com", Role: role}, nil
		},
		UpdateRoleFn: func(_ context.Context, userID, r string) error {
			if userID != "uid-1" {
				t.Fatalf("userID=%s", userID)
			}
			role = r
			return nil
		},
	}
	svc := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour))
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if pair.Role != "user" {
		t.Fatalf("login role=%s", pair.Role)
	}
	if _, _, _, err := svc.ValidateToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("pre-role validate: %v", err)
	}

	if err := svc.UpdateRole(context.Background(), "uid-1", "author"); err != nil {
		t.Fatal(err)
	}

	_, _, _, err = svc.ValidateToken(context.Background(), pair.AccessToken)
	if !errors.Is(err, model.ErrSessionRevoked) {
		t.Fatalf("want old access revoked, got %v", err)
	}

	// Refresh must still work (not revoked) and mint author from DB.
	next, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh after role change: %v", err)
	}
	if next.Role != "author" {
		t.Fatalf("want role=author after refresh, got %q", next.Role)
	}
	uid, _, gotRole, err := svc.ValidateToken(context.Background(), next.AccessToken)
	if err != nil || uid != "uid-1" || gotRole != "author" {
		t.Fatalf("new access validate uid=%s role=%s err=%v", uid, gotRole, err)
	}
}

func TestGetUser_CacheHit(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	u := &model.User{ID: "uid-1", Email: "a@b.c", Role: "author"}
	if err := cache.SetUserCache(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByIDFn: func(context.Context, string) (*model.User, error) {
			t.Fatal("repo must not be called on cache hit")
			return nil, nil
		},
	}
	got, err := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)).
		GetUser(context.Background(), "uid-1")
	if err != nil || got.Role != "author" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestUpdateRole_InvalidatesCache(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	if err := cache.SetUserCache(context.Background(), &model.User{ID: "uid-1", Email: "a@b.c", Role: "user"}); err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(context.Context, string, string) error { return nil },
		GetByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "a@b.c", Role: "author"}, nil
		},
	}
	svc := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour))
	if err := svc.UpdateRole(context.Background(), "uid-1", "author"); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetUser(context.Background(), "uid-1")
	if err != nil || got.Role != "author" {
		t.Fatalf("after invalidate should load author from repo: %+v err=%v", got, err)
	}
}
===== repository =====
--- services/auth/internal/repository/redis_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss сигнализирует о промахе кеша (ключ не найден) — это не ошибка Redis,
// вызывающая сторона должна сходить в Postgres.
var ErrCacheMiss = errors.New("cache miss")

// RedisAuthRepo — Redis-хранилище для Auth Service.
// Два независимых назначения одного клиента:
//  1. Cache-Aside для user:{id} (TTL 5 минут) — снижает нагрузку на Postgres.
//  2. Хранилище активных сессий (auth:session:{jti}) — авторизационные ключи
//     ОБЯЗАНЫ жить в Redis (см. confluence/history/2026-08/13.08.2026/THE_VOICE_MESSAGE.md),
//     что даёт возможность отзыва токена (logout) без ожидания истечения exp.
type RedisAuthRepo struct {
	client  *redis.Client
	userTTL time.Duration
}

func NewRedisAuthRepo(addr string, userTTL time.Duration) *RedisAuthRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		PoolSize: 10,
	})
	return &RedisAuthRepo{client: client, userTTL: userTTL}
}

func (r *RedisAuthRepo) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisAuthRepo) Close() error {
	return r.client.Close()
}

// --- Cache-Aside для User ---

func userCacheKey(id string) string {
	return fmt.Sprintf("auth:user:%s", id)
}

func (r *RedisAuthRepo) GetUserCache(ctx context.Context, id string) (*User, error) {
	data, err := r.client.Get(ctx, userCacheKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *RedisAuthRepo) SetUserCache(ctx context.Context, u *User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, userCacheKey(u.ID), data, r.userTTL).Err()
}

func (r *RedisAuthRepo) InvalidateUserCache(ctx context.Context, id string) error {
	return r.client.Del(ctx, userCacheKey(id)).Err()
}

// --- Хранилище сессий (авторизационные ключи) ---

func sessionKey(jti string) string {
	return fmt.Sprintf("auth:session:%s", jti)
}

func userSessionSetKey(userID string) string {
	return fmt.Sprintf("auth:user_session:%s", userID)
}

// CreateSession регистрирует выданный JWT (по jti) как активный, TTL = сроку жизни токена.
// Also tracks jti in auth:user_session:{userID} so UpdateRole can revoke all access tokens.
func (r *RedisAuthRepo) CreateSession(ctx context.Context, jti, userID string, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, sessionKey(jti), userID, ttl)
	if userID != "" {
		pipe.SAdd(ctx, userSessionSetKey(userID), jti)
		pipe.Expire(ctx, userSessionSetKey(userID), ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// SessionExists проверяет, не отозвана ли сессия (logout/бан удаляют ключ раньше TTL).
func (r *RedisAuthRepo) SessionExists(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, sessionKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// DeleteSession отзывает одну access-сессию (logout).
func (r *RedisAuthRepo) DeleteSession(ctx context.Context, jti string) error {
	return r.client.Del(ctx, sessionKey(jti)).Err()
}

// DeleteSessionForUser removes one access jti and drops it from the user's session set.
func (r *RedisAuthRepo) DeleteSessionForUser(ctx context.Context, jti, userID string) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, sessionKey(jti))
	if userID != "" {
		pipe.SRem(ctx, userSessionSetKey(userID), jti)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteAllSessionsForUser revokes every access session for the user (role change).
// Refresh tokens are intentionally left alive so FE can 401 → refresh → new role.
func (r *RedisAuthRepo) DeleteAllSessionsForUser(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	jtis, err := r.client.SMembers(ctx, userSessionSetKey(userID)).Result()
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, sessionKey(jti))
	}
	pipe.Del(ctx, userSessionSetKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}
--- services/auth/internal/repository/refresh_store.go ---
package repository

import (
	"context"
	"fmt"
	"time"
)

func refreshKey(jti string) string {
	return fmt.Sprintf("auth:refresh:%s", jti)
}

func userRefreshSetKey(userID string) string {
	return fmt.Sprintf("auth:user_refresh:%s", userID)
}

// SaveRefresh stores a refresh token jti in Redis (Week 6).
func (r *RedisAuthRepo) SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, refreshKey(jti), userID, ttl)
	pipe.SAdd(ctx, userRefreshSetKey(userID), jti)
	pipe.Expire(ctx, userRefreshSetKey(userID), ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisAuthRepo) RefreshExists(ctx context.Context, jti string) (bool, error) {
	n, err := r.client.Exists(ctx, refreshKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisAuthRepo) DeleteRefresh(ctx context.Context, jti, userID string) error {
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, refreshKey(jti))
	if userID != "" {
		pipe.SRem(ctx, userRefreshSetKey(userID), jti)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// DeleteAllRefreshForUser revokes every refresh token for the user (password change).
func (r *RedisAuthRepo) DeleteAllRefreshForUser(ctx context.Context, userID string) error {
	jtis, err := r.client.SMembers(ctx, userRefreshSetKey(userID)).Result()
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, refreshKey(jti))
	}
	pipe.Del(ctx, userRefreshSetKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}
--- services/auth/internal/repository/user_repo.go ---
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/pkg/sqb"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

type User = model.User

// UserRepository is the auth user store.
// SQL is built with pkg/sqb. The statement text holds $1, $2; emails and password hashes stay in the args slice.
//
//go:generate mockery --name=UserRepository --dir=. --output=mocks --outpkg=mocks --filename=user_repository_mock.go --with-expecter
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, role string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	UpdateRole(ctx context.Context, userID, role string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
	ListUsers(ctx context.Context, query string, limit, offset int) ([]*User, int64, error)
}

type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(ctx context.Context, email, passwordHash, role string) (string, error) {
	nickname := strings.Split(email, "@")[0]
	query, args, err := sqb.Insert("users").
		Columns("email", "password_hash", "nickname", "role").
		Values(email, passwordHash, nickname, role).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", err
	}
	var userID string
	if err := r.db.QueryRow(ctx, query, args...).Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := sqb.Select("id", "email", "password_hash", "nickname", "role", "created_at").
		From("users").
		Where("email = $1", email).
		ToSql()
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	query, args, err := sqb.Select("id", "email", "password_hash", "nickname", "role", "created_at").
		From("users").
		Where("id = $1", id).
		ToSql()
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepo) UpdateNickname(ctx context.Context, userID, nickname string) error {
	query, args, err := sqb.Update("users").
		Set("nickname", nickname).
		SetRaw("updated_at = CURRENT_TIMESTAMP").
		Where("id = $1", userID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *PostgresUserRepo) UpdateRole(ctx context.Context, userID, role string) error {
	query, args, err := sqb.Update("users").
		Set("role", role).
		SetRaw("updated_at = CURRENT_TIMESTAMP").
		Where("id = $1", userID).
		ToSql()
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *PostgresUserRepo) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
	query, args, err := sqb.Select("game_id", "MAX(score) as best_score").
		From("scores").
		Where("user_id = $1", userID).
		GroupBy("game_id").
		ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bestScores := make(map[string]int32)
	var totalScore int32
	for rows.Next() {
		var gameID string
		var best int32
		if err := rows.Scan(&gameID, &best); err != nil {
			return nil, 0, err
		}
		bestScores[gameID] = best
		totalScore += best
	}
	return bestScores, totalScore, rows.Err()
}

func (r *PostgresUserRepo) ListUsers(ctx context.Context, query string, limit, offset int) ([]*User, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := strings.TrimSpace(query)
	var (
		total     int64
		listSQL   string
		listArgs  []any
		countSQL  string
		countArgs []any
	)

	if q == "" {
		countSQL = `SELECT COUNT(*) FROM users`
		listSQL = `SELECT id, email, password_hash, nickname, role, created_at
			FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		listArgs = []any{limit, offset}
	} else {
		pattern := "%" + q + "%"
		countSQL = `SELECT COUNT(*) FROM users WHERE email ILIKE $1`
		countArgs = []any{pattern}
		listSQL = `SELECT id, email, password_hash, nickname, role, created_at
			FROM users WHERE email ILIKE $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		listArgs = []any{pattern, limit, offset}
	}

	if err := r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*User, 0, limit)
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Role, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}
===== converter =====
--- services/auth/internal/converter/user.go ---
package converter

import (
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

// UserToProto maps a domain user to GetUserResponse fields (without scores).
func UserToProto(u *model.User) *pb.GetUserResponse {
	if u == nil {
		return nil
	}
	return &pb.GetUserResponse{
		UserId:   u.ID,
		Email:    u.Email,
		Nickname: u.Nickname,
		Role:     u.Role,
	}
}

// UserToProtoWithScores attaches best scores to the response.
func UserToProtoWithScores(u *model.User, bestScores map[string]int32, totalScore int32) *pb.GetUserResponse {
	resp := UserToProto(u)
	if resp == nil {
		return nil
	}
	resp.BestScores = bestScores
	resp.TotalScore = totalScore
	return resp
}
--- services/auth/internal/converter/user_test.go ---
package converter

import (
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

func TestUserToProtoWithScores(t *testing.T) {
	u := &model.User{ID: "u1", Email: "a@b.c", Nickname: "neo", Role: "author", CreatedAt: time.Now()}
	resp := UserToProtoWithScores(u, map[string]int32{"hexagon": 10}, 10)
	if resp == nil || resp.UserId != "u1" || resp.Role != "author" || resp.BestScores["hexagon"] != 10 {
		t.Fatalf("unexpected: %+v", resp)
	}
}
===== jwt =====
--- services/auth/internal/jwt/manager.go ---
package jwtauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims carries EH identity fields.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

type Generator interface {
	GenerateAccessToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
	GenerateRefreshToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
}

type Validator interface {
	ValidateAccessToken(token string) (*Claims, error)
	ValidateRefreshToken(token string) (*Claims, error)
}

type Manager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) GenerateAccessToken(userID, email, role string) (string, string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeAccess, m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID, email, role string) (string, string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeRefresh, m.refreshTTL)
}

func (m *Manager) generate(userID, email, role, typ string, ttl time.Duration) (string, string, time.Time, error) {
	jti := uuid.NewString()
	exp := time.Now().Add(ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   typ,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(m.secret)
	return s, jti, exp, err
}

func (m *Manager) ValidateAccessToken(token string) (*Claims, error) {
	return m.validate(token, TokenTypeAccess)
}

func (m *Manager) ValidateRefreshToken(token string) (*Claims, error) {
	return m.validate(token, TokenTypeRefresh)
}

func (m *Manager) validate(tokenString, wantType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != wantType {
		// Legacy access tokens (pre-W6) had no type claim.
		if !(wantType == TokenTypeAccess && claims.Type == "") {
			return nil, fmt.Errorf("unexpected token type %q", claims.Type)
		}
	}
	if claims.UserID == "" {
		return nil, errors.New("missing user_id")
	}
	if claims.Role == "" {
		claims.Role = "user"
	}
	return claims, nil
}

// ParseUnverified extracts claims without signature check (logout helper).
func ParseUnverified(tokenString string) (*Claims, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
--- services/auth/internal/jwt/manager_test.go ---
package jwtauth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessAndRefreshRoundTrip(t *testing.T) {
	m := NewManager("secret", 0, 0)
	access, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	refresh, _, _, err := m.GenerateRefreshToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	ac, err := m.ValidateAccessToken(access)
	if err != nil || ac.UserID != "u1" || ac.Type != TokenTypeAccess {
		t.Fatalf("access: %+v %v", ac, err)
	}
	rc, err := m.ValidateRefreshToken(refresh)
	if err != nil || rc.Type != TokenTypeRefresh {
		t.Fatalf("refresh: %+v %v", rc, err)
	}
	if _, err := m.ValidateAccessToken(refresh); err == nil {
		t.Fatal("refresh must not validate as access")
	}
}

func TestDefaultTTLs(t *testing.T) {
	m := NewManager("x", 0, 0)
	if m.AccessTTL() != 15*time.Minute {
		t.Fatalf("access ttl %v", m.AccessTTL())
	}
	if m.RefreshTTL() != 7*24*time.Hour {
		t.Fatalf("refresh ttl %v", m.RefreshTTL())
	}
}

func TestExpiredAccessRejected(t *testing.T) {
	m := NewManager("secret", time.Millisecond, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := m.ValidateAccessToken(tok); err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "admin")
	if err != nil {
		t.Fatal(err)
	}
	bad := tok + "x"
	if _, err := m.ValidateAccessToken(bad); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestLegacyAccessWithoutTypeClaim(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	claims := Claims{
		UserID: "u1",
		Email:  "a@b.c",
		Role:   "",
		Type:   "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.ValidateAccessToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != "user" {
		t.Fatalf("default role want user, got %q", got.Role)
	}
}

func TestParseUnverified(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u9", "x@y.z", "author")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseUnverified(tok)
	if err != nil || c.UserID != "u9" || c.Role != "author" {
		t.Fatalf("%+v %v", c, err)
	}
}

func TestMissingUserIDRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	claims := Claims{
		UserID: "",
		Email:  "a@b.c",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ValidateAccessToken(raw); err == nil {
		t.Fatal("expected missing user_id")
	}
}

func TestWrongSecretRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	other := NewManager("other", time.Hour, time.Hour)
	if _, err := other.ValidateAccessToken(tok); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestParseUnverifiedBadToken(t *testing.T) {
	if _, err := ParseUnverified("not-a-jwt"); err == nil {
		t.Fatal("expected error")
	}
}
===== interceptor =====
--- services/auth/internal/interceptor/auth.go ---
package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
)

type ctxKey string

const UserIDKey ctxKey = "auth_user_id"
const RoleKey ctxKey = "auth_role"

// Auth returns a unary interceptor that validates Bearer access tokens from metadata.
// Skip paths that must stay public (Register/Login/RefreshToken).
func Auth(tokens *jwtauth.Manager, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authz := first(md.Get("authorization"))
		token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if token == "" {
			token = strings.TrimSpace(strings.TrimPrefix(authz, "bearer "))
		}
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}
		claims, err := tokens.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		return handler(ctx, req)
	}
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
--- services/auth/internal/interceptor/logger.go ---
package interceptor

import (
	"context"
	"log/slog"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Logger returns a unary server interceptor that logs method, status, and duration.
func Logger() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := path.Base(info.FullMethod)
		start := time.Now()

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			slog.Info("gRPC request",
				"method", method,
				"code", st.Code().String(),
				"duration_ms", duration.Milliseconds(),
				"err", err.Error(),
			)
		} else {
			slog.Info("gRPC request",
				"method", method,
				"code", "OK",
				"duration_ms", duration.Milliseconds(),
			)
		}
		return resp, err
	}
}
--- services/auth/internal/interceptor/recovery.go ---
package interceptor

import (
	"context"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recovery returns a unary server interceptor that converts panics into Internal errors.
func Recovery() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("gRPC panic recovered method=%s panic=%v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
--- services/auth/internal/interceptor/validate.go ---
package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type protoValidator interface {
	Validate() error
}

// Validate rejects requests that implement Validate() and fail their rules.
// Matches Kozirev week_1 pattern (req.Validate() → InvalidArgument).
func Validate() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if v, ok := req.(protoValidator); ok {
			if err := v.Validate(); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
			}
		}
		return handler(ctx, req)
	}
}
===== config =====
--- services/auth/internal/config/config.go ---
package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort            string
    MetricsPort         string
    DBHost              string
    DBPort              string
    DBUser              string
    DBPassword          string
    DBName              string
    JWTSecret           string
    JWTExpHours         int // legacy; prefer AccessTokenMinutes + RefreshTokenDays
    AccessTokenMinutes  int
    RefreshTokenDays    int
    RedisAddr           string
    UserCacheTTLMinutes int
}

func Load() *Config {
    return &Config{
        GRPCPort:            getEnv("GRPC_PORT", "50051"),
        MetricsPort:         getEnv("METRICS_PORT", "9091"),
        DBHost:              getEnv("DB_HOST", "localhost"),
        DBPort:              getEnv("DB_PORT", "5460"),
        DBUser:              getEnv("DB_USER", "eventhorizon"),
        DBPassword:          getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:              getEnv("DB_NAME", "eventhorizon"),
        JWTSecret:           getEnv("JWT_SECRET", "your-secret-key-change-me"),
        JWTExpHours:         getEnvAsInt("JWT_EXP_HOURS", 24),
        AccessTokenMinutes:  getEnvAsInt("JWT_ACCESS_MINUTES", 15),
        RefreshTokenDays:    getEnvAsInt("JWT_REFRESH_DAYS", 7),
        RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
        UserCacheTTLMinutes: getEnvAsInt("USER_CACHE_TTL_MINUTES", 5),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}
--- services/auth/internal/config/config_test.go ---
package config_test

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
)

func TestConfigProviderViews(t *testing.T) {
	cfg := config.Load()
	var _ config.ConfigProvider = cfg

	if cfg.GRPCConfig().Port() == "" {
		t.Fatal("empty grpc port")
	}
	if cfg.PostgresConfig().DSN() == "" {
		t.Fatal("empty dsn")
	}
	if cfg.JWTConfig().Secret() == "" {
		t.Fatal("empty jwt")
	}
	if cfg.LoggerConfig().Level() == "" {
		t.Fatal("empty log level")
	}
}
--- services/auth/internal/config/interfaces.go ---
package config

// ConfigProvider exposes typed config views for DI and tests.
type ConfigProvider interface {
	GRPCConfig() GRPCConfig
	MetricsConfig() MetricsConfig
	PostgresConfig() PostgresConfig
	RedisConfig() RedisConfig
	JWTConfig() JWTConfig
	LoggerConfig() LoggerConfig
}

type GRPCConfig interface {
	Port() string
}

type MetricsConfig interface {
	Port() string
}

type PostgresConfig interface {
	Host() string
	Port() string
	User() string
	Password() string
	Database() string
	DSN() string
}

type RedisConfig interface {
	Addr() string
	UserCacheTTLMinutes() int
}

type JWTConfig interface {
	Secret() string
	ExpHours() int
	AccessMinutes() int
	RefreshDays() int
}

type LoggerConfig interface {
	Level() string
	Format() string
}

type grpcView struct{ port string }

func (g grpcView) Port() string { return g.port }

type metricsView struct{ port string }

func (m metricsView) Port() string { return m.port }

type postgresView struct {
	host, port, user, password, database string
}

func (p postgresView) Host() string     { return p.host }
func (p postgresView) Port() string     { return p.port }
func (p postgresView) User() string     { return p.user }
func (p postgresView) Password() string { return p.password }
func (p postgresView) Database() string { return p.database }
func (p postgresView) DSN() string {
	return "postgres://" + p.user + ":" + p.password + "@" + p.host + ":" + p.port + "/" + p.database
}

type redisView struct {
	addr string
	ttl  int
}

func (r redisView) Addr() string             { return r.addr }
func (r redisView) UserCacheTTLMinutes() int { return r.ttl }

type jwtView struct {
	secret        string
	hours         int
	accessMinutes int
	refreshDays   int
}

func (j jwtView) Secret() string     { return j.secret }
func (j jwtView) ExpHours() int      { return j.hours }
func (j jwtView) AccessMinutes() int { return j.accessMinutes }
func (j jwtView) RefreshDays() int   { return j.refreshDays }

type loggerView struct {
	level  string
	format string
}

func (l loggerView) Level() string  { return l.level }
func (l loggerView) Format() string { return l.format }

func (c *Config) GRPCConfig() GRPCConfig {
	return grpcView{port: c.GRPCPort}
}

func (c *Config) MetricsConfig() MetricsConfig {
	return metricsView{port: c.MetricsPort}
}

func (c *Config) PostgresConfig() PostgresConfig {
	return postgresView{
		host: c.DBHost, port: c.DBPort, user: c.DBUser,
		password: c.DBPassword, database: c.DBName,
	}
}

func (c *Config) RedisConfig() RedisConfig {
	return redisView{addr: c.RedisAddr, ttl: c.UserCacheTTLMinutes}
}

func (c *Config) JWTConfig() JWTConfig {
	access := c.AccessTokenMinutes
	if access <= 0 {
		access = 15
	}
	refresh := c.RefreshTokenDays
	if refresh <= 0 {
		refresh = 7
	}
	return jwtView{
		secret:        c.JWTSecret,
		hours:         c.JWTExpHours,
		accessMinutes: access,
		refreshDays:   refresh,
	}
}

func (c *Config) LoggerConfig() LoggerConfig {
	return loggerView{level: getEnv("LOG_LEVEL", "info"), format: getEnv("LOG_FORMAT", "text")}
}
[denismatveev@c0der event_horizon]$ 
