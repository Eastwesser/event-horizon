# STATUS 29.04.2026 - 20:17

Основные сервисы (уже в плане)

Сервис	        Статус	        Назначение

Auth	        ✅ Готов	       JWT, регистрация, логин
Gateway	        🔜 Завтра	    HTTP → gRPC, WebSocket, rate limiting
Game	        📋 В плане	    Гексагональный пазл, Flappy Bird, валидация
Billing	        📋 В плане	    Лампочки, билетики, этажи, пассивный доход
Leaderboard	    📋 В плане	    Топ-10, Redis Sorted Set, NATS подписка

Дополнительные сервисы (обсуждали в начале)

1. Notification — 🔔 Был в списке

Push-уведомления (Firebase FCM)
Email (через SMTP)
"Твой рекорд побили!", "Забери бонус!"
Слушает NATS: score.beaten, daily.reward.ready
2. Analytics — 📊 Был в списке

Сбор метрик: DAU, MAU, retention
События: игры, покупки, логины
Хранилище: ClickHouse или PostgreSQL (схема events)
Слушает NATS: user.*, game.*, payment.*
3. Payment — 💰 Был в списке (интерфейс в Billing)

Интеграция с Boosty/Stripe
Вебхуки для подтверждения платежей
Публикует в NATS: payment.completed

4. Social/Friend — 👥 Был в списке (опционально)

Добавление друзей
Лента рекордов друзей
Вызовы "обидеть друга"



Твой уточнённый план (после твоего сообщения)

"Мы делаем gateway, потом game, потом billing, потом leaderboard. Потом оставшиеся микросервисы."

Получается такая очерёдность:

text
1. Auth ✅
2. Gateway 🔜
3. Game
4. Billing
5. Leaderboard (с NATS, Redis)
6. Notification (NATS слушатель)
7. Analytics (NATS слушатель + ClickHouse)
8. Payment (в Billing интерфейс, не отдельным)
9. Social (опционально, в конце)

NATS кластер (сразу 3 ноды)

Ты сказал: "можем сразу три сделать". Вот docker-compose секция для 3 нод NATS:

yaml
nats-1:
  image: nats:2.10-alpine
  container_name: event-horizon-nats-1
  command: ["-js", "--cluster_name", "nats-cluster", "--cluster", "nats://0.0.0.0:6222", "--routes", "nats://nats-1:6222,nats://nats-2:6222,nats://nats-3:6222"]
  ports:
    - "4222:4222"
    - "8222:8222"

nats-2:
  image: nats:2.10-alpine
  container_name: event-horizon-nats-2
  command: ["-js", "--cluster_name", "nats-cluster", "--cluster", "nats://0.0.0.0:6222", "--routes", "nats://nats-1:6222,nats://nats-2:6222,nats://nats-3:6222"]
  ports:
    - "4223:4222"
    - "8223:8222"

nats-3:
  image: nats:2.10-alpine
  container_name: event-horizon-nats-3
  command: ["-js", "--cluster_name", "nats-cluster", "--cluster", "nats://0.0.0.0:6222", "--routes", "nats://nats-1:6222,nats-2:6222,nats-3:6222"]
  ports:
    - "4224:4222"
    - "8224:8222"
Для старта можешь оставить 1 ноду (проще). Кластер добавишь, когда будешь тестировать отказоустойчивость.

Outbox паттерн (потрогать)

Для Billing (покупка этажей):

sql
-- Таблица outbox
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска
CREATE INDEX idx_outbox_unpublished ON outbox(published, created_at);
Воркер раз в секунду забирает published=false и отправляет в NATS.







Итог по дизайну (финальная версия)

text
┌─────────────────────────────────────────────────────────────┐
│                      Load Balancer (nginx)                  │
│                         (потом добавим)                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Gateway (Go + Gin)                             │
│   - Rate limiting (самописный)                              │
│   - HTTP → gRPC прокси                                      │
│   - WebSocket для leaderboard                               │
│   - JWT проверка                                            │
└─────────────────────────────────────────────────────────────┘
          │              │              │              │
          gRPC           gRPC           gRPC           WebSocket
          ▼              ▼              ▼              ▼
    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌──────────┐
    │ Auth ✅ │    │ Game    │    │ Billing │    │Leaderboard│
    │ :50051  │    │ :50052  │    │ :50053  │    │ :50054    │
    └─────────┘    └─────────┘    └─────────┘    └──────────┘
         │              │              │              │
         └──────────────┼──────────────┼──────────────┘
                        │              │
                        ▼              ▼
              ┌─────────────────────────────┐
              │      NATS JetStream          │
              │   (3 ноды кластер опционально)│
              └─────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
   ┌─────────┐    ┌──────────┐    ┌──────────┐
   │Notify   │    │Analytics │    │ (Social) │
   │(потом)  │    │(потом)   │    │(опционально)
   └─────────┘    └──────────┘    └──────────┘

Твои базы данных (каждый сервис со своим)

Сервис	        PostgreSQL	                Redis
Auth	        :5433 (users)	            :6379 (сессии, JWT blacklist)
Game	        :5434 (игровые сессии)	    :6380 (текущие игры)
Billing	        :5435 (валюты, этажи)	    :6381 (балансы, идемпотентность)
Leaderboard	    :5436 (архив топа)	        :6382 (sorted set)
Gateway	        -	                        :6383 (rate limiting)

Про биллинг и сервис оплаты (не путать. Один игровой, другой на бусти в реале)

Payment: Смешивать игровую валюту (лампочки/билетики) с реальными деньгами — антипаттерн. 

У них разные:
    SLA (платёж должен быть надёжным, игра может подождать)
    Регуляторные требования (ФНС, PSD2, PCI DSS)
    Частота (платежей мало по сравнению с игровыми действиями)

Payment — отдельный микросервис.

Исправленный порядок сервисов

text
1. Auth ✅
2. Gateway 🔜
3. Game
4. Billing (игровая валюта: лампочки, билетики, этажи)
5. Leaderboard (NATS + Redis)

Допы:
6. Payment (Boosty/Stripe, отдельно)
7. Notification (NATS слушатель)
8. Analytics (NATS слушатель + ClickHouse)
9. Social (опционально)
