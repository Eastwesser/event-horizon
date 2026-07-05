package main

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "strconv"
    "strings"
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
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

    "github.com/Eastwesser/event-horizon/services/gateway/internal/cache"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/client"
    "github.com/Eastwesser/event-horizon/services/gateway/internal/config"
    // "github.com/Eastwesser/event-horizon/services/gateway/internal/middleware"
    // "github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
    authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
    gamePb "github.com/Eastwesser/event-horizon/services/game/proto"
    leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
    billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
    profilePb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

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

func getUserIDFromToken(tokenString string) (string, error) {
    parts := strings.Split(tokenString, " ")
    if len(parts) == 2 && parts[0] == "Bearer" {
        tokenString = parts[1]
    }

    parts = strings.Split(tokenString, ".")
    if len(parts) != 3 {
        return "", fmt.Errorf("invalid token format")
    }

    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return "", err
    }

    var claims map[string]interface{}
    if err := json.Unmarshal(payload, &claims); err != nil {
        return "", err
    }

    userID, ok := claims["user_id"].(string)
    if !ok {
        return "", fmt.Errorf("user_id not found in token")
    }

    return userID, nil
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

func main() {
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
        // limiter := ratelimit.NewRateLimiter(rdb)
        // r.Use(middleware.RateLimitMiddleware(limiter))
    }

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

        userId := resp.UserId
        if userId == "" {
            parts := strings.Split(resp.AccessToken, ".")
            if len(parts) == 3 {
                payload, err := base64.RawURLEncoding.DecodeString(parts[1])
                if err == nil {
                    var claims map[string]interface{}
                    json.Unmarshal(payload, &claims)
                    if uid, ok := claims["user_id"].(string); ok {
                        userId = uid
                    }
                }
            }
        }

        c.JSON(http.StatusOK, gin.H{
            "access_token": resp.AccessToken,
            "token_type":   resp.TokenType,
            "expires_in":   resp.ExpiresIn,
            "user_id":      userId,
        })
    })

    scoreCache := cache.NewScoreCache(2 * time.Second)

    r.GET("/api/auth/user", func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
            return
        }

        userID, err := getUserIDFromToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        resp, err := authClient.GetClient().GetUser(c.Request.Context(), &authPb.GetUserRequest{
            UserId: userID,
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
        })
    })

    r.GET("/api/profile", func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        userID, err := getUserIDFromToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

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

    r.GET("/api/billing/balance/all", func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
            return
        }

        userID, err := getUserIDFromToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        resp, err := billingClient.GetAllBalances(c.Request.Context(), &billingPb.GetAllBalancesRequest{
            UserId: userID,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

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

    r.POST("/api/auth/update-nickname", func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        userID, err := getUserIDFromToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        
        var req struct {
            Nickname string `json:"nickname"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        _, err = authClient.GetClient().UpdateNickname(c.Request.Context(), &authPb.UpdateNicknameRequest{
            UserId:   userID,
            Nickname: req.Nickname,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"success": true, "message": "nickname updated"})
    })

    r.POST("/api/game/submit", func(c *gin.Context) {
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
            UserId:    req.UserID,
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
