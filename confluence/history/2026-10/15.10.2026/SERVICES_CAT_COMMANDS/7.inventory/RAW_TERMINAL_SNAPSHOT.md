[denismatveev@c0der event_horizon]$ find services/inventory -type f -name '*.go' | sort
echo "===== main ====="
cat services/inventory/cmd/main.go
echo "===== app ====="
for f in services/inventory/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/inventory/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/inventory/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/inventory/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/inventory/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== seed ====="
for f in services/inventory/internal/seed/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/inventory/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/inventory/cmd/main.go
services/inventory/internal/app/app.go
services/inventory/internal/app/di.go
services/inventory/internal/config/config.go
services/inventory/internal/converter/item.go
services/inventory/internal/converter/item_test.go
services/inventory/internal/handler/grpc_handler.go
services/inventory/internal/interceptor/logger.go
services/inventory/internal/interceptor/recovery.go
services/inventory/internal/interceptor/role.go
services/inventory/internal/interceptor/validate.go
services/inventory/internal/model/errors.go
services/inventory/internal/model/errors_test.go
services/inventory/internal/model/item.go
services/inventory/internal/model/stats.go
services/inventory/internal/repository/cached_repo.go
services/inventory/internal/repository/cached_repo_test.go
services/inventory/internal/repository/mongo_repo.go
services/inventory/internal/repository/mongo_repo_test.go
services/inventory/internal/repository/outbox.go
services/inventory/internal/repository/postgres_integration_test.go
services/inventory/internal/repository/postgres_repo.go
services/inventory/internal/repository/redis_repo.go
services/inventory/internal/repository/repository.go
services/inventory/internal/service/inventory_service.go
services/inventory/internal/service/inventory_service_test.go
services/inventory/internal/worker/outbox_worker.go
services/inventory/migrations/embed.go
services/inventory/proto/inventory_grpc.pb.go
services/inventory/proto/inventory.pb.go
services/inventory/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init inventory app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("inventory stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/inventory/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/config"
)

// App is the runnable inventory service process (gRPC + metrics HTTP).
type App struct {
	cfg           *config.Config
	log           *slog.Logger
	closer        *closer.Closer
	di            *diContainer
	grpcServer    *grpc.Server
	metricsServer *http.Server
	listener      net.Listener
}

// New wires dependencies and prepares servers (does not Serve yet).
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
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
		a.log.Info("inventory gRPC listening",
			"port", a.cfg.GRPCPort,
			"driver", a.cfg.Driver,
		)
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
--- services/inventory/internal/app/di.go ---
package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/config"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/handler"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/service"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/worker"
	invMigrations "github.com/Eastwesser/event-horizon/services/inventory/migrations"
	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

type diContainer struct {
	cfg *config.Config

	db          *sql.DB
	mongoClient *mongo.Client
	nc          *nats.Conn
	js          nats.JetStreamContext
	baseRepo    repository.InventoryRepository
	cache       *repository.RedisCacheRepo
	repo        repository.InventoryRepository
	outbox      repository.ItemOutboxWriter
	svc         *service.InventoryService
	api         pb.InventoryServiceServer

	workerCancel context.CancelFunc
}

func newDiContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (a *App) init(ctx context.Context) error {
	steps := []func(context.Context) error{
		a.initLogger,
		a.initCloser,
		a.initTracer,
		a.initRepository,
		a.initNATS,
		a.initDomain,
		a.initGRPC,
		a.initMetricsHTTP,
		a.initOutboxWorker,
	}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	a.log = logger.New(logger.DefaultConfig())
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	a.closer = closer.New(a.log)
	return nil
}

func (a *App) initTracer(ctx context.Context) error {
	endpoint := getenv("JAEGER_ENDPOINT", "localhost:4317")
	shutdown, err := tracing.Init(ctx, tracing.Config{
		ServiceName: "inventory",
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

func (a *App) initRepository(ctx context.Context) error {
	switch a.cfg.Driver {
	case "postgres":
		a.log.Info("using PostgreSQL driver")
		repo, db, err := a.connectPostgres()
		if err != nil {
			return err
		}
		a.di.baseRepo = repo
		a.di.repo = repo
		a.di.db = db
		a.closer.AddNamed("postgres", func(context.Context) error {
			return db.Close()
		})
	case "mongo":
		a.log.Info("using MongoDB driver")
		repo, client, err := a.connectMongo(ctx)
		if err != nil {
			return err
		}
		a.di.baseRepo = repo
		a.di.repo = repo
		a.di.outbox = nil
		a.di.mongoClient = client
		a.closer.AddNamed("mongodb", func(ctx context.Context) error {
			return client.Disconnect(ctx)
		})
	default:
		return fmt.Errorf("unknown driver: %s (use 'postgres' or 'mongo')", a.cfg.Driver)
	}
	return nil
}

func (a *App) connectPostgres() (repository.InventoryRepository, *sql.DB, error) {
	db, err := sql.Open("postgres", a.cfg.PGDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("postgres ping: %w", err)
	}
	if err := migrator.Up(db, invMigrations.FS); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("migrations: %w", err)
	}
	a.log.Info("postgres ready, migrations applied")
	return repository.NewPostgresRepo(db), db, nil
}

func (a *App) connectMongo(ctx context.Context) (repository.InventoryRepository, *mongo.Client, error) {
	mongoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(a.cfg.MongoURI))
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(mongoCtx, nil); err != nil {
		_ = client.Disconnect(mongoCtx)
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}
	a.log.Info("mongodb ready")
	return repository.NewMongoRepo(client.Database(a.cfg.MongoDBName)), client, nil
}

func (a *App) initNATS(_ context.Context) error {
	nc, err := nats.Connect(a.cfg.NATSURL)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("jetstream: %w", err)
	}
	a.di.nc = nc
	a.di.js = js
	a.closer.AddNamed("nats", func(context.Context) error {
		return nc.Drain()
	})
	a.log.Info("nats connected", "url", a.cfg.NATSURL)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	cache := repository.NewRedisCacheRepo(a.cfg.RedisAddr, 5*time.Minute)
	a.di.cache = cache
	a.closer.AddNamed("redis cache", func(context.Context) error {
		return cache.Close()
	})
	a.di.repo = repository.NewCachedRepository(a.di.baseRepo, cache)
	if ow, ok := a.di.repo.(repository.ItemOutboxWriter); ok {
		a.di.outbox = ow
	}
	a.di.svc = service.NewInventoryService(a.di.repo, a.di.outbox, a.di.js)
	a.di.api = handler.NewGRPCHandler(a.di.svc)
	return nil
}

func (a *App) initGRPC(_ context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	a.listener = lis
	a.closer.AddNamed("tcp listener", func(context.Context) error {
		return lis.Close()
	})

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Logger(),
			interceptor.Validate(),
			interceptor.RequireRoles(
				[]string{"author", "admin"},
				[]string{
					"/CreateItem", "/CreateItems", "/UpdateItem", "/DeleteItem",
					"/SoftDeleteItem", "/RestoreItem", "/GetStats",
				},
			),
			// Shop purchase/cancel stock sync — any authenticated role.
			interceptor.RequireRoles(
				[]string{"user", "author", "admin"},
				[]string{"/ReserveItem", "/ReleaseItem"},
			),
			metrics.UnaryServerInterceptor("inventory"),
		),
	)
	pb.RegisterInventoryServiceServer(grpcServer, a.di.api)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("inventory")
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	cfg := a.cfg
	db := a.di.db
	mongoClient := a.di.mongoClient

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"inventory","driver":"%s"}`, cfg.Driver)
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := "ready"
		httpStatus := http.StatusOK

		if cfg.Driver == "postgres" && db != nil {
			if err := db.Ping(); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				a.log.Warn("ready check: postgres ping failed", "err", err)
			}
		}
		if cfg.Driver == "mongo" && mongoClient != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := mongoClient.Ping(pingCtx, nil); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				a.log.Warn("ready check: mongodb ping failed", "err", err)
			}
		}

		w.WriteHeader(httpStatus)
		fmt.Fprintf(w, `{"status":"%s","service":"inventory","driver":"%s"}`, status, cfg.Driver)
	})

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: mux}
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

func (a *App) initOutboxWorker(_ context.Context) error {
	if a.cfg.Driver != "postgres" || a.di.db == nil {
		return nil
	}
	workerCtx, cancel := context.WithCancel(context.Background())
	a.di.workerCancel = cancel
	a.closer.AddNamed("outbox worker", func(context.Context) error {
		cancel()
		return nil
	})
	outboxWorker := worker.NewOutboxWorker(a.di.db, a.di.js)
	go 	outboxWorker.Start(workerCtx)
	a.log.Info("outbox worker started")
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
===== handler =====
--- services/inventory/internal/handler/grpc_handler.go ---
package handler

import (
    "context"
    "errors"
    "strconv"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/converter"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

type GRPCHandler struct {
    pb.UnimplementedInventoryServiceServer
    service *service.InventoryService
}

func NewGRPCHandler(svc *service.InventoryService) *GRPCHandler {
    return &GRPCHandler{
        service: svc,
    }
}

func (h *GRPCHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.ItemResponse, error) {
    item := converter.ItemFromCreateRequest(req)

    if err := h.service.CreateItem(ctx, item); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create item: %v", err)
    }

    protoItem, err := converter.ItemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.ItemResponse, error) {
    item, err := h.service.GetItem(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
    }

    protoItem, err := converter.ItemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.ItemResponse, error) {
    item := converter.ItemFromUpdateRequest(req)

    if err := h.service.UpdateItem(ctx, item); err != nil {
        return nil, mapInventoryErr(err, "failed to update item")
    }

    protoItem, err := converter.ItemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.DeleteItem(ctx, req.Id); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to delete item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

func (h *GRPCHandler) SearchItems(ctx context.Context, req *pb.SearchItemsRequest) (*pb.SearchItemsResponse, error) {
    filters := make(map[string]interface{})
    for k, v := range req.Filters {
        if k == "price_min" || k == "price_max" {
            if val, err := strconv.ParseFloat(v, 64); err == nil {
                filters[k] = val
            }
        } else {
            filters[k] = v
        }
    }

    items, total, err := h.service.SearchItems(ctx, filters, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to search items: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := converter.ItemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

func (h *GRPCHandler) GetByAuthor(ctx context.Context, req *pb.GetByAuthorRequest) (*pb.SearchItemsResponse, error) {
    items, total, err := h.service.GetByAuthor(ctx, req.AuthorId, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get items by author: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := converter.ItemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

func (h *GRPCHandler) GetByType(ctx context.Context, req *pb.GetByTypeRequest) (*pb.SearchItemsResponse, error) {
    items, total, err := h.service.GetByType(ctx, req.Type, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get items by type: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := converter.ItemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

// BulkCreateItems - массовое создание товаров
func (h *GRPCHandler) BulkCreateItems(ctx context.Context, req *pb.BulkCreateItemsRequest) (*pb.BulkCreateItemsResponse, error) {
    items := make([]*model.Item, len(req.Items))
    for i, pbItem := range req.Items {
        items[i] = converter.ItemFromCreateRequest(pbItem)
    }

    if err := h.service.BulkCreateItems(ctx, items); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to bulk create items: %v", err)
    }

    return &pb.BulkCreateItemsResponse{
        Success: true,
        Count:   int32(len(items)),
    }, nil
}

// ReserveItem - резервирование товара
func (h *GRPCHandler) ReserveItem(ctx context.Context, req *pb.ReserveItemRequest) (*pb.ReserveItemResponse, error) {
    remaining, err := h.service.ReserveItem(ctx, req.Id, int(req.Quantity))
    if err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        if err == model.ErrNotEnoughStock {
            return nil, status.Errorf(codes.FailedPrecondition, "not enough stock: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to reserve item: %v", err)
    }

    return &pb.ReserveItemResponse{
        Success:        true,
        RemainingStock: int32(remaining),
    }, nil
}

// ReleaseItem - возврат товара на склад (refund / cancel)
func (h *GRPCHandler) ReleaseItem(ctx context.Context, req *pb.ReleaseItemRequest) (*pb.ReleaseItemResponse, error) {
    remaining, err := h.service.ReleaseItem(ctx, req.Id, int(req.Quantity))
    if err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to release item: %v", err)
    }

    return &pb.ReleaseItemResponse{
        Success:        true,
        RemainingStock: int32(remaining),
    }, nil
}

// SoftDeleteItem - мягкое удаление
func (h *GRPCHandler) SoftDeleteItem(ctx context.Context, req *pb.SoftDeleteItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.SoftDeleteItem(ctx, req.Id); err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to soft delete item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

// RestoreItem - восстановление после мягкого удаления
func (h *GRPCHandler) RestoreItem(ctx context.Context, req *pb.RestoreItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.RestoreItem(ctx, req.Id); err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to restore item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

// GetStats - статистика
func (h *GRPCHandler) GetStats(ctx context.Context, req *pb.EmptyRequest) (*pb.StatsResponse, error) {
    stats, err := h.service.GetStats(ctx)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
    }

    top := make([]*pb.TopItem, 0, len(stats.TopExpensive))
    for _, item := range stats.TopExpensive {
        if item == nil {
            continue
        }
        top = append(top, &pb.TopItem{
            Id:       item.ID,
            Name:     item.Name,
            Price:    item.Price,
            AuthorId: item.AuthorID,
        })
    }
    return &pb.StatsResponse{
        TotalItems:   stats.TotalItems,
        ByType:       stats.ByType,
        ByAuthor:     stats.ByAuthor,
        TotalStock:   stats.TotalStock,
        TopExpensive: top,
    }, nil
}
func mapInventoryErr(err error, fallback string) error {
	switch {
	case errors.Is(err, model.ErrVersionConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, model.ErrItemNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrNotEnoughStock):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, model.ErrInvalidItem):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "%s: %v", fallback, err)
	}
}
===== service =====
--- services/inventory/internal/service/inventory_service.go ---
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/nats-io/nats.go"
)

// InventoryService — бизнес-логика сервиса инвентаря.
// Кеширование — в repository.CachedRepository (decorator над InventoryRepository).
type InventoryService struct {
	repo   repository.InventoryRepository
	outbox repository.ItemOutboxWriter // PostgreSQL outbox; nil for MongoDB
	js     nats.JetStreamContext
}

// NewInventoryService создает новый сервис инвентаря.
func NewInventoryService(
	repo repository.InventoryRepository,
	outbox repository.ItemOutboxWriter,
	js nats.JetStreamContext,
) *InventoryService {
	return &InventoryService{
		repo:   repo,
		outbox: outbox,
		js:     js,
	}
}

// GetItem получает товар по ID (cache-aside в CachedRepository).
func (s *InventoryService) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.repo.GetItem(ctx, id)
}

// CreateItem создает товар. Для PostgreSQL использует Outbox, для MongoDB — прямой Insert.
func (s *InventoryService) CreateItem(ctx context.Context, item *model.Item) error {
	if item.AuthorID == "" {
		return fmt.Errorf("author_id is required")
	}
	if item.Type == "" {
		return fmt.Errorf("type is required")
	}
	if item.Name == "" {
		return fmt.Errorf("name is required")
	}
	if item.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if item.Stock < 0 {
		return fmt.Errorf("stock cannot be negative")
	}

	if s.outbox != nil {
		event := map[string]interface{}{
			"event":      "item.created",
			"item_id":    "",
			"author_id":  item.AuthorID,
			"type":       item.Type,
			"name":       item.Name,
			"price":      item.Price,
			"stock":      item.Stock,
			"attributes": item.Attributes,
			"images":     item.Images,
			"timestamp":  time.Now().Unix(),
		}
		eventPayload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		return s.outbox.CreateItemWithOutbox(ctx, item, "inventory.item.created", eventPayload)
	}

	return s.repo.CreateItem(ctx, item)
}

// UpdateItem обновляет товар.
func (s *InventoryService) UpdateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		return fmt.Errorf("id is required")
	}
	if item.Version <= 0 {
		cur, err := s.repo.GetItem(ctx, item.ID)
		if err != nil {
			return err
		}
		item.Version = cur.Version
	}
	return s.repo.UpdateItem(ctx, item)
}

// DeleteItem удаляет товар.
func (s *InventoryService) DeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.DeleteItem(ctx, id)
}

// SearchItems выполняет поиск (кеш — в CachedRepository).
func (s *InventoryService) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, filters, limit, offset)
}

// GetByAuthor возвращает товары автора (с пагинацией).
func (s *InventoryService) GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, limit, offset)
}

// GetByType возвращает товары по типу (с пагинацией).
func (s *InventoryService) GetByType(ctx context.Context, itemType string, limit, offset int) ([]*model.Item, int64, error) {
	return s.repo.SearchItems(ctx, map[string]interface{}{"type": itemType}, limit, offset)
}

// BulkCreateItems - массовое создание товаров
func (s *InventoryService) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if item.AuthorID == "" {
			return fmt.Errorf("author_id is required for all items")
		}
		if item.Type == "" {
			return fmt.Errorf("type is required for all items")
		}
		if item.Name == "" {
			return fmt.Errorf("name is required for all items")
		}
		if item.Price < 0 {
			return fmt.Errorf("price cannot be negative")
		}
		if item.Stock < 0 {
			return fmt.Errorf("stock cannot be negative")
		}
	}

	return s.repo.BulkCreateItems(ctx, items)
}

// ReserveItem - резервирование товара (уменьшение stock)
func (s *InventoryService) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	if id == "" {
		return 0, fmt.Errorf("id is required")
	}
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	return s.repo.ReserveItem(ctx, id, quantity)
}

// ReleaseItem - возврат товара на склад (увеличение stock)
func (s *InventoryService) ReleaseItem(ctx context.Context, id string, quantity int) (int, error) {
	if id == "" {
		return 0, fmt.Errorf("id is required")
	}
	if quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	return s.repo.ReleaseItem(ctx, id, quantity)
}

// SoftDeleteItem - мягкое удаление
func (s *InventoryService) SoftDeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.SoftDeleteItem(ctx, id)
}

// RestoreItem - восстановление после мягкого удаления
func (s *InventoryService) RestoreItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.RestoreItem(ctx, id)
}

// GetStats - статистика по товарам
func (s *InventoryService) GetStats(ctx context.Context) (*model.Stats, error) {
	return s.repo.GetStats(ctx)
}
--- services/inventory/internal/service/inventory_service_test.go ---
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

type memRepo struct {
	items          map[string]*model.Item
	lastFilters    map[string]interface{}
	lastLimit      int
	lastOffset     int
	reserveCalls   int
	releaseCalls   int
	softDeleteIDs  []string
	restoreIDs     []string
	bulkCalls      int
	searchErr      error
	reserveErr     error
	releaseErr     error
	softErr        error
	restoreErr     error
}

func (m *memRepo) CreateItem(_ context.Context, item *model.Item) error {
	if m.items == nil {
		m.items = map[string]*model.Item{}
	}
	cp := *item
	m.items[item.ID] = &cp
	return nil
}
func (m *memRepo) GetItem(_ context.Context, id string) (*model.Item, error) {
	it, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *it
	return &cp, nil
}
func (m *memRepo) UpdateItem(_ context.Context, item *model.Item) error {
	cp := *item
	m.items[item.ID] = &cp
	return nil
}
func (m *memRepo) DeleteItem(context.Context, string) error { return nil }
func (m *memRepo) SearchItems(_ context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	m.lastFilters = filters
	m.lastLimit = limit
	m.lastOffset = offset
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	includeDeleted, _ := filters["include_deleted"].(bool)
	var out []*model.Item
	for _, it := range m.items {
		if author, ok := filters["author_id"].(string); ok && author != "" && it.AuthorID != author {
			continue
		}
		if typ, ok := filters["type"].(string); ok && typ != "" && it.Type != typ {
			continue
		}
		if id, ok := filters["id"].(string); ok && id != "" && it.ID != id {
			continue
		}
		if !includeDeleted && it.DeletedAt != nil {
			continue
		}
		cp := *it
		out = append(out, &cp)
	}
	return out, int64(len(out)), nil
}
func (m *memRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
	return m.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}
func (m *memRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
	return m.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}
func (m *memRepo) BulkCreateItems(_ context.Context, items []*model.Item) error {
	m.bulkCalls++
	for _, it := range items {
		_ = m.CreateItem(context.Background(), it)
	}
	return nil
}
func (m *memRepo) ReserveItem(_ context.Context, id string, quantity int) (int, error) {
	m.reserveCalls++
	if m.reserveErr != nil {
		return 0, m.reserveErr
	}
	it, ok := m.items[id]
	if !ok {
		return 0, errors.New("not found")
	}
	it.Stock -= quantity
	return it.Stock, nil
}
func (m *memRepo) ReleaseItem(_ context.Context, id string, quantity int) (int, error) {
	m.releaseCalls++
	if m.releaseErr != nil {
		return 0, m.releaseErr
	}
	it, ok := m.items[id]
	if !ok {
		return 0, errors.New("not found")
	}
	it.Stock += quantity
	return it.Stock, nil
}
func (m *memRepo) SoftDeleteItem(_ context.Context, id string) error {
	m.softDeleteIDs = append(m.softDeleteIDs, id)
	if m.softErr != nil {
		return m.softErr
	}
	it, ok := m.items[id]
	if !ok {
		return errors.New("not found")
	}
	now := time.Now().UTC()
	it.DeletedAt = &now
	return nil
}
func (m *memRepo) RestoreItem(_ context.Context, id string) error {
	m.restoreIDs = append(m.restoreIDs, id)
	if m.restoreErr != nil {
		return m.restoreErr
	}
	it, ok := m.items[id]
	if !ok {
		return errors.New("not found")
	}
	it.DeletedAt = nil
	return nil
}
func (m *memRepo) GetStats(context.Context) (*model.Stats, error) { return &model.Stats{}, nil }

func TestUpdateItem_LoadsVersionWhenOmitted(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Name: "old", Version: 3, CreatedAt: now, UpdatedAt: now},
	}}
	svc := NewInventoryService(repo, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "i1", Name: "new", Version: 0})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Version != 3 || got.Name != "new" {
		t.Fatalf("%+v", got)
	}
}

func TestUpdateItem_KeepsExplicitVersion(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Name: "old", Version: 3, CreatedAt: now, UpdatedAt: now},
	}}
	svc := NewInventoryService(repo, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "i1", Name: "new", Version: 9})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Version != 9 {
		t.Fatalf("version=%d", got.Version)
	}
}

func TestUpdateItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.UpdateItem(context.Background(), &model.Item{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateItem_LoadVersionRepoError(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	err := svc.UpdateItem(context.Background(), &model.Item{ID: "missing", Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	_, err := svc.GetItem(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetItem_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{"i1": {ID: "i1", Name: "n"}}}
	svc := NewInventoryService(repo, nil, nil)
	got, err := svc.GetItem(context.Background(), "i1")
	if err != nil || got.Name != "n" {
		t.Fatalf("got %+v err=%v", got, err)
	}
}

func TestCreateItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	base := &model.Item{AuthorID: "a1", Type: "t", Name: "n", Price: 1, Stock: 1}
	cases := []struct {
		name string
		item *model.Item
	}{
		{"missing_author", func() *model.Item { i := *base; i.AuthorID = ""; return &i }()},
		{"missing_type", func() *model.Item { i := *base; i.Type = ""; return &i }()},
		{"missing_name", func() *model.Item { i := *base; i.Name = ""; return &i }()},
		{"negative_price", func() *model.Item { i := *base; i.Price = -1; return &i }()},
		{"negative_stock", func() *model.Item { i := *base; i.Stock = -1; return &i }()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := svc.CreateItem(context.Background(), tc.item); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCreateItem_ViaRepo(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{}}
	svc := NewInventoryService(repo, nil, nil)
	item := &model.Item{ID: "i1", AuthorID: "a1", Type: "t", Name: "n", Price: 10, Stock: 5}
	if err := svc.CreateItem(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.GetItem(context.Background(), "i1")
	if got.Name != "n" {
		t.Fatalf("%+v", got)
	}
}

type stubOutbox struct {
	called bool
}

func (s *stubOutbox) CreateItemWithOutbox(_ context.Context, _ *model.Item, eventType string, _ []byte) error {
	s.called = true
	if eventType != "inventory.item.created" {
		return errors.New("bad event type")
	}
	return nil
}

func TestCreateItem_ViaOutbox(t *testing.T) {
	ob := &stubOutbox{}
	svc := NewInventoryService(&memRepo{}, ob, nil)
	item := &model.Item{AuthorID: "a1", Type: "t", Name: "n", Price: 1, Stock: 1}
	if err := svc.CreateItem(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	if !ob.called {
		t.Fatal("outbox not called")
	}
}

func TestDeleteItem_RequiresID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.DeleteItem(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestReserveItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if _, err := svc.ReserveItem(context.Background(), "", 1); err == nil {
		t.Fatal("empty id")
	}
	if _, err := svc.ReserveItem(context.Background(), "i1", 0); err == nil {
		t.Fatal("zero qty")
	}
	if _, err := svc.ReserveItem(context.Background(), "i1", -1); err == nil {
		t.Fatal("negative qty")
	}
}

func TestSoftDeleteAndRestore_RequireID(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if err := svc.SoftDeleteItem(context.Background(), ""); err == nil {
		t.Fatal("soft delete")
	}
	if err := svc.RestoreItem(context.Background(), ""); err == nil {
		t.Fatal("restore")
	}
}

func TestBulkCreateItems_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	ok := &model.Item{AuthorID: "a", Type: "t", Name: "n", Price: 0, Stock: 0}
	cases := []struct {
		name  string
		items []*model.Item
	}{
		{"empty_slice", nil},
		{"missing_author", []*model.Item{{Type: "t", Name: "n"}}},
		{"missing_type", []*model.Item{{AuthorID: "a", Name: "n"}}},
		{"missing_name", []*model.Item{{AuthorID: "a", Type: "t"}}},
		{"negative_price", []*model.Item{{AuthorID: "a", Type: "t", Name: "n", Price: -1}}},
		{"negative_stock", []*model.Item{{AuthorID: "a", Type: "t", Name: "n", Stock: -1}}},
		{"valid", []*model.Item{ok}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.BulkCreateItems(context.Background(), tc.items)
			if tc.name == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if tc.name == "empty_slice" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGetStats_DelegatesToRepo(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	stats, err := svc.GetStats(context.Background())
	if err != nil || stats == nil {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
}

func TestSearchItems_DelegatesFilters(t *testing.T) {
	now := time.Now().UTC()
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "auth-1", Type: "card", Name: "live"},
		"b": {ID: "b", AuthorID: "auth-1", Type: "card", Name: "gone", DeletedAt: &now},
		"c": {ID: "c", AuthorID: "other", Type: "merch", Name: "x"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	got, total, err := svc.SearchItems(context.Background(), map[string]interface{}{
		"author_id": "auth-1", "include_deleted": true,
	}, 20, 0)
	if err != nil || total != 2 || len(got) != 2 {
		t.Fatalf("got=%d total=%d err=%v", len(got), total, err)
	}
	if repo.lastLimit != 20 || repo.lastOffset != 0 {
		t.Fatalf("limit=%d offset=%d", repo.lastLimit, repo.lastOffset)
	}
}

func TestGetByAuthor_UsesSearchFilters(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "auth-1", Type: "card"},
		"b": {ID: "b", AuthorID: "auth-2", Type: "card"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	got, total, err := svc.GetByAuthor(context.Background(), "auth-1", 10, 5)
	if err != nil || total != 1 || len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("got=%+v total=%d err=%v", got, total, err)
	}
	if repo.lastFilters["author_id"] != "auth-1" || repo.lastLimit != 10 || repo.lastOffset != 5 {
		t.Fatalf("filters=%v limit=%d offset=%d", repo.lastFilters, repo.lastLimit, repo.lastOffset)
	}
}

func TestGetByType_UsesSearchFilters(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"a": {ID: "a", AuthorID: "x", Type: "брелок"},
		"b": {ID: "b", AuthorID: "x", Type: "card"},
	}}
	got, total, err := NewInventoryService(repo, nil, nil).GetByType(context.Background(), "брелок", 50, 0)
	if err != nil || total != 1 || got[0].Type != "брелок" {
		t.Fatalf("got=%+v total=%d err=%v", got, total, err)
	}
	if repo.lastFilters["type"] != "брелок" {
		t.Fatalf("filters=%v", repo.lastFilters)
	}
}

func TestReserveAndRelease_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", Stock: 10},
	}}
	svc := NewInventoryService(repo, nil, nil)
	left, err := svc.ReserveItem(context.Background(), "i1", 3)
	if err != nil || left != 7 || repo.reserveCalls != 1 {
		t.Fatalf("left=%d err=%v calls=%d", left, err, repo.reserveCalls)
	}
	left, err = svc.ReleaseItem(context.Background(), "i1", 2)
	if err != nil || left != 9 || repo.releaseCalls != 1 {
		t.Fatalf("left=%d err=%v calls=%d", left, err, repo.releaseCalls)
	}
}

func TestReleaseItem_Validation(t *testing.T) {
	svc := NewInventoryService(&memRepo{}, nil, nil)
	if _, err := svc.ReleaseItem(context.Background(), "", 1); err == nil {
		t.Fatal("empty id")
	}
	if _, err := svc.ReleaseItem(context.Background(), "i1", 0); err == nil {
		t.Fatal("zero qty")
	}
}

func TestSoftDeleteAndRestore_Success(t *testing.T) {
	repo := &memRepo{items: map[string]*model.Item{
		"i1": {ID: "i1", AuthorID: "a", Name: "n"},
	}}
	svc := NewInventoryService(repo, nil, nil)
	if err := svc.SoftDeleteItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
	if repo.items["i1"].DeletedAt == nil || len(repo.softDeleteIDs) != 1 {
		t.Fatal("expected soft delete")
	}
	// default search hides soft-deleted
	got, total, err := svc.GetByAuthor(context.Background(), "a", 10, 0)
	if err != nil || total != 0 || len(got) != 0 {
		t.Fatalf("should hide deleted: got=%d total=%d", len(got), total)
	}
	// include_deleted shows it
	got, total, err = svc.SearchItems(context.Background(), map[string]interface{}{
		"author_id": "a", "include_deleted": true,
	}, 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("include_deleted: total=%d err=%v", total, err)
	}
	if err := svc.RestoreItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
	if repo.items["i1"].DeletedAt != nil || len(repo.restoreIDs) != 1 {
		t.Fatal("expected restore")
	}
}

func TestDeleteItem_Success(t *testing.T) {
	if err := NewInventoryService(&memRepo{}, nil, nil).DeleteItem(context.Background(), "i1"); err != nil {
		t.Fatal(err)
	}
}

func TestReserveRelease_RepoErrorsPropagate(t *testing.T) {
	repo := &memRepo{
		items:      map[string]*model.Item{"i1": {ID: "i1", Stock: 1}},
		reserveErr: errors.New("reserve fail"),
		releaseErr: errors.New("release fail"),
	}
	svc := NewInventoryService(repo, nil, nil)
	if _, err := svc.ReserveItem(context.Background(), "i1", 1); err == nil {
		t.Fatal("expected reserve error")
	}
	repo.reserveErr = nil
	if _, err := svc.ReleaseItem(context.Background(), "i1", 1); err == nil {
		t.Fatal("expected release error")
	}
}
===== repository =====
--- services/inventory/internal/repository/cached_repo.go ---
package repository

import (
	"context"
	"fmt"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

// CachedRepository decorates InventoryRepository with Redis cache-aside (Kozirev W6 pattern).
// RedisCacheRepo remains the low-level storage adapter; this type is the repository decorator.
type CachedRepository struct {
	next  InventoryRepository
	cache *RedisCacheRepo
}

// NewCachedRepository wraps repo with Redis caching. Pass nil cache to disable caching.
func NewCachedRepository(next InventoryRepository, cache *RedisCacheRepo) InventoryRepository {
	if cache == nil || next == nil {
		return next
	}
	return &CachedRepository{next: next, cache: cache}
}

func (c *CachedRepository) CreateItem(ctx context.Context, item *model.Item) error {
	if err := c.next.CreateItem(ctx, item); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

// CreateItemWithOutbox delegates to the wrapped repo when it supports outbox, then warms cache.
func (c *CachedRepository) CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error {
	w, ok := c.next.(ItemOutboxWriter)
	if !ok {
		return fmt.Errorf("underlying repository does not support outbox")
	}
	if err := w.CreateItemWithOutbox(ctx, item, eventType, eventPayload); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) GetItem(ctx context.Context, id string) (*model.Item, error) {
	if item, err := c.cache.GetItem(ctx, id); err == nil {
		return item, nil
	}
	item, err := c.next.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if item != nil {
		_ = c.cache.SetItem(ctx, item)
	}
	return item, nil
}

func (c *CachedRepository) UpdateItem(ctx context.Context, item *model.Item) error {
	if err := c.next.UpdateItem(ctx, item); err != nil {
		return err
	}
	c.warmItem(ctx, item)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) DeleteItem(ctx context.Context, id string) error {
	if err := c.next.DeleteItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	queryKey, _ := filters["query"].(string)
	if queryKey != "" && len(filters) == 1 {
		if cached, err := c.cache.GetSearchResult(ctx, queryKey, limit, offset); err == nil {
			return cached, int64(len(cached)), nil
		}
	}

	items, total, err := c.next.SearchItems(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if queryKey != "" && len(filters) == 1 {
		_ = c.cache.SetSearchResult(ctx, queryKey, limit, offset, items)
	}
	return items, total, nil
}

func (c *CachedRepository) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
	return c.next.GetByAuthor(ctx, authorID)
}

func (c *CachedRepository) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
	return c.next.GetByType(ctx, itemType)
}

func (c *CachedRepository) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if err := c.next.BulkCreateItems(ctx, items); err != nil {
		return err
	}
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	remaining, err := c.next.ReserveItem(ctx, id, quantity)
	if err != nil {
		return 0, err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return remaining, nil
}

func (c *CachedRepository) ReleaseItem(ctx context.Context, id string, quantity int) (int, error) {
	remaining, err := c.next.ReleaseItem(ctx, id, quantity)
	if err != nil {
		return 0, err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return remaining, nil
}

func (c *CachedRepository) SoftDeleteItem(ctx context.Context, id string) error {
	if err := c.next.SoftDeleteItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) RestoreItem(ctx context.Context, id string) error {
	if err := c.next.RestoreItem(ctx, id); err != nil {
		return err
	}
	_ = c.cache.DeleteItem(ctx, id)
	c.invalidateSearch(ctx)
	return nil
}

func (c *CachedRepository) GetStats(ctx context.Context) (*model.Stats, error) {
	return c.next.GetStats(ctx)
}

func (c *CachedRepository) warmItem(ctx context.Context, item *model.Item) {
	if item != nil {
		_ = c.cache.SetItem(ctx, item)
	}
}

func (c *CachedRepository) invalidateSearch(ctx context.Context) {
	_ = c.cache.InvalidateSearchCache(ctx)
}
--- services/inventory/internal/repository/cached_repo_test.go ---
package repository_test

import (
	"context"
	"testing"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	item *model.Item
}

func (s *stubRepo) CreateItem(context.Context, *model.Item) error { return nil }
func (s *stubRepo) GetItem(_ context.Context, id string) (*model.Item, error) {
	if s.item != nil && s.item.ID == id {
		return s.item, nil
	}
	return nil, context.Canceled
}
func (s *stubRepo) UpdateItem(context.Context, *model.Item) error   { return nil }
func (s *stubRepo) DeleteItem(context.Context, string) error        { return nil }
func (s *stubRepo) SearchItems(context.Context, map[string]interface{}, int, int) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) GetByAuthor(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) GetByType(context.Context, string) ([]*model.Item, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) BulkCreateItems(context.Context, []*model.Item) error { return nil }
func (s *stubRepo) ReserveItem(context.Context, string, int) (int, error) {
	return 0, nil
}
func (s *stubRepo) ReleaseItem(context.Context, string, int) (int, error) {
	return 0, nil
}
func (s *stubRepo) SoftDeleteItem(context.Context, string) error { return nil }
func (s *stubRepo) RestoreItem(context.Context, string) error    { return nil }
func (s *stubRepo) GetStats(context.Context) (*model.Stats, error) {
	return &model.Stats{}, nil
}

func TestNewCachedRepository_nilCachePassthrough(t *testing.T) {
	base := &stubRepo{item: &model.Item{ID: "1", Name: "x"}}
	repo := repository.NewCachedRepository(base, nil)
	require.Equal(t, base, repo)
}

func TestCachedRepository_implementsInterface(t *testing.T) {
	base := &stubRepo{item: &model.Item{ID: "1", Name: "x"}}
	var _ repository.InventoryRepository = repository.NewCachedRepository(base, nil)
}
--- services/inventory/internal/repository/mongo_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepo — адаптер для MongoDB, реализующий интерфейс InventoryRepository.
// Использует гибкую схему с attributes map[string]interface{} для хранения
// разнородных товаров (брелоки, картины, фенечки).
type MongoRepo struct {
	collection *mongo.Collection
	client     *mongo.Client // Сохраняем клиент для транзакций и health check
}

// NewMongoRepo создает новый репозиторий и настраивает индексы.
// Индексы критичны для производительности: author_id, type, полнотекстовый поиск по attributes.
func NewMongoRepo(db *mongo.Database) *MongoRepo {
	collection := db.Collection("inventory_items")
	client := db.Client()

	// Создаем индексы для частых запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		// Составной индекс для получения товаров автора с сортировкой по дате
		{
			Keys: bson.D{
				{Key: "author_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		// Индекс для фильтрации по типу товара
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		// Полнотекстовый индекс по атрибутам (веса для релевантности)
		{
			Keys: bson.D{
				{Key: "attributes", Value: "text"},
			},
			Options: options.Index().
				SetDefaultLanguage("russian").
				SetWeights(bson.M{
					"attributes.name":        10,
					"attributes.description": 5,
					"attributes.material":    3,
				}),
		},
		// Составной индекс для фильтрации по типу + цене (сортировка)
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "price", Value: 1},
			},
		},
		// Индекс для поиска по полю id (уникальный)
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Printf("⚠️ Failed to create MongoDB indexes: %v", err)
	}

	return &MongoRepo{
		collection: collection,
		client:     client,
	}
}

// ---------------------- Вспомогательные функции ----------------------

// toMongoFilter преобразует map фильтров в bson.M для MongoDB.
// Поддерживает фильтрацию по author_id, type, price_min/max, query (текст),
// а также вложенные атрибуты с операторами $in, $gt, $lt.
func toMongoFilter(filters map[string]interface{}) bson.M {
	m := bson.M{}

	// Фильтр по автору
	if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
		m["author_id"] = authorID
	}

	// Фильтр по одному типу
	if itemType, ok := filters["type"].(string); ok && itemType != "" {
		m["type"] = itemType
	}

	// Фильтр по нескольким типам (массив)
	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		m["type"] = bson.M{"$in": types}
	}

	// Диапазон цен
	if priceMin, ok := filters["price_min"].(float64); ok && priceMin > 0 {
		m["price"] = bson.M{"$gte": priceMin}
	}
	if priceMax, ok := filters["price_max"].(float64); ok && priceMax > 0 {
		if _, ok := m["price"]; ok {
			m["price"].(bson.M)["$lte"] = priceMax
		} else {
			m["price"] = bson.M{"$lte": priceMax}
		}
	}

	// Фильтр по вложенным атрибутам (attributes.color, attributes.size и т.д.)
	if attrs, ok := filters["attributes"].(map[string]interface{}); ok {
		for k, v := range attrs {
			// Поддержка операторов: {"color": {"$in": ["red", "blue"]}}
			if conditions, ok := v.(map[string]interface{}); ok {
				for op, val := range conditions {
					if op == "$in" || op == "$nin" {
						m["attributes."+k] = bson.M{op: val}
					} else if op == "$gt" || op == "$gte" || op == "$lt" || op == "$lte" {
						m["attributes."+k] = bson.M{op: val}
					} else {
						m["attributes."+k] = v // fallback
					}
				}
			} else {
				m["attributes."+k] = v
			}
		}
	}

	// Полнотекстовый поиск по name и description (через $or)
	if query, ok := filters["query"].(string); ok && query != "" {
		m["$or"] = []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	includeDeleted := false
	switch v := filters["include_deleted"].(type) {
	case string:
		includeDeleted = v == "1" || v == "true" || v == "TRUE"
	case bool:
		includeDeleted = v
	}
	if !includeDeleted {
		m["deleted_at"] = nil
	}

	if id, ok := filters["id"].(string); ok && id != "" {
		m["id"] = id
	}

	return m
}

// mapToItem преобразует bson.M (документ MongoDB) в модель Item.
// Использует хелперы для безопасного извлечения значений разных типов.
func mapToItem(m bson.M) *model.Item {
	item := &model.Item{
		ID:          getString(m, "id"),
		AuthorID:    getString(m, "author_id"),
		Type:        getString(m, "type"),
		Name:        getString(m, "name"),
		Description: getString(m, "description"),
		Price:       getFloat64(m, "price"),
		Stock:       getInt(m, "stock"),
		Images:      getStringSlice(m, "images"),
		CreatedAt:   getTime(m, "created_at"),
		UpdatedAt:   getTime(m, "updated_at"),
	}

	// Извлекаем attributes (хранятся как вложенный bson.M)
	if attrs, ok := m["attributes"].(bson.M); ok {
		item.Attributes = make(map[string]interface{})
		for k, v := range attrs {
			item.Attributes[k] = v
		}
	} else if attrs, ok := m["attributes"].(map[string]interface{}); ok {
		item.Attributes = attrs
	}
	if item.Attributes == nil {
		item.Attributes = map[string]interface{}{}
	}
	if t := getTimePtr(m, "deleted_at"); t != nil {
		item.DeletedAt = t
		item.Attributes["_deleted"] = true
		item.Attributes["_deleted_at"] = t.Format(time.RFC3339)
	}

	return item
}

// ---------------------- Хелперы для безопасного извлечения данных ----------------------

func getString(m bson.M, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat64(m bson.M, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func getInt(m bson.M, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func getStringSlice(m bson.M, key string) []string {
	if v, ok := m[key].(primitive.A); ok {
		result := make([]string, len(v))
		for i, val := range v {
			if s, ok := val.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return []string{}
}

func getTime(m bson.M, key string) time.Time {
	if v, ok := m[key].(primitive.DateTime); ok {
		return v.Time()
	}
	return time.Time{}
}

func getTimePtr(m bson.M, key string) *time.Time {
	if v, ok := m[key].(primitive.DateTime); ok {
		t := v.Time().UTC()
		return &t
	}
	if v, ok := m[key].(time.Time); ok && !v.IsZero() {
		t := v.UTC()
		return &t
	}
	return nil
}

// ---------------------- CRUD-методы (реализация интерфейса) ----------------------

// CreateItem создает новый товар в коллекции.
// Генерирует ID в формате hex, устанавливает временные метки.
func (r *MongoRepo) CreateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		item.ID = primitive.NewObjectID().Hex()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	doc := bson.M{
		"id":          item.ID,
		"author_id":   item.AuthorID,
		"type":        item.Type,
		"name":        item.Name,
		"description": item.Description,
		"price":       item.Price,
		"stock":       item.Stock,
		"attributes":  item.Attributes,
		"images":      item.Images,
		"created_at":  item.CreatedAt,
		"updated_at":  item.UpdatedAt,
		"deleted_at":  nil, // явно указываем, что не удален
	}

	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

// GetItem возвращает товар по ID (игнорирует удаленные).
func (r *MongoRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
	var result bson.M
	err := r.collection.FindOne(ctx, bson.M{"id": id, "deleted_at": nil}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return mapToItem(result), nil
}

// UpdateItem обновляет все поля товара (стратегия PUT).
func (r *MongoRepo) UpdateItem(ctx context.Context, item *model.Item) error {
	item.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"author_id":   item.AuthorID,
			"type":        item.Type,
			"name":        item.Name,
			"description": item.Description,
			"price":       item.Price,
			"stock":       item.Stock,
			"attributes":  item.Attributes,
			"images":      item.Images,
			"updated_at":  item.UpdatedAt,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"id": item.ID, "deleted_at": nil}, update)
	return err
}

// DeleteItem выполняет жесткое удаление товара.
func (r *MongoRepo) DeleteItem(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// SoftDeleteItem выполняет мягкое удаление — устанавливает deleted_at и status.
func (r *MongoRepo) SoftDeleteItem(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"deleted_at": time.Now(),
				"status":     "deleted",
			},
		},
	)
	return err
}

// SearchItems выполняет поиск товаров с фильтрами, пагинацией и сортировкой.
// Автоматически исключает мягко удаленные записи (deleted_at: nil).
func (r *MongoRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	mongoFilter := toMongoFilter(filters)

	// Общее количество (без учета пагинации)
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []*model.Item
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, err
		}
		items = append(items, mapToItem(result))
	}
	return items, total, nil
}

// GetByAuthor возвращает все товары автора (с пагинацией по умолчанию 1000).
func (r *MongoRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}

// GetByType возвращает все товары указанного типа (с пагинацией по умолчанию 1000).
func (r *MongoRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
    return r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}

// ---------------------- Дополнительные методы для MongoDB ----------------------

// ReleaseItem — возвращает quantity на склад (обратная операция к ReserveItem).
func (r *MongoRepo) ReleaseItem(ctx context.Context, itemID string, quantity int) (int, error) {
	res, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": itemID, "deleted_at": nil},
		bson.M{"$inc": bson.M{"stock": quantity}},
	)
	if err != nil {
		return 0, err
	}
	if res.MatchedCount == 0 {
		return 0, model.ErrItemNotFound
	}
	var updated bson.M
	if err := r.collection.FindOne(ctx, bson.M{"id": itemID}).Decode(&updated); err != nil {
		return 0, err
	}
	return getInt(updated, "stock"), nil
}

// ReserveItem — пример использования транзакций в MongoDB.
// Бронирует товар: проверяет остаток, уменьшает stock и создает запись в истории.
// Возвращает новый остаток.
func (r *MongoRepo) ReserveItem(ctx context.Context, itemID string, quantity int) (int, error) {
	session, err := r.client.StartSession()
	if err != nil {
		return 0, err
	}
	defer session.EndSession(ctx)

	var remaining int
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// 1. Проверяем наличие и текущий остаток
		var result bson.M
		err := r.collection.FindOne(sessCtx, bson.M{"id": itemID, "deleted_at": nil}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, model.ErrItemNotFound
			}
			return nil, err
		}

		stock := getInt(result, "stock")
		if stock < quantity {
			return nil, model.ErrNotEnoughStock
		}

		// 2. Уменьшаем остаток
		_, err = r.collection.UpdateOne(
			sessCtx,
			bson.M{"id": itemID},
			bson.M{"$inc": bson.M{"stock": -quantity}},
		)
		if err != nil {
			return nil, err
		}

		// 3. Получаем новый остаток
		var updatedResult bson.M
		err = r.collection.FindOne(sessCtx, bson.M{"id": itemID}).Decode(&updatedResult)
		if err != nil {
			return nil, err
		}
		remaining = getInt(updatedResult, "stock")

		// 4. Создаём запись в истории (другая коллекция)
		historyColl := r.collection.Database().Collection("reservation_history")
		_, err = historyColl.InsertOne(sessCtx, bson.M{
			"item_id":     itemID,
			"quantity":    quantity,
			"remaining":   remaining,
			"reserved_at": time.Now(),
		})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return 0, err
	}

	return remaining, nil
}

// SearchByText — полнотекстовый поиск по индексу attributes.
// Использует $text для поиска по полям, указанным в индексе (name, description, material).
func (r *MongoRepo) SearchByText(ctx context.Context, query string, limit, offset int) ([]*model.Item, int64, error) {
	filter := bson.M{
		"$text":      bson.M{"$search": query},
		"deleted_at": nil,
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []*model.Item
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, 0, err
		}
		items = append(items, mapToItem(result))
	}

	return items, total, nil
}

// BulkCreateItems — массовая вставка товаров для начальной загрузки или импорта.
// Использует BulkWrite для производительности.
func (r *MongoRepo) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	var operations []mongo.WriteModel

	for _, item := range items {
		if item.ID == "" {
			item.ID = primitive.NewObjectID().Hex()
		}
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		doc := bson.M{
			"id":          item.ID,
			"author_id":   item.AuthorID,
			"type":        item.Type,
			"name":        item.Name,
			"description": item.Description,
			"price":       item.Price,
			"stock":       item.Stock,
			"attributes":  item.Attributes,
			"images":      item.Images,
			"created_at":  item.CreatedAt,
			"updated_at":  item.UpdatedAt,
			"deleted_at":  nil,
		}

		operations = append(operations, mongo.NewInsertOneModel().SetDocument(doc))
	}

	_, err := r.collection.BulkWrite(ctx, operations)
	return err
}

// WatchChanges — альтернатива Outbox для MongoDB.
// Слушает изменения коллекции (insert/update) и публикует события в NATS.
// Это позволяет реализовать событийную архитектуру без дополнительной таблицы outbox.
func (r *MongoRepo) WatchChanges(ctx context.Context, js nats.JetStreamContext) error {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType", Value: bson.D{{Key: "$in", Value: []string{"insert", "update"}}}},
		}}},
	}

	stream, err := r.collection.Watch(ctx, pipeline)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var event bson.M
		if err := stream.Decode(&event); err != nil {
			log.Printf("Failed to decode change event: %v", err)
			continue
		}

		payload, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue
		}

		_, err = js.Publish("inventory.item.changed", payload)
		if err != nil {
			log.Printf("Failed to publish change event: %v", err)
		} else {
			log.Printf("📡 Published change event for item: %s", event)
		}
	}

	return stream.Err()
}

// GetItemWithAuthor — пример агрегации с $lookup.
// Объединяет товар с информацией об авторе из коллекции authors.
func (r *MongoRepo) GetItemWithAuthor(ctx context.Context, itemID string) (*model.Item, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"id": itemID, "deleted_at": nil}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "authors"},
			{Key: "localField", Value: "author_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "as", Value: "author"},
		}}},
		{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$author"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "price", Value: 1},
			{Key: "author_name", Value: "$author.name"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, model.ErrItemNotFound
	}

	return mapToItem(results[0]), nil
}

// GetStats — возвращает статистику по товарам в MongoDB.
func (r *MongoRepo) GetStats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{
		ByType:       make(map[string]int64),
		ByAuthor:     make(map[string]int64),
		TopExpensive: []*model.TopItem{},
	}

	// 1. Общее количество + суммарный stock
	pipelineTotals := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "stock", Value: bson.D{{Key: "$sum", Value: "$stock"}}},
		}}},
	}
	curTotals, err := r.collection.Aggregate(ctx, pipelineTotals)
	if err != nil {
		return nil, err
	}
	defer curTotals.Close(ctx)
	var totals []struct {
		Total int64 `bson:"total"`
		Stock int64 `bson:"stock"`
	}
	if err := curTotals.All(ctx, &totals); err != nil {
		return nil, err
	}
	if len(totals) > 0 {
		stats.TotalItems = totals[0].Total
		stats.TotalStock = totals[0].Stock
	}

	// 2. Группировка по типу
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$type"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, row := range results {
		stats.ByType[row.ID] = row.Count
	}

	// 3. Группировка по автору (топ 10)
	pipeline = mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$author_id"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err = r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var authorResults []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	if err := cursor.All(ctx, &authorResults); err != nil {
		return nil, err
	}

	for _, row := range authorResults {
		stats.ByAuthor[row.ID] = row.Count
	}

	// 4. Топ-5 самых дорогих
	pipelineTop := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		{{Key: "$sort", Value: bson.D{{Key: "price", Value: -1}, {Key: "name", Value: 1}}}},
		{{Key: "$limit", Value: 5}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "price", Value: 1},
			{Key: "author_id", Value: 1},
		}}},
	}
	curTop, err := r.collection.Aggregate(ctx, pipelineTop)
	if err != nil {
		return nil, err
	}
	defer curTop.Close(ctx)
	var topRows []struct {
		ID       string  `bson:"id"`
		Name     string  `bson:"name"`
		Price    float64 `bson:"price"`
		AuthorID string  `bson:"author_id"`
	}
	if err := curTop.All(ctx, &topRows); err != nil {
		return nil, err
	}
	for _, row := range topRows {
		stats.TopExpensive = append(stats.TopExpensive, &model.TopItem{
			ID: row.ID, Name: row.Name, Price: row.Price, AuthorID: row.AuthorID,
		})
	}

	return stats, nil
}

// RestoreItem — восстанавливает мягко удаленный товар.
func (r *MongoRepo) RestoreItem(ctx context.Context, id string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"id": id, "deleted_at": bson.M{"$ne": nil}},
		bson.M{
			"$set": bson.M{
				"deleted_at": nil,
				"status":     "active",
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return model.ErrItemNotFound
	}
	return nil
}--- services/inventory/internal/repository/mongo_repo_test.go ---
//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestMongo поднимает тестовую MongoDB (использует testcontainers или локальный экземпляр).
// Для CI/CD рекомендуется использовать testcontainers.
func setupTestMongo(t *testing.T) *MongoRepo {
	// Используем локальную MongoDB (для разработки)
	// В CI можно заменить на testcontainers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Создаем временную базу данных для тестов
	db := client.Database("test_inventory_" + time.Now().Format("20060102150405"))
	repo := NewMongoRepo(db)

	// Очищаем после тестов
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})

	return repo
}

// TestMongoRepo_CreateItem тестирует создание товара.
func TestMongoRepo_CreateItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	item := &model.Item{
		AuthorID:    "author-123",
		Type:        "брелок",
		Name:        "Ключница 'Дракон'",
		Description: "Металлическая ключница с гравировкой",
		Price:       1500.00,
		Stock:       10,
		Attributes: map[string]interface{}{
			"material": "металл",
			"weight":   "150g",
			"color":    "золотой",
		},
		Images: []string{"image1.jpg", "image2.jpg"},
	}

	err := repo.CreateItem(ctx, item)
	assert.NoError(t, err)
	assert.NotEmpty(t, item.ID)
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())

	// Проверяем, что товар сохранился
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.Name, saved.Name)
	assert.Equal(t, item.Price, saved.Price)
	assert.Equal(t, item.Attributes["material"], saved.Attributes["material"])
}

// TestMongoRepo_UpdateItem тестирует обновление товара.
func TestMongoRepo_UpdateItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "картина",
		Name:     "Закат над морем",
		Price:    5000.00,
		Stock:    5,
		Attributes: map[string]interface{}{
			"size":      "50x70cm",
			"technique": "масло",
		},
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Обновляем
	item.Name = "Закат над морем (обновлено)"
	item.Price = 5500.00
	item.Attributes["frame"] = "дерево"

	err = repo.UpdateItem(ctx, item)
	assert.NoError(t, err)

	// Проверяем
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Закат над морем (обновлено)", saved.Name)
	assert.Equal(t, 5500.00, saved.Price)
	assert.Equal(t, "дерево", saved.Attributes["frame"])
}

// TestMongoRepo_SoftDeleteItem тестирует мягкое удаление.
func TestMongoRepo_SoftDeleteItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "фенечка",
		Name:     "Фенечка с бусинами",
		Price:    300.00,
		Stock:    20,
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Мягко удаляем
	err = repo.SoftDeleteItem(ctx, item.ID)
	assert.NoError(t, err)

	// Прямой поиск по ID должен вернуть ошибку (deleted_at != nil)
	_, err = repo.GetItem(ctx, item.ID)
	assert.Error(t, err)

	// Поиск с фильтрами не должен возвращать удаленный товар
	items, total, err := repo.SearchItems(ctx, map[string]interface{}{"author_id": "author-123"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, items, 0)
}

// TestMongoRepo_SearchItems тестирует поиск с фильтрами.
func TestMongoRepo_SearchItems(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем несколько товаров
	items := []*model.Item{
		{
			AuthorID: "author-123",
			Type:     "брелок",
			Name:     "Ключница 'Дракон'",
			Price:    1500.00,
			Stock:    10,
			Attributes: map[string]interface{}{
				"material": "металл",
				"color":    "золотой",
			},
		},
		{
			AuthorID: "author-123",
			Type:     "брелок",
			Name:     "Ключница 'Сова'",
			Price:    1200.00,
			Stock:    8,
			Attributes: map[string]interface{}{
				"material": "дерево",
				"color":    "коричневый",
			},
		},
		{
			AuthorID: "author-456",
			Type:     "картина",
			Name:     "Закат над морем",
			Price:    5000.00,
			Stock:    5,
		},
	}

	for _, item := range items {
		err := repo.CreateItem(ctx, item)
		require.NoError(t, err)
	}

	// Тест 1: Поиск по author_id
	result, total, err := repo.SearchItems(ctx, map[string]interface{}{"author_id": "author-123"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)

	// Тест 2: Поиск по type
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{"type": "брелок"}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Тест 3: Поиск по price_min
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{"price_min": 1300.0}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // 1500 и 5000

	// Тест 4: Поиск по атрибутам
	result, total, err = repo.SearchItems(ctx, map[string]interface{}{
		"attributes": map[string]interface{}{
			"material": "металл",
		},
	}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Ключница 'Дракон'", result[0].Name)
}

// TestMongoRepo_ReserveItem тестирует транзакцию резервирования.
func TestMongoRepo_ReserveItem(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	// Создаем товар с остатком 10
	item := &model.Item{
		AuthorID: "author-123",
		Type:     "брелок",
		Name:     "Тестовый брелок",
		Price:    1000.00,
		Stock:    10,
	}
	err := repo.CreateItem(ctx, item)
	require.NoError(t, err)

	// Резервируем 3 штуки
	left, err := repo.ReserveItem(ctx, item.ID, 3)
	assert.NoError(t, err)
	assert.Equal(t, 7, left)

	// Проверяем остаток
	saved, err := repo.GetItem(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, 7, saved.Stock)

	// Пытаемся зарезервировать больше, чем есть
	_, err = repo.ReserveItem(ctx, item.ID, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough stock")
}

// TestMongoRepo_BulkCreateItems тестирует массовую вставку.
func TestMongoRepo_BulkCreateItems(t *testing.T) {
	repo := setupTestMongo(t)
	ctx := context.Background()

	items := []*model.Item{
		{AuthorID: "author-1", Type: "тип1", Name: "Товар 1", Price: 100, Stock: 5},
		{AuthorID: "author-2", Type: "тип2", Name: "Товар 2", Price: 200, Stock: 10},
		{AuthorID: "author-3", Type: "тип3", Name: "Товар 3", Price: 300, Stock: 15},
	}

	err := repo.BulkCreateItems(ctx, items)
	assert.NoError(t, err)

	// Проверяем, что все создались
	for _, item := range items {
		saved, err := repo.GetItem(ctx, item.ID)
		assert.NoError(t, err)
		assert.Equal(t, item.Name, saved.Name)
	}
}--- services/inventory/internal/repository/outbox.go ---
package repository

import (
	"context"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

// ItemOutboxWriter persists an item and outbox event in one transaction (PostgreSQL).
type ItemOutboxWriter interface {
	CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error
}
--- services/inventory/internal/repository/postgres_integration_test.go ---
//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
	"github.com/Eastwesser/event-horizon/services/inventory/migrations"
)

func inventoryDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("INVENTORY_TEST_DATABASE_URL"); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Fatal(err)
		}
		if err := migrator.Up(db, migrations.FS); err != nil {
			_ = db.Close()
			t.Fatalf("migrate: %v", err)
		}
		return db, func() { _ = db.Close() }
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "inventory_test"},
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
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/inventory_test?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatal(err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := migrator.Up(db, migrations.FS); err != nil {
		_ = db.Close()
		_ = c.Terminate(ctx)
		t.Fatalf("migrate: %v", err)
	}
	return db, func() {
		_ = db.Close()
		_ = c.Terminate(ctx)
	}
}

func TestInventory_ReserveAndVersionConflict(t *testing.T) {
	db, cleanup := inventoryDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(db)
	id := uuid.NewString()
	author := uuid.NewString()
	now := time.Now().UTC()
	item := &model.Item{
		ID: id, AuthorID: author, Type: "брелок", Name: "test",
		Description: "d", Price: 10, Stock: 5, Attributes: map[string]any{},
		Images: []string{}, CreatedAt: now, UpdatedAt: now, Version: 1,
	}
	if err := repo.CreateItem(ctx, item); err != nil {
		t.Fatalf("create: %v", err)
	}

	left, err := repo.ReserveItem(ctx, id, 2)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if left != 3 {
		t.Fatalf("stock left=%d want 3", left)
	}

	got, err := repo.GetItem(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version < 2 {
		t.Fatalf("version after reserve=%d want >=2", got.Version)
	}

	// stale version update must conflict
	stale := *got
	stale.Version = 1
	stale.Name = "stale"
	if err := repo.UpdateItem(ctx, &stale); err != model.ErrVersionConflict {
		t.Fatalf("want ErrVersionConflict, got %v", err)
	}

	_, err = repo.ReserveItem(ctx, id, 100)
	if err != model.ErrNotEnoughStock {
		t.Fatalf("want ErrNotEnoughStock, got %v", err)
	}
}
--- services/inventory/internal/repository/postgres_repo.go ---
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresRepo — адаптер для PostgreSQL, реализующий интерфейс InventoryRepository.
// Использует JSONB для гибких атрибутов и поддерживает транзакции с Outbox.
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo создает новый репозиторий PostgreSQL.
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// ---------------------- Вспомогательные функции ----------------------

// toPostgresFilter turns a fixed set of filter keys into placeholders.
// Column names are written here, never taken from the request. Values go in args.
func toPostgresFilter(filters map[string]interface{}) ([]string, []interface{}, int) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if authorID, ok := filters["author_id"].(string); ok && authorID != "" {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argCount))
		args = append(args, authorID)
		argCount++
	}

	if id, ok := filters["id"].(string); ok && id != "" {
		conditions = append(conditions, fmt.Sprintf("id = $%d", argCount))
		args = append(args, id)
		argCount++
	}

	if itemType, ok := filters["type"].(string); ok && itemType != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
		args = append(args, itemType)
		argCount++
	}

	if types, ok := filters["types"].([]string); ok && len(types) > 0 {
		placeholders := make([]string, len(types))
		for i := range types {
			placeholders[i] = fmt.Sprintf("$%d", argCount)
			args = append(args, types[i])
			argCount++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ", ")))
	}

	if priceMin, ok := filters["price_min"].(float64); ok && priceMin > 0 {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argCount))
		args = append(args, priceMin)
		argCount++
	}
	if priceMax, ok := filters["price_max"].(float64); ok && priceMax > 0 {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argCount))
		args = append(args, priceMax)
		argCount++
	}

	if attrs, ok := filters["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
		attrsJSON, err := json.Marshal(attrs)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("attributes @> $%d", argCount))
			args = append(args, string(attrsJSON))
			argCount++
		}
	}

	if query, ok := filters["query"].(string); ok && query != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
		searchTerm := "%" + query + "%"
		args = append(args, searchTerm)
		argCount++
	}

	includeDeleted := false
	switch v := filters["include_deleted"].(type) {
	case string:
		includeDeleted = v == "1" || strings.EqualFold(v, "true")
	case bool:
		includeDeleted = v
	}
	if !includeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	return conditions, args, argCount
}

// ---------------------- CRUD-методы (реализация интерфейса) ----------------------

// CreateItem создает новый товар в PostgreSQL.
func (r *PostgresRepo) CreateItem(ctx context.Context, item *model.Item) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	query := `
		INSERT INTO inventory_items (
			id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = r.db.ExecContext(ctx, query,
		item.ID, item.AuthorID, item.Type, item.Name, item.Description,
		item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
		item.CreatedAt, item.UpdatedAt,
	)
	return err
}

// GetItem возвращает товар по ID (игнорирует удаленные).
func (r *PostgresRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
	query := `
		SELECT id, author_id, type, name, description, price, stock, COALESCE(version,1), attributes, images, created_at, updated_at
		FROM inventory_items WHERE id = $1 AND deleted_at IS NULL
	`
	var item model.Item
	var attrsJSON []byte
	var images []string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
		&item.Price, &item.Stock, &item.Version, &attrsJSON, pq.Array(&images),
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrItemNotFound
		}
		return nil, err
	}

	if len(attrsJSON) > 0 {
		if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}
	item.Images = images
	return &item, nil
}

// UpdateItem обновляет все поля товара (стратегия PUT) with optimistic locking.
func (r *PostgresRepo) UpdateItem(ctx context.Context, item *model.Item) error {
	item.UpdatedAt = time.Now()

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	expected := item.Version
	if expected <= 0 {
		expected = 1
	}
	query := `
		UPDATE inventory_items SET
			author_id = $1, type = $2, name = $3, description = $4,
			price = $5, stock = $6, attributes = $7, images = $8, updated_at = $9,
			version = version + 1
		WHERE id = $10 AND deleted_at IS NULL AND version = $11
	`
	res, err := r.db.ExecContext(ctx, query,
		item.AuthorID, item.Type, item.Name, item.Description,
		item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
		item.UpdatedAt, item.ID, expected,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrVersionConflict
	}
	item.Version = expected + 1
	return nil
}

// DeleteItem выполняет жесткое удаление товара.
func (r *PostgresRepo) DeleteItem(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM inventory_items WHERE id = $1", id)
	return err
}

// SoftDeleteItem выполняет мягкое удаление.
func (r *PostgresRepo) SoftDeleteItem(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE inventory_items SET deleted_at = NOW() WHERE id = $1
	`, id)
	return err
}

// SearchItems выполняет поиск с фильтрами.
func (r *PostgresRepo) SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error) {
	conditions, args, argCount := toPostgresFilter(filters)

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_items %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, author_id, type, name, description, price, stock, COALESCE(version,1), attributes, images, created_at, updated_at, deleted_at
		FROM inventory_items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	paginationArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, paginationArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		var item model.Item
		var attrsJSON []byte
		var images []string
		var deletedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.AuthorID, &item.Type, &item.Name, &item.Description,
			&item.Price, &item.Stock, &item.Version, &attrsJSON, pq.Array(&images),
			&item.CreatedAt, &item.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(attrsJSON) > 0 {
			if err := json.Unmarshal(attrsJSON, &item.Attributes); err != nil {
				return nil, 0, err
			}
		}
		if item.Attributes == nil {
			item.Attributes = map[string]interface{}{}
		}
		item.Images = images
		if deletedAt.Valid {
			t := deletedAt.Time.UTC()
			item.DeletedAt = &t
			item.Attributes["_deleted"] = true
			item.Attributes["_deleted_at"] = t.Format(time.RFC3339)
		}
		items = append(items, &item)
	}

	return items, total, nil
}

// BulkCreateItems — массовая вставка товаров для начальной загрузки или импорта.
// Использует транзакцию для атомарности.
func (r *PostgresRepo) BulkCreateItems(ctx context.Context, items []*model.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO inventory_items (
			id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		attrsJSON, err := json.Marshal(item.Attributes)
		if err != nil {
			return fmt.Errorf("marshal attributes: %w", err)
		}

		_, err = tx.ExecContext(ctx, query,
			item.ID, item.AuthorID, item.Type, item.Name, item.Description,
			item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
			item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ReserveItem — atomic stock decrement with optimistic version bump (single conditional UPDATE).
func (r *PostgresRepo) ReserveItem(ctx context.Context, id string, quantity int) (int, error) {
	var remaining int
	err := r.db.QueryRowContext(ctx, `
		UPDATE inventory_items
		SET stock = stock - $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL AND stock >= $1
		RETURNING stock
	`, quantity, id).Scan(&remaining)
	if err != nil {
		if err == sql.ErrNoRows {
			var exists bool
			_ = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_items WHERE id = $1 AND deleted_at IS NULL)`, id).Scan(&exists)
			if !exists {
				return 0, model.ErrItemNotFound
			}
			return 0, model.ErrNotEnoughStock
		}
		return 0, err
	}
	return remaining, nil
}

// ReleaseItem — atomic stock increment (cancel / refund path).
func (r *PostgresRepo) ReleaseItem(ctx context.Context, id string, quantity int) (int, error) {
	var remaining int
	err := r.db.QueryRowContext(ctx, `
		UPDATE inventory_items
		SET stock = stock + $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING stock
	`, quantity, id).Scan(&remaining)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, model.ErrItemNotFound
		}
		return 0, err
	}
	return remaining, nil
}

// RestoreItem — восстанавливает мягко удаленный товар.
func (r *PostgresRepo) RestoreItem(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE inventory_items SET deleted_at = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrItemNotFound
	}

	return nil
}

// GetStats — возвращает статистику по товарам.
func (r *PostgresRepo) GetStats(ctx context.Context) (*model.Stats, error) {
	stats := &model.Stats{
		ByType:       make(map[string]int64),
		ByAuthor:     make(map[string]int64),
		TopExpensive: []*model.TopItem{},
	}

	// Общее количество + суммарный stock
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(stock), 0)
		FROM inventory_items WHERE deleted_at IS NULL
	`).Scan(&stats.TotalItems, &stats.TotalStock)
	if err != nil {
		return nil, err
	}

	// Группировка по типу
	rows, err := r.db.QueryContext(ctx, `
		SELECT type, COUNT(*) FROM inventory_items
		WHERE deleted_at IS NULL
		GROUP BY type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var typ string
		var count int64
		if err := rows.Scan(&typ, &count); err != nil {
			return nil, err
		}
		stats.ByType[typ] = count
	}

	// Группировка по автору (топ 10)
	rows, err = r.db.QueryContext(ctx, `
		SELECT author_id, COUNT(*) FROM inventory_items
		WHERE deleted_at IS NULL
		GROUP BY author_id
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var authorID string
		var count int64
		if err := rows.Scan(&authorID, &count); err != nil {
			return nil, err
		}
		stats.ByAuthor[authorID] = count
	}

	// Топ-5 самых дорогих SKU
	topRows, err := r.db.QueryContext(ctx, `
		SELECT id, name, price, author_id FROM inventory_items
		WHERE deleted_at IS NULL
		ORDER BY price DESC, name ASC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer topRows.Close()

	for topRows.Next() {
		item := &model.TopItem{}
		if err := topRows.Scan(&item.ID, &item.Name, &item.Price, &item.AuthorID); err != nil {
			return nil, err
		}
		stats.TopExpensive = append(stats.TopExpensive, item)
	}

	return stats, nil
}

// GetByAuthor возвращает все товары автора (с пагинацией по умолчанию 1000).
func (r *PostgresRepo) GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error) {
	return r.SearchItems(ctx, map[string]interface{}{"author_id": authorID}, 1000, 0)
}

// GetByType возвращает все товары указанного типа (с пагинацией по умолчанию 1000).
func (r *PostgresRepo) GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error) {
	return r.SearchItems(ctx, map[string]interface{}{"type": itemType}, 1000, 0)
}

// ---------------------- Outbox методы (только для PostgreSQL) ----------------------

// CreateOutboxEvent создает событие в таблице outbox.
func (r *PostgresRepo) CreateOutboxEvent(ctx context.Context, eventType string, payload []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
	`, eventType, payload)
	return err
}

// CreateItemWithOutbox — сохраняет товар и событие в outbox в одной транзакции.
// Этот метод НЕ входит в интерфейс InventoryRepository, он специфичен для PostgreSQL.
func (r *PostgresRepo) CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	attrsJSON, err := json.Marshal(item.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	query := `
		INSERT INTO inventory_items (
			id, author_id, type, name, description, price, stock, attributes, images, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, query,
		item.ID, item.AuthorID, item.Type, item.Name, item.Description,
		item.Price, item.Stock, attrsJSON, pq.Array(item.Images),
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return err
	}

	var event map[string]interface{}
	if err := json.Unmarshal(eventPayload, &event); err != nil {
		return err
	}
	event["item_id"] = item.ID
	updatedPayload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox (event_type, payload) VALUES ($1, $2)
	`, eventType, updatedPayload)
	if err != nil {
		return err
	}

	return tx.Commit()
}
--- services/inventory/internal/repository/redis_repo.go ---
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	"github.com/redis/go-redis/v9"
)

// RedisCacheRepo — кеш для товаров и результатов поиска.
// Использует TTL для автоматической инвалидации.
type RedisCacheRepo struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacheRepo создает новый Redis кеш.
func NewRedisCacheRepo(addr string, ttl time.Duration) *RedisCacheRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		PoolSize: 10, // размер пула соединений
	})
	return &RedisCacheRepo{client: client, ttl: ttl}
}

// GetItem получает товар из кеша по ID.
func (r *RedisCacheRepo) GetItem(ctx context.Context, id string) (*model.Item, error) {
	key := fmt.Sprintf("inventory:item:%s", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var item model.Item
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// SetItem сохраняет товар в кеше.
func (r *RedisCacheRepo) SetItem(ctx context.Context, item *model.Item) error {
	key := fmt.Sprintf("inventory:item:%s", item.ID)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// DeleteItem удаляет товар из кеша (при обновлении или удалении).
func (r *RedisCacheRepo) DeleteItem(ctx context.Context, id string) error {
	key := fmt.Sprintf("inventory:item:%s", id)
	return r.client.Del(ctx, key).Err()
}

// GetSearchResult получает закешированный результат поиска.
func (r *RedisCacheRepo) GetSearchResult(ctx context.Context, query string, limit, offset int) ([]*model.Item, error) {
	key := fmt.Sprintf("inventory:search:%s:%d:%d", query, limit, offset)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var items []*model.Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// SetSearchResult сохраняет результат поиска в кеше.
func (r *RedisCacheRepo) SetSearchResult(ctx context.Context, query string, limit, offset int, items []*model.Item) error {
	key := fmt.Sprintf("inventory:search:%s:%d:%d", query, limit, offset)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// InvalidateSearchCache удаляет все закешированные результаты поиска (при изменении данных).
func (r *RedisCacheRepo) InvalidateSearchCache(ctx context.Context) error {
	// Используем паттерн для удаления всех ключей с префиксом "inventory:search:*"
	iter := r.client.Scan(ctx, 0, "inventory:search:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Close закрывает соединение с Redis.
func (r *RedisCacheRepo) Close() error {
	return r.client.Close()
}
--- services/inventory/internal/repository/repository.go ---
package repository

import (
    "context"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

type InventoryRepository interface {
    CreateItem(ctx context.Context, item *model.Item) error
    GetItem(ctx context.Context, id string) (*model.Item, error)
    UpdateItem(ctx context.Context, item *model.Item) error
    DeleteItem(ctx context.Context, id string) error
    SearchItems(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*model.Item, int64, error)
    GetByAuthor(ctx context.Context, authorID string) ([]*model.Item, int64, error)
    GetByType(ctx context.Context, itemType string) ([]*model.Item, int64, error)
    BulkCreateItems(ctx context.Context, items []*model.Item) error
    ReserveItem(ctx context.Context, id string, quantity int) (int, error)
    ReleaseItem(ctx context.Context, id string, quantity int) (int, error)
    SoftDeleteItem(ctx context.Context, id string) error
    RestoreItem(ctx context.Context, id string) error
    GetStats(ctx context.Context) (*model.Stats, error)
}===== converter =====
--- services/inventory/internal/converter/item.go ---
package converter

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func ItemFromCreateRequest(req *pb.CreateItemRequest) *model.Item {
	if req == nil {
		return nil
	}
	attrs := map[string]interface{}{}
	if req.Attributes != nil {
		attrs = req.Attributes.AsMap()
	}
	return &model.Item{
		AuthorID:    req.AuthorId,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Attributes:  attrs,
		Images:      req.Images,
	}
}

func ItemFromUpdateRequest(req *pb.UpdateItemRequest) *model.Item {
	if req == nil {
		return nil
	}
	attrs := map[string]interface{}{}
	if req.Attributes != nil {
		attrs = req.Attributes.AsMap()
	}
	return &model.Item{
		ID:          req.Id,
		AuthorID:    req.AuthorId,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Attributes:  attrs,
		Images:      req.Images,
		Version:     int(req.Version),
	}
}

func ItemToProto(item *model.Item) (*pb.Item, error) {
	if item == nil {
		return nil, fmt.Errorf("item is nil")
	}
	var attrs *structpb.Struct
	var err error
	if item.Attributes != nil {
		attrs, err = structpb.NewStruct(item.Attributes)
		if err != nil {
			return nil, err
		}
	}
	return &pb.Item{
		Id:          item.ID,
		AuthorId:    item.AuthorID,
		Type:        item.Type,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Stock:       int32(item.Stock),
		Attributes:  attrs,
		Images:      item.Images,
		CreatedAt:   item.CreatedAt.Format(timeRFC3339),
		UpdatedAt:   item.UpdatedAt.Format(timeRFC3339),
		Version:     int32(item.Version),
	}, nil
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
--- services/inventory/internal/converter/item_test.go ---
package converter

import (
	"testing"
	"time"

	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func TestItemRoundTrip(t *testing.T) {
	req := &pb.CreateItemRequest{AuthorId: "a1", Type: "брелок", Name: "Star", Price: 10.5, Stock: 3}
	item := ItemFromCreateRequest(req)
	if item.AuthorID != "a1" {
		t.Fatal(item.AuthorID)
	}
	item.ID = "id-1"
	item.Version = 4
	item.CreatedAt = time.Unix(0, 0).UTC()
	item.UpdatedAt = item.CreatedAt
	proto, err := ItemToProto(item)
	if err != nil {
		t.Fatal(err)
	}
	if proto.Id != "id-1" || proto.Name != "Star" || proto.Stock != 3 || proto.Version != 4 {
		t.Fatalf("%+v", proto)
	}
}

func TestItemFromUpdateRequest_Version(t *testing.T) {
	item := ItemFromUpdateRequest(&pb.UpdateItemRequest{
		Id: "x", Name: "n", Price: 1, Stock: 2, Version: 7,
	})
	if item.Version != 7 || item.ID != "x" {
		t.Fatalf("%+v", item)
	}
}
===== seed =====
--- services/inventory/internal/seed/*.go ---
cat: 'services/inventory/internal/seed/*.go': Нет такого файла или каталога
===== worker =====
--- services/inventory/internal/worker/outbox_worker.go ---
// services/inventory/internal/worker/outbox_worker.go
package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

type OutboxWorker struct {
	db *sql.DB
	js nats.JetStreamContext
}

func NewOutboxWorker(db *sql.DB, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	// Берём 100 необработанных событий
	rows, err := w.db.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox 
        WHERE processed = false 
        ORDER BY created_at 
        LIMIT 100
    `)
	if err != nil {
		slog.Error("outbox: query", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var eventType string
		var payload []byte

		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			slog.Error("outbox: scan", "err", err)
			continue
		}

		// Публикуем в NATS
		_, err := w.js.Publish(eventType, payload)
		if err != nil {
			slog.Error("outbox: publish", "event_type", eventType, "err", err)
			continue
		}

		// Помечаем как обработанное
		_, err = w.db.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW() 
            WHERE id = $1
        `, id)
		if err != nil {
			slog.Error("outbox: mark processed", "err", err)
		} else {
			slog.Info("outbox: published", "event_type", eventType, "id", id)
		}
	}
}
[denismatveev@c0der event_horizon]$ 

