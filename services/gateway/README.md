# 2. API GATEWAY SERVICE

```bash
.
├── cmd
│   └── main.go
├── Dockerfile
├── go.mod
├── internal
├── proto
└── README.md
```


Следующий шаг: gateway (REST → gRPC прокси)

Теперь нам нужен HTTP gateway, чтобы фронтенд (Angular) мог отправлять JSON, а мы внутри ходили в gRPC.

План для gateway:

Создадим структуру services/gateway/
Напишем HTTP хендлеры (Gin или стандартный net/http)
Подключимся к auth по gRPC (клиент)
Проксируем запросы: POST /api/auth/register → auth.AuthService/Register

--

2. Gateway: самописный + Envoy потом

Отличный план. Go gateway даст тебе:

Rate limiting (покажешь код на собеседовании)
Маршрутизация HTTP → gRPC
WebSocket прокси для leaderboard
CORS, заголовки, логирование
Envoy (или Traefik) потом — когда понадобится:

L7 балансировка с retries/circuit breakers
gRPC load balancing без дополнительного hop
Let's Encrypt автоматический
Для старта: Go gateway на 300 строк. Envoy добавим как "production hardening" позже.