🎯 ПОЛНЫЙ СПИСОК ПОВТОРЕНИЯ (ЧЕК-ЛИСТ)
🔴 КРИТИЧЕСКИ ВАЖНО (были вопросы)
TCP vs UDP — различия, когда что использовать (логи vs стриминг, батчинг для критичных логов)

TLS Handshake — пошагово: ClientHello → ServerHello → сертификат → обмен ключами → готово

HTTP vs HTTPS — разница, порты (80/443), шифрование

sync пакет — WaitGroup (ждёт горутины), Mutex (защита ресурса), Once (однократный вызов), RWMutex

Каналы — буферизированные/небуферизированные, close(), range по каналу, кто закрывает (писатель), select

Map — ключи должны быть сравнимыми (не слайс, не мапа, не функция), конкурентный доступ (sync.Map или мьютекс)

ООП в Go — встраивание (embedding), композиция, инкапсуляция (экспорт по букве), полиморфизм (интерфейсы, утиная типизация)

Порты — 22 (SSH), 80 (HTTP), 443 (HTTPS), 5432 (Postgres), 6379 (Redis), 8080 (dev), 9090 (Prometheus)

SQL — SELECT с GROUP BY, COUNT(*), ORDER BY

🟡 ВАЖНО (могут спросить)
Контекст — context.WithCancel, WithTimeout, WithDeadline, WithValue. Отмена горутин, передача метаданных

Ошибки — кастомная ошибка (тип реализующий error), errors.Is, errors.As, wrapping (%w). БЕЗ ПАКЕТА ERRORS (через struct)

Базы данных — транзакция (ACID), SAVEPOINT, индексы (B-tree, GiST в Postgres), MVCC

Типы БД — реляционные (Postgres), ключ-значение (Redis), колоночные (ClickHouse), документоориентированные (MongoDB)

Интерфейсы — пустой интерфейс (any), Stringer, error, когда использовать

Структуры — методы на указателях vs значениях, встраивание (*user vs user), String()

🟢 ДЛЯ ГЛУБИНЫ (для сеньорского уровня)
NATS — Core vs JetStream, queue group, ack'и

Kafka — retention, партиционирование, consumer groups, exactly-once

gRPC vs REST — HTTP/2, protobuf, строгий контракт, streaming

Профилирование — pprof, go tool pprof -http, flame graph, CPU/memory профиль

K8s — Helm, HPA, canary vs blue-green, Istio (VirtualService, destination weights)

Outbox pattern — как работает, гарантии доставки, идемпотентность, очистка таблицы

Observability — OpenTelemetry, метрики, трейсы, логи, алерты

🟣 ТВОИ КОНКРЕТНЫЕ ПРОВАЛЫ НА СОБЕСАХ
TCP vs UDP — подробно, не просто "TCP надёжный, UDP быстрый", а с примерами (логи → TCP с батчингом, стриминг/игры → UDP)

TLS handshake — пошагово без воды

sync.WaitGroup — Add/Done/Wait, что будет если забыть Done

sync.Mutex — Lock/Unlock, отличие от RWMutex, data race

sync.Once — для чего, пример (синглтон, инициализация)

Закрытие каналов — кто закрывает (писатель), что будет если читать из закрытого, range по закрытому

GoLand — отключить автодополнение, чтобы не позориться

📋 ШПАРГАЛКА (КОРОТКО)
TCP vs UDP
TCP	UDP
Соединение	Да (3-way handshake)	Нет
Надёжность	Да (подтверждение)	Нет (отправил и забыл)
Порядок	Да	Нет
Скорость	Медленнее	Быстрее
Примеры	HTTP, SSH, БД	DNS, VoIP, игры, стриминг, логи (если можно потерять)
Когда использовать	Критичные данные, логи (с батчингом)	Некритичные данные, высокая скорость
TLS Handshake (коротко)
ClientHello → (версия TLS, шифры, random)

ServerHello → (выбранные шифры, random, сертификат)

Client проверяет сертификат (CA, срок, домен)

Обмен ключами (Diffie-Hellman или RSA)

Готово → симметричное шифрование

sync пакет
go
// WaitGroup — ждём горутины
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()
wg.Wait()

// Mutex — защита ресурса
var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()

// RWMutex — много читателей, один писатель
var rwmu sync.RWMutex
rwmu.RLock()   // много
rwmu.RUnlock()
rwmu.Lock()    // эксклюзивно
rwmu.Unlock()

// Once — однократный вызов
var once sync.Once
once.Do(func() {
    // выполнится 1 раз
})
Каналы (важное)
go
ch := make(chan int)       // небуферизированный (блокируется)
ch := make(chan int, 10)   // буферизированный (до 10)

close(ch)   // закрывает ПИСАТЕЛЬ

// Читать
val, ok := <-ch  // ok == false если закрыт и пуст

// range требует close()
for val := range ch {
    fmt.Println(val)
}

// select — неблокирующая работа с каналами
select {
case val := <-ch1:
    // обработали
case ch2 <- 42:
    // отправили
case <-ctx.Done():
    return ctx.Err()
default:
    // неблокирующая проверка
}
Кастомная ошибка (без пакета errors)
go
type MyError struct {
    Code    int
    Message string
}

func (e MyError) Error() string {
    return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

// Использование
func Do() error {
    return MyError{Code: 404, Message: "not found"}
}
🗓️ ПЛАН ПОВТОРЕНИЯ
Сегодня (вечер):

TCP vs UDP — прочитать, запомнить примеры

TLS handshake — пересказать по шагам

sync пакет — написать примеры с WaitGroup, Mutex, Once

Завтра (утро):

Каналы — close, range, select

Контекст — WithCancel, WithTimeout, WithValue

Ошибки — кастомная ошибка, errors.Is, errors.As

Перед следующим собеседованием:

Прогнать всё ещё раз (15 минут)

GoLand — отключить автодополнение

Подготовить легенду (опыт, outbox, архитектура)

