Привет! Я работаю над проектом Event Horizon — это игровая платформа с микросервисной архитектурой на Go. Мне нужна помощь как senior-разработчику.

## ТЕКУЩИЙ СТЕК:
- 7 микросервисов: Auth, Game, Billing, Leaderboard, Profile, Shop, Gateway
- Balancer (Least Connections)
- PostgreSQL (7 БД) + Redis (кеш, сессии, лидерборд)
- NATS кластер (3 ноды) + JetStream
- Prometheus + Grafana + Jaeger
- Docker Compose (планируется переход на k3s)
- Фронтенд: React + TypeScript (4 игры: Flappy, Hexagon, Towers, Memory)
- WebSocket для real-time лидерборда
- K6 для нагрузочного тестирования

## ТЕКУЩАЯ ВЕРСИЯ: v1.0.5 (15.07.2026)
- ✅ Shop Service (магазин, инвентарь, покупки)
- ✅ NATS кластер 3 ноды + мониторинг
- ✅ Скины во всех играх (космические блины, радужные трубы, карточки со зверями)
- ✅ Redis кеширование баланса
- ✅ WebSocket через балансировщик
- ✅ Даты покупок в инвентаре

## БЛИЖАЙШИЕ ЗАДАЧИ (TODO):
1. Добить K6 тесты (p95 < 500ms, ошибки < 1%)
2. Добавить индексы в БД (highscores, user_currencies, inventory, purchases)
3. Раскомментировать Rate Limiter в Gateway
4. Написать OpenAPI для всех сервисов
5. README для каждого сервиса
6. Юнит-тесты (покрытие 70%+)

## ПРОБЛЕМЫ, КОТОРЫЕ МЫ РЕШИЛИ:
- NATS не резолвил короткие имена контейнеров (решили через полные имена event-horizon-nats-1)
- Billing не кешировал баланс в Redis (добавили SetBalance с TTL 5 минут)
- K6 отправлял test-user-id (заменили на реальный UUID из логина)
- Leaderboard падал из-за NATS подписки (пересобрали NATS Hub)

## КЛЮЧЕВЫЕ ФАЙЛЫ:
- services/shop/proto/shop.proto — gRPC контракт
- services/nats-hub/main.go — создание Stream EVENTS
- deployments/docker-compose.cluster.yml — вся инфраструктура
- deployments/k6/loadtest.js — нагрузочные тесты
- frontend/src/store/shopStore.ts — Zustand store для магазина

## ЗАПРОС:
Помоги мне с [опиши конкретную задачу]. Я хочу решение, которое:
1. Соответствует существующей архитектуре
2. Не ломает текущую функциональность
3. Может быть протестировано через K6 или curl

Дай код, объясни изменения и покажи, как проверить.