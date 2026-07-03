Сегодня Jaeger + k6 (нагрузка). DevOps (бэкапы, CI/CD, nginx, k8s) — оставляем на следующие дни.

План на сегодня:

Jaeger — добавляем трейсинг во все 5 сервисов
k6 — пишем нагрузочный тест на 4 игры
Тестируем — прогоняем нагрузку, смотрим метрики в Grafana и трейсы в Jaeger
🧩 Шаг 1: Jaeger — добавляем трейсинг во все сервисы

1.1. Обновляем go.mod во всех сервисах:

bash
cd /home/denismatveev/event_horizon/services/auth
go get go.opentelemetry.io/otel@v1.34.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
go get go.opentelemetry.io/otel/sdk@v1.34.0
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@v0.59.0
go mod tidy

# Повторить для billing, game, gateway, leaderboard
1.2. Добавляем трейсинг в каждый main.go

Для gRPC-сервисов (Auth, Billing, Game, Leaderboard):

go
// Добавить в импорты
import (
    // ... существующие
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Функция initTracer (уже есть, просто проверь)
func initTracer(ctx context.Context) (func(context.Context) error, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("auth"), // меняем на свой
        )),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp.Shutdown, nil
}

// В main:
func main() {
    ctx := context.Background()
    shutdown, err := initTracer(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize tracer: %v", err)
    }
    defer shutdown(ctx)

    // gRPC сервер с интерсепторами
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    // ... остальной код
}
Для Gateway (Gin):

go
import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// В main:
r := gin.Default()

// Мидлвара для трейсинга (ПЕРВАЯ!)
r.Use(otelgin.Middleware("gateway"))

// ... остальные middleware и маршруты
🧪 Шаг 2: k6 — нагрузочный тест

2.1. Устанавливаем k6:

bash
sudo apt-get update
sudo apt-get install -y k6
# Или через snap:
sudo snap install k6
2.2. Создаём скрипт loadtest.js:

javascript
// loadtest.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '30s', target: 10 },   // разогрев до 10 юзеров
        { duration: '30s', target: 50 },   // до 50
        { duration: '30s', target: 100 },  // до 100
        { duration: '30s', target: 500 },  // до 500
        { duration: '30s', target: 0 },    // спад
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% запросов < 500ms
        http_req_failed: ['rate<0.01'],   // ошибок < 1%
    },
};

const BASE_URL = 'http://localhost:8080';
const USER_EMAIL = `test_${__VU}@example.com`;
const USER_PASSWORD = 'secret123';

export default function () {
    // 1. Регистрация
    const registerRes = http.post(`${BASE_URL}/api/auth/register`, JSON.stringify({
        email: USER_EMAIL,
        password: USER_PASSWORD,
    }), { headers: { 'Content-Type': 'application/json' } });

    let userId = '';
    if (registerRes.status === 200) {
        userId = registerRes.json('userId');
    }

    // 2. Логин
    const loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
        email: USER_EMAIL,
        password: USER_PASSWORD,
    }), { headers: { 'Content-Type': 'application/json' } });

    const token = loginRes.json('access_token');
    check(loginRes, { 'login successful': (r) => r.status === 200 });

    if (!token) return;

    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    };

    // 3. Hexagon (Блинопёк)
    const hexRes = http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: userId,
        game_id: 'hexagon',
        level: 1,
        score: Math.floor(Math.random() * 500) + 50,
        user_email: USER_EMAIL,
        nickname: `Player_${__VU}`,
        seed: `seed_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers });

    check(hexRes, { 'hexagon submitted': (r) => r.status === 200 });

    // 4. Memory (Мемония)
    const memRes = http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: userId,
        game_id: 'memory',
        level: 1,
        score: Math.floor(Math.random() * 900) + 100,
        user_email: USER_EMAIL,
        nickname: `Player_${__VU}`,
        seed: `memory_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers });

    check(memRes, { 'memory submitted': (r) => r.status === 200 });

    // 5. Flappy Bird
    const flapRes = http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: userId,
        game_id: 'flappy',
        level: 1,
        score: Math.floor(Math.random() * 50) + 10,
        user_email: USER_EMAIL,
        nickname: `Player_${__VU}`,
        seed: `flappy_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers });

    check(flapRes, { 'flappy submitted': (r) => r.status === 200 });

    // 6. Towers (Башенки)
    const towRes = http.post(`${BASE_URL}/api/game/submit`, JSON.stringify({
        user_id: userId,
        game_id: 'towers',
        level: 1,
        score: Math.floor(Math.random() * 500) + 50,
        user_email: USER_EMAIL,
        nickname: `Player_${__VU}`,
        seed: `towers_${__VU}_${Date.now()}`,
        moves: [],
    }), { headers });

    check(towRes, { 'towers submitted': (r) => r.status === 200 });

    // Пауза между итерациями
    sleep(1);
}
2.3. Запускаем тест:

bash
# Проверить, что все сервисы запущены
make status

# Запустить нагрузку
k6 run loadtest.js

# Или с выводом в Grafana (через influxdb)
k6 run --out influxdb=http://localhost:8086/k6 loadtest.js
🖥️ Шаг 3: Смотрим результаты

3.1. Grafana дашборд для k6

Добавь Data Source: InfluxDB (если используешь)
Импортируй дашборд k6: ID 2587
3.2. Проверяем метрики вручную

bash
# RPS Gateway
curl -s "http://localhost:9090/api/v1/query?query=rate(gateway_requests_total[$__rate_interval])" | jq .

# Latency
curl -s "http://localhost:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(gateway_request_duration_seconds_bucket[$__rate_interval])) by (le, path))" | jq .

# Ошибки
curl -s "http://localhost:9090/api/v1/query?query=rate(gateway_requests_total{status=~\"5..\"}[$__rate_interval])" | jq .
3.3. Jaeger UI

Открой в браузере: http://localhost:16686

Выбери сервис: gateway
Нажми Find Traces
Увидишь полный путь запроса с таймингами
✅ Чек-лист на сегодня

Задача	Статус
OpenTelemetry во все 5 сервисов	⬜️
Jaeger UI показывает трейсы	⬜️
k6 скрипт написан	⬜️
Нагрузка 500 пользователей	⬜️
Метрики в Grafana обновляются	⬜️
Jaeger трейсы видны	⬜️

🧩 План действий (сегодня)

Настроить Jaeger (OpenTelemetry во все 5 сервисов) — 1 час
Запустить k6 с небольшой нагрузкой (50-100 юзеров) — увидеть трейсы в Jaeger
Увеличить нагрузку (500-1000 юзеров) — посмотреть, как система дышит
Собрать метрики в Grafana — RPS, latency, ошибки
Сравнить — что показывает Jaeger vs Grafana
🚀 Начинаем с Jaeger

Шаг 1: Обновляем зависимости во всех сервисах

bash
cd /home/denismatveev/event_horizon

for svc in auth billing game gateway leaderboard; do
  cd services/$svc
  go get go.opentelemetry.io/otel@v1.34.0
  go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.34.0
  go get go.opentelemetry.io/otel/sdk@v1.34.0
  go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@v0.59.0
  go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@v0.59.0
  go mod tidy
  cd ../..
done
Шаг 2: Проверяем, что Jaeger запущен

bash
docker ps | grep jaeger
# Должен быть event-horizon-jaeger
Шаг 3: Добавляем трейсинг в каждый сервис

Покажи мне один main.go (например, services/auth/cmd/main.go), и я напишу готовый код с трейсингом для всех 5 сервисов.

Капитан, давай сначала доделаем Jaeger, а потом врубим k6 на полную мощность. 💪


