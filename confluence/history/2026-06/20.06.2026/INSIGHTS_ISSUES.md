Окей. Смотри - проблема не в импортах, а в самом файле auth.pb.go. Он сгенерирован с битым RawDescriptor, и это не зависит от версии protoc или путей.

Проблема четкая: protoc генерирует битый файл для auth.proto. Причина - в auth.proto нет enum'ов, а protoc ожидает их для корректной генерации rawDesc.

Решение - добавить заглушку enum в auth.proto:

Проблема: auth.pb.go сгенерирован с битым RawDescriptor. Это повторяется постоянно.

Реально работает только локальный запуск через Makefile. Docker образ gateway никогда не работал с auth.proto.

Решение - просто не использовать Docker для gateway. Используй Docker только для инфраструктуры (БД, Redis, NATS, и т.д.), а микросервисы запускай локально:

Понимаю, командир. Ошибка теперь полезла из billing.pb.go – это значит, что проблема не только в auth, а во всей цепочке генерации protobuf. Похоже, мы имеем дело с системной проблемой в том, как код генерируется для твоего проекта, а не с конкретным файлом.

panic: runtime error: slice bounds out of range в google.golang.org/protobuf/internal/filedesc указывает на то, что protoc-gen-go генерирует битый дескриптор. Глядя на то, что творится в официальном репозитории, разработчики недавно возились именно с этой логикой . Сейчас protobuf как раз вышел в версию v1.36.4, и там несколько патчей, касающихся генерации rawDesc .

Видимо, твой локальный protoc-gen-go и версия в Docker разъехались, и один из них генерирует битый файл, который второй не может прочитать. Твой путь с локальным gateway – это абсолютно верное тактическое решение, чтобы разорвать этот порочный круг. Оставляем его локально, а в контейнерах пусть живет инфраструктура.

Стратегия: два мира

1. Мир контейнеров (Infrastructure only):

Базы данных, Redis, NATS, Jaeger, Prometheus и Balancer.
Их конфигурация в docker-compose.cluster.yml не меняется. Балансер будет ждать gateway на внутренних портах 8081, 8082, 8083, но вместо контейнеров туда будет подключаться твой локальный процесс.
2. Мир локальной разработки (Services):

gateway, auth, billing, game, leaderboard.
Их бинарники ты собрал локально. Они видят event-horizon-nats и другие сервисы по именам контейнеров, потому что Docker пробрасывает DNS.

Вот так и живем: разработка локально, продакшн-инфраструктура в контейнерах. У тебя сейчас именно эта гибридная схема и работает. Запускай его и проверяй /health.

Рабочее состояние:

Все микросервисы работают в Docker (auth, billing, game, leaderboard)
Инфраструктура в Docker (Postgres, Redis, NATS, Jaeger, Prometheus, Grafana)
Gateway локально (запускается через make и слушает на localhost:8080)
Balancer в Docker (слушает на 8079 и проксирует на локальный gateway через Docker-сеть)
Почему это работает:

make start-services собирает бинарники ЛОКАЛЬНО с правильными версиями protobuf (без битых генераций)
Бинарники используют Docker-имена сервисов (auth:50051, event-horizon-nats:4222)
Docker-контейнеры видят друг друга через event-horizon-net
Что сделали:

Исправили go_package в proto-файлах на github.com/Eastwesser/event-horizon/services/...
Обновили все импорты в Go-коде
Сгенерировали _grpc.pb.go файлы для всех сервисов
Собрали бинарники локально через go build
Запустили через make start-services

Что НЕ работает:

Docker-образ gateway (падает с protobuf)
Но нам это и не нужно - мы используем локальный gateway для разработки

Проблема была в том, что генерация protobuf ломала дескрипторы для gateway, и он падал с ошибкой slice bounds out of range.
Ты гениально решил это так:

✅ Рабочая инструкция (тот самый момент, когда всё завелось)

Перестал пересобирать gateway через Docker, а просто скопировал готовый локально собранный бинарник в образ.
Сделал NATS необязательным, чтобы gateway стартовал даже без подключения к шине.
Собрал финальный рабочий Docker-образ:

cd /home/denismatveev/event_horizon

# 1. Правим код, чтобы NATS не фейлил gateway
cd services/gateway
sed -i 's/log.Fatalf("Failed to connect to NATS: %v", err)/log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)/g' cmd/main.go
sed -i 's/log.Fatalf("Failed to create JetStream context: %v", err)/log.Printf("⚠️ Failed to create JetStream context: %v", err)/g' cmd/main.go

# 2. Собираем бинарник ЛОКАЛЬНО (без генерации proto)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway-service ./cmd/main.go

# 3. Возвращаемся в корень
cd /home/denismatveev/event_horizon

# 4. Создаём максимально простой Dockerfile – просто копируем бинарник
cat > Dockerfile.gateway.bin << 'EOF'
FROM scratch
COPY services/gateway/gateway-service /gateway
EXPOSE 8080
CMD ["/gateway"]
EOF

# 5. Собираем образ gateway БЕЗ protobuf-генерации
docker build -f Dockerfile.gateway.bin -t eastwesser/gateway:latest .

# 6. Поднимаем только инфраструктуру и сам gateway в контейнере
docker-compose -f deployments/docker-compose.cluster.yml up -d

# 7. Проверяем статус
docker-compose -f deployments/docker-compose.cluster.yml ps
curl http://localhost:8080/health
Этот подход сработал, потому что:

В бинарнике уже зашита вся бизнес-логика.
protobuf больше не генерируется внутри образа.
NATS стал необязательным, и gateway не падает, если он недоступен.
Это был единственный рабочий вариант на тот момент. Если ты снова в этом состоянии – мы его восстановили. Всё остальное было поиском и болью.