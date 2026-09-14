ListUrls := []string{
	"https://my.mts-link.ru/120194511/16181950806/record-new/15271534466", // 18 марта разбор задач (хз че там)
	"https://my.mts-link.ru/120194511/15666290009/record-new/14785495405", // 12 марта разбор задач (Видел, в начале прикол)
	"https://my.mts-link.ru/120194511/14413515235/record-new/13588589288", // 04 марта разбор задач с Сергеем, (они ниже)
}

"https://my.mts-link.ru/120194511/14413515235/record-new/13588589288", // 04 марта разбор задач с Сергеем, (они ниже)

Нужно вывести:
Вывести уникальные комбинации пользователя и id товара(SKU) для всех покупок, 
совершенных пользователями до того, как их забанили. 
Отсортировать сначала по имени пользователя, потом по SKU
Найти пользователей, которые совершили покупок на сумму больше 5000р. 
Вывести их имена в формате id пользователя | имя | фамилия | сумма покупок

-- user
id | firstname | lastname | birth
1  | Ivan      | Petrov   | 1996-05-01
2  | Anna      | Petrova  | 1999-06-01      
3  | Anna      | Petrova  | 1990-10-02
      
-- purchase
sku| price | user_id | date
1  | 5500  | 1       | 2021-02-15
1  | 5700  | 1       | 2021-01-15
2  | 4000  | 1       | 2021-02-14
3  | 8000  | 2       | 2021-03-01
4  | 400   | 2       | 2021-03-02


-- ban_list
user_id | date_from   
1       | 2021-03-08

-- Часть 1: уникальные комбинации пользователь + SKU для покупок до бана
SELECT DISTINCT
    u.id,
    u.firstname,
    u.lastname,
    p.sku
FROM user u
JOIN purchase p 
	ON u.id = p.user_id
LEFT JOIN ban_list b 
	ON u.id = b.user_id
WHERE b.user_id IS NULL OR p.date < b.date_from
ORDER BY u.firstname, u.lastname, p.sku;

-- Часть 2: пользователи, которые совершили покупок на сумму > 5000 (до бана)
SELECT 
    u.id,
    u.firstname,
    u.lastname,
    SUM(p.price) AS total_amount
FROM user u
JOIN purchase p 
	ON u.id = p.user_id
LEFT JOIN ban_list b 
	ON u.id = b.user_id
WHERE b.user_id IS NULL OR p.date < b.date_from
GROUP BY u.id, u.firstname, u.lastname
HAVING SUM(p.price) > 5000
ORDER BY u.id;

---
id, sku, name, lastname, price
1   1
1   1
1   1
2   1
2   1

Ответ:
1) Здесь мы проверяем что человек умеет в джоины, distinct, where, order
2) Здесь мы проверяем что человек умеет в HAVING, знает, чем having отличается от WHERE

-----

// Требуется реализовать функцию uniqRandn, которая генерирует слайс длины n УНИКАЛЬНЫХ, рандомных чисел.

import (
    "fmt"
    "math/rand/v2"
    "time"
)

func main() {
    fmt.Println(uniqRandn(11))
}

func uniqRandn(n int) []int {
    if n <= 0 {
	return []int{}
    }

    mapU :=make(map[int]struct{}, n)
    sl := make([]int, 0, n)

    for len(mapU)<n {
        rnd := rand.IntN(1000)
        if _, ok := mapU[rnd]; !ok {
            mapU[rnd] = struct{}{}
            sl = append(sl, rnd)
        }
    }

    return sl
}

---------------

На примере создания заказа. 
Есть запрос на сервис и есть ответ, между этими двумя действиями мы сладываем в аналитику товары которые заказали 
(например для подсчета популярности товаров). Сервис аналитики переодически работает медленно или вовсе таймаутит
, и мы не успеваем ответить, теряем заказы.

 Что делать, что бы перестать терять заказы, и деньги соответсвенно?

+-------+         +---------------+    +-------------------+
| User  |         | OrderService  |    | AnalyticsService  |
+-------+         +---------------+    +-------------------+
    |                     |                      |
    | CreateOrder         |                      |
    |-------------------->|                      |
    |                     |                      |
    |                     | TrackOrder           |
    |                     |--------------------->|
    |                     |                      |
    |                     |                      | Too long action
    |                     |                      |----------------
    |                     |                      |               |
    |                     |                      |<---------------
    |                     |                      |
    |                     |                      |
    |                     |<---------------------|
    |                     |                      |
    |                     |                      |
    |<--------------------|                      |
    |                     |                      |


user -> OrderService -> outbox -> Kafka,RabbitMQ,NATS(topic) <- AnalyticsService
Отдел аналитики жалуется на высокий рпс от нас когда они под нагрузкой мы их сервис кладем. Как решить проблему?
Ответ:
Нужно придумать экспоненциальное повышение или понижение нагрузки в зависимости от состояния аналитики
Circuit Breaker

17 грейд джун хз

Кандидат предлагает вынести запрос в аналитику в асинхронную рутину.

18 грейд миддл 300к до налогов

Предложил один из вариантов вариантов асинхронной отложенной обработки, например: придумать как распаралелить заказ и аналитику, брокер сообщений

19 грейд синьк 450к до налогов

Кандидат предложил OUTBOX. Понимает и может объяснить:

как организовать асинхронное вычитывание журнала с блокировкой:
на уровне PSQL;
через сторонний ресурс;
чем обеспечить гарантию отправки данных во внешние сервисы.
Также соискатель может обозначить минусы выноса отправки в "асинхронное" решение:

необходимость администрировать эти рутины, особенно при выключении сервиса;
проблема с ретраями и нагрузкой внутри рутин;
проблема с "гарантией отправки".





Это классическая задача про асинхронную обработку, гарантии доставки и отказоустойчивость. Разберем по уровням.

Проблема
text
User → OrderService → AnalyticsService
                       ↓
                   медленно/таймаутит
                       ↓
              OrderService ждет ответа
                       ↓
                 User получает таймаут
                       ↓
                 Заказ теряется ❌
Ключевая проблема: синхронный вызов медленного/ненадежного внешнего сервиса внутри критического пути.

Решения по уровням
17 грейд (Junior) — асинхронная рутина
go
go func() {
    analyticsService.TrackOrder(order)
}()
Что говорит: «Вынесу запрос в аналитику в отдельную горутину, чтобы не ждать ответа».

Проблемы:

При падении сервиса данные теряются (нет персистентности)

Нет гарантии доставки

При рестарте сервиса данные не восстановятся

Нет контроля нагрузки

18 грейд (Middle) — брокер сообщений
go
// OrderService отправляет сообщение в Kafka/RabbitMQ/NATS
broker.Publish("analytics.track", order)

// AnalyticsService подписывается и обрабатывает в своем темпе
broker.Subscribe("analytics.track", func(msg Message) {
    processAnalytics(msg)
})
Что говорит: «Использую брокер сообщений для асинхронной обработки».

Плюсы:

Разделение сервисов

AnalyticsService может обрабатывать в своем темпе

Брокер буферизирует нагрузку

Проблемы:

Если OrderService упал до отправки в брокер — данные потеряны

Нет гарантии "хотя бы раз" (at-least-once) без дополнительных механизмов

Нужно администрировать брокер

19 грейд (Senior) — паттерн Outbox
text
┌─────────────────────────────────────────────────────────────────┐
│                      OrderService                               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐ │
│  │   Order     │    │   Outbox    │    │   Outbox Poller     │ │
│  │   Creation  │───▶│   Table     │◀───│   (background)      │ │
│  └─────────────┘    └─────────────┘    └──────────┬──────────┘ │
│         │                                          │           │
│         └──────────────┼───────────────────────────┘           │
└────────────────────────┼───────────────────────────────────────┘
                         │
                         ▼
                  ┌─────────────┐
                  │   Kafka/    │
                  │   RabbitMQ  │
                  └─────────────┘
                         │
                         ▼
                  ┌─────────────────┐
                  │ AnalyticsService│
                  └─────────────────┘
Реализация Outbox:
sql
-- Таблица outbox
CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    processed_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' -- pending, processing, completed, failed
);

-- Индекс для быстрого поиска необработанных
CREATE INDEX idx_outbox_pending ON outbox (created_at) WHERE processed_at IS NULL;
go
// OrderService — в той же транзакции, что и создание заказа
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    tx := s.db.Begin()
    
    // 1. Создаем заказ
    order, err := s.createOrderInTx(tx, req)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    // 2. Пишем в outbox в ТОЙ ЖЕ транзакции
    err = s.createOutboxEventInTx(tx, "order_created", order)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    
    // 3. Возвращаем ответ пользователю сразу (не ждем аналитику!)
    return nil
}

// Фоновый процесс — вычитывает outbox и отправляет в брокер
func (p *OutboxPoller) Run() {
    for {
        events := p.getPendingEvents(100) // SELECT ... FOR UPDATE SKIP LOCKED
        
        for _, event := range events {
            // Отправляем в брокер с ретраями
            err := p.sendToBroker(event)
            if err != nil {
                p.markAsFailed(event.ID, err)
                continue
            }
            
            // Помечаем как обработанное
            p.markAsProcessed(event.ID)
        }
        
        time.Sleep(time.Second)
    }
}

// SELECT ... FOR UPDATE SKIP LOCKED — для конкурентной обработки
func (p *OutboxPoller) getPendingEvents(limit int) []OutboxEvent {
    rows, _ := p.db.Query(`
        SELECT id, event_type, payload 
        FROM outbox 
        WHERE processed_at IS NULL 
        ORDER BY created_at 
        LIMIT $1 
        FOR UPDATE SKIP LOCKED
    `, limit)
    // ...
}
Что должен знать Senior (19 грейд)
1. Гарантии доставки
Гарантия	Реализация
At-least-once	Outbox + retry + идемпотентность на стороне потребителя
Exactly-once	Идемпотентность + транзакционный брокер (Kafka с idempotent producer)
Ordering	Порядок гарантируется через ORDER BY created_at + один consumer
2. Варианты реализации вычитывания outbox
sql
-- Вариант 1: FOR UPDATE SKIP LOCKED (PostgreSQL)
SELECT * FROM outbox 
WHERE processed_at IS NULL 
ORDER BY id 
LIMIT 100 
FOR UPDATE SKIP LOCKED;
Не блокирует строки, обрабатываемые другими воркерами

Позволяет масштабировать обработку горизонтально

sql
-- Вариант 2: Логическое удаление с интервалами
UPDATE outbox SET status = 'processing' 
WHERE id IN (
    SELECT id FROM outbox 
    WHERE status = 'pending' 
    LIMIT 100 
    FOR UPDATE SKIP LOCKED
) RETURNING *;
3. Минусы асинхронного подхода
Минус	Решение
Дополнительная инфраструктура	Нужно администрировать брокер, poller, мониторинг
Проблемы при рестарте	Состояние outbox должно восстанавливаться, poller должен корректно стартовать
Ретраи и нагрузка	Нужен exponential backoff, ограничение количества попыток, dead letter queue
Гарантии доставки	Требует идемпотентности на стороне AnalyticsService
Задержки	Данные попадают в аналитику не мгновенно (обычно приемлемо)
Дубликаты	Возможны при ретраях → нужна идемпотентность
4. Альтернативные подходы
Подход	Плюсы	Минусы
Change Data Capture (CDC)	Автоматическое вычитывание из WAL (Debezium)	Сложнее настройка
Circuit Breaker	Защита от перегрузки, fallback	Все еще синхронный вызов
Bulkhead	Изоляция пулов потоков	Не решает проблему потери данных
Ответ для собеседования (Senior)
*«Для решения проблемы потери заказов при таймаутах аналитики я бы использовал паттерн Outbox. При создании заказа в одной транзакции сохраняю событие в outbox-таблицу. Отвечаю пользователю сразу, не дожидаясь аналитики. Фоновый процесс вычитывает outbox с блокировкой FOR UPDATE SKIP LOCKED и отправляет события в Kafka/RabbitMQ. AnalyticsService подписывается на топики и обрабатывает в своем темпе. Это дает гарантию at-least-once доставки, изолирует заказы от проблем аналитики и позволяет масштабировать обработку.»

«Важные детали: нужно обеспечить идемпотентность на стороне аналитики, чтобы ретраи не дублировали данные; настроить exponential backoff для ретраев; добавить мониторинг очереди outbox. Минусы — дополнительная инфраструктура и задержка попадания данных в аналитику, что обычно приемлемо для аналитических задач.»

Итоговая таблица грейдов
Грейд	Решение	Зарплата (до налогов)
17 (Junior)	go func() или time.AfterFunc	~200-250к
18 (Middle)	Брокер сообщений (Kafka/RabbitMQ)	~300к
19 (Senior)	Outbox + гарантии доставки + идемпотентность	~450к+
Ключевое отличие Senior: понимает, что просто брокер не решает проблему потери данных при падении сервиса до отправки. Outbox обеспечивает гарантию через атомарность с бизнес-транзакцией.















-----------

Построить оптимальный индекс для
SELECT * FROM employee 
WHERE sex = 'm' AND salary > 300000 AND age = 20 
ORDER BY created_at




sex - хорошо селективный 
create index (sex, salary, age, created_at)

18 грейд middle

Для PostgreSQL: 
Наиболее оптимальными будут вариации составных индексов в порядке ухудшения селективности. 
!!!!!!!!!!Такие индексы, как известно, работают эффективно до первого неравенства.!!!!! !!!!!!!!!!

Наилучшая реализация зависит от того, какие кандидату будут даны уточнения относительно данных в таблице. В
 зависимости от характера данных это могут быть следующие вариации: 

В случае, если условие sex = 'm' дает хорошую селективность

CREATE INDEX ... ON employee (sex, age, salary)
В случае, если мы не знаем о характере данных и только ограничиваемся предположением о кардинальности

SQL
CREATE INDEX ... ON employee (age, salary)
На деле может быть не сильно эффективнее чем предыдущая версия, но демонстрирует что кандидат знает, как можно оптимизировать сортировку

SQL
CREATE INDEX CONCURRENTLY ON employee (age, salary, created_at)
- Знает про составные индексы 
- Знает про EXPLAIN ANALYZE 
- Знает про селективность. Учитывает характер данных при построении индекса 

19 грейд senior
18+:
- Умеет создать индекс без блокировки таблицы CONCURRENTLY 
- Может подробнее рассказать про EXPLAIN ANALYZE, привести несколько кейсов, что полезное для оптимизации запросов там есть 
- Знает, что можно создать индекс только для части данных 
- Знает, что существует такая вещь как покрывающий индекс 



Опционально: 
- Знает недостатки создания индекса через CONCURRENTLY 
- Знает, что PostgreSQL может использовать несколько индексов совместно в одном запросе (Bitmap Scan) 
- Знает, что делает команда CLUSTER и для чего ее применяют 
- Знает, что за свойство fillfactor и зачем оно нужно










Отличная задача на знание индексов в PostgreSQL. Давай разберем по уровням.

Условия запроса
sql
SELECT * FROM employee 
WHERE sex = 'm' AND salary > 300000 AND age = 20 
ORDER BY created_at
Селективность полей:

sex = 'm' — высокая селективность (примерно 50% записей)

age = 20 — средняя/высокая (зависит от данных)

salary > 300000 — неравенство, низкая селективность

ORDER BY created_at — сортировка

Базовый ответ (middle, 18 грейд)
sql
CREATE INDEX idx_employee_sex_age_salary_created_at 
ON employee (sex, age, salary, created_at);
Почему так:

sex — первым, потому что условие равенства и высокая селективность

age — вторым, тоже равенство

salary — третьим, неравенство (индекс работает до первого неравенства)

created_at — последним, для сортировки

Важное замечание про неравенства:
В PostgreSQL составной индекс работает эффективно только до первого поля с неравенством. После него индекс используется только для фильтрации, но не для сортировки.

Улучшенный вариант (senior, 19 грейд)
sql
-- 1. Создаем индекс без блокировки таблицы
CREATE INDEX CONCURRENTLY idx_employee_cover 
ON employee (sex, age, salary, created_at);

-- 2. Покрывающий индекс (если часто выбираем только эти поля)
CREATE INDEX CONCURRENTLY idx_employee_cover_include 
ON employee (sex, age, salary) 
INCLUDE (created_at);

-- 3. Частичный индекс (если 'm' — самый частый запрос)
CREATE INDEX CONCURRENTLY idx_employee_male 
ON employee (age, salary, created_at) 
WHERE sex = 'm';
Что нужно знать на собеседовании (по уровням)
18 грейд (middle):
Знание	Комментарий
Составные индексы	Порядок полей важен: сначала равенства, потом неравенства
Индекс до первого неравенства	После > или < индекс не используется для сортировки
EXPLAIN ANALYZE	Нужно показать, как анализировать план запроса
Селективность	Поля с высокой селективностью должны идти первыми
19 грейд (senior):
Знание	Комментарий
CREATE INDEX CONCURRENTLY	Создает индекс без блокировки таблицы на запись (но дольше)
Покрывающий индекс (INCLUDE)	Хранит дополнительные поля в листьях, позволяет избежать обращения к таблице
Частичный индекс (WHERE)	Индекс только на подмножество строк
Bitmap Scan	PostgreSQL может комбинировать несколько индексов
CLUSTER	Физически переупорядочивает таблицу по индексу (ускоряет диапазонные запросы)
fillfactor	Заполненность страниц индекса (для частых UPDATE)
Оптимальный ответ для senior:
sql
-- Вариант 1: Если sex высокоселективный
CREATE INDEX CONCURRENTLY idx_employee_sex_age_salary 
ON employee (sex, age, salary);

-- Вариант 2: Если нужна сортировка без сортировки
CREATE INDEX CONCURRENTLY idx_employee_age_salary_created 
ON employee (age, salary, created_at);

-- Вариант 3: Покрывающий индекс (если не нужны все поля)
CREATE INDEX CONCURRENTLY idx_employee_cover 
ON employee (sex, age, salary) 
INCLUDE (created_at);

-- Вариант 4: Частичный индекс для мужчин
CREATE INDEX CONCURRENTLY idx_employee_male_active 
ON employee (age, salary, created_at) 
WHERE sex = 'm';
Почему несколько вариантов:

«Оптимальный индекс зависит от реального распределения данных. Я бы проанализировал существующие запросы через pg_stat_statements, посмотрел план через EXPLAIN ANALYZE и выбрал наиболее эффективный вариант. В production создавал бы индекс через CONCURRENTLY, чтобы не блокировать запись.»

Дополнительные вопросы, которые могут задать:
1. Недостатки CONCURRENTLY
Дольше создается (может идти часами)

Может не успеть завершиться в рамках транзакции

Требует больше ресурсов

При ошибке оставляет "невалидный" индекс

2. Что такое Bitmap Scan
PostgreSQL может использовать несколько индексов, объединяя их битовые карты:

sql
-- Может использовать два индекса: на age и на salary
SELECT * FROM employee WHERE age = 20 OR salary > 300000;
3. Зачем нужен CLUSTER
sql
CLUSTER employee USING idx_employee_age;
Физически переупорядочивает таблицу по индексу. Ускоряет диапазонные запросы (BETWEEN, >, <), но требует периодического обновления.

4. fillfactor
sql
CREATE INDEX CONCURRENTLY idx_employee_age 
ON employee (age) 
WITH (fillfactor = 80);
Оставляет место для будущих UPDATE, чтобы избежать раздувания индекса. По умолчанию 90% для индексов.

Итоговая оценка
Грейд	Что нужно показать
18	Составной индекс с порядком полей (равенства → неравенства → сортировка)
19	CONCURRENTLY, покрывающие индексы, частичные индексы, EXPLAIN ANALYZE
20+	CLUSTER, fillfactor, Bitmap Scan, работа с pg_stat_statements, анализ планов
Твой ответ про CREATE INDEX (sex, salary, age, created_at) — не совсем оптимальный, потому что:

salary — неравенство, после него age уже не используется для фильтрации эффективно

age лучше поставить перед salary, чтобы оба равенства были в начале

Правильный порядок: (sex, age, salary) или (sex, age, salary, created_at)














--------

Является ли слайс монотонным?
Монотонная функция - функция одной переменной, определённая на некотором подмножестве действительных чисел, 
которая либо везде (на области своего определения) не убывает, либо везде не возрастает.

    {1,7} - true
    {1,1} - true
    {3,3,1} - true
    {9,5,1} - true
    {23,5,23} - false

[i-1]

func isMonotonic(in []int) bool {
    if len(in) <= 2 {
        return true
    }
    
    inc := true
    dec := true
    
    for i := 1; i < len(in); i++ {
        if in[i] < in[i-1] {
            inc = false // нарушено неубывание
        }
        if in[i] > in[i-1] {
            dec = false // нарушено невозрастание
        }
        // если оба флага уже false, можно выйти раньше
        if !inc && !dec {
            return false
        }
    }
    
    return increasing || decreasing
}


Компактнее:

func isMonotonic(in []int) bool {
    inc, dec := true, true
    
    for i := 1; i < len(in); i++ {
        inc = inc && in[i] >= in[i-1]
        dec = dec && in[i] <= in[i-1]
    }
    
    return inc || dec
}


---------

### Чемпионат по шагам

Недавно мы устроили чемпионат по шагам. И вот настало время подводить итоги! 
Необходимо определить userids участников, которые прошли наибольшее количество шагов steps за все дни, не пропустив ни одного дня соревнований.

// Пример 1  ввод
statistics = [
    [{ userId: 1, steps: 1000 }, { userId: 2, steps: 1500 }],
    [{ userId: 2, steps: 1000 }],
]
// вывод
champions = { userIds: [2], steps: 2500 }

// Пример 2
statistics = [
    [{ userId: 1, steps: 2000 }, { userId: 2, steps: 1500 }],
    [{ userId: 2, steps: 4000 }, { userId: 1, steps: 3500 }],
]
// вывод
champions = { userIds: [1, 2], steps: 5500 }

type Statistic struct {
    UserID int
    Steps int
}

type Result struct{
    UserIDs []int
    Steps int
}

func getChampions(statistics [][]Statistic) Result {
    // ...
}

------
func getChampions(statistics [][]Statistic) Result {
    // Карта: userID -> общее количество шагов
    stepsSum := make(map[int]int)
    // Карта: userID -> количество дней участия
    daysPresent := make(map[int]int)
    totalDays := len(statistics)

    // 1. Проходим по всем дням
    for _, day := range statistics {
        // Множество пользователей, которые были в этот день
        today := make(map[int]bool)
        
        for _, stat := range day {
            stepsSum[stat.UserID] += stat.Steps
            today[stat.UserID] = true
        }
        
        // Увеличиваем счетчик дней для каждого, кто был сегодня
        for userID := range today {
            daysPresent[userID]++
        }
    }

    // 2. Находим максимальную сумму среди тех, кто был все дни
    maxSteps := 0
    for userID, sum := range stepsSum {
        if daysPresent[userID] == totalDays && sum > maxSteps {
            maxSteps = sum
        }
    }

    // 3. Собираем всех чемпионов
    champions := []int{}
    for userID, sum := range stepsSum {
        if daysPresent[userID] == totalDays && sum == maxSteps {
            champions = append(champions, userID)
        }
    }

    return Result{
        UserIDs: champions,
        Steps:   maxSteps,
    }
}

------

Логика (запоминается легко):
Считаем сумму шагов → stepsSum[userID] += steps

Считаем дни участия → через временное множество today, чтобы не дублировать одного пользователя дважды в один день

Находим максимум → только среди тех, у кого daysPresent == totalDays

Собираем чемпионов → кто набрал maxSteps

Почему этот вариант хорош для собеседования:

Почему			Что дает
Минимум структур	Только две мапы и одно множество на день
Один проход		За один цикл по дням собираем все данные
Четкие названия		stepsSum, daysPresent, today — понятно без комментариев
Пошаговая логика	1 → 2 → 3 → 4, легко проговаривать вслух
Краевые случаи		Если нет участников во все дни — maxSteps останется 0, champions пустой

Алгоритм:
	Собираем всех участников — нужно знать, кто вообще участвовал
	Считаем сумму шагов для каждого участника
	Отслеживаем, в каких днях был участник — чтобы отсеять тех, кто пропустил хотя бы один день
	Находим максимальную сумму шагов среди отфильтрованных участников
	Собираем всех, у кого сумма равна максимуму

Что сказать интервьюеру:
"Сначала собираю для каждого пользователя сумму шагов и количество дней, в которых он участвовал"
"Потом нахожу максимальную сумму среди тех, кто не пропустил ни одного дня"
"Потом собираю всех, кто набрал эту сумму"
"Сложность O(N * M) по времени, O(K) по памяти, где K — количество уникальных пользователей"

=================================



XS - в отчеты добавить новую колонку (неделька) epic - проект, небольшая задачка
а какая новая?


1 раз поправить поле в контрактах - 1 день (10 минут-час)
сделать логику в ручке - 4 дня
создать топик и отправлять туда сообщения 4 дня







