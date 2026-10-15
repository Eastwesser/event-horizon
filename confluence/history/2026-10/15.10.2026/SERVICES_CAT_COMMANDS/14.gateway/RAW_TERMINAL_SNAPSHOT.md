[denismatveev@c0der event_horizon]$ find services/gateway -type f -name '*.go' | sort
echo "===== main ====="
cat services/gateway/cmd/main.go
echo "===== app ====="
for f in services/gateway/internal/app/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== middleware ====="
for f in services/gateway/internal/middleware/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== ratelimit ====="
for f in services/gateway/internal/ratelimit/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== circuit ====="
for f in services/gateway/internal/circuit/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== cache ====="
for f in services/gateway/internal/cache/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== client ====="
for f in services/gateway/internal/client/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== dto ====="
for f in services/gateway/internal/dto/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== websocket ====="
for f in services/gateway/internal/websocket/*.go; do echo "--- $f ---"; cat "$f"; done
echo "===== webhook ====="
for f in services/gateway/internal/webhook/*.go; do echo "--- $f ---"; cat "$f"; done
services/gateway/api/doc.go
services/gateway/api/swagger.go
services/gateway/cmd/main.go
services/gateway/internal/app/app.go
services/gateway/internal/app/gateway.go
services/gateway/internal/app/upload.go
services/gateway/internal/cache/response_cache.go
services/gateway/internal/cache/score_cache.go
services/gateway/internal/circuit/breaker.go
services/gateway/internal/circuit/breaker_test.go
services/gateway/internal/client/auth_client.go
services/gateway/internal/client/dial.go
services/gateway/internal/client/retry.go
services/gateway/internal/client/retry_test.go
services/gateway/internal/config/config.go
services/gateway/internal/dto/inventory.go
services/gateway/internal/dto/inventory_test.go
services/gateway/internal/dto/shop.go
services/gateway/internal/dto/shop_test.go
services/gateway/internal/dto/twins.go
services/gateway/internal/middleware/auth.go
services/gateway/internal/middleware/auth_role_test.go
services/gateway/internal/middleware/docs.go
services/gateway/internal/middleware/docs_test.go
services/gateway/internal/middleware/ratelimit.go
services/gateway/internal/middleware/security.go
services/gateway/internal/ratelimit/limiter.go
services/gateway/internal/webhook/verify.go
services/gateway/internal/websocket/hub.go
===== main =====
package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init gateway app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("gateway stopped with error: %v", err)
		os.Exit(1)
	}
}
===== app =====
--- services/gateway/internal/app/app.go ---
package app

import (
	"context"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"log/slog"
)

// App is the API gateway process. HTTP route wiring lives in gateway.go
// (extracted from the former cmd/main monolith). New/RunUntilSignal match other services.
type App struct {
	log    *slog.Logger
	closer *closer.Closer
}

// New prepares logging/closer; servers start in RunUntilSignal via runGateway.
func New(ctx context.Context) (*App, error) {
	_ = ctx
	log := logger.NewFrom("info", "text")
	return &App{
		log:    log,
		closer: closer.New(log),
	}, nil
}

// RunUntilSignal blocks in the gateway HTTP server until process signal.
func (a *App) RunUntilSignal(ctx context.Context) error {
	_ = ctx
	a.log.Info("starting gateway bootstrap")
	runGateway()
	return nil
}

// Close is mostly a no-op; runGateway owns its shutdown path today.
func (a *App) Close(ctx context.Context) error {
	if a.closer == nil {
		return nil
	}
	return a.closer.CloseAll(ctx)
}
--- services/gateway/internal/app/gateway.go ---
package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	analyticsPb "github.com/Eastwesser/event-horizon/services/analytics/proto"
	authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
	authorsPb "github.com/Eastwesser/event-horizon/services/authors/proto"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	gamePb "github.com/Eastwesser/event-horizon/services/game/proto"
	"github.com/Eastwesser/event-horizon/services/gateway/api"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/cache"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/circuit"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/client"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/config"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/dto"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/middleware"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
	gwhook "github.com/Eastwesser/event-horizon/services/gateway/internal/webhook"
	historyPb "github.com/Eastwesser/event-horizon/services/history/proto"
	inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
	leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
	notificationPb "github.com/Eastwesser/event-horizon/services/notification/proto"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	profilePb "github.com/Eastwesser/event-horizon/services/profile/proto"
	shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

const (
	RoleUser   = "user"
	RoleAuthor = "author"
	RoleAdmin  = "admin"
)

func parseBillingCurrency(raw string) (billingPb.CurrencyType, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "lamps":
		return billingPb.CurrencyType_LAMPS, true
	case "tickets":
		return billingPb.CurrencyType_TICKETS, true
	default:
		return billingPb.CurrencyType_CURRENCY_UNSPECIFIED, false
	}
}

func billingCurrencyName(c billingPb.CurrencyType) string {
	switch c {
	case billingPb.CurrencyType_LAMPS:
		return "lamps"
	case billingPb.CurrencyType_TICKETS:
		return "tickets"
	default:
		return ""
	}
}

func withUserRole(ctx context.Context, c *gin.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"x-user-role", middleware.Role(c),
		"x-user-id", middleware.UserID(c),
	)
}

func newServiceBreaker(name string) *circuit.Breaker {
	return circuit.New(circuit.Settings{
		Name:          name,
		MaxRequests:   3,
		Timeout:       10 * time.Second,
		MaxConcurrent: 32, // bulkhead: cap in-flight calls per downstream
		ReadyToTrip: func(counts circuit.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
}

// appendPartialJSON adds "_partial": true to a cached whoami JSON object.
func appendPartialJSON(cached []byte) []byte {
	var m map[string]any
	if err := json.Unmarshal(cached, &m); err != nil {
		return cached
	}
	m["_partial"] = true
	b, err := json.Marshal(m)
	if err != nil {
		return cached
	}
	return b
}

// throughBreaker runs fn under the circuit breaker + bulkhead.
// On open circuit or full bulkhead writes 503 and returns the sentinel error.
func throughBreaker(b *circuit.Breaker, c *gin.Context, fn func() (any, error)) (any, error) {
	out, err := b.Execute(fn)
	if err == circuit.ErrOpen {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable", "circuit": b.Name()})
		return nil, err
	}
	if err == circuit.ErrBulkheadFull {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service overloaded", "bulkhead": b.Name()})
		return nil, err
	}
	return out, err
}

// maskEmail hides local-part for author sales lists (a***@domain).
func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return "hidden"
	}
	local, domain := email[:at], email[at:]
	if len(local) == 1 {
		return local + "***" + domain
	}
	return string(local[0]) + "***" + domain
}

// listAdminUserIDs pages Auth.ListUsers and keeps role=admin. Best-effort for event enrichment.
func listAdminUserIDs(ctx context.Context, authClient *client.AuthClient, authCB *circuit.Breaker) []string {
	const page = int32(100)
	var out []string
	var offset int32
	for {
		respAny, err := authCB.Execute(func() (any, error) {
			return authClient.GetClient().ListUsers(ctx, &authPb.ListUsersRequest{
				Limit:  page,
				Offset: offset,
			})
		})
		if err != nil {
			log.Printf("listAdminUserIDs: %v", err)
			break
		}
		resp := respAny.(*authPb.ListUsersResponse)
		users := resp.GetUsers()
		if len(users) == 0 {
			break
		}
		for _, u := range users {
			if strings.EqualFold(u.GetRole(), RoleAdmin) && u.GetUserId() != "" {
				out = append(out, u.GetUserId())
			}
		}
		offset += int32(len(users))
		if int64(offset) >= resp.GetTotal() || len(users) < int(page) {
			break
		}
		if offset > 5000 { // safety cap
			break
		}
	}
	return out
}

// writeGRPCError maps common gRPC codes to HTTP (see architecture/STATUS_CODES.md).
func writeGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		// otel / transport wrappers may lose the typed status; Convert still recovers Code when possible.
		st = status.Convert(err)
	}
	if st.Code() == codes.Unknown {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "InvalidArgument"):
			c.JSON(http.StatusBadRequest, gin.H{"error": msg, "grpc_code": "InvalidArgument"})
			return
		case strings.Contains(msg, "AlreadyExists"):
			c.JSON(http.StatusConflict, gin.H{"error": msg, "grpc_code": "AlreadyExists"})
			return
		case strings.Contains(msg, "NotFound"):
			c.JSON(http.StatusNotFound, gin.H{"error": msg, "grpc_code": "NotFound"})
			return
		case strings.Contains(msg, "Unauthenticated"):
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg, "grpc_code": "Unauthenticated"})
			return
		case strings.Contains(msg, "PermissionDenied"):
			c.JSON(http.StatusForbidden, gin.H{"error": msg, "grpc_code": "PermissionDenied"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
	}
	code := http.StatusInternalServerError
	switch st.Code() {
	case codes.InvalidArgument:
		code = http.StatusBadRequest
	case codes.Unauthenticated:
		code = http.StatusUnauthorized
	case codes.PermissionDenied:
		code = http.StatusForbidden
	case codes.NotFound:
		code = http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		code = http.StatusConflict
	case codes.FailedPrecondition:
		code = http.StatusConflict
	case codes.ResourceExhausted:
		code = http.StatusTooManyRequests
	case codes.Unavailable:
		code = http.StatusServiceUnavailable
	}
	body := gin.H{"error": st.Message(), "grpc_code": st.Code().String()}
	if st.Message() == "subscription_required" {
		body["code"] = "subscription_required"
	}
	c.JSON(code, body)
}

// handleRPCError maps gRPC errors to HTTP via writeGRPCError. Returns true if caller should return.
func handleRPCError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	writeGRPCError(c, err)
	return true
}

// writePaymentConfirmError maps ConfirmPayment RPC errors; duplicate webhook delivery → HTTP 200.
func writePaymentConfirmError(c *gin.Context, err error) {
	if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": st.Message()})
		return
	}
	writeGRPCError(c, err)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: originAllowed,
}

// originAllowed accepts the page that loaded the app (same host), localhost, and WS_ALLOWED_ORIGINS.
// An empty Origin is a non-browser client; the handler still requires a bearer token or a one-time ticket.
func originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) || strings.HasPrefix(u.Host, "localhost:") || u.Host == "localhost" || strings.HasPrefix(u.Host, "127.0.0.1:") || u.Host == "127.0.0.1" {
		return true
	}
	for _, allowed := range strings.Split(os.Getenv("WS_ALLOWED_ORIGINS"), ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}

// issueWSTicket stores a random 30-second Redis key.
// The browser puts that ticket in the socket URL and keeps the JWT in the Authorization header.
func issueWSTicket(ctx context.Context, rdb *redis.Client, userID string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(buf)
	if err := rdb.Set(ctx, "ws:ticket:"+ticket, userID, 30*time.Second).Err(); err != nil {
		return "", err
	}
	return ticket, nil
}

// consumeWSTicket deletes the key on the first read, so the same ticket cannot open a second socket.
func consumeWSTicket(ctx context.Context, rdb *redis.Client, ticket string) (string, error) {
	if len(ticket) < 16 || len(ticket) > 64 {
		return "", redis.Nil
	}
	return rdb.GetDel(ctx, "ws:ticket:"+ticket).Result()
}

type WSClient struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			slog.Info("websocket client connected", "total", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			slog.Info("websocket client disconnected", "total", len(h.clients))
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *WSClient) writePump() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func initTracer(ctx context.Context) (func(context.Context) error, error) {
	endpoint := os.Getenv("JAEGER_ENDPOINT")
	if endpoint == "" {
		endpoint = "jaeger:4317" // ← для Docker используем имя контейнера
	}
	log.Printf("🔄 Initializing Jaeger tracer with endpoint: %s", endpoint)

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("gateway"),
			attribute.String("environment", "development"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Println("✅ Jaeger tracer initialized for Gateway")
	return tp.Shutdown, nil
}

func runGateway() {
	cfg := config.Load()
	ctx := context.Background()

	// Инициализация Jaeger
	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to initialize Jaeger tracer: %v", err)
	} else {
		defer shutdown(ctx)
		log.Println("✅ Jaeger tracer initialized for Gateway")
	}

	log.Printf("🚀 Starting Gateway with config:")
	log.Printf("   Port: %s", cfg.Port)
	log.Printf("   Metrics: %s", cfg.MetricsPort)
	log.Printf("   NATS: %s", cfg.NATSUrl)
	log.Printf("   Redis: %s", cfg.RedisAddr)
	log.Printf("   Auth: %s", cfg.AuthAddr)
	log.Printf("   Game: %s", cfg.GameAddr)
	log.Printf("   Billing: %s", cfg.BillingAddr)
	log.Printf("   Leaderboard: %s", cfg.LeaderboardAddr)

	hub := NewHub()
	go hub.Run()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("📊 Metrics endpoint: http://0.0.0.0:%s/metrics", cfg.MetricsPort)
		if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
			log.Printf("API Gateway metrics server error: %v", err)
		}
	}()

	authClient, err := client.NewAuthClient(cfg.AuthAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth: %v", err)
	}
	defer authClient.Close()
	authCB := newServiceBreaker("auth")

	gameConn, err := client.Dial(cfg.GameAddr)
	if err != nil {
		log.Fatalf("Failed to connect to game: %v", err)
	}
	defer gameConn.Close()
	gameClient := gamePb.NewGameServiceClient(gameConn)
	gameCB := newServiceBreaker("game")

	profileConn, err := client.Dial(cfg.ProfileAddr)
	if err != nil {
		log.Fatalf("Failed to connect to profile: %v", err)
	}
	defer profileConn.Close()
	profileClient := profilePb.NewProfileServiceClient(profileConn)
	profileCB := newServiceBreaker("profile")

	leaderboardConn, err := client.Dial(cfg.LeaderboardAddr)
	if err != nil {
		log.Fatalf("Failed to connect to leaderboard: %v", err)
	}
	defer leaderboardConn.Close()
	leaderboardClient := leaderboardPb.NewLeaderboardServiceClient(leaderboardConn)
	leaderboardCB := newServiceBreaker("leaderboard")

	// NATS SECTION
	var js nats.JetStreamContext
	nc, err := nats.Connect(cfg.NATSUrl)
	if err != nil {
		log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)
	} else {
		defer nc.Drain()
		js, err = nc.JetStream()
		if err != nil {
			log.Printf("⚠️ Failed to create JetStream context: %v", err)
		} else {
			_, err = js.AddStream(&nats.StreamConfig{
				Name:     "EVENTS",
				Subjects: []string{"event.>", "score.updated"},
				Storage:  nats.FileStorage,
			})
			if err != nil {
				log.Printf("Stream might already exist: %v", err)
			}

			_, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
				hub.Broadcast(msg.Data)
				log.Printf("📡 Broadcasted score update to WebSocket clients")
			}, nats.Durable("gateway-websocket"))
			if err != nil {
				log.Printf("Failed to subscribe to score.updated: %v", err)
			}
		}
	}

	// GIN ROUTER SECTION
	// gin.Default logs the raw URL, which would keep a socket ticket in the access log.
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Printf("%s %s %d %s", c.Request.Method, path, c.Writer.Status(), time.Since(start).Round(time.Millisecond))
	})
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RejectUnsafeRequest())
	r.Use(middleware.LimitJSONBody())

	// Canonical HTTP prefix is /api/v1/*. Legacy /api/* is rewritten and re-dispatched
	// (Gin matches routes before middleware, so mutating Path alone is not enough).
	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") && !strings.HasPrefix(p, "/api/v1/") {
			c.Request.URL.Path = "/api/v1/" + strings.TrimPrefix(p, "/api/")
			r.HandleContext(c)
			c.Abort()
			return
		}
		c.Next()
	})

	r.Use(otelgin.Middleware("gateway"))

	gatewayRequestsTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of HTTP requests to Gateway",
		},
		[]string{"method", "path", "status"},
	)

	gatewayRequestDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		path := c.Request.URL.Path
		method := c.Request.Method

		gatewayRequestsTotal.WithLabelValues(method, path, status).Inc()
		gatewayRequestDuration.WithLabelValues(method, path).Observe(duration)
	})

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("⚠️ Redis connection failed: %v", err)
	} else {
		log.Println("✅ Redis connected for rate limiter")
		limiter := ratelimit.NewRateLimiter(rdb)
		r.Use(middleware.RateLimitMiddleware(limiter))
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	r.GET("/ready", func(c *gin.Context) {
		readyCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(readyCtx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "reason": "redis unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "service": "gateway"})
	})

	registerUploadRoutes(r, cfg, authClient)

	// OpenAPI stays off the public internet unless DOCS_PASSWORD is set.
	docsGate := middleware.DocsGate(cfg.DocsUser, cfg.DocsPassword)
	r.GET("/openapi.yaml", docsGate, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.OpenAPIYAML)
	})
	r.GET("/docs/swagger-ui.css", docsGate, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/css; charset=utf-8", api.SwaggerCSS)
	})
	r.GET("/docs/swagger-ui-bundle.js", docsGate, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/javascript; charset=utf-8", api.SwaggerJS)
	})
	r.GET("/docs/swagger-init.js", docsGate, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/javascript; charset=utf-8", api.SwaggerInit)
	})
	r.GET("/docs", docsGate, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(api.DocsHTML))
	})

	r.POST("/api/v1/ws-ticket", middleware.RequireAuth(authClient), func(c *gin.Context) {
		userID := middleware.UserID(c)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
			return
		}
		ticket, err := issueWSTicket(c.Request.Context(), rdb, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "ticket store unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ticket": ticket, "expires_in": 30})
	})

	r.GET("/ws/leaderboard", func(c *gin.Context) {
		if !originAllowed(c.Request) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			return
		}
		if strings.TrimSpace(c.Query("token")) != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "use a one-time ticket from POST /api/v1/ws-ticket"})
			return
		}
		if ticket := strings.TrimSpace(c.Query("ticket")); ticket != "" {
			if _, err := consumeWSTicket(c.Request.Context(), rdb, ticket); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired ticket"})
				return
			}
		} else {
			token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing ticket"})
				return
			}
			resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
			if err != nil {
				if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
					c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "session store unavailable"})
					return
				}
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			if resp == nil || !resp.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				return
			}
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		client := &WSClient{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		hub.register <- client
		go client.writePump()
		go client.readPump()
	})

	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req authPb.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Public registration is always role=user. Reject self-serve author/admin.
		if role := strings.TrimSpace(req.Role); role != "" && role != "user" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "public registration only allows role=user"})
			return
		}
		req.Role = ""
		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().Register(c.Request.Context(), &req)
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.RegisterResponse)
		eventData := map[string]interface{}{
			"event":   "user.registered",
			"user_id": resp.UserId,
			"email":   resp.Email,
		}
		eventJSON, _ := json.Marshal(eventData)
		if js != nil {
			// Profile listens on event.user.registered; history also on user.registered.
			js.Publish("event.user.registered", eventJSON)
			js.Publish("user.registered", eventJSON)
			log.Printf("📡 Published event: user.registered for %s", resp.Email)
		} else {
			log.Printf("⚠️ NATS not available, event not published")
		}
		c.JSON(http.StatusOK, resp)
	})

	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req authPb.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().Login(c.Request.Context(), &req)
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.LoginResponse)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"token_type":    resp.TokenType,
			"expires_in":    resp.ExpiresIn,
			"user_id":       resp.UserId,
			"email":         resp.Email,
			"role":          resp.Role,
		})
	})

	r.POST("/api/v1/auth/refresh", func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
			return
		}
		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().RefreshToken(c.Request.Context(), &authPb.RefreshTokenRequest{
				RefreshToken: req.RefreshToken,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.RefreshTokenResponse)
		c.JSON(http.StatusOK, gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"token_type":    resp.TokenType,
			"expires_in":    resp.ExpiresIn,
			"user_id":       resp.UserId,
			"role":          resp.Role,
		})
	})

	scoreCache := cache.NewScoreCache(2 * time.Second)
	authReadCache := cache.NewResponseCache(5 * time.Second)

	cacheWhoami := func(token string, payload gin.H) {
		if b, err := json.Marshal(payload); err == nil {
			authReadCache.Set("whoami:"+token, b)
		}
	}

	r.GET("/api/v1/auth/whoami", middleware.RequireAuth(authClient), func(c *gin.Context) {
		token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))
		if cached, ok := authReadCache.Get("whoami:" + token); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().Whoami(c.Request.Context(), &authPb.WhoamiRequest{AccessToken: token})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.WhoamiResponse)
		payload := gin.H{
			"user_id":  resp.UserId,
			"email":    resp.Email,
			"role":     resp.Role,
			"nickname": resp.Nickname,
		}
		cacheWhoami(token, payload)
		c.JSON(http.StatusOK, payload)
	})

	r.POST("/api/v1/auth/logout", middleware.RequireAuth(authClient), func(c *gin.Context) {
		token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))
		_, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().Logout(c.Request.Context(), &authPb.LogoutRequest{Token: token})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "logged out"})
	})

	// Admin-only: change a user's role (user/author/admin).
	r.POST("/api/v1/auth/update-role", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().UpdateRole(c.Request.Context(), &authPb.UpdateRoleRequest{
				UserId: req.UserID,
				Role:   req.Role,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.UpdateRoleResponse)

		c.JSON(http.StatusOK, gin.H{"success": resp.Success, "message": resp.Message})
	})

	r.GET("/api/v1/auth/user", middleware.RequireAuth(authClient), func(c *gin.Context) {
		userID := middleware.UserID(c)
		if cached, ok := authReadCache.Get("user:" + userID); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().GetUser(c.Request.Context(), &authPb.GetUserRequest{
				UserId: userID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authPb.GetUserResponse)
		payload := gin.H{
			"user_id":     resp.UserId,
			"email":       resp.Email,
			"nickname":    resp.Nickname,
			"best_scores": resp.BestScores,
			"total_score": resp.TotalScore,
			"role":        resp.Role,
		}
		if b, err := json.Marshal(payload); err == nil {
			authReadCache.Set("user:"+userID, b)
		}
		c.JSON(http.StatusOK, payload)
	})

	r.GET("/api/v1/profile", middleware.RequireAuth(authClient), func(c *gin.Context) {
		userID := middleware.UserID(c)
		token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))

		if cached, ok := authReadCache.Get("profile:" + userID); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		out, err := throughBreaker(profileCB, c, func() (any, error) {
			return profileClient.GetProfile(c.Request.Context(), &profilePb.GetProfileRequest{
				UserId: userID,
			})
		})
		if err == circuit.ErrOpen {
			if cached, ok := authReadCache.Get("whoami:" + token); ok {
				c.Data(http.StatusOK, "application/json", appendPartialJSON(cached))
				return
			}
			return
		}
		if handleRPCError(c, err) {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
				if cached, ok := authReadCache.Get("whoami:" + token); ok {
					c.Data(http.StatusOK, "application/json", appendPartialJSON(cached))
					return
				}
			}
			return
		}
		resp := out.(*profilePb.GetProfileResponse)

		b, err := json.Marshal(resp)
		if err == nil {
			authReadCache.Set("profile:"+userID, b)
		}
		c.JSON(http.StatusOK, resp)
	})

	billingConn, err := client.Dial(cfg.BillingAddr)
	if err != nil {
		log.Fatalf("Failed to connect to billing: %v", err)
	}
	defer billingConn.Close()
	billingClient := billingPb.NewBillingServiceClient(billingConn)
	billingCB := newServiceBreaker("billing")

	r.GET("/api/v1/billing/balance/all", middleware.RequireAuth(authClient), func(c *gin.Context) {
		out, err := throughBreaker(billingCB, c, func() (any, error) {
			return billingClient.GetAllBalances(c.Request.Context(), &billingPb.GetAllBalancesRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*billingPb.GetAllBalancesResponse)

		var lamps, tickets int32
		for _, b := range resp.Balances {
			if b.Currency == billingPb.CurrencyType_LAMPS {
				lamps = b.Balance
			} else if b.Currency == billingPb.CurrencyType_TICKETS {
				tickets = b.Balance
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"lamps":   lamps,
			"tickets": tickets,
		})
	})

	r.GET("/api/v1/billing/balance", middleware.RequireAuth(authClient), func(c *gin.Context) {
		currency, ok := parseBillingCurrency(c.Query("currency"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be lamps or tickets"})
			return
		}
		out, err := throughBreaker(billingCB, c, func() (any, error) {
			return billingClient.GetBalance(c.Request.Context(), &billingPb.GetBalanceRequest{
				UserId:   middleware.UserID(c),
				Currency: currency,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*billingPb.GetBalanceResponse)
		c.JSON(http.StatusOK, gin.H{
			"currency":   billingCurrencyName(resp.GetCurrency()),
			"balance":    resp.GetBalance(),
			"updated_at": resp.GetUpdatedAt(),
		})
	})

	r.GET("/api/v1/billing/transactions", middleware.RequireAuth(authClient), func(c *gin.Context) {
		currency, ok := parseBillingCurrency(c.Query("currency"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be lamps or tickets"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if offset < 0 {
			offset = 0
		}
		out, err := throughBreaker(billingCB, c, func() (any, error) {
			return billingClient.GetTransactionHistory(c.Request.Context(), &billingPb.GetTransactionHistoryRequest{
				UserId:   middleware.UserID(c),
				Currency: currency,
				Limit:    int32(limit),
				Offset:   int32(offset),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*billingPb.GetTransactionHistoryResponse)
		items := make([]gin.H, 0, len(resp.GetTransactions()))
		for _, tx := range resp.GetTransactions() {
			items = append(items, gin.H{
				"id":            tx.GetId(),
				"currency":      billingCurrencyName(tx.GetCurrency()),
				"amount":        tx.GetAmount(),
				"balance_after": tx.GetBalanceAfter(),
				"reason":        tx.GetReason(),
				"created_at":    tx.GetCreatedAt(),
			})
		}
		c.JSON(http.StatusOK, gin.H{"transactions": items, "total": resp.GetTotal()})
	})

	// Inventory client is needed by shop purchase (decrement catalog stock) and
	// by /api/inventory/* routes below — create once, reuse.
	inventoryConn, err := client.Dial(cfg.InventoryAddr)
	if err != nil {
		log.Fatalf("Failed to connect to inventory: %v", err)
	}
	defer inventoryConn.Close()
	inventoryClient := inventoryPb.NewInventoryServiceClient(inventoryConn)
	inventoryCB := newServiceBreaker("inventory")

	shopConn, err := client.Dial(cfg.ShopAddr)
	if err != nil {
		log.Fatalf("Failed to connect to shop: %v", err)
	}
	defer shopConn.Close()
	shopClient := shopPb.NewShopServiceClient(shopConn)
	shopCB := newServiceBreaker("shop")

	// Добавь эндпоинты:
	r.GET("/api/v1/shop/items", middleware.RequireAuth(authClient), func(c *gin.Context) {
		category := c.Query("category")
		gameID := c.Query("game_id")

		out, err := throughBreaker(shopCB, c, func() (any, error) {
			return shopClient.GetItems(c.Request.Context(), &shopPb.GetItemsRequest{
				UserId:   middleware.UserID(c),
				Category: category,
				GameId:   gameID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*shopPb.GetItemsResponse)
		c.JSON(http.StatusOK, dto.ShopItemsCatalog(resp.GetItems()))
	})

	r.POST("/api/v1/shop/purchase", middleware.RequireAuth(authClient), func(c *gin.Context) {
		var req struct {
			ItemID string `json:"item_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		out, err := throughBreaker(shopCB, c, func() (any, error) {
			return shopClient.PurchaseItem(c.Request.Context(), &shopPb.PurchaseItemRequest{
				UserId: middleware.UserID(c),
				ItemId: req.ItemID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*shopPb.PurchaseItemResponse)

		// Catalog cards live in inventory; shop.items.stock is a synced mirror
		// (often default 999). Decrement the inventory source of truth so the
		// shop UI "В наличии: N" stays correct. Ignore NotFound (pure shop skins).
		// Shop purchase already committed — never write 503 from inventory CB here.
		var remainingStock *int32
		if req.ItemID != "" {
			resOut, resErr := inventoryCB.Execute(func() (any, error) {
				return inventoryClient.ReserveItem(withUserRole(c.Request.Context(), c), &inventoryPb.ReserveItemRequest{
					Id:       req.ItemID,
					Quantity: 1,
				})
			})
			if resErr == nil {
				if rresp, ok := resOut.(*inventoryPb.ReserveItemResponse); ok {
					v := rresp.GetRemainingStock()
					remainingStock = &v
				}
			} else if st, ok := status.FromError(resErr); ok && st.Code() == codes.NotFound {
				// Shop-only item (skin/theme) — no inventory row.
			} else {
				log.Printf("shop purchase: inventory reserve failed item=%s: %v", req.ItemID, resErr)
			}
		}

		body := gin.H{
			"success":     resp.GetSuccess(),
			"message":     resp.GetMessage(),
			"new_balance": resp.GetNewBalance(),
		}
		if remainingStock != nil {
			body["remaining_stock"] = *remainingStock
		}
		c.JSON(http.StatusOK, body)
	})

	r.POST("/api/v1/shop/purchase/:id/cancel", middleware.RequireAuth(authClient), func(c *gin.Context) {
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		out, err := throughBreaker(shopCB, c, func() (any, error) {
			return shopClient.CancelPurchase(c.Request.Context(), &shopPb.CancelPurchaseRequest{
				UserId: middleware.UserID(c),
				ItemId: itemID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*shopPb.CancelPurchaseResponse)

		var remainingStock *int32
		if !resp.GetAlreadyRefunded() && itemID != "" {
			// Shop cancel already committed — never write 503 from inventory CB here.
			relOut, relErr := inventoryCB.Execute(func() (any, error) {
				return inventoryClient.ReleaseItem(withUserRole(c.Request.Context(), c), &inventoryPb.ReleaseItemRequest{
					Id:       itemID,
					Quantity: 1,
				})
			})
			if relErr == nil {
				if rresp, ok := relOut.(*inventoryPb.ReleaseItemResponse); ok {
					v := rresp.GetRemainingStock()
					remainingStock = &v
				}
			} else if st, ok := status.FromError(relErr); ok && st.Code() == codes.NotFound {
				// Shop-only item — no inventory catalog row.
			} else {
				log.Printf("shop cancel: inventory release failed item=%s: %v", itemID, relErr)
			}
		}

		body := gin.H{
			"success":          resp.GetSuccess(),
			"message":          resp.GetMessage(),
			"new_balance":      resp.GetNewBalance(),
			"refunded_amount":  resp.GetRefundedAmount(),
			"already_refunded": resp.GetAlreadyRefunded(),
		}
		if remainingStock != nil {
			body["remaining_stock"] = *remainingStock
		}
		c.JSON(http.StatusOK, body)
	})

	r.GET("/api/v1/shop/inventory", middleware.RequireAuth(authClient), func(c *gin.Context) {
		out, err := throughBreaker(shopCB, c, func() (any, error) {
			return shopClient.GetInventory(c.Request.Context(), &shopPb.GetInventoryRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*shopPb.GetInventoryResponse)
		c.JSON(http.StatusOK, dto.ShopItems(resp.GetItems()))
	})

	// --- Payment (Boosty subscription) ---
	paymentConn, err := client.Dial(cfg.PaymentAddr)
	if err != nil {
		log.Fatalf("Failed to connect to payment: %v", err)
	}
	defer paymentConn.Close()
	paymentClient := paymentPb.NewPaymentServiceClient(paymentConn)
	paymentCB := newServiceBreaker("payment")

	r.POST("/api/v1/payment/checkout", middleware.RequireAuth(authClient), func(c *gin.Context) {
		var req struct {
			Plan string `json:"plan"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Plan == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan is required (present|future)"})
			return
		}
		out, err := throughBreaker(paymentCB, c, func() (any, error) {
			return paymentClient.CreateCheckout(c.Request.Context(), &paymentPb.CreateCheckoutRequest{
				UserId: middleware.UserID(c),
				Plan:   req.Plan,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*paymentPb.CreateCheckoutResponse)
		c.JSON(http.StatusOK, gin.H{
			"payment_id":   resp.PaymentId,
			"checkout_url": resp.CheckoutUrl,
			"amount_rub":   resp.AmountRub,
			"plan":         resp.Plan,
			"status":       resp.Status,
		})
	})

	r.GET("/api/v1/payment/subscription", middleware.RequireAuth(authClient), func(c *gin.Context) {
		out, err := throughBreaker(paymentCB, c, func() (any, error) {
			return paymentClient.GetSubscription(c.Request.Context(), &paymentPb.GetSubscriptionRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*paymentPb.GetSubscriptionResponse)
		c.JSON(http.StatusOK, gin.H{
			"active":          resp.Active,
			"plan":            resp.Plan,
			"status":          resp.Status,
			"expires_at_unix": resp.ExpiresAtUnix,
			"amount_rub":      resp.AmountRub,
		})
	})

	// Admin-only user directory with billing + subscription enrichment.
	r.GET("/api/v1/admin/users", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		q := c.Query("q")
		limit := int32(50)
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = int32(parsed)
			}
		}
		offset := int32(0)
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = int32(parsed)
			}
		}

		out, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().ListUsers(c.Request.Context(), &authPb.ListUsersRequest{
				Query:  q,
				Limit:  limit,
				Offset: offset,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		list := out.(*authPb.ListUsersResponse)

		type adminUser struct {
			UserID       string `json:"user_id"`
			Email        string `json:"email"`
			Role         string `json:"role"`
			Nickname     string `json:"nickname"`
			CreatedAt    string `json:"created_at"`
			Lamps        int32  `json:"lamps"`
			Tickets      int32  `json:"tickets"`
			Subscription gin.H  `json:"subscription"`
		}

		rows := make([]adminUser, len(list.Users))
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, 8)
		ctx := c.Request.Context()

		for i, u := range list.Users {
			i, u := i, u
			rows[i] = adminUser{
				UserID:    u.GetUserId(),
				Email:     u.GetEmail(),
				Role:      u.GetRole(),
				Nickname:  u.GetNickname(),
				CreatedAt: u.GetCreatedAt(),
				Subscription: gin.H{
					"active": false,
					"plan":   "",
					"status": "none",
				},
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var lamps, tickets int32
				if bal, err := billingClient.GetAllBalances(ctx, &billingPb.GetAllBalancesRequest{UserId: u.GetUserId()}); err == nil {
					for _, b := range bal.GetBalances() {
						switch b.GetCurrency() {
						case billingPb.CurrencyType_LAMPS:
							lamps = b.GetBalance()
						case billingPb.CurrencyType_TICKETS:
							tickets = b.GetBalance()
						}
					}
				}
				subH := gin.H{"active": false, "plan": "", "status": "none"}
				if sub, err := paymentClient.GetSubscription(ctx, &paymentPb.GetSubscriptionRequest{UserId: u.GetUserId()}); err == nil && sub != nil {
					subH = gin.H{
						"active": sub.GetActive(),
						"plan":   sub.GetPlan(),
						"status": sub.GetStatus(),
					}
				}

				mu.Lock()
				rows[i].Lamps = lamps
				rows[i].Tickets = tickets
				rows[i].Subscription = subH
				mu.Unlock()
			}()
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"users":  rows,
			"total":  list.GetTotal(),
			"limit":  limit,
			"offset": offset,
		})
	})

	r.GET("/api/v1/payment/can-purchase-merch", middleware.RequireAuth(authClient), func(c *gin.Context) {
		out, err := throughBreaker(paymentCB, c, func() (any, error) {
			return paymentClient.CanPurchaseMerch(c.Request.Context(), &paymentPb.CanPurchaseMerchRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*paymentPb.CanPurchaseMerchResponse)
		c.JSON(http.StatusOK, gin.H{"allowed": resp.Allowed, "reason": resp.Reason})
	})

	r.POST("/api/v1/payment/webhook", func(c *gin.Context) {
		raw, err := c.GetRawData()
		if err != nil || len(raw) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))

		sig := c.GetHeader("X-Boosty-Signature")
		if sig == "" {
			sig = c.GetHeader("X-Hub-Signature-256")
		}

		var req struct {
			PaymentID       string `json:"payment_id"`
			ProviderRef     string `json:"provider_ref"`
			WebhookSecret   string `json:"webhook_secret"`
			WebhookSecretCW string `json:"webhookSecret"`
		}
		if err := json.Unmarshal(raw, &req); err != nil || req.PaymentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
			return
		}
		webhookSecret := req.WebhookSecret
		if webhookSecret == "" {
			webhookSecret = req.WebhookSecretCW
		}
		if webhookSecret == "" {
			webhookSecret = c.GetHeader("X-Webhook-Secret")
		}
		if webhookSecret == "" {
			webhookSecret = c.GetHeader("X-Boosty-Webhook-Secret")
		}
		if webhookSecret == "" {
			webhookSecret = c.Query("webhook_secret")
		}
		if err := gwhook.Authorize(cfg.PaymentWebhookSecret, webhookSecret, raw, sig); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		out, err := throughBreaker(paymentCB, c, func() (any, error) {
			return paymentClient.ConfirmPayment(c.Request.Context(), &paymentPb.ConfirmPaymentRequest{
				PaymentId:     req.PaymentID,
				ProviderRef:   req.ProviderRef,
				WebhookSecret: webhookSecret,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if err != nil {
			writePaymentConfirmError(c, err)
			return
		}
		resp := out.(*paymentPb.ConfirmPaymentResponse)
		c.JSON(http.StatusOK, gin.H{
			"success":         resp.Success,
			"message":         resp.Message,
			"subscription_id": resp.SubscriptionId,
			"expires_at_unix": resp.ExpiresAtUnix,
		})
	})

	// YooKassa webhook stub (local dev): maps a simplified YooKassa notification
	// to our existing payment activation flow.
	//
	// Expected minimal payload (for smoke tests):
	// {
	//   "event": "payment.succeeded",
	//   "object": { "id": "<payment_id>", "status": "succeeded" }
	// }
	r.POST("/api/v1/payment/yookassa/webhook", func(c *gin.Context) {
		var req struct {
			Event         string `json:"event"`
			WebhookSecret string `json:"webhook_secret"`
			Object        struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"object"`
		}

		if err := c.ShouldBindJSON(&req); err != nil || req.Object.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "object.id (payment_id) is required"})
			return
		}

		// Extract webhook secret from body/header/query.
		webhookSecret := req.WebhookSecret
		if webhookSecret == "" {
			webhookSecret = c.GetHeader("X-Webhook-Secret")
		}
		if webhookSecret == "" {
			webhookSecret = c.Query("webhook_secret")
		}

		// Ignore non-success events for the stub.
		if req.Object.Status != "" && req.Object.Status != "succeeded" {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "ignored (status != succeeded)", "event": req.Event})
			return
		}

		out, err := throughBreaker(paymentCB, c, func() (any, error) {
			return paymentClient.ConfirmPayment(c.Request.Context(), &paymentPb.ConfirmPaymentRequest{
				PaymentId:     req.Object.ID,
				ProviderRef:   req.Object.ID,
				WebhookSecret: webhookSecret,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if err != nil {
			writePaymentConfirmError(c, err)
			return
		}

		resp := out.(*paymentPb.ConfirmPaymentResponse)
		c.JSON(http.StatusOK, gin.H{
			"success":         resp.Success,
			"message":         resp.Message,
			"subscription_id": resp.SubscriptionId,
			"expires_at_unix": resp.ExpiresAtUnix,
		})
	})

	// --- Authors ---
	authorsConn, err := client.Dial(cfg.AuthorsAddr)
	if err != nil {
		log.Fatalf("Failed to connect to authors: %v", err)
	}
	defer authorsConn.Close()
	authorsClient := authorsPb.NewAuthorsServiceClient(authorsConn)
	authorsCB := newServiceBreaker("authors")

	r.PUT("/api/v1/authors/me", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		var req struct {
			DisplayName string `json:"display_name"`
			Bio         string `json:"bio"`
			AvatarURL   string `json:"avatar_url"`
			Portfolio   string `json:"portfolio"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.DisplayName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "display_name is required"})
			return
		}
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.UpsertProfile(c.Request.Context(), &authorsPb.UpsertProfileRequest{
				UserId:      middleware.UserID(c),
				DisplayName: req.DisplayName,
				Bio:         req.Bio,
				AvatarUrl:   req.AvatarURL,
				Portfolio:   req.Portfolio,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.UpsertProfileResponse)
		c.JSON(http.StatusOK, dto.Author(resp.Author))
	})

	r.GET("/api/v1/authors/me", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.GetAuthor(c.Request.Context(), &authorsPb.GetAuthorRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.GetAuthorResponse)
		c.JSON(http.StatusOK, dto.Author(resp.Author))
	})

	// Wave 3 C3 — author sales (read-only aggregates + purchase rows).
	r.GET("/api/v1/authors/me/sales", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		authorID := middleware.UserID(c)
		if middleware.Role(c) == RoleAdmin {
			if q := strings.TrimSpace(c.Query("author_id")); q != "" {
				authorID = q
			}
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if limit <= 0 {
			limit = 50
		}

		invOut, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.SearchItems(c.Request.Context(), &inventoryPb.SearchItemsRequest{
				Filters: map[string]string{"author_id": authorID, "include_deleted": "1"},
				Limit:   100,
				Offset:  0,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		invResp := invOut.(*inventoryPb.SearchItemsResponse)
		itemIDs := make([]string, 0, len(invResp.GetItems()))
		names := make(map[string]string, len(invResp.GetItems()))
		for _, it := range invResp.GetItems() {
			itemIDs = append(itemIDs, it.GetId())
			names[it.GetId()] = it.GetName()
		}
		// Page remaining item ids if author has >100 items.
		totalItems := invResp.GetTotal()
		for off := int32(100); int64(off) < totalItems; off += 100 {
			moreOut, err := throughBreaker(inventoryCB, c, func() (any, error) {
				return inventoryClient.SearchItems(c.Request.Context(), &inventoryPb.SearchItemsRequest{
					Filters: map[string]string{"author_id": authorID, "include_deleted": "1"},
					Limit:   100,
					Offset:  off,
				})
			})
			if err == circuit.ErrOpen {
				return
			}
			if handleRPCError(c, err) {
				return
			}
			more := moreOut.(*inventoryPb.SearchItemsResponse)
			for _, it := range more.GetItems() {
				itemIDs = append(itemIDs, it.GetId())
				names[it.GetId()] = it.GetName()
			}
		}

		if len(itemIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"sales_count":    0,
				"tickets_earned": 0,
				"purchases":      []any{},
				"total":          0,
				"author_id":      authorID,
			})
			return
		}

		shopOut, err := throughBreaker(shopCB, c, func() (any, error) {
			return shopClient.ListPurchasesByItemIDs(c.Request.Context(), &shopPb.ListPurchasesByItemIDsRequest{
				ItemIds: itemIDs,
				Limit:   int32(limit),
				Offset:  int32(offset),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		sales := shopOut.(*shopPb.ListPurchasesByItemIDsResponse)

		emailCache := map[string]string{}
		purchases := make([]gin.H, 0, len(sales.GetPurchases()))
		for _, p := range sales.GetPurchases() {
			email := emailCache[p.GetUserId()]
			if email == "" {
				if uOut, uErr := throughBreaker(authCB, c, func() (any, error) {
					return authClient.GetClient().GetUser(c.Request.Context(), &authPb.GetUserRequest{UserId: p.GetUserId()})
				}); uErr == nil {
					if ur, ok := uOut.(*authPb.GetUserResponse); ok {
						email = maskEmail(ur.GetEmail())
						emailCache[p.GetUserId()] = email
					}
				}
				if email == "" {
					email = "hidden"
				}
			}
			purchases = append(purchases, gin.H{
				"id":           p.GetId(),
				"item_id":      p.GetItemId(),
				"item_name":    names[p.GetItemId()],
				"buyer_email":  email,
				"price":        p.GetPrice(),
				"status":       p.GetStatus(),
				"purchased_at": p.GetPurchasedAt(),
				"refunded_at":  p.GetRefundedAt(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"sales_count":    sales.GetSalesCount(),
			"tickets_earned": sales.GetTicketsEarned(),
			"purchases":      purchases,
			"total":          sales.GetTotal(),
			"author_id":      authorID,
			"limit":          limit,
			"offset":         offset,
		})
	})

	// Wave 3 C1 — author application (approval is C2).
	r.POST("/api/v1/authors/apply", middleware.RequireAuth(authClient), func(c *gin.Context) {
		var req struct {
			DisplayName  string `json:"display_name"`
			Portfolio    string `json:"portfolio"`
			Motivation   string `json:"motivation"`
			ContactEmail string `json:"contact_email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.TrimSpace(req.DisplayName) == "" || strings.TrimSpace(req.Motivation) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "display_name and motivation are required"})
			return
		}
		contact := strings.TrimSpace(req.ContactEmail)
		if contact == "" {
			contact = middleware.Email(c)
		}
		if contact == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contact_email is required"})
			return
		}
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.SubmitApplication(c.Request.Context(), &authorsPb.SubmitApplicationRequest{
				UserId:       middleware.UserID(c),
				CallerRole:   middleware.Role(c),
				DisplayName:  strings.TrimSpace(req.DisplayName),
				Portfolio:    strings.TrimSpace(req.Portfolio),
				Motivation:   strings.TrimSpace(req.Motivation),
				ContactEmail: contact,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.SubmitApplicationResponse)
		app := resp.GetApplication()
		if js != nil && app != nil {
			adminIDs := listAdminUserIDs(c.Request.Context(), authClient, authCB)
			payload, _ := json.Marshal(map[string]any{
				"event":          "author.application.submitted",
				"application_id": app.GetId(),
				"user_id":        app.GetUserId(),
				"display_name":   app.GetDisplayName(),
				"admin_ids":      adminIDs,
			})
			if _, err := js.Publish("author.application.submitted", payload); err != nil {
				log.Printf("publish author.application.submitted: %v", err)
			}
		}
		c.JSON(http.StatusOK, resp.Application)
	})

	r.GET("/api/v1/authors/me/application", middleware.RequireAuth(authClient), func(c *gin.Context) {
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.GetMyApplication(c.Request.Context(), &authorsPb.GetMyApplicationRequest{
				UserId: middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.GetMyApplicationResponse)
		c.JSON(http.StatusOK, resp.Application)
	})

	// Wave 3 C2 — admin application review (orchestrates Auth.UpdateRole on approve).
	r.GET("/api/v1/authors/applications", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		statusFilter := strings.TrimSpace(c.DefaultQuery("status", "pending"))
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.ListApplications(c.Request.Context(), &authorsPb.ListApplicationsRequest{
				Status: statusFilter,
				Limit:  int32(limit),
				Offset: int32(offset),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.ListApplicationsResponse)
		c.JSON(http.StatusOK, gin.H{
			"applications": dto.AuthorApplications(resp.GetApplications()),
			"total":        resp.GetTotal(),
			"limit":        limit,
			"offset":       offset,
			"status":       statusFilter,
		})
	})

	r.POST("/api/v1/authors/applications/:id/approve", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		appID := c.Param("id")
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.ApproveApplication(c.Request.Context(), &authorsPb.ApproveApplicationRequest{
				ApplicationId: appID,
				ReviewerId:    middleware.UserID(c),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.ApproveApplicationResponse)
		userID := ""
		if resp.GetApplication() != nil {
			userID = resp.GetApplication().GetUserId()
		}
		if userID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "approved application missing user_id"})
			return
		}

		_, roleErr := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().UpdateRole(c.Request.Context(), &authPb.UpdateRoleRequest{
				UserId: userID,
				Role:   RoleAuthor,
			})
		})
		if roleErr == circuit.ErrOpen {
			_, _ = authorsClient.RevertApplication(c.Request.Context(), &authorsPb.RevertApplicationRequest{
				ApplicationId: appID,
			})
			return
		}
		if roleErr != nil {
			_, _ = authorsClient.RevertApplication(c.Request.Context(), &authorsPb.RevertApplicationRequest{
				ApplicationId: appID,
			})
			if handleRPCError(c, roleErr) {
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to update role after approve"})
			return
		}

		if js != nil {
			payload, _ := json.Marshal(map[string]any{
				"event":          "author.application.approved",
				"application_id": appID,
				"user_id":        userID,
			})
			if _, err := js.Publish("author.application.approved", payload); err != nil {
				log.Printf("publish author.application.approved: %v", err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"application": dto.AuthorApplication(resp.GetApplication()),
			"author":      dto.Author(resp.GetAuthor()),
		})
	})

	r.POST("/api/v1/authors/applications/:id/reject", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		var req struct {
			ReviewerNote string `json:"reviewer_note"`
		}
		_ = c.ShouldBindJSON(&req)
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.RejectApplication(c.Request.Context(), &authorsPb.RejectApplicationRequest{
				ApplicationId: c.Param("id"),
				ReviewerId:    middleware.UserID(c),
				ReviewerNote:  strings.TrimSpace(req.ReviewerNote),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.RejectApplicationResponse)
		c.JSON(http.StatusOK, dto.AuthorApplication(resp.GetApplication()))
	})

	r.GET("/api/v1/authors/:user_id", func(c *gin.Context) {
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.GetAuthor(c.Request.Context(), &authorsPb.GetAuthorRequest{
				UserId: c.Param("user_id"),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.GetAuthorResponse)
		c.JSON(http.StatusOK, resp.Author)
	})

	r.GET("/api/v1/authors", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		out, err := throughBreaker(authorsCB, c, func() (any, error) {
			return authorsClient.ListAuthors(c.Request.Context(), &authorsPb.ListAuthorsRequest{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*authorsPb.ListAuthorsResponse)
		c.JSON(http.StatusOK, gin.H{"authors": dto.Authors(resp.GetAuthors()), "total": resp.GetTotal()})
	})

	// --- Notification (in-app inbox) ---
	notificationConn, err := client.Dial(cfg.NotificationAddr)
	if err != nil {
		log.Fatalf("Failed to connect to notification: %v", err)
	}
	defer notificationConn.Close()
	notificationClient := notificationPb.NewNotificationServiceClient(notificationConn)
	notificationCB := newServiceBreaker("notification")

	r.GET("/api/v1/notifications", middleware.RequireAuth(authClient), func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		unreadOnly := c.Query("unread_only") == "1" || strings.EqualFold(c.Query("unread_only"), "true")
		out, err := throughBreaker(notificationCB, c, func() (any, error) {
			return notificationClient.ListNotifications(c.Request.Context(), &notificationPb.ListNotificationsRequest{
				UserId:     middleware.UserID(c),
				Limit:      int32(limit),
				Offset:     int32(offset),
				UnreadOnly: unreadOnly,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*notificationPb.ListNotificationsResponse)
		c.JSON(http.StatusOK, gin.H{
			"notifications": dto.Notifications(resp.GetNotifications()),
			"total":         resp.GetTotal(),
			"unread_count":  resp.GetUnreadCount(),
		})
	})

	r.POST("/api/v1/notifications/:id/read", middleware.RequireAuth(authClient), func(c *gin.Context) {
		id := c.Param("id")
		if id == "all" {
			id = ""
		}
		out, err := throughBreaker(notificationCB, c, func() (any, error) {
			return notificationClient.MarkRead(c.Request.Context(), &notificationPb.MarkReadRequest{
				UserId:         middleware.UserID(c),
				NotificationId: id,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*notificationPb.MarkReadResponse)
		c.JSON(http.StatusOK, gin.H{"marked": resp.GetMarked()})
	})

	// --- History ---
	historyConn, err := client.Dial(cfg.HistoryAddr)
	if err != nil {
		log.Fatalf("Failed to connect to history: %v", err)
	}
	defer historyConn.Close()
	historyClient := historyPb.NewHistoryServiceClient(historyConn)
	historyCB := newServiceBreaker("history")

	r.GET("/api/v1/history", middleware.RequireAuth(authClient), func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		out, err := throughBreaker(historyCB, c, func() (any, error) {
			return historyClient.ListEvents(c.Request.Context(), &historyPb.ListEventsRequest{
				UserId:    middleware.UserID(c),
				EventType: c.Query("event_type"),
				Limit:     int32(limit),
				Offset:    int32(offset),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*historyPb.ListEventsResponse)
		c.JSON(http.StatusOK, gin.H{"events": dto.HistoryEvents(resp.GetEvents()), "total": resp.GetTotal()})
	})

	// --- Analytics ---
	analyticsConn, err := client.Dial(cfg.AnalyticsAddr)
	if err != nil {
		log.Fatalf("Failed to connect to analytics: %v", err)
	}
	defer analyticsConn.Close()
	analyticsClient := analyticsPb.NewAnalyticsServiceClient(analyticsConn)
	analyticsCB := newServiceBreaker("analytics")

	r.GET("/api/v1/analytics/dau", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		out, err := throughBreaker(analyticsCB, c, func() (any, error) {
			return analyticsClient.GetDAU(c.Request.Context(), &analyticsPb.GetDAURequest{Days: int32(days)})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*analyticsPb.GetDAUResponse)
		c.JSON(http.StatusOK, gin.H{"days": dto.DayCounts(resp.GetDays())})
	})

	r.GET("/api/v1/analytics/mau", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		out, err := throughBreaker(analyticsCB, c, func() (any, error) {
			return analyticsClient.GetMAU(c.Request.Context(), &analyticsPb.GetMAURequest{Days: int32(days)})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*analyticsPb.GetMAUResponse)
		c.JSON(http.StatusOK, gin.H{"mau": resp.Mau, "window_days": resp.WindowDays})
	})

	r.GET("/api/v1/analytics/retention", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		cohort, _ := strconv.Atoi(c.DefaultQuery("cohort_days_ago", "7"))
		window, _ := strconv.Atoi(c.DefaultQuery("window_days", "7"))
		out, err := throughBreaker(analyticsCB, c, func() (any, error) {
			return analyticsClient.GetRetention(c.Request.Context(), &analyticsPb.GetRetentionRequest{
				CohortDaysAgo: int32(cohort),
				WindowDays:    int32(window),
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*analyticsPb.GetRetentionResponse)
		c.JSON(http.StatusOK, gin.H{
			"cohort_day":  resp.GetCohortDay(),
			"cohort_size": resp.GetCohortSize(),
			"points":      dto.RetentionPoints(resp.GetPoints()),
		})
	})

	// Inventory gRPC client created earlier (shared with shop purchase).

	// GET /api/inventory/items — список товаров с фильтрами
	r.GET("/api/v1/inventory/items", middleware.RequireAuth(authClient), func(c *gin.Context) {
		filters := make(map[string]string)
		if authorID := c.Query("author_id"); authorID != "" {
			filters["author_id"] = authorID
		}
		if itemType := c.Query("type"); itemType != "" {
			filters["type"] = itemType
		}
		if priceMin := c.Query("price_min"); priceMin != "" {
			filters["price_min"] = priceMin
		}
		if priceMax := c.Query("price_max"); priceMax != "" {
			filters["price_max"] = priceMax
		}
		if query := c.Query("query"); query != "" {
			filters["query"] = query
		}
		includeDeleted := c.Query("include_deleted") == "1" || strings.EqualFold(c.Query("include_deleted"), "true")
		if includeDeleted {
			role := middleware.Role(c)
			if role != RoleAuthor && role != RoleAdmin {
				c.JSON(http.StatusForbidden, gin.H{"error": "include_deleted requires author or admin"})
				return
			}
			filters["include_deleted"] = "1"
			// Non-admin authors may only see their own deleted rows.
			if role != RoleAdmin {
				filters["author_id"] = middleware.UserID(c)
			}
		}

		limit := int32(20)
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = int32(parsed)
			}
		}
		offset := int32(0)
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = int32(parsed)
			}
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.SearchItems(c.Request.Context(), &inventoryPb.SearchItemsRequest{
				Filters: filters,
				Limit:   limit,
				Offset:  offset,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.SearchItemsResponse)
		// Catalog list DTO: omit card_text / flavor_text (detail GET stays full).
		c.JSON(http.StatusOK, gin.H{
			"items": dto.InventoryItemsCatalog(resp.GetItems()),
			"total": resp.GetTotal(),
		})
	})

	// POST /api/inventory/items — создать товар (только author/admin)
	r.POST("/api/v1/inventory/items", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)

		var req struct {
			Type        string                 `json:"type"`
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Price       float64                `json:"price"`
			Stock       int32                  `json:"stock"`
			Attributes  map[string]interface{} `json:"attributes"`
			Images      []string               `json:"images"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		attrs, err := structpb.NewStruct(req.Attributes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attributes: " + err.Error()})
			return
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.CreateItem(withUserRole(c.Request.Context(), c), &inventoryPb.CreateItemRequest{
				AuthorId:    userID,
				Type:        req.Type,
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				Stock:       req.Stock,
				Attributes:  attrs,
				Images:      req.Images,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.ItemResponse)
		c.JSON(http.StatusOK, dto.InventoryItemResponse(resp))
	})

	// POST /api/inventory/items/bulk — массовое создание (author/admin)
	// Must be registered before /items/:id so "bulk" is not captured as an id.
	r.POST("/api/v1/inventory/items/bulk", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)

		var req struct {
			Items []struct {
				Type        string                 `json:"type"`
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				Price       float64                `json:"price"`
				Stock       int32                  `json:"stock"`
				Attributes  map[string]interface{} `json:"attributes"`
				Images      []string               `json:"images"`
			} `json:"items"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "items array is required"})
			return
		}

		pbItems := make([]*inventoryPb.CreateItemRequest, 0, len(req.Items))
		for _, item := range req.Items {
			attrs, err := structpb.NewStruct(item.Attributes)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attributes: " + err.Error()})
				return
			}
			pbItems = append(pbItems, &inventoryPb.CreateItemRequest{
				AuthorId:    userID,
				Type:        item.Type,
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				Stock:       item.Stock,
				Attributes:  attrs,
				Images:      item.Images,
			})
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.BulkCreateItems(c.Request.Context(), &inventoryPb.BulkCreateItemsRequest{
				Items: pbItems,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.BulkCreateItemsResponse)

		c.JSON(http.StatusOK, gin.H{"success": resp.Success, "count": resp.Count})
	})

	// GET /api/inventory/items/:id — получить товар по ID
	r.GET("/api/v1/inventory/items/:id", middleware.RequireAuth(authClient), func(c *gin.Context) {
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{
				Id: itemID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.ItemResponse)
		c.JSON(http.StatusOK, dto.InventoryItemResponse(resp))
	})

	// PUT /api/inventory/items/:id — обновить товар (author владелец или admin)
	r.PUT("/api/v1/inventory/items/:id", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)

		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		if middleware.Role(c) != RoleAdmin {
			out, err := throughBreaker(inventoryCB, c, func() (any, error) {
				return inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
			})
			if err == circuit.ErrOpen {
				return
			}
			if handleRPCError(c, err) {
				return
			}
			existing := out.(*inventoryPb.ItemResponse)
			if existing.Item == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
				return
			}
			if existing.Item.AuthorId != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own items"})
				return
			}
		}

		var req struct {
			Type        string                 `json:"type"`
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Price       float64                `json:"price"`
			Stock       int32                  `json:"stock"`
			Version     int32                  `json:"version"`
			Attributes  map[string]interface{} `json:"attributes"`
			Images      []string               `json:"images"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		attrs, err := structpb.NewStruct(req.Attributes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attributes"})
			return
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.UpdateItem(withUserRole(c.Request.Context(), c), &inventoryPb.UpdateItemRequest{
				Id:          itemID,
				AuthorId:    userID,
				Type:        req.Type,
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				Stock:       req.Stock,
				Version:     req.Version,
				Attributes:  attrs,
				Images:      req.Images,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.ItemResponse)
		c.JSON(http.StatusOK, dto.InventoryItemResponse(resp))
	})

	// DELETE /api/inventory/items/:id — удалить товар (author владелец или admin)
	r.DELETE("/api/v1/inventory/items/:id", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)

		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		if middleware.Role(c) != RoleAdmin {
			out, err := throughBreaker(inventoryCB, c, func() (any, error) {
				return inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
			})
			if err == circuit.ErrOpen {
				return
			}
			if handleRPCError(c, err) {
				return
			}
			existing := out.(*inventoryPb.ItemResponse)
			if existing.Item == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
				return
			}
			if existing.Item.AuthorId != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own items"})
				return
			}
		}

		_, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.DeleteItem(withUserRole(c.Request.Context(), c), &inventoryPb.DeleteItemRequest{
				Id: itemID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "item deleted"})
	})

	// POST /api/inventory/items/:id/reserve — уменьшить stock
	r.POST("/api/v1/inventory/items/:id/reserve", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		var req struct {
			Quantity int32 `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be >= 1"})
			return
		}

		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.ReserveItem(withUserRole(c.Request.Context(), c), &inventoryPb.ReserveItemRequest{
				Id:       itemID,
				Quantity: req.Quantity,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.ReserveItemResponse)

		c.JSON(http.StatusOK, gin.H{
			"success":         resp.Success,
			"remaining_stock": resp.RemainingStock,
		})
	})

	// DELETE /api/inventory/items/:id/soft — мягкое удаление (author владелец или admin)
	r.DELETE("/api/v1/inventory/items/:id/soft", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		if middleware.Role(c) != RoleAdmin {
			out, err := throughBreaker(inventoryCB, c, func() (any, error) {
				return inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
			})
			if err == circuit.ErrOpen {
				return
			}
			if handleRPCError(c, err) {
				return
			}
			existing := out.(*inventoryPb.ItemResponse)
			if existing.Item == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
				return
			}
			if existing.Item.AuthorId != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "you can only soft-delete your own items"})
				return
			}
		}

		if _, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.SoftDeleteItem(withUserRole(c.Request.Context(), c), &inventoryPb.SoftDeleteItemRequest{Id: itemID})
		}); err == circuit.ErrOpen {
			return
		} else if handleRPCError(c, err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "item soft deleted"})
	})

	// POST /api/inventory/items/:id/restore — восстановить после soft delete (author владелец или admin)
	r.POST("/api/v1/inventory/items/:id/restore", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		userID := middleware.UserID(c)
		itemID := c.Param("id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
			return
		}

		if middleware.Role(c) != RoleAdmin {
			out, err := throughBreaker(inventoryCB, c, func() (any, error) {
				return inventoryClient.SearchItems(c.Request.Context(), &inventoryPb.SearchItemsRequest{
					Filters: map[string]string{
						"id":              itemID,
						"author_id":       userID,
						"include_deleted": "1",
					},
					Limit:  1,
					Offset: 0,
				})
			})
			if err == circuit.ErrOpen {
				return
			}
			if handleRPCError(c, err) {
				return
			}
			found := out.(*inventoryPb.SearchItemsResponse)
			if found.GetTotal() == 0 || len(found.GetItems()) == 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "you can only restore your own items"})
				return
			}
		}

		if _, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.RestoreItem(withUserRole(c.Request.Context(), c), &inventoryPb.RestoreItemRequest{Id: itemID})
		}); err == circuit.ErrOpen {
			return
		} else if handleRPCError(c, err) {
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "item restored"})
	})

	// GET /api/inventory/stats — статистика по товарам (только admin)
	r.GET("/api/v1/inventory/stats", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
		out, err := throughBreaker(inventoryCB, c, func() (any, error) {
			return inventoryClient.GetStats(withUserRole(c.Request.Context(), c), &inventoryPb.EmptyRequest{})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*inventoryPb.StatsResponse)

		byType := resp.GetByType()
		if byType == nil {
			byType = map[string]int64{}
		}
		byAuthor := resp.GetByAuthor()
		if byAuthor == nil {
			byAuthor = map[string]int64{}
		}
		top := resp.GetTopExpensive()
		if top == nil {
			top = []*inventoryPb.TopItem{}
		}

		// Enrich author UUIDs → email (best-effort; keep UUID on miss).
		// Call auth directly — do not use throughBreaker (it writes HTTP on open).
		authorEmails := make(map[string]string, len(byAuthor))
		for authorID := range byAuthor {
			u, uErr := authClient.GetClient().GetUser(c.Request.Context(), &authPb.GetUserRequest{UserId: authorID})
			if uErr != nil || u == nil || u.GetEmail() == "" {
				continue
			}
			authorEmails[authorID] = u.GetEmail()
		}

		topJSON := make([]gin.H, 0, len(top))
		for _, item := range top {
			if item == nil {
				continue
			}
			topJSON = append(topJSON, gin.H{
				"id":        item.GetId(),
				"name":      item.GetName(),
				"price":     item.GetPrice(),
				"author_id": item.GetAuthorId(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"total_items":   resp.GetTotalItems(),
			"total_stock":   resp.GetTotalStock(),
			"by_type":       byType,
			"by_author":     byAuthor,
			"author_emails": authorEmails,
			"top_expensive": topJSON,
		})
	})

	r.GET("/api/v1/leaderboard", func(c *gin.Context) {
		gameID := c.Query("game_id")
		limit := c.Query("limit")
		levelQ := c.Query("level")

		var limitInt int32 = 10
		if limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				limitInt = int32(l)
			}
		}
		var levelInt int32 = 1
		if levelQ != "" {
			if l, err := strconv.Atoi(levelQ); err == nil && l >= 1 {
				levelInt = int32(l)
			}
		}

		out, err := throughBreaker(leaderboardCB, c, func() (any, error) {
			return leaderboardClient.GetTopScores(c.Request.Context(), &leaderboardPb.GetTopScoresRequest{
				GameId: gameID,
				Limit:  limitInt,
				Level:  levelInt,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*leaderboardPb.GetTopScoresResponse)
		c.JSON(http.StatusOK, gin.H{"entries": dto.ScoreEntries(resp.GetEntries())})
	})

	r.POST("/api/v1/auth/update-nickname", middleware.RequireAuth(authClient), func(c *gin.Context) {
		var req struct {
			Nickname string `json:"nickname"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := middleware.UserID(c)
		_, err := throughBreaker(authCB, c, func() (any, error) {
			return authClient.GetClient().UpdateNickname(c.Request.Context(), &authPb.UpdateNicknameRequest{
				UserId:   userID,
				Nickname: req.Nickname,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}

		// Refresh LB nicknames by user_id (all games) via NATS.
		if js != nil {
			payload, _ := json.Marshal(map[string]string{
				"user_id":  userID,
				"nickname": req.Nickname,
			})
			if _, pubErr := js.Publish("event.user.nickname.updated", payload); pubErr != nil {
				log.Printf("⚠️ publish event.user.nickname.updated: %v", pubErr)
			}
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "nickname updated"})
	})

	// /api/game/submit previously trusted the "user_id" field from the request body
	// with no auth check at all, letting anyone submit scores for any account.
	// It now requires a valid token and always uses the authenticated user's id.
	r.POST("/api/v1/game/submit", middleware.RequireAuth(authClient), func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		cacheKey := string(body)
		if cached, ok := scoreCache.Get(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		var req struct {
			UserID    string `json:"user_id"`
			GameID    string `json:"game_id"`
			Level     int32  `json:"level"`
			Score     int32  `json:"score"`
			UserEmail string `json:"user_email"`
			Nickname  string `json:"nickname"`
			Seed      string `json:"seed"`
			BoostID   string `json:"boost_id"`
			Moves     []struct {
				FromX     int32 `json:"fromX"`
				FromY     int32 `json:"fromY"`
				ToX       int32 `json:"toX"`
				ToY       int32 `json:"toY"`
				Timestamp int32 `json:"timestamp"`
			} `json:"moves"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		moves := make([]*gamePb.Move, len(req.Moves))
		for i, m := range req.Moves {
			moves[i] = &gamePb.Move{
				FromX:     m.FromX,
				FromY:     m.FromY,
				ToX:       m.ToX,
				ToY:       m.ToY,
				Timestamp: m.Timestamp,
			}
		}

		out, err := throughBreaker(gameCB, c, func() (any, error) {
			return gameClient.SubmitScore(c.Request.Context(), &gamePb.SubmitScoreRequest{
				UserId:    middleware.UserID(c),
				GameId:    req.GameID,
				Level:     req.Level,
				Score:     req.Score,
				UserEmail: req.UserEmail,
				Nickname:  req.Nickname,
				Seed:      req.Seed,
				Moves:     moves,
				BoostId:   req.BoostID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*gamePb.SubmitScoreResponse)

		// Explicit map keeps ranked=false (proto3 omitempty would drop it).
		respBody := gin.H{
			"success":        resp.GetSuccess(),
			"new_highscore":  resp.GetNewHighscore(),
			"rank":           resp.GetRank(),
			"message":        resp.GetMessage(),
			"lamps_earned":   resp.GetLampsEarned(),
			"tickets_earned": resp.GetTicketsEarned(),
			"ranked":         resp.GetRanked(),
		}
		respJSON, _ := json.Marshal(respBody)
		scoreCache.Set(cacheKey, respJSON)
		c.JSON(http.StatusOK, respBody)
	})

	// Tamagotchi daily gift: +1000 tickets (idempotent per UTC day) + FE still adds care points.
	r.POST("/api/v1/game/companion/daily-gift", middleware.RequireAuth(authClient), func(c *gin.Context) {
		userID := middleware.UserID(c)
		day := time.Now().UTC().Format("2006-01-02")
		ref := "companion-daily-" + userID + "-" + day
		out, err := throughBreaker(billingCB, c, func() (any, error) {
			return billingClient.AddCurrency(c.Request.Context(), &billingPb.AddCurrencyRequest{
				UserId:      userID,
				Currency:    billingPb.CurrencyType_TICKETS,
				Amount:      1000,
				Reason:      "companion_daily_gift",
				ReferenceId: ref,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if err != nil {
			// Unique reference_id → already claimed today.
			msg := err.Error()
			if strings.Contains(strings.ToLower(msg), "duplicate") || strings.Contains(msg, "unique") {
				c.JSON(http.StatusOK, gin.H{
					"tickets_granted": 0,
					"already_claimed": true,
					"message":         "already claimed today",
				})
				return
			}
			if handleRPCError(c, err) {
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		resp := out.(*billingPb.AddCurrencyResponse)
		c.JSON(http.StatusOK, gin.H{
			"tickets_granted": 1000,
			"already_claimed": false,
			"new_balance":     resp.GetNewBalance(),
			"message":         resp.GetMessage(),
		})
	})

	r.POST("/api/v1/game/boost/start", middleware.RequireAuth(authClient), func(c *gin.Context) {
		var req struct {
			GameID string `json:"game_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.GameID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
			return
		}
		out, err := throughBreaker(gameCB, c, func() (any, error) {
			return gameClient.StartBoost(c.Request.Context(), &gamePb.StartBoostRequest{
				UserId: middleware.UserID(c),
				GameId: req.GameID,
			})
		})
		if err == circuit.ErrOpen {
			return
		}
		if handleRPCError(c, err) {
			return
		}
		resp := out.(*gamePb.StartBoostResponse)
		c.JSON(http.StatusOK, gin.H{
			"boost_id":    resp.GetBoostId(),
			"boosted":     resp.GetBoosted(),
			"cost":        resp.GetCost(),
			"new_balance": resp.GetNewBalance(),
			"message":     resp.GetMessage(),
		})
	})

	srv := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 Gateway listening on :%s", cfg.Port)
	log.Printf("   WebSocket endpoint: ws://localhost:%s/ws/leaderboard", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gateway gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	authClient.Close()
	gameConn.Close()
	profileConn.Close()
	leaderboardConn.Close()
	billingConn.Close()
	if nc != nil {
		nc.Drain()
	}

	log.Println("Gateway stopped gracefully")
}
--- services/gateway/internal/app/upload.go ---
package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/client"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/config"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/middleware"
)

var allowedUploadExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func ensureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// registerUploadRoutes serves local inventory images and accepts multipart uploads.
// Any authenticated user may upload (avatars + author/admin catalog images).
func registerUploadRoutes(r *gin.Engine, cfg *config.Config, authClient *client.AuthClient) {
	if err := ensureUploadDir(cfg.UploadDir); err != nil {
		log.Printf("⚠️ upload dir %s: %v (uploads disabled)", cfg.UploadDir, err)
		return
	}
	log.Printf("📁 Uploads: dir=%s max=%d bytes", cfg.UploadDir, cfg.UploadMaxBytes)

	r.Static("/uploads", cfg.UploadDir)

	r.POST("/api/v1/uploads", middleware.RequireAuth(authClient), func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.UploadMaxBytes+1024)
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required (multipart field \"file\")"})
			return
		}
		if file.Size <= 0 || file.Size > cfg.UploadMaxBytes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file size must be 1..%d bytes", cfg.UploadMaxBytes),
			})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if _, ok := allowedUploadExt[ext]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only .jpg, .jpeg, .png, .webp allowed"})
			return
		}
		// Normalize jpeg → .jpg
		if ext == ".jpeg" {
			ext = ".jpg"
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read upload"})
			return
		}
		defer src.Close()

		head := make([]byte, 512)
		n, _ := io.ReadFull(src, head)
		detected := http.DetectContentType(head[:n])
		if !strings.HasPrefix(detected, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is not an image"})
			return
		}

		name := uuid.New().String() + ext
		destPath := filepath.Join(cfg.UploadDir, name)
		dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot store upload"})
			return
		}
		defer dst.Close()

		if _, err := dst.Write(head[:n]); err != nil {
			_ = os.Remove(destPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
			return
		}
		if _, err := io.Copy(dst, src); err != nil {
			_ = os.Remove(destPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
			return
		}

		url := "/uploads/" + name
		c.JSON(http.StatusOK, gin.H{"url": url, "size": file.Size})
	})
}
===== middleware =====
--- services/gateway/internal/middleware/auth.go ---
// services/gateway/internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

const (
	CtxUserID = "user_id"
	CtxEmail  = "email"
	CtxRole   = "role"
)

// ExtractBearerToken pulls the raw JWT out of an "Authorization: Bearer <token>" header.
func ExtractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1]), true
	}
	// Some clients send the raw token without the "Bearer " prefix.
	return strings.TrimSpace(header), true
}

// AuthClient matches client.AuthClient's gRPC surface needed here (avoids an import cycle).
type AuthClient interface {
	GetClient() authPb.AuthServiceClient
}

// RequireAuth validates the bearer token against the real Auth service (ValidateToken RPC,
// which itself checks the Redis session) and stores user_id/email/role in the gin context.
// This replaces the old local base64 JWT decode, which trusted the token's claims blindly
// and never checked for revocation (logout) or the source of truth in Postgres/Redis.
func RequireAuth(authClient AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "session store unavailable"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth service unavailable"})
			return
		}
		if !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, resp.UserId)
		c.Set(CtxEmail, resp.Email)
		c.Set(CtxRole, resp.Role)
		c.Next()
	}
}

// RequireRole must run after RequireAuth. It rejects the request unless the
// authenticated user's role is one of the allowed roles.
func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		roleStr, _ := role.(string)
		if !allowedSet[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// UserID is a small helper for handlers to read the authenticated user id
// set by RequireAuth, instead of re-parsing the token.
func UserID(c *gin.Context) string {
	v, _ := c.Get(CtxUserID)
	s, _ := v.(string)
	return s
}

// Role is a small helper for handlers to read the authenticated user's role.
func Role(c *gin.Context) string {
	v, _ := c.Get(CtxRole)
	s, _ := v.(string)
	return s
}

// Email is a small helper for handlers to read the authenticated user's email.
func Email(c *gin.Context) string {
	v, _ := c.Get(CtxEmail)
	s, _ := v.(string)
	return s
}
--- services/gateway/internal/middleware/auth_role_test.go ---
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRole_ForbiddenForUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(CtxRole, "user")
		c.Next()
	}, RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestRequireRole_AllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(CtxRole, "admin")
		c.Next()
	}, RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
--- services/gateway/internal/middleware/docs.go ---
package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocsGate hides OpenAPI and Swagger.
// An empty password answers 404, so a public host without DOCS_PASSWORD does not serve the spec.
// A configured password uses HTTP Basic. The browser prompts; the secret is not put in the URL.
func DocsGate(user, password string) gin.HandlerFunc {
	userB := []byte(user)
	passB := []byte(password)
	return func(c *gin.Context) {
		if len(passB) == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		gotUser, gotPass, ok := c.Request.BasicAuth()
		match := subtle.ConstantTimeCompare([]byte(gotUser), userB) &
			subtle.ConstantTimeCompare([]byte(gotPass), passB)
		if !ok || match != 1 {
			c.Header("WWW-Authenticate", `Basic realm="Event Horizon docs", charset="UTF-8"`)
			c.Header("Cache-Control", "no-store")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Header("Cache-Control", "private, no-store")
		c.Next()
	}
}
--- services/gateway/internal/middleware/docs_test.go ---
package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocsGate_HiddenWithoutPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", ""), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestDocsGate_RejectsWrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", "s3cret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	req.SetBasicAuth("docs", "nope")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("missing WWW-Authenticate")
	}
}

func TestDocsGate_AllowsPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", "s3cret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	req.SetBasicAuth("docs", "s3cret")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestRejectUnsafeRequest_Traversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RejectUnsafeRequest())
	r.GET("/files/*rest", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/ok", nil)
	req.URL.RawPath = "/files/../secret"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestSecurityHeaders_CSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/docs", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") {
		t.Fatalf("api csp: %s", w.Header().Get("Content-Security-Policy"))
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src 'self'") {
		t.Fatalf("docs csp: %s", w.Header().Get("Content-Security-Policy"))
	}
}
--- services/gateway/internal/middleware/ratelimit.go ---
// services/gateway/internal/middleware/ratelimit.go
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
)

// RateLimitMiddleware enforces per-route and global (~100 req/s) rate limits.
// Subject is user_id when RequireAuth already ran; otherwise a hash of the Bearer
// token (same session → same bucket) or the client IP.
func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		subject := rateLimitSubject(c)

		var allowed bool

		switch {
		case path == "/api/v1/game/submit" && method == "POST":
			allowed = limiter.AllowSubmit(subject)

		case path == "/api/v1/auth/login" && method == "POST":
			allowed = limiter.AllowLogin(c.ClientIP())

		case path == "/ws/leaderboard":
			allowed = limiter.AllowWebSocket(c.ClientIP())

		case path == "/docs" || path == "/openapi.yaml" || strings.HasPrefix(path, "/docs/"):
			allowed = limiter.AllowDocs(c.ClientIP())

		case path == "/health" || path == "/ready" || path == "/metrics":
			allowed = true

		default:
			allowed = limiter.AllowGlobal(subject)
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please try again later.",
				"retry_after": 1,
			})
			return
		}

		c.Next()
	}
}

func rateLimitSubject(c *gin.Context) string {
	if uid := UserID(c); uid != "" {
		return "u:" + uid
	}
	if tok, ok := ExtractBearerToken(c.GetHeader("Authorization")); ok && tok != "" {
		sum := sha256.Sum256([]byte(tok))
		return "t:" + hex.EncodeToString(sum[:8])
	}
	return "ip:" + c.ClientIP()
}
--- services/gateway/internal/middleware/security.go ---
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxJSONBody = 1 << 20 // 1 MiB. Uploads use their own limit.

// SecurityHeaders sets the baseline headers that do not depend on cookies.
// The API is Bearer-token, so a CSRF token is not added here.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		h.Set("Content-Security-Policy", contentSecurityPolicy(c.Request.URL.Path))
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func contentSecurityPolicy(path string) string {
	if path == "/docs" || strings.HasPrefix(path, "/docs/") {
		return "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'"
	}
	return "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"
}

// RejectUnsafeRequest drops path traversal and methods this API does not use.
func RejectUnsafeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Request.URL.RawPath
		if raw == "" {
			raw = c.Request.URL.Path
		}
		if strings.Contains(raw, "..") || strings.ContainsRune(raw, 0) || strings.Contains(raw, "\\") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		default:
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		c.Next()
	}
}

// LimitJSONBody caps ordinary request bodies. Upload routes keep UploadMaxBytes.
func LimitJSONBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || strings.Contains(c.Request.URL.Path, "/uploads") {
			c.Next()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBody)
		c.Next()
	}
}
===== ratelimit =====
--- services/gateway/internal/ratelimit/limiter.go ---
// services/gateway/internal/ratelimit/limiter.go
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Default global API budget (~100 req/s per user or IP).
const (
	GlobalPerUserLimit = 100
	GlobalWindow       = time.Second
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// Allow проверяет, не превышен ли лимит (sliding window алгоритм)
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := rl.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := cmds[2].(*redis.IntCmd).Val()
	return count <= int64(limit), nil
}

// AllowGlobal ~100 req/s per authenticated user (or IP when anonymous).
func (rl *RateLimiter) AllowGlobal(id string) bool {
	key := fmt.Sprintf("rl:global:%s", id)
	allowed, err := rl.Allow(context.Background(), key, GlobalPerUserLimit, GlobalWindow)
	if err != nil {
		// Fail open on Redis blips — do not take the API down with the limiter.
		return true
	}
	return allowed
}

// AllowSubmit для POST /api/v1/game/submit (10 запросов в секунду)
func (rl *RateLimiter) AllowSubmit(userID string) bool {
	key := fmt.Sprintf("rl:submit:%s", userID)
	allowed, _ := rl.Allow(context.Background(), key, 10, time.Second)
	return allowed
}

// AllowLogin для POST /api/v1/auth/login (5 запросов в секунду с одного IP)
func (rl *RateLimiter) AllowLogin(ip string) bool {
	key := fmt.Sprintf("rl:login:%s", ip)
	allowed, _ := rl.Allow(context.Background(), key, 5, time.Second)
	return allowed
}

// AllowDocs limits guesses against the documentation password (20 requests per minute per IP).
func (rl *RateLimiter) AllowDocs(ip string) bool {
	key := fmt.Sprintf("rl:docs:%s", ip)
	allowed, _ := rl.Allow(context.Background(), key, 20, time.Minute)
	return allowed
}

// AllowWebSocket для WebSocket соединений (100 соединений в минуту с одного IP)
func (rl *RateLimiter) AllowWebSocket(ip string) bool {
	key := fmt.Sprintf("rl:ws:%s", ip)
	allowed, _ := rl.Allow(context.Background(), key, 100, time.Minute)
	return allowed
}
===== circuit =====
--- services/gateway/internal/circuit/breaker.go ---
package circuit

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the breaker is open (fail-fast).
var ErrOpen = errors.New("circuit breaker open")

// ErrBulkheadFull is returned when MaxConcurrent in-flight calls are already running.
var ErrBulkheadFull = errors.New("bulkhead full")

// Settings mirrors sony/gobreaker-style knobs used by the gateway.
type Settings struct {
	Name          string
	MaxRequests   uint32        // half-open probes
	Timeout       time.Duration // how long to stay open
	Interval      time.Duration // clear consecutive failures in closed state
	MaxConcurrent uint32        // bulkhead slots; 0 = unlimited
	ReadyToTrip   func(counts Counts) bool
}

type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

// Breaker is a minimal circuit breaker + optional bulkhead (no external deps).
type Breaker struct {
	name          string
	maxRequests   uint32
	timeout       time.Duration
	interval      time.Duration
	readyToTrip   func(Counts) bool
	sem           chan struct{} // bulkhead; nil = unlimited

	mu         sync.Mutex
	state      state
	generation uint64
	counts     Counts
	expiry     time.Time
}

func New(s Settings) *Breaker {
	if s.Timeout <= 0 {
		s.Timeout = 10 * time.Second
	}
	if s.MaxRequests == 0 {
		s.MaxRequests = 3
	}
	if s.ReadyToTrip == nil {
		s.ReadyToTrip = func(c Counts) bool { return c.ConsecutiveFailures >= 5 }
	}
	b := &Breaker{
		name:        s.Name,
		maxRequests: s.MaxRequests,
		timeout:     s.Timeout,
		interval:    s.Interval,
		readyToTrip: s.ReadyToTrip,
	}
	if s.MaxConcurrent > 0 {
		b.sem = make(chan struct{}, s.MaxConcurrent)
	}
	return b
}

func (b *Breaker) Name() string { return b.name }

func (b *Breaker) Execute(fn func() (any, error)) (any, error) {
	if b.sem != nil {
		select {
		case b.sem <- struct{}{}:
			defer func() { <-b.sem }()
		default:
			return nil, ErrBulkheadFull
		}
	}

	generation, err := b.beforeRequest()
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			b.afterRequest(generation, false)
			panic(r)
		}
	}()
	result, err := fn()
	b.afterRequest(generation, err == nil)
	return result, err
}

func (b *Breaker) beforeRequest() (uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	switch b.state {
	case stateOpen:
		if now.After(b.expiry) {
			b.toHalfOpen()
			return b.generation, nil
		}
		return 0, ErrOpen
	case stateHalfOpen:
		if b.counts.Requests >= b.maxRequests {
			return 0, ErrOpen
		}
	case stateClosed:
		if !b.expiry.IsZero() && now.After(b.expiry) {
			b.counts = Counts{}
			b.setIntervalExpiry(now)
		}
	}
	b.counts.Requests++
	return b.generation, nil
}

func (b *Breaker) afterRequest(generation uint64, success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.generation != generation {
		return
	}
	if success {
		b.onSuccess()
	} else {
		b.onFailure()
	}
}

func (b *Breaker) onSuccess() {
	b.counts.TotalSuccesses++
	b.counts.ConsecutiveSuccesses++
	b.counts.ConsecutiveFailures = 0
	if b.state == stateHalfOpen && b.counts.ConsecutiveSuccesses >= b.maxRequests {
		b.toClosed()
	}
}

func (b *Breaker) onFailure() {
	b.counts.TotalFailures++
	b.counts.ConsecutiveFailures++
	b.counts.ConsecutiveSuccesses = 0
	switch b.state {
	case stateHalfOpen:
		b.toOpen()
	case stateClosed:
		if b.readyToTrip(b.counts) {
			b.toOpen()
		}
	}
}

func (b *Breaker) toClosed() {
	b.state = stateClosed
	b.counts = Counts{}
	b.setIntervalExpiry(time.Now())
}

func (b *Breaker) toOpen() {
	b.state = stateOpen
	b.generation++
	b.counts = Counts{}
	b.expiry = time.Now().Add(b.timeout)
}

func (b *Breaker) toHalfOpen() {
	b.state = stateHalfOpen
	b.generation++
	b.counts = Counts{}
}

func (b *Breaker) setIntervalExpiry(now time.Time) {
	if b.interval > 0 {
		b.expiry = now.Add(b.interval)
	} else {
		b.expiry = time.Time{}
	}
}
--- services/gateway/internal/circuit/breaker_test.go ---
package circuit

import (
	"errors"
	"testing"
	"time"
)

func TestBreakerOpensAfterFailures(t *testing.T) {
	b := New(Settings{
		Name:        "billing",
		MaxRequests: 3,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool { return c.ConsecutiveFailures >= 2 },
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	_, _ = b.Execute(fail)
	_, _ = b.Execute(fail)
	_, err := b.Execute(fail)
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("want ErrOpen, got %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	ok := false
	_, err = b.Execute(func() (any, error) { ok = true; return "ok", nil })
	if err != nil || !ok {
		t.Fatalf("half-open probe failed: %v", err)
	}
}

func TestNew_DefaultSettings(t *testing.T) {
	b := New(Settings{Name: "shop"})
	if b.Name() != "shop" {
		t.Fatalf("name=%q", b.Name())
	}
	result, err := b.Execute(func() (any, error) { return "ok", nil })
	if err != nil || result != "ok" {
		t.Fatalf("execute failed: %v", err)
	}
}

func TestBreaker_ClosedSuccessResetsFailures(t *testing.T) {
	trips := 0
	b := New(Settings{
		Name:        "inventory",
		MaxRequests: 2,
		Timeout:     10 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool {
			trips++
			return c.ConsecutiveFailures >= 3
		},
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	ok := func() (any, error) { return "ok", nil }
	_, _ = b.Execute(fail)
	_, _ = b.Execute(fail)
	_, _ = b.Execute(ok)
	_, err := b.Execute(fail)
	if err == nil {
		t.Fatal("expected failure")
	}
	if trips == 0 {
		t.Fatal("ReadyToTrip never called")
	}
}

func TestBreaker_BulkheadRejectsWhenFull(t *testing.T) {
	b := New(Settings{
		Name:          "shop",
		MaxConcurrent: 1,
	})
	started := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_, _ = b.Execute(func() (any, error) {
			close(started)
			<-release
			return nil, nil
		})
	}()
	<-started
	_, err := b.Execute(func() (any, error) { return "nope", nil })
	if !errors.Is(err, ErrBulkheadFull) {
		t.Fatalf("want ErrBulkheadFull, got %v", err)
	}
	close(release)
}

func TestBreaker_HalfOpenProbeLimit(t *testing.T) {
	b := New(Settings{
		Name:        "auth",
		MaxRequests: 1,
		Timeout:     50 * time.Millisecond,
		ReadyToTrip: func(c Counts) bool { return c.ConsecutiveFailures >= 1 },
	})
	fail := func() (any, error) { return nil, errors.New("down") }
	_, _ = b.Execute(fail) // trip open
	time.Sleep(60 * time.Millisecond)
	_, _ = b.Execute(fail) // half-open probe fails, back to open
	_, err := b.Execute(fail)
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("want ErrOpen while breaker open, got %v", err)
	}
}
===== cache =====
--- services/gateway/internal/cache/response_cache.go ---
package cache

import (
	"sync"
	"time"
)

// ResponseCache stores short-lived JSON HTTP responses (whoami, profile, leaderboard submit).
type ResponseCache struct {
	cache map[string]entry
	mu    sync.RWMutex
	ttl   time.Duration
}

type entry struct {
	response  []byte
	expiresAt time.Time
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	c := &ResponseCache{
		cache: make(map[string]entry),
		ttl:   ttl,
	}
	go c.cleanup()
	return c
}

// NewScoreCache is an alias for game submit score caching (2s default at call site).
func NewScoreCache(ttl time.Duration) *ResponseCache {
	return NewResponseCache(ttl)
}

func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.cache[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.response, true
}

func (c *ResponseCache) Set(key string, response []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = entry{
		response:  response,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ResponseCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, e := range c.cache {
			if now.After(e.expiresAt) {
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
	}
}
--- services/gateway/internal/cache/score_cache.go ---
package cache

// ScoreCache is kept for backward compatibility; use ResponseCache.
type ScoreCache = ResponseCache
===== client =====
--- services/gateway/internal/client/auth_client.go ---
package client

import (
	"fmt"
	"log"

	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
	"google.golang.org/grpc"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	log.Printf("Connected to Auth gRPC server at %s", addr)

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) GetClient() pb.AuthServiceClient {
	return c.client
}
--- services/gateway/internal/client/dial.go ---
package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial opens a gRPC client connection with insecure transport and unary retry+jitter.
func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryRetryInterceptor(DefaultRetryConfig())),
	}
	return grpc.NewClient(addr, append(base, opts...)...)
}
--- services/gateway/internal/client/retry.go ---
package client

import (
	"context"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig controls unary client retries (exponential backoff + full jitter).
type RetryConfig struct {
	MaxAttempts int           // total attempts including the first call
	BaseDelay   time.Duration // backoff base; wait = rand(0, base*2^i)
	MaxDelay    time.Duration // cap on exponential term before jitter
}

// DefaultRetryConfig is tuned for short gateway→service hops.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   40 * time.Millisecond,
		MaxDelay:    400 * time.Millisecond,
	}
}

// UnaryRetryInterceptor retries transient gRPC failures with full jitter.
// Does not retry permanent codes (InvalidArgument, NotFound, PermissionDenied, …).
func UnaryRetryInterceptor(cfg RetryConfig) grpc.UnaryClientInterceptor {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 40 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 400 * time.Millisecond
	}

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var last error
		for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
			last = invoker(ctx, method, req, reply, cc, opts...)
			if last == nil {
				return nil
			}
			if !isRetryable(last) || attempt == cfg.MaxAttempts-1 {
				return last
			}
			if err := waitBackoff(ctx, cfg, attempt); err != nil {
				return last
			}
		}
		return last
	}
}

func isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

func waitBackoff(ctx context.Context, cfg RetryConfig, attempt int) error {
	exp := cfg.BaseDelay * time.Duration(1<<attempt)
	if exp > cfg.MaxDelay {
		exp = cfg.MaxDelay
	}
	wait := time.Duration(rand.Int63n(int64(exp) + 1)) // full jitter
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
--- services/gateway/internal/client/retry_test.go ---
package client

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryRetryInterceptor_retriesUnavailable(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		n := calls.Add(1)
		if n < 3 {
			return status.Error(codes.Unavailable, "down")
		}
		return nil
	}

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestUnaryRetryInterceptor_noRetryOnInvalidArgument(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls.Add(1)
		return status.Error(codes.InvalidArgument, "bad")
	}

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 4,
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	})

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
}

func TestUnaryRetryInterceptor_respectsContextCancel(t *testing.T) {
	var calls atomic.Int32
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls.Add(1)
		return status.Error(codes.Unavailable, "down")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	interceptor := UnaryRetryInterceptor(RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
	})

	err := interceptor(ctx, "/svc/Method", nil, nil, nil, invoker)
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("want Unavailable (last RPC err), got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1 (no further attempts after cancel)", calls.Load())
	}
}

func TestIsRetryable(t *testing.T) {
	if !isRetryable(status.Error(codes.Unavailable, "x")) {
		t.Fatal("Unavailable should retry")
	}
	if !isRetryable(status.Error(codes.ResourceExhausted, "x")) {
		t.Fatal("ResourceExhausted should retry")
	}
	if isRetryable(status.Error(codes.PermissionDenied, "x")) {
		t.Fatal("PermissionDenied must not retry")
	}
}
===== dto =====
--- services/gateway/internal/dto/inventory.go ---
// Package dto maps gRPC/proto responses to stable public JSON (gin.H).
// Proto3 omitempty drops zeros and empty slices; public HTTP must keep them.
package dto

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/structpb"

	inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

// InventoryItem maps a proto Item to gin.H with explicit zeros and [].
// Nil item → nil (caller wraps as {"item": null} or 404).
func InventoryItem(item *inventoryPb.Item) gin.H {
	if item == nil {
		return nil
	}
	images := item.GetImages()
	if images == nil {
		images = []string{}
	}
	attrs := attributesMap(item.GetAttributes())
	deleted := false
	if v, ok := attrs["_deleted"].(bool); ok && v {
		deleted = true
	}
	h := gin.H{
		"id":          item.GetId(),
		"author_id":   item.GetAuthorId(),
		"type":        item.GetType(),
		"name":        item.GetName(),
		"description": item.GetDescription(),
		"price":       item.GetPrice(),
		"stock":       item.GetStock(),
		"attributes":  attrs,
		"images":      images,
		"created_at":  item.GetCreatedAt(),
		"updated_at":  item.GetUpdatedAt(),
		"version":     item.GetVersion(),
		"deleted":     deleted,
	}
	if deletedAt, ok := attrs["_deleted_at"].(string); ok && deletedAt != "" {
		h["deleted_at"] = deletedAt
	}
	return h
}

// InventoryItemResponse wraps a single item the way FE expects: {"item": {...}}.
func InventoryItemResponse(resp *inventoryPb.ItemResponse) gin.H {
	if resp == nil {
		return gin.H{"item": nil}
	}
	return gin.H{"item": InventoryItem(resp.GetItem())}
}

// InventoryItems maps a slice; nil → empty [] for JSON.
func InventoryItems(items []*inventoryPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if h := InventoryItem(it); h != nil {
			out = append(out, h)
		}
	}
	return out
}

// catalogAttrOmit are bulky / internal keys not needed for grid filters.
var catalogAttrOmit = map[string]struct{}{
	"card_text":       {},
	"flavor_text":     {},
	"noiz_review":     {},
	"idempotency_key": {},
	"market_rub":      {},
}

// InventoryItemsCatalog is a smaller list DTO for SearchItems:
// keeps filterable attributes + images, drops card/flavor text.
// Detail GET still uses InventoryItem (full).
func InventoryItemsCatalog(items []*inventoryPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		h := InventoryItem(it)
		if h == nil {
			continue
		}
		if attrs, ok := h["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
			slim := make(map[string]interface{}, len(attrs))
			for k, v := range attrs {
				if _, drop := catalogAttrOmit[k]; drop {
					continue
				}
				slim[k] = v
			}
			h["attributes"] = slim
		}
		out = append(out, h)
	}
	return out
}

func attributesMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return s.AsMap()
}
--- services/gateway/internal/dto/inventory_test.go ---
package dto

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func TestInventoryItemKeepsZeros(t *testing.T) {
	item := &inventoryPb.Item{
		Id:     "a",
		Name:   "free",
		Price:  0,
		Stock:  0,
		Images: nil,
	}
	h := InventoryItem(item)
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["price"]; !ok {
		t.Fatalf("price omitted: %s", b)
	}
	if _, ok := m["stock"]; !ok {
		t.Fatalf("stock omitted: %s", b)
	}
	imgs, ok := m["images"].([]any)
	if !ok {
		t.Fatalf("images not array: %s", b)
	}
	if len(imgs) != 0 {
		t.Fatalf("want empty images, got %v", imgs)
	}
}

func TestInventoryItemsNilIsEmpty(t *testing.T) {
	out := InventoryItems(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}

func TestInventoryItemsCatalogOmitsHeavyAttrs(t *testing.T) {
	attrs, err := structpb.NewStruct(map[string]any{
		"element":         "woods",
		"card_text":       "long rules text",
		"flavor_text":     "flavor",
		"idempotency_key": "k",
		"market_rub":      100,
		"hp":              float64(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	item := &inventoryPb.Item{
		Id:         "a",
		Name:       "x",
		Attributes: attrs,
	}
	out := InventoryItemsCatalog([]*inventoryPb.Item{item})
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
	am, ok := out[0]["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("attrs type %T", out[0]["attributes"])
	}
	if _, ok := am["card_text"]; ok {
		t.Fatal("card_text should be omitted")
	}
	if _, ok := am["flavor_text"]; ok {
		t.Fatal("flavor_text should be omitted")
	}
	if _, ok := am["idempotency_key"]; ok {
		t.Fatal("idempotency_key should be omitted")
	}
	if _, ok := am["market_rub"]; ok {
		t.Fatal("market_rub should be omitted")
	}
	if am["element"] != "woods" {
		t.Fatalf("element=%v", am["element"])
	}
	if am["hp"] != float64(5) && am["hp"] != 5 {
		// structpb may decode numbers as float64
		if v, ok := am["hp"].(float64); !ok || v != 5 {
			t.Fatalf("hp=%v (%T)", am["hp"], am["hp"])
		}
	}
}
--- services/gateway/internal/dto/shop.go ---
package dto

import (
	"github.com/gin-gonic/gin"

	shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

// ShopItem maps a shop proto Item to gin.H with explicit zeros / empty strings.
// Nil item → nil.
func ShopItem(item *shopPb.Item) gin.H {
	if item == nil {
		return nil
	}
	return gin.H{
		"id":               item.GetId(),
		"name":             item.GetName(),
		"description":      item.GetDescription(),
		"price":            item.GetPrice(),
		"category":         item.GetCategory(),
		"game_id":          item.GetGameId(),
		"image_url":        item.GetImageUrl(),
		"available":        item.GetAvailable(),
		"owned":            item.GetOwned(),
		"purchased_at":     item.GetPurchasedAt(),
		"purchase_price":   item.GetPurchasePrice(),
		"purchase_id":      item.GetPurchaseId(),
		"refundable_until": item.GetRefundableUntil(),
		"fulfilled_at":     item.GetFulfilledAt(),
		"can_cancel":       item.GetCanCancel(),
	}
}

// ShopItems maps a slice; nil → empty [] for JSON.
func ShopItems(items []*shopPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if h := ShopItem(it); h != nil {
			out = append(out, h)
		}
	}
	return out
}

// ShopItemsCatalog is a smaller list DTO for GET /api/shop/items:
// keeps grid fields, drops description and purchase-only fields.
// Shop inventory / cancel paths should use ShopItems (full) when wired.
func ShopItemsCatalog(items []*shopPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, gin.H{
			"id":        it.GetId(),
			"name":      it.GetName(),
			"price":     it.GetPrice(),
			"category":  it.GetCategory(),
			"game_id":   it.GetGameId(),
			"image_url": it.GetImageUrl(),
			"available": it.GetAvailable(),
			"owned":     it.GetOwned(),
		})
	}
	return out
}
--- services/gateway/internal/dto/shop_test.go ---
package dto

import (
	"encoding/json"
	"testing"

	shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

func TestShopItemKeepsZeros(t *testing.T) {
	item := &shopPb.Item{
		Id:            "a",
		Name:          "free",
		Price:         0,
		Available:     false,
		Owned:         false,
		PurchasePrice: 0,
	}
	h := ShopItem(item)
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"price", "available", "owned", "description", "purchase_price", "purchased_at", "purchase_id"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("%s omitted: %s", key, b)
		}
	}
}

func TestShopItemsNilIsEmpty(t *testing.T) {
	out := ShopItems(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}

func TestShopItemsCatalogOmitsHeavyFields(t *testing.T) {
	item := &shopPb.Item{
		Id:            "a",
		Name:          "skin",
		Description:   "long description text",
		Price:         0,
		Category:      "game_skin",
		GameId:        "flappy",
		ImageUrl:      "",
		Available:     true,
		Owned:         false,
		PurchasedAt:   "2026-01-01T00:00:00Z",
		PurchasePrice: 100,
		PurchaseId:    "p1",
	}
	out := ShopItemsCatalog([]*shopPb.Item{item})
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
	b, err := json.Marshal(out[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"description", "purchased_at", "purchase_price", "purchase_id"} {
		if _, ok := m[key]; ok {
			t.Fatalf("%s should be omitted: %s", key, b)
		}
	}
	for _, key := range []string{"id", "name", "price", "category", "game_id", "image_url", "available", "owned"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("%s missing: %s", key, b)
		}
	}
	if m["price"].(float64) != 0 {
		t.Fatalf("price zero dropped: %s", b)
	}
	if m["owned"].(bool) != false {
		t.Fatalf("owned false dropped: %s", b)
	}
}

func TestShopItemsCatalogNilIsEmpty(t *testing.T) {
	out := ShopItemsCatalog(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}
--- services/gateway/internal/dto/twins.go ---
package dto

import (
	"github.com/gin-gonic/gin"

	analyticsPb "github.com/Eastwesser/event-horizon/services/analytics/proto"
	authorsPb "github.com/Eastwesser/event-horizon/services/authors/proto"
	historyPb "github.com/Eastwesser/event-horizon/services/history/proto"
	leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
	notificationPb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

// Author maps authors proto with explicit zeros (active/timestamps never omit).
func Author(a *authorsPb.Author) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":               a.GetId(),
		"user_id":          a.GetUserId(),
		"display_name":     a.GetDisplayName(),
		"bio":              a.GetBio(),
		"avatar_url":       a.GetAvatarUrl(),
		"portfolio":        a.GetPortfolio(),
		"active":           a.GetActive(),
		"created_at_unix":  a.GetCreatedAtUnix(),
		"updated_at_unix":  a.GetUpdatedAtUnix(),
		"verified_at_unix": a.GetVerifiedAtUnix(),
	}
}

func Authors(list []*authorsPb.Author) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		if h := Author(a); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func AuthorApplication(a *authorsPb.AuthorApplication) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":               a.GetId(),
		"user_id":          a.GetUserId(),
		"status":           a.GetStatus(),
		"display_name":     a.GetDisplayName(),
		"portfolio":        a.GetPortfolio(),
		"motivation":       a.GetMotivation(),
		"contact_email":    a.GetContactEmail(),
		"created_at_unix":  a.GetCreatedAtUnix(),
		"reviewed_at_unix": a.GetReviewedAtUnix(),
		"reviewed_by":      a.GetReviewedBy(),
		"reviewer_note":    a.GetReviewerNote(),
	}
}

func AuthorApplications(list []*authorsPb.AuthorApplication) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		if h := AuthorApplication(a); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func HistoryEvent(e *historyPb.HistoryEvent) gin.H {
	if e == nil {
		return nil
	}
	return gin.H{
		"id":              e.GetId(),
		"user_id":         e.GetUserId(),
		"event_type":      e.GetEventType(),
		"payload_json":    e.GetPayloadJson(),
		"created_at_unix": e.GetCreatedAtUnix(),
	}
}

func HistoryEvents(list []*historyPb.HistoryEvent) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		if h := HistoryEvent(e); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func Notification(n *notificationPb.Notification) gin.H {
	if n == nil {
		return nil
	}
	return gin.H{
		"id":              n.GetId(),
		"user_id":         n.GetUserId(),
		"title":           n.GetTitle(),
		"body":            n.GetBody(),
		"link":            n.GetLink(),
		"created_at_unix": n.GetCreatedAtUnix(),
		"read_at_unix":    n.GetReadAtUnix(),
	}
}

func Notifications(list []*notificationPb.Notification) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, n := range list {
		if h := Notification(n); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func DayCount(d *analyticsPb.DayCount) gin.H {
	if d == nil {
		return nil
	}
	return gin.H{
		"day":   d.GetDay(),
		"count": d.GetCount(),
	}
}

func DayCounts(list []*analyticsPb.DayCount) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, d := range list {
		if h := DayCount(d); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func RetentionPoint(p *analyticsPb.RetentionPoint) gin.H {
	if p == nil {
		return nil
	}
	return gin.H{
		"day_n": p.GetDayN(),
		"rate":  p.GetRate(),
	}
}

func RetentionPoints(list []*analyticsPb.RetentionPoint) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, p := range list {
		if h := RetentionPoint(p); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func ScoreEntry(e *leaderboardPb.ScoreEntry) gin.H {
	if e == nil {
		return nil
	}
	level := e.GetLevel()
	if level < 1 {
		level = 1
	}
	return gin.H{
		"rank":       e.GetRank(),
		"user_id":    e.GetUserId(),
		"user_email": e.GetUserEmail(),
		"nickname":   e.GetNickname(),
		"score":      e.GetScore(),
		"updated_at": e.GetUpdatedAt(),
		"level":      level,
	}
}

func ScoreEntries(list []*leaderboardPb.ScoreEntry) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		if h := ScoreEntry(e); h != nil {
			out = append(out, h)
		}
	}
	return out
}
===== websocket =====
--- services/gateway/internal/websocket/hub.go ---
package websocket

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub управляет всеми WebSocket соединениями
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			n := len(h.clients)
			h.mu.Unlock()
			slog.Info("websocket client connected", "total", n)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			n := len(h.clients)
			h.mu.Unlock()
			slog.Info("websocket client disconnected", "total", n)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем клиентам
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}
===== webhook =====
--- services/gateway/internal/webhook/verify.go ---
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// Authorize verifies HMAC header (preferred) or shared secret equality.
func Authorize(configuredSecret, providedSecret string, rawBody []byte, signatureHeader string) error {
	if configuredSecret == "" {
		return nil
	}
	if signatureHeader != "" {
		if VerifyHMACSHA256(configuredSecret, rawBody, signatureHeader) {
			return nil
		}
		return ErrInvalidSignature
	}
	if providedSecret != "" && subtle.ConstantTimeCompare([]byte(providedSecret), []byte(configuredSecret)) == 1 {
		return nil
	}
	return ErrUnauthorized
}

var (
	ErrUnauthorized     = errString("webhook unauthorized")
	ErrInvalidSignature = errString("webhook signature invalid")
)

type errString string

func (e errString) Error() string { return string(e) }

func VerifyHMACSHA256(secret string, rawBody []byte, signatureHeader string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	sig := strings.TrimSpace(signatureHeader)
	if i := strings.Index(sig, "="); i >= 0 {
		sig = sig[i+1:]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	sum := mac.Sum(nil)
	got, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(sum, got) == 1
}
[denismatveev@c0der event_horizon]$ 

