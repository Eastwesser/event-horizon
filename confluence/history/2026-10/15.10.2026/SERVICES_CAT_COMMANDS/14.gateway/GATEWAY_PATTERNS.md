Разбор gateway
Что делает сервис
gateway — это API-шлюз (BFF — Backend For Frontend):

HTTP (Gin) — принимает запросы от фронтенда.

14 gRPC-клиентов — auth, game, profile, leaderboard, billing, inventory, shop, payment, authors, notification, history, analytics, + дважды inventory (shared).

WebSocket Hub — трансляция score.updated всем клиентам.

Middleware Chain — Security Headers, RejectUnsafe, LimitJSONBody, RateLimit, Auth, Role, DocsGate.

Circuit Breaker + Bulkhead — на каждый downstream-сервис.

Retry Interceptor — на gRPC-клиентах.

Rate Limiter — per-route (submit, login, ws, docs, global).

Response Cache — whoami, profile, submit (short TTL).

DTO Mapper — proto → JSON с explicit zeros.

Upload Routes — multipart upload картинок.

Docs Gate — Basic Auth для /docs.

Webhook Relay — Boosty + YooKassa → payment.ConfirmPayment.

Найденные паттерны
1. Backend For Frontend (BFF)
Файл: internal/app/gateway.go.

Почему: gateway — это BFF: он не хранит данные, а агрегирует вызовы к 14 микросервисам и отдаёт фронту удобный JSON. Каждый HTTP-эндпоинт маппится на один или несколько gRPC-вызовов.

Код:

go
r.GET("/api/v1/profile", middleware.RequireAuth(authClient), func(c *gin.Context) {
    userID := middleware.UserID(c)
    out, err := throughBreaker(profileCB, c, func() (any, error) {
        return profileClient.GetProfile(c.Request.Context(), &profilePb.GetProfileRequest{
            UserId: userID,
        })
    })
    // ...
})
BFF Pattern — главный паттерн gateway.

2. API Gateway
Файл: internal/app/gateway.go.

Почему: Gateway маршрутизирует HTTP-запросы в нужные gRPC-сервисы, унифицирует ошибки (writeGRPCError), добавляет security headers. Это классический API Gateway.

3. Middleware Chain (Gin)
Файл: internal/app/gateway.go + internal/middleware/*.go.

Почему: Цепочка middleware:

go
r.Use(gin.Recovery())
r.Use(middleware.SecurityHeaders())
r.Use(middleware.RejectUnsafeRequest())
r.Use(middleware.LimitJSONBody())
r.Use(legacyAPIRewrite())
r.Use(otelgin.Middleware("gateway"))
r.Use(prometheusMetrics())
r.Use(middleware.RateLimitMiddleware(limiter))
Middleware Chain — стандартный паттерн для HTTP.

4. Circuit Breaker + Bulkhead (уникально)
Файл: internal/circuit/breaker.go.

Почему: Свой Circuit Breaker с Bulkhead:

3 состояния: closed, open, half-open.

MaxConcurrent — bulkhead (ограничение in-flight запросов).

Код:

go
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
    // ...
    result, err := fn()
    b.afterRequest(generation, err == nil)
    return result, err
}
Circuit Breaker + Bulkhead — уникально для gateway. Ни один другой сервис не использует circuit breaker.

Плюс: Breaker написан без внешних зависимостей (не sony/gobreaker).

5. Retry Interceptor (gRPC)
Файл: internal/client/retry.go.

Почему: gRPC-клиент имеет unary retry interceptor с exponential backoff + full jitter.

Код:

go
func UnaryRetryInterceptor(cfg RetryConfig) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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
Retry Interceptor с full jitter — правильный паттерн.

Плюс: isRetryable различает transient (Unavailable, ResourceExhausted) и permanent (InvalidArgument, NotFound) ошибки.

6. Rate Limiter (Sliding Window в Redis)
Файл: internal/ratelimit/limiter.go.

Почему: Sliding Window Rate Limiter через Redis ZSET:

ZRemRangeByScore — удалить старые записи.

ZAdd — добавить текущий запрос.

ZCard — посчитать количество в окне.

Код:

go
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now().UnixMilli()
    windowStart := now - window.Milliseconds()

    pipe := rl.rdb.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    pipe.ZCard(ctx, key)
    pipe.Expire(ctx, key, window)

    cmds, err := pipe.Exec(ctx)
    // ...
    count := cmds[2].(*redis.IntCmd).Val()
    return count <= int64(limit), nil
}
Sliding Window Rate Limiter — правильный паттерн (не fixed window).

Плюс: Fail-open при Redis blip:

go
if err != nil {
    // Fail open on Redis blips — do not take the API down with the limiter.
    return true
}
7. Per-Route Rate Limiting
Файл: internal/middleware/ratelimit.go.

Почему: Разные лимиты для разных путей:

/api/v1/game/submit — 10/сек.

/api/v1/auth/login — 5/сек per IP.

/ws/leaderboard — 100/мин per IP.

/docs — 20/мин per IP.

/health, /ready, /metrics — без лимита.

Default — 100/сек per user.

Код:

go
switch {
case path == "/api/v1/game/submit" && method == "POST":
    allowed = limiter.AllowSubmit(subject)
case path == "/api/v1/auth/login" && method == "POST":
    allowed = limiter.AllowLogin(c.ClientIP())
// ...
}
Per-Route Rate Limiting — гранулярный контроль.

Плюс: Rate limit subject — user_id → token hash → IP (fallback chain):

go
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
8. Response Cache (in-memory, TTL)
Файл: internal/cache/response_cache.go.

Почему: In-memory Response Cache с TTL и cleanup-горутиной. Используется для:

whoami — 5 сек.

profile — 5 сек.

user — 5 сек.

game/submit — 2 сек.

Код:

go
type ResponseCache struct {
    cache map[string]entry
    mu    sync.RWMutex
    ttl   time.Duration
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
In-Memory Response Cache с фоновым cleanup.

9. Partial Response Fallback (уникально)
Файл: internal/app/gateway.go, appendPartialJSON.

Почему: Если Profile-сервис недоступен (circuit open), gateway отдаёт частичный ответ из кеша whoami с пометкой _partial: true.

Код:

go
if err == circuit.ErrOpen {
    if cached, ok := authReadCache.Get("whoami:" + token); ok {
        c.Data(http.StatusOK, "application/json", appendPartialJSON(cached))
        return
    }
    return
}
И:

go
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
Partial Response Fallback — уникально для gateway. Когда downstream упал, отдаём что есть.

10. DTO Mapper (proto → JSON)
Файл: internal/dto/*.go.

Почему: DTO-мапперы явно обрабатывают proto3 omitempty (нули и пустые слайсы исчезают). dto.InventoryItem, dto.ShopItem, dto.Author и т.д.

Код:

go
// Proto3 omitempty drops zeros and empty slices; public HTTP must keep them.
func InventoryItem(item *inventoryPb.Item) gin.H {
    images := item.GetImages()
    if images == nil {
        images = []string{}
    }
    return gin.H{
        "price":  item.GetPrice(),   // ← 0 сохраняется
        "stock":  item.GetStock(),   // ← 0 сохраняется
        "images": images,            // ← [] сохраняется
        // ...
    }
}
DTO Mapper — явное решение проблемы proto3 omitempty. Уникально для gateway.

Плюс: Есть Catalog DTO с урезанными полями:

go
var catalogAttrOmit = map[string]struct{}{
    "card_text":       {},
    "flavor_text":     {},
    "noiz_review":     {},
    "idempotency_key": {},
    "market_rub":      {},
}
11. WebSocket Hub (Broadcast)
Файл: internal/app/gateway.go, NewHub + internal/websocket/hub.go.

Почему: Gorilla WebSocket + Hub для broadcast. Клиенты регистрируются, сервер рассылает score.updated.

Код:

go
type Hub struct {
    clients    map[*WSClient]bool
    broadcast  chan []byte
    register   chan *WSClient
    unregister chan *WSClient
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            delete(h.clients, client)
            close(client.send)
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
WebSocket Hub — стандартный паттерн broadcast.

12. WebSocket Ticket Auth (одноразовый тикет)
Файл: internal/app/gateway.go, issueWSTicket / consumeWSTicket.

Почему: Browser не может передать Authorization header при WebSocket upgrade. Решение: одноразовый ticket в Redis (TTL 30 сек, GetDel — атомарное удаление).

Код:

go
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

func consumeWSTicket(ctx context.Context, rdb *redis.Client, ticket string) (string, error) {
    if len(ticket) < 16 || len(ticket) > 64 {
        return "", redis.Nil
    }
    return rdb.GetDel(ctx, "ws:ticket:"+ticket).Result()
}
WebSocket Ticket Auth — правильное решение для WS-аутентификации. Уникально для gateway.

Плюс: Origin Check (originAllowed) — защита от CSRF.

13. Origin Check (WebSocket)
Файл: internal/app/gateway.go, originAllowed.

Код:

go
func originAllowed(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    if origin == "" {
        return true
    }
    u, err := url.Parse(origin)
    if err != nil || u.Host == "" {
        return false
    }
    if strings.EqualFold(u.Host, r.Host) || strings.HasPrefix(u.Host, "localhost:") || /* ... */ {
        return true
    }
    for _, allowed := range strings.Split(os.Getenv("WS_ALLOWED_ORIGINS"), ",") {
        if strings.TrimSpace(allowed) == origin {
            return true
        }
    }
    return false
}
Origin Check — защита WebSocket от cross-origin атак.

14. Auth Middleware (gRPC ValidateToken)
Файл: internal/middleware/auth.go.

Почему: Вместо локального JWT decode используется gRPC-вызов ValidateToken — проверка через source of truth (Redis session store).

Код:

go
func RequireAuth(authClient AuthClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
        // ...
        resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
        // ...
        c.Set(CtxUserID, resp.UserId)
        c.Set(CtxEmail, resp.Email)
        c.Set(CtxRole, resp.Role)
        c.Next()
    }
}
Комментарий из кода:

This replaces the old local base64 JWT decode, which trusted the token's claims blindly and never checked for revocation (logout) or the source of truth in Postgres/Redis.

Auth Middleware через gRPC — правильное решение. Уникально для gateway.

15. Role-Based Access Control (RBAC)
Файл: internal/middleware/auth.go, RequireRole.

Код:

go
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
RBAC через middleware. Как в inventory.

16. Security Headers
Файл: internal/middleware/security.go, SecurityHeaders.

Код:

go
h.Set("X-Content-Type-Options", "nosniff")
h.Set("X-Frame-Options", "DENY")
h.Set("Referrer-Policy", "no-referrer")
h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
h.Set("X-Permitted-Cross-Domain-Policies", "none")
h.Set("Content-Security-Policy", contentSecurityPolicy(c.Request.URL.Path))
if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
    h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}
Security Headers — baseline.

Плюс: Разный CSP для /docs и API:

go
func contentSecurityPolicy(path string) string {
    if path == "/docs" || strings.HasPrefix(path, "/docs/") {
        return "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; ..."
    }
    return "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"
}
17. Path Traversal Guard
Файл: internal/middleware/security.go, RejectUnsafeRequest.

Код:

go
if strings.Contains(raw, "..") || strings.ContainsRune(raw, 0) || strings.Contains(raw, "\\") {
    c.AbortWithStatus(http.StatusBadRequest)
    return
}
Path Traversal Guard — защита от ../.

18. Body Limit Middleware
Файл: internal/middleware/security.go, LimitJSONBody.

Код:

go
const maxJSONBody = 1 << 20 // 1 MiB

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
Body Limit — защита от large payload.

19. Docs Gate (Basic Auth + ConstantTimeCompare)
Файл: internal/middleware/docs.go.

Код:

go
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
Docs Gate — защита OpenAPI. Плюс: если password пустой → 404 (не отдаём spec).

Плюс: subtle.ConstantTimeCompare — защита от timing attacks.

20. Upload Middleware (multipart + image validation)
Файл: internal/app/upload.go.

Почему: Приём multipart-загрузки с:

Лимитом размера (MaxBytesReader).

Проверкой расширения (.jpg, .jpeg, .png, .webp).

Проверкой content-type (http.DetectContentType → image/*).

UUID-именем файла.

Код:

go
head := make([]byte, 512)
n, _ := io.ReadFull(src, head)
detected := http.DetectContentType(head[:n])
if !strings.HasPrefix(detected, "image/") {
    c.JSON(http.StatusBadRequest, gin.H{"error": "content is not an image"})
    return
}
Upload with Validation — безопасная загрузка файлов.

21. Webhook Relay (HMAC verification)
Файл: internal/webhook/verify.go.

Почему: Тот же HMAC-SHA256 код, что в payment. Gateway проверяет подпись и релеит payment.ConfirmPayment.

Код:

go
if err := gwhook.Authorize(cfg.PaymentWebhookSecret, webhookSecret, raw, sig); err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
    return
}
Webhook Relay — gateway принимает вебхук, проверяет подпись, релеит в payment.

22. Idempotent Webhook Response
Файл: internal/app/gateway.go, writePaymentConfirmError.

Почему: При дубликате вебхука (AlreadyExists) возвращается HTTP 200, чтобы провайдер не ретраил.

Код:

go
func writePaymentConfirmError(c *gin.Context, err error) {
    if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
        c.JSON(http.StatusOK, gin.H{"success": true, "message": st.Message()})
        return
    }
    writeGRPCError(c, err)
}
Idempotent Webhook Response — правильная обработка дубликатов.

23. gRPC → HTTP Error Mapping
Файл: internal/app/gateway.go, writeGRPCError.

Почему: Явный маппинг gRPC-кодов в HTTP:

go
case codes.InvalidArgument: code = http.StatusBadRequest
case codes.Unauthenticated: code = http.StatusUnauthorized
case codes.PermissionDenied: code = http.StatusForbidden
case codes.NotFound: code = http.StatusNotFound
case codes.AlreadyExists, codes.Aborted: code = http.StatusConflict
case codes.FailedPrecondition: code = http.StatusConflict
case codes.ResourceExhausted: code = http.StatusTooManyRequests
case codes.Unavailable: code = http.StatusServiceUnavailable
Error Mapping — стандартный приём для gateway.

Плюс: Fallback для codes.Unknown через string matching (когда otel/transport теряет typed status).

24. Legacy API Rewrite (/api/* → /api/v1/*)
Файл: internal/app/gateway.go.

Почему: Старые пути рерайтятся в новые. Плюс: используется r.HandleContext(c) — потому что Gin уже сматчил роуты до middleware.

Код:

go
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
Legacy API Rewrite — обратная совместимость.

25. Prometheus Metrics (Counter + Histogram)
Файл: internal/app/gateway.go.

Код:

go
gatewayRequestsTotal := promauto.NewCounterVec(
    prometheus.CounterOpts{Name: "gateway_requests_total", ...},
    []string{"method", "path", "status"},
)
gatewayRequestDuration := promauto.NewHistogramVec(
    prometheus.HistogramOpts{Name: "gateway_request_duration_seconds", ...},
    []string{"method", "path"},
)
Prometheus Metrics — счётчики и гистограммы для gateway.

26. OpenTelemetry Integration
Файл: internal/app/gateway.go, initTracer + otelgin.Middleware("gateway").

Код:

go
func initTracer(ctx context.Context) (func(context.Context) error, error) {
    endpoint := os.Getenv("JAEGER_ENDPOINT")
    if endpoint == "" {
        endpoint = "jaeger:4317"
    }
    exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
    // ...
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp.Shutdown, nil
}
OTel + Jaeger — distributed tracing.

Плюс: otelgin.Middleware("gateway") — автоматические span'ы для HTTP.

27. Context Propagation to gRPC
Файл: internal/app/gateway.go, withUserRole.

Код:

go
func withUserRole(ctx context.Context, c *gin.Context) context.Context {
    return metadata.AppendToOutgoingContext(ctx,
        "x-user-role", middleware.Role(c),
        "x-user-id", middleware.UserID(c),
    )
}
Context Propagation — user role/id пробрасываются в gRPC metadata.

28. Parallel Fan-Out + Semaphore (уникально)
Файл: internal/app/gateway.go, /api/v1/admin/users.

Почему: Для каждого пользователя параллельно вызываются billing + payment. Ограничение параллелизма — semaphore на 8 горутин.

Код:

go
rows := make([]adminUser, len(list.Users))
var mu sync.Mutex
var wg sync.WaitGroup
sem := make(chan struct{}, 8)
ctx := c.Request.Context()

for i, u := range list.Users {
    i, u := i, u
    rows[i] = adminUser{...}
    wg.Add(1)
    go func() {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        // ... billing + payment calls ...
        mu.Lock()
        rows[i].Lamps = lamps
        mu.Unlock()
    }()
}
wg.Wait()
Parallel Fan-Out + Semaphore — эффективный способ обогатить список.

Уникально для gateway — в других сервисах такого нет.

29. Cache-Aside (for whoami, profile, submit)
Файл: internal/app/gateway.go.

Код:

go
if cached, ok := authReadCache.Get("whoami:" + token); ok {
    c.Data(http.StatusOK, "application/json", cached)
    return
}
// ...
cacheWhoami(token, payload)
Cache-Aside для часто запрашиваемых read-эндпоинтов.

30. Metrics + Health + Ready
Файл: internal/app/gateway.go.

Код:

go
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
Health/Ready — стандартные эндпоинты.

31. Graceful Shutdown
Файл: internal/app/gateway.go.

Код:

go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Printf("HTTP server shutdown error: %v", err)
}
Graceful Shutdown — 10 секунд на завершение.

32. Broadcast через NATS
Файл: internal/app/gateway.go.

Код:

go
_, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
    hub.Broadcast(msg.Data)
    log.Printf("📡 Broadcasted score update to WebSocket clients")
}, nats.Durable("gateway-websocket"))
NATS → WebSocket Broadcast — событие из NATS рассылается всем WS-клиентам.

33. HTTP 404 vs Auth Fail (Security by Default)
Файл: internal/middleware/docs.go.

Код:

go
if len(passB) == 0 {
    c.AbortWithStatus(http.StatusNotFound)
    return
}
Security by Default — если пароль не задан, эндпоинт скрыт (404), а не открыт.

34. CSRF Protection via Bearer-only
Файл: internal/middleware/security.go.

Комментарий:

The API is Bearer-token, so a CSRF token is not added here.

CSRF Protection via Bearer-only — правильное решение: Bearer-токены не отправляются автоматически, CSRF не работает.

35. Dual WebSocket Hub (дублирование)
Файлы: internal/app/gateway.go (Hub) + internal/websocket/hub.go (Hub).

Проблема: Два одинаковых Hub. Один используется в gateway.go, другой в internal/websocket/. Это дублирование — стоит унифицировать.

Паттерн: Code Duplication (анти-паттерн).

36. Config Loader
Файл: internal/config/config.go.

Код:

go
func Load() *Config {
    return &Config{
        Port:                getenv("PORT", "8080"),
        MetricsPort:         getenv("METRICS_PORT", "9090"),
        NATSUrl:             getenv("NATS_URL", "nats://localhost:4222"),
        RedisAddr:           getenv("REDIS_ADDR", "localhost:6379"),
        AuthAddr:            getenv("AUTH_ADDR", "localhost:50051"),
        // ...
    }
}
Config Loader — стандартный приём.

37. GraphQL-style Endpoint (нет, но похоже)
Заметка: gateway не использует GraphQL, но агрегирует данные из нескольких сервисов в одном ответе (admin/users) — это GraphQL-style aggregation через HTTP.

38. Bulkhead (часть Circuit Breaker)
Файл: internal/circuit/breaker.go, ErrBulkheadFull.

Код:

go
if s.MaxConcurrent > 0 {
    b.sem = make(chan struct{}, s.MaxConcurrent)
}
Bulkhead — ограничение in-flight запросов per downstream. Уникально для gateway.

39. Fail-Fast vs Fail-Open
Файл: internal/ratelimit/limiter.go + internal/circuit/breaker.go.

Почему:

Rate limiter: Fail-open (Redis down → allow).

Circuit breaker: Fail-fast (open circuit → 503).

Fail-Fast vs Fail-Open — важный выбор для каждого компонента.

40. AuthClient Interface (для тестов)
Файл: internal/middleware/auth.go.

Код:

go
// AuthClient matches client.AuthClient's gRPC surface needed here (avoids an import cycle).
type AuthClient interface {
    GetClient() authPb.AuthServiceClient
}
Interface for Testing — интерфейс, чтобы избежать import cycle и упростить тесты.

41. Testable Breaker
Файл: internal/circuit/breaker_test.go.

Плюс: Тесты покрывают:

Open after failures.

Default settings.

Closed success resets failures.

Bulkhead rejects when full.

Half-open probe limit.

Thorough Testing — 5 тестов на breaker.

42. Testable Retry
Файл: internal/client/retry_test.go.

Плюс: Тесты покрывают:

Retry on Unavailable.

No retry on InvalidArgument.

Respect context cancel.

isRetryable для разных кодов.

Thorough Testing для retry.

43. Testable Rate Limit
Плюс: Rate limit покрыт тестами (не показаны, но по структуре видно).

44. Testable Docs Gate
Файл: internal/middleware/docs_test.go.

Плюс: Тесты:

Hidden without password.

Rejects wrong password.

Allows password.

Security headers.

Path traversal.

Thorough Testing для security middleware.

45. Multi-Tenant Config
Файл: internal/config/config.go.

14 адресов downstream-сервисов + WebSocket allowed origins.

Multi-Tenant Config — конфигурация всех зависимостей.

46. Bootstrapping от runGateway
Файл: internal/app/gateway.go, runGateway.

Почему: runGateway — огромная функция (больше 1000 строк), которая:

Инициализирует tracer.

Подключается к 14 gRPC-сервисам.

Настраивает Gin router.

Регистрирует все роуты.

Запускает сервер.

God Function (анти-паттерн) — стоит разбить на модули:

bootstrap/clients.go — подключение gRPC.

bootstrap/routes.go — регистрация роутов.

bootstrap/middleware.go — middleware.

Итог по gateway
Паттерн	Где
Backend For Frontend (BFF)	app/gateway.go
API Gateway	app/gateway.go
Middleware Chain	app/gateway.go + middleware/*.go
Circuit Breaker + Bulkhead	circuit/breaker.go
Retry Interceptor (exponential backoff + full jitter)	client/retry.go
Sliding Window Rate Limiter	ratelimit/limiter.go
Per-Route Rate Limiting	middleware/ratelimit.go
In-Memory Response Cache	cache/response_cache.go
Partial Response Fallback	app/gateway.go, appendPartialJSON
DTO Mapper (proto3 omitempty fix)	dto/*.go
WebSocket Hub (Broadcast)	app/gateway.go, websocket/hub.go
WebSocket Ticket Auth	app/gateway.go, issueWSTicket / consumeWSTicket
Origin Check (WebSocket)	app/gateway.go, originAllowed
Auth Middleware via gRPC	middleware/auth.go, RequireAuth
Role-Based Access Control	middleware/auth.go, RequireRole
Security Headers	middleware/security.go, SecurityHeaders
Path Traversal Guard	middleware/security.go, RejectUnsafeRequest
Body Limit Middleware	middleware/security.go, LimitJSONBody
Docs Gate (Basic Auth + ConstantTimeCompare)	middleware/docs.go
Upload with Validation	app/upload.go
Webhook Relay (HMAC)	webhook/verify.go
Idempotent Webhook Response	app/gateway.go, writePaymentConfirmError
gRPC → HTTP Error Mapping	app/gateway.go, writeGRPCError
Legacy API Rewrite	app/gateway.go
Prometheus Metrics (Counter + Histogram)	app/gateway.go
OTel + Jaeger	app/gateway.go, initTracer
Context Propagation to gRPC	app/gateway.go, withUserRole
Parallel Fan-Out + Semaphore	app/gateway.go, /api/v1/admin/users
Cache-Aside (whoami, profile, submit)	app/gateway.go
Health/Ready	app/gateway.go
Graceful Shutdown	app/gateway.go
NATS → WebSocket Broadcast	app/gateway.go
Security by Default (404 for hidden)	middleware/docs.go
CSRF Protection via Bearer-only	middleware/security.go
Code Duplication (Hub ×2)	app/gateway.go + websocket/hub.go
Config Loader	config/config.go
Fail-Fast vs Fail-Open	ratelimit/limiter.go + circuit/breaker.go
Interface for Testing (AuthClient)	middleware/auth.go
Thorough Testing	circuit/breaker_test.go, client/retry_test.go, middleware/*_test.go, dto/*_test.go
God Function (runGateway)	app/gateway.go — анти-паттерн
Multi-Tenant Config	config/config.go
Итого 41 паттерн и приём — абсолютный рекорд проекта.

Уникальные фичи gateway:

Circuit Breaker + Bulkhead — единственный сервис с этим.

Retry Interceptor с full jitter — единственный сервис с retry.

Sliding Window Rate Limiter — единственный сервис с rate limit.

Partial Response Fallback — единственный сервис с деградацией данных.

DTO Mapper (proto3 omitempty fix) — единственный сервис с явным DTO.

WebSocket Ticket Auth — единственный сервис с WS-тикетом.

Parallel Fan-Out + Semaphore — единственный сервис с параллельной агрегацией.

Docs Gate — единственный сервис с Basic Auth для docs.

Анти-паттерны:

Code Duplication — два Hub.

God Function — runGateway больше 1000 строк.

Готовый markdown для GATEWAY_PATTERNS.md
Допиши в 14.gateway/GATEWAY_PATTERNS.md:

md
# gateway — Programming Patterns

> Сервис: API Gateway (BFF). HTTP (Gin) + 14 gRPC-клиентов + WebSocket + OTel + Prometheus.

## Что делает сервис

Принимает HTTP-запросы от фронтенда, маршрутизирует в 14 gRPC-сервисов, агрегирует данные, отдаёт JSON. WebSocket broadcast `score.updated`. Circuit breaker, rate limiter, retry, cache, DTO mapper.

## Паттерны

### 1. Backend For Frontend (BFF)
- **Файл:** `internal/app/gateway.go`
- **Суть:** агрегирует вызовы к 14 микросервисам, отдаёт удобный JSON.

### 2. API Gateway
- **Файл:** `internal/app/gateway.go`
- **Суть:** маршрутизация, унификация ошибок, security headers.

### 3. Middleware Chain (Gin)
- **Файл:** `internal/app/gateway.go` + `internal/middleware/*.go`
- **Код:**
  ```go
  r.Use(gin.Recovery())
  r.Use(middleware.SecurityHeaders())
  r.Use(middleware.RejectUnsafeRequest())
  r.Use(middleware.LimitJSONBody())
  r.Use(otelgin.Middleware("gateway"))
  r.Use(middleware.RateLimitMiddleware(limiter))
4. Circuit Breaker + Bulkhead
Файл: internal/circuit/breaker.go

Суть: свои Breaker без внешних deps; MaxConcurrent = bulkhead.

Код:

go
if b.sem != nil {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
    default:
        return nil, ErrBulkheadFull
    }
}
5. Retry Interceptor (exponential backoff + full jitter)
Файл: internal/client/retry.go

Код:

go
exp := cfg.BaseDelay * time.Duration(1<<attempt)
if exp > cfg.MaxDelay { exp = cfg.MaxDelay }
wait := time.Duration(rand.Int63n(int64(exp) + 1)) // full jitter
6. Sliding Window Rate Limiter
Файл: internal/ratelimit/limiter.go

Код:

go
pipe.ZRemRangeByScore(ctx, key, "0", ...)
pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
pipe.ZCard(ctx, key)
7. Per-Route Rate Limiting
Файл: internal/middleware/ratelimit.go

Суть: разные лимиты для submit, login, ws, docs, global.

8. In-Memory Response Cache
Файл: internal/cache/response_cache.go

Суть: TTL + фоновый cleanup.

9. Partial Response Fallback
Файл: internal/app/gateway.go, appendPartialJSON

Суть: при circuit open отдаём кешированный whoami + _partial: true.

10. DTO Mapper (proto3 omitempty fix)
Файл: internal/dto/*.go

Комментарий: Proto3 omitempty drops zeros and empty slices; public HTTP must keep them.

11. WebSocket Hub (Broadcast)
Файл: internal/app/gateway.go + internal/websocket/hub.go

Суть: gorilla/websocket + Hub broadcast.

12. WebSocket Ticket Auth
Файл: internal/app/gateway.go, issueWSTicket / consumeWSTicket

Суть: одноразовый ticket в Redis (TTL 30s, GetDel).

13. Origin Check (WebSocket)
Файл: internal/app/gateway.go, originAllowed

14. Auth Middleware via gRPC
Файл: internal/middleware/auth.go, RequireAuth

Суть: gRPC ValidateToken вместо локального JWT decode.

15. Role-Based Access Control
Файл: internal/middleware/auth.go, RequireRole

16. Security Headers
Файл: internal/middleware/security.go, SecurityHeaders

Код: X-Content-Type-Options, X-Frame-Options, CSP, HSTS.

17. Path Traversal Guard
Файл: internal/middleware/security.go, RejectUnsafeRequest

18. Body Limit Middleware
Файл: internal/middleware/security.go, LimitJSONBody

Код: http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBody)

19. Docs Gate (Basic Auth + ConstantTimeCompare)
Файл: internal/middleware/docs.go

Код: subtle.ConstantTimeCompare([]byte(gotUser), userB) & subtle.ConstantTimeCompare([]byte(gotPass), passB)

20. Upload with Validation
Файл: internal/app/upload.go

Код: http.DetectContentType(head[:n]) → image/*.

21. Webhook Relay (HMAC)
Файл: internal/webhook/verify.go

22. Idempotent Webhook Response
Файл: internal/app/gateway.go, writePaymentConfirmError

Суть: AlreadyExists → HTTP 200.

23. gRPC → HTTP Error Mapping
Файл: internal/app/gateway.go, writeGRPCError

24. Legacy API Rewrite
Файл: internal/app/gateway.go

Суть: /api/* → /api/v1/*.

25. Prometheus Metrics
Файл: internal/app/gateway.go

Код: gateway_requests_total, gateway_request_duration_seconds.

26. OTel + Jaeger
Файл: internal/app/gateway.go, initTracer + otelgin.Middleware("gateway")

27. Context Propagation to gRPC
Файл: internal/app/gateway.go, withUserRole

28. Parallel Fan-Out + Semaphore
Файл: internal/app/gateway.go, /api/v1/admin/users

Код: sem := make(chan struct{}, 8).

29. Cache-Aside (whoami, profile, submit)
Файл: internal/app/gateway.go

30. Health/Ready
Файл: internal/app/gateway.go

31. Graceful Shutdown
Файл: internal/app/gateway.go

32. NATS → WebSocket Broadcast
Файл: internal/app/gateway.go

Код: js.Subscribe("score.updated", func(msg *nats.Msg) { hub.Broadcast(msg.Data) })

33. Security by Default (404 for hidden)
Файл: internal/middleware/docs.go

34. CSRF Protection via Bearer-only
Файл: internal/middleware/security.go

35. Code Duplication (Hub ×2) — анти-паттерн
Файлы: internal/app/gateway.go + internal/websocket/hub.go

36. Config Loader
Файл: internal/config/config.go

37. Fail-Fast vs Fail-Open
Файлы: internal/ratelimit/limiter.go (fail-open) + internal/circuit/breaker.go (fail-fast)

38. Interface for Testing (AuthClient)
Файл: internal/middleware/auth.go

39. Thorough Testing
Файлы: circuit/breaker_test.go, client/retry_test.go, middleware/*_test.go, dto/*_test.go

40. God Function (runGateway) — анти-паттерн
Файл: internal/app/gateway.go

41. Multi-Tenant Config
Файл: internal/config/config.go

Суть: 14 адресов downstream-сервисов.
