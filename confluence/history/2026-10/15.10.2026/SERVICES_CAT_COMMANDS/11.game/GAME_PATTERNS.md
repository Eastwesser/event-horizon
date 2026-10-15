Разбор game
Что делает сервис
game — это игровое ядро:

SubmitScore — приём очков от игр, валидация, начисление наград.

Strategy по играм — switch req.GameID с разной логикой для каждой из 8 игр.

Validator — для hexagon и memory есть серверная валидация ходов.

Boost System — StartBoost, ConsumeActiveRunBoost. Буст = 10 ламп.

Outbox — score.updated через outbox (с fallback на прямой NATS publish).

Billing Client — gRPC-вызов SpendCurrency при старте буста.

Prometheus Metrics — game_submits_total, game_score_histogram.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg           config.ConfigProvider
    db            *sql.DB
    nc            *nats.Conn
    js            nats.JetStreamContext
    billingClient billingPb.BillingServiceClient
    repo          repository.GameRepository
    svc           service.GameService
    api           pb.GameServiceServer
}
Тут gRPC-клиент к billing — важно.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer, a.initPostgres,
    a.initNATS, a.initDomain, a.initOutbox, a.initGRPC, a.initMetricsHTTP,
}
initDomain включает grpc.NewClient(billingAddr) — dial к billing.

3. Strategy по играм (главный паттерн)
Файл: internal/service/game_service.go, SubmitScore.

Почему: Вся логика разбита по switch req.GameID. Каждая игра — своя стратегия валидации и наград.

Код:

go
switch req.GameID {
case "hexagon":
    valid, valScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, req.Score)
    // ...
    lampsEarned = 10
    ticketsEarned = validatedScore / 100
case "memory":
    game := memory.NewMemoryGame(req.Seed)
    // ...
    validatedScore = valScore
    lampsEarned, ticketsEarned = game.CalculateRewards(validatedScore)
case "flappy":
    // ...
case "towers":
    // ...
case "hanoi":
    // ...
case "twenty48":
    // ...
case "gears":
    // ...
case "companion":
    // ...
default:
    return &SubmitScoreResponse{Success: false, Message: fmt.Sprintf("unknown game_id: %s", req.GameID)}, nil
}
Strategy Pattern — каждая игра имеет свои правила валидации и наград. Ключевой паттерн game.

4. Domain Validators (hexagon, memory)
Файлы: games/hexagons/validator.go, games/memory/game.go.

Почему: Для двух игр вынесена серверная валидация ходов в отдельные пакеты. Это Domain Services — независимые от инфраструктуры.

Код (hexagon):

go
valid, valScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, req.Score)
Код (memory):

go
game := memory.NewMemoryGame(req.Seed)
var memoryMoves []memory.Move
for _, m := range req.Moves {
    memoryMoves = append(memoryMoves, memory.Move{
        CardIndex1: int(m.FromX),
        CardIndex2: int(m.ToX),
    })
}
valid, valScore, err := game.ValidateMoves(req.Seed, memoryMoves, req.Score)
Domain Validator — важный паттерн для анти-читерства. Уникально для game.

5. Boost System (3-фазный цикл)
Файлы: internal/service/game_service.go, StartBoost + internal/repository/game_repo.go.

Почему: Boost — это 3-фазный цикл:

StartBoost — списать 10 ламп через billing, создать run_boost запись.

GetActiveRunBoost — при SubmitScore найти активный буст.

ConsumeActiveRunBoost — пометить буст использованным.

Код (StartBoost):

go
spendResp, err := s.billingClient.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
    UserId:      userID,
    Currency:    billingPb.CurrencyType_LAMPS,
    Amount:      boostCostLamps,
    Reason:      "game_boost",
    ReferenceId: "boost:" + boostID,
})
if err != nil { /* ... */ }

if err := s.repo.CreateRunBoost(ctx, boostID, userID, gameID); err != nil {
    return nil, status.Errorf(codes.Internal, "record boost: %v", err)
}
Код (ConsumeActiveRunBoost):

go
func (r *PostgresGameRepo) ConsumeActiveRunBoost(ctx context.Context, userID, gameID, boostID string) (string, bool, error) {
    var id string
    var err error
    if boostID != "" {
        err = r.db.QueryRowContext(ctx, `
            UPDATE run_boosts
            SET consumed_at = NOW()
            WHERE id = $1::uuid AND user_id = $2 AND game_id = $3 AND consumed_at IS NULL
            RETURNING id::text`,
            boostID, userID, gameID,
        ).Scan(&id)
    }
    // ...
}
Boost System — важная фича. Плюс: буст можно переиспользовать, если предыдущий не был consumed (existing boost reused).

6. Idempotent Boost Reuse
Файл: internal/service/game_service.go, StartBoost.

Код:

go
// Reuse an unpaid-for active boost (crash / reconnect) without double-charging.
if existing, ok, err := s.repo.GetActiveRunBoost(ctx, userID, gameID); err != nil {
    return nil, status.Errorf(codes.Internal, "lookup boost: %v", err)
} else if ok {
    return &StartBoostResponse{
        BoostID:    existing,
        Boosted:    true,
        Cost:       0,
        NewBalance: -1,
        Message:    "existing boost reused",
    }, nil
}
Idempotent Boost Reuse — если буст уже активен, повторный StartBoost не списывает лампы. Уникально для game.

7. Transactional Outbox с Fallback
Файл: internal/service/game_service.go, SubmitScore.

Почему: Есть два пути публикации score.updated:

Primary — SaveHighscoreAndEnqueueOutbox (транзакция).

Fallback — SaveHighscore + publishScoreUpdatedDirect (прямой NATS publish).

Код:

go
if isNewRecord {
    if err := s.repo.SaveHighscoreAndEnqueueOutbox(ctx, req.UserID, req.GameID, validatedScore, level, "score.updated", eventData); err != nil {
        log.Printf("outbox+highscore failed, falling back to SaveHighscore+Publish: %v", err)
        if saveErr := s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore, level); saveErr != nil {
            log.Printf("Failed to save highscore: %v", saveErr)
        }
        s.publishScoreUpdatedDirect(eventData, req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
    } else {
        log.Printf("📬 Enqueued score.updated (with highscore): user=%s game=%s level=%d score=%d", req.UserID, req.GameID, level, validatedScore)
    }
}
Outbox with Fallback — если outbox не работает (например, миграция не применена), используется прямой publish. Уникально для game.

8. Outbox Worker с FOR UPDATE SKIP LOCKED
Файл: internal/worker/outbox_worker.go.

Код:

go
rows, err := tx.QueryContext(ctx, `
    SELECT id, event_type, payload FROM outbox
    WHERE processed = false
    ORDER BY created_at
    LIMIT 20
    FOR UPDATE SKIP LOCKED
`)
FOR UPDATE SKIP LOCKED — многопоточная безопасность. Как в shop.

9. Prometheus Metrics (Counter + Histogram)
Файл: internal/handler/grpc_handler.go.

Код:

go
var (
    gameSubmitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_submits_total",
            Help: "Total number of score submissions",
        },
        []string{"game_id", "status"},
    )

    gameScoreHistogram = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "game_score_histogram",
            Help:    "Distribution of scores by game",
            Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
        },
        []string{"game_id"},
    )
)
Prometheus Metrics — счётчик + гистограмма. Уникально для game (в billing metrics были в сервисе, здесь — в handler).

10. Converter (Score)
Файл: internal/converter/score.go.

Код:

go
func SubmitScoreFromProto(req *pb.SubmitScoreRequest) model.ScoreSubmission {
    if req == nil {
        return model.ScoreSubmission{}
    }
    return model.ScoreSubmission{
        UserID:   req.UserId,
        GameID:   req.GameId,
        Level:    req.Level,
        Score:    req.Score,
        Nickname: req.Nickname,
    }
}
Простой converter.

11. Mock Object (Tests)
Файл: internal/service/game_service_test.go.

Два мока:

mockGameRepo — для GameRepository.

stubBilling — для billingPb.BillingServiceClient.

failEnqueueRepo — для тестирования fallback.

Три мока — важно.

12. Failover Test
Файл: internal/service/game_service_test.go, TestSubmitScore_MoreGamesAndOutboxFallback.

Почему: Тест проверяет, что при ошибке outbox сервис переходит на fallback (SaveHighscore + publishScoreUpdatedDirect).

Код:

go
failOnce := &failEnqueueRepo{mockGameRepo: mockGameRepo{highscore: 0}}
svc = NewGameService(failOnce, nil, nil)
resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
    UserID: "u", GameID: "flappy", Score: 30, Level: 1,
})
if err != nil || !resp.Success || !failOnce.savedHS {
    t.Fatalf("fallback %+v saved=%v err=%v", resp, failOnce.savedHS, err)
}
Failover Test — тест на деградацию. Редко встречается.

13. Multi-Level Highscores (Flappy)
Файл: internal/repository/game_repo.go.

Почему: Highscores хранятся по (user_id, game_id, level). Только Flappy использует разные уровни, остальные — level=1.

Код:

go
func normalizeHighscoreLevel(level int) int {
    if level < 1 {
        return 1
    }
    return level
}
Multi-Level Highscores — упрощённая система уровней.

14. Graceful Degradation (Billing)
Файл: internal/app/di.go, initDomain.

Код:

go
addr := a.cfg.BillingAddr()
billingConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    a.log.Warn("billing dial failed; boosts disabled", "addr", addr, "err", err)
} else {
    a.di.billingClient = billingPb.NewBillingServiceClient(billingConn)
    // ...
}
Graceful Degradation — если billing недоступен, бусты отключаются, но остальная логика работает.

15. Boost Started Event (Publish)
Файл: internal/service/game_service.go, publishBoostStarted.

Код:

go
func (s *gameService) publishBoostStarted(boostID, userID, gameID string) {
    body, err := json.Marshal(map[string]any{
        "event_uuid": uuid.NewString(),
        "event_type": "BoostStarted",
        "user_uuid":  userID,
        "game_id":    gameID,
        "boost_uuid": boostID,
        "cost_lamps": boostCostLamps,
    })
    s.js.Publish("boost.started", body)
}
Event Publish — событие boost.started публикуется напрямую в NATS (не через outbox). Это fallback без гарантий.

16. Tracer OTel
Файл: internal/app/di.go, initTracer.

Как в leaderboard, shop, billing.

17. Interceptor Chain
Файл: internal/app/di.go, initGRPC.

Код:

go
grpc.ChainUnaryInterceptor(
    interceptor.Recovery(),
    interceptor.Logger(),
    interceptor.Validate(),
)
Стандартный набор.

18. ConfigProvider (View-интерфейсы)
Файл: internal/config/interfaces.go (ты его не скинул, но он есть).

Как в auth, billing, shop.

19. Graceful Shutdown / Closer
Файл: internal/app/di.go.

20. Converter Test
Файл: internal/converter/score_test.go.

Простой тест.

Итог по game
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go
Strategy по играм	service/game_service.go, SubmitScore
Domain Validators (hexagon, memory)	games/hexagons/, games/memory/
Boost System (3-фазный)	service/game_service.go, StartBoost + ConsumeActiveRunBoost
Idempotent Boost Reuse	service/game_service.go, StartBoost
Transactional Outbox с Fallback	service/game_service.go, SubmitScore
Outbox Worker с FOR UPDATE SKIP LOCKED	worker/outbox_worker.go
Prometheus Metrics (Counter + Histogram)	handler/grpc_handler.go
Converter (Score)	converter/score.go
Mock Object (×3)	service/game_service_test.go
Failover Test	service/game_service_test.go
Multi-Level Highscores	repository/game_repo.go
Graceful Degradation (Billing)	app/di.go
Boost Started Event	service/game_service.go, publishBoostStarted
Tracer OTel	app/di.go, initTracer
Interceptor Chain	app/di.go
ConfigProvider	config/interfaces.go
Graceful Shutdown	app/di.go
Converter Test	converter/score_test.go
Итого 20 паттернов и приёмов.

Уникальные фичи game:

Strategy по играм — switch req.GameID — главный паттерн.

Domain Validators — для hexagon и memory есть серверная валидация.

Boost System — 3-фазный цикл с идемпотентным переиспользованием.

Transactional Outbox с Fallback — если outbox не работает, есть прямой publish.

Prometheus Metrics — счётчики и гистограммы в handler.

Failover Test — тест на деградацию.

Готовый markdown для GAME_PATTERNS.md
md
# game — Programming Patterns

> Сервис: игровое ядро (submit score, boost, валидация).
> Postgres + NATS + gRPC (billing) + OTel + Prometheus.

## Что делает сервис

Принимает очки от игр, валидирует ходы (hexagon, memory), начисляет награды (лампы, билетики), управляет бустами. Публикует `score.updated` через Outbox (с fallback на прямой publish). Публикует `boost.started`.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`

### 3. Strategy по играм
- **Файл:** `internal/service/game_service.go`, `SubmitScore`
- **Код:** `switch req.GameID { case "hexagon": ... case "memory": ... }`

### 4. Domain Validators
- **Файлы:** `games/hexagons/validator.go`, `games/memory/game.go`
- **Суть:** серверная валидация ходов для hexagon и memory.

### 5. Boost System (3-фазный)
- **Файл:** `internal/service/game_service.go`, `StartBoost` + `ConsumeActiveRunBoost`

### 6. Idempotent Boost Reuse
- **Файл:** `internal/service/game_service.go`, `StartBoost`
- **Суть:** если буст активен, повторный StartBoost не списывает лампы.

### 7. Transactional Outbox с Fallback
- **Файл:** `internal/service/game_service.go`, `SubmitScore`
- **Код:**
  ```go
  if err := s.repo.SaveHighscoreAndEnqueueOutbox(...); err != nil {
      s.repo.SaveHighscore(...)
      s.publishScoreUpdatedDirect(...)
  }
8. Outbox Worker с FOR UPDATE SKIP LOCKED
Файл: internal/worker/outbox_worker.go

9. Prometheus Metrics
Файл: internal/handler/grpc_handler.go

Код: promauto.NewCounterVec("game_submits_total", ...) + promauto.NewHistogramVec("game_score_histogram", ...)

10. Converter (Score)
Файл: internal/converter/score.go

11. Mock Object (×3)
Файл: internal/service/game_service_test.go, mockGameRepo, stubBilling, failEnqueueRepo

12. Failover Test
Файл: internal/service/game_service_test.go, TestSubmitScore_MoreGamesAndOutboxFallback

13. Multi-Level Highscores
Файл: internal/repository/game_repo.go

Код: normalizeHighscoreLevel(level)

14. Graceful Degradation (Billing)
Файл: internal/app/di.go, initDomain

Суть: если billing недоступен, бусты отключаются.

15. Boost Started Event
Файл: internal/service/game_service.go, publishBoostStarted

16. Tracer OTel
Файл: internal/app/di.go, initTracer

17. Interceptor Chain
Файл: internal/app/di.go

18. ConfigProvider
Файл: internal/config/interfaces.go

19. Graceful Shutdown
Файл: internal/app/di.go

20. Converter Test
Файл: internal/converter/score_test.go
t
