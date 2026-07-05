У нас:

Gateway → сервисы (Auth, Game, Billing, Leaderboard) — через gRPC (синхронно).

Сервисы → сервисы (Game → Leaderboard, Game → Billing, Auth → Profile) — через NATS (асинхронно, события).

NATS — это шина для событий, а не для синхронных RPC-вызовов.

Важно: NATS — это не замена gRPC, а дополнительный канал для асинхронных уведомлений.

🧠 Как это работает в Event Horizon
Канал	Протокол	Для чего
gRPC	Синхронный	Запрос-ответ (Auth, Game, Billing, Leaderboard)
NATS	Асинхронный	События (score.updated, user.registered)
Пример:

Клиент → Gateway (HTTP)

Gateway → Game (gRPC) — сохранить рекорд

Game → NATS (publish score.updated)

Leaderboard → NATS (subscribe) — обновляет топ

Billing → NATS (subscribe) — начисляет валюту

Profile → NATS (subscribe) — обновляет профиль

✅ Так что план — правильный

Stream создаёт только Gateway (или отдельный хаб).

Остальные сервисы только подписываются и публикуют.

gRPC остаётся для синхронных вызовов.

📊 Даст ли это прирост?
Аспект	Что даст
Надёжность	        Stream'ы создаются один раз, независимо от сервисов.
Простота	        Каждый сервис знает только о своих subject'ах.
Масштабируемость	Добавляешь новый subject → обновляешь nats-hub → перезапускаешь.
Тестируемость	    Можно запускать nats-hub отдельно и проверять Stream'ы.


cd ~/event_horizon

# Game
cd services/game && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o game-service ./cmd/main.go
cd ../..
docker build -f Dockerfile.game.bin -t eastwesser/game:latest .

# Leaderboard
cd services/leaderboard && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o leaderboard-service ./cmd/main.go
cd ../..
docker build -f Dockerfile.leaderboard.bin -t eastwesser/leaderboard:latest .

# Profile
cd services/profile && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o profile-service ./cmd/main.go
cd ../..
docker build -f Dockerfile.profile.bin -t eastwesser/profile:latest .

# Billing (если убирал создание Stream)
cd services/billing && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o billing-service ./cmd/main.go
cd ../..
docker build -f Dockerfile.billing.bin -t eastwesser/billing:latest .