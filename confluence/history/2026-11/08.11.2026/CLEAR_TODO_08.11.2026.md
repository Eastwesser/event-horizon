🔥 Что мы сделали сегодня (08.11, зафиксировать)
Это уже в коде, не трогаем, только помним:

✅ AdminCancelPurchase — proto + gRPC + Gateway + БД, проверено вживую (REFUNDED в БД).

✅ checkAuthorAccess — админ вседозволенный (PUT под админом прошёл).

✅ UpdateItem — обнаружен баг обнуления полей (в TODO ниже).

✅ Nginx, SERVICE_WEIGHT, REDIS_SPOF — документация, не код.

🟥 P0 — Блокеры. Без них не «прод», но и не «срочно сегодня»
Порядок: сверху вниз. Это то, что ты уже знаешь и что точно надо сделать. Не распыляйся.

#	Сервис	Задача	Почему P0
P0-1	Fulfillment	Внедрить Transactional Outbox — сейчас событие теряется между purchase.paid и purchase.fulfilled	Если упадёт — заказ «завис», fulfilled_at null. Техдолг, но деньги/товар.
P0-2	Fulfillment	Заменить time.After(10s) на реальный WMS API (или хотя бы stub с идемпотентным ключом)	Заглушка, но блокирует честный outbox.
P0-3	Gateway	Разбить runGateway (God Function >1000 строк) на bootstrap/clients.go, routes.go, middleware.go	Долгая инициализация, сложно поддерживать.
P0-4	Shop	resumeSagas не берёт строку started > 60s → деньги списаны, компенсации нет. Воркер должен брать started, проверять Billing.GetTransactions(spend_ref)	Деньги. Это самый настоящий P0.
P0-5	Billing	RewardConsumer использует time.Now().UnixNano() для reference_id → двойное начисление при redelivery. Использовать event.ReferenceID	Деньги.
P0-6	Authors	RevertApplication hard delete — FK падает. Soft delete + check FKs	Approve не откатывается, если у автора есть заказы.
P0-7	Balancer	Single point of failure — нет резервного. 2 Balancer + VIP/DNS (или Nginx перед ними — у тебя уже есть NGINX_KEYIDEA.md)	Всё упало — всё недоступно.
P0-8	Inventory	IDOR в UpdateItem/DeleteItem — на бэке нет проверки author_id. Фронт скрывает, но API открыт	Проверено: checkAuthorAccess есть в grpc_handler.go, но надо проверить, что оно реально вызывается в UpdateItem/DeleteItem, а не только в SoftDelete/Restore.
P0-9	Shop	paymentClient может быть nil — мерч-гейт деградирует. Метрика merch_gate_disabled	Без подписки можно купить мерч.


P0-11	Shop	AdminCancelPurchase должен обходить fulfilled_at для админа. Сейчас 400 «Товар уже отправлен»	Ты сам записал в TODO. Проверено вживую.

Итого P0: 11. Это не «сегодня», это «следующие 3–5 рабочих сессий». Бери по одной, не параллель.

🟧 P1 — Надёжность / безопасность
Это можно делать после P0, или параллельно, если P0-задача упёрлась в чужой код.

#	Сервис	Задача
P1-1	Gateway	Два Hub (app/gateway.go + websocket/hub.go) — унифицировать
P1-2	Gateway	Async gRPC в admin/users без CB — обернуть в throughBreaker
P1-3	Inventory	Два драйвера (PG + Mongo) — определиться с одним. Ты уже решил: PG for all, Mongo — testing. Зафиксировать в коде/README.
P1-4	Inventory	toPostgresFilter через fmt.Sprintf — переписать без fmt.Sprintf для имён колонок
P1-5	Inventory	Outbox через type assertion — явный интерфейс TransactionalRepository
P1-6	MCP	SQL Sandbox string matching — обходится. Read-only роль + pgx prepared
P1-7	MCP	err.Error() — утечка внутренних деталей. Санитизировать
P1-8	Balancer	/ready не проверяет бэкенды. Реальная проверка хотя бы одного
P1-9	Balancer	Static registry — нет service discovery. etcd/consul/k8s
P1-10	Balancer	Нет health checks бэкендов. Периодический ping /health
P1-11	Notification	Poison message из Kafka — бесконечный retry. Ack + DLQ
P1-12	Notification	Fan-Out без rate limit
P1-13	Analytics	fmt.Sprintf для db name — параметризовать
P1-14	Leaderboard	RestoreFromPostgres при падении PG — пустой leaderboard. Retry/backoff
P1-15	Leaderboard	Нет rate limit на record_beaten — лавина при массовых рекордах
P1-16	Payment	Нет сверки суммы webhook с payments.amount — сверять amount
P1-17	History	RetentionWorker в процессе — при рестарте сброс. K8s CronJob
P1-18	Auth	Нет ротации JWT_SECRET
P1-19	Auth	reuse detection для refresh — если старый jti использован повторно → отозвать все сессии
P1-20	Shop	kafkaProd создаётся, не используется — убрать или подключить
P1-21	Shop	cancelPurchase reference_id = cancel-{id} — может задвоить. Idempotency Key
P1-22	Shop / Billing / Inventory / Payment	Single-flight на cache stampede
P1-23	Все	Нет WAF — WAF перед Gateway
Итого P1: 23.

🟨 P2 — Гигиена / UX (бэкенд) — быстрые фиксы
#	Сервис	Задача
P2-1	Gateway	log.Printf везде — перейти на slog
P2-2	Balancer	Пустые пакеты backend/, proxy/ — удалить или наполнить
P2-3	Balancer	log.Printf в hot path — structured logging
P2-4	Fulfillment	newUUID (crypto/rand) vs uuid.NewString() — унифицировать
P2-5	Game	redis-game подключить anti-cheat по GAME_REDIS.md (INCR submit, SETNX dedup, GET score cache, fail-open). Решено: подключаем, оставляем.
P2-6	Billing	GetBalance возвращает -1 при cache miss — (int, bool)
P2-7	Auth	boosty.to URL захардкожен — Config
P2-8	Payment	subscriptionDays default 30 — явно в config
P2-9	Authors	UX gap — JWT старый после approve — автоперевыпуск JWT или инструкция
P2-10	Analytics	5 субъектов в одном consumer'е — разделить
P2-11	History	5 субъектов в одном consumer'е — разделить
P2-12	Notification	Retry Fixed Delay для Telegram — exponential backoff
P2-13	Profile	UpdateProfile наружу не открыт — отдельный gRPC без total_score
P2-14	Gateway	Gin matches routes before middleware v1 — legacy /api/* rewrite хрупок. Проверить в новых Gin, добавить интеграционный тест
P2-15	DevOps	make seed-v110 заметка — описать в DAILY_OPS.md / README.md
P2-16	Ник в профиле	На фронте при смене ника берётся захардкоженный ник. На бэке правильно (в админке видно darkstar_2026 → Darkstar). Починить инвалидацию кэша профиля (GET /profile заново). Это наш NICK-1 из 06.11.
Итого P2 (бэкенд): 16.

🟨 P2 — Фронтенд: карточки, мобилки, auth UX
#	Задача
FE-CARD-1	Стоимость (прочерк) — добавить cost: market_rub в attributes
FE-CARD-2	Ход (прочерк) — добавить поле move в форму или автозаполнение
FE-CARD-3	Бейдж «Товар» вместо «Карточка» в магазине — починить shopItemMap.ts
FE-CARD-4	Класс карты (прочерк) — добавить поле class в форму (массив строк)
FE-MOB-1	Убрать горизонтальный скролл (overflow-x-hidden)
FE-MOB-2	Отступы по бокам (лонгрид)
FE-MOB-3	Центрирование модалок и колокольчика
FE-MOB-4	Свайпы для 2048 и Pancaker
FE-AUTH-1	Автозаполнение email/пароля после регистрации при переходе на вкладку входа
FE-AUTH-2	Глазик (показать/скрыть пароль) на форме регистрации/входа
FE-AUTH-3	Смена пароля — отдельный сценарий auth (старый пароль, bcrypt, сброс сессий)
Итого FE P2: 11.

🟧 P1 — Новые сценарии: модерация, уведомления, подписки
Это твой сценарий 1–3 из 06.11. Я убрал MOD-6 (аппрув), потому что ты сам решил: постмодерация, не аппрув.

#	Задача	Домен
MOD-1	Уведомление автору при удалении товара админом — «ваш товар снят с продажи / отозван администратором» + «подробнее» → окно «что отозвано и причину»	Бэк + Фронт
MOD-2	Админ: при редактировании/удалении товара спрашивать причину — записывать в окно, товар soft-delete	Фронт + Бэк
MOD-3	Автор видит причину удаления и может заново создать товар (фото + описание)	Фронт + Бэк
MOD-4	Уведомление автору, если админ поменял описание или цену товара	Бэк + Фронт
MOD-5	Админ: кнопка «Дать подписку» / «Отозвать» рядом со статусом подписки (активна–неактивна)	Фронт + Бэк
Итого модерация: 5. (MOD-6 аппрув — убран, постмодерация.)

🟧 P1 — История (History service)
Это твой «в записях ничего не отображается». Плюс окно 10 записей.

#	Задача	Домен
HIST-1	Починить историю на фронте — ничего не отображается. Проверить, настроен ли бэк на сохранение и выдачу	Фронт + Бэк
HIST-2	История — 10 последних записей на пользователя (действия на аккаунте), каждому отдельно	Бэк + Фронт
HIST-3	Общая статистика — показывать в окне, где пользователь открывает, актуальную под вкладку	Бэк + Фронт
HIST-4	У админа история в JSON (регистрация, рекорды, покупки, оплата, авторы) — служебная. Проверить, у авторов и пользователей	Бэк + Фронт
HIST-5	Проверить, почему история сломана — возможно, потому что служебная и скрыта	Бэк
Итого история: 5.

🟧 P1 — Кейсы «бабуля натыкала 30 пицц» + SAGA
Это твоя отдельная мысль. Разбираем как задачи-исследования, не как код «сегодня».

#	Кейс	Что проверить
CASE-1	30 пицц в одном заказе	Есть ли алерт/звонок менеджера при >10 в заказе?
CASE-2	Бабуля словила лаг, нажала 30 раз «оформить»	Что с сервисами? Легли? Идемпотентность?
CASE-3	Два игрока купили последнюю легендарную карту	Saga compensating → compensated работает? Написать тест: два игрока → последняя карта → один откатан → «товар закончился».
Итого кейсы: 3.

🟨 P2 — Админка: статистика инвентаря
#	Задача
ADMIN-INV-1	Переименовать: «Всего товаров» → «Уникальных позиций»; «Сток» → «Общее кол-во товаров в наличии»
ADMIN-INV-2	Починить «По типу» — пропуск: 4 карточки (новые, под авторами) с пустым типом (блоб с пустым местом). Проверить маппинг type для новых карточек
ADMIN-INV-3	Починить «По автору» — засеяно одним юзером 108 авторов → «засевный автор 1» + остальные. Дедупликация/разделение. Уточнено: 108 — статика, реальных 11.
Итого админка: 3.

🟦 P3 — Roadmap (не трогаем сейчас)
#	Что
RM-1	i18n (RU/EN/DE/ES) — i18next + react-i18next
RM-2	Пагинация каталога на сервере — до 1000+ карт
RM-3	Уровни лидерборда 2–3 — сейчас все level=1
RM-4	Порог покрытия в CI ≥70%
RM-5	Retry с джиттером, bulkhead
RM-6	108 авторов — витрина без текстов/картинок. Locked until product OK
RM-7	3D — нет ассетов и клиента. Locked
RM-8	Helm как повседневный деплой — чарт неполный. Locked
🟩 P4 — Долгосрок / парковка (не трогаем)
Loki / Tempo / node-exporter / SLO burn, Consul, Alertmanager → Telegram, WAF / пентест, C4 payouts, Miro-стикеры, Единый хром мини-игр.

✅ Уже закрыто (не тащим)
v1.1.3.1 патч (AppNavbar, Modal, InventoryEditModal isCard/card_text/цена/version/двойной updateItem, validateItem) · v1.1.3 · v1.1.2 · v1.1.1 · v1.1.0 · v1.0.9 · v1.0.8 · v1.0.7 · v1.0.6 · v1.0.5 · GEARS_ISSUE · FRONTEND_AUTHORS_FIX блоки 1–4.

📊 Сводка по числам
Приоритет	Кол-во
P0 (блокеры)	11
P1 (надёжность)	23
P2 (бэкенд)	16
P2 (фронт)	11
P1 (модерация)	5
P1 (история)	5
P1 (кейсы)	3
P2 (админка)	3
P3 (roadmap)	8
P4 (парковка)	6
Всего активных	~94
Критичных (P0)	11

🎯 Что делать прямо сейчас — следующая сессия
Не распыляйся. Возьми одну из этих трёх, они самые «дешёвые» и разблокируют остальное:

P0-11 — AdminCancelPurchase обходит fulfilled_at для админа. Ты уже проверил вживую — 400. Маленький дифф в RefundPurchase.

P2-16 / NICK-1 — ник в профиле. На бэке правильно, на фронте — кэш. Это самая заметная UX-бага и самая дешёвая.

Мой совет: 1 → 2 → 3. За одну сессию закроешь три «горячих» и почувствуешь, что каша разгребается.