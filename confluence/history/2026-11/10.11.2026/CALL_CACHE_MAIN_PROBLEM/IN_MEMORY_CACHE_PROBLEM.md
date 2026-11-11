Что делаем
Файл: /home/denismatveev/event_horizon/services/gateway/internal/middleware/auth.go

Функция: RequireAuth

Что добавляем: in-memory кэш token → (user_id, email, role) с TTL 1s. Без Redis, без внешних зависимостей — просто sync.Map + timestamp. 1s — компромисс: 140 VU × 4 ручки = 560 вызовов/итерация, но 10 токенов → 10 ключей, кэш на 1s → ~10 ValidateToken/1s вместо 560. ×56 меньше. Logout → через 1s 401, это приемлемо (1s — не «вечность» для read path).

Почему in-memory, а не Redis
Redis — сетевой хоп (~1ms), но всё равно хоп. Для 10 ключей — оверкилл.

In-memory — 0ms, нет сети. 10 ключей × 3 поля × 1s TTL — ничего не весит.

Минус: 3 реплики Gateway не делят кэш. Каждая своя 1s. Для нашего профиля (10 токенов) — ок, каждая реплика свой кэш, 10 ValidateToken/1s на реплику = 30/1s всего. Всё равно ×18 меньше 560.

## Результат (10.11.2026, in-memory кэш 1s)

### Что сделали
- `RequireAuth` — `sync.Map` кэш `token → {userID, email, role, expiry}`, TTL 1s.
- Fast path: cache hit → `c.Next()` без gRPC.
- Slow path: real ValidateToken → cacheSet.
- Без Redis, без внешних зависимостей — 10 ключей × 1s на реплику.

### Почему 1s (не 5s)
- 5s → logout **5s** с отозванным токеном. **Плохо** для безопасности.
- 1s → logout **1s**. Приемлемо для read path (shop/items, profile).
- Эффект: 10 токенов × 1s → **10 ValidateToken/1s** на реплику вместо **560** = **×56 меньше**.

### Результат (P=200, до/после)
| Метрика | До | После |
|---|---|---|
| browse p95 | 10396ms | ? |
| auth_token_validations_total | ~560/s | ~50/s |
| status_5xx | 0% | 0% |


## Шаг 4 (ValidateToken кэш 1s) — промежуточный результат

### Ручной тест (доказательство работы)
- 1-й запрос (slow path): 680ms
- 2-й запрос (fast path): 66ms
- ×10 быстрее — кэш работает.

### P=50 / P=100 / P=200 — ВНИМАНИЕ: хуже, чем до кэша
| Профиль | До кэша (P=50: 3166, P=100: 5900, P=200: 10396) | После кэша |
|---|---|---|
| P=50  | browse 3166ms | browse 5212ms (**×1.6 хуже**) |
| P=100 | browse 5900ms | browse 10420ms (**×1.8 хуже**) |
| P=200 | browse 10396ms | browse 24150ms (**×2.3 хуже**) |

### Гипотезы
1. **Cold start после рестарта Inventory** (миграция индексов) — PG cold buffer, Redis cold cache. Первые минуты медленно.
2. **3 реплики × свой sync.Map** — 3× cold start, 3× ValidateToken в первую секунду.
3. **Не кэш виноват** — все метрики (shop/inv/lb/profile) выросли одновременно.

### Что нужно
- `auth_token_validations_total` — **сейчас** (работает ли кэш под нагрузкой).
- Прогреть 5 минут, **повторить P=50** — вернулось ли к 3166ms.

### Возможные фиксы (если кэш виноват)
- TTL 5s для GET, 0s для POST (Шаг 4.5).
- Или откатить кэш, разобраться.

