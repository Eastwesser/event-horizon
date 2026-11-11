🟢 УРОВЕНЬ 0 — «5–15 минут, одна правка»
#	Сервис	Задача	Собеседный вес	Статус
P2-7	Auth	boosty.to URL захардкожен → Config	Config View / 12-factor	открыто
P2-8	Payment	subscriptionDays default 30 → явно в config	Конфиг vs магия	открыто
P2-6	Billing	GetBalance возвращает -1 при cache miss → (int, bool)	Cache miss vs ошибка	открыто
P2-15	DevOps	make seed-v110 описать в DAILY_OPS.md	Документация как код	открыто
P2-2	Balancer	Пустые пакеты backend/, proxy/ — удалить	Мёртвый код	открыто
P2-1	Gateway	log.Printf → slog (в Balancer уже сделано)	Structured logging	открыто
P2-4	Fulfillment	newUUID vs uuid.NewString() — унифицировать	Консистентность crypto	открыто
P0-9	Shop	Проверить/добавить метрику merch_gate_disabled	Метрики деградации	⚠️ частично закрыт
P2-3	Balancer	slog в hot path	—	✅ закрыт v1.1.3.6
🟡 УРОВЕНЬ 1 — «30 мин – 1 час, правка + тест»
#	Сервис	Задача	Собеседный вес	Статус
BUG-08-2	Inventory	price=0 неотличим от «не передали» → optional/*int в proto3	proto3 nullable — часто спрашивают	🆕 открыто
BUG-08-3	Inventory	stock=0 неотличим от «не передали» → optional/*int	proto3 nullable	🆕 открыто
BUG-08-6	Inventory	Images в UpdateItem приходит [] вместо nil — затирает	proto3 merge semantics	🆕 открыто
P2-9	Authors	UX gap: JWT старый после approve → автоперевыпуск / инструкция	JWT lifecycle	открыто
P2-14	Gateway	Gin legacy /api/* rewrite хрупок → интеграционный тест	Route matching order	открыто
NICK-1	Frontend	Фронт-кэш профиля после смены ника (бэк закрыт v1.1.3.3)	Cache invalidation	⚠️ уточнить
ADMIN-INV-3	Frontend	«По автору» 108 vs 11 (дедупликация)	—	⚠️ проверить UI
—	Gateway	Retry с джиттером + bulkhead	Retry full jitter / Bulkhead	✅ закрыт v1.1.3.7 — вычеркнуть
🟠 УРОВЕНЬ 2 — «1–3 часа, но ценно»
#	Сервис	Задача	Собеседный вес	Статус
P1-16	Payment	Сверка amount webhook с payments.amount	Webhook security / replay	открыто
P2-5	Game	Подключить redis-game anti-cheat (INCR, SETNX, GET, fail-open)	Anti-cheat + fail-open	открыто
P1-3	Inventory	Два драйвера (PG + Mongo) — зафиксировать PG for all, Mongo testing	Strategy / polyglot persistence	открыто
P1-4	Inventory	toPostgresFilter через fmt.Sprintf → без fmt.Sprintf для имён колонок	SQL injection / whitelist	открыто
P1-5	Inventory	Outbox через type assertion → явный TransactionalRepository	Interface segregation	открыто
P1-20	Shop	kafkaProd создаётся, не используется — убрать/подключить	Мёртвый код	открыто
P1-13	Analytics	fmt.Sprintf для db name → параметризовать	SQL injection	открыто
P1-7	MCP	err.Error() — санитизировать	Info leak	открыто
P2-12	Notification	Retry Fixed Delay Telegram → exponential backoff	Retry strategy	открыто
P2-13	Profile	UpdateProfile наружу → отдельный gRPC без total_score	Least privilege	открыто
P1-1	Gateway	Два Hub (app/gateway.go + websocket/hub.go) — унифицировать	Code duplication	открыто
P1-2	Gateway	Async gRPC в admin/users без CB → throughBreaker	Circuit breaker consistency	открыто
P2-10	Analytics	5 субъектов в 1 consumer → разделить	Consumer isolation / lag	открыто
P2-11	History	5 субъектов в 1 consumer → разделить	Consumer isolation	открыто
🔴 УРОВЕНЬ 3 — «день+, P0-блокеры»
#	Сервис	Задача	Собеседный вес	Статус
P0-4	Shop	resumeSagas не берёт started > 60s → деньги списаны, компенсации нет	Saga resume / eventual consistency	открыто
P0-1	Fulfillment	Transactional Outbox (событие теряется между paid и fulfilled)	Outbox / dual-write	открыто
P0-2	Fulfillment	Заменить time.After(10s) на WMS stub с идемпотентным ключом	Idempotency / external API	зависит от P0-1
P0-3	Gateway	Разбить runGateway (God Function >1000 строк) → bootstrap/clients.go, routes.go, middleware.go	God Object decomposition	открыто
P1-6	MCP	SQL Sandbox (string matching) → read-only роль + pgx prepared	SQL injection — сильный ответ	открыто
P1-11	Notification	Poison message из Kafka → Ack + DLQ	Poison message / DLQ	открыто
P1-12	Notification	Fan-Out без rate limit	Throttling	открыто
P1-14	Leaderboard	RestoreFromPostgres при падении PG → retry/backoff	Retry / backoff	открыто
P1-15	Leaderboard	Rate limit на record_beaten	Throttling / event storm	открыто
P1-17	History	RetentionWorker в процессе → K8s CronJob	Scheduled job / reliability	открыто
P1-18	Auth	Ротация JWT_SECRET	Key rotation	открыто
P1-19	Auth	reuse detection для refresh (старый jti повторно → отозвать все сессии)	Refresh token rotation	открыто
P1-21	Shop	cancelPurchase reference_id cancel-{id} может задвоить → Idempotency Key	Idempotency	открыто
P1-22	Shop/Billing/Inventory/Payment	Single-flight на cache stampede	Cache stampede / single-flight	открыто
P1-23	Все	WAF перед Gateway	WAF / perimeter	открыто
🟣 СЦЕНАРИИ (CLEAR_TODO 08.11) — P1
#	Домен	Задача	Собеседный вес	Статус
MOD-1	Модерация	Уведомление автору при удалении/редактировании товара админом + причина	Event-driven notifications	🆕 открыто
MOD-2	Модерация	Админ: причина при редактировании/удалении, soft-delete	Soft-delete / audit	🆕 открыто
MOD-3	Модерация	Автор видит причину, может пересоздать товар	UX + soft-delete	🆕 открыто
MOD-4	Модерация	Уведомление автору при смене цены/описания админом	Event-driven	🆕 открыто
MOD-5	Модерация	Кнопка «Дать/Отозвать подписку» в админке	—	✅ закрыт v1.1.3.3
HIST-1	History	История на фронте ничего не отображается — БАГ	Consumer / durable name	🆕 открыто
HIST-2	History	История — 10 последних записей на пользователя	Pagination / keyset	🆕 открыто
HIST-3	History	Общая статистика под вкладку	Aggregation	🆕 открыто
HIST-4	History	У админа история в JSON — служебная	RBAC	🆕 открыто
HIST-5	History	Почему история сломана — разбор	Debugging	🆕 открыто
CASE-1	Исследование	30 пицц в одном заказе — алерт?	Business rules	🆕 открыто
CASE-2	Исследование	Бабуля натыкала 30 раз — идемпотентность?	Idempotency	🆕 открыто
CASE-3	Исследование	Два игрока купили последнюю карту — Saga compensating тест	Saga compensating	🆕 открыто
🔵 ФРОНТЕНД (P2) — «без мутотени, тык-тык»
#	Задача	Что проверить	Статус
FE-MOB-4	Свайпы для 2048 и Pancaker	Свайп → ход; ручка дёргается	открыто
FE-AUTH-1	Автозаполнение email/пароля после регистрации при переходе на вход	Поле заполнено	открыто
FE-AUTH-2	Глазик (показать/скрыть пароль)	Тык → видно; ещё → скрыто	открыто
FE-AUTH-3	Смена пароля — отдельный сценарий auth (старый пароль, bcrypt, сброс сессий)	Ручка POST /auth/change-password	открыто
⚪ P3 — ROADMAP (не трогаем)
#	Задача
RM-1	Порог покрытия в CI ≥70%
RM-2	Уровни лидерборда 2–3 (сейчас все level=1)
RM-3	BILLING FOR GAMES STRICT TICKET-PER-GAME RULES
RM-4	Единый хром мини-игр + арт тамагочи
RM-5	Batch owned, pgbouncer, 6 Gateway — v1.2 (batch owned ✅ v1.1.3.7) ⚠️
RM-6	Смена пароля (пересекается с FE-AUTH-3)
RM-7	Витрина 108 авторов + 3D (Locked)
RM-8	ORDER-сервис: корзина/заказы/доставка, резерв на время корзины — v1.2+
RM-9	Пагинация каталога на сервере (до 1000+ карт)
RM-10	i18n (RU/EN/DE/ES)
RM-11	API GATEWAY GOD OBJECT DECOMPOSITION (пересекается с P0-3)
RM-12	Ротация JWT_SECRET (пересекается с P1-18)
RM-13	200ms на P=200 → 32 GiB, pgbouncer, batch owned, 6 Gateway — v1.2
RM-14	Upload to prod → eastwessers-event-horizon.com
⚫ P4 — ПАРКОВКА (не трогаем)
#	Задача
—	Loki / Tempo / node-exporter / SLO burn
—	Consul
—	Helm как повседневный деплой
—	Alertmanager → Telegram
—	WAF / пентест (отчёт человека)
—	C4 payouts
—	Miro-стикеры
—	Доля 85/15 — журнал, перевода автору нет
✅ ЗАКРЫТО — вычеркнуть (v1.1.3.5–3.7)
Задача	Где закрыто
P0-5 Billing reference_id	v1.1.3.4
P0-6 Authors soft-delete	v1.1.3.4/3.5
P0-8 Inventory IDOR	v1.1.3.4/3.5
P0-10 Inventory merge	v1.1.3.6
P0-11 AdminCancelPurchase fulfilled	v1.1.3.6
P0-7 Balancer SPOF (2 balancer + health checks)	v1.1.3.6
P1-8 Balancer /ready	v1.1.3.6
P1-9 Balancer static registry	v1.1.3.6 (частично) ⚠️
P1-10 Balancer health checks	v1.1.3.6
P2-1/P2-3 Balancer slog	v1.1.3.6 (Balancer; Gateway — открыто)
P2-16 NICK-1 (бэк)	v1.1.3.3
FE-CARD-1..4	v1.1.3.5
FE-MOB-1..3	v1.1.3.5
ADMIN-INV-1	v1.1.3.5
ADMIN-INV-2 (частично)	v1.1.3.5
MOD-5 кнопка подписки	v1.1.3.3
Retry full jitter + bulkhead Gateway	v1.1.3.7
Gateway ValidateToken кэш 1s	v1.1.3.7
Shop deadlock / purchase reserve / batch owned	v1.1.3.7
Inventory GIN / pg_prewarm / pool	v1.1.3.7
Nginx p99 36s → 8s	v1.1.3.7
Pool 11 сервисов	v1.1.3.7
Итого открыто:

🟢 Уровень 0: 8 (1 ⚠️)

🟡 Уровень 1: 7 (2 ⚠️)

🟠 Уровень 2: 14

🔴 Уровень 3: 15

🟣 Сценарии: 12 (1 ✅ вычеркнуть)

🔵 Фронт: 4

⚪ P3: 14 (1 ⚠️, 2 пересекаются)

⚫ P4: 8

✅ Закрыто: 22

Суммарно активных: ~60 (P0 4 + P1 ~18 + P2 ~17 + сценарии ~12 + фронт 4 + P3 14, минус пересечения).


ADDITIONAL TASKS


### Игровая оптимизация (Приоритет 3)
| # | Задача | Статус |
|---|--------|--------|
| GAME-PERF-1 | Memonia: 30 карточек — `React.memo`, `useCallback` | открыто |
| GAME-PERF-2 | Flappy: 60 FPS стабильно — проверить steady state | открыто |
| GAME-PERF-3 | Gears: физика через `requestAnimationFrame` — проверить steady state | открыто |
| GAME-PERF-4 | Hanoi: анимация через `transform`, не `top/left` | открыто |
| GAME-UI-1 | Единый `GameShell` для всех игр (пересекается с «единый хром») | открыто |
| GAME-UI-2 | Кнопка «На главную» слева, буст в `controls` последним | открыто |
| GAME-UI-3 | `GameHelp` с per-game текстами | открыто |
| GAME-LB-1 | Отображение ранга игрока после submit | открыто |
| GAME-LB-2 | Единый формат очков | открыто |
| GAME-MECH-1 | Memonia: level-up механики | открыто |
| GAME-MECH-2 | Gears: визуальные подсказки | открыто |
| GAME-MECH-3 | Hanoi: улучшить анимацию перекладывания | открыто |
| GAME-NEW-1 | Новая игра для разнообразия (возможно) | идея |

### Пользователи и продакшн (Приоритет 4)
| # | Задача | Статус |
|---|--------|--------|
| PROD-METRIC-1 | LTV / ARPU (DAU/MAU/retention/когорты — уже есть в analytics) | открыто |
| PROD-FEEDBACK-1 | Форма фидбека | открыто |
| PROD-FEEDBACK-2 | In-app опрос | открыто |
| PROD-FEEDBACK-3 | Discord/Slack канал | открыто |
| PROD-MKT-1 | SEO-метатеги | открыто |
| PROD-MKT-2 | Open Graph | открыто |
| PROD-MKT-3 | Twitter/X-бот для анонсов | открыто |
| PROD-LEGAL-1 | Terms of Service | открыто |
| PROD-LEGAL-2 | Privacy Policy | открыто |
| PROD-LEGAL-3 | GDPR compliance | открыто |
| PROD-LEGAL-4 | Cookie banner | открыто |

### Observability (Приоритет 5) — расширение
| # | Задача | Статус |
|---|--------|--------|
| OBS-TRACE-1 | OTel/Jaeger в остальные 10 сервисов (сейчас в 7) | открыто |
| OBS-TRACE-2 | Sampling policy | открыто |
| OBS-LOG-1 | Log aggregation (Loki) — пересекается с P4 | открыто |
| OBS-SLO-1 | Определить SLO: 99.9% uptime, p99 < 500ms | открыто |
| OBS-SLO-2 | Мониторить error budget | открыто |
| OBS-SLO-3 | Incident response | открыто |
| OBS-GRAF-1 | Grafana dashboards (есть, но не все — проверить) | ⚠️ уточнить |

### Безопасность (Приоритет 5) — расширение
| # | Задача | Статус |
|---|--------|--------|
| SEC-AUTH-1 | MFA | открыто |
| SEC-API-1 | WAF — пересекается с P1-23 | дубль |
| SEC-API-2 | DDoS protection | открыто |
| SEC-DATA-1 | Encryption at rest | открыто |
| SEC-DATA-2 | Encryption in transit (TLS) — не везде | открыто |
| SEC-DATA-3 | Secrets management (Vault) | открыто |
| SEC-AUDIT-1 | Audit log | открыто |
| SEC-AUDIT-2 | PII detection | открыто |
| SEC-TEST-1 | Penetration testing — пересекается с P4 | дубль |
| SEC-TEST-2 | Fuzzing | открыто |
| SEC-TEST-3 | SAST / DAST | открыто |

### Идеи на потом (из FUTURE_EVENT_HORIZON.md → «Идеи на потом»)
| # | Задача | Статус |
|---|--------|--------|
| IDEA-1 | Сервер за 50$ (8 vCPU / 32 GiB — Hetzner/DO/Vultr) | идея |
| IDEA-2 | Довести механику лампочек (экономика, визуализация, ежедневный бонус) | идея |
| IDEA-3 | `make ride-on` — deploy + npm install + npm run dev + «ALL SYSTEMS NOMINAL» | идея |
| IDEA-4 | Фича «Ваш рекорд побили» — уведомление в BELL (record_beaten уже есть — проверить фронт) | ⚠️ уточнить |
| IDEA-5 | `make seed-v110` заметка — пересекается с P2-15 | дубль |

⚠️ Что надо поправить (5 мест)
P2-1 Gateway log.Printf → slog — в P2-1/P2-3 закрытых стоит «Balancer slog ✅, Gateway — открыто». Но OBS-LOG-1 в ADDITIONAL говорит «Gateway там log.Printf» — не дублируй, оставь P2-1 единственным местом, а в OBS-LOG-1 убери упоминание Gateway.

GAME-UI-1 «Единый GameShell» — пересекается с RM-4 «единый хром мини-игр» (P3) и с парковкой. Сведи в одну строку: RM-4 = GAME-UI-1 + GAME-UI-2 + GAME-UI-3. Иначе три задачи про одно.

IDEA-4 «Ваш рекорд побили» — помечено ⚠️, но по факту: record_beaten в Leaderboard есть (v1.1.1, Event Chaining), Notification его потребляет, in-app bell есть. Не хватает только deep-link на фронте (/leaderboard?game=X&level=Y&highlight=me). Переформулируй: IDEA-4: record_beaten → deep-link на фронте (бэк готов).

OBS-GRAF-1 ⚠️ — в README «Grafana provision fix» (v1.1.0), в FUTURE_EVENT_HORIZON «dashboards нет». Противоречие. Проверь monitoring/grafana/provisioning/ — если дашборды провижены, вычеркни; если нет — оставь открытым.

SEC-AUTH-1 MFA — это не «безопасность расширение», это фича продукта. MFA требует UI (QR-код, recovery codes), flow в Auth, хранение TOTP-секрета. Поставь в P3 (roadmap), а не в SEC-*.

🟡 Что не хватает (2 места)
P2-GAME-STRATEGY — SubmitScore god-function switch на 8 игр (из FUTURE_EVENT_HORIZON #17 средних). В твоём списке его нет. В P2-5 только anti-cheat. Добавь отдельной строкой в Уровень 2: «Game: вынести per-game strategy interface (Strategy pattern, собеседный вес высокий)».

«Не улучшать» блок — если в MEGA_TECH_DEBT.md его нет, добавь снизу:

markdown
## 🚫 Не улучшать (зафиксировано)
- Saga в Shop — работает.
- Circuit Breaker в Gateway — минимальный, работает (+ bulkhead v1.1.3.7).
- Least Connections в Balancer — правильный.
- Double-Entry Ledger в Billing — фундамент.
- Формулы очков, стоимость бустов (10 ламп), награды — сбалансированы.
- Свой crypto / свой auth — не делаем.
📊 Итоговая сводка — сходится?
Уровень	В твоём файле	Сверка
🟢 Уровень 0	8 (1 ⚠️)	✅ ок
🟡 Уровень 1	7 (2 ⚠️)	⚠️ P2-GAME-STRATEGY не хватает → станет 8
🟠 Уровень 2	14	⚠️ +P2-GAME-STRATEGY → 15
🔴 Уровень 3	15	✅ ок
🟣 Сценарии	12 (1 ✅)	✅ ок (MOD-5 вычеркнуть)
🔵 Фронт	4	✅ ок
⚪ P3	14	⚠️ RM-4 = GAME-UI-1..3 (свести), SEC-AUTH-1 MFA сюда
⚫ P4	8	✅ ок
🆕 ADDITIONAL	13+11+7+11+5 = 47	⚠️ GAME-UI-1 дубль RM-4, OBS-LOG-1 дубль P2-1, IDEA-4 переформулировать
Активных сходится: ~60 + 47 additional − дубли (GAME-UI-1, OBS-LOG-1, IDEA-5) ≈ 104. Но это уже не «60 активных», а полная карта. Нормально — теперь ты видишь всё.