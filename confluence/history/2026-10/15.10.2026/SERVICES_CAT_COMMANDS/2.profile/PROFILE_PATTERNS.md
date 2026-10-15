Разбор profile
Что делает сервис
profile — это агрегатор пользовательских данных: профиль (ник, total_score, best_scores, лампы, билетики) + достижения (achievements). Сервис слушает события из NATS (event.user.registered, score.updated, purchase.paid, boost.started) и обновляет профиль + разблокирует достижения. Отдаёт gRPC API для чтения/обновления профиля.

Найденные паттерны
1. DI Container (как в auth)
Файл: internal/app/di.go, тип diContainer.

Почему: Всё те же зависимости в одном месте. Только добавились nc (NATS), js (JetStream), achievements.

Код:

go
type diContainer struct {
    cfg       *config.Config
    dbpool    *pgxpool.Pool
    nc        *nats.Conn
    js        nats.JetStreamContext
    redisRepo *repository.RedisProfileRepo
    repo      *repository.PostgresProfileRepo
    svc       service.ProfileService
    api       pb.ProfileServiceServer
}
2. Pipeline / Init Chain (как в auth)
Файл: internal/app/di.go, функция App.init.

Почему: Список шагов инициализации, включая initSubscriptions в конце — NATS-подписки стартуют после всех остальных.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initPostgres, a.initNATS,
    a.initRedis, a.initDomain, a.initGRPC, a.initMetricsHTTP,
    a.initSubscriptions,   // ← подписки стартуют последними
}
3. Фасад (Facade)
Файл: internal/service/profile_service.go, интерфейс ProfileService.

Почему: ProfileService скрывает работу с тремя подсистемами: ProfileRepository, AchievementRepository, RedisProfileRepo. Хендлер не знает ни про одну из них.

Код:

go
type profileService struct {
    repo         repository.ProfileRepository
    achievements repository.AchievementRepository
    cache        *repository.RedisProfileRepo
    cacheTTL     time.Duration
}
4. Репозиторий (Repository) — два репозитория
Файлы:

internal/repository/profile_repo.go, интерфейс ProfileRepository.

internal/repository/achievement_repo.go, интерфейс AchievementRepository.

Почему: Два отдельных репозитория — профиль и достижения. Это Interface Segregation: сервис может использовать только тот репозиторий, который ему нужен.

Код:

go
type ProfileRepository interface {
    GetProfile(ctx context.Context, userID string) (*UserProfile, error)
    UpsertProfile(ctx context.Context, profile *UserProfile) error
}

type AchievementRepository interface {
    ListUnlocked(ctx context.Context, userID string) ([]Achievement, error)
    Unlock(ctx context.Context, userID string, codes []string) (int64, error)
}
5. Конвертер (Converter / Mapper)
Файл: internal/converter/profile.go, ProfileToProto и ProfileFromUpdateRequest.

Почему: Есть оба направления: domain → proto и proto → domain. Это правильный Mapper.

Код:

go
func ProfileToProto(p *model.Profile) *pb.GetProfileResponse { /* ... */ }
func ProfileFromUpdateRequest(req *pb.UpdateProfileRequest) *model.Profile { /* ... */ }
6. Event-Driven Consumer (Event Subscriber)
Файл: internal/app/di.go, функция initSubscriptions.

Почему: Сервис подписывается на 4 события NATS:

event.user.registered — создаёт профиль.

score.updated — обновляет best_scores / total_score / lamps / tickets + разблокирует достижения.

purchase.paid — разблокирует first_purchase.

boost.started — разблокирует first_boost.

Это Event-Driven Consumer — сервис реагирует на события из шины.

Код:

go
_, err := js.Subscribe("event.user.registered", func(msg *nats.Msg) {
    // создаём профиль
    _ = msg.Ack()
}, nats.Durable("profile-user-registered"), nats.ManualAck())
7. Durable Consumer (Persistent Subscription)
Файл: internal/app/di.go, initSubscriptions.

Почему: Каждая подписка имеет durable name (profile-user-registered, profile-score-updated, etc.). Это гарантирует, что при перезапуске сервиса подписка продолжится с того же места, не потеряет сообщения.

Код:

go
nats.Durable("profile-user-registered")
nats.Durable("profile-score-updated")
nats.Durable("profile-purchase-paid")
nats.Durable("profile-boost-started")
Это Durable Consumer — часть паттерна Event Sourcing / Message Queue.

8. Manual Ack (Ручное подтверждение)
Файл: internal/app/di.go, nats.ManualAck() + _ = msg.Ack().

Почему: Ты явно подтверждаешь обработку сообщения только после успешной обработки. Если обработка упала — сообщение не будет подтверждено и вернётся.

Код:

go
if err := svc.UpdateProfile(ctx, profile); err != nil {
    a.log.Error("failed to upsert profile", "user", event.UserID, "err", err)
    // не ack-аем! сообщение вернётся
} else {
    a.log.Info("profile created", "user", event.UserID)
}
_ = msg.Ack()   // ← ack только после успеха
Правда, у тебя есть нюанс: в некоторых местах ты делаешь _ = msg.Ack() даже при ошибке (например, при unmarshal). Это ок для poison message (сообщение, которое никогда не обработается).

9. Cache-Aside (как в auth)
Файл: internal/service/profile_service.go, GetProfile + UpdateProfile.

Почему: Сначала Redis, при промахе — Postgres, результат в Redis. При обновлении — инвалидация кеша.

Код:

go
func (s *profileService) GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error) {
    if s.cache != nil {
        cached, err := s.cache.GetProfile(ctx, userID)
        if err == nil && cached != nil {
            return cached, nil
        }
    }
    profile, err := s.repo.GetProfile(ctx, userID)
    if err == nil && profile != nil && s.cache != nil {
        s.cache.SetProfile(ctx, profile, s.cacheTTL)
    }
    return profile, nil
}
10. Idempotent Write (ON CONFLICT DO NOTHING)
Файл: internal/repository/achievement_repo.go, Unlock.

Почему: Разблокировка достижения идемпотентна: ON CONFLICT (user_id, code) DO NOTHING. Можно вызывать сколько угодно раз — дубликатов не будет.

Код:

go
INSERT INTO user_achievements (user_id, code, unlocked_at)
SELECT $1, c.code, NOW()
FROM UNNEST($2::text[]) AS c(code)
INNER JOIN achievements a ON a.code = c.code
ON CONFLICT (user_id, code) DO NOTHING
Это Idempotent Write — важный паттерн в event-driven системах.

11. Domain Rule Engine (Rules / Achievements)
Файл: internal/service/achievements.go, функция EvaluateAchievements.

Почему: Это чистая функция, которая берёт bestScores, totalScore, gameID, level и возвращает список кодов достижений. Логика разблокировки инкапсулирована в одном месте. Это Rule Engine в миниатюре.

Код:

go
func EvaluateAchievements(bestScores map[string]int32, totalScore int32, gameID string, level int) []string {
    codes := make([]string, 0, 40)
    if bestScores == nil {
        bestScores = map[string]int32{}
    }
    for _, g := range mainGames {
        if _, ok := bestScores[g]; ok {
            codes = append(codes, "first_play_"+g)
        }
    }
    // ... пороги по играм, по total, по level
    return uniqueStrings(codes)
}
Это Domain Service + Rule Engine.

12. Idempotent Producer / Outbox (в зачатке)
Файл: internal/app/di.go, подписка на purchase.paid и boost.started.

Почему: Профиль реагирует на события, но не публикует свои. Это односторонний поток. Однако если добавишь outbox для profile events (например, achievement.unlocked), это станет полноценным Outbox Pattern.

Пока — это просто Event Consumer.

13. Graceful Degradation
Файлы: initRedis, initNATS — попытки подключения с retry + предупреждение вместо падения.

Почему: Если Redis недоступен — сервис продолжает работать без кеша. Если NATS недоступен — сервис стартует, но подписки не работают. Это Graceful Degradation: сервис не падает целиком из-за отказа одной подсистемы.

Код:

go
if err := redisRepo.Ping(ctx); err != nil {
    a.log.Warn("redis unavailable; profile will run without cache", "err", err)
} else {
    a.log.Info("redis connected", "addr", a.cfg.RedisAddr)
}
14. Config View (как в auth)
Файл: internal/config/config.go — простой Config, но с view-методами. Отличается от auth: здесь нет ConfigProvider с view-интерфейсами, а просто *Config с полями. Это упрощение.

Паттерн: Config Struct (не view-интерфейсы, как в auth).

15. Achievement Unlock Pipeline (пайплайн разблокировки)
Файл: internal/service/achievements.go, EvaluateAndUnlock → UnlockCodes → achievements.Unlock.

Почему: Разблокировка достижений — это пайплайн:

EvaluateAchievements — вычисляет список кодов.

UnlockCodes — вызывает репозиторий.

Unlock — идемпотентно пишет в БД.

Код:

go
func (s *profileService) EvaluateAndUnlock(ctx context.Context, userID string, bestScores map[string]int32, totalScore int32, gameID string, level int) error {
    codes := EvaluateAchievements(bestScores, totalScore, gameID, level)
    return s.UnlockCodes(ctx, userID, codes)
}
Это Pipeline — цепочка шагов обработки.

Итог по profile
Паттерн	Где
DI Container	app/di.go, diContainer
Pipeline / Init Chain	app/di.go, App.init
Facade	service/profile_service.go, profileService
Repository (×2)	repository/profile_repo.go, repository/achievement_repo.go
Converter ( двусторонний)	converter/profile.go
Event-Driven Consumer	app/di.go, initSubscriptions
Durable Consumer	nats.Durable(...)
Manual Ack	nats.ManualAck() + msg.Ack()
Cache-Aside	service/profile_service.go, GetProfile
Idempotent Write	repository/achievement_repo.go, Unlock
Rule Engine / Domain Service	service/achievements.go, EvaluateAchievements
Pipeline (unlock)	EvaluateAndUnlock → UnlockCodes → Unlock
Graceful Degradation	initRedis, initNATS
Config Struct (упрощённый Config View)	config/config.go
Итого 14 паттернов + архитектурных приёмов.

Ключевое отличие от auth: event-driven consumer через NATS JetStream, durable-подписки, manual ack, идемпотентная запись, graceful degradation.

Готовый markdown для PROFILE_PATTERNS.md
md
# profile — Programming Patterns

> Сервис: агрегация профиля пользователя + достижения.
> Postgres + Redis + NATS JetStream (event-driven consumer) + gRPC.

## Что делает сервис

Хранит агрегированный профиль пользователя (ник, total_score, best_scores, лампы, билетики) и достижения. Слушает события из NATS (`event.user.registered`, `score.updated`, `purchase.paid`, `boost.started`) и обновляет профиль + разблокирует достижения. Отдаёт gRPC API для чтения/обновления профиля.

## Паттерны

### 1. DI Container
- **Файл:** `internal/app/di.go`, `diContainer`
- **Суть:** зависимости в одном контейнере, включая NATS (`nc`, `js`).

### 2. Pipeline / Init Chain
- **Файл:** `internal/app/di.go`, `App.init`
- **Суть:** шаги инициализации, `initSubscriptions` стартует последним.

### 3. Facade
- **Файл:** `internal/service/profile_service.go`, `profileService`
- **Суть:** скрывает `ProfileRepository`, `AchievementRepository`, `RedisProfileRepo`.

### 4. Repository (×2)
- **Файлы:** `repository/profile_repo.go`, `repository/achievement_repo.go`
- **Суть:** два независимых репозитория.

### 5. Converter (двусторонний)
- **Файл:** `internal/converter/profile.go`
- **Суть:** `ProfileToProto` + `ProfileFromUpdateRequest`.

### 6. Event-Driven Consumer
- **Файл:** `internal/app/di.go`, `initSubscriptions`
- **Суть:** подписки на 4 события NATS.

### 7. Durable Consumer
- **Файл:** `internal/app/di.go`
- **Суть:** `nats.Durable("profile-user-registered")` и др.

### 8. Manual Ack
- **Файл:** `internal/app/di.go`
- **Суть:** `nats.ManualAck()` + `msg.Ack()` после успешной обработки.

### 9. Cache-Aside
- **Файл:** `internal/service/profile_service.go`, `GetProfile`
- **Суть:** Redis → Postgres → Redis.

### 10. Idempotent Write
- **Файл:** `internal/repository/achievement_repo.go`, `Unlock`
- **Суть:** `ON CONFLICT (user_id, code) DO NOTHING`.

### 11. Rule Engine / Domain Service
- **Файл:** `internal/service/achievements.go`, `EvaluateAchievements`
- **Суть:** чистая функция, вычисляющая коды достижений.

### 12. Pipeline (unlock)
- **Файл:** `internal/service/achievements.go`
- **Суть:** `EvaluateAndUnlock` → `UnlockCodes` → `Unlock`.

### 13. Graceful Degradation
- **Файл:** `internal/app/di.go`, `initRedis`, `initNATS`
- **Суть:** при отказе Redis/NATS сервис не падает.

### 14. Config Struct (упрощённый Config View)
- **Файл:** `internal/config/config.go`
- **Суть:** `*Config` с полями, без view-интерфейсов (в отличие от `auth`).