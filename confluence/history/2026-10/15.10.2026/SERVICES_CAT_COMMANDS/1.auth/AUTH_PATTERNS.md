Разбор auth
Что делает сервис
auth — это управление идентификацией: регистрация, логин, refresh, logout, whoami, validate, update role/nickname, list users. Использует Postgres (пользователи) + Redis (кеш + сессии) + JWT (access + refresh). Отдаёт gRPC API.

Найденные паттерны
1. Dependency Injection Container (DI-контейнер)
Файл: internal/app/di.go, тип diContainer + функция newDiContainer.

Почему: Все зависимости (dbpool, redisRepo, userRepo, svc, api) хранятся в одном контейнере, а init() собирает их пошагово. Хендлер не знает, откуда берётся сервис — он получает его через конструктор NewAuthHandler(a.di.svc).

Код:

go
type diContainer struct {
    cfg       config.ConfigProvider
    dbpool    *pgxpool.Pool
    redisRepo *repository.RedisAuthRepo
    userRepo  repository.UserRepository
    svc       service.AuthService
    api       pb.AuthServiceServer
}
Это классический DI-контейнер. Ты собираешь приложение как конструктор.

2. Конструктор с шагами (Pipeline / Init Chain)
Файл: internal/app/di.go, функция App.init.

Почему: Инициализация разбита на упорядоченные шаги, каждый из которых — отдельная функция. Если шаг упал — приложение не стартует. Это Pipeline (конвейер инициализации).

Код:

go
func (a *App) init(ctx context.Context) error {
    steps := []func(context.Context) error{
        a.initLogger,
        a.initCloser,
        a.initTracer,
        a.initPostgres,
        a.initRedis,
        a.initDomain,
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
Это Chain of Responsibility для старта. Ты называешь это «инициализация», но по сути это Pipeline из шагов, каждый из которых может прервать цепочку.

3. Фасад (Facade)
Файл: internal/service/auth_service.go, тип authService + интерфейс AuthService.

Почему: Хендлер (grpc_handler.go) не знает ни про UserRepository, ни про RedisAuthRepo, ни про jwtauth.Manager. Он дёргает h.authService.Register(...), и всё. Сервис — это фасад над тремя подсистемами: репозиторий, кеш, JWT.

Код:

go
type authService struct {
    repo   repository.UserRepository
    cache  *repository.RedisAuthRepo
    tokens *jwtauth.Manager
}
Классический Facade: единый интерфейс к сложной подсистеме.

4. Репозиторий (Repository)
Файл: internal/repository/user_repo.go, интерфейс UserRepository.

Почему: Интерфейс UserRepository описывает что нужно (Create, GetByEmail, GetByID, UpdateNickname, UpdateRole, GetUserScores, ListUsers), а PostgresUserRepo — как это делается через Postgres. Сервис зависит от интерфейса, не от Postgres.

Код:

go
type UserRepository interface {
    Create(ctx context.Context, email, passwordHash, role string) (string, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    // ...
}

type PostgresUserRepo struct {
    db *pgxpool.Pool
}
5. Конвертер (Converter / Mapper)
Файл: internal/converter/user.go.

Почему: Функция UserToProto переводит доменную модель (model.User) в proto-ответ (pb.GetUserResponse). Это изолирует хендлер от деталей модели и от gRPC-структур.

Код:

go
func UserToProto(u *model.User) *pb.GetUserResponse {
    if u == nil {
        return nil
    }
    return &pb.GetUserResponse{
        UserId:   u.ID,
        Email:    u.Email,
        Nickname: u.Nickname,
        Role:     u.Role,
    }
}
6. Стратегия (Strategy) — JWT
Файл: internal/jwt/manager.go, интерфейсы Generator и Validator.

Почему: Ты определил два интерфейса:

Generator — умеет генерировать токены (access + refresh).

Validator — умеет валидировать токены.

Это Стратегия: можно подменить реализацию (например, RS256 вместо HS256, или вообще OAuth-провайдер). Сервис зависит от интерфейсов, а не от Manager.

Код:

go
type Generator interface {
    GenerateAccessToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
    GenerateRefreshToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
}

type Validator interface {
    ValidateAccessToken(token string) (*Claims, error)
    ValidateRefreshToken(token string) (*Claims, error)
}
Правда, в di.go ты создаёшь *jwtauth.Manager, а не интерфейс. Так что это заготовка под Стратегию, но пока не используется на полную. Если ты передашь tokens как Generator + Validator — паттерн заиграет.

7. Адаптер (Adapter) — над golang-jwt
Файл: internal/jwt/manager.go, тип Manager.

Почему: Manager оборачивает библиотеку github.com/golang-jwt/jwt/v5 и предоставляет свой API (GenerateAccessToken, ValidateAccessToken). Наружу не торчат jwt.Token, jwt.RegisteredClaims — только твой Claims.

Код:

go
type Manager struct {
    secret     []byte
    accessTTL  time.Duration
    refreshTTL time.Duration
}
Это Адаптер: ты приспособил чужую библиотеку под свой интерфейс.

8. Middleware Chain (цепочка интерсепторов)
Файл: internal/app/di.go, initGRPC.

Почему: gRPC-сервер собирает цепочку unary-интерсепторов:

go
grpc.ChainUnaryInterceptor(
    interceptor.Recovery(),
    interceptor.Logger(),
    interceptor.Validate(),
    metrics.UnaryServerInterceptor("auth"),
)
Каждый интерсептор — отдельный слой: recovery → logging → validation → metrics. Это Chain of Responsibility (или Middleware Chain).

Код (интерсептор Validate):

go
func Validate() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if v, ok := req.(protoValidator); ok {
            if err := v.Validate(); err != nil {
                return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
            }
        }
        return handler(ctx, req)
    }
}
9. Cache-Aside (паттерн кеширования)
Файл: internal/service/auth_service.go, функция GetUser.

Почему: Сначала идём в Redis, если промах — идём в Postgres и записываем результат в Redis. Это классический Cache-Aside.

Код:

go
func (s *authService) GetUser(ctx context.Context, userID string) (*model.User, error) {
    if s.cache != nil {
        cached, err := s.cache.GetUserCache(ctx, userID)
        if err == nil {
            return cached, nil   // ← попадание
        }
        if !errors.Is(err, repository.ErrCacheMiss) {
            log.Printf("auth: Redis GetUserCache failed: %v", err)
        }
    }
    user, err := s.repo.GetByID(ctx, userID)   // ← промах, идём в БД
    if err != nil {
        return nil, err
    }
    if user != nil && s.cache != nil {
        if cacheErr := s.cache.SetUserCache(ctx, user); cacheErr != nil {   // ← кладём в кеш
            log.Printf("auth: Redis SetUserCache failed: %v", cacheErr)
        }
    }
    return user, nil
}
10. Session Store (хранилище сессий на Redis)
Файл: internal/repository/redis_repo.go, функции CreateSession, SessionExists, DeleteSession, DeleteAllSessionsForUser.

Почему: Ты хранишь активные JWT (по jti) в Redis, чтобы иметь возможность отозвать токен до истечения exp. Это не паттерн из «банды четырёх», но это устоявшийся архитектурный приём — stateful-сессии поверх stateless-JWT.

Код:

go
func (r *RedisAuthRepo) CreateSession(ctx context.Context, jti, userID string, ttl time.Duration) error {
    pipe := r.client.TxPipeline()
    pipe.Set(ctx, sessionKey(jti), userID, ttl)
    if userID != "" {
        pipe.SAdd(ctx, userSessionSetKey(userID), jti)
        pipe.Expire(ctx, userSessionSetKey(userID), ttl)
    }
    _, err := pipe.Exec(ctx)
    return err
}
Это Session Store — можно называть его так.

11. Value Object через View-интерфейсы (ConfigProvider)
Файл: internal/config/interfaces.go, интерфейс ConfigProvider + view-типы.

Почему: Ты не отдаёшь наружу Config со всеми полями. Ты отдаёшь узкие view-интерфейсы (GRPCConfig, MetricsConfig, PostgresConfig, RedisConfig, JWTConfig, LoggerConfig). Каждый потребитель видит только то, что ему нужно.

Код:

go
type ConfigProvider interface {
    GRPCConfig() GRPCConfig
    MetricsConfig() MetricsConfig
    PostgresConfig() PostgresConfig
    // ...
}

type GRPCConfig interface {
    Port() string
}
Это Interface Segregation (принцип SOLID) в чистом виде. Можно назвать Config View или Narrow Interface.

12. Graceful Shutdown через Closer
Файл: internal/app/di.go, везде a.closer.AddNamed(...) + App.RunUntilSignal.

Почему: Каждый ресурс (Postgres pool, Redis, gRPC server, metrics HTTP) регистрирует свой close-метод в closer. При получении SIGINT/SIGTERM все они закрываются по цепочке. Это реестр закрывашек — можно назвать Closer Registry или Graceful Shutdown Pattern.

Код:

go
a.closer.AddNamed("postgres pool", func(context.Context) error {
    dbpool.Close()
    return nil
})
13. SQL Builder (собственный)
Файл: internal/repository/user_repo.go использует pkg/sqb.

Почему: Ты используешь собственный pkg/sqb (Select, Insert, Update, Where, Set) для сборки SQL. Это Builder (строитель) — паттерн из GoF.

Код:

go
query, args, err := sqb.Insert("users").
    Columns("email", "password_hash", "nickname", "role").
    Values(email, passwordHash, nickname, role).
    Suffix("RETURNING id").
    ToSql()
Итог по auth
Паттерн	Где
DI Container	app/di.go, diContainer
Pipeline / Init Chain	app/di.go, App.init
Facade	service/auth_service.go, authService
Repository	repository/user_repo.go, UserRepository
Converter	converter/user.go, UserToProto
Strategy (заготовка)	jwt/manager.go, Generator / Validator
Adapter	jwt/manager.go, Manager над golang-jwt
Middleware Chain	app/di.go, ChainUnaryInterceptor
Cache-Aside	service/auth_service.go, GetUser
Session Store	repository/redis_repo.go, CreateSession
Config View / Interface Segregation	config/interfaces.go, ConfigProvider
Closer Registry / Graceful Shutdown	app/di.go, closer.AddNamed
SQL Builder	repository/user_repo.go, sqb.Insert
Итого 13 паттернов и архитектурных приёмов. Ты написал auth с плотной архитектурой — теперь у тебя есть словарь для описания того, что ты сделал.

Готовый markdown для AUTH_PATTERNS.md
md
# auth — Programming Patterns

> Сервис: аутентификация (register, login, refresh, logout, whoami, validate, update role/nickname, list users).
> Postgres + Redis + JWT (access + refresh) + gRPC.

## Что делает сервис

Управляет идентификацией пользователя: регистрация, логин, выдача access/refresh токенов, отзыв сессий, кеш пользователей. Хранит пользователей в Postgres, активные JWT (`jti`) в Redis для возможности отзыва. Отдаёт gRPC API.

## Паттерны

### 1. DI Container

- **Файл:** `internal/app/di.go`, тип `diContainer`
- **Суть:** все зависимости собраны в одном контейнере; `App.init` собирает их пошагово.
- **Код:**
  ```go
  type diContainer struct {
      cfg       config.ConfigProvider
      dbpool    *pgxpool.Pool
      redisRepo *repository.RedisAuthRepo
      userRepo  repository.UserRepository
      svc       service.AuthService
      api       pb.AuthServiceServer
  }
2. Pipeline / Init Chain
Файл: internal/app/di.go, функция App.init

Суть: инициализация — упорядоченный список шагов; падение шага прерывает старт.

Код:

go
steps := []func(context.Context) error{
    a.initLogger, a.initCloser, a.initTracer,
    a.initPostgres, a.initRedis, a.initDomain,
    a.initGRPC, a.initMetricsHTTP,
}
3. Facade
Файл: internal/service/auth_service.go, тип authService

Суть: хендлер не знает про репозиторий, кеш, JWT — он видит только AuthService.

Код:

go
type authService struct {
    repo   repository.UserRepository
    cache  *repository.RedisAuthRepo
    tokens *jwtauth.Manager
}
4. Repository
Файл: internal/repository/user_repo.go, интерфейс UserRepository

Суть: интерфейс описывает «что», PostgresUserRepo — «как».

Код:

go
type UserRepository interface {
    Create(ctx context.Context, email, passwordHash, role string) (string, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    // ...
}
5. Converter (Mapper)
Файл: internal/converter/user.go, UserToProto

Суть: переводит доменную модель в proto-ответ; хендлер не знает о модели.

Код:

go
func UserToProto(u *model.User) *pb.GetUserResponse {
    return &pb.GetUserResponse{UserId: u.ID, Email: u.Email, /* ... */ }
}
6. Strategy (заготовка) — JWT
Файл: internal/jwt/manager.go, интерфейсы Generator и Validator

Суть: можно подменить реализацию (HS256 → RS256, OAuth и т.п.).

Код:

go
type Generator interface {
    GenerateAccessToken(userID, email, role string) (string, string, time.Time, error)
    GenerateRefreshToken(userID, email, role string) (string, string, time.Time, error)
}
7. Adapter — над golang-jwt
Файл: internal/jwt/manager.go, тип Manager

Суть: оборачивает github.com/golang-jwt/jwt/v5, наружу отдаёт только свой API.

Код:

go
type Manager struct {
    secret     []byte
    accessTTL  time.Duration
    refreshTTL time.Duration
}
8. Middleware Chain
Файл: internal/app/di.go, initGRPC

Суть: цепочка unary-интерсепторов: recovery → logger → validate → metrics.

Код:

go
grpc.ChainUnaryInterceptor(
    interceptor.Recovery(),
    interceptor.Logger(),
    interceptor.Validate(),
    metrics.UnaryServerInterceptor("auth"),
)
9. Cache-Aside
Файл: internal/service/auth_service.go, GetUser

Суть: сначала Redis, при промахе — Postgres, результат кладём в Redis.

Код:

go
cached, err := s.cache.GetUserCache(ctx, userID)
if err == nil { return cached, nil }
user, err := s.repo.GetByID(ctx, userID)
// ...
s.cache.SetUserCache(ctx, user)
10. Session Store
Файл: internal/repository/redis_repo.go, CreateSession / SessionExists / DeleteSession

Суть: активные JWT (jti) хранятся в Redis для возможности отзыва до exp.

Код:

go
pipe.Set(ctx, sessionKey(jti), userID, ttl)
pipe.SAdd(ctx, userSessionSetKey(userID), jti)
11. Config View / Interface Segregation
Файл: internal/config/interfaces.go, ConfigProvider

Суть: каждый потребитель видит только нужный кусок конфига (GRPCConfig, JWTConfig, ...).

Код:

go
type ConfigProvider interface {
    GRPCConfig() GRPCConfig
    MetricsConfig() MetricsConfig
    // ...
}
12. Closer Registry / Graceful Shutdown
Файл: internal/app/di.go, closer.AddNamed(...)

Суть: каждый ресурс регистрирует свой close-метод; всё закрывается по сигналу.

Код:

go
a.closer.AddNamed("postgres pool", func(context.Context) error {
    dbpool.Close()
    return nil
})
13. SQL Builder
Файл: internal/repository/user_repo.go, pkg/sqb

Суть: SQL собирается через builder: .Insert(...).Columns(...).Values(...).ToSql().

Код:

go
query, args, err := sqb.Insert("users").
    Columns("email", "password_hash", "nickname", "role").
    Values(email, passwordHash, nickname, role).
    Suffix("RETURNING id").
    ToSql()