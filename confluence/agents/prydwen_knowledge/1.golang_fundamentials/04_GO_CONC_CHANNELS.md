# Go: каналы

Канал — типизированная очередь для передачи значений между горутинами. На senior-уровне ждут: семантика блокировок, close, select, и когда канал хуже mutex.

## Unbuffered vs buffered

| | Unbuffered `make(chan T)` | Buffered `make(chan T, N)` |
|--|---------------------------|----------------------------|
| send | блокируется, пока другой не recv | блокируется, если буфер полон |
| recv | блокируется, пока не будет send | блокируется, если буфер пуст |
| синхронизация | rendezvous (happens-before) | асинхронная очередь до N |

- Unbuffered — синхронизация «передал = принял».
- Buffered — backpressure до N; `N` без расчёта = скрытая unbounded очередь (утечка памяти под нагрузкой).

## Направление типов

```go
chan T       // двусторонний
chan<- T     // только send
<-chan T     // только recv
```

Сужайте тип на границе API — меньше шансов закрыть/писать «не той» стороной.

## Close

- Закрывает **только отправитель** (один owner). Двойной close → panic.
- Recv из closed: сразу zero value + `ok == false`.
- Send в closed → panic.
- `close` не обязателен для GC: канал собирается, если на него нет ссылок; close нужен как сигнал «данных больше не будет» (часто для `range`).

```go
v, ok := <-ch
for v := range ch { /* до close */ }
```

## Select

- Ждёт готовности нескольких case; если несколько готовы — **случайный** выбор (fairness).
- `default` — non-blocking poll (осторожно: hot-loop жрёт CPU).
- Паттерн отмены:

```go
select {
case job := <-jobs:
    handle(job)
case <-ctx.Done():
    return ctx.Err()
}
```

- Timeout: `time.After` в select удобно, но в горячем цикле лучше `time.NewTimer` + `Stop` (иначе leak таймеров).

## Nil channel

- Send/recv на `nil` channel **блокируются навсегда**.
- Полезно в select: отключить case, присвоив `ch = nil`.

## Типичные ошибки

- Забыть close → `range` висит навсегда (goroutine leak).
- Close со стороны consumer / из нескольких sender без координации.
- Использовать channel как mutex «потому что идиоматично» — для shared state часто `Mutex` проще и быстрее.
- Unbounded `go` + unbuffered без читателя → deadlock.
- Большие struct по каналу — копируются; для тяжёлых данных передавайте указатель и продумайте ownership.

## Channel vs Mutex (эвристика)

- **Channel:** передача владения, pipeline, fan-in/out, сигналы жизни/смерти.
- **Mutex:** защита in-place структуры (кэш map, счётчики с логикой).
- **atomic:** один счётчик/флаг без сложной инварианты.

## Event Horizon

Workers (outbox, analytics NATS→ClickHouse): job channel или подписка + `ctx.Done()` в select; при shutdown не оставлять send без recv. gRPC stream / gateway — backpressure через контекст запроса, не через бесконечный буфер.

## Типичные вопросы на собесе

- Чем buffered отличается от unbuffered на уровне блокировок?
- Кто должен делать `close`? Что будет при send в closed / double close?
- Как работает `select` при двух готовых case?
- Зачем nil channel в select?
- Когда предпочтёте mutex, а когда channel?
- Как корректно сделать graceful stop consumer’а на канале?
