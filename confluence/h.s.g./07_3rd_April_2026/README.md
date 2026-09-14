# Урок 07 — 3 апреля 2026

System Design чата: расчёт нагрузки, LB/Gateway, JWT, **WebSocket connection management**, Redis Pub/Sub fan-out, thundering herd / singleflight / rate limit при reconnect.

## Задачи / что говорить

### 1. WS connection map
Инстанс хранит `userID → conn`. Sticky sessions на ALB, чтобы не прыгать между подами. При падении — клиент reconnect.

### 2. Pub/Sub fan-out
Сообщение в Redis PubSub → все WS-ноды; нода шлёт только если user локальный. Persistent storage отдельно от realtime.

### 3. Reconnect storm
10k сокетов одновременно: rate limiter + singleflight на «load history», иначе дудос своего же чата.

Шпаргалка цифр: RPS = DAU×actions/86400; 80/20 → кэш; >1000 RPS → LB.

## Как запускать

```bash
export GOWORK=off
cd code/01_ws_connection_map && go run main.go
```
