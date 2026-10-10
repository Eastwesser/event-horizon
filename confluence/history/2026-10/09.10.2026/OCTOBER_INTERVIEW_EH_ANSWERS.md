# Event Horizon — October interview answers

> **Denis / EH architect voice.** Answers are Event Horizon–specific prep for the Shortcut 50-question bank — **not** a copy of Shortcut эталоны.  
> Source TOC / pasted guide: [`OCTOBER_INTERVIEW.md`](./OCTOBER_INTERVIEW.md).  
> Grounding: Clean Architecture, gRPC mesh, NATS JetStream, Outbox, JWT+Redis, Gateway Gin `/api/v1`, Prometheus/Grafana/Jaeger, balancer, k6, postgres-per-service.

---

## Если спросят про проект

- **Event Horizon v1.1.0** — игровая микросервисная платформа: 8 мини-игр, билетики/лампы, магазин авторов, real-time лидерборд; целевая модель ~10k DAU.
- **Edge:** React → Balancer `:8079` (least-conn) → Gateway ×3 `:8081–8083` (Gin, JWT, HTTP→gRPC, OpenAPI `/docs`).
- **Mesh:** Auth `:50051`, Game `:50052`, Billing `:50053`, Leaderboard `:50054`, Shop `:50055`, Inventory `:50059`, Profile `:50060`, Payment `:50058`, Authors `:50061`, History `:50062`, Analytics `:50057` + Fulfillment / Notification / NATS Hub.
- **Data:** PostgreSQL per service (host ports ~5460+); Redis for sessions (Auth), cache (Billing/Shop), LB Sorted Set; inventory also keeps Mongo where course requires it.
- **Async:** NATS JetStream stream `EVENTS` — `score.updated`, `purchase.paid` / fulfillment, shop/history subjects; Kafka opt-in (`make deploy-heavy`).
- **Outbox** (DB commit + worker → NATS): Game, Billing, Shop, Inventory, Payment, Authors.
- **Observability:** Prometheus / Grafana / Jaeger; `/health` + `/ready` on metrics HTTP; load via k6 (`deployments/k6/`).
- **Security:** JWT sessions in Redis, roles `user|author|admin`, bcrypt cost 12, secrets via env (`internal/config`).

---

## РАЗДЕЛ 01 — Go Core (1–8)

### 01. Generics vs интерфейсы

**Вопрос:** Работал ли с дженериками? В чём отличие от интерфейсов?

**Ответ:**  
В EH дженерики почти не в доменном слое — контракты идут через proto/gRPC и маленькие интерфейсы репозиториев (`UserRepository`, `BalanceRepository`). Где generic уместен — утилиты вроде `Min[T cmp.Ordered]`, generic helpers для тестов/коллекций.

Отличие: **дженерик** фиксирует тип на компиляции (constraint / type set), без boxing в интерфейс и без type assertion. **Интерфейс** — runtime-полиморфизм: в handler/service я принимаю интерфейс репозитория, а в DI подставляю Postgres или mock. Это разные оси: compile-time параметризация vs duck typing на методах.

В Clean Architecture EH я сознательно держу **accept interfaces, return structs** на границах service↔repo; дженерики туда не тащу, чтобы не усложнять gRPC-маппинг.

⚠️ **Частая ошибка:** сказать «дженерики = сахар над интерфейсами» или что Go всегда мономорфизует как C++ templates.

---

### 02. Неявная реализация интерфейсов

**Вопрос:** Что означает неявная реализация интерфейсов в Go? Как реализовать интерфейс?

**Ответ:**  
Тип реализует интерфейс, если у него есть все нужные методы — без `implements`. В EH интерфейсы объявляю **на стороне потребителя**: в `internal/service` — `type InventoryRepo interface { Get…; Add… }`, а Postgres-реализация живёт в `internal/repository` и «просто подходит».

Следствия: легко мокать в unit-тестах; слой handler не знает про SQL. Важно: если методы на pointer receiver — интерфейсу удовлетворяет `*T`, не `T`. Поэтому репозитории и сервисы у меня почти всегда `*XxxRepo` / `*XxxService`.

⚠️ Путают с динамической типизацией: проверка всё равно на компиляции.

---

### 03. Value vs pointer receiver

**Вопрос:** В чём отличие метода для указателя от метода для значения структуры?

**Ответ:**  
Value receiver работает с копией — мутации не видны снаружи; pointer — с оригиналом. В EH сервисы/репозитории с состоянием (DB pool, Redis client, NATS js) — **всегда pointer receivers**. Value — только для маленьких immutable DTO / domain value objects без mutex.

Правило: если тип содержит `sync.Mutex` / пул / клиент — только `*T`, иначе копирование mutex = баг. Method set: интерфейсу с pointer-методом удовлетворяет только `*T` — на этом ловятся ошибки при регистрации gRPC handler’ов.

⚠️ Копируют struct с mutex value-receiver’ом «потому что маленький».

---

### 04. nil-интерфейс vs typed nil

**Вопрос:** Что будет выведено при сравнении nil-интерфейса с typed nil?

**Ответ:**  
Интерфейс = пара (type, data). `nil` только когда оба пусты. `var e *MyError = nil; return e` кладёт в `error` typed nil → `err != nil` == true. В EH это критично на границах repo→service→handler: объявляю `var err error` и возвращаю голый `nil`, а не `return (*RepoError)(nil)`.

gRPC Recovery interceptor ловит panic; для бизнес-ошибок — явный `error` + status codes, не panic. Typed nil как раз маскирует «ошибки нет» и ломает gateway mapping.

⚠️ Уверенно говорят `true`, не зная iface/eface.

---

### 05. map: бакеты и эвакуация

**Вопрос:** Расскажи про map. Почему к элементу map нельзя взять указатель? Что такое эвакуация?

**Ответ:**  
Классическая hmap: бакеты по ~8 слотов + tophash; при росте — инкрементальная **эвакуация** в новую таблицу. Поэтому `&m[k]` запрещён: элемент может переехать. Map — ссылочный тип (копия переменной делит данные), **не потокобезопасна**.

В EH in-process map почти не как shared cache между горутинами gateway: кэш балансов/сессий — **Redis**. Локальные map — request-scoped или под mutex (например, in-memory registry в balancer). Concurrent write → fatal от runtime.

⚠️ Объясняют chaining «список на каждый ключ», не overflow-бакеты; не связывают `&m[k]` с эвакуацией.

---

### 06. Слайсы: header, len/cap, append

**Вопрос:** Что такое слайс? Зачем capacity в make? Как работает append и почему O(1) амортизированно? Изменится ли оригинал при передаче в функцию?

**Ответ:**  
Header: ptr, len, cap. Передача копирует header, массив общий → `s[i]=` видно снаружи; `append` меняет локальный header — снаружи len не растёт, пока не вернули слайс. `make(..., cap)` режет реаллокации (геометрический рост → амортизированно O(1)).

В EH: батчи outbox poller’а, protobuf↔model slices, k6-несвязанные буферы в workers — заранее `make([]T, 0, batchSize)`, чтобы не дробить heap под GC. Осторожен с sub-slice sharing при reuse буферов для NATS payload.

⚠️ «Слайс передаётся по ссылке» → ждут, что append всегда виден снаружи.

---

### 07. Конкурентный доступ к map

**Вопрос:** Code review: глобальная map, 80% чтений / 20% записей. Проблемы и решение?

**Ответ:**  
Без синхронизации — data race / fatal. Чинить: структура `{ mu sync.RWMutex; m map[...] }` — RLock на read, Lock на write; get-or-create целиком под Lock. `defer Unlock`. При очень hot write — шарды по hash ключа.

В EH аналог read-heavy: конфиг/feature flags in-process, rate-limit counters на gateway/balancer. Для session/balance — не map, а Redis. `sync.Map` — только если ключи стабильны и профиль реально read-heavy; иначе RWMutex проще рассуждать на ревью.

⚠️ Защищают только write; getOrCreate через RLock→Lock с окном гонки.

---

### 08. Указатели: `*p = v` vs `p = &v`

**Вопрос:** `setAge` принимает `*Person` и внутри создаёт новый `Person`. Как это повлияет на оригинал?

**Ответ:**  
Указатель копируется. `*p = Person{…}` пишет в чужую память — видно снаружи. `p = &Person{…}` меняет локальную копию — оригинал нет. В EH при маппинге proto→model пишу в поля существующего struct или возвращаю новый указатель; не «переназначаю» аргумент.

⚠️ «Раз pointer — любое присваивание p видно снаружи».

---

## РАЗДЕЛ 02 — Конкурентность (9–20)

### 09. Горутина vs OS-поток

**Вопрос:** Что такое горутина? Чем отличается от потока ОС?

**Ответ:**  
Горутина — единица планирования Go runtime: стартовый стек ~2KB, растёт; тысячи штук дёшево. OS-thread — 1–8MB стек, переключение через ядро. Runtime мультиплексирует G на M через P (GMP).

В EH: на каждый gRPC unary — горутина из пула сервера; outbox workers — ограниченный пул; WS fan-out на LB — не «горутина на каждого игрока без backpressure». Цель ~10k DAU, не миллион голых потоков.

⚠️ «Горутина = поток ОС» или «бесплатно запускать go на каждый URL без лимита».

---

### 10. Конкурентность vs параллелизм vs асинхронность

**Вопрос:** Чем отличается асинхронное исполнение от параллельного? А конкурентность от параллелизма?

**Ответ:**  
**Конкурентность** — структура (много задач в прогрессе, возможно на 1 CPU). **Параллелизм** — одновременное исполнение на нескольких ядрах. **Асинхронность** — не блокировать caller: callback/future/event; в Go часто каналы/context + go.

В EH: purchase path асинхронный — Shop коммитит + outbox, NATS доставляет в Fulfillment/Notification (eventual). gRPC submit score конкурентен под нагрузкой; параллелизм даёт `GOMAXPROCS` и несколько gateway replicas.

⚠️ Смешивают все три слова в одно.

---

### 11. Каналы: buffered / unbuffered

**Вопрос:** Какие бывают каналы? Чем отличаются буферизованные и небуферизованные? Как закрывать?

**Ответ:**  
Unbuffered — синхронная передача (rendezvous). Buffered — очередь на N. Закрывает только отправитель; читать из closed — zero + ok=false; писать в closed — panic.

В EH каналы — внутри workers (outbox batch, fan-in метрик), не как публичный API между сервисами (там NATS/gRPC). Для «первый результат» — buffered `len(sources)`, чтобы writers не утекли.

⚠️ Закрывают канал с двух сторон; range без закрытия → deadlock.

---

### 12. select, default, nil channel

**Вопрос:** Как работает select? Что меняет default? Что если канал nil?

**Ответ:**  
select ждёт готовности case’ов; при нескольких — случайный. `default` — неблокирующий poll. nil-channel case никогда не готов — удобно отключать ветку.

В EH: worker loop `select { case <-ctx.Done(); case job := <-jobs; case <-ticker.C }` для outbox poller и health. default в hot loop без sleep — CPU spin (плохо для биллинга).

⚠️ Крутят default в busy-loop «для скорости».

---

### 13. context

**Вопрос:** Для чего нужен context, какие проблемы он решает?

**Ответ:**  
Отмена, дедлайны, request-scoped values (trace id). В EH: Gateway → gRPC передаёт ctx; service дергает PG/Redis/NATS с тем же ctx; таймаут на `/ready` ping. `defer cancel()` обязателен. Values — только cross-cutting (auth user id, trace), не «опции функции».

Jaeger spans живут на ctx между gateway и mesh.

⚠️ WithTimeout без cancel; ждут, что cancel убьёт горутину без проверки `ctx.Done()`.

---

### 14. Первый результат + отмена остальных

**Вопрос:** Запустить поиск по нескольким источникам параллельно, вернуть первый успех, остальных отменить.

**Ответ (скелет, как на собесе):**  
`ctx, cancel := context.WithCancel(parent); defer cancel()`; `ch := make(chan Result, len(srcs))`; в каждой go — работа с ctx, `select { case ch <- r; case <-ctx.Done() }`; main читает первый и cancel. Буфер = N или select на send — против leak.

В EH похожий паттерн мысли: multi-replica gateway health / «первый ответивший» не в проде как fan-out поиска, но для resilient client к нескольким backends идея та же. Межсервисно — gRPC + timeout, не самописный search.

⚠️ Небуферизованный ch + одна read → проигравшие висят на send.

---

### 15. Mutex vs RWMutex

**Вопрос:** Расскажи про sync.Mutex. Что такое RWMutex и когда выигрывает?

**Ответ:**  
Mutex — эксклюзив; всегда `defer Unlock`; не копировать; не reentrant. RWMutex — много RLock / один Lock; выигрыш при доминирующем read и не микроскопической секции.

В EH: in-memory least-conn state в balancer — кандидаты на RWMutex (частые чтения счётчиков коннектов, редкий update). Для balance — Redis, не локальный mutex на всю платформу.

⚠️ «RWMutex всегда быстрее»; мутация под RLock.

---

### 16. Data race на max; atomic vs Mutex

**Вопрос:** 1000 горутин обновляют общую `max`. Как убрать race? Когда atomic?

**Ответ:**  
Race: read-modify-write без синхронизации. Mutex на весь compare-and-set + WaitGroup. Atomic CAS — если одна ячейка; составной инвариант — Mutex. Всегда гонять `-race` в CI на unit-тестах auth/billing.

В EH критичные счётчики в hot path лучше в Prometheus/Redis; in-process — atomic для простых counters метрик.

⚠️ Защищают только запись max; забывают WaitGroup.

---

### 17. Примитивы синхронизации; sync.Once

**Вопрос:** Какие примитивы знаешь? Как устроен sync.Once и почему безопасен?

**Ответ:**  
Mutex/RWMutex, WaitGroup, Cond, Once, atomic, каналы как sync. `Once` — atomic flag + mutex: Do выполняет f ровно один раз, даже при панике внутри (повтор не вызывает — важно знать семантику).

В EH: `sync.Once` на init tracer/metrics register, lazy connect к JetStream в hub-клиентах. Не для бизнес «один раз на пользователя» — там Redis/DB unique.

⚠️ Once для «идемпотентности покупки» вместо unique reference_id.

---

### 18. Deadlock и goroutine leak

**Вопрос:** Что такое deadlock и goroutine leak? Детектит ли рантайм? Миллион горутин?

**Ответ:**  
Deadlock — все ждут друг друга; runtime детектит только **глобальный** deadlock (все G спят). Leak — горутина живёт вечно (забытый ctx, send в unbuffered). Миллион горутин возможен по памяти стеков, но убьёт scheduler/FD — в EH лимитируем workers (outbox, NATS consumers).

Инструменты: pprof goroutine, метрики `go_goroutines` в Grafana.

⚠️ «Runtime всегда найдёт deadlock»; бесконечный go на каждый HTTP без лимита.

---

### 19. Worker pool вместо go на каждый URL

**Вопрос:** Обойти большой список URL. Горутина на каждый — плохо. Какие паттерны?

**Ответ:**  
Фиксированный пул воркеров + jobs channel; или semaphore (`chan struct{}, N`); errgroup с лимитом. Backpressure: не загружать все URL в буфер сразу.

В EH: outbox worker читает batch из PG и публикует в NATS с ограниченной конкуренцией; k6/load не равен «миллион go». Gateway rate limit на login (~burst) — тоже форма backpressure.

⚠️ errgroup без SetLimit на 100k задач.

---

### 20. Fan-in / fan-out и pipeline

**Вопрос:** Реализуй `joinChannels` (fan-in). Pipeline и типичные дедлоки?

**Ответ:**  
Fan-out: много workers читают один in. Fan-in: N горутин читают свои ch и пишут в out; WaitGroup + close(out) когда все закончили. Pipeline: stage1 | stage2 | stage3 через каналы.

Дедлоки: никто не close; stage блокируется на send при мёртвом consumer; забытый ctx. В EH «pipeline» продуктовый — Game → outbox → NATS → Billing/LB/History, а не чистые каналы между процессами.

⚠️ Close(out) до завершения writers → panic.

---

## РАЗДЕЛ 03 — Runtime / GMP (21–24)

### 21. GMP и work stealing

**Вопрос:** Объясни GMP. Какую проблему решает? Переключение горутин?

**Ответ:**  
G — горутина, M — OS thread, P — логический процессор (очередь runnable). Лимит параллелизма ≈ `GOMAXPROCS` P. Work stealing: пустой P крадёт половину очереди у другого — балансировка. Переключение в user-space дешевле syscall.

Проблема: не плодить OS-thread на каждую задачу при I/O-bound gateway/gRPC. EH выигрывает на тысячах concurrent requests при малом числе M.

⚠️ Путают G и M; думают, что всегда 1:1 с потоками.

---

### 22. Блокирующий syscall и планировщик

**Вопрос:** Пример синхронного syscall; что делает планировщик?

**Ответ:**  
Пример: синхронный `read` файла / DNS без cgo-path в старых кейсах. При блокирующем syscall M «отлипает» от P; runtime может создать/взять другой M, чтобы P продолжал крутить G. Иначе один slow syscall стопорит весь P.

В EH hot path — сеть (netpoller) и PG (часто через poller/async-ish); долгие bcrypt(cost 12) на Auth — CPU, не syscall, но жрёт P → rate limit на login оправдан (видели на k6).

⚠️ «Любой I/O останавливает весь процесс».

---

### 23. Netpoller

**Вопрос:** Что такое netpoller? Чем сетевой I/O отличается от блокирующего syscall?

**Ответ:**  
Netpoller (epoll/kqueue) — runtime паркует G, пока FD не готов, **не удерживая** M на всём wait. Поэтому десятки тысяч conn на gateway/WS возможны без 1 thread/conn.

EH: Balancer `:8079`, Gateway, WS `/ws/leaderboard` — классический netpoller workload. PG драйверы тоже интегрируются с poller.

⚠️ Считают, что `conn.Read` всегда занимает OS thread на всё время wait.

---

### 24. GOMAXPROCS и asynchronous preemption

**Вопрос:** Что задаёт GOMAXPROCS? Как вытесняют тугой цикл?

**Ответ:**  
`GOMAXPROCS` ≈ число P (параллельных OS threads под Go-код). С Go 1.14+ async preemption: даже тугой цикл без вызовов функций могут вытеснить (сигналы/safe-points), чтобы не голодать GC и другие G.

В контейнерах EH смотрю cgroup CPU vs GOMAXPROCS (авто из quota в новых Go). Для Auth bcrypt — CPU-bound: лишние P не спасут без горизонтального scale.

⚠️ «Без вызова функции горутину нельзя вытеснить» (устарело).

---

## РАЗДЕЛ 04 — GC и память (25–30)

### 25. Трёхцветный GC

**Вопрос:** Как работает трёхцветный алгоритм GC в Go?

**Ответ:**  
Белые — кандидаты на сбор; серые — найдены, дети не просканированы; чёрные — живые, дети просмотрены. Concurrent mark-sweep: мутатор работает параллельно с mark; write barrier не даёт «потерять» указатель white←black. В конце белые — мусор.

В EH при росте alloc (JSON в gateway, большие shop catalogs) смотрю `go_memstats_*` и pprof heap, не «выключаю GC».

⚠️ Описывают stop-the-world stop-copy как единственную модель Go.

---

### 26. GOGC и GOMEMLIMIT

**Вопрос:** Что такое GOGC? Сколько резервирует? Чем помогает GOMEMLIMIT?

**Ответ:**  
`GOGC=100` (default): цель — куча растёт ~на 100% от live перед следующим циклом (упрощённо). Меньше GOGC → чаще GC, меньше RSS пики. `GOMEMLIMIT` — soft limit: GC старается удержать процесс ниже лимита (важно в k8s/cgroup), снижает OOM Killer риск.

В EH docker/k3s: лучше явный memory limit + GOMEMLIMIT, чем верить default на шумном shop/game.

⚠️ «GOGC=off всегда быстрее в проде».

---

### 27. Write barrier

**Вопрос:** Что такое write barrier в GC и зачем?

**Ответ:**  
Код, вставляемый при записи указателей во время concurrent mark: если чёрный объект начинает указывать на белый, белый помечается серым (или эквивалент Dijkstra/Yuasa-варианта в Go). Иначе GC «потерял бы» живой объект.

Для интервью достаточно: barrier корректен при параллельном мутаторе; цена — чуть дороже pointer writes.

⚠️ Путают с memory barrier CPU (store/load fence).

---

### 28. STW в Go GC

**Вопрос:** Когда STW и почему паузы короткие? Отдельные потоки GC?

**Ответ:**  
Короткие STW на старте/конце цикла (enable barrier, sweep termination и т.п.); mark в основном concurrent. GC использует mark workers на P + background. Паузы миллисекунды/суб-мс на нормальной куче — не «секунды как old Java».

В Grafana смотрю GC pause metrics / gogc; если p99 gateway растёт вместе с GC — режу alloc (буферы, меньше копий proto↔JSON).

⚠️ «Go GC полностью без STW» или «всегда long pause».

---

### 29. sync.Pool и GC

**Вопрос:** Как Pool взаимодействует с GC? Victim pool? Почему не кэш?

**Ответ:**  
Pool — переиспользование буферов между запросами; при GC содержимое может быть выброшено (victim — «полуочищенный» слой между циклами). Не гарантия хранения → **не кэш бизнес-данных**.

В EH: Pool уместен для `[]byte` под JSON encode в gateway; балансы/сессии — Redis. Иначе «пропал объект из Pool» удивляет на собесе и в проде.

⚠️ Используют Pool как LRU кэш пользователей.

---

### 30. SetH: переназначение указателя

**Вопрос:** `SetH(p *Person)` делает `p = &Person{…}` — что снаружи? Как починить?

**Ответ:**  
Снаружи без изменений (копия указателя). Фикс: менять поля `p.Name=…`, или `**Person`, или return new pointer. Escape analysis решит heap vs stack.

Связь с EH: при «обновить профиль» в service мутируем entity / пишем в repo, не переназначаем аргумент handler’а.

⚠️ Путают с Q08/Q30 одно и то же — нормально связать, показать рисунок памяти.

---

## РАЗДЕЛ 05 — Ошибки (31–32)

### 31. Panic vs error

**Вопрос:** В чём отличие паники от ошибки?

**Ответ:**  
`error` — ожидаемый сбой (нет денег на покупку, invalid JWT, not found) → return + gRPC status. `panic` — нарушен инвариант / баг; процесс не должен «ловить бизнес через panic».

В EH: shop purchase → error codes; Recovery interceptor на каждом gRPC server (`ChainUnaryInterceptor(Recovery(), Logger(), Validate())`) — последний рубеж, логирует и не валит весь процесс.

⚠️ panic для «user not found».

---

### 32. recover: где вызывать

**Вопрос:** Как обработать панику? Где вызывать recover?

**Ответ:**  
Только в defer той горутины, где panic. На границе: HTTP middleware / gRPC Recovery. Внутри business — почти никогда.

В EH: `pkg/interceptor` / per-service `Recovery()`; gateway Gin recovery. После recover — метрика/лог + 500/Internal, не silent swallow без сигнала.

⚠️ recover в обычной функции без defer; «recover в main спасёт все горутины».

---

## РАЗДЕЛ 06 — БД и транзакции (33–38)

### 33. MVCC в Postgres

**Вопрос:** Зачем Postgres MVCC? Что это?

**Ответ:**  
Несколько версий строки: читатели не блокируют писателей. UPDATE создаёт новую версию; старые чистит VACUUM. Цена — bloat / autovacuum.

В EH покупка/списание билетиков и запись outbox в одной TX опираются на MVCC+ACID: параллельные submit score / purchase не стопорят весь сервис на table lock.

⚠️ «MVCC = нет vacuum никогда».

---

### 34. Уровни изоляции и аномалии

**Вопрос:** Какие уровни изоляции и какие аномалии закрывают?

**Ответ:**  
Read Uncommitted / Read Committed (PG default) / Repeatable Read / Serializable. Аномалии: dirty read, non-repeatable, phantom, serialization anomaly. PG RC не даёт dirty read; RR/Serializable — строже для money-like.

В EH для billing/shop TX обычно default RC + явные constraints/unique (`reference_id` идемпотентности). Где нужна строгая сериализация спорных обновлений — `SELECT … FOR UPDATE` на строке баланса.

⚠️ Зубрят ANSI, не зная отличий Postgres.

---

### 35. Блокировки, долгие TX, пул соединений

**Вопрос:** Связь блокировок Postgres и исчерпания коннектов? Как пул?

**Ответ:**  
Долгая TX держит locks + занимает слот пула → другие запросы ждут → каскадный timeout → «утечка» busy conns. Лечить: короткие TX, таймауты ctx, не ходить в NATS внутри TX (только outbox row).

EH стандарт пулов: MaxOpen/MaxConns=25, MinIdle/MinConns=10, MaxConnLifetime=5m (`database/sql` или pgxpool в сервисах). Postgres per service — blast radius меньше.

⚠️ Открывают TX на весь HTTP request с внешними вызовами.

---

### 36. Покрывающий индекс

**Вопрос:** Что такое covering index? Виды индексов?

**Ответ:**  
Индекс, из которого запрос берёт все нужные колонки без heap fetch (Index Only Scan). В PG помогают INCLUDE-колонки / составные индексы под WHERE+ORDER BY. Виды: B-tree (default), Hash, GiST/GIN (FTS/jsonb/geo), BRIN.

EH: индексы под `(user_id)`, `(game_id, score DESC)` на LB/game; billing `user_currencies(user_id)`. Смотрю `EXPLAIN ANALYZE`, не гадаю.

⚠️ «Любой индекс покрывающий».

---

### 37. Шардирование 100M товаров

**Вопрос:** Как шардировать БД со 100M товаров? Ключ? Hot key?

**Ответ:**  
Сначала вертикально (каталог vs purchases) и read-cache. Шард key: `item_id` / `hash(item_id)` для равномерности; для user-centric — `user_id` (inventory). Hot key (вирусный товар) — отдельный кеш, возможно выделенный шард/replica.

В EH сейчас шардинг не нужен (~10k DAU): shop PG + Redis cache. На рост — шардировать inventory/purchases по `user_id`, каталог — отдельный read model.

⚠️ Шардируют по `created_at` и ловят cross-shard joins на всё.

---

### 38. Полнотекстовый поиск

**Вопрос:** Как сделать FTS? Движок? Как работает?

**Ответ:**  
Для EH-масштаба: Postgres `tsvector` + GIN достаточно для поиска по shop items / authors. Выше нагрузка / сложная релевантность — OpenSearch/ES. Индексация: документ → токены → posting lists; запрос → rank.

Не тащить ES «на будущее» без боли: у нас каталог и авторы умещаются в PG.

⚠️ Сразу Kafka+ES на 1k товаров.

---

## РАЗДЕЛ 07 — Брокеры (39–41)  
*(в банке — Kafka; в EH канон — NATS JetStream, Kafka opt-in)*

### 39. Партиции / порядок (Kafka) ↔ JetStream subjects

**Вопрос:** Сколько consumer’ов в группе при 3 партициях? Где offsets? Порядок при увеличении партиций?

**Ответ:**  
В Kafka: >3 consumers в группе на 1 topic с 3 partitions — idle consumers; порядок **внутри партиции**; offsets в `__consumer_offsets` (или клиент). Добавление партиций ломает ключ→partition для старых ключей → порядок глобально не сохранить.

**Как отвечаю про EH:** основной путь — JetStream stream `EVENTS`, durable consumers (billing/fulfillment/history). «Партиции» ≈ subject design + consumer replicas; порядок per-subject/stream sequence, не «глобальный total order». Kafka — `make deploy-heavy`, не thin path.

⚠️ Обещают global ordering по всему топику.

---

### 40. Лаг 1M сообщений

**Вопрос:** Топик с 3 партициями, лаг 1M. Как разгрести?

**Ответ:**  
Сначала метрики: скорость produce vs consume, slow handler, poison message. Scale consumers до #partitions; ускорить обработку (batch, меньше sync PG); временно pause producers; для poison — DLQ. Добавить партиций — только если ключи позволяют rebalance.

В EH: Grafana JetStream lag / pending; порог «>1000 = bad» в ops watchlist. Outbox + ack/nak; idempotency по `reference_id`, чтобы при replay не удвоить баланс.

⚠️ Слепо крутят replicas >> partitions.

---

### 41. DLQ и Outbox

**Вопрос:** Зачем DLQ? Что такое Outbox и почему DBA его любит?

**Ответ:**  
DLQ — сообщения, которые не обработались после N попыток: не блокируют очередь, можно расследовать. Outbox: в одной TX с бизнес-строкой пишем `outbox(event_type, payload)`; worker публикует в NATS и помечает sent — нет «коммит без события» / «событие без коммита».

EH Outbox: **Game, Billing, Shop, Inventory, Payment, Authors**. Purchase: Shop → NATS `purchase.paid` → Fulfillment/Notification. DBA любит: целостность в PG, без 2PC с брокером.

⚠️ Publish в NATS внутри TX до commit; или commit без outbox «потом опубликую».

---

## РАЗДЕЛ 08 — System Design (42–44)

### 42. Кеш в архитектуре

**Вопрос:** Где кеш? Что кешировать? Ключ? Инвалидация?

**Ответ:**  
Слои: CDN/static → gateway → Redis beside service → DB. EH: Auth sessions Redis; Billing balance Cache-Aside (PG write → delete Redis key); Leaderboard hot tops — Redis Sorted Set; Shop catalog — короткий TTL/invalidate on change.

Ключи: `balance:{user}:{currency}`, session by token id. Инвалидация: delete-on-write; TTL как safety net. Не кешировать «истину» денег без идемпотентности на write path.

⚠️ Кеш как единственный source of truth; длинный TTL без инвалидации на purchase.

---

### 43. CQRS при 100K RPS (20% write / 80% read)

**Вопрос:** Как спроектировать write/read товаров? Имя паттерна? Trade-offs?

**Ответ:**  
CQRS: отдельная write-модель (нормализованный PG shop/inventory) и read-модель (денормализованный Redis/CH/replica) под каталог и LB. Пишем в write → событие (outbox/NATS) → обновляем read. Trade-off: eventual consistency, сложность синка, выигрыш по scale read.

В EH зачаток CQRS: write path shop/billing vs read LB из Redis SS + Analytics в ClickHouse. При 100K RPS — горизонтально gateway, шарды, read replicas; не один монолитный SQL на всё.

⚠️ «CQRS = обязательно Event Sourcing».

---

### 44. Transactional Outbox для уведомлений/событий

**Вопрос:** Надёжная отправка событий при создании заказа?

**Ответ:**  
1) BEGIN  
2) INSERT order/purchase + INSERT outbox  
3) COMMIT  
4) Worker SELECT FOR UPDATE SKIP LOCKED / batch → JetStream Publish → mark sent  

Consumer идемпотентен. В EH то же для purchase/balance/score. Не слать notification в TX; Notification — отдельный consumer.

⚠️ «Сначала Kafka, потом DB» или dual-write без outbox.

---

## РАЗДЕЛ 09 — OS / CS (45–47)

### 45. Виртуальная vs физическая память, OOM, swap

**Вопрос:** Почему VIRT 1.4TB при малой RAM? На что смотрит OOM Killer? Swap?

**Ответ:**  
Виртуальная память — адресное пространство (резервы, mmap); RSS/физическая — реально занятые страницы. OOM Killer смотрит на **физическое** давление (RSS/oom_score_adj), не на «VIRT красивый в top». Swap откладывает OOM ценой latency thrash — для latency-sensitive gateway/gRPC обычно плохо.

В EH контейнерах важны cgroup limits + `GOMEMLIMIT`, мониторинг RSS в Grafana, не пугаться высокого VIRT у Go.

⚠️ «VIRT = сколько реально съели».

---

### 46. Почему OS thread дорогой

**Вопрос:** Почему создание потока ОС дорого (кроме стека)? Связь с Go?

**Ответ:**  
Syscall clone, kernel structures, scheduler accounting, TLB/cache effects, FD/guard pages. Поэтому Go держит пул M и дешёвые G. Важно: не обходить модель, создавая OS threads вручную (`LockOSThread` только когда надо).

EH: тысячи concurrent requests ≠ тысячи threads; bcrypt/CPU — scale replicas, не `runtime.LockOSThread` на каждый login.

⚠️ «Стек — единственная цена потока».

---

### 47. Системные вызовы

**Вопрос:** Что такое syscall? Пример sync и async?

**Ответ:**  
Переход user→kernel: open/read/write/socket. Синхронный: блокирующий `read` файла. «Асинхронный» I/O: epoll/io_uring — регистрируем интерес, ядро будит позже (на этом стоит netpoller).

В EH почти весь сетевой путь — non-blocking + poller; метрики/логи — не делать sync disk flush на каждый request.

⚠️ Путают async syscall с «горутиной».

---

## РАЗДЕЛ 10 — Алгоритмы (48–50)

### 48. Дерево по preorder + inorder

**Вопрос:** Построить бинарное дерево по preorder и inorder; вернуть корень.

**Ответ:**  
Preorder[0] — корень; в inorder делим на левое/правое поддерево; рекурсивно на срезах (или индексный hashmap value→inorder idx для O(n)). Время O(n), память O(n) на map + стек рекурсии.

К EH не привязано — чисто алгоритмический билет; могу сравнить с AST/expression trees, но в проде игр не строю деревья так.

⚠️ O(n²) от линейного поиска корня в inorder на каждом шаге без map.

---

### 49. Big O: время и память

**Вопрос:** Временная и пространственная сложность? Big O vs Ω vs Θ? N vs N log N?

**Ответ:**  
O — верхняя асимптотическая оценка; Ω — нижняя; Θ — плотно. Складываем доминирующий член. N log N растёт быстрее N; log N — почти константа на практике для больших N, но не «равно O(1)».

Примеры: max в массиве O(n)/O(1); binary search O(log n)/O(1); outbox batch publish O(k) на batch. На собесе отдельно говорю про амортизацию append и про то, что сеть/БД константы важнее micro Big O в EH.

⚠️ «N и N log N почти одно».

---

### 50. Паттерны: палиндром, скобки, группировка

**Вопрос:** Палиндром; валидация скобок; группировка слов по длине.

**Ответ:**  
- **Палиндром:** два указателя с концов, `[]rune` для Unicode — O(n)/O(1) доп.  
- **Скобки:** стек открывающих; при закрывающей — pop и match; конец — стек пуст — O(n).  
- **Группировка по длине:** `map[int][]string` — O(n) по словам.

Связь с EH слабая (валидация nested JSON/proto реже стеком); на собесе показываю чистый код + сложности. Для production validation у нас proto `validate.rules` + interceptor, не самописный скобочный парсер.

⚠️ Палиндром через reverse всей строки без нужды; скобки без стека «счётчиком» для разных типов.

---

## Как пользоваться этим файлом

1. Читай вопрос в [`OCTOBER_INTERVIEW.md`](./OCTOBER_INTERVIEW.md) (формулировка как на моке).  
2. Отвечай **своими** словами из секции выше + 1 конкретный EH-пример (сервис/порт/паттерн).  
3. Если уводят в Kafka-only — честно: thin path = NATS JetStream, Kafka opt-in.  
4. Не выдумывай метрики: опирайся на README / LOAD_POSTURE / факты k6 из `confluence/interview/5.testing`.
