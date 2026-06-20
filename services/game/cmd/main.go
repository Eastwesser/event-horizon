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

    "github.com/nats-io/nats.go"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

    "github.com/Eastwesser/event-horizon/services/game/internal/config"
    "github.com/Eastwesser/event-horizon/services/game/internal/handler"
    "github.com/Eastwesser/event-horizon/services/game/internal/repository"
    "github.com/Eastwesser/event-horizon/services/game/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

// Инициализация OpenTelemetry для Jaeger
func initTracer(ctx context.Context) (func(context.Context) error, error) {
    // Читаем эндпоинт из переменной окружения
    endpoint := os.Getenv("JAEGER_ENDPOINT")
    if endpoint == "" {
        endpoint = "localhost:4317"
    }
    log.Printf("🔄 Initializing Jaeger tracer with endpoint: %s", endpoint)

    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        log.Printf("❌ Failed to create exporter: %v", err)
        return nil, err
    }
    log.Println("✅ Jaeger exporter created")

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("game"),
            attribute.String("environment", "development"),
        )),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    log.Println("✅ Jaeger tracer initialized")
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

    // Подключаемся к NATS
    nc, err := nats.Connect(cfg.NATSUrl)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Drain()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём stream для score.updated
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "SCORES",
        Subjects: []string{"score.updated"},
        Storage:  nats.FileStorage,
        MaxAge:   24 * time.Hour,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    }

    // Репозиторий и сервис
    gameRepo := repository.NewPostgresGameRepo()
    gameService := service.NewGameService(gameRepo, js)
    gameHandler := handler.NewGameHandler(gameService)

    // gRPC сервер с интерсепторами для трейсинга
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    pb.RegisterGameServiceServer(grpcServer, gameHandler)
    reflection.Register(grpcServer)

    // Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:9092/metrics")
        if err := http.ListenAndServe(":9092", nil); err != nil {
            log.Printf("Game metrics server error: %v", err)
        }
    }()

    // Listener
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Запуск gRPC сервера в горутине
    go func() {
        log.Printf("🎮 Game service listening on :%s", cfg.GRPCPort)
        log.Printf("   NATS: %s", cfg.NATSUrl)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down game service gracefully...")

    grpcServer.GracefulStop()
    nc.Drain()

    log.Println("Game service stopped gracefully")
}
