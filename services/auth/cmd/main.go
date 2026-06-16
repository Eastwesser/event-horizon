package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    // "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

    "event_horizon/services/auth/internal/config"
    "event_horizon/services/auth/internal/handler"
    "event_horizon/services/auth/internal/repository"
    "event_horizon/services/auth/internal/service"
    pb "event_horizon/services/auth/proto"
)

// Инициализация OpenTelemetry для Jaeger
func initTracer(ctx context.Context) (func(context.Context) error, error) {
    // Создаём экспортёр для OTLP gRPC (Jaeger)
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    // Создаём TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("auth"),
            attribute.String("environment", "development"),
        )),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp.Shutdown, nil
}

func main() {
    cfg := config.Load()

    ctx := context.Background()

    // Инициализация Jaeger
    shutdown, err := initTracer(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize tracer: %v", err)
    }
    defer shutdown(ctx)

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

    // Создаём gRPC сервер с трейсингом
    // grpcServer := grpc.NewServer(
    //     grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
    //     grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    // )
    grpcServer := grpc.NewServer()
    pb.RegisterAuthServiceServer(grpcServer, authHandler)
    reflection.Register(grpcServer)

    // Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:9091/metrics")
        if err := http.ListenAndServe(":9091", nil); err != nil {
            log.Printf("Metrics server error: %v", err)
        }
    }()

    // Запускаем
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Запуск в горутине
    go func() {
        log.Printf("🔐 Auth service listening on :%s", cfg.GRPCPort)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down auth service gracefully...")

    // Даём время завершить текущие запросы
    time.Sleep(2 * time.Second)

    grpcServer.GracefulStop()
    dbpool.Close()

    log.Println("Auth service stopped gracefully")
}
