Перед рестартом — напомню последовательность, чтобы всё завелось с полпинка:

1. Включаем виртуалку

Запускаешь VM, логинишься.
2. Поднимаем инфраструктуру в Docker

bash
cd /home/denismatveev/event_horizon
docker-compose -f deployments/docker-compose.cluster.yml up -d
Проверяем:

bash
docker-compose -f deployments/docker-compose.cluster.yml ps
Должны быть: postgres (4 шт), redis (4 шт), nats, jaeger, prometheus, grafana, balancer.

3. Запускаем микросервисы локально

bash
make start-services
Или если make ругается — по отдельности:

bash
cd services/auth && ./auth-service &
cd services/billing && ./billing-service &
cd services/game && ./game-service &
cd services/leaderboard && ./leaderboard-service &
cd services/gateway && ./gateway-service &
4. Проверяем

bash
make status
curl http://localhost:8080/health
curl http://localhost:8079/health
Про твой вопрос про раскидывание сервисов по разным серверам

Ты хотел спросить про распределение микросервисов по разным физическим/виртуальным машинам.

Коротко:
Это называется горизонтальное масштабирование. Если у тебя несколько серверов — ты можешь:

Раскидать микросервисы по разным тачкам (чтобы они не дрались за ресурсы);
Или поднять на каждой тачке по копии каждого сервиса (для отказоустойчивости).
Для этого нужно:

Чтобы сервисы общались по сети (gRPC через IP/домены), а не через localhost.
Чтобы каждый сервис знал адреса других (через Service Discovery — типа Consul, или через переменные окружения с IP адресами).
Чтобы балансировщик (твой balancer) знал, где живут инстансы gateway.
Но для пэтки — это оверхед

Ты на виртуалке. Если хочешь поэкспериментировать с распределением — можно поднять второй экземпляр сервиса на другом порту или в другой VM, и прописать его в balancer.

Но по сути, для одной VM — всё в одном месте, и это ок.

Запускайся, проверяй — я рядом. Как поднимешь — отпишись.