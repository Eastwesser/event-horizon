[denismatveev@c0der event_horizon]$ find services/shop -type f -name '*.go' | sort
echo "===== main ====="
cat services/shop/cmd/main.go
echo "===== app ====="
for f in services/shop/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== handler ====="
for f in services/shop/internal/handler/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== service ====="
for f in services/shop/internal/service/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== repository ====="
for f in services/shop/internal/repository/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== converter ====="
for f in services/shop/internal/converter/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== worker ====="
for f in services/shop/internal/worker/*.go; do echo "--- $f ---"; cat "$f"; done
services/shop/cmd/main.go
services/shop/internal/app/app.go
services/shop/internal/app/di.go
services/shop/internal/config/config.go
services/shop/internal/config/interfaces.go
services/shop/internal/converter/item.go
services/shop/internal/converter/item_test.go
services/shop/internal/handler/grpc_handler.go
services/shop/internal/interceptor/logger.go
services/shop/internal/interceptor/recovery.go
services/shop/internal/interceptor/validate.go
services/shop/internal/model/errors.go
services/shop/internal/model/item.go
services/shop/internal/model/physical.go
services/shop/internal/model/physical_test.go
services/shop/internal/repository/postgres_integration_test.go
services/shop/internal/repository/postgres_repo.go
services/shop/internal/repository/redis_repo.go
services/shop/internal/service/merch_gate_test.go
services/shop/internal/service/shop_helpers_test.go
services/shop/internal/service/shop_service.go
services/shop/internal/service/shop_service_test.go
services/shop/internal/worker/outbox_worker.go
services/shop/migrations/embed.go
services/shop/proto/shop_grpc.pb.go
services/shop/proto/shop.pb.go
services/shop/proto/validate.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/shop/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init shop app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("shop stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/shop/internal/app/app.go ---
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/services/shop/internal/config"
)

// App is the runnable shop service process (gRPC + metrics HTTP).
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
		a.log.Info("shop gRPC listening", "port", a.cfg.GRPCConfig().Port())
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
--- services/shop/internal/app/di.go ---
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/pkg/migrator"
	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	"github.com/Eastwesser/event-horizon/platform/pkg/tracing"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/config"
	"github.com/Eastwesser/event-horizon/services/shop/internal/handler"
	"github.com/Eastwesser/event-horizon/services/shop/internal/interceptor"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"github.com/Eastwesser/event-horizon/services/shop/internal/service"
	"github.com/Eastwesser/event-horizon/services/shop/internal/worker"
	"github.com/Eastwesser/event-horizon/services/shop/migrations"
	pb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

type diContainer struct {
	cfg config.ConfigProvider

	db        *sql.DB
	redisRepo *repository.RedisShopRepo
	nc        *nats.Conn
	js        nats.JetStreamContext
	pgRepo    *repository.PostgresShopRepo
	billing   billingPb.BillingServiceClient
	svc       service.ShopService
	api       pb.ShopServiceServer
	kafkaProd kafka.Producer
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
		a.initNATS,
		a.initKafka,
		a.initDomain,
		a.initOutbox,
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
		ServiceName: "shop",
		Endpoint:    endpoint,
		Environment: getenv("ENVIRONMENT", "development"),
		Insecure:    true,
	})
	if err != nil {
		a.log.Warn("tracer init failed; continuing without OTLP", "err", err)
		return nil
	}
	a.closer.AddNamed("otel tracer", shutdown)
	return nil
}

func (a *App) initPostgres(_ context.Context) error {
	db, err := sql.Open("postgres", a.cfg.PostgresConfig().DSN())
	if err != nil {
		return fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := migrator.Up(db, migrations.FS); err != nil {
		_ = db.Close()
		return fmt.Errorf("migrations: %w", err)
	}

	a.di.db = db
	a.closer.AddNamed("postgres", func(context.Context) error {
		return db.Close()
	})
	a.log.Info("postgres ready, migrations applied")
	return nil
}

func (a *App) initRedis(_ context.Context) error {
	rc := a.cfg.RedisConfig()
	a.di.redisRepo = repository.NewRedisShopRepo(rc.Addr(), rc.DB())
	a.log.Info("redis repo ready", "addr", rc.Addr(), "db", rc.DB())
	return nil
}

func (a *App) initNATS(_ context.Context) error {
	url := a.cfg.NATSConfig().URL()
	var nc *nats.Conn
	var err error
	for i := 0; i < 30; i++ {
		nc, err = nats.Connect(url)
		if err == nil {
			break
		}
		a.log.Warn("nats connection attempt failed", "attempt", i+1, "of", 30, "err", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("nats connect after 30 attempts: %w", err)
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
	a.log.Info("nats connected", "url", url)
	return nil
}

func (a *App) initKafka(_ context.Context) error {
	cfg := kafka.LoadConfig()
	a.di.kafkaProd = kafka.NewProducer(cfg, kafka.TopicPurchasePaid, a.log)
	a.closer.AddNamed("kafka producer", func(context.Context) error {
		return a.di.kafkaProd.Close()
	})
	a.log.Info("kafka producer ready", "enabled", cfg.Enabled, "topic", kafka.TopicPurchasePaid)
	return nil
}

func (a *App) initDomain(_ context.Context) error {
	a.di.pgRepo = repository.NewPostgresShopRepo(a.di.db)

	billingConn, err := grpc.Dial(a.cfg.BillingConfig().Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial billing: %w", err)
	}
	a.closer.AddNamed("billing grpc", func(context.Context) error { return billingConn.Close() })
	billingClient := billingPb.NewBillingServiceClient(billingConn)
	a.di.billing = billingClient

	var paymentClient paymentPb.PaymentServiceClient
	if addr := a.cfg.PaymentConfig().Addr(); addr != "" {
		pconn, perr := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if perr != nil {
			a.log.Warn("payment dial failed (merch gate degraded)", "err", perr)
		} else {
			a.closer.AddNamed("payment grpc", func(context.Context) error { return pconn.Close() })
			paymentClient = paymentPb.NewPaymentServiceClient(pconn)
		}
	}

	subscribeInventorySync(a.di.js, a.di.pgRepo, a.log)
	subscribePurchaseFulfilled(a.di.js, a.di.pgRepo, a.log)

	a.di.svc = service.New(a.di.pgRepo, a.di.redisRepo, a.di.js, billingClient, paymentClient)
	a.di.svc.SetKafkaProducer(a.di.kafkaProd)
	a.di.api = handler.NewShopHandler(a.di.svc)
	return nil
}

func subscribeInventorySync(js nats.JetStreamContext, pg *repository.PostgresShopRepo, log *slog.Logger) {
	if js == nil || pg == nil {
		return
	}
	log.Info("subscribing to inventory.item.created")
	_, err := js.Subscribe("inventory.item.created", func(msg *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Warn("inventory event parse failed", "err", err)
			_ = msg.Nak()
			return
		}
		itemID, ok := event["item_id"].(string)
		if !ok || itemID == "" {
			log.Warn("inventory event missing item_id", "event", event)
			_ = msg.Nak()
			return
		}
		name, _ := event["name"].(string)
		description, _ := event["description"].(string)
		price, _ := event["price"].(float64)
		stock := 1
		switch v := event["stock"].(type) {
		case float64:
			stock = int(v)
		case int:
			stock = v
		case int32:
			stock = int(v)
		case int64:
			stock = int(v)
		}
		if err := pg.CreateItemFromInventory(context.Background(), itemID, name, description, price, stock); err != nil {
			log.Warn("create shop item from inventory failed", "item_id", itemID, "err", err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	}, nats.Durable("shop-inventory-sync"))
	if err != nil {
		log.Warn("inventory.item.created subscribe failed", "err", err)
		return
	}
	log.Info("subscribed to inventory.item.created")
}

func subscribePurchaseFulfilled(js nats.JetStreamContext, pg *repository.PostgresShopRepo, log *slog.Logger) {
	if js == nil || pg == nil {
		return
	}
	log.Info("subscribing to purchase.fulfilled")
	_, err := js.Subscribe(kafka.TopicPurchaseFulfilled, func(msg *nats.Msg) {
		fulfilled, err := events.UnmarshalPurchaseFulfilled(msg.Data)
		if err != nil {
			log.Warn("purchase.fulfilled parse failed", "err", err)
			_ = msg.Nak()
			return
		}
		if fulfilled.PurchaseUUID == "" {
			log.Warn("purchase.fulfilled missing purchase_uuid")
			_ = msg.Nak()
			return
		}
		if err := pg.MarkPurchaseFulfilled(context.Background(), fulfilled.PurchaseUUID); err != nil {
			log.Warn("mark purchase fulfilled failed", "purchase_id", fulfilled.PurchaseUUID, "err", err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	}, nats.Durable("shop-purchase-fulfilled"))
	if err != nil {
		log.Warn("purchase.fulfilled subscribe failed", "err", err)
		return
	}
	log.Info("subscribed to purchase.fulfilled")
}

func (a *App) initOutbox(_ context.Context) error {
	outboxCtx, cancelOutbox := context.WithCancel(context.Background())
	outboxWorker := worker.NewOutboxWorker(a.di.db, a.di.js, func(ctx context.Context, payload []byte) error {
		if a.di.billing == nil {
			return fmt.Errorf("billing client missing")
		}
		var body struct {
			UserID      string `json:"user_id"`
			Amount      int    `json:"amount"`
			ReferenceID string `json:"reference_id"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return err
		}
		_, err := a.di.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
			UserId:      body.UserID,
			Currency:    billingPb.CurrencyType_TICKETS,
			Amount:      int32(body.Amount),
			Reason:      "shop_purchase_refund",
			ReferenceId: body.ReferenceID,
		})
		return err
	})
	go outboxWorker.Start(outboxCtx)
	a.closer.AddNamed("outbox worker", func(context.Context) error {
		cancelOutbox()
		return nil
	})
	a.log.Info("outbox worker started")
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
			metrics.UnaryServerInterceptor("shop"),
		),
	)
	pb.RegisterShopServiceServer(grpcServer, a.di.api)
	reflection.Register(grpcServer)

	a.grpcServer = grpcServer
	metrics.SetHealthy("shop")
	a.closer.AddNamed("grpc server", func(context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})
	return nil
}

func (a *App) initMetricsHTTP(_ context.Context) error {
	db := a.di.db
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "shop"})
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "reason": "postgres unreachable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "service": "shop"})
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
--- services/shop/internal/handler/grpc_handler.go ---
package handler

import (
    "context"
    "errors"
    "time"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/Eastwesser/event-horizon/services/shop/proto"
    "github.com/Eastwesser/event-horizon/services/shop/internal/model"
    "github.com/Eastwesser/event-horizon/services/shop/internal/service"
)

type ShopHandler struct {
    pb.UnimplementedShopServiceServer
    shopService service.ShopService
}

func NewShopHandler(svc service.ShopService) *ShopHandler {
    return &ShopHandler{shopService: svc}
}

func (h *ShopHandler) GetItems(ctx context.Context, req *pb.GetItemsRequest) (*pb.GetItemsResponse, error) {
    items, err := h.shopService.GetItems(ctx, req.Category, req.GameId, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbItems := make([]*pb.Item, len(items))
    for i, item := range items {
        gameID := ""
        if item.GameID != nil {
            gameID = *item.GameID
        }
        
        // Конвертируем PurchasedAt в строку
        purchasedAt := ""
        if item.PurchasedAt != nil {
            purchasedAt = item.PurchasedAt.Format(time.RFC3339)
        }
        
        pbItems[i] = &pb.Item{
            Id:          item.ID,
            Name:        item.Name,
            Description: item.Description,
            Price:       int32(item.Price),
            Category:    item.Category,
            GameId:      gameID,
            ImageUrl:    item.ImageURL,
            Available:   item.Available,
            Owned:       item.Owned,
            PurchasedAt: purchasedAt,
        }
    }

    return &pb.GetItemsResponse{Items: pbItems}, nil
}

func (h *ShopHandler) PurchaseItem(ctx context.Context, req *pb.PurchaseItemRequest) (*pb.PurchaseItemResponse, error) {
    newBalance, err := h.shopService.PurchaseItem(ctx, req.UserId, req.ItemId)
    if err != nil {
        return nil, mapShopErr(err)
    }

    return &pb.PurchaseItemResponse{
        Success:     true,
        Message:     "Purchase successful",
        NewBalance:  newBalance,
    }, nil
}

func (h *ShopHandler) CancelPurchase(ctx context.Context, req *pb.CancelPurchaseRequest) (*pb.CancelPurchaseResponse, error) {
    result, err := h.shopService.CancelPurchase(ctx, req.UserId, req.ItemId)
    if err != nil {
        return nil, mapShopErr(err)
    }

    msg := "Purchase cancelled"
    if result.AlreadyRefunded {
        msg = "Already refunded"
    }
    return &pb.CancelPurchaseResponse{
        Success:         true,
        Message:         msg,
        NewBalance:      result.NewBalance,
        RefundedAmount:  result.RefundedAmount,
        AlreadyRefunded: result.AlreadyRefunded,
    }, nil
}

func (h *ShopHandler) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
    items, err := h.shopService.GetInventory(ctx, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbItems := make([]*pb.Item, len(items))
    for i, item := range items {
        gameID := ""
        if item.GameID != nil {
            gameID = *item.GameID
        }
        
        // Конвертируем PurchasedAt в строку
        purchasedAt := ""
        if item.PurchasedAt != nil {
            purchasedAt = item.PurchasedAt.Format(time.RFC3339)
        }
        refundableUntil := ""
        if item.RefundableUntil != nil {
            refundableUntil = item.RefundableUntil.UTC().Format(time.RFC3339)
        }
        fulfilledAt := ""
        if item.FulfilledAt != nil {
            fulfilledAt = item.FulfilledAt.UTC().Format(time.RFC3339)
        }

        pbItems[i] = &pb.Item{
            Id:              item.ID,
            Name:            item.Name,
            Description:     item.Description,
            Price:           int32(item.Price),
            Category:        item.Category,
            GameId:          gameID,
            ImageUrl:        item.ImageURL,
            Available:       item.Available,
            Owned:           item.Owned,
            PurchasedAt:     purchasedAt,
            PurchasePrice:   int32(item.PurchasePrice),
            PurchaseId:      item.PurchaseID,
            RefundableUntil: refundableUntil,
            FulfilledAt:     fulfilledAt,
            CanCancel:       item.CanCancel,
        }
    }

    return &pb.GetInventoryResponse{Items: pbItems}, nil
}

func (h *ShopHandler) ListPurchasesByItemIDs(ctx context.Context, req *pb.ListPurchasesByItemIDsRequest) (*pb.ListPurchasesByItemIDsResponse, error) {
    rows, total, salesCount, tickets, err := h.shopService.ListPurchasesByItemIDs(ctx, req.GetItemIds(), int(req.GetLimit()), int(req.GetOffset()))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    out := make([]*pb.PurchaseRow, 0, len(rows))
    for _, p := range rows {
        row := &pb.PurchaseRow{
            Id:          p.ID,
            UserId:      p.UserID,
            ItemId:      p.ItemID,
            Price:       int32(p.Price),
            Status:      p.Status,
            PurchasedAt: p.PurchasedAt.Format(time.RFC3339),
        }
        if p.RefundedAt != nil {
            row.RefundedAt = p.RefundedAt.Format(time.RFC3339)
        }
        out = append(out, row)
    }
    return &pb.ListPurchasesByItemIDsResponse{
        Purchases:     out,
        Total:         total,
        SalesCount:    salesCount,
        TicketsEarned: tickets,
    }, nil
}

func mapShopErr(err error) error {
    switch {
    case errors.Is(err, model.ErrSubscriptionRequired):
        return status.Error(codes.PermissionDenied, "subscription_required")
    case errors.Is(err, model.ErrItemNotFound), errors.Is(err, model.ErrPurchaseNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, model.ErrAlreadyOwned):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, model.ErrRefundWindowExpired), errors.Is(err, model.ErrAlreadyFulfilled):
        // InvalidArgument → HTTP 400 (FailedPrecondition is mapped to 409 for funds/stock).
        return status.Error(codes.InvalidArgument, err.Error())
    case errors.Is(err, model.ErrInsufficientFunds), errors.Is(err, model.ErrItemUnavailable):
        return status.Error(codes.FailedPrecondition, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}
===== service =====
--- services/shop/internal/service/merch_gate_test.go ---
package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"google.golang.org/grpc"
)

type stubPayment struct {
	allowed bool
	reason  string
	err     error
}

func (s stubPayment) CreateCheckout(context.Context, *paymentPb.CreateCheckoutRequest, ...grpc.CallOption) (*paymentPb.CreateCheckoutResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) ConfirmPayment(context.Context, *paymentPb.ConfirmPaymentRequest, ...grpc.CallOption) (*paymentPb.ConfirmPaymentResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) GetSubscription(context.Context, *paymentPb.GetSubscriptionRequest, ...grpc.CallOption) (*paymentPb.GetSubscriptionResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) CanPurchaseMerch(context.Context, *paymentPb.CanPurchaseMerchRequest, ...grpc.CallOption) (*paymentPb.CanPurchaseMerchResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &paymentPb.CanPurchaseMerchResponse{Allowed: s.allowed, Reason: s.reason}, nil
}

func TestCheckMerchAllowed_NilClient(t *testing.T) {
	s := &shopService{}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "payment service") {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_Blocked(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: false, reason: "no subscription"}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if !errors.Is(err, model.ErrSubscriptionRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_Allowed(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: true}}
	if err := s.checkMerchAllowed(context.Background(), "u1"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckMerchAllowed_RPCError(t *testing.T) {
	s := &shopService{payment: stubPayment{err: errors.New("down")}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "subscription check failed") {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_DefaultReason(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: false, reason: ""}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if !errors.Is(err, model.ErrSubscriptionRequired) {
		t.Fatalf("got %v", err)
	}
}
--- services/shop/internal/service/shop_helpers_test.go ---
package service

import (
	"regexp"
	"testing"
)

func TestNewEventID_Format(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{32}$`)
	for i := 0; i < 5; i++ {
		id := newEventID()
		if !re.MatchString(id) {
			t.Fatalf("bad id %q", id)
		}
	}
}
--- services/shop/internal/service/shop_service.go ---
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
)

// CancelResult is returned by CancelPurchase.
type CancelResult struct {
	NewBalance      int32
	RefundedAmount  int32
	AlreadyRefunded bool
}

// ShopService is the interface for the shop service.
type ShopService interface {
	GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error)
	PurchaseItem(ctx context.Context, userID, itemID string) (int32, error)
	CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error)
	GetInventory(ctx context.Context, userID string) ([]repository.Item, error)
	ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error)
	SetKafkaProducer(p kafka.Producer)
}

// ShopStore is the PostgreSQL port used by shopService.
type ShopStore interface {
	GetItems(ctx context.Context, category, gameID string) ([]repository.Item, error)
	GetItemByID(ctx context.Context, itemID string) (*repository.Item, error)
	IsItemOwned(ctx context.Context, userID, itemID string) (bool, error)
	ListOwnedItemIDs(ctx context.Context, userID string) ([]string, error)
	PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []repository.OutboxRecord) error
	NextSpendCycle(ctx context.Context, userID, itemID string) (int, error)
	BumpSpendCycle(ctx context.Context, userID, itemID string) error
	SavePurchaseSaga(ctx context.Context, saga repository.PurchaseSaga) (string, string, error)
	MarkPurchaseSaga(ctx context.Context, id, state string) error
	EnqueueOutbox(ctx context.Context, eventType string, payload []byte) (string, error)
	MarkOutboxProcessed(ctx context.Context, id string) error
	GetUserInventory(ctx context.Context, userID string) ([]repository.Item, error)
	RefundPurchase(ctx context.Context, userID, itemID string) (*repository.RefundResult, error)
	MarkPurchaseFulfilled(ctx context.Context, purchaseID string) error
	ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error)
	CreateItemFromInventory(ctx context.Context, itemID, name, description string, price float64, stock int) error
}

// ShopCache is the Redis port for catalog lists / invalidation.
type ShopCache interface {
	GetItems(ctx context.Context, key string) ([]repository.Item, error)
	SetItems(ctx context.Context, key string, items []repository.Item, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Note: kafka is replaced with NATS in our project
type shopService struct {
	pgRepo    ShopStore
	redisRepo ShopCache
	js        nats.JetStreamContext
	billing   billingPb.BillingServiceClient
	payment   paymentPb.PaymentServiceClient
	kafkaProd kafka.Producer // optional (noop if unset / KAFKA_BROKERS empty)
}

// New constructs the shop service. gRPC dial / NATS inventory sync live in app wiring.
func New(
	pg ShopStore,
	redis ShopCache,
	js nats.JetStreamContext,
	billing billingPb.BillingServiceClient,
	payment paymentPb.PaymentServiceClient,
) ShopService {
	return &shopService{
		pgRepo:    pg,
		redisRepo: redis,
		js:        js,
		billing:   billing,
		payment:   payment,
	}
}

func (s *shopService) GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error) {
	cacheKey := fmt.Sprintf("shop:items:%s:%s", category, gameID)

	// Пытаемся получить из Redis
	items, err := s.redisRepo.GetItems(ctx, cacheKey)
	if err == nil {
		s.applyOwnedFlags(ctx, userID, items)
		return items, nil
	}

	// Если нет в кеше — из PostgreSQL
	items, err = s.pgRepo.GetItems(ctx, category, gameID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в Redis (TTL 5 минут)
	_ = s.redisRepo.SetItems(ctx, cacheKey, items, 5*time.Minute)

	s.applyOwnedFlags(ctx, userID, items)
	return items, nil
}

// applyOwnedFlags sets Owned from one inventory query (avoids N+1 IsItemOwned).
func (s *shopService) applyOwnedFlags(ctx context.Context, userID string, items []repository.Item) {
	if len(items) == 0 || userID == "" {
		return
	}
	ids, err := s.pgRepo.ListOwnedItemIDs(ctx, userID)
	if err != nil {
		return
	}
	owned := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		owned[id] = struct{}{}
	}
	for i := range items {
		_, items[i].Owned = owned[items[i].ID]
	}
}

// PurchaseItem spends tickets in billing, then writes stock and the outbox in the shop database.
// The returned int32 is the ticket balance after a successful spend.
func (s *shopService) PurchaseItem(ctx context.Context, userID, itemID string) (int32, error) {
	item, err := s.pgRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return 0, model.ErrItemNotFound
	}
	if !item.Available {
		return 0, model.ErrItemUnavailable
	}

	if item.Category == "merch" {
		if err := s.checkMerchAllowed(ctx, userID); err != nil {
			return 0, err
		}
	}

	owned, err := s.pgRepo.IsItemOwned(ctx, userID, itemID)
	if err != nil {
		return 0, err
	}
	if owned {
		return 0, model.ErrAlreadyOwned
	}

	cycle, err := s.pgRepo.NextSpendCycle(ctx, userID, itemID)
	if err != nil {
		return 0, fmt.Errorf("spend cycle: %w", err)
	}
	spendRef := fmt.Sprintf("shop:%s:%s:%d", userID, itemID, cycle)
	refundRef := fmt.Sprintf("refund:shop:%s:%s:%d", userID, itemID, cycle)
	sagaID, purchaseID, err := s.pgRepo.SavePurchaseSaga(ctx, repository.PurchaseSaga{
		ID: uuid.NewString(), UserID: userID, ItemID: itemID, PurchaseID: uuid.NewString(),
		SpendRef: spendRef, RefundRef: refundRef, Price: item.Price, State: "started",
	})
	if err != nil {
		return 0, fmt.Errorf("purchase saga: %w", err)
	}

	spendResp, err := s.billing.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(item.Price),
		Reason:      "shop_purchase",
		ReferenceId: spendRef,
		CheckOnly:   false,
	})
	alreadySpent := false
	if err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, model.ErrInsufficientFunds
		}
		if !isDuplicateRef(err) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, fmt.Errorf("failed to spend tickets: %w", err)
		}
		alreadySpent = true
	} else if spendResp == nil || !spendResp.GetSuccess() {
		msg := "spend rejected"
		if spendResp != nil && spendResp.GetMessage() != "" {
			msg = spendResp.GetMessage()
		}
		if strings.Contains(msg, "insufficient balance") {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, model.ErrInsufficientFunds
		}
		if !isDuplicateRef(errors.New(msg)) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, fmt.Errorf("failed to spend tickets: %s", msg)
		}
		alreadySpent = true
	}
	_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "spent")

	event := map[string]interface{}{
		"user_id":     userID,
		"item_id":     itemID,
		"item_name":   item.Name,
		"price":       item.Price,
		"category":    item.Category,
		"purchase_id": purchaseID,
		"timestamp":   time.Now().Unix(),
	}
	eventData, _ := json.Marshal(event)
	paid := events.PurchasePaid{
		EventUUID:    uuid.NewString(),
		PurchaseUUID: purchaseID,
		UserUUID:     userID,
		ItemUUID:     itemID,
		Price:        int32(item.Price),
	}
	paidBody, mErr := paid.Marshal()
	if mErr != nil {
		return 0, fmt.Errorf("marshal purchase.paid: %w", mErr)
	}

	if err := s.pgRepo.PurchaseItemWithStock(ctx, userID, itemID, purchaseID, item.Price, []repository.OutboxRecord{
		{EventType: "shop.purchased", Payload: eventData},
		{EventType: kafka.TopicPurchasePaid, Payload: paidBody},
	}); err != nil {
		if isDuplicateRef(err) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
			return 0, model.ErrAlreadyOwned
		}
		_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensating")
		log.Printf("CRITICAL: tickets spent but purchase failed for user %s, item %s: %v — refunding", userID, itemID, err)
		outboxID := s.enqueueRefund(ctx, userID, itemID, item.Price, refundRef, err)
		_ = s.pgRepo.BumpSpendCycle(ctx, userID, itemID)
		if _, refundErr := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
			UserId:      userID,
			Currency:    billingPb.CurrencyType_TICKETS,
			Amount:      int32(item.Price),
			Reason:      "shop_purchase_refund",
			ReferenceId: refundRef,
		}); refundErr != nil && !isDuplicateRef(refundErr) {
			metrics.PurchaseRefundFailures.Inc()
			log.Printf("CRITICAL: refund also failed for user %s, item %s: %v", userID, itemID, refundErr)
			return 0, fmt.Errorf("failed to record purchase: %w", err)
		}
		if outboxID != "" {
			_ = s.pgRepo.MarkOutboxProcessed(ctx, outboxID)
		}
		_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensated")
		return 0, fmt.Errorf("failed to record purchase: %w", err)
	}
	_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
	if alreadySpent {
		log.Printf("shop: replayed spend %s for user %s item %s", spendRef, userID, itemID)
	}

	gameID := ""
	if item.GameID != nil {
		gameID = *item.GameID
	}
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
	_ = s.redisRepo.Delete(ctx, "shop:items:all:")
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))

	metrics.RecordOrder(float64(item.Price))

	if spendResp != nil {
		return spendResp.NewBalance, nil
	}
	bal, balErr := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
		UserId:   userID,
		Currency: billingPb.CurrencyType_TICKETS,
	})
	if balErr != nil || bal == nil {
		return 0, nil
	}
	return bal.GetBalance(), nil
}

// enqueueRefund stores a billing.refund outbox row so a crashed process can still return the tickets.
// An empty id means the insert failed and only the in-process refund remains.
func (s *shopService) enqueueRefund(ctx context.Context, userID, itemID string, price int, referenceID string, cause error) string {
	body, _ := json.Marshal(map[string]any{
		"user_id":      userID,
		"item_id":      itemID,
		"amount":       price,
		"reference_id": referenceID,
		"error":        cause.Error(),
	})
	id, err := s.pgRepo.EnqueueOutbox(ctx, "billing.refund", body)
	if err != nil {
		log.Printf("shop: enqueue refund %s: %v", referenceID, err)
		return ""
	}
	return id
}

// creditCancel pays tickets back for a cancelled purchase.
// reference_id is cancel-{purchaseID}, so a repeated cancel does not credit twice.
func (s *shopService) creditCancel(ctx context.Context, userID string, refund *repository.RefundResult) (int32, error) {
	if refund == nil || refund.PurchaseID == "" || refund.Price <= 0 {
		bal, err := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
			UserId:   userID,
			Currency: billingPb.CurrencyType_TICKETS,
		})
		if err != nil || bal == nil {
			return 0, err
		}
		return bal.GetBalance(), nil
	}
	refID := fmt.Sprintf("cancel-%s", refund.PurchaseID)
	addResp, err := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(refund.Price),
		Reason:      "shop_purchase_cancel",
		ReferenceId: refID,
	})
	if err != nil {
		if isDuplicateRef(err) {
			bal, balErr := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
				UserId:   userID,
				Currency: billingPb.CurrencyType_TICKETS,
			})
			if balErr != nil || bal == nil {
				return 0, nil
			}
			return bal.GetBalance(), nil
		}
		return 0, err
	}
	if addResp == nil {
		return 0, nil
	}
	return addResp.GetNewBalance(), nil
}

// isDuplicateRef is true when Postgres or billing already stored this reference_id (SQLSTATE 23505).
func isDuplicateRef(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "23505") || strings.Contains(s, "unique constraint")
}

func (s *shopService) GetInventory(ctx context.Context, userID string) ([]repository.Item, error) {
	items, err := s.pgRepo.GetUserInventory(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range items {
		items[i].CanCancel = canCancelInventoryItem(items[i], now)
	}
	return items, nil
}

// canCancelInventoryItem is the player-facing refund window.
// Physical goods close once fulfillment has set fulfilled_at.
func canCancelInventoryItem(item repository.Item, now time.Time) bool {
	if item.PurchaseID == "" || item.RefundableUntil == nil {
		return false
	}
	if now.After(item.RefundableUntil.UTC()) {
		return false
	}
	if item.FulfilledAt != nil && model.IsPhysicalItem(item.Category) {
		return false
	}
	return true
}

func (s *shopService) CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error) {
	item, err := s.pgRepo.GetItemByID(ctx, itemID)
	if err != nil {
		// Still allow refund if purchase exists but shop item row is gone.
		item = &repository.Item{ID: itemID, Category: "merch"}
	}
	_ = item

	refund, err := s.pgRepo.RefundPurchase(ctx, userID, itemID)
	if err != nil {
		if errors.Is(err, model.ErrRefundWindowExpired) || errors.Is(err, model.ErrAlreadyFulfilled) {
			return nil, err
		}
		if strings.Contains(err.Error(), "purchase not found") {
			return nil, model.ErrPurchaseNotFound
		}
		return nil, err
	}
	if !refund.AlreadyRefunded {
		if bumpErr := s.pgRepo.BumpSpendCycle(ctx, userID, itemID); bumpErr != nil {
			log.Printf("shop: bump spend cycle user=%s item=%s: %v", userID, itemID, bumpErr)
		}
	}

	newBal, err := s.creditCancel(ctx, userID, refund)
	if err != nil {
		log.Printf("CRITICAL: purchase refunded in shop but AddCurrency failed user=%s item=%s purchase=%s: %v",
			userID, itemID, refund.PurchaseID, err)
		return nil, fmt.Errorf("failed to refund tickets: %w", err)
	}
	if refund.AlreadyRefunded {
		return &CancelResult{
			NewBalance:      newBal,
			AlreadyRefunded: true,
		}, nil
	}

	gameID := ""
	if item.GameID != nil {
		gameID = *item.GameID
	}
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
	_ = s.redisRepo.Delete(ctx, "shop:items:all:")
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))

	return &CancelResult{
		NewBalance:     newBal,
		RefundedAmount: int32(refund.Price),
	}, nil
}

// checkMerchAllowed asks payment whether this player may buy a physical item.
// Without an active subscription the purchase stops before any tickets are spent.
func (s *shopService) checkMerchAllowed(ctx context.Context, userID string) error {
	if s.payment == nil {
		return fmt.Errorf("merch purchases require payment service")
	}
	gate, err := s.payment.CanPurchaseMerch(ctx, &paymentPb.CanPurchaseMerchRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("subscription check failed: %w", err)
	}
	if !gate.GetAllowed() {
		return model.ErrSubscriptionRequired
	}
	return nil
}

func (s *shopService) ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error) {
	return s.pgRepo.ListPurchasesByItemIDs(ctx, itemIDs, limit, offset)
}

// SetKafkaProducer attaches an optional Kafka producer for PurchasePaid events (Week 5).
func (s *shopService) SetKafkaProducer(p kafka.Producer) {
	s.kafkaProd = p
}

func newEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
--- services/shop/internal/service/shop_service_test.go ---
package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"google.golang.org/grpc"
)

type mockShopStore struct {
	item          *repository.Item
	getItemErr    error
	owned         bool
	ownedErr      error
	ownedIDs      []string
	listOwnedErr  error
	purchaseErr   error
	purchaseCalls int
	refund        *repository.RefundResult
	refundErr     error
	refundCalls   int
	listRows      []repository.PurchaseRecord
	listTotal     int64
	listSales     int64
	listTickets   int64
	listErr       error
	inventory     []repository.Item
	items         []repository.Item
	getItemsErr   error
}

func (m *mockShopStore) GetItems(context.Context, string, string) ([]repository.Item, error) {
	if m.getItemsErr != nil {
		return nil, m.getItemsErr
	}
	return m.items, nil
}
func (m *mockShopStore) GetItemByID(context.Context, string) (*repository.Item, error) {
	if m.getItemErr != nil {
		return nil, m.getItemErr
	}
	if m.item == nil {
		return nil, errors.New("not found")
	}
	cp := *m.item
	return &cp, nil
}
func (m *mockShopStore) IsItemOwned(context.Context, string, string) (bool, error) {
	return m.owned, m.ownedErr
}
func (m *mockShopStore) ListOwnedItemIDs(context.Context, string) ([]string, error) {
	if m.listOwnedErr != nil {
		return nil, m.listOwnedErr
	}
	if m.ownedIDs != nil {
		return m.ownedIDs, nil
	}
	return nil, nil
}
func (m *mockShopStore) PurchaseItemWithStock(_ context.Context, _, _, _ string, _ int, _ []repository.OutboxRecord) error {
	m.purchaseCalls++
	return m.purchaseErr
}
func (m *mockShopStore) NextSpendCycle(context.Context, string, string) (int, error) { return 1, nil }
func (m *mockShopStore) BumpSpendCycle(context.Context, string, string) error        { return nil }
func (m *mockShopStore) SavePurchaseSaga(_ context.Context, s repository.PurchaseSaga) (string, string, error) {
	return s.ID, s.PurchaseID, nil
}
func (m *mockShopStore) MarkPurchaseSaga(context.Context, string, string) error { return nil }
func (m *mockShopStore) EnqueueOutbox(context.Context, string, []byte) (string, error) {
	return "ob1", nil
}
func (m *mockShopStore) MarkOutboxProcessed(context.Context, string) error { return nil }
func (m *mockShopStore) GetUserInventory(context.Context, string) ([]repository.Item, error) {
	return m.inventory, nil
}
func (m *mockShopStore) RefundPurchase(context.Context, string, string) (*repository.RefundResult, error) {
	m.refundCalls++
	if m.refundErr != nil {
		return nil, m.refundErr
	}
	if m.refund == nil {
		return nil, errors.New("purchase not found")
	}
	cp := *m.refund
	return &cp, nil
}
func (m *mockShopStore) MarkPurchaseFulfilled(context.Context, string) error {
	return nil
}
func (m *mockShopStore) ListPurchasesByItemIDs(context.Context, []string, int, int) ([]repository.PurchaseRecord, int64, int64, int64, error) {
	return m.listRows, m.listTotal, m.listSales, m.listTickets, m.listErr
}
func (m *mockShopStore) CreateItemFromInventory(context.Context, string, string, string, float64, int) error {
	return nil
}

type mockShopCache struct {
	deleted []string
	items   []repository.Item
	getErr  error
}

func (m *mockShopCache) GetItems(context.Context, string) ([]repository.Item, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.items, nil
}
func (m *mockShopCache) SetItems(context.Context, string, []repository.Item, time.Duration) error {
	return nil
}
func (m *mockShopCache) Delete(_ context.Context, key string) error {
	m.deleted = append(m.deleted, key)
	return nil
}

type stubBilling struct {
	spendBal     int32
	spendOK      bool
	spendMsg     string
	spendErr     error
	spendCalls   int
	addBal       int32
	addErr       error
	addCalls     int
	getBal       int32
	getBalErr    error
	lastSpendRef string
}

func (b *stubBilling) GetBalance(context.Context, *billingPb.GetBalanceRequest, ...grpc.CallOption) (*billingPb.GetBalanceResponse, error) {
	if b.getBalErr != nil {
		return nil, b.getBalErr
	}
	return &billingPb.GetBalanceResponse{Balance: b.getBal}, nil
}
func (b *stubBilling) GetAllBalances(context.Context, *billingPb.GetAllBalancesRequest, ...grpc.CallOption) (*billingPb.GetAllBalancesResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) AddCurrency(context.Context, *billingPb.AddCurrencyRequest, ...grpc.CallOption) (*billingPb.AddCurrencyResponse, error) {
	b.addCalls++
	if b.addErr != nil {
		return nil, b.addErr
	}
	return &billingPb.AddCurrencyResponse{NewBalance: b.addBal, Success: true}, nil
}
func (b *stubBilling) SpendCurrency(_ context.Context, req *billingPb.SpendCurrencyRequest, _ ...grpc.CallOption) (*billingPb.SpendCurrencyResponse, error) {
	b.spendCalls++
	b.lastSpendRef = req.GetReferenceId()
	if b.spendErr != nil {
		return nil, b.spendErr
	}
	return &billingPb.SpendCurrencyResponse{
		Success:    b.spendOK,
		Message:    b.spendMsg,
		NewBalance: b.spendBal,
	}, nil
}
func (b *stubBilling) GetTransactionHistory(context.Context, *billingPb.GetTransactionHistoryRequest, ...grpc.CallOption) (*billingPb.GetTransactionHistoryResponse, error) {
	return nil, errors.New("unused")
}

func newTestShop(store *mockShopStore, cache *mockShopCache, billing *stubBilling) *shopService {
	if cache == nil {
		cache = &mockShopCache{getErr: errors.New("miss")}
	}
	return &shopService{
		pgRepo:    store,
		redisRepo: cache,
		billing:   billing,
	}
}

func withPayment(s *shopService, p stubPayment) *shopService {
	s.payment = p
	return s
}

func TestNew_ConstructsService(t *testing.T) {
	svc := New(&mockShopStore{}, &mockShopCache{}, nil, &stubBilling{}, nil)
	if svc == nil {
		t.Fatal("nil service")
	}
}

func TestPurchaseItem_NotFoundUnavailableOwned(t *testing.T) {
	billing := &stubBilling{spendOK: true, spendBal: 90}
	store := &mockShopStore{getItemErr: errors.New("missing")}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrItemNotFound) {
		t.Fatalf("not found: %v", err)
	}

	store = &mockShopStore{item: &repository.Item{ID: "i1", Available: false, Price: 10, Category: "card"}}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrItemUnavailable) {
		t.Fatalf("unavailable: %v", err)
	}

	store = &mockShopStore{item: &repository.Item{ID: "i1", Available: true, Price: 10, Category: "card"}, owned: true}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrAlreadyOwned) {
		t.Fatalf("owned: %v", err)
	}
	if billing.spendCalls != 0 {
		t.Fatalf("should not spend when owned")
	}
}

func TestPurchaseItem_MerchBlocked(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "m1", Available: true, Price: 50, Category: "merch"}}
	billing := &stubBilling{spendOK: true}
	svc := withPayment(newTestShop(store, nil, billing), stubPayment{allowed: false})
	_, err := svc.PurchaseItem(context.Background(), "u1", "m1")
	if !errors.Is(err, model.ErrSubscriptionRequired) {
		t.Fatalf("%v", err)
	}
	if billing.spendCalls != 0 || store.purchaseCalls != 0 {
		t.Fatalf("merch gate should short-circuit")
	}
}

func TestPurchaseItem_InsufficientFunds(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "i1", Available: true, Price: 100, Category: "card", Name: "Card"}}
	billing := &stubBilling{spendErr: errors.New("insufficient balance")}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrInsufficientFunds) {
		t.Fatalf("%v", err)
	}

	billing = &stubBilling{spendOK: false, spendMsg: "insufficient balance"}
	_, err = newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrInsufficientFunds) {
		t.Fatalf("msg path: %v", err)
	}
}

func TestPurchaseItem_Success(t *testing.T) {
	store := &mockShopStore{item: &repository.Item{ID: "item-uuid-1", Available: true, Price: 25, Category: "card", Name: "Keychain"}}
	cache := &mockShopCache{getErr: errors.New("miss")}
	billing := &stubBilling{spendOK: true, spendBal: 75}
	bal, err := newTestShop(store, cache, billing).PurchaseItem(context.Background(), "u1", "item-uuid-1")
	if err != nil || bal != 75 {
		t.Fatalf("bal=%d err=%v", bal, err)
	}
	if billing.spendCalls != 1 || store.purchaseCalls != 1 {
		t.Fatalf("spend=%d purchase=%d", billing.spendCalls, store.purchaseCalls)
	}
	if billing.lastSpendRef != "shop:u1:item-uuid-1:1" {
		t.Fatalf("spend ref %q", billing.lastSpendRef)
	}
	if len(cache.deleted) == 0 {
		t.Fatal("expected cache invalidation")
	}
}

func TestPurchaseItem_RecordFailureRefunds(t *testing.T) {
	store := &mockShopStore{
		item:        &repository.Item{ID: "i1", Available: true, Price: 10, Category: "card", Name: "x"},
		purchaseErr: errors.New("db down"),
	}
	billing := &stubBilling{spendOK: true, spendBal: 0, addBal: 10}
	_, err := newTestShop(store, nil, billing).PurchaseItem(context.Background(), "u1", "i1")
	if err == nil || !strings.Contains(err.Error(), "failed to record purchase") {
		t.Fatalf("%v", err)
	}
	if billing.addCalls != 1 {
		t.Fatalf("expected refund AddCurrency, got %d", billing.addCalls)
	}
}

func TestCancelPurchase_NotFound(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "card"},
		refundErr: errors.New("purchase not found"),
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrPurchaseNotFound) {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_AlreadyRefundedIdempotent(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{AlreadyRefunded: true, PurchaseID: "p1", Price: 25},
	}
	billing := &stubBilling{addBal: 42}
	got, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err != nil || !got.AlreadyRefunded || got.NewBalance != 42 || got.RefundedAmount != 0 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if billing.addCalls != 1 {
		t.Fatal("retry must send the same cancel reference")
	}
}

func TestCancelPurchase_Success(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{PurchaseID: "p1", Price: 25},
	}
	cache := &mockShopCache{}
	billing := &stubBilling{addBal: 100}
	got, err := newTestShop(store, cache, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err != nil || got.AlreadyRefunded || got.RefundedAmount != 25 || got.NewBalance != 100 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if billing.addCalls != 1 || store.refundCalls != 1 || len(cache.deleted) == 0 {
		t.Fatalf("add=%d refund=%d deleted=%v", billing.addCalls, store.refundCalls, cache.deleted)
	}
}

func TestCancelPurchase_MissingCatalogItemStillRefunds(t *testing.T) {
	store := &mockShopStore{
		getItemErr: errors.New("gone"),
		refund:     &repository.RefundResult{PurchaseID: "p1", Price: 7},
	}
	billing := &stubBilling{addBal: 7}
	got, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "orphan")
	if err != nil || got.RefundedAmount != 7 {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestListPurchasesByItemIDs_Delegates(t *testing.T) {
	store := &mockShopStore{
		listRows:    []repository.PurchaseRecord{{ID: "p1", ItemID: "i1", Price: 10}},
		listTotal:   1,
		listSales:   1,
		listTickets: 10,
	}
	rows, total, sales, tickets, err := newTestShop(store, nil, &stubBilling{}).
		ListPurchasesByItemIDs(context.Background(), []string{"i1"}, 20, 0)
	if err != nil || total != 1 || sales != 1 || tickets != 10 || len(rows) != 1 {
		t.Fatalf("rows=%v total=%d sales=%d tickets=%d err=%v", rows, total, sales, tickets, err)
	}
}

func TestGetInventory_Delegates(t *testing.T) {
	store := &mockShopStore{inventory: []repository.Item{{ID: "i1", Name: "n"}}}
	got, err := newTestShop(store, nil, &stubBilling{}).GetInventory(context.Background(), "u1")
	if err != nil || len(got) != 1 || got[0].ID != "i1" {
		t.Fatalf("%v err=%v", got, err)
	}
}

func TestGetItems_CacheHitSetsOwned(t *testing.T) {
	store := &mockShopStore{ownedIDs: []string{"i1"}}
	cache := &mockShopCache{items: []repository.Item{{ID: "i1", Name: "cached"}}}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err != nil || len(got) != 1 || !got[0].Owned {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestGetItems_CacheMissLoadsPostgres(t *testing.T) {
	store := &mockShopStore{
		items:    []repository.Item{{ID: "i2", Name: "db"}},
		ownedIDs: []string{},
	}
	cache := &mockShopCache{getErr: errors.New("miss")}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "card", "hexagon", "u1")
	if err != nil || len(got) != 1 || got[0].Name != "db" || got[0].Owned {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestGetItems_BatchOwnedMixed(t *testing.T) {
	store := &mockShopStore{
		ownedIDs: []string{"i2"},
	}
	cache := &mockShopCache{items: []repository.Item{
		{ID: "i1", Name: "a"},
		{ID: "i2", Name: "b"},
		{ID: "i3", Name: "c"},
	}}
	got, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err != nil || len(got) != 3 {
		t.Fatalf("%+v err=%v", got, err)
	}
	if got[0].Owned || !got[1].Owned || got[2].Owned {
		t.Fatalf("owned flags: %+v %+v %+v", got[0], got[1], got[2])
	}
}

func TestGetItems_PostgresError(t *testing.T) {
	store := &mockShopStore{getItemsErr: errors.New("db")}
	cache := &mockShopCache{getErr: errors.New("miss")}
	_, err := newTestShop(store, cache, &stubBilling{}).GetItems(context.Background(), "all", "", "u1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelPurchase_AddCurrencyFails(t *testing.T) {
	store := &mockShopStore{
		item:   &repository.Item{ID: "i1", Category: "card"},
		refund: &repository.RefundResult{PurchaseID: "p1", Price: 25},
	}
	billing := &stubBilling{addErr: errors.New("billing down")}
	_, err := newTestShop(store, nil, billing).CancelPurchase(context.Background(), "u1", "i1")
	if err == nil || !strings.Contains(err.Error(), "failed to refund tickets") {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_RefundWindowExpired(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "game_skin"},
		refundErr: model.ErrRefundWindowExpired,
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrRefundWindowExpired) {
		t.Fatalf("%v", err)
	}
}

func TestCancelPurchase_AlreadyFulfilledPhysical(t *testing.T) {
	store := &mockShopStore{
		item:      &repository.Item{ID: "i1", Category: "merch"},
		refundErr: model.ErrAlreadyFulfilled,
	}
	_, err := newTestShop(store, nil, &stubBilling{}).CancelPurchase(context.Background(), "u1", "i1")
	if !errors.Is(err, model.ErrAlreadyFulfilled) {
		t.Fatalf("%v", err)
	}
}

func TestGetInventory_ComputesCanCancel(t *testing.T) {
	until := time.Now().UTC().Add(48 * time.Hour)
	store := &mockShopStore{
		inventory: []repository.Item{{
			ID:              "i1",
			Category:        "game_skin",
			PurchaseID:      "p1",
			RefundableUntil: &until,
		}},
	}
	got, err := newTestShop(store, nil, &stubBilling{}).GetInventory(context.Background(), "u1")
	if err != nil || len(got) != 1 || !got[0].CanCancel {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestSetKafkaProducer(t *testing.T) {
	s := &shopService{}
	s.SetKafkaProducer(nil)
	if s.kafkaProd != nil {
		t.Fatal("expected nil producer")
	}
}
===== repository =====
--- services/shop/internal/repository/postgres_integration_test.go ---
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
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"github.com/Eastwesser/event-horizon/services/shop/migrations"
)

func shopDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	if dsn := os.Getenv("SHOP_TEST_DATABASE_URL"); dsn != "" {
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
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "shop_test"},
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
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/shop_test?sslmode=disable", host, port.Port())
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

func TestShop_PurchaseItemWithStock_AndVersionBump(t *testing.T) {
	db, cleanup := shopDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresShopRepo(db)

	itemID := uuid.NewString()
	userID := uuid.NewString()
	_, err := db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, price, category, available, stock, version)
		VALUES ($1, 'merch hat', 't', 50, 'merch', true, 2, 1)`, itemID)
	if err != nil {
		t.Fatalf("seed item: %v", err)
	}

	purchaseID := uuid.NewString()
	if err := repo.PurchaseItemWithStock(ctx, userID, itemID, purchaseID, 50, []repository.OutboxRecord{{
		EventType: "shop.purchased",
		Payload:   []byte(`{"user_id":"` + userID + `","item_id":"` + itemID + `"}`),
	}}); err != nil {
		t.Fatalf("purchase: %v", err)
	}

	var outboxN int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox WHERE event_type='shop.purchased' AND processed=false`).Scan(&outboxN); err != nil {
		t.Fatalf("outbox count: %v", err)
	}
	if outboxN != 1 {
		t.Fatalf("outbox rows=%d want 1", outboxN)
	}

	var stock, version int
	if err := db.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id=$1`, itemID).Scan(&stock, &version); err != nil {
		t.Fatal(err)
	}
	if stock != 1 {
		t.Fatalf("stock=%d want 1", stock)
	}
	if version != 2 {
		t.Fatalf("version=%d want 2", version)
	}

	var owned int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory WHERE user_id=$1 AND item_id=$2`, userID, itemID).Scan(&owned)
	if owned != 1 {
		t.Fatalf("owned=%d", owned)
	}

	// exhaust stock
	u2 := uuid.NewString()
	if err := repo.PurchaseItemWithStock(ctx, u2, itemID, uuid.NewString(), 50, nil); err != nil {
		t.Fatalf("second purchase: %v", err)
	}
	err = repo.PurchaseItemWithStock(ctx, uuid.NewString(), itemID, uuid.NewString(), 50, nil)
	if err == nil {
		t.Fatal("expected out of stock")
	}
}
--- services/shop/internal/repository/postgres_repo.go ---
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
)

type Item struct {
	ID              string
	Name            string
	Description     string
	Price           int
	Category        string
	GameID          *string
	ImageURL        string
	Available       bool
	Owned           bool
	PurchasedAt     *time.Time
	PurchasePrice   int
	PurchaseID      string
	RefundableUntil *time.Time
	FulfilledAt     *time.Time
	CanCancel       bool
}

// RefundResult is the outcome of marking a completed purchase as refunded.
type RefundResult struct {
	PurchaseID      string
	Price           int
	AlreadyRefunded bool
}

type PostgresShopRepo struct {
	db *sql.DB
}

func NewPostgresShopRepo(db *sql.DB) *PostgresShopRepo {
	return &PostgresShopRepo{db: db}
}

func (r *PostgresShopRepo) GetItems(ctx context.Context, category, gameID string) ([]Item, error) {
	query := `SELECT id, name, description, price, category, game_id, image_url, available FROM items WHERE available = true`
	args := []interface{}{}
	argIdx := 1

	if category != "" && category != "all" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	if gameID != "" {
		query += fmt.Sprintf(" AND (game_id = $%d OR game_id IS NULL)", argIdx)
		args = append(args, gameID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var gameID sql.NullString
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Category,
			&gameID,
			&item.ImageURL,
			&item.Available,
		)
		if err != nil {
			return nil, err
		}
		if gameID.Valid {
			item.GameID = &gameID.String
		}
		items = append(items, item)
	}
	return items, nil
}

// CreateItemFromInventory создаёт товар из события инвентаря.
// stock mirrors inventory catalog stock (not the old default 999).
func (r *PostgresShopRepo) CreateItemFromInventory(ctx context.Context, itemID, name, description string, price float64, stock int) error {
	if stock < 0 {
		stock = 0
	}
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO items (id, name, description, price, category, game_id, image_url, available, stock)
        VALUES ($1, $2, $3, $4, 'merch', '', '', true, $5)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            price = EXCLUDED.price,
            stock = EXCLUDED.stock,
            available = true
    `, itemID, name, description, int(price), stock)
	return err
}

func (r *PostgresShopRepo) GetItemByID(ctx context.Context, itemID string) (*Item, error) {
	query := `SELECT id, name, description, price, category, game_id, image_url, available FROM items WHERE id = $1`
	var item Item
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.Name, &item.Description, &item.Price,
		&item.Category, &item.GameID, &item.ImageURL, &item.Available,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PostgresShopRepo) IsItemOwned(ctx context.Context, userID, itemID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM inventory WHERE user_id = $1 AND item_id = $2)`
	err := r.db.QueryRowContext(ctx, query, userID, itemID).Scan(&exists)
	return exists, err
}

// ListOwnedItemIDs returns all catalog item ids the user owns (one round-trip for GetItems).
func (r *PostgresShopRepo) ListOwnedItemIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT item_id FROM inventory WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostgresShopRepo) PurchaseItem(ctx context.Context, userID, itemID string, price int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем в инвентарь
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		return err
	}

	// Добавляем в историю покупок
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases (user_id, item_id, price) VALUES ($1, $2, $3)`, userID, itemID, price)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresShopRepo) GetUserInventory(ctx context.Context, userID string) ([]Item, error) {
	query := `
        SELECT i.id, i.name, i.description, i.price, i.category, i.game_id, i.image_url, i.available,
               inv.purchased_at,
               COALESCE(p.price, 0) AS purchase_price,
               COALESCE(p.id::text, '') AS purchase_id,
               p.refundable_until,
               p.fulfilled_at
        FROM inventory inv
        JOIN items i ON inv.item_id = i.id
        LEFT JOIN LATERAL (
            SELECT id, price, refundable_until, fulfilled_at
            FROM purchases
            WHERE user_id = inv.user_id
              AND item_id = inv.item_id
              AND status = 'COMPLETED'
              AND refunded_at IS NULL
            ORDER BY COALESCE(completed_at, purchased_at) DESC
            LIMIT 1
        ) p ON true
        WHERE inv.user_id = $1
        ORDER BY inv.purchased_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		var gameID sql.NullString
		var purchasedAt time.Time
		var refundableUntil sql.NullTime
		var fulfilledAt sql.NullTime
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Category,
			&gameID,
			&item.ImageURL,
			&item.Available,
			&purchasedAt,
			&item.PurchasePrice,
			&item.PurchaseID,
			&refundableUntil,
			&fulfilledAt,
		)
		if err != nil {
			return nil, err
		}
		if gameID.Valid {
			item.GameID = &gameID.String
		}
		item.PurchasedAt = &purchasedAt
		if refundableUntil.Valid {
			t := refundableUntil.Time
			item.RefundableUntil = &t
		}
		if fulfilledAt.Valid {
			t := fulfilledAt.Time
			item.FulfilledAt = &t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresShopRepo) CreatePendingPurchase(ctx context.Context, userID, itemID string, price int) (string, error) {
	var purchaseID string
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO purchases (user_id, item_id, price, status) 
        VALUES ($1, $2, $3, 'PENDING') 
        RETURNING id
    `, userID, itemID, price).Scan(&purchaseID)
	return purchaseID, err
}

func (r *PostgresShopRepo) CompletePurchase(ctx context.Context, purchaseID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE purchases SET status = 'COMPLETED', completed_at = NOW() 
        WHERE id = $1 AND status = 'PENDING'
    `, purchaseID)
	return err
}

func (r *PostgresShopRepo) CancelPendingPurchase(ctx context.Context, purchaseID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE purchases SET status = 'CANCELLED' 
        WHERE id = $1 AND status = 'PENDING'
    `, purchaseID)
	return err
}

// RefundPurchase marks the latest completed purchase as REFUNDED, removes the
// inventory row, and restores shop.items stock. Idempotent via refunded_at.
// Enforces 7-day refundable_until and blocks physical items after fulfilled_at.
func (r *PostgresShopRepo) RefundPurchase(ctx context.Context, userID, itemID string) (*RefundResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var purchaseID string
	var price int
	var refundableUntil time.Time
	var fulfilledAt sql.NullTime
	var category string
	err = tx.QueryRowContext(ctx, `
		SELECT p.id, p.price, p.refundable_until, p.fulfilled_at, i.category
		FROM purchases p
		JOIN items i ON i.id = p.item_id
		WHERE p.user_id = $1 AND p.item_id = $2
		  AND p.status = 'COMPLETED' AND p.refunded_at IS NULL
		ORDER BY COALESCE(p.completed_at, p.purchased_at) DESC
		LIMIT 1
		FOR UPDATE OF p
	`, userID, itemID).Scan(&purchaseID, &price, &refundableUntil, &fulfilledAt, &category)
	if err == sql.ErrNoRows {
		var alreadyID string
		var alreadyPrice int
		qerr := tx.QueryRowContext(ctx, `
			SELECT id, price FROM purchases
			WHERE user_id = $1 AND item_id = $2
			  AND (status = 'REFUNDED' OR refunded_at IS NOT NULL)
			ORDER BY refunded_at DESC NULLS LAST
			LIMIT 1
		`, userID, itemID).Scan(&alreadyID, &alreadyPrice)
		if qerr == nil {
			return &RefundResult{PurchaseID: alreadyID, Price: alreadyPrice, AlreadyRefunded: true}, nil
		}
		return nil, fmt.Errorf("purchase not found")
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if now.After(refundableUntil) {
		return nil, model.ErrRefundWindowExpired
	}
	if fulfilledAt.Valid && model.IsPhysicalItem(category) {
		return nil, model.ErrAlreadyFulfilled
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE purchases
		SET status = 'REFUNDED', refunded_at = NOW()
		WHERE id = $1 AND status = 'COMPLETED' AND refunded_at IS NULL
	`, purchaseID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &RefundResult{PurchaseID: purchaseID, Price: price, AlreadyRefunded: true}, nil
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM inventory WHERE user_id = $1 AND item_id = $2
	`, userID, itemID); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE items SET stock = stock + 1, version = version + 1 WHERE id = $1
	`, itemID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &RefundResult{PurchaseID: purchaseID, Price: price}, nil
}

// MarkPurchaseFulfilled sets fulfilled_at when fulfillment publishes PurchaseFulfilled.
func (r *PostgresShopRepo) MarkPurchaseFulfilled(ctx context.Context, purchaseID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE purchases
		SET fulfilled_at = NOW()
		WHERE id = $1 AND fulfilled_at IS NULL
	`, purchaseID)
	return err
}

type OutboxRecord struct {
	EventType string
	Payload   []byte
}

// PurchaseSaga is the orchestrator row for one buy that touches billing and shop.
type PurchaseSaga struct {
	ID         string
	UserID     string
	ItemID     string
	PurchaseID string
	SpendRef   string
	RefundRef  string
	Price      int
	State      string
}

// SavePurchaseSaga inserts the orchestrator row for this spend key.
// A retry of the same key returns the original saga id and purchase id.
func (r *PostgresShopRepo) SavePurchaseSaga(ctx context.Context, s PurchaseSaga) (string, string, error) {
	var id, purchaseID string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO purchase_sagas (id, user_id, item_id, purchase_id, spend_ref, refund_ref, price, state)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (spend_ref) DO UPDATE SET updated_at = purchase_sagas.updated_at
		RETURNING id::text, purchase_id::text
	`, s.ID, s.UserID, s.ItemID, s.PurchaseID, s.SpendRef, s.RefundRef, s.Price, s.State).Scan(&id, &purchaseID)
	return id, purchaseID, err
}

// MarkPurchaseSaga moves the orchestrator to the next state.
func (r *PostgresShopRepo) MarkPurchaseSaga(ctx context.Context, id, state string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE purchase_sagas SET state = $2, updated_at = NOW() WHERE id = $1
	`, id, state)
	return err
}

// NextSpendCycle returns the stable billing cycle for this player+item.
// Parallel clicks share it. BumpSpendCycle moves it after a refund.
func (r *PostgresShopRepo) NextSpendCycle(ctx context.Context, userID, itemID string) (int, error) {
	var cycle int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO purchase_claims (user_id, item_id, cycle)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, item_id) DO UPDATE SET cycle = purchase_claims.cycle
		RETURNING cycle
	`, userID, itemID).Scan(&cycle)
	return cycle, err
}

// BumpSpendCycle advances the claim so the next purchase uses a new reference_id.
func (r *PostgresShopRepo) BumpSpendCycle(ctx context.Context, userID, itemID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE purchase_claims SET cycle = cycle + 1 WHERE user_id = $1 AND item_id = $2
	`, userID, itemID)
	return err
}

// EnqueueOutbox writes one event outside the purchase transaction (compensation).
func (r *PostgresShopRepo) EnqueueOutbox(ctx context.Context, eventType string, payload []byte) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO outbox (event_type, payload) VALUES ($1, $2) RETURNING id
	`, eventType, payload).Scan(&id)
	return id, err
}

// MarkOutboxProcessed closes a row the worker would otherwise retry.
func (r *PostgresShopRepo) MarkOutboxProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox SET processed = true, processed_at = NOW() WHERE id = $1
	`, id)
	return err
}

// PurchaseItemWithStock — покупка с проверкой стока. purchaseID is the row id
// and the PurchaseUUID of every outbox event written in this transaction.
func (r *PostgresShopRepo) PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []OutboxRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Pessimistic row lock + optimistic version bump on stock change.
	var stock, version int
	err = tx.QueryRowContext(ctx, `SELECT stock, version FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&stock, &version)
	if err != nil {
		return err
	}
	if stock <= 0 {
		return fmt.Errorf("item out of stock")
	}

	res, err := tx.ExecContext(ctx, `
        UPDATE items SET stock = stock - 1, version = version + 1 WHERE id = $1 AND version = $2`, itemID, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("optimistic lock conflict on item %s", itemID)
	}

	// Добавляем в инвентарь
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory (user_id, item_id) VALUES ($1, $2)`, userID, itemID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO purchases (id, user_id, item_id, price, status, refundable_until)
        VALUES ($1, $2, $3, $4, 'COMPLETED', NOW() + INTERVAL '7 days')`, purchaseID, userID, itemID, price)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.EventType == "" || len(rec.Payload) == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
            INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
			rec.EventType, rec.Payload); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// PurchaseRecord is a row from purchases for author sales.
type PurchaseRecord struct {
	ID          string
	UserID      string
	ItemID      string
	Price       int
	Status      string
	PurchasedAt time.Time
	RefundedAt  *time.Time
}

// ListPurchasesByItemIDs returns purchases for the given catalog item ids (newest first).
func (r *PostgresShopRepo) ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) (rowsOut []PurchaseRecord, total, salesCount, ticketsEarned int64, err error) {
	if len(itemIDs) == 0 {
		return nil, 0, 0, 0, nil
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	arr := pq.Array(itemIDs)
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM purchases WHERE item_id::text = ANY($1)`, arr).Scan(&total); err != nil {
		return nil, 0, 0, 0, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(price), 0)
		FROM purchases
		WHERE item_id::text = ANY($1)
		  AND status = 'COMPLETED'
		  AND refunded_at IS NULL`, arr).Scan(&salesCount, &ticketsEarned); err != nil {
		return nil, 0, 0, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, item_id::text, price, COALESCE(status, 'COMPLETED'),
		       purchased_at, refunded_at
		FROM purchases
		WHERE item_id::text = ANY($1)
		ORDER BY purchased_at DESC
		LIMIT $2 OFFSET $3`, arr, limit, offset)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	out := make([]PurchaseRecord, 0)
	for rows.Next() {
		var p PurchaseRecord
		var refunded sql.NullTime
		if err := rows.Scan(&p.ID, &p.UserID, &p.ItemID, &p.Price, &p.Status, &p.PurchasedAt, &refunded); err != nil {
			return nil, 0, 0, 0, err
		}
		if refunded.Valid {
			t := refunded.Time
			p.RefundedAt = &t
		}
		out = append(out, p)
	}
	return out, total, salesCount, ticketsEarned, rows.Err()
}
--- services/shop/internal/repository/redis_repo.go ---
package repository

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisShopRepo struct {
    client *redis.Client
}

func NewRedisShopRepo(addr string, db int) *RedisShopRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisShopRepo{client: client}
}

func (r *RedisShopRepo) GetItems(ctx context.Context, key string) ([]Item, error) {
    data, err := r.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, err
    }
    var items []Item
    if err := json.Unmarshal(data, &items); err != nil {
        return nil, err
    }
    return items, nil
}

func (r *RedisShopRepo) SetItems(ctx context.Context, key string, items []Item, ttl time.Duration) error {
    data, err := json.Marshal(items)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisShopRepo) Delete(ctx context.Context, key string) error {
    return r.client.Del(ctx, key).Err()
}
===== converter =====
--- services/shop/internal/converter/item.go ---
package converter

import (
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

func ItemToProto(item model.Item) *pb.Item {
	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Category:    item.Category,
		GameId:      item.GameID,
		ImageUrl:    item.ImageURL,
		Available:   item.Available,
	}
}

func ItemsToProto(items []model.Item) []*pb.Item {
	out := make([]*pb.Item, len(items))
	for i, it := range items {
		out[i] = ItemToProto(it)
	}
	return out
}
--- services/shop/internal/converter/item_test.go ---
package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
)

func TestItemToProto(t *testing.T) {
	p := ItemToProto(model.Item{ID: "1", Name: "skin", Price: 100, Available: true})
	if p.Id != "1" || p.Price != 100 || !p.Available {
		t.Fatalf("%+v", p)
	}
}
===== worker =====
--- services/shop/internal/worker/outbox_worker.go ---
package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// CompensationEvent is an outbox row the shop worker settles by calling billing,
// not by publishing to NATS.
const CompensationEvent = "billing.refund"

// CompensateFunc retries a ticket refund. A duplicate reference_id is success.
type CompensateFunc func(ctx context.Context, payload []byte) error

type OutboxWorker struct {
	db         *sql.DB
	js         nats.JetStreamContext
	compensate CompensateFunc
}

func NewOutboxWorker(db *sql.DB, js nats.JetStreamContext, compensate CompensateFunc) *OutboxWorker {
	return &OutboxWorker{db: db, js: js, compensate: compensate}
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
			w.resumeSagas(ctx)
		}
	}
}

type outboxRow struct {
	id        string
	eventType string
	payload   []byte
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("shop outbox: begin", "err", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 20
        FOR UPDATE SKIP LOCKED
    `)
	if err != nil {
		slog.Error("shop outbox: query", "err", err)
		return
	}

	var batch []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.id, &row.eventType, &row.payload); err != nil {
			slog.Error("shop outbox: scan", "err", err)
			continue
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("shop outbox: rows", "err", err)
		return
	}

	for _, row := range batch {
		if err := w.deliver(ctx, row); err != nil {
			slog.Error("shop outbox: deliver", "event_type", row.eventType, "id", row.id, "err", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, row.id); err != nil {
			slog.Error("shop outbox: mark processed", "id", row.id, "err", err)
			return
		}
		slog.Info("shop outbox: published", "event_type", row.eventType, "id", row.id)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("shop outbox: commit", "err", err)
	}
}

func (w *OutboxWorker) deliver(ctx context.Context, row outboxRow) error {
	if row.eventType == CompensationEvent {
		if w.compensate == nil {
			return errNoCompensator
		}
		err := w.compensate(ctx, row.payload)
		if err != nil && !isDuplicateRef(err) {
			return err
		}
		return nil
	}
	if w.js == nil {
		return errNoJetStream
	}
	_, err := w.js.Publish(row.eventType, row.payload)
	return err
}

var (
	errNoCompensator = errString("shop compensation handler is not wired")
	errNoJetStream   = errString("jetstream is not connected")
)

type errString string

func (e errString) Error() string { return string(e) }

func isDuplicateRef(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "23505") || strings.Contains(s, "unique constraint")
}

type sagaResume struct {
	id         string
	userID     string
	itemID     string
	purchaseID string
	refundRef  string
	price      int
}

// resumeSagas finishes a purchase that died after billing spent tickets
// and before the shop row was committed. The row lock is released before the billing call.
func (w *OutboxWorker) resumeSagas(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("shop saga: begin", "err", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, user_id, item_id, purchase_id::text, refund_ref, price
		FROM purchase_sagas
		WHERE state IN ('spent', 'compensating')
		  AND updated_at < NOW() - INTERVAL '30 seconds'
		ORDER BY updated_at
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		slog.Error("shop saga: query", "err", err)
		return
	}

	var refunds []sagaResume
	for rows.Next() {
		var s sagaResume
		if err := rows.Scan(&s.id, &s.userID, &s.itemID, &s.purchaseID, &s.refundRef, &s.price); err != nil {
			slog.Error("shop saga: scan", "err", err)
			continue
		}
		var stocked bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM purchases WHERE id = $1::uuid)
		`, s.purchaseID).Scan(&stocked); err != nil {
			slog.Error("shop saga: purchase lookup", "id", s.id, "err", err)
			continue
		}
		if stocked {
			if _, err := tx.ExecContext(ctx, `
				UPDATE purchase_sagas SET state = 'stocked', updated_at = NOW() WHERE id = $1::uuid
			`, s.id); err != nil {
				slog.Error("shop saga: mark stocked", "id", s.id, "err", err)
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE purchase_sagas SET state = 'compensating', updated_at = NOW() WHERE id = $1::uuid
		`, s.id); err != nil {
			slog.Error("shop saga: mark compensating", "id", s.id, "err", err)
			continue
		}
		refunds = append(refunds, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		slog.Error("shop saga: rows", "err", err)
		return
	}
	if err := tx.Commit(); err != nil {
		slog.Error("shop saga: commit", "err", err)
		return
	}

	for _, s := range refunds {
		if w.compensate == nil {
			slog.Error("shop saga: compensator is not wired", "id", s.id)
			continue
		}
		body, _ := json.Marshal(map[string]any{
			"user_id":      s.userID,
			"item_id":      s.itemID,
			"amount":       s.price,
			"reference_id": s.refundRef,
			"error":        "saga resume",
		})
		if err := w.compensate(ctx, body); err != nil && !isDuplicateRef(err) {
			slog.Error("shop saga: refund", "id", s.id, "err", err)
			continue
		}
		if _, err := w.db.ExecContext(ctx, `
			UPDATE purchase_claims SET cycle = cycle + 1
			WHERE user_id = $1::uuid AND item_id = $2::uuid
		`, s.userID, s.itemID); err != nil {
			slog.Error("shop saga: bump cycle", "id", s.id, "err", err)
			continue
		}
		if _, err := w.db.ExecContext(ctx, `
			UPDATE purchase_sagas SET state = 'compensated', updated_at = NOW() WHERE id = $1::uuid
		`, s.id); err != nil {
			slog.Error("shop saga: mark compensated", "id", s.id, "err", err)
		}
	}
}
[denismatveev@c0der event_horizon]$ 
