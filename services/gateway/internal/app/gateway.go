package app

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "log"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/nats-io/nats.go"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/redis/go-redis/v9"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/protobuf/types/known/structpb"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

    "github.com/Eastwesser/event-horizon/services/gateway/internal/cache"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/circuit"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/client"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/config"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/middleware"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
    "github.com/Eastwesser/event-horizon/services/gateway/api"
    authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
    gamePb "github.com/Eastwesser/event-horizon/services/game/proto"
    leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
    billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
    profilePb "github.com/Eastwesser/event-horizon/services/profile/proto"
    shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
    inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
    paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
    authorsPb "github.com/Eastwesser/event-horizon/services/authors/proto"
    historyPb "github.com/Eastwesser/event-horizon/services/history/proto"
    analyticsPb "github.com/Eastwesser/event-horizon/services/analytics/proto"
)

const (
    RoleUser   = "user"
    RoleAuthor = "author"
    RoleAdmin  = "admin"
)

func withUserRole(ctx context.Context, c *gin.Context) context.Context {
    return metadata.AppendToOutgoingContext(ctx,
        "x-user-role", middleware.Role(c),
        "x-user-id", middleware.UserID(c),
    )
}

func newServiceBreaker(name string) *circuit.Breaker {
    return circuit.New(circuit.Settings{
        Name:        name,
        MaxRequests: 3,
        Timeout:     10 * time.Second,
        ReadyToTrip: func(counts circuit.Counts) bool {
            return counts.ConsecutiveFailures >= 5
        },
    })
}

// throughBreaker runs fn under the circuit breaker. On open circuit writes 503 and returns ErrOpen.
func throughBreaker(b *circuit.Breaker, c *gin.Context, fn func() (any, error)) (any, error) {
    out, err := b.Execute(fn)
    if err == circuit.ErrOpen {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable", "circuit": b.Name()})
        return nil, err
    }
    return out, err
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
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
            log.Printf("🟢 WebSocket client connected. Total: %d", len(h.clients))
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            log.Printf("🔴 WebSocket client disconnected. Total: %d", len(h.clients))
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
        endpoint = "jaeger:4317"  // ← для Docker используем имя контейнера
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

    gameConn, err := grpc.NewClient(cfg.GameAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to game: %v", err)
    }
    defer gameConn.Close()
    gameClient := gamePb.NewGameServiceClient(gameConn)

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
    r := gin.Default()

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

    // Week-1: expose OpenAPI + Swagger UI (canonical HTTP contract)
    r.GET("/openapi.yaml", func(c *gin.Context) {
        c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.OpenAPIYAML)
    })
    r.GET("/docs", func(c *gin.Context) {
        c.Header("Content-Type", "text/html; charset=utf-8")
        _, _ = c.Writer.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <title>Event Horizon API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/openapi.yaml',
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>`))
    })

    r.GET("/ws/leaderboard", func(c *gin.Context) {
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

    r.POST("/api/auth/register", func(c *gin.Context) {
        var req authPb.RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        resp, err := authClient.GetClient().Register(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        eventData := map[string]interface{}{
            "event":   "user.registered",
            "user_id": resp.UserId,
            "email":   resp.Email,
        }
        eventJSON, _ := json.Marshal(eventData)
        if js != nil {
            js.Publish("event.user.registered", eventJSON)
            log.Printf("📡 Published event: user.registered for %s", resp.Email)
        } else {
            log.Printf("⚠️ NATS not available, event not published")
        }
        c.JSON(http.StatusOK, resp)
    })

    r.POST("/api/auth/login", func(c *gin.Context) {
        var req authPb.LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().Login(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

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

    r.POST("/api/auth/refresh", func(c *gin.Context) {
        var req struct {
            RefreshToken string `json:"refresh_token"`
        }
        if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
            return
        }
        resp, err := authClient.GetClient().RefreshToken(c.Request.Context(), &authPb.RefreshTokenRequest{
            RefreshToken: req.RefreshToken,
        })
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "access_token":  resp.AccessToken,
            "refresh_token": resp.RefreshToken,
            "token_type":    resp.TokenType,
            "expires_in":    resp.ExpiresIn,
            "user_id":       resp.UserId,
            "role":          resp.Role,
        })
    })

    r.GET("/api/auth/whoami", middleware.RequireAuth(authClient), func(c *gin.Context) {
        token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))
        resp, err := authClient.GetClient().Whoami(c.Request.Context(), &authPb.WhoamiRequest{AccessToken: token})
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "user_id":  resp.UserId,
            "email":    resp.Email,
            "role":     resp.Role,
            "nickname": resp.Nickname,
        })
    })

    r.POST("/api/auth/logout", middleware.RequireAuth(authClient), func(c *gin.Context) {
        token, _ := middleware.ExtractBearerToken(c.GetHeader("Authorization"))
        _, err := authClient.GetClient().Logout(c.Request.Context(), &authPb.LogoutRequest{Token: token})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"success": true, "message": "logged out"})
    })

    // Admin-only: change a user's role (user/author/admin).
    r.POST("/api/auth/update-role", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
        var req struct {
            UserID string `json:"user_id"`
            Role   string `json:"role"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().UpdateRole(c.Request.Context(), &authPb.UpdateRoleRequest{
            UserId: req.UserID,
            Role:   req.Role,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": resp.Success, "message": resp.Message})
    })

    scoreCache := cache.NewScoreCache(2 * time.Second)

    r.GET("/api/auth/user", middleware.RequireAuth(authClient), func(c *gin.Context) {
        resp, err := authClient.GetClient().GetUser(c.Request.Context(), &authPb.GetUserRequest{
            UserId: middleware.UserID(c),
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "user_id":     resp.UserId,
            "email":       resp.Email,
            "nickname":    resp.Nickname,
            "best_scores": resp.BestScores,
            "total_score": resp.TotalScore,
            "role":        resp.Role,
        })
    })

    r.GET("/api/profile", middleware.RequireAuth(authClient), func(c *gin.Context) {
        userID := middleware.UserID(c)

        conn, err := grpc.Dial(cfg.ProfileAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer conn.Close()

        client := profilePb.NewProfileServiceClient(conn)
        resp, err := client.GetProfile(c.Request.Context(), &profilePb.GetProfileRequest{
            UserId: userID,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    billingConn, err := grpc.NewClient(cfg.BillingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to billing: %v", err)
    }
    defer billingConn.Close()
    billingClient := billingPb.NewBillingServiceClient(billingConn)
    billingCB := newServiceBreaker("billing")

    r.GET("/api/billing/balance/all", middleware.RequireAuth(authClient), func(c *gin.Context) {
        out, err := throughBreaker(billingCB, c, func() (any, error) {
            return billingClient.GetAllBalances(c.Request.Context(), &billingPb.GetAllBalancesRequest{
                UserId: middleware.UserID(c),
            })
        })
        if err == circuit.ErrOpen {
            return
        }
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

    shopConn, err := grpc.NewClient(cfg.ShopAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to shop: %v", err)
    }
    defer shopConn.Close()
    shopClient := shopPb.NewShopServiceClient(shopConn)
    shopCB := newServiceBreaker("shop")
    _ = shopCB // used below

    // Добавь эндпоинты:
    r.GET("/api/shop/items", middleware.RequireAuth(authClient), func(c *gin.Context) {
        category := c.Query("category")
        gameID := c.Query("game_id")

        resp, err := shopClient.GetItems(c.Request.Context(), &shopPb.GetItemsRequest{
            UserId:   middleware.UserID(c),
            Category: category,
            GameId:   gameID,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp.Items)
    })

    r.POST("/api/shop/purchase", middleware.RequireAuth(authClient), func(c *gin.Context) {
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
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        resp := out.(*shopPb.PurchaseItemResponse)

        c.JSON(http.StatusOK, resp)
    })

    r.GET("/api/shop/inventory", middleware.RequireAuth(authClient), func(c *gin.Context) {
        resp, err := shopClient.GetInventory(c.Request.Context(), &shopPb.GetInventoryRequest{
            UserId: middleware.UserID(c),
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp.Items)
    })

    // --- Payment (Boosty subscription) ---
    paymentConn, err := grpc.NewClient(cfg.PaymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to payment: %v", err)
    }
    defer paymentConn.Close()
    paymentClient := paymentPb.NewPaymentServiceClient(paymentConn)
    paymentCB := newServiceBreaker("payment")

    r.POST("/api/payment/checkout", middleware.RequireAuth(authClient), func(c *gin.Context) {
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
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

    r.GET("/api/payment/subscription", middleware.RequireAuth(authClient), func(c *gin.Context) {
        out, err := throughBreaker(paymentCB, c, func() (any, error) {
            return paymentClient.GetSubscription(c.Request.Context(), &paymentPb.GetSubscriptionRequest{
                UserId: middleware.UserID(c),
            })
        })
        if err == circuit.ErrOpen {
            return
        }
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

    r.GET("/api/payment/can-purchase-merch", middleware.RequireAuth(authClient), func(c *gin.Context) {
        out, err := throughBreaker(paymentCB, c, func() (any, error) {
            return paymentClient.CanPurchaseMerch(c.Request.Context(), &paymentPb.CanPurchaseMerchRequest{
                UserId: middleware.UserID(c),
            })
        })
        if err == circuit.ErrOpen {
            return
        }
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        resp := out.(*paymentPb.CanPurchaseMerchResponse)
        c.JSON(http.StatusOK, gin.H{"allowed": resp.Allowed, "reason": resp.Reason})
    })

    r.POST("/api/payment/webhook", func(c *gin.Context) {
        var req struct {
            PaymentID     string `json:"payment_id"`
            ProviderRef   string `json:"provider_ref"`
            WebhookSecret string `json:"webhook_secret"`
        }
        if err := c.ShouldBindJSON(&req); err != nil || req.PaymentID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
            return
        }
        resp, err := paymentClient.ConfirmPayment(c.Request.Context(), &paymentPb.ConfirmPaymentRequest{
            PaymentId:     req.PaymentID,
            ProviderRef:   req.ProviderRef,
            WebhookSecret: req.WebhookSecret,
        })
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "success":         resp.Success,
            "message":         resp.Message,
            "subscription_id": resp.SubscriptionId,
            "expires_at_unix": resp.ExpiresAtUnix,
        })
    })

    // --- Authors ---
    authorsConn, err := grpc.NewClient(cfg.AuthorsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to authors: %v", err)
    }
    defer authorsConn.Close()
    authorsClient := authorsPb.NewAuthorsServiceClient(authorsConn)

    r.PUT("/api/authors/me", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
        var req struct {
            DisplayName string `json:"display_name"`
            Bio         string `json:"bio"`
            AvatarURL   string `json:"avatar_url"`
        }
        if err := c.ShouldBindJSON(&req); err != nil || req.DisplayName == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "display_name is required"})
            return
        }
        resp, err := authorsClient.UpsertProfile(c.Request.Context(), &authorsPb.UpsertProfileRequest{
            UserId:      middleware.UserID(c),
            DisplayName: req.DisplayName,
            Bio:         req.Bio,
            AvatarUrl:   req.AvatarURL,
        })
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, resp.Author)
    })

    r.GET("/api/authors/:user_id", func(c *gin.Context) {
        resp, err := authorsClient.GetAuthor(c.Request.Context(), &authorsPb.GetAuthorRequest{
            UserId: c.Param("user_id"),
        })
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, resp.Author)
    })

    r.GET("/api/authors", func(c *gin.Context) {
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        resp, err := authorsClient.ListAuthors(c.Request.Context(), &authorsPb.ListAuthorsRequest{
            Limit:  int32(limit),
            Offset: int32(offset),
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"authors": resp.Authors, "total": resp.Total})
    })

    // --- History ---
    historyConn, err := grpc.NewClient(cfg.HistoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to history: %v", err)
    }
    defer historyConn.Close()
    historyClient := historyPb.NewHistoryServiceClient(historyConn)

    r.GET("/api/history", middleware.RequireAuth(authClient), func(c *gin.Context) {
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        resp, err := historyClient.ListEvents(c.Request.Context(), &historyPb.ListEventsRequest{
            UserId:    middleware.UserID(c),
            EventType: c.Query("event_type"),
            Limit:     int32(limit),
            Offset:    int32(offset),
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"events": resp.Events, "total": resp.Total})
    })

    // --- Analytics ---
    analyticsConn, err := grpc.NewClient(cfg.AnalyticsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to analytics: %v", err)
    }
    defer analyticsConn.Close()
    analyticsClient := analyticsPb.NewAnalyticsServiceClient(analyticsConn)

    r.GET("/api/analytics/dau", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
        days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
        resp, err := analyticsClient.GetDAU(c.Request.Context(), &analyticsPb.GetDAURequest{Days: int32(days)})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"days": resp.Days})
    })

    r.GET("/api/analytics/mau", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
        days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
        resp, err := analyticsClient.GetMAU(c.Request.Context(), &analyticsPb.GetMAURequest{Days: int32(days)})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"mau": resp.Mau, "window_days": resp.WindowDays})
    })

    r.GET("/api/analytics/retention", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
        cohort, _ := strconv.Atoi(c.DefaultQuery("cohort_days_ago", "7"))
        window, _ := strconv.Atoi(c.DefaultQuery("window_days", "7"))
        resp, err := analyticsClient.GetRetention(c.Request.Context(), &analyticsPb.GetRetentionRequest{
            CohortDaysAgo: int32(cohort),
            WindowDays:    int32(window),
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "cohort_day":  resp.CohortDay,
            "cohort_size": resp.CohortSize,
            "points":      resp.Points,
        })
    })

    // --- Inventory gRPC клиент ---
    inventoryConn, err := grpc.NewClient(cfg.InventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to inventory: %v", err)
    }
    defer inventoryConn.Close()
    inventoryClient := inventoryPb.NewInventoryServiceClient(inventoryConn)
    inventoryCB := newServiceBreaker("inventory")
    _ = inventoryCB

    // GET /api/inventory/items — список товаров с фильтрами
    r.GET("/api/inventory/items", middleware.RequireAuth(authClient), func(c *gin.Context) {
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

        resp, err := inventoryClient.SearchItems(c.Request.Context(), &inventoryPb.SearchItemsRequest{
            Filters: filters,
            Limit:   limit,
            Offset:  offset,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    // POST /api/inventory/items — создать товар (только author/admin)
    r.POST("/api/inventory/items", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
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

        resp, err := inventoryClient.CreateItem(withUserRole(c.Request.Context(), c), &inventoryPb.CreateItemRequest{
            AuthorId:    userID,
            Type:        req.Type,
            Name:        req.Name,
            Description: req.Description,
            Price:       req.Price,
            Stock:       req.Stock,
            Attributes:  attrs,
            Images:      req.Images,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    // POST /api/inventory/items/bulk — массовое создание (author/admin)
    // Must be registered before /items/:id so "bulk" is not captured as an id.
    r.POST("/api/inventory/items/bulk", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
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

        resp, err := inventoryClient.BulkCreateItems(c.Request.Context(), &inventoryPb.BulkCreateItemsRequest{
            Items: pbItems,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": resp.Success, "count": resp.Count})
    })

    // GET /api/inventory/items/:id — получить товар по ID
    r.GET("/api/inventory/items/:id", middleware.RequireAuth(authClient), func(c *gin.Context) {
        itemID := c.Param("id")
        if itemID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
            return
        }

        resp, err := inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{
            Id: itemID,
        })
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    // PUT /api/inventory/items/:id — обновить товар (author владелец или admin)
    r.PUT("/api/inventory/items/:id", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
        userID := middleware.UserID(c)

        itemID := c.Param("id")
        if itemID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
            return
        }

        if middleware.Role(c) != RoleAdmin {
            existing, err := inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
            if err != nil || existing.Item == nil {
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

        resp, err := inventoryClient.UpdateItem(withUserRole(c.Request.Context(), c), &inventoryPb.UpdateItemRequest{
            Id:          itemID,
            AuthorId:    userID,
            Type:        req.Type,
            Name:        req.Name,
            Description: req.Description,
            Price:       req.Price,
            Stock:       req.Stock,
            Attributes:  attrs,
            Images:      req.Images,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    // DELETE /api/inventory/items/:id — удалить товар (author владелец или admin)
    r.DELETE("/api/inventory/items/:id", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
        userID := middleware.UserID(c)

        itemID := c.Param("id")
        if itemID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
            return
        }

        if middleware.Role(c) != RoleAdmin {
            existing, err := inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
            if err != nil || existing.Item == nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
                return
            }
            if existing.Item.AuthorId != userID {
                c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own items"})
                return
            }
        }

        _, err := inventoryClient.DeleteItem(withUserRole(c.Request.Context(), c), &inventoryPb.DeleteItemRequest{
            Id: itemID,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true, "message": "item deleted"})
    })

    // POST /api/inventory/items/:id/reserve — уменьшить stock
    r.POST("/api/inventory/items/:id/reserve", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
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

        resp, err := inventoryClient.ReserveItem(withUserRole(c.Request.Context(), c), &inventoryPb.ReserveItemRequest{
            Id:       itemID,
            Quantity: req.Quantity,
        })
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "success":         resp.Success,
            "remaining_stock": resp.RemainingStock,
        })
    })

    // DELETE /api/inventory/items/:id/soft — мягкое удаление (author владелец или admin)
    r.DELETE("/api/inventory/items/:id/soft", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
        userID := middleware.UserID(c)
        itemID := c.Param("id")
        if itemID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
            return
        }

        if middleware.Role(c) != RoleAdmin {
            existing, err := inventoryClient.GetItem(c.Request.Context(), &inventoryPb.GetItemRequest{Id: itemID})
            if err != nil || existing.Item == nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
                return
            }
            if existing.Item.AuthorId != userID {
                c.JSON(http.StatusForbidden, gin.H{"error": "you can only soft-delete your own items"})
                return
            }
        }

        if _, err := inventoryClient.SoftDeleteItem(withUserRole(c.Request.Context(), c), &inventoryPb.SoftDeleteItemRequest{Id: itemID}); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true, "message": "item soft deleted"})
    })

    // POST /api/inventory/items/:id/restore — восстановить после soft delete (author владелец или admin)
    r.POST("/api/inventory/items/:id/restore", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
        itemID := c.Param("id")
        if itemID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
            return
        }

        if _, err := inventoryClient.RestoreItem(withUserRole(c.Request.Context(), c), &inventoryPb.RestoreItemRequest{Id: itemID}); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true, "message": "item restored"})
    })

    // GET /api/inventory/stats — статистика по товарам (только admin)
    r.GET("/api/inventory/stats", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAdmin), func(c *gin.Context) {
        resp, err := inventoryClient.GetStats(withUserRole(c.Request.Context(), c), &inventoryPb.EmptyRequest{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "total_items": resp.TotalItems,
            "by_type":     resp.ByType,
            "by_author":   resp.ByAuthor,
        })
    })

    r.GET("/api/leaderboard", func(c *gin.Context) {
        gameID := c.Query("game_id")
        limit := c.Query("limit")

        conn, err := grpc.Dial(cfg.LeaderboardAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer conn.Close()

        leaderboardClient := leaderboardPb.NewLeaderboardServiceClient(conn)

        var limitInt int32 = 10
        if limit != "" {
            if l, err := strconv.Atoi(limit); err == nil {
                limitInt = int32(l)
            }
        }

        resp, err := leaderboardClient.GetTopScores(c.Request.Context(), &leaderboardPb.GetTopScoresRequest{
            GameId: gameID,
            Limit:  limitInt,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"entries": resp.Entries})
    })

    r.POST("/api/auth/update-nickname", middleware.RequireAuth(authClient), func(c *gin.Context) {
        var req struct {
            Nickname string `json:"nickname"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        _, err := authClient.GetClient().UpdateNickname(c.Request.Context(), &authPb.UpdateNicknameRequest{
            UserId:   middleware.UserID(c),
            Nickname: req.Nickname,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"success": true, "message": "nickname updated"})
    })

    // /api/game/submit previously trusted the "user_id" field from the request body
    // with no auth check at all, letting anyone submit scores for any account.
    // It now requires a valid token and always uses the authenticated user's id.
    r.POST("/api/game/submit", middleware.RequireAuth(authClient), func(c *gin.Context) {
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

        resp, err := gameClient.SubmitScore(c.Request.Context(), &gamePb.SubmitScoreRequest{
            UserId:    middleware.UserID(c),
            GameId:    req.GameID,
            Level:     req.Level,
            Score:     req.Score,
            UserEmail: req.UserEmail,
            Nickname:  req.Nickname,
            Seed:      req.Seed,
            Moves:     moves,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        respJSON, _ := json.Marshal(resp)
        scoreCache.Set(cacheKey, respJSON)
        c.JSON(http.StatusOK, resp)
    })

    srv := &http.Server{
        Addr:    "0.0.0.0:" + cfg.Port,
        Handler: r,
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
    billingConn.Close()
    if nc != nil {
        nc.Drain()
    }

    log.Println("Gateway stopped gracefully")
}
