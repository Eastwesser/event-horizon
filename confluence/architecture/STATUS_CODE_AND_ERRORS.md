# Event Horizon — полный справочник статус-кодов

> Собрано из кода (`grep` по `services/`) + `STATUS_CODES.md`.
> v1.1.3 (октябрь 2026).
> Все статус-коды: HTTP (Gateway), gRPC (сервисы), numeric (c.JSON).

---

## 1. Маппинг gRPC → HTTP (Gateway `writeGRPCError`)

Источник: `services/gateway/internal/app/gateway.go:194–256`.

| gRPC code | HTTP code | Когда |
|-----------|-----------|-------|
| `OK` | 200 | Успех |
| `InvalidArgument` | **400** | Bind/validate, пустые поля |
| `Unauthenticated` | **401** | Нет/невалидный JWT, `ValidateToken` fail |
| `PermissionDenied` | **403** | Роль не подходит |
| `NotFound` | **404** | Author/item/user not found |
| `AlreadyExists` | **409** | Дубликат (unique key) |
| `Aborted` | **409** | Version conflict (optimistic lock) |
| `FailedPrecondition` | **409** | Например, stock=0, insufficient balance |
| `ResourceExhausted` | **429** | Rate limit / quota |
| `Unavailable` | **503** | Dependency down (Redis, NATS) |
| `Internal` | **500** | Unexpected |
| `Unknown` | **500** (fallback) | Строковый fallback: `InvalidArgument` → 400, `AlreadyExists` → 409, `NotFound` → 404, `Unauthenticated` → 401, `PermissionDenied` → 403 |
| **Особый случай** | **200** | `AlreadyExists` на `ConfirmPayment` → 200 (idempotent webhook) |
| **Особый случай** | **403 + `code: "subscription_required"`** | `PermissionDenied` с message `subscription_required` |

**Ключевой код:** `writeGRPCError` — единая точка маппинга. `writePaymentConfirmError` — особый случай для webhook.

---

## 2. HTTP-коды по сервисам (из кода)

### 2.1. Gateway

| Код | Где | Когда |
|-----|-----|-------|
| **200** | Везде | Успех |
| **400** | `gateway.go` | Bind/validate, пустые поля, invalid attributes, `item_id is required`, `currency must be lamps or tickets`, `file required`, `content is not an image`, `empty body`, `payment_id is required`, `plan is required`, `display_name is required`, `contact_email is required`, `game_id is required`, `quantity must be >= 1`, `refresh_token required` |
| **401** | `auth.go:47` | `missing authorization header` |
| **401** | `auth.go:61` | `invalid or expired token` |
| **401** | `auth.go:57` | `auth service unavailable` (странно — 401, а не 503, но так в коде) |
| **401** | `gateway.go:1432` | Payment webhook: HMAC fail |
| **401** | `gateway.go:646` | WS: `use a one-time ticket` |
| **401** | `gateway.go:651` | WS: `invalid or expired ticket` |
| **401** | `gateway.go:657` | WS: `missing ticket` |
| **403** | `auth.go:83` | `insufficient permissions` (роль) |
| **403** | `gateway.go:642` | WS: `origin not allowed` |
| **403** | `gateway.go:2093` | `include_deleted requires author or admin` |
| **403** | `gateway.go:2292` | `you can only edit your own items` |
| **403** | `gateway.go:2368` | `you can only delete your own items` |
| **403** | `gateway.go:2453` | `you can only soft-delete your own items` |
| **403** | `gateway.go:2498` | `you can only restore your own items` |
| **404** | `docs.go:18` | `/docs`, `/openapi.yaml` без `DOCS_PASSWORD` |
| **409** | `gateway.go:207` | gRPC `AlreadyExists` → 409 (fallback) |
| **429** | `ratelimit.go:47` | Rate limit (login 5/s IP, общий ~100/s) |
| **500** | `gateway.go` | Внутренние ошибки, `cannot read upload`, `cannot store upload`, `write failed`, `approved application missing user_id` |
| **502** | `gateway.go:1843` | `failed to update role after approve` (Auth down) |
| **503** | `gateway.go:132` | Circuit Breaker open |
| **503** | `gateway.go:136` | Bulkhead overloaded |
| **503** | `gateway.go:600` | `/ready` degraded (Redis unreachable) |
| **503** | `gateway.go:634` | WS ticket store unavailable |
| **503** | `gateway.go:663` | WS session store unavailable (Auth `Unavailable`) |
| **503** | `auth.go:54` | `session store unavailable` (Auth `Unavailable`) |
| **405** | `security.go:51` | Method not allowed |
| **400** | `security.go:45` | Path traversal (`..`, `\`) |

**Особые ответы:**
- `gateway.go:261` — `AlreadyExists` на `ConfirmPayment` → **200** `{"success": true, "message": ...}`.
- `gateway.go:1492` — YooKassa webhook со статусом != succeeded → **200** `{"ignored": true}`.
- `gateway.go:912`, `:920` — `_partial` profile (Cache hit / `Unavailable`) → **200** с `_partial: true`.

### 2.2. Balancer

| Код | Где | Когда |
|-----|-----|-------|
| **200** | `metrics.go:14, 19` | `/health`, `/metrics` |
| **502** | `least_conn.go:44` | Backend error (`connection refused`) |
| **503** | `least_conn.go:94` | `No available backends` |

### 2.3. Auth

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:38` | `public registration only allows role=user` |
| `InvalidArgument` | `grpc_handler.go:44, 226, 229, 233` | Validation (email, password 8–128) |
| `Unauthenticated` | `grpc_handler.go:65, 88, 106` | Login fail, refresh fail, validate fail |
| `Unauthenticated` | `auth.go:29, 37, 41` | Missing metadata, missing bearer, invalid token |
| `NotFound` | `grpc_handler.go:109, 158` | User not found |
| `InvalidArgument` | `grpc_handler.go:102, 121, 143, 153, 181` | Missing required fields |
| `Unavailable` | `grpc_handler.go:85, 126` | Redis down (refresh, validate) |
| `Internal` | `grpc_handler.go:46, 146, 173, 184, 201` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |
| `InvalidArgument` | `validate.go:26` | Validation interceptor |

### 2.4. Game

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:51, 110, 130` | `user_id and game_id are required` |
| `NotFound` | `grpc_handler.go:135` | Game not found |
| `InvalidArgument` | `game_service.go:356, 361` | Request/user_id/game_id required |
| `FailedPrecondition` | `game_service.go:364, 393, 396, 405` | Boost not available, billing fail |
| `Unavailable` | `game_service.go:367` | Billing unavailable |
| `Internal` | `game_service.go:372, 398, 409` | Lookup/spend/record fail |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.5. Billing

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:30, 49, 73, 92, 122` | Missing user_id/amount |
| `FailedPrecondition` | `grpc_handler.go:108` | Insufficient balance |
| `Internal` | `grpc_handler.go:37, 54, 59, 80, 110, 133` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.6. Leaderboard

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:28, 61, 78` | `game_id` / `user_id` required |
| `Internal` | `grpc_handler.go:40, 67, 94` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.7. Shop

| gRPC | Где | Когда |
|------|-----|-------|
| `Internal` | `grpc_handler.go:28, 96, 145, 184` | Unexpected |
| `PermissionDenied` | `grpc_handler.go:173` | `subscription_required` (мерч) |
| `NotFound` | `grpc_handler.go:175` | Item not found |
| `AlreadyExists` | `grpc_handler.go:177` | Duplicate purchase |
| `InvalidArgument` | `grpc_handler.go:180` | Bad input |
| `FailedPrecondition` | `grpc_handler.go:182` | Stock=0, insufficient |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.8. Inventory

| gRPC | Где | Когда |
|------|-----|-------|
| `Internal` | `grpc_handler.go:32, 37, 51, 66, 74, 93, 100, 114, 121, 135, 142, 161, 180, 196, 211, 222, 231, 265` | Unexpected |
| `NotFound` | `grpc_handler.go:46, 175, 194, 209, 220, 259` | Item not found |
| `FailedPrecondition` | `grpc_handler.go:178, 261` | Not enough stock |
| `Aborted` | `grpc_handler.go:257` | Version conflict |
| `InvalidArgument` | `grpc_handler.go:263` | Bad input |
| `PermissionDenied` | `role.go:38` | `insufficient permissions` (RBAC) |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.9. Profile

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:28` | `user_id required` |
| `NotFound` | `grpc_handler.go:37` | `profile not found` |
| `Internal` | `grpc_handler.go:33, 46, 84` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.10. Payment

| gRPC | Где | Когда |
|------|-----|-------|
| `NotFound` | `grpc_handler.go:77` | Payment not found |
| `InvalidArgument` | `grpc_handler.go:79` | Bad input |
| `PermissionDenied` | `grpc_handler.go:81` | `subscription_required` |
| `AlreadyExists` | `grpc_handler.go:83` | Idempotent webhook (→ 200 на Gateway) |
| `Internal` | `grpc_handler.go:85` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.11. Authors

| gRPC | Где | Когда |
|------|-----|-------|
| `NotFound` | `grpc_handler.go:155` | Author/application not found |
| `InvalidArgument` | `grpc_handler.go:159` | Bad input |
| `Internal` | `grpc_handler.go:161` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.12. History

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:47` | Bad input |
| `Internal` | `grpc_handler.go:49` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.13. Analytics

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:69` | Bad input |
| `Internal` | `grpc_handler.go:71` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.14. Notification

| gRPC | Где | Когда |
|------|-----|-------|
| `InvalidArgument` | `grpc_handler.go:63` | Bad input |
| `NotFound` | `grpc_handler.go:65` | Notification not found |
| `Internal` | `grpc_handler.go:67` | Unexpected |
| `Internal` | `recovery.go:24` | Panic recovery |

### 2.15. NATS Hub

| HTTP | Где | Когда |
|------|-----|-------|
| **200** | `app.go:101, 110` | `/health`, `/metrics` |
| **503** | `app.go:105, 114` | NATS disconnected |

### 2.16. Fulfillment

**Нет HTTP-ручек, нет gRPC.** Только NATS/Kafka consumer + HTTP `/health`, `/metrics` (коды — стандартные `200` / `503`, как у NATS Hub).

### 2.17. MCP

**Нет HTTP.** stdio (JSON-RPC). Ошибки — текстом в JSON-RPC, не HTTP-кодами.

---

## 3. Все gRPC-коды (уникальные, из вывода)

| Код | Где встречается |
|-----|-----------------|
| `InvalidArgument` | Все сервисы (валидация) |
| `Internal` | Все сервисы (unexpected) |
| `Unauthenticated` | Auth (login, refresh, validate) |
| `PermissionDenied` | Inventory (RBAC), Payment (subscription), Shop (subscription) |
| `NotFound` | Auth, Game, Inventory, Profile, Payment, Authors, Notification |
| `AlreadyExists` | Payment (webhook), Shop (purchase), Gateway (fallback) |
| `Aborted` | Inventory (version conflict), Gateway (fallback → 409) |
| `FailedPrecondition` | Billing (insufficient), Game (boost), Inventory (stock), Shop (stock) |
| `ResourceExhausted` | Gateway (retry interceptor), редко |
| `Unavailable` | Auth (Redis), Game (Billing), Gateway (Auth) |
| `Unknown` | Gateway (fallback) |
| `Unimplemented` | **Автогенерированный код** в `*_grpc.pb.go` — не реальный кейс, а заглушка для нереализованных методов. **Игнорировать на собесе.** |

**Важно:** `Unimplemented` встречается ~100 раз в выводе — это **автогенерация** из `protoc`, не реальные коды. Не путай с реальными.

---

## 4. Особые случаи (на собесе)

| Случай | Код | Почему |
|--------|-----|--------|
| **`AlreadyExists` на `ConfirmPayment`** | **200** | Idempotent webhook: повтор Boosty → 200, не открывает подписку второй раз |
| **YooKassa webhook (status != succeeded)** | **200** `{ignored: true}` | Не ошибка, просто игнор |
| **`PermissionDenied` с message `subscription_required`** | **403** + `code: "subscription_required"` | Мерч-гейт: клиент видит доменный код |
| **`_partial` profile** | **200** с `_partial: true` | Graceful Degradation (Billing/Game down) |
| **`/docs`, `/openapi.yaml` без `DOCS_PASSWORD`** | **404** | Docs Gate |
| **WS: `?token=` с JWT** | **401** | Защита от утечки JWT в query |
| **Auth `Unavailable` (Redis down)** | **503** | Не 401 — «не можем проверить», а не «не знаем кто» |
| **Balancer: backend dead** | **502** | `connection refused` |
| **Balancer: все backends dead** | **503** | `No available backends` |
| **Path traversal** | **400** | `..` / `\` |
| **Method not allowed** | **405** | Вне GET/HEAD/POST/PUT/PATCH/DELETE/OPTIONS |

---

## 5. Сводная таблица HTTP-кодов (уникальные)

| Код | Значение | Где |
|-----|----------|-----|
| **200** | OK | Везде (успех) |
| **400** | Bad Request | Gateway (validate, bind), все сервисы (InvalidArgument) |
| **401** | Unauthorized | Gateway (auth), Auth (Unauthenticated), Payment (HMAC fail) |
| **403** | Forbidden | Gateway (role, subscription), Shop/Payment (`subscription_required`), Inventory (RBAC) |
| **404** | Not Found | Gateway (docs gate, not found), все сервисы (NotFound) |
| **405** | Method Not Allowed | Gateway (security) |
| **409** | Conflict | Gateway (AlreadyExists, Aborted, FailedPrecondition) |
| **429** | Too Many Requests | Gateway (rate limit, ResourceExhausted) |
| **500** | Internal Server Error | Gateway, все сервисы (Internal, Unknown) |
| **502** | Bad Gateway | Gateway (role update fail), Balancer (backend error) |
| **503** | Service Unavailable | Gateway (CB, bulkhead, ready, WS ticket, session store), Balancer (no backends), NATS Hub (disconnected) |

---

## 6. Сводная таблица gRPC-кодов (реальные, без Unimplemented)

| Код | HTTP | Где |
|-----|------|-----|
| `InvalidArgument` | 400 | Все сервисы (валидация) |
| `Unauthenticated` | 401 | Auth |
| `PermissionDenied` | 403 | Inventory, Payment, Shop |
| `NotFound` | 404 | Auth, Game, Inventory, Profile, Payment, Authors, Notification |
| `AlreadyExists` | 409 (200 для webhook) | Payment, Shop |
| `Aborted` | 409 | Inventory (version) |
| `FailedPrecondition` | 409 | Billing, Game, Inventory, Shop |
| `ResourceExhausted` | 429 | Gateway (retry) |
| `Unavailable` | 503 | Auth (Redis), Game (Billing) |
| `Internal` | 500 | Все сервисы |
| `Unknown` | 500 (fallback) | Gateway |

---

## 7. Что сказать на собесе

> «Gateway — единственная точка маппинга gRPC → HTTP. `writeGRPCError`: `InvalidArgument` → 400, `Unauthenticated` → 401, `PermissionDenied` → 403, `NotFound` → 404, `AlreadyExists`/`Aborted`/`FailedPrecondition` → 409, `ResourceExhausted` → 429, `Unavailable` → 503, `Internal`/`Unknown` → 500. Особые случаи: `AlreadyExists` на `ConfirmPayment` → 200 (idempotent webhook), `PermissionDenied` с `subscription_required` → 403 + доменный код, `_partial` profile → 200, `/docs` без пароля → 404, WS `?token=` → 401, Auth `Unavailable` → 503 (не 401). Balancer: backend dead → 502, все dead → 503. `Unimplemented` в выводе — автогенерация `protoc`, не реальный кейс.»

---

## 8. Что НЕ говорить

1. **«`Unimplemented` — реальный код»** — нет, автогенерация.
2. **«401 на role»** — нет, 403.
3. **«500 на всё»** — нет, маппинг.
4. **«404 на пустой список»** — нет, 200 + `[]`.
5. **«429 на CB»** — нет, 503.
6. **«Auth `Unavailable` → 401»** — нет, 503.

---

## Итого

| Что | Количество |
|-----|-----------|
| HTTP-коды (уникальные) | 10 (200, 400, 401, 403, 404, 405, 409, 429, 500, 502, 503) — 11 |
| gRPC-коды (реальные) | 11 (OK, InvalidArgument, Unauthenticated, PermissionDenied, NotFound, AlreadyExists, Aborted, FailedPrecondition, ResourceExhausted, Unavailable, Internal, Unknown) — 12 |
| Особые случаи | 11 |
| Маппинг | `writeGRPCError` (gateway.go:194) |