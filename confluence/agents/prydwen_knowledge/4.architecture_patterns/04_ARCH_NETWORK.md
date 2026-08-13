# Сеть и протоколы: gRPC, REST, gateway, TLS, timeouts, circuit breaker

Senior backend обязан уверенно говорить про транспорт между сервисами, таймауты и деградацию. В Event Horizon: **Gin gateway** снаружи, **gRPC** между сервисами, OpenAPI/`/docs` для HTTP-контракта.

## gRPC vs REST (когда что)

- **REST/JSON**: человекочитаемо, кэшируется на HTTP-семантике, удобно фронту/публичному API, отлично для CRUD и BFF.
- **gRPC/HTTP2 + Protobuf**: строгий контракт, стриминг, меньше overhead на внутренних RPC, кодоген stub'ов.
- EH правило: не изобретать второй HTTP-gateway поверх всего; публичный вход — Gin + `docs/openapi.yaml`; внутренности — gRPC.
- Не смешивай без нужды: браузер ↔ gateway REST/WS; gateway ↔ auth/shop/billing — gRPC.

## API Gateway в EH

- Единая точка: auth middleware (JWT), роутинг, агрегация, rate limit/circuit breaker.
- Отдаёт `GET /openapi.yaml` и `GET /docs` (Swagger UI) — контракт для фронта.
- Gateway **не** место для тяжёлой доменной логики (Shop/Billing остаются владельцами правил).
- Ошибки наружу: осмысленные HTTP status; внутрь — gRPC codes (`InvalidArgument`, `NotFound`, `Unavailable`).

## TLS и безопасность канала

- TLS: шифрование + проверка сертификата; HTTPS снаружи обязательно на проде.
- TLS handshake: ClientHello → ServerHello/Certificate → ключи → Application Data. На собесе достаточно уровней: аутентификация сервера, опционально mTLS.
- **mTLS** между сервисами в mesh/k8s — сильный prod-паттерн; в локальном compose часто plaintext + private network.
- JWT: bearer на gateway; роли `user` | `author` | `admin`; сессии auth в Redis. Секреты — только env/K8s Secret, не в git.

## Timeouts, retries, deadline

- У каждого исходящего вызова — **явный timeout/deadline** (context в Go). Без deadline запросы копятся и роняют процесс.
- Иерархия: client timeout ≥ gateway ≥ downstream; иначе «висящие» горутины.
- Retries только на **идемпотентных** или с Idempotency-Key; на Spend без ключа — опасно.
- Backoff + jitter; ограничить max attempts. Не ретраить `InvalidArgument`.
- gRPC: `context.WithTimeout`; смотри `codes.DeadlineExceeded` vs `Unavailable`.

## Circuit breaker

- Защита caller'а от каскадного отказа: после N подряд ошибок — **open** (сразу fail), затем **half-open** пробные запросы, при успехе — **closed**.
- В EH gateway: `internal/circuit`, Timeout≈10s, MaxRequests=3 на half-open, быстрый 503 в open.
- Сочетать с bulkhead (лимит параллелизма) и правильными таймаутами.
- Метрики: state transitions, rejected calls — обязательны в Prometheus.

## OSI / TCP vs UDP (кратко для собеса)

- Для backend-собеса часто достаточно: L3 IP, L4 TCP/UDP, L7 HTTP/gRPC.
- **TCP**: надёжная доставка, порядок, stream — основа HTTP/gRPC.
- **UDP**: без гарантий; DNS, QUIC/HTTP3 снизу, realtime.
- HTTP/1.1 vs HTTP/2: мультиплекс в H2 критичен для gRPC.
- WebSocket: длинное двунаправленное соединение (игровой/realtime контур через gateway).

## Health на сетевом периметре

- `/health` — liveness (процесс жив).
- `/ready` — dependency ping (PG/Redis/NATS); не готов — убрать из балансировки.
- Не путать: kill по ready = ложные рестарты; только liveness для restart policy.

## Антипаттерны

- `http.DefaultClient` без timeout (запрещено правилами EH).
- Бесконечные retries на non-idempotent purchase.
- Один глобальный timeout 60s на всё.
- Игнор `Unavailable` и отсутствие circuit breaker на критичных зависимостях.
- Светить внутренние gRPC порты в интернет без gateway/auth.

## Типичные вопросы на собесе

- Когда выберете gRPC, а когда REST?
- Как устроен API Gateway и что нельзя в него тащить?
- Объясните TLS handshake и зачем mTLS.
- Как расставить timeouts в цепочке client→gateway→billing?
- Как работает circuit breaker и какие параметры крутить?
- Чем `/health` отличается от `/ready`?
- Почему retry на списании баланса опасен и как сделать безопасно?
