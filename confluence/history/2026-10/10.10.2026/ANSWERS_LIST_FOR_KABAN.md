# Как устроен Event Horizon

Сверено с кодом после правок v1.1.1 (10.10.2026). Что именно поменялось относительно первого разбора — внизу, в разделе «До и после».

Игрок бьёт в балансировщик `:8079`. Дальше HTTP принимает gateway и синхронно зовёт домен по gRPC. Деньги, склад и рекорд пишутся в свой Postgres сервиса. То, что можно пережить с задержкой — очки на табло, колокольчик, «сборка» заказа, история, аналитика — уходит в NATS JetStream.

## Авторизация

Регистрация разрешает только роль `user`. Роль `author` или `admin` с публичной формы не ставится. Пароль уходит в bcrypt с cost 12, в базе лежит хеш.

Логин сверяет хеш и выпускает пару. Access по умолчанию живёт 15 минут, refresh — 7 дней. Оба подписаны HMAC-SHA256 общим секретом из env. В claims лежат `user_id`, `email`, `role`, тип токена (`access` или `refresh`) и `jti` — уникальный id этого конкретного токена.

`jti` access-токена пишется в Redis на срок жизни access. `jti` refresh — отдельно, на срок refresh. Пока ключ есть, сессия жива.

Каждый закрытый запрос gateway не разбирает подпись сам. `RequireAuth` вырезает `Authorization: Bearer …` и зовёт Auth по gRPC `ValidateToken`. Auth проверяет подпись, тип `access` и наличие `jti` в Redis. Нет ключа — ответ «сессия отозвана», gateway отдаёт 401. Роль кладётся в контекст запроса. `RequireRole` после этого пускает только нужную роль на конкретные ручки.

Refresh одноразовый: старый refresh-`jti` удаляется, пара выпускается заново, роль перечитывается из Postgres. Поэтому смена роли подхватывается без вечного старого access.

Logout удаляет `jti` из Redis. Подпись у токена остаётся верной до `exp`, но проверка сессии его больше не пускает.

Если Redis на проверке ответил ошибкой, Auth возвращает gRPC `Unavailable`, gateway отвечает **503**. Подпись без живого Redis больше не считается сессией. То же на refresh.

Логин с одного IP режется 5 запросами в секунду. При ошибке Redis этот лимит закрывается (пускать перестаёт). Общий лимит ~100 запросов в секунду на пользователя при ошибке Redis наоборот открывается.

## Идемпотентность

Где она есть:

- Биллинг. `transactions.reference_id` уникален. Повтор `AddCurrency` / `SpendCurrency` с тем же ключом не проходит второй вставкой.
- Покупка. Ключ списания `shop:{user}:{item}:{cycle}`. Два клика делят один cycle и один ключ. После возврата cycle растёт, следующая покупка того же товара получает новый ключ.
- Отмена. Ключ `cancel-{id покупки}`. Повторный cancel снова зовёт `AddCurrency`. Вторая проводка упирается в уникальный ключ и деньги не удваивает.
- Срыв записи покупки после списания. В outbox ложится `billing.refund` с ключом `refund:shop:{user}:{item}:{cycle}`. Запрос сразу пробует вернуть деньги. Если процесс умер, воркер дожимает тот же ключ.
- Инвентарь магазина: первичный ключ `(user_id, item_id)`. Второй раз тот же скин тому же игроку не вставить.
- Уведомление: `ON CONFLICT (user_id, source_event) DO NOTHING`.
- Рекорд: `ON CONFLICT` обновляет лучший счёт.
- Вебхук оплаты: повторный confirm с уже известным платежом gateway отвечает 200.
- Буст: если активный буст уже есть, лампочки второй раз не списываются.

## NATS JetStream и Kafka

JetStream — лог поверх NATS. Сообщение лежит в стриме на диске, пока consumer его не подтвердит (`Ack`).

Кластер из трёх нод (`nats-1/2/3`). Стрим один, имя `EVENTS`, хранение файловое, возраст 7 дней, потолок 1 ГБ. Его создаёт `nats-hub`. Субъекты: `score.updated`, `user.registered`, `shop.purchased`, `shop.purchase.failed`, `payment.completed`, `purchase.paid`, `purchase.fulfilled`, `boost.started`, события авторов и инвентаря, плюс шаблон `event.>`.

| Кто публикует | Субъект | Кто читает |
|---|---|---|
| Game, через outbox | `score.updated` | Leaderboard, Billing, Profile, History, Analytics, Gateway → вебсокет |
| Shop outbox | `purchase.paid` | Fulfillment, Notification |
| Fulfillment | `purchase.fulfilled` | Shop (ставит `fulfilled_at` по `purchases.id`), Notification |
| Shop outbox | `shop.purchased` | History, Analytics |
| Inventory | `inventory.item.created` | Shop создаёт строку каталога |

Партиций у NATS нет. Параллелизм задаёт durable-имя подписки (`gateway-websocket`, `shop-inventory-sync`, `analytics-score-updated`). Одно durable-имя делит поток. Разные durable на один субъект получают каждое сообщение сами — как несколько consumer group Kafka на один топик.

Kafka в обычном `make deploy` выключена: producer без `KAFKA_BROKERS` ничего не шлёт. Топики, если брокер включили профилем `make deploy-heavy`, называются так же: `purchase.paid`, `purchase.fulfilled`, `boost.started`. Ключ сообщения — id покупки. Брокер в compose один, replication factor 1, без шифрования. Это учебный контур.

Если JetStream жив, fulfillment и notification Kafka не читают. Один заказ не собирается дважды.

## Сага покупки и отмена

Две локальные транзакции и компенсация, которую можно дожать.

Покупка:

1. Shop берёт cycle из `purchase_claims` и синхронно зовёт Billing: `UPDATE … WHERE balance >= amount`, вставка в `transactions` с ключом `shop:{user}:{item}:{cycle}`, outbox `balance.updated`, commit.
2. Транзакция Shop: `SELECT stock FOR UPDATE`, `stock - 1` с проверкой `version`, вставка в `inventory`, строка `purchases` с заранее известным id, статус `COMPLETED`, `refundable_until = now + 7 days`, две строки outbox — `shop.purchased` и `purchase.paid` с этим же id. Commit.
3. Воркер outbox забирает строки `FOR UPDATE SKIP LOCKED` и публикует их в JetStream. Прямого `Publish` после commit больше нет.

Если шаг 2 упал не из-за «уже куплено», Shop пишет `billing.refund` и зовёт `AddCurrency` с тем же ключом возврата. Повтор этого ключа деньги не удваивает. Если и этот вызов упал, счётчик `shop_purchase_refund_failures_total` растёт, строка outbox остаётся, воркер повторяет.

Отмена внутри Shop — одна транзакция: блокировка покупки, отказ если окно 7 дней вышло или физический товар уже `fulfilled_at`, статус `REFUNDED`, удаление из `inventory`, `stock + 1`. Потом `AddCurrency` с `cancel-{id}`. Повтор отмены делает тот же вызов.

Мерч перед покупкой синхронно спрашивает Payment: `CanPurchaseMerch`. Нет подписки — покупка не начинается.

## Безопасность

CSRF. Сессия живёт в заголовке `Authorization`, не в cookie. Чужой сайт не подставит этот заголовок. Отдельного CSRF-токена нет: при этой схеме он не нужен. Если JWT когда-нибудь ляжет в cookie, понадобятся SameSite и токен.

SQL. Запросы в репозиториях идут с `$1, $2`. Фильтры инвентаря подставляют номер плейсхолдера, значение уходит аргументом.

XSS. Во фронте нет `dangerouslySetInnerHTML`. React экранирует текст.

Роли. Запись каталога автора и админские ручки закрыты `RequireRole`.

Чего нет: WAF, Content-Security-Policy, пентест, ротация JWT-секрета. Swagger UI тянется с `unpkg.com`. Rate limit на логин есть. Секреты из env.

## Вебсокеты

Один эндпоинт: `GET /ws/leaderboard?token=`. До upgrade gateway проверяет Origin (тот же хост, localhost или список `WS_ALLOWED_ORIGINS`) и `ValidateToken`. Без токена соединения нет.

Game кладёт `score.updated` в outbox. Воркер публикует в JetStream. Gateway подписан durable-именем `gateway-websocket` и рассылает JSON подключённым. Клиент в сокет ничего не пишет: read-цикл ловит обрыв.

Лимит: 100 новых соединений в минуту с одного IP. Табло по-прежнему читается обычным HTTP.

## Fulfillment

Имитация сборки, не склад. Сервис читает `purchase.paid`, ждёт 10 секунд, публикует `purchase.fulfilled` с тем же id покупки. Shop ставит `fulfilled_at`. Notification пишет в колокольчик. Пока `fulfilled_at` пустой и не вышли 7 дней, отмена возможна. Для физического товара после сборки отмена закрыта.

## Observability

Grafana поднимается со стеком и сама забирает дашборд **Event Horizon — Ops Overview** из `deployments/grafana/dashboards/event-horizon.json`. Руками его собирать не нужно.

На нём: живые таргеты, RPS gateway, доля 5xx, p95/p99, сабмиты игр, заказы, NATS, горутины, Redis, Postgres auth-экспортёра, heap, плюс ряд v1.1.1: скорость `purchase.paid` и `purchase.fulfilled`, p95 сборки, счётчик сорвавшихся мгновенных возвратов.

Prometheus скрейпит `/metrics`. Jaeger получает трейсы gateway по OTLP `:4317`. Loki и Tempo в этот дашборд не входят.

## Синхронно и асинхронно

Синхронно, игрок ждёт ответ:

- Gateway → Auth `ValidateToken` на каждом закрытом запросе.
- Gateway → Game `SubmitScore`, `StartBoost`.
- Gateway → Shop покупка, отмена, каталог.
- Shop → Billing списание и возврат.
- Shop → Payment `CanPurchaseMerch` для мерча.
- Game → Billing списание лампочек на буст.
- Gateway → Leaderboard / Profile / Inventory / Authors по их ручкам.

Асинхронно, ответ игроку уже ушёл:

- `score.updated` → табло, награда, достижения, история, аналитика, сокет.
- `shop.purchased` → история и ClickHouse.
- `purchase.paid` → сборка и колокольчик.
- `purchase.fulfilled` → отметка в Shop и второе уведомление.
- `inventory.item.created` → строка в каталоге Shop.

Общий человек — `user_id` из Auth. Общие деньги — Billing: игра начисляет билетики и лампочки по `score.updated`, буст списывает лампочки синхронно, магазин списывает билетики синхронно. Скин лежит в shop-inventory. Каталог автора живёт в Inventory и событием появляется в Shop. Прямого gRPC Game → Shop нет.

## MCP

Процесс `services/mcp` говорит с Cursor по stdin/stdout. В браузер игрока он не смотрит и в compose игрового стека не входит.

Шесть инструментов: список стримов JetStream, список consumer'ов, `SELECT` в Postgres (запись запрещена, лимит строк), `GET` и `SCAN` Redis, поиск по базе знаний Prydwen (TF-IDF по markdown, без Qdrant).

Это инструмент разработки. В оценку продукта для инвестора его класть не стоит. В сеть его выпускать нельзя: DSN, который ему дали, он читает целиком.

## Сервисы

- **Balancer** — вход `:8079`, несколько реплик gateway за ним.
- **Gateway** — единственный HTTP. JWT через Auth, лимиты, circuit breaker на gRPC, OpenAPI на `/docs`, сокет, метрики.
- **Auth** — пользователи, пароли, пара токенов, сессии в Redis, роли.
- **Game** — приём счёта, буст за лампочки, рекорд и outbox `score.updated` в одной транзакции. Бустнутый забег в таблицу не идёт.
- **Billing** — билетики и лампочки. Списание условием `balance >= amount`. Проводка и outbox в одной транзакции.
- **Leaderboard** — Redis sorted set и Postgres. Слушает `score.updated`. Доски читают `level=1`, кроме Flappy.
- **Shop** — витрина, покупка, инвентарь игрока, отмена 7 дней.
- **Inventory** — каталог автора, Postgres и Mongo. Пишет события, Shop их подхватывает.
- **Profile** — анкета и достижения по `score.updated`.
- **Payment** — подписка и шлагбаум мерча. Повтор вебхука идемпотентен.
- **Authors** — карточки авторов. Выплаты авторам (C4) закрыты.
- **History** — лента событий из NATS.
- **Analytics** — те же события в ClickHouse.
- **Fulfillment** — таймер 10 секунд и `purchase.fulfilled`.
- **Notification** — колокольчик в приложении. Повтор того же события не плодит строку.
- **Nats-hub** — создаёт стрим `EVENTS`. Игровой логики нет.

## Технологии

Go — все сервисы. gRPC — вызов между ними. Gin — только gateway. Postgres на сервис — чужая схема не общая. Redis — сессии, лимиты, кеш витрины, топ. Mongo — атрибуты карточек. NATS JetStream — лог событий с подтверждением. Kafka — второй брокер, в обычном деплое не кормит покупку. ClickHouse — аналитика. JWT — кто ты. bcrypt — проверка пароля. Outbox — событие в той же транзакции, что и данные. Circuit breaker — gateway перестаёт долбить мёртвый gRPC. OpenTelemetry / Jaeger — след запроса. Prometheus / Grafana — цифры и готовый дашборд. k3s / Ansible — задел выкладки. Повседневный запуск — Docker Compose. Первый `make deploy` после перезагрузки ждёт, пока Postgres ответит, и только потом гоняет goose.

## Что спросят люди с деньгами

Продукт — игровая витрина: аккаунты, мини-игры, таблица, магазин скинов и карточек, внутренняя валюта. Выручки авторов нет, процент с продажи не включён.

Масштаб, под который считали железо, — около 10 тысяч людей в день, десятки запросов в секунду. Локальный прогон это подтверждает на 20 виртуальных пользователях. Десять тысяч одновременных соединений никто не обещал и не мерил.

Деньги внутри системы — билетики и лампочки. Реальный эквайринг — вебхук Payment.

Покупка больше не теряет id заказа и не списывает дважды с двух кликов. Возврат можно повторить тем же ключом. Это ещё не банковская сага с оркестратором: списание и склад по-прежнему в двух базах, их связывает ключ и воркер.

## До и после v1.1.1

| | До | После |
|---|---|---|
| Id в `purchase.paid` | случайные 16 байт, `fulfilled_at` не ставился | `purchases.id` |
| Повторный клик «купить» | два списания | один ключ на cycle |
| Срыв записи покупки | возврат только если процесс жив | outbox `billing.refund` + воркер |
| Повтор отмены | деньги не досылались | тот же `cancel-{id}` |
| Публикация `purchase.paid` | прямой publish после commit | outbox в транзакции, `SKIP LOCKED` |
| Вебсокет | открыт с любого сайта | Origin + токен |
| NATS и Kafka вместе | заказ собирался дважды | Kafka молчит, пока жив JetStream |
| Redis недоступен | токен считался живым | 503 |
| Grafana | без пути покупки | paid, fulfilled, p95 сборки, срывы возврата |

Подробности по файлам: `FIX_AFTER_ESSAY.md` рядом.
