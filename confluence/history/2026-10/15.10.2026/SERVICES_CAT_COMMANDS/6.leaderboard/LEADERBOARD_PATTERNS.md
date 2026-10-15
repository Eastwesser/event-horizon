Разбор leaderboard
Что делает сервис
leaderboard — это топ игроков по играм и уровням. Особенности:

Основное хранилище — Redis (ZSET для сортировки, HASH для метаданных). Postgres — только backup (leaderboard_backup).

Consumer score.updated — обновляет топ при новом рекорде.

Consumer nickname.updated — обновляет ник во всех играх/уровнях.

Publish event.leaderboard.record_beaten — когда кто-то обгоняет #1, публикуется событие (его ловит notification).

Ack Batching — подтверждения NATS собираются в батчи для скорости.

RestoreFromPostgres при старте — восстанавливает Redis из Postgres, если тот упал.

Найденные паттерны
1. DI Container
Файл: internal/app/di.go, diContainer.

Код:

go
type diContainer struct {
    cfg *config.Config
    dbpool *pgxpool.Pool
    nc     *nats.Conn
    js     nats.JetStreamContext
    repo   *repository.RedisLeaderboardRepo
    svc    service.LeaderboardService
    api    pb.LeaderboardServiceServer
    ackCancel context.CancelFunc  // ← уникально
}
Обрати внимание: ackCancel — часть DI-контейнера. Это функция отмены для ack batcher.

2. Pipeline / Init Chain
Файл: internal/app/di.go, App.init.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer,
    a.initPostgres, a.initRedis, a.initDomain,
    a.initNATS, a.initGRPC, a.initMetricsHTTP,
}
Тут есть initTracer — отдельный шаг для OTel/Jaeger. В других сервисах трассировка либо встроена, либо отсутствует.

3. Port (Hexagonal) — LeaderboardRepository
Файл: internal/repository/redis_repo.go, интерфейс LeaderboardRepository.

Код:

go
type LeaderboardRepository interface {
    UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error)
    UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error
    GetTopScores(ctx context.Context, gameID string, limit, level int) ([]ScoreEntry, error)
    GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error)
    SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error
}
Тот же приём, что в history/analytics: сервис общается с внешним миром через порт.

4. Facade — LeaderboardService
Файл: internal/service/leaderboard_service.go.

Код:

go
type leaderboardService struct {
    repo repository.LeaderboardRepository
}
Тонкий фасад с валидацией (limit clamp, gameID/userID required).

5. Event-Driven Consumer (×2)
Файл: internal/app/di.go, initNATS.

Код:

go
js.Subscribe("score.updated", func(msg *nats.Msg) {
    // обновление топа + publish record_beaten
}, nats.Durable("leaderboard-durable"), nats.ManualAck())

js.Subscribe("event.user.nickname.updated", func(msg *nats.Msg) {
    // обновление ника во всех играх/уровнях
}, nats.Durable("leaderboard-nickname"), nats.ManualAck())
Два консьюмера с разными durable names.

6. Durable Consumer
Код:

go
nats.Durable("leaderboard-durable")
nats.Durable("leaderboard-nickname")
7. Manual Ack + Nak (×2)
Файл: internal/app/di.go, initNATS.

В nickname.updated:

go
if err := json.Unmarshal(msg.Data, &event); err != nil || event.UserID == "" || event.Nickname == "" {
    _ = msg.Ack()  // ← poison message — ack, чтобы не застревало
    return
}
// ... обработка ...
_ = msg.Ack()
В score.updated — не Nak, а Ack через batcher (см. ниже).

8. Ack Batcher (уникальный паттерн!)
Файл: internal/app/di.go, initNATS.

Почему это интересно: вместо того чтобы делать msg.Ack() сразу после обработки, ты складываешь сообщения в канал ackChan, а отдельная горутина подтверждает их батчами по 100 штук или раз в 100 мс.

Код:

go
ackChan := make(chan *nats.Msg, 1000)
ticker := time.NewTicker(100 * time.Millisecond)
go func() {
    defer ticker.Stop()
    batch := make([]*nats.Msg, 0, 100)
    flush := func() {
        for _, m := range batch {
            _ = m.Ack()
        }
        batch = batch[:0]
    }
    for {
        select {
        case <-ackCtx.Done():
            flush()
            return
        case msg := <-ackChan:
            batch = append(batch, msg)
            if len(batch) >= 100 {
                flush()
            }
        case <-ticker.C:
            if len(batch) > 0 {
                flush()
            }
        }
    }
}()
Это Ack Batching — паттерн оптимизации: вместо N запросов к NATS делается 1 запрос на N сообщений. Очень редкий и ценный приём.

9. Publish внутри Consumer (Event Chaining)
Файл: internal/app/di.go, initNATS, score.updated.

Почему: Внутри consumer'а score.updated ты публикуешь новое событие event.leaderboard.record_beaten, если новый счёт побил рекорд #1.

Код:

go
if top, topErr := svc.GetTopScores(msgCtx, event.GameID, 1, level); topErr == nil && len(top) > 0 {
    prev := top[0]
    if prev.UserID != "" && prev.UserID != event.UserID && event.Score > prev.Score {
        payload, _ := json.Marshal(map[string]any{
            "event":             "leaderboard.record_beaten",
            "beaten_user_id":    prev.UserID,
            "new_user_id":       event.UserID,
            "new_nickname":      event.Nickname,
            // ...
        })
        if _, pubErr := js.Publish("event.leaderboard.record_beaten", payload); pubErr != nil {
            a.log.Error("publish record_beaten", "err", pubErr)
        }
    }
}
Это Event Chaining / Choreography — событие A вызывает публикацию события B, которое обрабатывает другой сервис (notification). Ключевой паттерн микросервисов.

10. Read-Before-Write (Optimistic Check)
Файл: internal/app/di.go, score.updated.

Почему: Перед обновлением топа ты читаешь текущий топ #1, чтобы понять, побит ли рекорд. Это Read-Before-Write.

Код:

go
if top, topErr := svc.GetTopScores(msgCtx, event.GameID, 1, level); topErr == nil && len(top) > 0 {
    prev := top[0]
    // ...
}
11. Adapter над Redis — RedisLeaderboardRepo
Файл: internal/repository/redis_repo.go.

Почему: RedisLeaderboardRepo инкапсулирует работу с Redis (ZSET, HASH) и автоматически дублирует записи в Postgres (leaderboard_backup) для durability.

Код:

go
func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error) {
    // 1. Пишем в Redis HASH (email, nickname)
    r.client.HSet(ctx, ik, userID+"_email", userEmail)
    // 2. Пишем в Redis ZSET (score)
    r.client.ZAddArgs(ctx, key, redis.ZAddArgs{GT: true, Members: []redis.Z{{Score: float64(score), Member: userID}}})
    // 3. Пишем в Postgres (backup)
    r.db.Exec(ctx, `INSERT INTO leaderboard_backup ...`)
    // 4. Читаем rank
    rank, _ := r.client.ZRevRank(ctx, key, userID).Result()
    return int(rank) + 1, nil
}
Это Write-Through Cache + Backup Storage: Redis — primary, Postgres — durable backup.

12. Write-Through Cache (архитектурный приём)
Файл: internal/repository/redis_repo.go, UpdateScore.

Почему: Обновление идёт сразу в Redis (primary) и в Postgres (backup). Это Write-Through.

Плюс: GT: true в ZAddArgs — обновляет только если новый скор больше текущего. Это Max-Only Write — оптимизация.

13. Restore-on-Startup (Warm-Up Cache)
Файл: internal/app/di.go, initRedis.

Почему: При старте сервиса ты восстанавливаешь Redis из Postgres для всех игр и уровней. Это Cache Warm-Up / Restore-on-Startup.

Код:

go
for _, gameID := range games {
    maxLevel := 1
    if gameID == "flappy" {
        maxLevel = 10
    }
    for level := 1; level <= maxLevel; level++ {
        if err := redisRepo.RestoreFromPostgres(ctx, gameID, level); err != nil {
            a.log.Warn("failed to restore leaderboard", "game", gameID, "level", level, "err", err)
        }
    }
}
Это Restore-on-Startup — важно для leaderboard, потому что Redis может быть не durable.

14. Multi-Level Key Structure
Файл: internal/repository/redis_repo.go, boardKey, infoKey, emailKey.

Почему: Для каждой игры и уровня — три ключа:

leaderboard:{game}:{level} — ZSET с очками.

leaderboard:{game}:{level}:info — HASH с email и nickname.

leaderboard:{game}:{level}:emails — HASH с email (устаревший).

Код:

go
func boardKey(gameID string, level int) string {
    return fmt.Sprintf("leaderboard:%s:%d", gameID, normalizeLevel(level))
}
func infoKey(gameID string, level int) string {
    return boardKey(gameID, level) + ":info"
}
Это Multi-Level Key Structure — паттерн для разнесения данных по ключам.

15. Tracer / OTel Integration (уникально для leaderboard)
Файл: internal/app/di.go, initTracer.

Почему: Только в leaderboard есть явная инициализация OTel-трейсера с Jaeger:

Код:

go
exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint(endpoint),
    otlptracegrpc.WithInsecure(),
)
tp := trace.NewTracerProvider(
    trace.WithBatcher(exporter),
    trace.WithResource(resource.NewWithAttributes(
        semconv.SchemaVersion,
        semconv.ServiceNameKey.String("leaderboard"),
    )),
)
otel.SetTracerProvider(tp)
Это Distributed Tracing — редкий паттерн в твоих сервисах.

16. OTel gRPC StatsHandler
Файл: internal/app/di.go, initGRPC.

Код:

go
grpcServer := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()),  // ← OTel
    grpc.ChainUnaryInterceptor(...),
)
Это OTel gRPC StatsHandler — автоматическое создание span'ов для каждого gRPC-вызова. Уникально для leaderboard (в auth тоже есть, но там OTel импортируется явно).

17. Graceful Shutdown / Closer
Файл: internal/app/di.go.

18. Mock Object (Tests)
Файл: internal/service/leaderboard_service_test.go, mockLB.

Код:

go
type mockLB struct {
    top          []repository.ScoreEntry
    topLimit     int
    rank         int
    score        int
    updateRank   int
    // ...
}
19. Clamp / Normalize (валидация)
Файл: internal/repository/redis_repo.go, normalizeLevel.

Код:

go
func normalizeLevel(level int) int {
    if level < 1 {
        return 1
    }
    if level > 100 {
        return 100
    }
    return level
}
Это Clamp — ограничение значения диапазоном. Также limit в сервисе clamp'ится до 10-100.

20. Coalesce на уровне БД
Файл: internal/repository/redis_repo.go, UpdateScore.

Код:

sql
ON CONFLICT (game_id, level, user_id) DO UPDATE
SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
    user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
    updated_at = NOW()
GREATEST — берём максимум. COALESCE(NULLIF(...)) — обновляем email только если новый не пустой. Это SQL-level Merge с идемпотентностью.

Итог по leaderboard
Паттерн	Где
DI Container	app/di.go
Pipeline / Init Chain	app/di.go, App.init
Port (Hexagonal)	repository/redis_repo.go, LeaderboardRepository
Facade	service/leaderboard_service.go
Event-Driven Consumer (×2)	app/di.go, initNATS
Durable Consumer	nats.Durable("leaderboard-durable"), leaderboard-nickname
Manual Ack + Nak	app/di.go
Ack Batcher	app/di.go, initNATS
Event Chaining / Choreography	app/di.go, publish record_beaten
Read-Before-Write	app/di.go, score.updated
Write-Through Cache	repository/redis_repo.go, UpdateScore
Max-Only Write	ZAddArgs{GT: true}
Restore-on-Startup / Warm-Up Cache	app/di.go, initRedis
Multi-Level Key Structure	repository/redis_repo.go
Distributed Tracing (OTel)	app/di.go, initTracer
OTel gRPC StatsHandler	app/di.go, initGRPC
Clamp / Normalize	repository/redis_repo.go, normalizeLevel
SQL-level Merge (GREATEST + COALESCE)	repository/redis_repo.go, UpdateScore
Graceful Shutdown	app/di.go
Mock Object	service/leaderboard_service_test.go
Итого 20 паттернов и приёмов. Это самый плотный сервис из всех.

Ключевые уникальности:

Ack Batcher — оптимизация ack для NATS.

Event Chaining / Choreography — публикация события из consumer'а.

OTel / Jaeger — единственный сервис с явной трассировкой (помимо auth и gateway).

Restore-on-Startup — warm-up кеша.

Write-Through Cache — Redis + Postgres backup.

SQL-level Merge — GREATEST + COALESCE(NULLIF(...)).

Готовый markdown для LEADERBOARD_PATTERNS.md
md
# leaderboard — Programming Patterns

> Сервис: топ игроков по играм и уровням.
> Redis (primary) + Postgres (backup) + NATS JetStream + gRPC + OTel.

## Что делает сервис

Хранит топ игроков по играм и уровням. Primary — Redis (`ZSET` + `HASH`), backup — Postgres. Реагирует на события `score.updated` и `event.user.nickname.updated`. Публикует `event.leaderboard.record_beaten` при смене #1. Восстанавливает Redis из Postgres при старте.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`
- **Особенность:** содержит `ackCancel context.CancelFunc`.

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`
- **Особенность:** отдельный шаг `initTracer` для OTel.

### 3. Port (Hexagonal)
- **Файл:** `internal/repository/redis_repo.go`, `LeaderboardRepository`

### 4. Facade
- **Файл:** `internal/service/leaderboard_service.go`

### 5. Event-Driven Consumer (×2)
- **Файл:** `internal/app/di.go`, `initNATS`
- **События:** `score.updated`, `event.user.nickname.updated`.

### 6. Durable Consumer
- **Код:** `nats.Durable("leaderboard-durable")`, `leaderboard-nickname`

### 7. Manual Ack + Nak
- **Файл:** `internal/app/di.go`

### 8. Ack Batcher
- **Файл:** `internal/app/di.go`, `initNATS`
- **Суть:** подтверждения собираются в батчи по 100 или раз в 100 мс.
- **Код:**
  ```go
  ackChan := make(chan *nats.Msg, 1000)
  ticker := time.NewTicker(100 * time.Millisecond)
  // ...
  if len(batch) >= 100 {
      flush()
  }
9. Event Chaining / Choreography
Файл: internal/app/di.go, score.updated

Суть: внутри consumer'а публикуется event.leaderboard.record_beaten.

10. Read-Before-Write
Файл: internal/app/di.go, score.updated

Суть: читаем топ-1, чтобы понять, побит ли рекорд.

11. Write-Through Cache
Файл: internal/repository/redis_repo.go, UpdateScore

Суть: обновление идёт и в Redis, и в Postgres.

12. Max-Only Write
Файл: internal/repository/redis_repo.go

Код: ZAddArgs{GT: true, ...}

13. Restore-on-Startup
Файл: internal/app/di.go, initRedis

Суть: загружаем Redis из Postgres при старте.

14. Multi-Level Key Structure
Файл: internal/repository/redis_repo.go

Ключи: leaderboard:{game}:{level}, ...:info.

15. Distributed Tracing (OTel)
Файл: internal/app/di.go, initTracer

16. OTel gRPC StatsHandler
Файл: internal/app/di.go, initGRPC

Код: grpc.StatsHandler(otelgrpc.NewServerHandler())

17. Clamp / Normalize
Файл: internal/repository/redis_repo.go, normalizeLevel

18. SQL-level Merge
Файл: internal/repository/redis_repo.go, UpdateScore

Код:

sql
SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
    user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email)
19. Graceful Shutdown
Файл: internal/app/di.go

20. Mock Object
Файл: internal/service/leaderboard_service_test.go, mockLB