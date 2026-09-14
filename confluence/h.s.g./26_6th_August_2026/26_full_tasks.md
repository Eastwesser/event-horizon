Блок 1: Задача "Произведение всех элементов кроме i-го"
Изображение 1:

go
1 ---
2 ПЕРВАЯ
3 // Есть массив целых чисел, нужно написать функцию которая вернет массив такого же размера,
4 // где на каждой i-ой позиции будет произведение всех элементов кроме i-го
5 // [1, 2, 3] => [2*3, 1*3, 1*2] => [6, 3, 2]
6 ---
Изображение 2: Решение

go
1 ---
2 ПЕРВАЯ
3 // Есть массив целых чисел, нужно написать функцию которая вернет массив такого же размера,
4 // где на каждой i-ой позиции будет произведение всех элементов кроме i-го
5 // [1, 2, 3] => [2*3, 1*3, 1*2] => [6, 3, 2]
6
7 [3, 5, 2] ---> [10, 6, 15]
8
9 // Делить нельзя
10 ---
11
12 func Foo(nums []int) []int {
13     n := len(nums)
14     result := make([]int, n) // 1*1 1*2 2*3 // 1, 1, 2
15
16     // [1, 2, 3]
17     left := 1 // i:0 = 1, i:1 = 1, i:2 = 2 == result
18     for i := 0; i < n; i++ {
19         result[i] = left
20         left *= nums[i]
21     }
22     // 2 1 1
23     right := 1 // i:2 =
24     for i := n - 1; i >= 0; i-- {
25         result[i] *= right
26         right *= nums[i]
27     }
28
29     return result
30 }
Блок 2: Структура проекта и задача Crawler
Изображение 3: Структура проекта

text
go-interview-problems ~/GolandProjects/go-inte
> 01-first-successful-key-lookup
> 02-equivalent-binary-trees
> 03-web-crawler
> 04-non-blocking-cache
> 05-costly-connections-with-unsafe-storage
> 06-rate-limiter
> 07-ttl-cache
> 08-request-with-failover
> 09-merge-channels
> 10-concurrent-queue
> 11-concurrent-queue-ii
> 12-concurrent-queue-iii
v 13-rate-tracker
    > solution
        M README.md
        task.go
        task_test.go
    > playground
        have-fun.go
M CONTRIBUTING.md
go.mod
LICENSE
M README.md
Изображение 4: Условие задачи Crawler

go
39 // Хотим делать параллельно но не больше k запросов одновременно
40 ---
41
42 worker pool
43 semaphore
44 // map[partition_id][]string
45 var urls = []string{
46     "https://www.lamoda.ru/p/mp002xw0lvkd/clothes-tomollyfromjames-plate/",
47     "https://www.lamoda.ru/p/mp002xw14uf2/clothes-tomollyfromjames-plate/",
48     "https://www.lamoda.ru/p/rtladr746901/clothes-iceberg-plate/",
49     "https://www.lamoda.ru/p/mp002xw18h9d/clothes-victoriaveisbrut-plate/",
50     "https://www.lamoda.ru/p/mp002xw004x4/clothes-clanvi-plate/",
51     "https://www.lamoda.ru/p/mp002xw0zfxy/clothes-glvr-plate/",
52     "https://www.lamoda.ru/p/mp002xw0slmg/clothes-snezhnayakoroleva-plate-kozhanoe/",
53     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
54     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
55     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
56     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
57     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
58     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
59     "https://www.lamoda.ru/p/mp002xw132c3/clothes-auranna-plate/",
60     ......
61 }
62
63
64 func crawl(urls []string, k int) []int {
65     for range
66 }
Изображение 5: Реализация Crawler (черновик)

go
60     ......
61 }
62
63 type Task struct {
64     idx int,
65     url string,
66 }
67
68 // 1.25
69 func crawl(urls []string, k int) []int {
70     in := make(chan Task, k)
71     for range k {
72         go func() {
73             task := <- in
74             response := http.Get(task.url)
75             res[task.idx] = response.StatusCode
76         }()
77     }
78
79     n := len(urls)
80     var res := make([]int, n)
81     var wg sync.WaitGroup
82
83     wg.Add(n)
84     for i, url := range urls {
85         go func() {
86             defer wg.Done()
87             response := http.Get(url)
88
89             res[i] = response.StatusCode
90         }()
91     }
92
93     wg.Wait()
Изображение 6: Продолжение реализации

go
95     close(in)
96     wg.Wait()
97 }
98
99 func main() {
100     result := crawl(urls, 5)
101     fmt.Println("All done")
102 }
103
104
105
106
107 ==========================================
108
109 if ban {
110     accessLevel &^= users.MonetizationAccess
111 } else {
112     accessLevel |= users.MonetizationAccess
113 }
114
115
116 ==========================================
117
118
119
120
121
122
123
124
125 func(r *Repository) WithdrawBalance(userId int32, amount int32) {
126     tx := BeginTransaction("ISOLATION LEVEL READ COMMITTED")
127     currentUserBalance := tx.exec("SELECT balance FROM users WHERE user_id = $1", userId)
128     if oldBalance >= amount {
Блок 3: Работа с битовыми масками
Изображение 7:

go
106
107 ==========================================
108
109 if ban {
110     accessLevel &^= users.MonetizationAccess // хор с отрицанием
111 } else {
112     accessLevel |= users.MonetizationAccess // побитовое или
113 }
114
115
116 ==========================================
117
118
Блок 4: Работа с транзакциями БД (эволюция решения)
Изображение 8: Первый вариант

go
123
124 func(r *Repository) WithdrawBalance(userId int64, amount float64) {
125     tx := BeginTransaction("ISOLATION LEVEL READ COMMITTED")
126     currentUserBalance := tx.exec("SELECT balance FROM users WHERE user_id = $1", userId)
127     if oldBalance >= amount {
128         tx.exec("UPDATE users SET balance = balance - $2 WHERE user_id = $1", userId, amount)
129     }
130     newBalance := tx.exec("SELECT balance FROM users WHERE user_id = $1", userId)
131
132     tx.Commit()
133
134     return newBalance
135 }
136
Изображение 9: Второй вариант (с откатом и возвратом)

go
121 struct db {
122
123 }
124
125
126 func(r *Repository) WithdrawBalance(userId int64, amount float64) (int, err) { // перевод всех денег в копейки(int, decimal)
127     tx := BeginTransaction("ISOLATION LEVEL READ COMMITTED") // вынести на уровень выше
128
129     defer tx.Rollback()
130     /*
131     oldBalance := tx.exec("SELECT balance FROM users WHERE user_id = $1", userId,)
132     if oldBalance >= amount {
133         tx.exec("UPDATE users SET balance = balance - $2 WHERE user_id = $1", userId, amount)
134     }
135     newBalance := tx.exec("SELECT balance FROM users WHERE user_id = $1", userId)
136     */
137     // 1000 2000 = -1000
138     // 1000 > 2000
139
140     balance, err := tx.exec("UPDATE users SET balance = balance - $2 WHERE user_id = $1 AND balance >= $2 RETURNING balance FOR UPDATE", userId, amount)
141     if err != nil {
142         return balance, err
143     }
144
145     tx.Commit()
146     return newBalance, nil
147 }
148
Блок 5: SQL (Основные операторы)
Изображение 10:

text
148
149 group,
150 order
151 join
152 limit
153 offset
154
155
156
Блок 6: Вопросы по Go (Что выведет код?)
Изображение 11: Вопрос 6 (Интерфейсы и nil)

go
157
158 6) Что выведет код?
159
160 func main() {
161     handler := func(string) error {
162         return errors.New("oi")
163     }
164
165     handler = nil
166     check(handler)
167 }
168
169 func check(v any) {
170     if v == nil {
171         fmt.Println("is nil")
172         return
173     }
174
175     if fn, ok := v.(func(string) error); ok {
176         fmt.Println("is func:", fn("test"))
177     }
178
179     if s, ok := v.(string); ok {
180         fmt.Println("is string:", s)
181     }
182 }
183
Изображение 12: Вопрос 3 (Замыкания в цикле)

go
189 3) Что выведет код?
190
191 func main() {
192     names := [4]string{"alpha", "beta", "gamma", "delta"}
193
194     var wg sync.WaitGroup
195     var k int
196
197     wg.Add(len(names))
198
199     for k = range names {
200         go func() {
201             defer wg.Done()
202             fmt.Println(names[k])
203         }()
204     }
205
206     wg.Wait()
207 }
208
Изображение 13: Вопрос 5 (Каналы и паника)

go
217 5) Что выведет код?
218
219 func main() {
220     ch := make(chan string)
221
222     go func() {
223         ch <- "hello"
224         ch <- "world" // panic
225     }()
226
227     time.Sleep(100 * time.Millisecond)
228     close(ch)
229
230     for msg := range ch {
231         fmt.Print(msg + " ") //
232     }
233
234     fmt.Println()
235 }
236
Изображение 14: Вопрос 4 (Мапы и структуры) - Ошибка

go
241 4) Что выведет код:
242
243 type User struct {
244     id   int
245     role string
246 }
247
248 func main() {
249     users := make(map[int]User)
250
251     users[1] = User{id: 1, role: "guest"}
252     users[1].role = "admin"
253
254     fmt.Println(users[1]) //
255 }
256
257
258
Изображение 15: Вопрос 4 - Разбор ошибки

go
241 4) Что выведет код:
242
243 type User struct {
244     id   int
245     role string
246 }
247
248 func main() {
249     users := make(map[int]*User) // <- не храните в мапах структуры как значения, ата-та
250
251     users[1] = &User{id: 1, role: "guest"}
252     users[1].role = "admin"
253
254     fmt.Println(users[1]) //
255 }
256
257 // cannot assign to struct field users[1].role in map
258
259241 4) Что выведет код:
242
243 type User struct {
244     id   int
245     role string
246 }
247
248 func main() {
249     users := make(map[int]User) // <- не храните в мапах структуры как значения, ата-та
250
251     users[1] = User{id: 1, role: "guest"}
252     users[1].role = "admin"
253
254     fmt.Println(users[1]) // cannot assign to struct field users[1].role in map
255 }
256
257 Прямая причина: элемент map не addressable. users[1] — это rvalue (копия),
258 не место в памяти, куда можно писать поле. Компилятор режет на этапе compile.
259
260 Почему так сделали: внутри map бакеты могут переезжать (grow / evacuation).
261 Если бы разрешили &users[1] или users[1].role = ..., указатель/запись
262 могли бы быть в уже переехавший бакет -> use-after-move
263
264 скриншот -
265 алго
266 3 задачки многопоточку
267
268 middle -> senior(архитектурные этап)
269
270
271 wb - не спрашивали
272 lamoda - трудовую
273 yandex - не спрашивали