# Security basics — шпаргалка senior (JWT, bcrypt, OWASP, RBAC)

## Угрозы, о которых спрашивают

Утечка токенов, слабый хеш паролей, секреты в репо, IDOR, injection, сломанный access control (OWASP A01), SSRF, зависимость от «безопасности через неясность».

## JWT

- Access token: короткий TTL, claims: `sub`, `role` (`user|author|admin`), `exp`, `iat`.
- Refresh: длиннее, хранение/ротация; в EH сессии/refresh завязаны на **Redis** (logout = revoke).
- Подпись HMAC/RS;* секрет только из env (`internal/config`), не в коде.
- На Gateway: `RequireAuth` парсит JWT; downstream может получить `x-user-id` / `x-user-role`.
- Не класть PII/пароли в claims. Не использовать JWT как «шифроконтейнер» больших данных.

Компрометация access → ждать expiry или blacklist/session version; поэтому короткий TTL важен.

## bcrypt cost 12

`bcrypt.GenerateFromPassword(password, 12)` — стандарт EH Auth. Cost 12 = разумный баланс brute-force vs latency на логине. Не хранить plaintext, не сравнивать строки хешей вручную — только `CompareHashAndPassword`.

На собесе: почему не MD5/SHA без salt; почему cost не 4 «чтобы тесты быстрее» в проде.

## Секреты в env

- JWT secret, DB DSN, Redis password — env / secret manager.
- Не коммитить `.env`; шаблоны — `deployments/env/*.template`.
- Разные секреты per env; ротация без хардкода в образе.
- MCP: отдельный read-only DSN, не write-prod.

## RBAC

Роли: `user` | `author` | `admin`. Enforcement на границе (Gateway) + повторная проверка на чувствительных gRPC (Inventory role interceptor). Принцип: **deny by default**.

IDOR: проверка не только роли, но и владения ресурсом (`user_id` в токене == owner), где нужно.

## OWASP Top-10 — сжатый маппинг на backend

| Риск | Практика в EH-стиле |
|------|---------------------|
| Broken Access Control | RBAC + тесты 401/403, без утечки чужих id |
| Cryptographic Failures | bcrypt 12, TLS на периметре, секреты в env |
| Injection | parameterized SQL; ClickHouse named params; validate proto |
| Insecure Design | Outbox/идемпотентность платежей, circuit breaker |
| Security Misconfiguration | нет дефолтных паролей в compose для прод-рассказа |
| Vulnerable Components | go.mod audit, минимальные образы |
| Auth Failures | JWT expiry, Redis session revoke |
| Software/Data Integrity | миграции, не exec из user input |
| Logging/Monitoring | не логировать пароли/токены; trace id |
| SSRF | не открывать произвольные URL из пользовательского ввода |

## Дополнительно по EH

- Interceptor `Validate()` на gRPC — первая линия против мусора в полях.
- Payment merch gate — бизнес-ограничение как security/бизнес invariant.
- MCP `postgres_query` — SELECT-only allowlist (агенты ≠ суперюзеры).

## Чеклист перед демо на собесе

«Секреты в env, bcrypt 12, роли на gateway, параметры в SQL, revoke сессий в Redis, circuit не роняет каскадом, агентский SQL только read-only.»

## Типичные вопросы на собесе

1. Как устроен JWT access vs refresh и revoke?
2. Почему bcrypt и почему cost 12?
3. 401 vs 403 в контексте RBAC?
4. Как предотвратить SQL/ClickHouse injection?
5. Что сделаете, если JWT secret утёк в git?
6. Как тестировать access control (unit policy vs e2e)?
7. Чем IDOR отличается от «просто забыли RequireRole»?
8. Какие секреты нельзя отдавать MCP/LLM-агенту?
