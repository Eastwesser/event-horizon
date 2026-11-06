🧾 TODOLISTISHE — агрегированный (Event Horizon, ноябрь 2026)
Легенда:

🟥 P0 — блокер

🟧 P1 — надёжность/безопасность

🟨 P2 — гигиена/UX

🟦 P3 — roadmap

🟩 P4 — долгосрок/парковка

✅ Закрыто — не тащим

[дубль] — слито из нескольких источников

[противоречие] — расхождение источников

[упущено] — не в README, но есть в бреже

[проверить] — статус неясен

[проверки] — из аудита, отдельной секцией

🟥 P0 — Блокеры (нельзя в прод)
#	Задача	Домен	Источник	Пометка
P0-1	Fulfillment: внедрить Transactional Outbox — событие теряется между purchase.paid и purchase.fulfilled	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README (не отражено)	[упущено] [дубль]
P0-2	Fulfillment: заменить time.After(10s) на реальный WMS API	Бэкенд	TECH_DEBT_15.10, README («заменится на WMS»)	[дубль]
P0-3	Gateway: разбить runGateway God Function (>1000 строк) на bootstrap/clients.go, routes.go, middleware.go	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P0-4	Shop: Saga started > 60s не берётся resumeSagas — деньги списаны, компенсации нет. Воркер должен брать started, проверять Billing.GetTransactions(spend_ref)	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, FINAL_AUDIT	[дубль] [упущено]
P0-5	Billing: RewardConsumer использует time.Now().UnixNano() для reference_id → двойное начисление при redelivery. Использовать event.ReferenceID	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] [упущено]
P0-6	Authors: RevertApplication hard delete — FK падает. Soft delete + check FKs	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль] [упущено]
P0-7	Balancer: single point of failure — нет резервного. Несколько Balancer + VIP/DNS	Девопс	BACKEND_BIG_TECH_DEBT, EH_BOARDER_CASES	[дубль] [упущено]
P0-8	Inventory: IDOR в UpdateItem/DeleteItem — на бэке нет проверки author_id. Фронт скрывает кнопки, но API открыт	Бэкенд	PATCH_NOTES_1.1.3.1, AUTHORS_AUDIT, README (упомянут в патче)	[дубль] [проверить] — бэк не починен
P0-9	Shop: paymentClient может быть nil — мерч-гейт деградирует. Метрика merch_gate_disabled	Бэкенд	TECH_DEBT_15.10, EVENT_HORIZON_PATTERNS	[упущено]
Итого P0: 9 задач (было 6 в одном файле, 7 в другом — после слияния и добавления IDOR + paymentClient).

🟧 P1 — Надёжность / безопасность
#	Задача	Домен	Источник	Пометка
P1-1	Gateway: два Hub (app/gateway.go + websocket/hub.go) — унифицировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-2	Gateway: async gRPC в admin/users без CB — обернуть в throughBreaker	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-3	Inventory: два драйвера (PG + Mongo) — определиться с одним	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-4	Inventory: toPostgresFilter через fmt.Sprintf — переписать без fmt.Sprintf для имён колонок	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-5	Inventory: Outbox через type assertion — явный интерфейс TransactionalRepository	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-6	MCP: SQL Sandbox string matching — обходится. Read-only роль + pgx prepared	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-7	MCP: err.Error() — утечка внутренних деталей — санитизировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-8	Balancer: /ready не проверяет бэкенды — реальная проверка хотя бы одного	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-9	Balancer: static registry — нет service discovery. etcd/consul/k8s	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-10	Balancer: нет health checks бэкендов — периодический ping /health	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-11	Notification: poison message из Kafka — бесконечный retry. Ack + DLQ	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-12	Analytics: fmt.Sprintf для db name — параметризовать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-13	Leaderboard: RestoreFromPostgres при падении PG — пустой leaderboard. Retry/backoff	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-14	Payment: нет сверки суммы webhook с payments.amount — сверять amount	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-15	History: RetentionWorker в процессе — при рестарте сброс. K8s CronJob	Бэкенд/Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-16	Auth: нет ротации JWT_SECRET — ротация	Бэкенд/Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README	[дубль]
P1-17	Все: нет WAF — WAF перед Gateway	Девопс	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT, README	[дубль]
P1-18	Shop: kafkaProd создаётся, не используется — убрать или подключить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-19	Shop: cancelPurchase reference_id = cancel-{id} — может задвоить. Idempotency Key	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P1-20	Shop: не хватает single-flight на cache stampede (Billing, Inventory, Shop, Payment)	Бэкенд	EH_BOARDER_CASES, FINAL_AUDIT	[упущено]
P1-21	Leaderboard: нет rate limit на record_beaten — лавина при массовых рекордах	Бэкенд	EH_BOARDER_CASES, FINAL_AUDIT	[упущено]
P1-22	Notification: Fan-Out без rate limit	Бэкенд	EH_BOARDER_CASES	[упущено]
P1-23	Auth: reuse detection для refresh — если старый jti использован повторно → отозвать все сессии	Бекенд	EH_SHATTER_AND_ISSUE_SOLVING	[упущено]
Итого P1: 23 задачи (было 20 в одном файле, 12 в другом — после слияния и добавления single-flight, rate limit, reuse detection).

🟨 P2 — Гигиена / UX (бэкенд)
#	Задача	Домен	Источник	Пометка
P2-1	Gateway: log.Printf везде — перейти на slog	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-2	Balancer: пустые пакеты backend/, proxy/ — удалить или наполнить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-3	Balancer: log.Printf в hot path — structured logging	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-4	Fulfillment: newUUID (crypto/rand) vs uuid.NewString() — унифицировать	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-5	Game: redis-game :6380 не подключён — реализовать anti-cheat по GAME_REDIS.md (INCR submit, SETNX dedup, GET score cache, fail-open). Оставляем.	Бэкенд	GAME_REDIS.md, GAME_REDIS_ANTICHEAT.md, README	[противоречие] README vs GAME_REDIS.md — решено: подключаем
P2-6	Billing: GetBalance возвращает -1 при cache miss — (int, bool)	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-7	Auth: boosty.to URL захардкожен — Config	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-8	Payment: subscriptionDays default 30 — явно в config	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-9	Authors: UX gap — JWT старый после approve — автоперевыпуск JWT или инструкция	Бэкенд/Фронт	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-10	Analytics: 5 субъектов в одном consumer'е — разделить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-11	History: 5 субъектов в одном consumer'е — разделить	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-12	Notification: Retry Fixed Delay для Telegram — exponential backoff	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-13	Profile: UpdateProfile наружу не открыт (умеет переписать total_score) — отдельный gRPC без total_score	Бэкенд	TECH_DEBT_15.10, BACKEND_BIG_TECH_DEBT	[дубль]
P2-14	FUTURE_TASKS: Gin matches routes before middleware v1 — legacy /api/* rewrite хрупок. Проверить в новых Gin, добавить интеграционный тест	Бэкенд	FUTURE_TASKS.md, BACKEND_BIG_TECH_DEBT (критичные)	[упущено] — в позднем P0/P1 нет
P2-15	FUTURE_TASKS: make seed-v110 заметка — описать в DAILY_OPS.md / README.md, что засеивает	Девопс/Док	FUTURE_TASKS.md, BACKEND_BIG_TECH_DEBT (мелкие)	[упущено]
P2-16	BACKEND_BIG_TECH_DEBT: poison message в Notification → Ack + лог — уже в P1-11, но пометить как [дубль]	—	—	[дубль]
Итого P2: 15 задач (было 12 в одном, 13 в другом — после слияния + Gin + seed-v110).

🟨 P2 — Фронтенд: мелкие баги карточек
Из FRONTEND_AUTHORS_FIX.md «Что осталось починить»:

#	Задача	Источник	Пометка
FE-CARD-1	Стоимость (прочерк) — добавить cost: market_rub в attributes	FRONTEND_AUTHORS_FIX	[упущено]
FE-CARD-2	Ход (прочерк) — добавить поле move в форму или автозаполнение для существ	FRONTEND_AUTHORS_FIX	[упущено]
FE-CARD-3	Бейдж «Товар» вместо «Карточка» в магазине — починить shopItemMap.ts	FRONTEND_AUTHORS_FIX	[упущено]
FE-CARD-4	Класс карты (прочерк) — добавить поле class в форму (массив строк)	FRONTEND_AUTHORS_FIX	[упущено]
🟨 P2 — Фронтенд: мобилки
Из FRONTEND_AUTHORS_FIX.md «Следующие блоки (Мобилки)»:

#	Задача	Источник	Пометка
FE-MOB-1	Убрать горизонтальный скролл (overflow-x-hidden)	FRONTEND_AUTHORS_FIX	[упущено]
FE-MOB-2	Отступы по бокам (лонгрид)	FRONTEND_AUTHORS_FIX	[упущено]
FE-MOB-3	Центрирование модалок и колокольчика	FRONTEND_AUTHORS_FIX	[упущено]
FE-MOB-4	Свайпы для 2048 и Pancaker	FRONTEND_AUTHORS_FIX	[упущено]
🟨 P2 — Фронтенд: Auth UX (твои правки)
#	Задача	Источник	Пометка
FE-AUTH-1	Автозаполнение email/пароля после регистрации при переходе на вкладку входа	Твои правки 06.11	новое
FE-AUTH-2	Глазик (показать/скрыть пароль) на форме регистрации/входа	Твои правки 06.11	новое
FE-AUTH-3	Смена пароля — отдельный сценарий auth (старый пароль, bcrypt, сброс сессий), не недостающий глагол	README, TECH_DEBT	[дубль]
🟨 P2 — Фронтенд/Бэкенд: Ник в профиле
Из твоего TODOLISTISHE:

#	Задача	Домен	Источник	Пометка
NICK-1	Научить ник сохраняться в профиле на фронте — на бэке правильно, на фронте при смене ника как будто захардкожено берётся с бэка. Проверить: есть ли nickname в бэке? В админке видно, что функционал есть (darkstar_2026 → Darkstar). Починить инвалидацию кэша профиля (GET /profile заново)	Фронт/Бэк	TODOLISTISHE 06.11, FRONTEND_AUTHORS_FIX (скриншот 7)	[дубль] с FRONTEND_AUTHORS_FIX
🟨 P2 — Админка: статистика инвентаря
Из твоего TODOLISTISHE:

#	Задача	Домен	Источник	Пометка
ADMIN-INV-1	Переименовать метрики: «Всего товаров» → «Уникальных позиций»; «Сток» → «Общее кол-во товаров в наличии»	Фронт	TODOLISTISHE 06.11	новое
ADMIN-INV-2	Починить «По типу»: пропуск — 4 карточки (новые, добавленные под авторами) отображаются с пустым типом (блоб с пустым местом). Проверить маппинг type для новых карточек	Фронт/Бэк	TODOLISTISHE 06.11	новое [проверить]
ADMIN-INV-3	Починить «По автору»: засеяно одним юзером 108 авторов → «засевный автор 1» + остальные. Дедупликация/разделение	Фронт/Бэк	TODOLISTISHE 06.11, FRONTEND_AUTHORS_FIX (108 — статика, реальных 11)	[дубль] [проверить]
🟧 P1 — Новые сценарии: уведомления авторам / админ-модерация
Из твоего TODOLISTISHE (сценарий 1–3):

#	Задача	Домен	Источник	Пометка
MOD-1	Уведомление автору при удалении товара админом — «ваш товар снят с продажи / отозван администратором» + «подробнее» → окно «что отозвано и причину»	Бэк/Фронт	TODOLISTISHE 06.11	новое
MOD-2	Админ: при редактировании/удалении товара спрашивать причину — записывать в окно, товар soft-delete	Фронт/Бэк	TODOLISTISHE 06.11	новое
MOD-3	Автор видит причину удаления и может заново создать товар (фото + описание)	Фронт/Бэк	TODOLISTISHE 06.11	новое
MOD-4	Уведомление автору, если админ поменял описание или цену товара	Бэк/Фронт	TODOLISTISHE 06.11	новое
MOD-5	Админ: кнопка «Дать подписку» / «Отозвать» рядом со статусом подписки (активна–неактивна)	Фронт/Бэк	TODOLISTISHE 06.11	новое
MOD-6	Аппрув товара админом перед публикацией (цензура, чтобы не постили хентай). Happy path — авторы публикуют	Бэк/Фронт	TODOLISTISHE 06.11	новое [проверить] — возможно, противоречит текущей модели «автор создаёт → Shop через Outbox»
🟧 P1 — История (History service)
Из твоего TODOLISTISHE:

#	Задача	Домен	Источник	Пометка
HIST-1	Починить историю на фронте — ничего не отображается. Проверить, настроен ли бэк на сохранение и выдачу	Фронт/Бэк	TODOLISTISHE 06.11, AUTHORS_AUDIT	новое [проверить]
HIST-2	История — 10 последних записей на пользователя (действия на аккаунте), показывать каждому отдельно	Бэк/Фронт	TODOLISTISHE 06.11	новое
HIST-3	Общая статистика — показывать в окне, где пользователь открывает, актуальную под вкладку	Бэк/Фронт	TODOLISTISHE 06.11	новое
HIST-4	У админа история в JSON (регистрация, рекорды, покупки, оплата, авторы) — служебная информация. Проверить, у авторов и пользователей	Бэк/Фронт	TODOLISTISHE 06.11	новое [проверить]
HIST-5	Проверить, почему история сломана — возможно, потому что служебная и скрыта	Бэк	TODOLISTISHE 06.11	[проверить]
🟧 P1 — Кейсы «бабуля натыкала 30 пицц» + SAGA
Из твоего TODOLISTISHE:

#	Задача	Домен	Источник	Пометка
CASE-1	Разобрать кейс: 30 пицц в одном заказе — нужен ли алерт/звонок менеджера при >10 в заказе?	Бэк/Фронт	TODOLISTISHE 06.11	новое [проверить]
CASE-2	Разобрать кейс: бабуля словила лаг, нажала 30 раз «оформить» — что с сервисами? Легли? Идемпотентность?	Бэк	TODOLISTISHE 06.11	новое [проверить]
CASE-3	Разобрать кейс: два игрока купили последнюю легендарную карту — одного откатим по SAGA, вернём средства + «товар закончился». Проверить, реализовано ли	Бэк	TODOLISTISHE 06.11	новое [проверить] — в бреже есть Shop Saga, но конкретный кейс «stock=0 → compensating → compensated» разобран, надо проверить на практике
🟨 P2 — Прочее (из FUTURE_TASKS / BOOST_PLAN / DIGITAL_MERCH)
#	Задача	Домен	Источник	Пометка
MISC-1	make ride-on — one-command DX: make deploy → make status → cd frontend && npm install && npm run dev → последнее сообщение ALL SYSTEMS NOMINAL	Девопс	FUTURE_TASKS.md	[упущено] — в README нет
MISC-2	«Ваш рекорд побили» — уведомление на BELL, когда другой игрок побил рекорд. Backend: leaderboard уже публикует record_beaten → notification слушает. Frontend: NotificationBell.tsx — иконка/цвет, ссылка /leaderboard?game=X&level=Y&highlight=me. Один рекорд — одно уведомление	Бэк/Фронт	FUTURE_TASKS.md	[упущено] — в README нет
MISC-3	Унификация интерфейса игр — единый GameShell, GameOverActions, GameHelp; одинаковое кол-во кнопок и позиции. Чек-лист: GAME_CHROME_CHECKLIST.md (создать)	Фронт	FUTURE_TASKS.md, README («единый хром»)	[дубль] с README «единый хром»
MISC-4	Сервер за $50 — 8 vCPU / 32 GiB (Hetzner/DO/Vultr), 50–100 VU стабильно, ~500 RPS peak. Когда появятся первые реальные пользователи	Девопс	FUTURE_TASKS.md, README («8 vCPU/32 GiB»)	[дубль]
MISC-5	Механика лампочек до ума — визуализация трат (профиль → «Мои лампы»), ежедневный бонус, серия 7 дней → +50 ламп. Буст не должен быть обязательным	Бэк/Фронт	BOOST_PLAN.md, FUTURE_TASKS.md	[упущено] — в README только «механика лампочек»
MISC-6	BOOST_PLAN: копирайт эффекта у Старта, одинаковое место галочки во всех GameShell, экран конца партии «в топ / не в топ», Tamagotchi без чекбокса буста	Фронт	BOOST_PLAN.md	[упущено]
MISC-7	DIGITAL_MERCH: isDigital учитывать не только game_id, но и category/type (game_skin, profile_theme, theme, скин, тема) + fallback-иконки для цифровых (иконка игры для скинов, sparkle для тем, не images из инвентаря)	Фронт	DIGITAL_MERCH_ISSUE.md	[проверить] — README говорит isDigitalItem сделан, но про category/type и fallback — нет
🟦 P3 — Roadmap (из README + TECH_DEBT)
#	Задача	Домен	Источник	Пометка
RM-1	i18n (RU/EN/DE/ES) — i18next + react-i18next, волны 1–4. Волна 1: AppNavbar, Login, Register, Home, PageHeader, GameShell. Волна 2: игры. Волна 3: магазин. Волна 4: admin	Фронт	I18_INSIGHTS.md, README, TECH_DEBT	[дубль]
RM-2	Пагинация каталога на сервере — до 1000+ карт	Бэк	README, TECH_DEBT	[дубль]
RM-3	Уровни лидерборда 2–3 — сейчас все level=1	Бэк	README, TECH_DEBT	[дубль]
RM-4	Порог покрытия в CI ≥70% — юнит-тесты есть, порог не включён	Девопс	README, TECH_DEBT	[дубль]
RM-5	Retry с джиттером, bulkhead — у gateway уже есть CB	Бэк	README, TECH_DEBT	[дубль]
RM-6	108 авторов — витрина без текстов/картинок/согласия. Locked until product OK. Реальных 11, 108 — статика	Бэк/Данные	README, TECH_DEBT, FRONTEND_AUTHORS_FIX	[дубль] [противоречие] 108 vs 11 — уточнено
RM-7	3D — нет ассетов и клиента. Locked	Фронт	README, TECH_DEBT	[дубль]
RM-8	Helm как повседневный деплой — чарт неполный (нет Redis, payment, authors, history, notification и баз). Locked	Девопс	README, TECH_DEBT	[дубль]
🟩 P4 — Долгосрок / парковка
#	Задача	Домен	Источник	Пометка
PARK-1	Loki / Tempo / node-exporter / SLO burn / multi-DB scrape — dream list	Девопс	README, DREAM_DEVOPS_IDEAS	[дубль]
PARK-2	Consul, Alertmanager → Telegram — парковка	Девопс	README	[дубль]
PARK-3	WAF / пентест — контур перед приложением + отчёт человека	Девопс	README, TECH_DEBT	[дубль]
PARK-4	C4 payouts — locked	Бэк	README	[дубль]
PARK-5	Miro-стикеры — IRL	Прочее	README	[дубль]
PARK-6	Единый хром мини-игр и арт тамагочи — soft polish	Фронт	README	[дубль] с MISC-3
🔍 Секция «Проверки» (аудит, из AUTHORS_AUDIT.md)
#	Проверка	Домен	Источник
AUD-1	IDOR-бэк: убедиться, что author_id проверяется на UpdateItem/DeleteItem (сейчас только фронт скрывает)	Бэк	AUTHORS_AUDIT, PATCH_NOTES
AUD-2	«Что если все товары у автора куплены» — что происходит у автора, у игрока	Бэк/Фронт	AUTHORS_AUDIT
AUD-3	Собрать метрики и сделать скриншоты (Grafana)	Девопс	AUTHORS_AUDIT
AUD-4	Собрать NATS-метрики	Девопс	AUTHORS_AUDIT
AUD-5	Что с History service — почему пусто на фронте, настроен ли бэк на выдачу	Бэк/Фронт	AUTHORS_AUDIT
AUD-6	Пройти чек-лист AUTHORS_AUDIT шаги 0–5: окружение, засев, синхронизация Inventory→Shop, покупка, история, метрики/админка	Все	AUTHORS_AUDIT
AUD-7	Проверить SQL по каждому шагу (Inventory, outbox, Billing, History, Analytics)	Бэк	AUTHORS_AUDIT
AUD-8	Проверить, что price = market_rub * 1000, card_text не пустой, author_id привязан	Бэк	AUTHORS_AUDIT
AUD-9	Проверить, что stock уменьшился, баланс списался, автору начислились 85%	Бэк	AUTHORS_AUDIT
✅ Закрыто (не тащим)
v1.1.3.1 патч (AppNavbar, Modal, InventoryEditModal isCard/card_text/цена/version/двойной updateItem, validateItem).

v1.1.3 (Patterns catalog, roadmap, фронт-полировка, BoostCheckbox, Pancaker, Gears canvas fix, Hanoi, Memonia, Flappy, isDigitalItem для 998 stock).

v1.1.2 (Tamagotchi, Builder, Gears, Memonia, Hanoi, 2048, Pancaker, Flappy).

v1.1.1 (Spend Cycle, purchase_sagas, WS ticket, Docs Gate, Revenue Split, Redis 503, Kafka молчит).

v1.1.0 (/api/v1, boost gate, Pancaker score fix, Hanoi multipliers, Gears skins, Companion daily gift, LB level-1, profile edit disclosure).

v1.0.9 (Berserk 280, cancel purchase, Noiz, CardImage, slim DTO).

v1.0.8 (secrets, thin deploy, purchase on NATS, response cache + _partial, Boosty HMAC, password min 8).

v1.0.7 (Payment/Authors/History/Analytics, CB, Game Outbox, MCP/RAG, OpenAPI sync).

v1.0.6 (Inventory, CI/CD, Ansible, k3s, Makefile).

v1.0.5 (Shop, блины, Memory скин, Flappy, Towers, purchased_at, NATS кластер, WS через балансировщик).

GEARS_ISSUE (skinsLoading в useEffect) — закрыто в v1.1.2.

FRONTEND_AUTHORS_FIX блоки 1–4 (картинки, профили, синхронизация, UI карточек) — [x].

📊 Сводка по числам
Приоритет	Кол-во	Было в источниках
P0	9	6 (TECH_DEBT_15.10) / 7 (BACKEND_BIG) → +IDOR, +paymentClient
P1	23	12 / 20 → +single-flight, +rate limit, +reuse detection
P2 (бэкенд)	15	12 / 13 → +Gin, +seed-v110
P2 (фронт: карточки)	4	новое из FRONTEND_AUTHORS_FIX
P2 (фронт: мобилки)	4	новое из FRONTEND_AUTHORS_FIX
P2 (фронт: auth UX)	3	новое (автозаполнение, глазик) + смена пароля
P2 (ник)	1	новое
P2 (админка: инвентарь)	3	новое
P1 (модерация/уведомления)	6	новое
P1 (история)	5	новое
P1 (кейсы)	3	новое
P2 (прочее)	7	FUTURE_TASKS, BOOST_PLAN, DIGITAL_MERCH
P3 (roadmap)	8	README + TECH_DEBT
P4 (парковка)	6	README
Проверки (аудит)	9	AUTHORS_AUDIT
Закрыто	~15 релизов + 2 фикса	—