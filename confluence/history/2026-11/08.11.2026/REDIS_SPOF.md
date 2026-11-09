# Redis SPOF: Sentinel vs Cluster

## Проблема
Каждый Redis (Auth, Billing, Shop, Inventory, Leaderboard, Profile, Authors, Payment) — это отдельный инстанс. Если он упадёт, сервис потеряет кэш (но данные в PG останутся).

## Два решения
### 1. Redis Sentinel (выбрано для MVP)
- 3 Sentinel-процесса + 1 реплика на каждый Redis.
- Автоматический failover за ~5-30 секунд.
- Go-клиенты используют `redis.NewFailoverClient`.
- **Плюс:** не нужно менять код, данные не шардируются.
- **Минус:** не решает проблему памяти.

### 2. Redis Cluster (на будущее)
- 3 мастера + 3 реплики, данные шардируются на 16384 слота.
- **Плюс:** горизонтальное масштабирование.
- **Минус:** сложно, нужен slot-aware код.

## План для Sentinel
1. Добавить по 1 реплике для Auth, Billing, Shop (критичные).
2. Поднять 3 контейнера `redis-sentinel`.
3. Настроить мониторинг мастеров.
4. В Go-сервисах заменить `redis.NewClient` на `redis.NewFailoverClient`.

## Когда переходить на Cluster
Когда кэш вырастет до гигабайт и перестанет влезать в память одного сервера.

---

# Redis SPOF: Sentinel для MVP

## Решение
Для MVP — **Redis Sentinel** для критичных Redis (Auth, Billing). Остальные — оставить как есть (fallback на PG).

## Критичные Redis
- **Auth** (:6379) — сессии, rate limit. Если упадёт — все 503.
- **Billing** (:6381) — кэш балансов. Если упадёт — медленнее, но работает.

## План (P1, не P0)
1. Добавить по 1 реплике для Auth и Billing.
2. Поднять 3 контейнера `redis-sentinel`.
3. Настроить мониторинг мастеров.
4. В Go-сервисах Auth и Billing заменить `redis.NewClient` на `redis.NewFailoverClient`.

## На будущее (P3)
Redis Cluster — когда кэш вырастет до гигабайт.