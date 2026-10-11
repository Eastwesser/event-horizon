# 11.10.2026 — что добили и что сознательно оставили

Рабочая заметка к v1.1.1. Тег стоит на коммите, в который входит эта волна.

## Закрыто в коде

### Рекорд игры больше не публикуется дважды

`services/game/internal/worker/outbox_worker.go` забирает необработанные строки так:

`SELECT … FROM outbox WHERE processed = false ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED`

Два процесса берут разные строки. Строка помечается обработанной только после успешного `Publish` в JetStream. Если JetStream не подключён, строка остаётся необработанной и уйдёт со следующей попытки.

### Сокет лидерборда без JWT в URL

Браузер делает `POST /api/v1/ws-ticket` с заголовком `Authorization: Bearer`. Gateway кладёт в Redis ключ `ws:ticket:{id}` на 30 секунд и отдаёт `{ticket, expires_in}`. Сокет открывается как `GET /ws/leaderboard?ticket=`. Билет удаляется первым `GETDEL`, второй раз он уже недействителен.

`?token=` с JWT gateway отвечает 401. Небраузерный клиент по-прежнему может передать `Authorization: Bearer` на сам upgrade. Origin проверяется как раньше.

Access-лог gateway пишет метод, путь и статус. Query в лог не попадает, поэтому билет не оседает в строке прокси из-за нашего логгера.

Фронт: `frontend/src/components/Leaderboard/Leaderboard.tsx`.

### Оркестратор покупки

Списание билетиков живёт в Postgres billing, склад — в Postgres shop. Общей транзакции на две базы нет и не будет: двухфазный коммит здесь ломает доступность ради иллюзии одной базы.

Оркестратор — строка `purchase_sagas` в базе shop (`services/shop/migrations/20261011011000_purchase_sagas.sql`). Ключ один на попытку списания: `spend_ref`. Повтор того же ключа возвращает ту же сагу и тот же `purchase_id`.

Состояния:

- `started` — строка создана, билетики ещё не списаны
- `aborted` — billing отказал (нет баланса или другая ошибка, это не повтор ключа)
- `spent` — билетики списаны или ключ уже был использован
- `stocked` — строка `purchases` записана
- `compensating` — склад не записался, идёт возврат
- `compensated` — возврат прошёл, cycle сдвинут, следующая покупка получит новый ключ

Если процесс умер в `spent` или `compensating` и строка старше 30 секунд, воркер shop (`resumeSagas`) берёт её `FOR UPDATE SKIP LOCKED`. Нашёл `purchases.id` — ставит `stocked` и ничего не возвращает. Не нашёл — вызывает тот же возврат, что и outbox `billing.refund`, сдвигает `purchase_claims.cycle` и ставит `compensated`. Лок строки снимается до RPC в billing.

### Доля 85/15 записана, перевод автору не делается

`services/payment/migrations/20261011010000_revenue_splits.sql`. Таблица применяется самим payment при старте (`migrator.Up` на встроенных SQL).

`CompletePaymentAndActivateSubscription` в одной транзакции: блокирует платёж, закрывает его, гасит старую подписку, вставляет новую, пишет outbox `payment.completed` и строку `revenue_splits`. Авторская доля — `amount * 85 / 100`, платформа — остаток, чтобы рубль не терялся на округлении: 200 → 170/30, 300 → 255/45.

Это журнал. Денег на карту автора код не отправляет: живая касса по-прежнему Boosty, своего эквайринга с автосплитом нет.

## Как у нас работают транзакции

Одна база — обычная транзакция Postgres. Несколько баз — сага и идемпотентный ключ. Распределённого координатора нет.

### Внутри одной базы

Billing `SpendBalance`: условный `UPDATE … WHERE balance >= $1`, строка в `transactions`, событие в `outbox`, один commit. Повтор того же `reference_id` упирается в уникальный ключ и не списывает второй раз.

Shop `PurchaseItemWithStock`: `SELECT stock FOR UPDATE`, склад, инвентарь игрока, `purchases`, несколько строк outbox, один commit. Отмена: блокировка покупки, окно 7 дней, запрет физического товара после `fulfilled_at`, статус `REFUNDED`, возврат штуки на склад.

Payment: завершение платежа, подписка, outbox и `revenue_splits` — один commit. Повторный вебхук видит статус `completed` и не открывает подписку второй раз.

Game: рекорд и строка outbox `score.updated` пишутся вместе. Публикация — отдельным воркером после commit.

### Между базами

Покупка скина: shop сначала пишет сагу `started`, потом зовёт billing. Успех или «такой ключ уже был» → `spent`. Потом shop пишет склад и `purchase.paid` в своей транзакции → `stocked`. Сбой склада → `compensating`, возврат с ключом `refund:shop:{user}:{item}:{cycle}`, затем `compensated`.

Воркер — это и есть продолжение оркестратора после падения процесса. Он не держит две базы в одной транзакции. Он дочитывает строку саги и делает один идемпотентный шаг.

События наружу идут из outbox той базы, которая их породила. Shop и game помечают строку обработанной только после публикации и только под `SKIP LOCKED`.

## Observability

Дашборд `deployments/grafana/dashboards/event-horizon.json`, uid `event-horizon`, version 4. Папки и Prometheus datasource уже провиженятся compose.

К панелям 1–16 добавлены:

- p95 задержки gateway по пути
- ошибки gRPC в секунду по сервису
- p95 задержки gRPC по сервису
- максимум паузы GC (`go_gc_duration_seconds`, quantile 1) у горячих сервисов
- CPU процесса
- сумма билетиков, списанных магазином (`orders_revenue_total` — это цена в билетиках, не рубли)

Loki и Tempo в повседневный `make deploy` не ставились. На хосте около 8 ГиБ, Jaeger уже принимает OTLP на `:4317`, Prometheus уже скрейпит сервисы. Второй сборщик логов и второй сборщик трейсов на этой машине конкурируют с Postgres и k6, а графики выше живут на метриках, которые уже есть.

## Защита

Сессия игрока — заголовок `Authorization: Bearer`, не cookie. Чужой сайт не может приложить этот заголовок к нашему API из браузера игрока: это и есть защита от CSRF для такой схемы. Отдельный CSRF-cookie не заводился, он понадобился бы, если бы сессия жила в cookie.

Что стоит на gateway:

- `SecurityHeaders`: `nosniff`, запрет фрейма, `Referrer-Policy: no-referrer`, отключение геолокации, микрофона и камеры
- `ReadHeaderTimeout` 5 секунд — обрыв медленного заголовка (slowloris). `ReadTimeout`/`WriteTimeout` на весь ответ не ставились: они рвут долгое WebSocket-соединение
- логин: 5 запросов в секунду с IP; ошибка Redis на этом лимите не пускает (функция возвращает false)
- общий потолок около 100 запросов в секунду на пользователя или IP; если Redis на общем лимите недоступен, API остаётся открытым, чтобы лимитер не погасил весь вход
- сокет: 100 новых соединений в минуту с IP
- SQL пишется плейсхолдерами `$1`, `$2`. Фильтры инвентаря подставляют номер плейсхолдера, не текст значения
- пароль — bcrypt, стоимость 12. Роли `user` | `author` | `admin` проверяются на ручках записи

Дальше по защите, 11 октября:

- `GET /docs`, `GET /openapi.yaml` и файлы Swagger отвечают 404, пока не задан `DOCS_PASSWORD`. С паролем браузер спрашивает Basic (`DOCS_USER`, по умолчанию `docs`). Пароль в URL не кладётся. На публичном `https://easwessers-event-horizon.com/docs` без этой переменной документации нет.
- Swagger UI 5.17.14 лежит в репозитории (`services/gateway/api/swagger/`). Страница больше не ходит на unpkg.
- CSP на ответах gateway: у API `default-src 'none'`, у `/docs` скрипты только со своего хоста. HSTS ставится, когда запрос пришёл по HTTPS или с `X-Forwarded-Proto: https`, чтобы локальный http не залипал.
- Тело обычного JSON не больше 1 МиБ. Загрузка картинок по-прежнему до `UPLOAD_MAX_BYTES`.
- Путь с `..` или обратным слэшем отклоняется. Методы вне GET/HEAD/POST/PUT/PATCH/DELETE/OPTIONS тоже.
- В `frontend/index.html` стоит CSP страницы: свои скрипты и стили, картинки ещё и по https (обложки авторов), сокеты `ws:`/`wss:`. `unsafe-inline` оставлен, потому что Vite в разработке вставляет скрипт обновления прямо в страницу. Чужой домен скрипт подключить не может.

WAF (ModSecurity, Cloudflare) и пентест сюда не входят. WAF — это отдельный фильтр перед приложением, его подпись надо крутить на живом трафике, иначе режутся нормальные ники и описания. Пентест — это человек, срок и отчёт, не флаг в коде. Ротации JWT-секрета по-прежнему нет.

## Swagger

Спека есть. Канонический текст — `docs/openapi.yaml`. Gateway вшивает копию `services/gateway/api/openapi.yaml` и отдаёт её как `GET /openapi.yaml`. Страница `GET /docs` — Swagger UI 5 со своими файлами на том же хосте. Обе ручки закрыты `DOCS_PASSWORD`.

Публичный HTTP только у gateway. Остальные сервисы говорят gRPC. В спеке нет отдельного REST на каждый сервис, и это совпадает с кодом.

Ручки в спеке: health, admin users, auth (register, login, refresh, whoami, logout, role, user, nickname), billing balance / balance-all / transactions, game submit / daily-gift / boost, leaderboard, profile, shop items / purchase / cancel / inventory, uploads, inventory (list, get, create, update, delete, bulk, reserve, soft, restore, stats), `POST /api/v1/ws-ticket`, `GET /ws/leaderboard`, payment checkout / subscription / can-purchase-merch / webhook / yookassa webhook, authors me / sales / apply / application / approve / reject / list / get, history, notifications + read, analytics dau / mau / retention.

## CRUD — где он есть и где он не нужен

Писали по действиям игрока и автора, не по таблице «на каждую сущность четыре глагола». Пустых ручек ради полного набора нет.

| Место | Что есть | Чего нет и почему |
|---|---|---|
| Inventory | list, get, create, update, delete, bulk, reserve, soft-delete, restore, stats | Это каталог автора, здесь CRUD уместен |
| Auth | регистрация, вход, refresh, whoami, logout, ник, смена роли админом, чтение пользователя, список пользователей админу | Игрок не редактирует чужой аккаунт. Удаления аккаунта нет |
| Shop | чтение витрины, покупка, отмена, инвентарь игрока | Карточку создаёт Inventory и событие, не `POST /shop/items` |
| Game | submit, boost, дневной подарок | Игра — клиентский рантайм, не строка, которую создают через API |
| Leaderboard | чтение топа | Пишет событие `score.updated` |
| Billing | оба баланса, один баланс, лента проводок | `GET /api/v1/billing/balance?currency=` и `GET /api/v1/billing/transactions` читают только id из токена. Смена пароля отдельным методом не заводилась: это новый сценарий auth (старый пароль, bcrypt, сброс сессий), не готовый RPC |
| Profile | чтение собранной анкеты | Имя меняется через `POST /api/v1/auth/update-nickname`. gRPC `UpdateProfile` наружу не открыт: он умеет переписать `total_score` |
| Payment | checkout, подписка, проверка мерча, вебхуки | Платёж создаёт checkout, закрывает вебхук. Отдельного customer CRUD нет |
| Authors | заявка, моя карточка, продажи, approve/reject, публичное чтение | Удаления автора нет |
| History | список | Журнал дописывается событиями |
| Notifications | список и «прочитано» | Удаления уведомления нет |
| Analytics | dau, mau, retention, только admin | Три чтения, не CRUD |
| Fulfillment, nats-hub, MCP | нет игрового HTTP | Внутренние процессы |

## README

В шапке уже было «8 мини-игр». Таблица `## 🎮 Игры` перечисляла пять. Сейчас в ней все восемь, в том порядке, что и `GAME_LB_TABS`: Flappy Bird, Pancaker, Builder, Memonia, Hanoi, 2048, Gears, Tamagotchi.

## Что не добили и почему это не дыра в коде

**108 авторов.** Витрина авторов без их текстов, картинок и согласия — набор пустых карточек. Сид не добавлялся.

**3D.** Нет ассетов и игрового клиента. Заглушка на сцене ничего не доказывает.

**Helm как повседневный деплой.** Чарт `deployments/helm/event-horizon` остаётся путём `make deploy-k3s`. В нём нет Redis, payment, authors, history, notification и их баз. `make deploy` по-прежнему compose. Переключение сломало бы локальный подъём.

**Loki и Tempo.** См. раздел Observability. Следы уже уходят в Jaeger, метрики — в Prometheus. Отдельный профиль на тонкую машину не включался в ежедневный подъём.

## SQL

Запросы в репозиториях остаются текстом с `$1`, `$2`. Squirrel сюда не тащили: `pkg/sqb` уже делает то же самое в Auth, значения всё равно лежат в аргументах, не в строке. Угроза была бы, если бы текст игрока попадал в SQL. В Inventory имена колонок зашиты в `toPostgresFilter`, а значения идут плейсхолдерами. В MCP `postgres_query` дополнительно запрещает `UNION` / `EXCEPT` / `INTERSECT`.

## MCP

Гайд: `services/mcp/README.md`. Это stdio-процесс для Cursor, не ручка сайта. `MCP_WHEN.md` помечен как устаревший.

## Порты

Таблица в README сверена с `docker-compose.cluster.yml` по портам хоста. Metrics gateway на хост не проброшены.

## Комментарии

Английские комментарии стоят на функциях, по имени которых не видно шага: сага покупки, билет сокета, окно отмены, мерч-гейт, фильтр инвентаря. `Save` / `Get` без комментария оставлены.
