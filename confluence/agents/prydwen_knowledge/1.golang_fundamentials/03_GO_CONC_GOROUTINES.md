# Go: горутины

Горутина — лёгкий поток выполнения, планируемый рантаймом Go (GMP). На собесе ждут не «это дешёвые треды», а понимание lifecycle, утечек и связи с `GOMAXPROCS`.

## Старт и стоимость

- `go f()` — ставит G в очередь текущего P; стек начинается маленьким (~2 KiB) и растёт/сжимается.
- Дешевле OS thread, но не бесплатно: каждая живая горутина — стек + scheduling overhead + возможные блокировки.
- Не спауньте «горутину на каждый байт сети» без backpressure (семафор, worker pool, buffered channel).

## Типичные паттерны

- **fan-out / fan-in:** N воркеров, один aggregator.
- **worker pool:** фиксированный N = `GOMAXPROCS` или кратно I/O-bound нагрузке.
- **per-request goroutine:** HTTP/gRPC handler уже в своей горутине (net/http, grpc server) — не плодите лишние без нужды.
- **background workers:** outbox publisher, NATS/Kafka consumers в EH — отдельная горутина + graceful shutdown через context.

## Утечки горутин (leaks)

Горутина жива, пока функция не вернулась. Утечка = блокировка навсегда или забытый цикл.

Частые причины:

- send/recv на channel, куда никто не пишет/не читает;
- `select` без `ctx.Done()` / timeout;
- `WaitGroup` без `Done` при early return;
- HTTP-клиент без timeout (`http.DefaultClient` — запрещён в EH rules);
- забыли отменить `context` у child workers при shutdown.

Симптомы: рост `runtime.NumGoroutine()`, память, деградация latency. Диагностика: `pprof` goroutine dump, `GODEBUG=schedtrace`.

**EH:** при `SIGTERM` — cancel root context → workers выходят → closer закрывает DB/gRPC. Не хранить context в struct навечно; передавать вниз по вызову.

## Panic в горутине

- Panic в дочерней горутине **не** ловится `recover` родителя — процесс падает (если не recover внутри самой горутины).
- На gRPC: unary interceptor `Recovery()` в chain (EH стандарт) ловит panic в handler, не в произвольных background jobs — для workers свой recover + лог/метрика.

## GOMAXPROCS

- `GOMAXPROCS` = число P (логических процессоров), одновременно выполняющих Go-код.
- По умолчанию = `runtime.NumCPU()` (с Go 1.5+). В контейнерах учитывайте cgroup CPU limit (Go 1.19+ / `automaxprocs` — иначе думаете, что у вас 64 ядра на shared node).
- CPU-bound: N ≈ GOMAXPROCS. I/O-bound: можно больше воркеров, но упираетесь в пул БД (в EH MaxConns=25) и в сеть.
- Крутить `GOMAXPROCS` «на глаз» без метрик — антипаттерн; смотрите CPU throttling в k8s.

## Когда не нужна горутина

- Короткий CPU-кусок в том же запросе.
- Уже есть пул (pgx pool сам мультиплексирует).
- Нужен строгий порядок без координации — проще синхронно.

## Типичные вопросы на собесе

- Чем горутина отличается от OS thread?
- Как детектить goroutine leak? Приведите 2–3 сценария утечки.
- Что делает `GOMAXPROCS`? Что будет в Kubernetes при лимите 0.5 CPU и дефолтном GOMAXPROCS на большой машине?
- Почему panic в `go func()` роняет процесс?
- Как правильно останавливать background worker в микросервисе (EH-style)?
