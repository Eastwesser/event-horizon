Что мы НЕ разобрали — честный список
Прошёл всё, что было в файлах и в наших ответах. Вот что осталось за кадром — по категориям.

1. Сервисы: что не разобрано
1.1. Game — не разобрано глубоко
Что	Почему важно
8 игр — каждая со своей логикой	Strategy по game_id — switch на 8 игр. Domain Validators только для Hexagon и Memory. Остальные 6 — базовая проверка. Что за игры? Flappy, Pancaker/Hexagon, Builder/Towers, Memonia, Hanoi, 2048, Gears, Tamagotchi. Каждая — свой score, свои правила.
Boost система — 3-фазный цикл	StartBoost → SpendCurrency → boost active. Бустнутый забег не в лидерборд. Но как именно работает 3-фазный цикл? Не разобрано.
Domain Validators — что проверяют	Hexagon: ходы, seed, score. Memory: ходы, seed, score. Остальные — базовая. Что именно проверяется? Не разобрано.
Tamagotchi daily gift	POST /game/companion/daily-gift → +1000 билетиков / день. Как работает? Idempotency по daily-gift:{user}:{date}. Не разобрано.
Leaderboard levels	Все доски читают level=1, кроме Flappy. Уровни 2–3 — в roadmap. Почему Flappy исключение? Не разобрано.
1.2. Leaderboard — не разобрано
Что	Почему важно
Multi-Level Key	Ключ включает game_id, level. Сейчас все level=1. Как работает multi-level? Не разобрано.
Restore-on-Startup	При старте читает PG → заливает Redis. Что если PG упал? Retry/backoff — в техдолге. Но как именно Restore работает? Не разобрано.
Ack Batcher	Подтверждения NATS батчем. Какой размер батча? Какой таймаут? Не разобрано.
1.3. Profile — не разобрано
Что	Почему важно
Rule Engine	EvaluateAchievements — чистая функция. Но какие достижения? Какие правила? Не разобрано.
Achievement Unlock Pipeline	EvaluateAndUnlock → UnlockCodes → Unlock. Как именно работает pipeline? Не разобрано.
Достижения	Какие есть? first_win, first_purchase, first_boost? Не разобрано.
1.4. Authors — не разобрано
Что	Почему важно
State Machine	pending → approved / rejected. Какие ещё состояния? aborted? Не разобрано.
Pessimistic Lock	FOR UPDATE на заявке. Но как именно? Не разобрано.
Compensating Action	RevertApplication — что именно откатывает? Hard delete authors — FK падает. Не разобрано.
1.5. History — не разобрано
Что	Почему важно
5 субъектов	score.updated, user.registered, event.user.registered, shop.purchased, payment.completed, author.upserted. Почему user.registered и event.user.registered — оба? Не разобрано.
RetentionWorker	Cron внутри процесса. Какой интервал? Какой размер пачки? Не разобрано.
Payload Sanitization	Невалидный JSON → {"raw": "..."}. Что именно считается невалидным? Не разобрано.
1.6. Analytics — не разобрано
Что	Почему важно
Cohort Retention Analysis	Классический SQL. Но какой именно? SELECT cohort, day, count(DISTINCT user_id) ... GROUP BY cohort, day. Не разобрано.
Parameterized Query	{name:Type} — named params. Как именно? Не разобрано.
Батчинг	Раз в N секунд или N событий. Какой N? Не разобрано.
1.7. Notification — не разобрано
Что	Почему важно
Fan-Out	1 → N уведомлений. Как именно resolve recipients? Не разобрано.
Telegram	Как подписывается пользователь? Как отправляется? Не разобрано.
In-app bell	FE компонент. Как работает? Не разобрано.
1.8. Payment — не разобрано
Что	Почему важно
Boosty	Как именно работает checkout? Redirect? Webhook? Не разобрано.
Revenue Split	85/15 — журнал, не перевод. Как именно записывается? Не разобрано.
Cache TTL based on Expiry	TTL из expires_at. Как именно? Не разобрано.
1.9. Fulfillment — не разобрано
Что	Почему важно
Delay-Based Simulation	time.After(10s). Что именно эмулируется? Не разобрано.
Dual-Transport	NATS primary, Kafka fallback. Как именно выбирается? Не разобрано.
Idempotency Guard	Map в памяти с TTL. Какой TTL? Не разобрано.
1.10. Gateway — не разобрано
Что	Почему важно
DTO Mapper	proto3 omitempty fix. Как именно? Не разобрано.
Partial Response Fallback	_partial: true. Как именно? Не разобрано.
Retry Interceptor	Full jitter. Какой base? Какой max? Не разобрано.
Circuit Breaker	Half-open через 10s, MaxRequests 3. Как именно? Не разобрано.
1.11. Balancer — не разобрано
Что	Почему важно
Least Connections	Atomic counter. Как именно? Код есть в SYSTEM_DESIGN.md. Не разобрано.
Reverse Proxy	httputil.ReverseProxy. Как именно? Не разобрано.
WS sticky	Как именно? Не разобрано.
1.12. NATS Hub — не разобрано
Что	Почему важно
Stream Config	MaxAge: 7d, MaxBytes: 1GB. Почему именно? Не разобрано.
Wildcard Subscription	event.>. Что именно логирует? Не разобрано.
Реальный health check	nc.IsConnected(). Как именно? Не разобрано.
1.13. MCP — не разобрано
Что	Почему важно
6 tools	search_prydwen, nats_list_streams, nats_list_consumers, postgres_query, redis_get, redis_keys. Как именно? Не разобрано.
Offline RAG (TF-IDF)	Как именно? Не разобрано.
SQL Sandbox	ValidateSelect. Что именно проверяет? Не разобрано.
2. Сквозные темы: что не разобрано
2.1. Deployment / DevOps
Что	Почему важно
Docker Compose	Как именно? deployments/docker-compose.cluster.yml. Не разобрано.
k3s	Как именно? make deploy-k3s. Не разобрано.
Helm	Как именно? deployments/helm/event-horizon. Не разобрано.
Ansible	Как именно? delivery/. Не разобрано.
GitHub Actions	Как именно? .github/workflows/. Не разобрано.
Thin vs Heavy	make deploy vs make deploy-heavy. Не разобрано.
Seeds	make seed-v110, seed-lb-demo, seed-card-artists. Не разобрано.
2.2. Observability
Что	Почему важно
Prometheus	Как именно скрейпит? prometheus.yml. Не разобрано.
Grafana	Как именно дашборд? event-horizon.json. Не разобрано.
Jaeger	Как именно трейсы? OTLP :4317. Не разобрано.
Alertmanager	Как именно алерты? Telegram? Не разобрано.
NATS Exporter	:7777. Не разобрано.
Postgres Exporter	:9187. Не разобрано.
Redis Exporter	:9121. Не разобрано.
2.3. Безопасность
Что	Почему важно
JWT	HMAC-SHA256, JWT_SECRET из env. Как именно? Не разобрано.
bcrypt	cost 12. Как именно? Не разобрано.
OAuth	Google/Yandex. Как именно? Не разобрано.
HMAC	Boosty webhook. Как именно? Не разобрано.
CSRF	Не применим (Bearer). Но как именно? Не разобрано.
XSS	React экранирует. Но как именно? Не разобрано.
SQL injection	Параметризованные $1, $2. Но как именно? Не разобрано.
Path Traversal	.. отклоняется. Но как именно? Не разобрано.
WAF	Нет. Техдолг. Не разобрано.
2.4. Frontend
Что	Почему важно
React	8 игр, магазин, профиль. Не разобрано.
Vite	:5173. Не разобрано.
API client	baseURL: '/api/v1'. Не разобрано.
WebSocket	ws://localhost:5173/ws/leaderboard. Не разобрано.
NotificationBell	FE компонент. Не разобрано.
ShopItemCard	isDigitalItem. Не разобрано.
GameShell	Единый chrome. Не разобрано.
2.5. Тестирование
Что	Почему важно
Unit	task test. Не разобрано.
Integration	testcontainers. Не разобрано.
k6	browse.js, purchase.js. Не разобрано.
Race	go test -race. Не разобрано.
Failover	Game. Не разобрано.
Thorough	Gateway (CB, retry). Не разобрано.
Coverage	task test-coverage. Не разобрано.
2.6. Load testing
Что	Почему важно
CORE k6	browse.js (20 VU), purchase.js (1–5 VU). Не разобрано.
Legacy 500 VU	Login blast. Не разобрано.
Load posture	LOAD_POSTURE.md. Не разобрано.
Suitable server	SUITABLE_SERVER.md. Не разобрано.
K6 verdict	K6_VERDICT.md. Не разобрано.
2.7. Roadmap / Tech Debt
Что	Почему важно
FUTURE_EVENT_HORIZON.md	i18n, tech debt, games polish. Не разобрано.
DREAM_DEVOPS_IDEAS.md	Loki, Tempo, SLO burn. Не разобрано.
LEFT_TO_FINISH.md	Что осталось. Не разобрано.
STILL_TECH_DEBT.md	Техдолг. Не разобрано.
3. Конкретные вопросы, которые мы не разобрали
3.1. «Что если …»
Вопрос	Разобрано?
Что если два клиента одновременно покупают последний товар?	❌ Не разобрано (есть version conflict, но не race)
Что если Boosty присылает webhook с неверной суммой?	❌ Не разобрано (техдолг: сверка суммы)
Что если игрок отправил score больше max?	❌ Не разобрано (Domain Validator)
Что если record_beaten приходит для себя?	❌ Не разобрано (Event Chaining)
Что если два админа approve одну заявку?	✅ Разобрано (Pessimistic Lock)
Что если Redis упал во время resumeSagas?	❌ Не разобрано
Что если NATS упал во время purchase.fulfilled?	❌ Не разобрано
Что если ClickHouse упал во время батча?	❌ Не разобрано (Nak)
Что если Telegram API вернул 429?	❌ Не разобрано (Retry Fixed Delay)
Что если MCP postgres_query с ; DROP?	✅ Разобрано (SQL Sandbox обходится)
3.2. «Как работает …»
Вопрос	Разобрано?
Как работает resumeSagas через 30s?	✅ Разобрано (sequence)
Как работает Ack Batcher?	❌ Не разобрано (размер батча, таймаут)
Как работает Restore-on-Startup?	❌ Не разобрано (что если PG упал)
Как работает Rule Engine?	❌ Не разобрано (какие правила)
Как работает Pipeline в Profile?	❌ Не разобрано
Как работает State Machine в Authors?	❌ Не разобрано (какие ещё состояния)
Как работает Payload Sanitization?	❌ Не разобрано (что считается невалидным)
Как работает Cohort Retention?	❌ Не разобрано (какой SQL)
Как работает Parameterized Query?	❌ Не разобрано (named params)
Как работает Fan-Out?	❌ Не разобрано (resolve recipients)
Как работает Revenue Split?	❌ Не разобрано (как записывается)
Как работает Cache TTL based on Expiry?	❌ Не разобрано
Как работает Delay-Based Simulation?	❌ Не разобрано (что эмулируется)
Как работает Dual-Transport?	❌ Не разобрано (как выбирается)
Как работает Idempotency Guard?	❌ Не разобрано (TTL)
Как работает DTO Mapper?	❌ Не разобрано (proto3 omitempty fix)
Как работает Partial Response Fallback?	❌ Не разобрано
Как работает Retry Interceptor?	❌ Не разобрано (base, max)
Как работает Circuit Breaker?	❌ Не разобрано (half-open, MaxRequests)
Как работает Least Connections?	❌ Не разобрано (код есть, но не разобрали)
Как работает WS sticky?	❌ Не разобрано
Как работает Stream Config?	❌ Не разобрано (MaxAge, MaxBytes)
Как работает Wildcard Subscription?	❌ Не разобрано (что логирует)
Как работает TF-IDF?	❌ Не разобрано
Как работает SQL Sandbox?	✅ Разобрано (обходится)
3.3. «Почему …»
Вопрос	Разобрано?
Почему MaxAge: 7d, а не 30?	✅ Разобрано
Почему MaxBytes: 1GB?	✅ Разобрано
Почему bcrypt cost 12?	✅ Разобрано
Почему Redis Sorted Set, а не PG?	✅ Разобрано
Почему Saga, а не 2PC?	✅ Разобрано
Почему Outbox, а не прямой publish?	✅ Разобрано
Почему ClickHouse, а не PG?	✅ Разобрано
Почему NATS, а не Kafka?	✅ Разобрано
Почему Least Connections, а не round-robin?	❌ Не разобрано
Почему Atomic Counter, а не Mutex?	❌ Не разобрано
Почему Write-Through, а не Cache-Aside в Leaderboard?	❌ Не разобрано
Почему Max-Only Write (GT: true)?	✅ Разобрано
Почему Spend Cycle?	✅ Разобрано
Почему Pessimistic + Optimistic Lock в одной TX?	✅ Разобрано
Почему Dual-Bus?	✅ Разобрано
Почему Dual-Transport?	✅ Разобрано
Почему Manual DI в Fulfillment, Balancer, NATS Hub?	❌ Не разобрано
Почему stdio для MCP?	✅ Разобрано
Почему No gRPC в Fulfillment?	✅ Разобрано
4. Что мы разобрали (сводка)
Категория	Разобрано	Не разобрано
Сервисы	17 (по одному кейсу на каждый)	Глубоко — 10 из 17 (Auth, Game, Billing, Leaderboard, Shop, Inventory, Profile, Payment, Authors, History, Analytics, Notification, Fulfillment, Gateway, Balancer, NATS Hub, MCP — все 17, но не все детали)
Sequence (happy path)	17	—
Sequence (edge cases)	17	—
Глубокие разборы	15	—
Thundering herd / Retry / Cache	3	Детали (размеры батчей, TTL)
Follow-up 10k / 100k	2	1M — поверхностно
Техдолг	40	—
Шпаргалка	1	—
FULL.md	1	—
5. Что можно разобрать дальше (если нужно)
5.1. Высокий приоритет
8 игр Game — каждая со своей логикой, Domain Validators, score.

Rule Engine Profile — какие достижения, какие правила.

State Machine Authors — какие ещё состояния.

Cohort Retention Analytics — какой SQL.

Fan-Out Notification — resolve recipients.

Revenue Split Payment — как записывается.

DTO Mapper Gateway — proto3 omitempty fix.

Least Connections Balancer — код, atomic counter.

WS sticky Balancer — как именно.

TF-IDF MCP — как именно.

5.2. Средний приоритет
RetentionWorker History — интервал, размер пачки.

Payload Sanitization History — что считается невалидным.

Parameterized Query Analytics — named params.

Батчинг Analytics — N секунд или N событий.

Cache TTL based on Expiry Payment — как именно.

Delay-Based Simulation Fulfillment — что эмулируется.

Dual-Transport Fulfillment — как выбирается.

Idempotency Guard Fulfillment — TTL.

Retry Interceptor Gateway — base, max.

Circuit Breaker Gateway — half-open, MaxRequests.

Ack Batcher Leaderboard — размер, таймаут.

Restore-on-Startup Leaderboard — что если PG упал.

Stream Config NATS Hub — MaxAge, MaxBytes.

Wildcard Subscription NATS Hub — что логирует.

SQL Sandbox MCP — что именно проверяет.

5.3. Низкий приоритет (DevOps, Frontend, Testing)
Docker Compose — как именно.

k3s — как именно.

Helm — как именно.

Ansible — как именно.

GitHub Actions — как именно.

Prometheus — как скрейпит.

Grafana — дашборд.

Jaeger — трейсы.

Alertmanager — алерты.

React — 8 игр, магазин.

Vite — proxy.

Unit / Integration / k6 — как именно.

Load posture — детали.

Suitable server — детали.

K6 verdict — детали.