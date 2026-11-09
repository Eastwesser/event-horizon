1. Вес контейнеров (для твоей карты проекта)
Из твоего вывода docker images (беру latest):

Сервис	Вес (latest)
balancer	15.2 MB
nats-hub	26.6 MB
analytics	27.2 MB
fulfillment	27.4 MB
game	29.9 MB
history	31.1 MB
notification	31.5 MB
shop	35.4 MB
authors	36.5 MB
payment	36.5 MB
profile	37.7 MB
auth	37.9 MB
inventory	38.8 MB
leaderboard	38.8 MB
billing	38.8 MB
gateway	44.2 MB

Инфраструктура:

nginx:alpine — 63 MB (виртуальный)

postgres:16-alpine — 294 MB

redis:7-alpine — 39.1 MB

nats:2.10-alpine — 25.5 MB

clickhouse:24.8-alpine — 857 MB (виртуальный)

grafana:latest — 1.16 GB

prometheus:latest — 250 MB

jaeger:all-in-one — 85.2 MB

Итого Go-сервисы (16 штук): ~550 MB. 

Это отлично для микросервисов (обычно один Go-бинарник в scratch — 10-30 MB). 

Укладываемся в < 50 MB на сервис. 
