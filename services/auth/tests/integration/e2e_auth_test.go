//go:build integration

package integration_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Eastwesser/event-horizon/services/auth/internal/handler"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/auth/migrations"
)

// E2E-style black-box: real Postgres + in-process gRPC Register→Login→Validate.
func TestAuthGRPC_RegisterLoginValidate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	if err := migrator.Up(stdlib.OpenDBFromPool(pool), migrations.FS); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewPostgresUserRepo(pool)
	svc := service.NewAuthService(repo, nil, "w4-test-secret", 1)
	api := handler.NewAuthHandler(svc)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, api)
	go func() { _ = s.Serve(lis) }()
	defer s.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewAuthServiceClient(conn)

	email := "w4-e2e-" + time.Now().Format("150405.000") + "@example.com"
	password := "secret-pass-12"
	if os.Getenv("AUTH_E2E_SKIP") == "1" {
		t.Skip("AUTH_E2E_SKIP=1")
	}

	reg, err := client.Register(ctx, &pb.RegisterRequest{Email: email, Password: password, Role: "user"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if reg.GetUserId() == "" {
		t.Fatal("empty user id")
	}

	login, err := client.Login(ctx, &pb.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.GetAccessToken() == "" {
		t.Fatal("empty token")
	}

	val, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: login.GetAccessToken()})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if val.GetUserId() != reg.GetUserId() {
		t.Fatalf("user mismatch: %s vs %s", val.GetUserId(), reg.GetUserId())
	}
}
