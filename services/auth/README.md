# 1. AUTH SERVICE

```text
services/auth/
├── cmd/
│   └── main.go                 # точка входа
├── internal/
│   ├── config/
│   │   └── config.go           # конфигурация (порты, БД, JWT секрет)
│   ├── repository/
│   │   └── user_repo.go        # работа с PostgreSQL
│   ├── service/
│   │   └── auth_service.go     # бизнес-логика
│   └── handler/
│       └── grpc_handler.go     # gRPC хендлер
├── proto/
│   └── auth.proto              # gRPC контракт
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

# CREATE GRPC FILES:
```bash
cd services/auth
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/auth.proto
```

# COMPILE:
```bash
go build -o auth-service ./cmd/main.go
```

# RECOMPILE ^^
```bash
cd ~/event_horizon/services/auth
go build -o auth-service ./cmd/main.go
./auth-service
```

# AUTH TEST SITE:
```bash
# 1. Регистрация нового пользователя
grpcurl -plaintext -d '{"email":"bob@example.com","password":"secret123"}' \
  localhost:50051 auth.AuthService/Register

# 2. Попробовать зарегистрировать того же (должна быть ошибка)
grpcurl -plaintext -d '{"email":"bob@example.com","password":"secret123"}' \
  localhost:50051 auth.AuthService/Register

# 3. Логин
grpcurl -plaintext -d '{"email":"bob@example.com","password":"secret123"}' \
  localhost:50051 auth.AuthService/Login
```

# FINAL:

ГОТОВО! Auth сервис работает как часы!

```text
✅ Регистрация — работает
✅ Защита от дубликатов — работает (user already exists)
✅ Логин — работает, JWT выдаётся
✅ Валидация паролей — работает

Компонент	              Статус
gRPC сервер	            ✅ слушает :50051
PostgreSQL подключение	✅ через pgxpool
Хэширование паролей	    ✅ bcrypt
JWT генерация	          ✅ HS256
User repository	        ✅ CRUD операции
Reflection API	        ✅ включил
```
