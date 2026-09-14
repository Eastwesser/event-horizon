# MANUSCRIPT_DOJO — Dojo ticket_to_go (M1–M6)

**Master-маршрут «собес завтра»:** [`MANUSCRIPT.md`](./MANUSCRIPT.md)  
Этот файл — деталь dojo. Теория: `prydwen_knowledge/`. Код: `old_…/0.dojo/exam/ticket_to_go/`.

Запуск везде: `export GOWORK=off && go run main.go`

---

## Сквозной ответ middle+ (30 сек)

> На входе — **rate limit**. Между инстансами — **load balancer**. На sync-вызовах наружу — **timeouts + retry с backoff/jitter**, снаружи **circuit breaker** (open → 503). Живость смотрим **health/ready**. Чтение ускоряем **cache** (у нас часто cache-aside, как в EH Inventory).

---

## M1 — Rate Limiter

| | |
|--|--|
| **Код** | `module_1_rate_limiters/` (`TODO.md`) |
| **Стратегия** | Worker pool ≠ rate limit. Для API — Token Bucket: capacity + refill + `Allow()`. |
| **Сказать** | Fixed Window прост, но burst на границе. Sliding точнее, дороже по памяти. Token Bucket даёт контролируемый burst после простоя. Leaky — ровный выход. |
| **Edge** | thread-safe Mutex; lazy refill vs ticker-горутина. |
| **Якорь** | strikes `11` |

---

## M2 — Load Balancer

| | |
|--|--|
| **Код** | `module_2_load_balancers/` |
| **Стратегия** | RR по умолчанию; Weighted при разной мощности; Least Conn при долгих запросах; IP Hash для sticky. |
| **Сказать** | `(i+1)%n` + atomic/Mutex. Worker pool: `close(jobs)` + WaitGroup. |
| **Edge** | пустой список серверов; unhealthy node (см. Yandex LB pack). |
| **Якорь** | strikes `12`, company `yandex/load_balancer` |

---

## M3 — Circuit Breaker

| | |
|--|--|
| **Код** | `module_3_circuit_breakers/` |
| **Стратегия** | Closed → (N fails) Open → (timeout) Half-Open → (успехи) Closed. |
| **Сказать** | Open = **fail fast 503**, не 500 и не hang. Mutex **не** держать на `fn()`. Retry ≠ CB. |
| **EH** | Gateway circuit → 503. |
| **Якорь** | strikes `13` · Prydwen `4.architecture_patterns/04_ARCH_NETWORK.md` |

---

## M4 — Health Checker

| | |
|--|--|
| **Код** | `module_4_health_checkers/` |
| **Стратегия** | ticker + `WithTimeout` на check; N сервисов = N loop + общий ctx. Backoff если dependency больна. |
| **Сказать** | **liveness ≠ readiness**. 99.9% ≈ 43 мин/мес. |
| **EH** | `/health` vs `/ready`. |
| **Якорь** | strikes `20` (timeout mindset) |

---

## M5 — Retry + Backoff

| | |
|--|--|
| **Код** | `module_5_retryers/` |
| **Стратегия** | `base * 2^i` + **jitter** (anti thundering herd) + maxRetries. Не ретраить permanent/неидемпотентное. |
| **Сказать** | 5xx/429 часто да; 400 нет. Комбо: retry внутри, CB снаружи. |
| **Якорь** | strikes `17` |

---

## M6 — LRU / Cache

| | |
|--|--|
| **Код** | `module_6_caches/` |
| **Стратегия** | map + doubly linked list, O(1). SafeLRU — Mutex (Get двигает список). |
| **Сказать** | Write-Through = consistency; Write-Back = скорость+риск; **Cache-Aside** = контроль в app (EH). |
| **Якорь** | strikes `16` · Trie в M6 leetcode |

---

## Ритуал «собес завтра» (90 мин)

Полный маршрут (strikes → dojo → company → oral → fail phrases): [`MANUSCRIPT.md`](./MANUSCRIPT.md).

Кратко для dojo-блока (20 мин): один модуль M1–M6 вслух + показать код.

---

## Статус

- [x] Dojo M1–M6 связи  
- [x] Ссылки на strikes / Yandex / Avito  
- [x] T-Bank / Ozon / WB — маршруты в [`MANUSCRIPT.md`](./MANUSCRIPT.md)  
- [x] Факап-фразы из `22_training_checklist` — в [`MANUSCRIPT.md`](./MANUSCRIPT.md)  

*Dojo layer done. Master = MANUSCRIPT.md.*
