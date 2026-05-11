проверим полный цикл Billing

bash
USER_ID="d37c56ef-0d8c-4aa8-99ae-655e531333bc"

# 1. Проверить баланс лампочек
grpcurl -plaintext -d '{"user_id":"'$USER_ID'","currency":1}' \
  localhost:50053 billing.BillingService/GetBalance

# 2. Добавить билетики
grpcurl -plaintext -d '{"user_id":"'$USER_ID'","currency":2,"amount":50,"reason":"record","reference_id":"record-1"}' \
  localhost:50053 billing.BillingService/AddCurrency

# 3. Получить все балансы
grpcurl -plaintext -d '{"user_id":"'$USER_ID'"}' \
  localhost:50053 billing.BillingService/GetAllBalances

# 4. Получить историю транзакций
grpcurl -plaintext -d '{"user_id":"'$USER_ID'","currency":1,"limit":10}' \
  localhost:50053 billing.BillingService/GetTransactionHistory
Что у нас теперь есть (полный стек)

Сервис	Порт	Статус	Назначение
Auth	50051	✅	Регистрация, JWT
Game	50052	✅	Валидация, подсчёт очков
Billing	50053	✅	Лампочки, билетики
Leaderboard	50054	✅	Топ-10, Redis
Gateway	8080	✅	HTTP, WebSocket, NATS
NATS	4222	✅	Событийная шина
Полный рабочий пайплайн

text
curl → Gateway (8080) → Game (50052) → NATS (4222) → Leaderboard (50054) → Redis (6382)
                                              ↓
                                        Billing (50053) → PostgreSQL (5462)
                                              ↓
                                        Начисление валюты