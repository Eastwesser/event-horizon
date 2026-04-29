package main

import (
	"context"
	"log"
	"net"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	
	"event_horizon/services/auth/internal/config"
	"event_horizon/services/auth/internal/handler"
	"event_horizon/services/auth/internal/repository"
	"event_horizon/services/auth/internal/service"
	pb "event_horizon/services/auth/proto"
)

func main() {
	cfg := config.Load()
	
	// Подключаемся к PostgreSQL
	dbURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()
	
	// Инициализируем слои
	userRepo := repository.NewPostgresUserRepo(dbpool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpHours)
	authHandler := handler.NewAuthHandler(authService)
	
	// Создаём gRPC сервер
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)
	
	// Запускаем
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	log.Printf("Auth service listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
