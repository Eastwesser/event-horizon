Блок 1: Примеры кода на Go (Асинхронность и Мьютексы)
Изображение 1: Последовательный запуск

go
1 package main
2
3 import (
4     "fmt"
5     "time"
6 )
7
8 var value int
9
10 func main() {
11     for range 1000 {
12         process()
13     }
14     fmt.Println(value)
15 }
16
17 func process() {
18     time.Sleep(1 * time.Second)
19     value++
20     fmt.Println(value)
21 }
Изображение 2: Параллельный запуск с WaitGroup и Mutex

go
4     "fmt"
5     "time"
6     "sync"
7 )
8
9 var value int
10 var mu sync.Mutex
11
12 func main() {
13     var wg sync.WaitGroup
14     for range 10 {
15         wg.Add(1)
16         go func(){
17             defer wg.Done()
18             process()
19         }()
20     }
21
22     wg.Wait()
23     fmt.Println(value)
24 }
25
26 func process() {
27     time.Sleep(1 * time.Second)
28     mu.Lock()
29     defer mu.Unlock()
30     value++
31     fmt.Println(value)
32 }
Блок 2: Реализация монитора (Monitor)
Изображение 3: Базовая структура

go
3 import (
4     "sync"
5     "time"
6 )
7
8 type monitor struct {
9     wg      sync.WaitGroup
10    ticker *time.Ticker
11    done    chan struct{}
12 }
13
14 func (m *monitor) Start() {
15     m.wg.Add(1)
16     go func() {
17         defer m.wg.Done()
18         for {
19             foo()
20             select {
21             case <-m.ticker.C:
22             case <-m.done:
23                 break
24             }
25         }
26     }()
27 }
28
29 func (m *monitor) Stop() {
30     close(m.done)
31     m.wg.Wait()
32     m.ticker.Stop()
33 }
Изображение 4: Монитор с комментариями

go
8 type monitor struct {
9     wg      sync.WaitGroup //
10    ticker *time.Ticker
11    done    chan struct{}
12 }
13
14 Start()
15
16 Stop()
17
18 func (m *monitor) Start() {
19     m.wg.Add(1)
20     go func() {
21         defer m.wg.Done()
22         for {
23             foo()
24             select {
25             case <-m.ticker.C:
26             case <-m.done:
27                 break
28             }
29         }
30     }()
31 }
32
33 // STOP func родительская функция и там бы обрабатывал закрытие
34 func (m *monitor) Stop() {
35     close(m.done) //
36     m.wg.Wait()
37     m.ticker.Stop()
38 }
Блок 3: Архитектура и Базы Данных (Партиционирование)
Изображение 5: Вставка в партиционирование

text
1. 200_000 записей как сделать вставку в партиционирование
джоба
2
3
4
5 package main
6
7 import (
8     "sync"
9     "time"
10 )
11
12 type monitor struct {
13     wg      sync.WaitGroup //
14     ticker *time.Ticker
15     done    chan struct{}
16 }
17
18 func (m *monitor) Start() {
19     m.wg.Add(1)
20     go func() {
21         defer m.wg.Done()
22         for {
23             foo() //
24             select {
25             case <-m.ticker.C:
26             case <-m.done:
27                 return
28             }
29         }
30     }()
31 }
Изображение 6: Агрегация и партиции

text
1. 200_000 записей как сделать агрегацию?
таблицу. триггеры -> таблицу
джоба в конце будет собирать данные.
скейл(увеличиваем кол-во) скейлим до 10 джоб.
2
3 1 джоба = 1 партиции
4 ...
5 10 джоба = 10 партиция
6
7 id | sku | bid | user_id | part_key |
8 1.  1.    1.    1.        1
9 0.  1.    1.    1.        2
10 2.  1.    1.    1.        3
11 1.  1.    1.    1.        4
12 1.  1.    1.    1.        1
13 4.  1.    1.    1.        1
14 1.  1.    1.    1.        2
15 1.  1.    1.    1.        3
16
17 ---
18
19 ring buffer
20 3 джоба = с одного и того же места
Блок 4: Задача на написание in-memory кэша
Изображение 7: Условие задачи

go
83 ----------
84 /*
85 Необходимо написать in-memory кэш.
86
87 Условия:
88 1. У кэша должен быть TTL:
89    - Должна быть возможность задавать кастомный TTL для каждого элемента.
90 2. Написать функции для работы с кэшем:
91    - Получение пользователя по его ID (при чтении не обновляем TTL).
92    - Добавление пользователя в кэш.
93    - Удаление пользователя из кэша по его ID.
94 3. Написать тестовые сценарии для проверки работы кэша.
95 */
96
97 type User struct {
98     ID   int64
99     Name string
100    // more fields...
101 }
Блок 5: Алгоритмическая задача (Чемпионат по шагам)
Изображение 8: Условие и примеры

text
1 Чемпионат по шагам
2
3 Недавно мы устроили чемпионат по шагам. И вот настало время подводить итоги!
4 Необходимо определить userIds участников, которые прошли наибольшее количество шагов за все дни,
5 не пропустив ни одного дня соревнований.
6
7 // Пример 1 ввод
8 statistics = [
9     [{ userId: 1, steps: 1000 }, { userId: 2, steps: 1500 }],
10    [{ userId: 2, steps: 1000 }],
11 ]
12
13 // вывод
14 champions = { userIds: [2], steps: 2500 }
15
16 // Пример 2
17 statistics = [
18    [{ userId: 1, steps: 2000 }, { userId: 2, steps: 1500 }],
19    [{ userId: 2, steps: 4000 }, { userId: 1, steps: 3500 }],
20 ]
21 // вывод
22 champions = { userIds: [1, 2], steps: 5500 }
23
24 type Statistic struct {
25     UserID int
26     Steps  int
27 }
28
29 type User struct {
30     Id    int
31     Steps int
32     Days  int
33 }
34
35 type Result struct {
36     UserIDs []int
37     Steps   int
38 }
Изображение 9: Решение (часть 1)

go
43 Граничные случаи:
44 == 1day
45 >2 day
46 user = 0
47 user > 0
48 champions = 0
49 */
50 func getChampions(statistics [][]Statistic) Result {
51     result := make(map[int]User, 1)
52     maxDays := len(statistics)
53     // maxSteps = 0 //
54     for indexDay, dayStats := range statistics {
55         for _, stats := range dayStats {
56             if indexDay == 0 { // 1 день
57                 result[stats.userId] = User{Steps: stats.Steps, Days: 1}
58                 continue//
59             }
60             value, ok := result[stats.userId]
61             if !ok {
62                 continue
63             }
64             value.Days++
65             value.Steps += stats.Steps
66             result[stats.userId] = value
67         }
68     }
69     maxSteps := 0
70     var winnerID int
71     for key,value := range result { // O(n)
72         if maxSteps < value.Steps && value.Days == maxDays {
73             maxSteps = value.Steps
74             winnerID = key
75         }
76     }
77     return result[winnerID]
78 }
Изображение 10: Решение (часть 2) и заметки по soft skills

go
70     maxSteps := 0
71     var winnerID int
72     for key,value := range result {
73         if maxSteps < value.Steps && value.Days == maxDays {
74             maxSteps = value.Steps
75             winnerID = key
76         }
77     }
78     return result[winnerID]
79 }
80
81
82
83
84 product owner = 100% прошел:
85 1. кем ты видишь себя через столько то лет
86
87 lead
88
89
90 Базовые вопросы которые должен спросить:
91 1. сколько кол-во человек в команде
92 2. чем я буду заниматься?
93 3. Как дорости до синьер позиции? -> тебе надо будет вести самостоятельно эпики(большая продуктовая задача)
94 4. Что на счет дежурство?
95 5.
96
97
98 Mвидео warehouse:
99 Эпик от продукта, что они внедряют новый склад.
100 Декомпозировать эпик: на мелкий задачки.
101
Изображение 11: Вопросы на собеседовании (Middle+)

text
83 Вопросы которые могут задать:
84 - Вел ли ты эпики? Если да, то какие и что по времени <- middle+
85 1. Чем занимался?
86 2. Какие задачи выполнял?
87 3. Если тебе что-то не понравилось в процессах команде, как влиял на нее? Например:
88 4.
89
90 5 задач в неделю. 7 задач
91
92 5 задач в неделю = 20 задач в неделю
93
94
95 88 4.
96
97 Базовые вопросы которые должен спросить:
98 1. сколько кол-во человек в команде
99 2. чем я буду заниматься?
100 3. Как дорости до синьер позиции? -> тебе надо будет вести самостоятельно эпики(большая продуктовая задача)
101 4. Что на счет дежурство?
102 5. Что по flow-выкатки в прод?
103
104
105 Mвидео warehouse:
106 Эпик от продукта, что они внедряют новый склад.
107 Декомпозировать эпик: на мелкий задачки.