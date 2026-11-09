🔴 Открытие №1 — Nginx уже ретраит, но криво
Смотри nginx.conf:

nginx
upstream balancer_backend {
    server balancer:8079;
    server balancer-2:8079;
    keepalive 32;
}
Здесь нет proxy_next_upstream. Это значит:

Дефолт Nginx — proxy_next_upstream error timeout. То есть connection error / timeout — ретраит на второй балансер. ✅

HTTP 502 / 503 / 504 — НЕ ретраит (по дефолту proxy_next_upstream не включает эти коды). ❌

max_fails / fail_timeout — по дефолту 1 / 10s. То есть если балансер вернул error один раз — он помечается dead на 10 секунд. Но error — это только connection error, не 502.

Плюс proxy_set_header Connection "" (пустая строка) — это правильный трюк для keepalive к upstream (не close, не upgrade). ОК.

Плюс отсутствует ; в конце блока location /:

nginx
proxy_set_header Connection ""
Нет точки с запятой. Это синтаксическая ошибка, Nginx с таким конфигом не стартует. Nginx упадёт на старте, контейнер в restart loop. Проверь — возможно, на проде конфиг другой, а этот — черновик. Или Nginx сейчас в crashloop. P0, проверь docker logs deployments-nginx.

Что это значит для балансера:

Nginx ретраит connection error на второй балансер (между балансерами).

Балансер не ретраит connection error на другой gateway (внутри одного балансера).

Если gateway-2 мёртв, а балансер выбрал его — 502. Nginx видит 502 — не ретраит (не входит в дефолтный proxy_next_upstream). Клиент получает 502.

Вывод: без ретрая в балансере и без proxy_next_upstream http_502 в Nginx — один мёртвый gateway даёт 502 клиенту. Это ровно то, что ты описал в TODO.

🔴 Открытие №2 — у Gateway уже есть Retry Interceptor
Смотри services/gateway/internal/client/retry.go:

go
func UnaryRetryInterceptor(cfg RetryConfig) grpc.UnaryClientInterceptor {
    // ...
    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        last = invoker(ctx, method, req, reply, cc, opts...)
        if last == nil { return nil }
        if !isRetryable(last) || attempt == cfg.MaxAttempts-1 { return last }
        if err := waitBackoff(ctx, cfg, attempt); err != nil { return last }
    }
}
DefaultRetryConfig: MaxAttempts: 3, BaseDelay: 40ms, MaxDelay: 400ms, full jitter. Ретраит Unavailable и ResourceExhausted.

То есть gateway уже ретраит gRPC-ошибки. Это третий слой ретрая (после Nginx → балансер). И это правильно — gRPC-уровень, где есть идемпотентность через circuit breaker + ретрай interceptor.

А вот RequireAuth НЕ обёрнут в breaker/retry — он зовёт ValidateToken напрямую:

go
resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
Это на каждый закрытый запрос. Если Auth тормозит — все запросы тормозят. Не про ретрай, но связано. P2.

🔴 Открытие №3 — Circuit Breaker + Bulkhead уже есть
internal/circuit/breaker.go — полноценный CB с bulkhead (MaxConcurrent). В runGateway:

go
func newServiceBreaker(name string) *circuit.Breaker {
    return circuit.New(circuit.Settings{
        Name:          name,
        MaxRequests:   3,        // half-open probes
        Timeout:       10 * time.Second,
        MaxConcurrent: 32,       // bulkhead
        ReadyToTrip: func(counts circuit.Counts) bool {
            return counts.ConsecutiveFailures >= 5
        },
    })
}
Значит мой аргумент «ретраить 5xx на POST не надо, это работа CB + идемпотентности» — подтверждается кодом. CB в gateway ловит ошибки downstream-сервисов и делает fail-fast 503. Балансеру лезть в 5xx не нужно.

🔴 Открытие №4 — Gateway создаёт Stream EVENTS сам
В runGateway:

go
_, err = js.AddStream(&nats.StreamConfig{
    Name:     "EVENTS",
    Subjects: []string{"event.>", "score.updated"},
    Storage:  nats.FileStorage,
})
Но по доке NATS Hub — единственный, кто создаёт Stream. А gateway тоже пытается. Это дублирование Bootstrap. AddStream вернёт «already exists» — не критично, но архитектурно грязно. P2 (не про ретрай, но в копилку «God Function»).

🔴 Открытие №5 — WS Hub дублируется
Ты уже отметил это в TODO. Подтверждаю:

internal/app/gateway.go → type Hub struct { ... } (inline, ~60 строк)

internal/websocket/hub.go → type Hub struct { ... } (отдельный файл, ~60 строк)

Оба не используются одновременно. В gateway.go используется inline NewHub(), а internal/websocket/hub.go — мёртвый код. P3.

🔴 Открытие №6 — writeGRPCError не мапит AlreadyExists на 409 для webhook
Смотри writePaymentConfirmError:

go
func writePaymentConfirmError(c *gin.Context, err error) {
    if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
        c.JSON(http.StatusOK, gin.H{"success": true, "message": st.Message()})
        return
    }
    writeGRPCError(c, err)
}
Это ок для webhook. Но в writeGRPCError codes.Aborted мапится в 409, а AlreadyExists — в 409. Хорошо.

🔴 Открытие №7 — RequireAuth в POST /api/v1/game/submit читается дважды
Смотри:

go
r.POST("/api/v1/game/submit", middleware.RequireAuth(authClient), func(c *gin.Context) {
    body, _ := c.GetRawData()
    c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

    cacheKey := string(body)
    if cached, ok := scoreCache.Get(cacheKey); ok {
        c.Data(http.StatusOK, "application/json", cached)
        return
    }
    // ...
    userInfo, authErr := authClient.GetClient().GetUser(...)  // ← ещё один вызов Auth!
RequireAuth уже сделал ValidateToken (который ходит в Redis), а тут — ещё один GetUser в Auth для ника. Итого 2 вызова Auth на каждый submit. Плюс scoreCache — 2 секунды TTL по сырому body, а не по user_id + game_id + score. То есть два разных юзера с одинаковым body → один кэш. P2, но по сути — баг кэша.

🔴 Открытие №8 — .openapi.yaml в gateway НЕ полная
Ты сам заметил. Сравнил с твоей основной (в docs/openapi.yaml). В gateway-версии НЕ хватает:

Чего нет	В основной есть
POST /api/v1/admin/shop/purchases/cancel	✅
POST /api/v1/admin/payments/grant-subscription	✅
POST /api/v1/admin/payments/revoke-subscription	✅
GET /api/v1/auth/user → role в ответе	❌ (в основной нет, но в коде есть)
GET /api/v1/authors/me — описание про soft-deleted	⚠️ в gateway нет описания, в основной есть (P0-6)
GET /api/v1/authors/{user_id} — описание 404 для soft-deleted	⚠️ нет в gateway
POST /api/v1/inventory/items/bulk — idempotency_key в attributes	⚠️ тест есть, спека нет
GET /api/v1/leaderboard — nickname в ответе	⚠️ в коде DTO ScoreEntry есть nickname, в спеке нет
GET /api/v1/notifications/{id}/read — id=all	⚠️ в gateway нет упоминания id=all
GET /api/v1/analytics/dau — параметр days	⚠️ в gateway нет
GET /api/v1/analytics/mau — параметр days	⚠️ в gateway нет
GET /api/v1/analytics/retention — cohort_days_ago, window_days	⚠️ в gateway нет
GET /api/v1/inventory/items — include_deleted	⚠️ в gateway нет (а в коде есть!)
POST /api/v1/inventory/items — version в UpdateItem	⚠️ в gateway нет, а в коде req.Version есть
Плюс расхождения:

POST /api/v1/auth/register — в gateway password.minLength: 6, а в коде Auth ≥ 8 (по доке Auth). Спека врёт.

GET /api/v1/authors/me — в gateway закомментирован блок и добавлен второй get: с GetMyAuthor. В основной — уже без комментария, с описанием P0-6.

GET /api/v1/leaderboard — в gateway game_id: enum [hexagon, flappy, memory, towers], а по README игр 8 (hexagon, flappy, memory, towers, hanoi, 2048, gears, tamagotchi). Спека врёт по games.

POST /api/v1/game/submit — то же, enum на 4 игры.

POST /api/v1/shop/purchase — в gateway 409 описан как «Conflict (already owned, insufficient tickets)». В коде — subscription_required → 403. Уже исправлено в gateway, ок.

Вывод: gateway-версия OpenAPI — устаревшая и неполная. Если она нужна для /docs (swagger-ui) — надо синхронизировать с docs/openapi.yaml. Или, проще, — читать из docs/openapi.yaml через //go:embed (если он попадёт в образ). Сейчас api.OpenAPIYAML — это embed из services/gateway/api/openapi.yaml. Можно заменить на симлинк/копию при сборке. P1.

🔴 Открытие №9 — README врёт про количество игр
В README — 8 игр. В openapi.yaml — 4 (hexagon, flappy, memory, towers). В game/submit — 4 в enum. Спека рассинхронизирована с реальностью. Не про ретрай, но при sync OpenAPI надо поправить.

🎯 Обновлённый план (с учётом всего)
P0 (этот заход)
Nginx: proxy_next_upstream error timeout http_502 http_503 http_504; + proxy_next_upstream_tries 2; для location / и location /ws/.

Это самый дешёвый ретрай на уровне, который уже есть.

Не требует пересборки балансера.

Ретраит любые запросы, включая POST — но только при connection error, когда бэкенд точно не обработал (Nginx видит connection error до отправки ответа). Это безопасно.

Для 502/503/504 — есть риск (балансер мог отдать 502 от мёртвого gateway, но gateway мог успеть обработать и умереть). Но на практике: 502 = connection error от балансера к gateway (через ErrorHandler), gateway не обработал. Ретрай 502 безопасен. 503 = circuit breaker открыт — бессмысленно ретраить, только если второй балансер к другому gateway. 504 — не ретраить (таймаут мог произойти после обработки).

Вердикт: proxy_next_upstream error timeout http_502; — ретраить connection error + 502. Не ретраить 503/504.

Плюс — исправить ; в location /.

Балансер: health checks бэкендов.

Горутина раз в 5s пингует /ready (не /health — /ready проверяет Redis, health — только процесс) каждого бэкенда.

atomic.Bool Healthy на Backend.

getLeastConnBackend пропускает !Healthy.

Метрика balancer_backend_healthy{backend="..."} 0|1.

Без этого Nginx будет ретраить 502 на второй балансер, а тот тоже выберет мёртвый gateway (least connections = 0). Порочный круг.

Балансер: retry на GET/HEAD/OPTIONS при connection error.

Свой ErrorHandler → пишет в statusRecorder → если ошибка и метод идемпотентный → повтор на другой бэкенд.

MaxAttempts = 2 (1 попытка + 1 ретрай).

POST/PUT/DELETE — не ретраить (нет Idempotency-Key в API, подтверждено OpenAPI).

Почему в балансере, если Nginx уже ретраит 502? Потому что Nginx ретраит 502 → второй балансер → а тот выберет тот же мёртвый gateway (least connections = 0). Nginx не знает, какой gateway мёртв. Только балансер с health checks знает. Это дублирующий слой, но нужный, потому что Nginx не различает gateway.

Альтернатива: не делать retry в балансере, а сделать только health checks. Тогда Nginx делает retry 502 на второй балансер, а второй балансер с health checks не выберет мёртвый gateway. Это чище. Но требует, чтобы health checks были синхронизированы между двумя балансерами (оба пингуют бэкенды). Это ок — оба пингуют независимо, оба знают, кто мёртв.

Мой выбор: health checks + Nginx proxy_next_upstream http_502. Балансерный retry не нужен — Nginx уже ретраит 502, а health checks гарантируют, что второй балансер выберет живого. Меньше кода, тот же результат. Если хочешь — можно добавить, но это второй слой без явной пользы.

Балансер: slog вместо log.Printf в hot path. Плюс atomic.LoadInt32 в логе. Плюс backend.URL в ErrorHandler.

Балансер: /ready осмысленный — проверяет, есть ли хотя бы один healthy backend. Сейчас всегда 200 ready. Это влияет на Nginx? Нет (Nginx не пингует /ready, только depends_on в compose). Но влияет на внешний мониторинг. P2.

Compose: healthcheck для balancer и balancer-2. Иначе Nginx может подняться раньше, чем балансер начнёт слушать :8079, и первые запросы уйдут в 502. P2.

P1 (следующим заходом)
Sync services/gateway/api/openapi.yaml с docs/openapi.yaml. Варианты:

(а) //go:embed ../../../docs/openapi.yaml — нельзя, go:embed не ходит выше пакета.

(б) cp docs/openapi.yaml services/gateway/api/openapi.yaml в Makefile перед сборкой.

(в) Симлинк.

(г) Оставить две копии, но синхронизировать вручную.

Мой выбор: (б) — Makefile таргет sync-openapi перед docker-build. Плюс CI-проверка diff docs/openapi.yaml services/gateway/api/openapi.yaml.

Заодно починить: password.minLength: 8 (не 6), enum games = 8 игр, include_deleted в inventory, nickname в leaderboard, days в analytics, id=all в notifications, version в UpdateItem, idempotency_key в bulk, GetMyAuthor описание, /{user_id} 404 для soft-deleted, POST /admin/shop/purchases/cancel, POST /admin/payments/grant-subscription, POST /admin/payments/revoke-subscription.

P2
RequireAuth обернуть в CB/retry (сейчас голый вызов Auth).

game/submit: убрать двойной вызов Auth — RequireAuth уже знает email, но не nickname. Можно добавить nickname в ValidateTokenResponse (сейчас есть только user_id, email, role). P2, требует правки proto Auth.

scoreCache — ключ по user_id + game_id + score, не по сырому body.

P3
Убрать Backend.Mutex (мёртвое поле).

Убрать internal/websocket/hub.go (мёртвый код, дублирует inline Hub).

X-Balancer-Backend — сейчас request header, а не response. Если нужен для debug — w.Header().Set до Proxy.ServeHTTP.

Итог: мой ответ на «куда ставим ретрай»
Двухслойно, но без retry в балансере:

Nginx proxy_next_upstream error timeout http_502; — connection error + 502 → второй балансер. Это и есть ретрай. Не нужен свой retry в балансере.

Балансер health checks — гарантирует, что второй балансер не выберет мёртвый gateway.

Retry в балансере на GET при connection error — не нужен, потому что Nginx уже ретраит на другой балансер, а health checks гарантируют, что тот выберет живого. Меньше кода, тот же результат.

Если хочешь третий слой (retry в балансере) — можно, но это избыточно. Я бы не делал.