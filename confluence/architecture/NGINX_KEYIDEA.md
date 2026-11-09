# Nginx Key Idea: Reverse Proxy перед балансерами

## Контекст
В Event Horizon есть 3 реплики API Gateway (`gateway`, `gateway-2`, `gateway-3`) и 2 самописных балансера (`balancer`, `balancer-2`). Балансеры распределяют трафик между гейтвеями по алгоритму Least Connections.

## Проблема
Балансеры — это SPOF (Single Point of Failure). Если один упадёт, половина трафика умрёт. Если оба — сайт ляжет.

## Решение
Добавить Nginx как reverse proxy перед балансерами.

### Схема

CLIENT (React :5173)
│
▼
NGINX (:80)
│
├──▶ BALANCER-1 (:8079) ──▶ GATEWAY-1 (:8080) ──▶ gRPC-сервисы
│
└──▶ BALANCER-2 (:8079) ──▶ GATEWAY-2 (:8080) ──▶ gRPC-сервисы
└──▶ GATEWAY-3 (:8080) ──▶ gRPC-сервисы



### Конфигурация
- `upstream balancer_backend { server balancer:8079; server balancer-2:8079; }`
- Nginx слушает `:80` и проксирует всё на `balancer_backend`.
- WebSocket (`/ws/`) проксируется так же, с `Upgrade` и `Connection: upgrade`.

## Почему это работает
1. **Failover:** если один балансер упал, Nginx шлёт трафик на второй.
2. **Least Connections:** балансеры продолжают работать как раньше, выбирая гейтвей по наименьшему числу соединений.
3. **Keep-alive:** Nginx держит соединения с балансерами, не открывая новые.

## Почему это не тормозит
- Nginx — это C-приложение, обрабатывает запрос за < 1 мс.
- Keep-alive уменьшает overhead на TCP-соединения.
- Задержка Nginx незначительна по сравнению с bcrypt (100–300 мс) и gRPC (10–50 мс).

## Альтернативы
- **NGINX вместо балансеров:** не подходит, потому что Nginx не умеет Least Connections (только Round Robin).
- **Только балансеры без Nginx:** SPOF, не подходит для продакшена.
- **Kubernetes Ingress:** не подходит для текущего деплоя (docker-compose).

## Вывод
Nginx + 2 балансера — это **стандартная** схема для отказоустойчивости. Она даёт failover без потери производительности.