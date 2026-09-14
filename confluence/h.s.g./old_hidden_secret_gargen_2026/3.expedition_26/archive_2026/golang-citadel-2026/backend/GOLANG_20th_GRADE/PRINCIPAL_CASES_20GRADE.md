Кейс 1: Почему make([]byte, 1000000) не всегда вызывает GC?
Короткий ответ
Потому что make аллоцирует память, но не триггерит GC. GC триггерится по объёму выделенной памяти с прошлого GC, а не по факту аллокации.

Развёрнуто
Правило запуска GC:

text
GC запускается, когда:
выделено_с_прошлого_GC >= heap_live * (GOGC/100)
Пример с make:

go
// Живая память: 10MB
// GOGC=100 (по умолчанию)

for i := 0; i < 1000; i++ {
_ = make([]byte, 1<<20) // 1MB
}
// Что произойдёт:
// - Первые 10 аллокаций: GC не запустится (10MB выделено, target=20MB)
// - 11-я аллокация: превысили target → запуск GC
// - GC удаляет все предыдущие []byte (на них нет ссылок)
// - Живая память: 10MB (долгоживущие объекты) + текущий 1MB
// - Цикл продолжается
Почему это может быть проблемой:

go
func badCase() {
for i := 0; i < 1_000_000; i++ {
// Создаём 1KB, но GC не запускается
_ = make([]byte, 1024)
// Через 1000 итераций выделили 1MB, target не достигнут
// Через 10000 итераций — 10MB, target достигнут → GC
// Через 1_000_000 итераций — GC запустится всего несколько раз
}
}
А если make внутри цикла с накоплением:

go
func dangerousCase() {
var bufs [][]byte
for i := 0; i < 1000; i++ {
buf := make([]byte, 1<<20) // 1MB
bufs = append(bufs, buf)   // сохраняем ссылку!
// GC не запустится, пока heap_live не удвоится
// Но heap_live растёт с каждым append!
// При 1000 итераций: heap_live = 1GB
// GC запустится только когда выделится ещё 1GB (ещё 1000 итераций)
// Итог: программа может упасть с OOM до первого GC
}
}
Как заставить GC запуститься чаще
go
// 1. Ручной вызов (не делай в проде без причины)
runtime.GC()

// 2. Уменьшить GOGC
debug.SetGCPercent(50) // запускать при +50% к живой памяти

// 3. Подсказать GC, что память больше не нужна
buf = nil // обнули ссылку
Кейс 2: Почему bytes.Buffer может быть опасным для GC?
Короткий ответ
bytes.Buffer хранит внутренний слайс buf []byte, который растёт при записи. Если ты используешь один Buffer долго, его внутренний буфер становится большим и никогда не уменьшается автоматически.

Развёрнуто
Как растёт bytes.Buffer:

go
// Внутри bytes.Buffer
type Buffer struct {
buf []byte
off int // сколько прочитано
}

// При записи больше чем cap(buf)
func (b *Buffer) grow(n int) {
// Новая ёмкость = max(2*cap(b.buf), n)
newcap := max(2*cap(b.buf), n)
newbuf := make([]byte, newcap)
copy(newbuf, b.buf[b.off:])
b.buf = newbuf
}
Проблема в коде:

go
// Плохо: буфер растёт и никогда не сжимается
var buffer bytes.Buffer
for {
buffer.WriteString("hello")          // растёт
buffer.WriteString("world")
result := buffer.String()
buffer.Reset()                        // Reset не уменьшает cap!
// buffer.buf всё ещё 1KB, 2KB, 1MB, 100MB...
// GC не может удалить эту память, потому что buffer.buf жива
process(result)
}
Что происходит с GC:

Буфер вырос до 1GB

Ты вызвал Reset() — buffer.buf всё ещё 1GB, просто buffer.off = 0

GC видит: buffer.buf — чёрный объект (достижим через глобальную или стековую переменную)

GC никогда не удалит эту память, пока buffer в области видимости

Программа жрёт 1GB RAM постоянно

Как исправить
go
// Способ 1: Создавать новый буфер каждый раз
for {
buffer := &bytes.Buffer{}          // новый, маленький
buffer.WriteString("hello")
result := buffer.String()
// buffer уходит из области видимости → GC удалит
}

// Способ 2: Периодически сбрасывать буфер через sync.Pool
var bufferPool = sync.Pool{
New: func() interface{} { return &bytes.Buffer{} },
}

for {
buffer := bufferPool.Get().(*bytes.Buffer)
buffer.WriteString("hello")
result := buffer.String()
buffer.Reset()
bufferPool.Put(buffer)              // переиспользуем, но не держим вечно
}

// Способ 3: Принудительное сжатие (редко)
func shrinkBuffer(b *bytes.Buffer) {
if b.Cap() > 1024*1024 { // больше 1MB
// Создаём новый с нужной ёмкостью
newBuf := bytes.NewBuffer(make([]byte, 0, b.Len()))
newBuf.Write(b.Bytes())
*b = *newBuf
}
}
Почему strings.Builder лучше?
go
// strings.Builder имеет ту же проблему, но его Reset сбрасывает буфер
var sb strings.Builder
sb.WriteString("hello")
sb.Reset()   // sb.buf = nil! память отдаётся GC
Вывод: Для кратковременных операций — strings.Builder. Для долгих — bytes.Buffer с пулом.

Кейс 3: Как читать pprof и находить утечки памяти
Короткий ответ
Утечки памяти в Go — это обычно не "забыл free" (как в C), а "держу ссылку дольше, чем надо".

Типы утечек
Тип	Симптом	Причина
Накопление в слайсе/мапе	Память растёт линейно	Забыл удалять элементы
Забытый указатель	Объект не удаляется, хотя не нужен	Глобальная переменная хранит ссылку
Закрытая горутина	Стек горутины не освобождается	Горутина ждёт канал вечно
Сборщик логов	Память растёт, но GC не помогает	Логи хранятся в глобальном буфере
Как читать pprof
Шаг 1: Включить профилирование

go
import _ "net/http/pprof"

func main() {
go func() {
log.Println(http.ListenAndServe("localhost:6060", nil))
}()
// ... твой код
}
Шаг 2: Забрать профиль памяти

bash
# Сравниваем два состояния (через 5 минут)
curl -o mem1.prof http://localhost:6060/debug/pprof/heap
sleep 300
curl -o mem2.prof http://localhost:6060/debug/pprof/heap

# Смотрим разницу
go tool pprof --base=mem1.prof mem2.prof
Шаг 3: Команды внутри pprof

text
(pprof) top10 -cum
# Показывает, какие функции больше всего нааллоцировали

(pprof) list dangerousFunc
# Показывает, какая строка кода жрёт память

(pprof) web
# Генерирует граф вызовов (требует graphviz)

(pprof) inuse_space
# Память, которая всё ещё используется (живая)

(pprof) alloc_space
# Вся память, которая когда-либо была выделена (включая удалённую)
Реальный пример: находим утечку
Плохой код:

go
var cache = make(map[string][]byte)

func handler(w http.ResponseWriter, r *http.Request) {
key := r.URL.Query().Get("key")

    // Утечка: cache никогда не чистится
    if _, ok := cache[key]; !ok {
        data := fetchFromDB(key) // 10MB
        cache[key] = data
    }
    w.Write(cache[key])
}
Что видим в pprof:

text
(pprof) top10
Showing nodes accounting for 10.50GB, 100% of 10.50GB total
flat  flat%   sum%        cum   cum%
5.20GB 49.52% 49.52%     5.20GB 49.52%  main.handler
4.80GB 45.71% 95.23%     4.80GB 45.71%  main.fetchFromDB
0.50GB  4.76%   100%     0.50GB  4.76%  runtime.mapassign

(pprof) list handler
Total: 10.50GB
ROUTINE ======================== main.handler
5.20GB     5.20GB (flat, cum) 49.52% of Total
.          .     20:   key := r.URL.Query().Get("key")
.          .     21:   
.          .     22:   if _, ok := cache[key]; !ok {
.          .     23:       data := fetchFromDB(key)
5.20GB     5.20GB     24:       cache[key] = data   // <-- УТЕЧКА
.          .     25:   }
.          .     26:   w.Write(cache[key])
Решение:

go
// Добавляем TTL или LRU
type CacheItem struct {
data      []byte
createdAt time.Time
}

var cache = make(map[string]CacheItem)

func cleanup() {
for k, v := range cache {
if time.Since(v.createdAt) > 5*time.Minute {
delete(cache, k)
}
}
}
// Или использовать github.com/hashicorp/golang-lru
Как найти утечку горутин
bash
# Забрать профиль горутин
curl -o goroutines.prof http://localhost:6060/debug/pprof/goroutine

# Смотрим
go tool pprof goroutines.prof
(pprof) traces
# Показывает стек всех живых горутин
Пример утечки:

go
func leakyFunc() {
ch := make(chan int)
go func() {
<-ch  // ждёт вечно
}()
// ch никогда не закрывается, горутина висит
}
Что видим:

text
(pprof) traces
goroutine 42 [chan receive]:
main.leakyFunc.func1()
/main.go:42
runtime.gopark()
/runtime/proc.go:366
Продвинутая техника: --base и аллокации
bash
# Замеряем до и после подозрительной операции
go tool pprof --base=before.prof after.prof

(pprof) top -cum
# Разница покажет, что нааллоцировалось за время между замерами
Кейс 4: Почему GC не возвращает память ОС (и когда это проблема)
Короткий ответ
Go держит память в mheap, потому что завтра она может снова понадобиться. Возврат памяти ОС — дорогая операция (munmap → page fault при следующем использовании).

Развёрнуто
Что видит администратор:

bash
$ top -p <pid>
PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
12345 app       20   0   10.5g   1.2g   1.1g S   2.0  15.2   1:23.45 myapp
# VIRT = 10.5GB (зарезервировано через mmap)
# RES  = 1.2GB  (реально используется)
Почему RES не падает после GC:

GC удалил объекты (теперь они белые)

Страницы памяти, где лежали эти объекты, помечены как "свободные в mheap"

Но эти страницы не возвращены ОС (нет munmap)

При следующей аллокации Go просто возьмёт их из mheap — без системного вызова

Когда память возвращается:

sysmon периодически вызывает scavenger

scavenger смотрит: страница не использовалась >5 минут?

Если да — вызывает madvise(MADV_DONTNEED) (Linux)

Ядро помечает страницу как "можно забрать"

При следующем обращении — page fault и выделение новой физической страницы

Принудительный возврат:

go
import "runtime/debug"

debug.FreeOSMemory() // вызывает GC + возвращает память (ДОРОГО!)
Когда это проблема:

Kubernetes: лимит памяти 2GB, процесс резервирует 1.5GB, но использует 500MB

K8s смотрит на RES (используемую), а не на VIRT

Если RES не падает после GC — может показаться, что память утекает

На самом деле — просто Go держит кэш страниц

Как уменьшить RES:

bash
# Настройка через GODEBUG
GODEBUG=madvdontneed=1 ./myapp
# Заставляет Go сразу вызывать MADV_DONTNEED при освобождении
Финальный тест (для Senior+)
Ты должен ответить:

Почему make([]byte, 1e9) не вызывает GC, если это 1GB аллокации?

GC триггерится по объёму с прошлого GC, а не по размеру одной аллокации

Почему bytes.Buffer может быть причиной OOM?

Внутренний буфер растёт и никогда не сжимается, даже после Reset

В чём разница между inuse_space и alloc_space в pprof?

inuse — живая память (то, что GC не удалил), alloc — вся память за всё время

Почему Go не возвращает память ОС сразу после GC?

Чтобы не платить за page fault при следующей аллокации

Как найти утечку горутины через pprof?

/debug/pprof/goroutine → traces → ищем горутины в состоянии chan receive или select