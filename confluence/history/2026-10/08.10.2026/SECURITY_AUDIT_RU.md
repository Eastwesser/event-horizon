# Аудит защиты и нагрузки Event Horizon (v1.1.0)

Для собеседований: цифры + что реально измерили. Не путать **целевую модель 10k DAU** с **локальным CORE k6**.

Дата прохода CSRF/XSS/SQL/AuthZ: **09.10.2026**.

---

## 1. Целевая модель (из `LOAD_TESTS/METRICS.md` + Miro правая колонка) — «как должно быть на проде»

| Метрика | Целевое значение | Комментарий |
|---------|------------------|-------------|
| MAU / DAU | 100k / **10k** | DAU ≈ 10% MAU |
| Сессии / юзер / день | 10–20 | в расчётах часто берут 15 |
| Средний RPS (грубо) | ~17 (пик ~35) | 10k×15×10 API / 86400; пик ×2 |
| Альтернативная оценка | 30–50 avg, пик ~100 | если сессия «тяжелее» (на Miro тоже) |
| Latency p50 / p90 / p95 / p99 | &lt;50 / &lt;150 / &lt;200 / &lt;300 ms | цель на проде |
| HTTP / WS одновременные | ~5k / ~2.5k (до 10k / 5k) | модель, не измерение |
| Нагрузка железа | CPU/RAM/IO/net &lt;70% на пике | |
| Бюджет Selectel (оценка) | ~$430 / мес (5 VM) | август/окт 2026 |
| Read:Write | ~2:1 … 3:1 | больше чтений |
| Сырые события | retention ~30 дней | history/analytics |

**Вывод для интервью:** «Мы проектируем под ~10k DAU → десятки RPS в среднем, пик порядка ×2; латентность целимся в сотни мс на p99». Это **capacity planning**, не результат одного локального k6.

Сверка с Miro: правая колонка **совпадает** с `METRICS.md` (см. `miro/MIRO_REVIEW.md`).

---

## 2. Что реально измерили локально (08.10.2026)

| Прогон | Результат |
|--------|-----------|
| CORE `make test-k6` (browse, ~20 VU, :8079) | checks **100%**, p95 ≈ **705 ms**, errors **0%**, ~35 HTTP req/s |
| CORE `make test-k6-purchase` (1 VU) | checks **100%**, p95 ≈ **535 ms**, errors **0%** |
| Legacy blast ~500 VU login | ~99% login fail — **ожидаемо** (см. лимиты ниже); shop/LB/submit часто green |

Локальный p95 ~700 ms **выше** целевых 200 ms — это ok для laptop + docker; на собеседовании разделяй: «цель прода vs факт стенда».

Файлы: `LOAD_RESULTS/browse-20261008.txt`, `purchase-20261008.txt`.

---

## 3. Защита (дословно, по коду)

### Аутентификация / сессии
- JWT + сессии в **Redis**
- Роли: `user` | `author` | `admin`
- Пароли: **bcrypt cost 12** (`services/auth`)
- Секреты только из env (`internal/config`), `.env` в gitignore

### Rate limit (Gateway → Redis)
Код: `services/gateway/internal/ratelimit/limiter.go`

| Ограничение | Лимит |
|-------------|--------|
| Login с одного IP | **5 req/s** |
| Game submit на user | **10 req/s** |
| Global на user/IP | **~100 req/s** |
| WebSocket с IP | **100 conn / мин** |

При падении Redis limiter **fail-open** (API не валим вместе с Redis) — осознанный trade-off availability vs strict deny.

### Валидация / границы
- gRPC: Recovery → Logger → **Validate**
- HTTP gateway → gRPC; канонический префикс **`/api/v1/`** (legacy `/api/*` переписывается в v1 один раз)
- FE: `baseURL: '/api/v1'` + относительные пути (фикс `/api/api/profile` сохранён)

### Данные / SQL
- Параметризованные запросы (`$1…` через pgx) в repository-слое
- Отдельный Postgres на сервис (blast radius)
- Shop cleanup: только **явные UUID**, Berserk `карточка` не трогаем массовыми WHERE

### Уведомления
- In-app bell; soft-fail при logout (без цикла 401)
- Не долбим игроков в Telegram по умолчанию

---

## 4. CSRF / XSS / SQL / AuthZ — чеклист (09.10.2026)

| Угроза | Как проверяли | Статус | Заметки |
|--------|---------------|--------|---------|
| **CSRF** | Auth = Bearer JWT в `Authorization`, не cookie-session SPA | **OK (низкий риск)** | Классический CSRF на state-changing cookie не применим. Admin mutations идут с тем же Bearer. Cookie CSRF token не внедряли — не нужен при текущей модели. |
| **XSS** | Поиск `dangerouslySetInnerHTML` в `frontend/src` | **OK** | Совпадений **0**. React text escaping по умолчанию. Не рендерим сырой HTML из API в DOM. |
| **SQL injection** | Выборочный обзор `**/repository/**/*.go` (auth/shop/inventory/authors/profile/…) | **OK** | Запросы с `$n` плейсхолдерами; нет `fmt.Sprintf` склейки user input в SQL в просмотренных местах. |
| **AuthZ** | Gateway: `RequireAuth` + `RequireRole(RoleAdmin\|RoleAuthor)` на write/admin | **OK (базовый)** | Admin users/apps/analytics, author inventory writes, uploads — под ролями. Публичны только явные read (часть authors list). Полный matrix «каждая ручка × роль» — follow-up при seed 108 authors. |
| **Secrets** | `.env` gitignore; config из env | **OK** | Не хардкодить JWT/DB в репо. |
| **DDoS / brute** | Login 5/s IP; global ~100/s; bcrypt 12 | **OK (осознанно)** | 500 VU login blast красный by design. |

**Вердикт pass:** для демо/курса / v1.1.0 tag — **baseline + checklist закрыты**.  
Не заявлено: WAF, CSP production policy, CSRF double-submit для будущих cookie flows, fuzzing, pen-test.

---

## 5. «Сколько одновременных пользователей выдержит?»

Честный ответ:

1. **Продуктовая цель:** модель 10k DAU ≈ десятки RPS avg → один thin-стек + нормальный хостинг реалистичен; пики планируем ×2 и rate limit на auth.
2. **Auth write path** (login/register) — **узкое место**: bcrypt 12 + 5 login/s на IP. 500 параллельных логинов специально «краснеют». Это защита от брутфорса/дудоса на `/login`, а не «сервис мёртв».
3. **Read path** (shop, inventory, LB, submit при живой сессии) на локальном CORE держится при 20 VU без ошибок.
4. **Одновременные WS:** цель в METRICS ~2.5k; локально не гоняли до этого потолка — на интервью: «лимит WS 100/мин с IP + бюджет соединений в модели».

Формулировка вслух:  
«Мы не обещаем 10k одновременных логинов. Мы обещаем модель 10k DAU при нормальном RPS и жёстком rate limit на auth; read-path проверен CORE k6.»

---

## 6. Что смотреть в Grafana ночью (из METRICS / Miro)

- p99 latency Auth / Gateway  
- HTTP 5xx rate  
- Free RAM  
- NATS JetStream **consumer lag** (лаг ≫ 1000 → проблема consumer’а)

Go: goroutines &lt;1000/процесс, heap разумный, mutex waits не секундами.  
DB: connections &lt;100/pool (у нас пулы ~25 open / 10 idle в правилах), slow queries, deadlocks = 0.

---

## 7. Вердикт

| Вопрос | Ответ |
|--------|--------|
| Baseline security для демо/курса? | **Да** (JWT Redis, bcrypt 12, validate, rate limit, secrets env) |
| CSRF/XSS/SQL checklist закрыт? | **Да** (09.10) — см. §4 |
| Production hardening закончен? | **Нет** — пентест/WAF/CSP/полный AuthZ matrix |
| 500 VU login fail = баг? | **Нет** — capacity + защита |
| CORE browse/purchase green? | **Да** (08.10) |
| API prefix | **`/api/v1/`** (+ legacy rewrite `/api/*`) |

Дальше: не использовать legacy blast как pass/fail; parked avatar / 108 authors / MCP / C4 — вне этого аудита.
