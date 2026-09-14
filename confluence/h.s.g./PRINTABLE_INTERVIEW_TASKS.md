# Полное руководство к собесу: Go + SQL (middle+)

Распечатай или держи на втором экране. У каждой Go-задачи есть **краткое решение** (код).

Как учиться:
1. Закрой блок «Краткое решение» рукой.
2. Таймер: Easy 10–15 мин · Medium 20–30 · Design 10–15 устно.
3. Вслух: идея → сложность → края → тест.
4. Сверь с кодом ниже.

```bash
export GOWORK=off
cd confluence/h.s.g./NN_…/code/01_… && go run main.go
```

---

# 1. Язык Go — классика на собесе

## 1.1. Invert map

**Условие.** Дана `map[string]int`. Нужно вернуть новую map, где ключи и значения «перевёрнуты»: ключ — бывшее значение (`int`), значение — слайс всех исходных ключей (`[]string`) с этим числом. Если несколько ключей имели одно value — все они попадают в один слайс. Порядок ключей в слайсе и порядок обхода map не важны.

**Сигнатура.** `func invert(m map[string]int) map[int][]string`

**Пример 1**
```
Input:  map[string]int{"a": 1, "b": 2, "c": 1}
Output: map[int][]string{1: {"a", "c"}, 2: {"b"}}
```
(допустимо `1: {"c","a"}` — порядок в слайсе не фиксирован)

**Пример 2**
```
Input:  map[string]int{}
Output: map[int][]string{}  // или пустая map
```

**Края.** Пустая map; все values уникальны; все values одинаковы.

**Как решать.** Один проход: `out[v] = append(out[v], k)`.

**Краткое решение.**
```go
func invert(m map[string]int) map[int][]string {
    out := make(map[int][]string, len(m))
    for k, v := range m {
        out[v] = append(out[v], k)
    }
    return out
}
```

---

## 1.2. Filter slice in-place

**Условие.** Напиши `Filter`, которая оставляет в слайсе только элементы, для которых `pred(x) == true`. **Нельзя** выделять новый массив через `make` / создавать новый слайс с новым backing специально под результат — работай write-index’ом по исходному буферу. Исходный слайс можно портить. Верни `nums[:w]`, где `w` — число оставшихся.

**Сигнатура.** `func Filter(nums []int, pred func(int) bool) []int`

**Пример 1** (чётные)
```
Input:  nums = [1, 2, 3, 4, 5, 6], pred = even
Output: [2, 4, 6]
Len: 3  (cap может остаться 6 — это нормально)
```

**Пример 2**
```
Input:  nums = [1, 3, 5], pred = even
Output: []
```

**Края.** Пустой вход; все подходят; ни один не подходит.

**Как решать.** `w := 0`; если подходит — `nums[w]=x; w++`; `return nums[:w]`.

**Краткое решение.**
```go
func Filter(nums []int, pred func(int) bool) []int {
    w := 0
    for _, x := range nums {
        if pred(x) {
            nums[w] = x
            w++
        }
    }
    return nums[:w]
}
```

---

## 1.3. Slice header и append

**Условие.** Объясни (и покажи кодом), что произойдёт:

```go
func mutate(s []int) {
    s[0] = 100
    s = append(s, 4)
    s[1] = 200
}

func main() {
    data := []int{1, 2, 3} // len=3, cap=3
    mutate(data)
    fmt.Println(data) // ???
}
```

**Ожидаемый Output**
```
[100 2 3]
```
Не `[100 200 3]` и не `[100 200 3 4]`.

**Почему.** Slice header `{ptr,len,cap}` передаётся **по значению**. `s[0]=100` меняет общий array → видно снаружи. `append` при `cap==len` аллоцирует **новый** array → `s[1]=200` уже не в `data`.

**Доп. вопрос на собесе.** Что будет, если заранее `data := make([]int, 3, 8)` и тот же `mutate`?

**Краткое решение (демо).**
```go
func mutate(s []int) {
    s[0] = 100          // видно снаружи (тот же array)
    s = append(s, 4)    // при cap==len — НОВЫЙ array
    s[1] = 200          // уже другой array
}
// data := []int{1,2,3}; mutate(data) → [100 2 3]
```
Slice = `{ptr,len,cap}` by value.

---

## 1.4. Замыкание в цикле + горутины

**Условие.** Что выведет код? (На собесе ждут классическую семантику **до Go 1.22**.)

```go
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i)
    }()
}
time.Sleep(time.Second)
```

**Output (классика / что говорить)**
```
3
3
3
```
(порядок строк может быть любым, значения — три раза последнее `i`)

**Почему.** Замыкание захватывает **переменную** `i`, а не копию. К печати цикл уже закончен, `i == 3`.

**Задача:** напиши исправленный вариант, который печатает `0 1 2` (в любом порядке).

**Краткое решение.**
```go
for i := 0; i < 3; i++ {
    go func(id int) { // копия значения
        fmt.Println(id)
    }(i)
}
```

---

## 1.5. defer = LIFO

**Условие.** Что напечатает программа? В каком порядке?

```go
func main() {
    defer fmt.Println(1)
    defer fmt.Println(2)
    fmt.Println("start")
}
```

**Output**
```
start
2
1
```

**Правило.** `defer` = стек (LIFO). Аргументы вычисляются **сразу**, тело — при выходе из функции.

**Доп.**
```go
x := 1
defer fmt.Println(x) // напечатает 1
x = 2
```

**Краткое решение.** Порядок: `start` → `2` → `1` (LIFO). Код из условия выше — уже эталон.

---

## 1.6. panic / recover

**Условие.**
1. Напиши функцию, которая делает `panic("boom")`, но `main` не падает — паника перехвачена.
2. Ответь: поймает ли `recover` в `main` панику из **другой** горутины?

**Пример поведения**
```
Input:  safe() вызывается из main
Output:
  recovered: boom
  main continues
```

**Правила.** `recover` только в `defer` и только в **той же** горутине. Код после `panic` в функции не выполняется.

**Краткое решение.**
```go
func safe() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()
    panic("boom")
}
```
Только defer + та же горутина. Чужую горутину `recover` в `main` **не** спасёт.

---

## 1.7. map: порядок и гонки

**Условие.**
1. Гарантирован ли порядок `for k, v := range m`?  
2. Что будет, если две горутины пишут в одну map без синхронизации?  
3. Напиши безопасную запись в map из N горутин.

**Пример (гонка — плохо)**
```go
m := map[int]int{}
for i := 0; i < 10; i++ {
    go func(i int) { m[i] = i }(i) // fatal: concurrent map writes
}
```

**Ожидаемое знание Output при гонке:** процесс падает с `fatal error: concurrent map writes`.

**Решение по смыслу:** `sync.Mutex` вокруг записи (или `sync.Map` — реже на middle-собесе).

**Краткое решение.**
```go
var mu sync.Mutex
m := map[string]int{}

mu.Lock()
m[key] = val
mu.Unlock()
```
`range` — unordered; concurrent write без mutex → fatal.

---

## 1.8. sync.Once

**Условие.** Есть дорогая инициализация (загрузка конфига). Её могут вызвать 100 горутин одновременно — выполниться она должна **ровно один раз**, все должны увидеть готовое значение.

**Пример**
```
Input:  5 горутин параллельно вызывают load()
Output: init once   // ровно одна строка
        ready
        ready
        ...
```

**Сигнатура-идея.** `func load() string` с `sync.Once` внутри пакета.

**Краткое решение.**
```go
var once sync.Once
var cfg string

func load() string {
    once.Do(func() { cfg = "ready" })
    return cfg
}
```

---

# 2. Алгоритмы (лайвкодинг)

## 2.1. Two Sum

**Условие.** Дан массив целых `nums` и число `target`. Верни индексы двух различных элементов, сумма которых равна `target`. Гарантируется ровно одно решение (или верни `nil`, если нет — уточни на собесе). Нельзя использовать один элемент дважды.

**Сигнатура.** `func twoSum(nums []int, target int) []int`

**Пример 1**
```
Input:  nums = [2, 7, 11, 15], target = 9
Output: [0, 1]
Explanation: nums[0] + nums[1] = 2 + 7 = 9
```

**Пример 2**
```
Input:  nums = [3, 2, 4], target = 6
Output: [1, 2]
```

**Пример 3**
```
Input:  nums = [3, 3], target = 6
Output: [0, 1]
```

**Сложность цели.** O(n) время, O(n) память (hash map). Brute O(n²) — упомяни и улучши.

**Краткое решение.**
```go
func twoSum(nums []int, target int) []int {
    seen := map[int]int{}
    for i, n := range nums {
        if j, ok := seen[target-n]; ok {
            return []int{j, i}
        }
        seen[n] = i
    }
    return nil
}
```
O(n) / O(n).

---

## 2.2. Valid Parentheses

**Условие.** Дана строка `s`, содержащая только `'('`, `')'`, `'{'`, `'}'`, `'['`, `']'`. Строка валидна, если:
1. Открытые скобки закрываются тем же типом.
2. Открытые скобки закрываются в правильном порядке.
3. Каждой закрывающей соответствует открывающая.

**Сигнатура.** `func isValid(s string) bool`

**Пример 1**
```
Input:  s = "()"
Output: true
```

**Пример 2**
```
Input:  s = "()[]{}"
Output: true
```

**Пример 3**
```
Input:  s = "(]"
Output: false
```

**Пример 4**
```
Input:  s = "([)]"
Output: false
```

**Пример 5**
```
Input:  s = "{[]}"
Output: true
```

**Края.** `""` → true; начинается с `)` → false.

**Краткое решение.**
```go
func isValid(s string) bool {
    st := []rune{}
    pair := map[rune]rune{')': '(', '}': '{', ']': '['}
    for _, ch := range s {
        if open, ok := pair[ch]; ok {
            if len(st) == 0 || st[len(st)-1] != open {
                return false
            }
            st = st[:len(st)-1]
        } else {
            st = append(st, ch)
        }
    }
    return len(st) == 0
}
```

---

## 2.3. Single Number (XOR)

**Условие.** Дан непустой массив `nums`, в котором каждый элемент встречается **дважды**, кроме одного — он встречается один раз. Найди этот элемент. Требование: линейное время и **константная** доп. память.

**Сигнатура.** `func singleNumber(nums []int) int`

**Пример 1**
```
Input:  nums = [2, 2, 1]
Output: 1
```

**Пример 2**
```
Input:  nums = [4, 1, 2, 1, 2]
Output: 4
```

**Пример 3**
```
Input:  nums = [1]
Output: 1
```

**Идея.** XOR: `a^a=0`, `a^0=a`, пары обнуляются.

**Краткое решение.**
```go
func singleNumber(nums []int) int {
    x := 0
    for _, v := range nums {
        x ^= v
    }
    return x
}
```

---

## 2.4. Maximum Average Subarray

**Условие.** Дан массив `nums` из `n` целых и целое `k`. Найди непрерывный подмассив длины **ровно** `k` с максимальным средним арифметическим. Верни это среднее (`float64`).

**Сигнатура.** `func findMaxAverage(nums []int, k int) float64`

**Пример 1**
```
Input:  nums = [1, 12, -5, -6, 50, 3], k = 4
Output: 12.75
Explanation: подмассив [12, -5, -6, 50], сумма 51, среднее 51/4 = 12.75
```

**Пример 2**
```
Input:  nums = [5], k = 1
Output: 5.0
```

**Подход.** Sliding window по сумме длины k, затем `/ k`.

**Краткое решение.**
```go
func findMaxAverage(nums []int, k int) float64 {
    sum := 0
    for i := 0; i < k; i++ {
        sum += nums[i]
    }
    best := sum
    for i := k; i < len(nums); i++ {
        sum += nums[i] - nums[i-k]
        if sum > best {
            best = sum
        }
    }
    return float64(best) / float64(k)
}
```

---

## 2.5. Product Except Self

**Условие.** Дан массив `nums` длины n. Верни массив `answer` той же длины, где `answer[i]` — произведение всех элементов `nums`, **кроме** `nums[i]`. Решать **без деления**, за O(n). Доп. массив кроме ответа желательно O(1) (не считая выхода).

**Сигнатура.** `func productExceptSelf(nums []int) []int`

**Пример 1**
```
Input:  nums = [1, 2, 3, 4]
Output: [24, 12, 8, 6]
Explanation:
  i=0: 2*3*4=24
  i=1: 1*3*4=12
  i=2: 1*2*4=8
  i=3: 1*2*3=6
```

**Пример 2**
```
Input:  nums = [-1, 1, 0, -3, 3]
Output: [0, 0, 9, 0, 0]
```

**Края.** Нули в массиве — алгоритм prefix/suffix всё равно корректен.

**Краткое решение.**
```go
func productExceptSelf(nums []int) []int {
    n := len(nums)
    out := make([]int, n)
    out[0] = 1
    for i := 1; i < n; i++ {
        out[i] = out[i-1] * nums[i-1]
    }
    r := 1
    for i := n - 1; i >= 0; i-- {
        out[i] *= r
        r *= nums[i]
    }
    return out
}
```

---

## 2.6. Merge Two Sorted Lists

**Условие.** Даны головы двух отсортированных по неубыванию связных списков. Слей их в один отсортированный список (можно переиспользовать узлы). Верни голову результата.

**Тип.** `type ListNode struct { Val int; Next *ListNode }`

**Сигнатура.** `func mergeTwoLists(a, b *ListNode) *ListNode`

**Пример 1**
```
Input:  a = [1,2,4], b = [1,3,4]
Output: [1,1,2,3,4,4]
```

**Пример 2**
```
Input:  a = [], b = []
Output: []
```

**Пример 3**
```
Input:  a = [], b = [0]
Output: [0]
```

**Подход.** Dummy head + два указателя.

**Краткое решение.**
```go
func mergeTwoLists(a, b *ListNode) *ListNode {
    dummy := &ListNode{}
    cur := dummy
    for a != nil && b != nil {
        if a.Val < b.Val {
            cur.Next, a = a, a.Next
        } else {
            cur.Next, b = b, b.Next
        }
        cur = cur.Next
    }
    if a != nil {
        cur.Next = a
    } else {
        cur.Next = b
    }
    return dummy.Next
}
```

---

## 2.7. Linked List Cycle

**Условие.** Дан head связного списка. Определи, есть ли в нём цикл (`true`/`false`). Цикл есть, если какой-то узел достижим снова при движении по `Next`.

**Сигнатура.** `func hasCycle(head *ListNode) bool`

**Пример 1**
```
Input:  head = [3,2,0,-4], pos = 1  // хвост указывает на индекс 1
Output: true
```

**Пример 2**
```
Input:  head = [1,2], pos = 0
Output: true
```

**Пример 3**
```
Input:  head = [1], pos = -1
Output: false
```

**Подход.** Floyd: slow +1, fast +2. Встретились → цикл. Без доп. set → O(1) память.

**Краткое решение.**
```go
func hasCycle(head *ListNode) bool {
    slow, fast := head, head
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
        if slow == fast {
            return true
        }
    }
    return false
}
```

---

## 2.8. LRU Cache

**Условие.** Спроектируй LRU-кэш с capacity. Операции:
- `Get(key int) int` — вернуть value или `-1`, если нет; обращение делает ключ «самым свежим».
- `Put(key, value int)` — вставить/обновить; если размер превысил capacity, вытеснить **least recently used**.

Обе операции — **O(1)** в среднем.

**Пример**
```
Input:
  LRUCache(2)
  put(1, 1)
  put(2, 2)
  get(1)       → 1
  put(3, 3)    // вытесняет key 2
  get(2)       → -1
  put(4, 4)    // вытесняет key 1
  get(1)       → -1
  get(3)       → 3
  get(4)       → 4
```

**Структура.** `map[key]*node` + doubly linked list + sentinels head/tail.

**Краткое решение (ядро).**
```go
type node struct{ k, v int; prev, next *node }
type LRU struct {
    cap int
    m map[int]*node
    head, tail *node // sentinels
}

func (l *LRU) Get(k int) int {
    n, ok := l.m[k]
    if !ok { return -1 }
    l.remove(n); l.front(n)
    return n.v
}
func (l *LRU) Put(k, v int) {
    if n, ok := l.m[k]; ok {
        n.v = v; l.remove(n); l.front(n); return
    }
    n := &node{k: k, v: v}
    l.m[k] = n; l.front(n)
    if len(l.m) > l.cap {
        x := l.tail.prev
        l.remove(x); delete(l.m, x.k)
    }
}
// remove / front — стандартная двусвязка с head/tail
```

---

## 2.9. Trie

**Условие.** Реализуй префиксное дерево для строк из `'a'..'z'`:
- `Insert(word)`
- `Search(word) bool` — слово целиком есть
- `StartsWith(prefix) bool` — есть слово с таким префиксом

**Пример**
```
Input:
  Insert("apple")
  Search("apple")   → true
  Search("app")     → false
  StartsWith("app") → true
  Insert("app")
  Search("app")     → true
```

**Краткое решение.**
```go
type TrieNode struct {
    ch  [26]*TrieNode
    end bool
}
type Trie struct{ root *TrieNode }

func (t *Trie) Insert(w string) {
    n := t.root
    for i := 0; i < len(w); i++ {
        j := w[i] - 'a'
        if n.ch[j] == nil { n.ch[j] = &TrieNode{} }
        n = n.ch[j]
    }
    n.end = true
}
func (t *Trie) walk(s string) *TrieNode {
    n := t.root
    for i := 0; i < len(s); i++ {
        j := s[i] - 'a'
        if n.ch[j] == nil { return nil }
        n = n.ch[j]
    }
    return n
}
func (t *Trie) Search(w string) bool {
    n := t.walk(w); return n != nil && n.end
}
func (t *Trie) StartsWith(p string) bool { return t.walk(p) != nil }
```

---

## 2.10. Ближайший товар (Avito-style)

**Условие.** Даны массив товаров `products` (числа) и массив потребностей покупателей `needs`. Для каждой потребности выбирается товар с **минимальным** |need − product| (если два равноудалены — любой / ближайший по правилу «влево или вправо» — зафиксируй на собесе). Количество каждого товара не ограничено. Верни **сумму** неудовлетворённостей (сумму модулей) по всем покупателям.

**Сигнатура.** `func totalUnmet(products, needs []int) int`

**Пример**
```
Input:  products = [1, 3, 10], needs = [2, 5, 9]
Output: 4
Explanation:
  need 2 → товар 1 или 3, |2-1|=1 или |2-3|=1 → 1
  need 5 → товар 3, |5-3|=2
  need 9 → товар 10, |9-10|=1
  сумма 1+2+1 = 4
```

**Подход.** Sort products + binary search на каждый need.

**Краткое решение.**
```go
func totalUnmet(products, needs []int) int {
    sort.Ints(products)
    sum := 0
    for _, need := range needs {
        i := sort.SearchInts(products, need)
        var best int
        switch {
        case i == 0:
            best = products[0]
        case i == len(products):
            best = products[len(products)-1]
        default:
            a, b := products[i-1], products[i]
            if need-a <= b-need { best = a } else { best = b }
        }
        if best > need { sum += best - need } else { sum += need - best }
    }
    return sum
}
```

---

## 2.11. Champions (шаги)

**Условие.** Соревнование идёт несколько дней. Каждый день — список `{UserId, Steps}`. Чемпионы:
1. Набрали **максимальную** сумму шагов за все дни.
2. Имеют запись **в каждом** дне (не пропустили день).

Верни всех чемпионов (ничья возможна) и их суммарные шаги.

**Типы.**
```go
type Statistics struct{ UserId, Steps int }
type Result struct{ UserIds []int; Steps int }
```

**Пример 1**
```
Input:
  day0: [{1,1000},{2,1500}]
  day1: [{2,1000}]
Output: UserIds=[2], Steps=2500
Explanation: user 1 пропустил day1; user 2: 1500+1000=2500
```

**Пример 2**
```
Input:
  day0: [{1,2000},{2,1500}]
  day1: [{2,4000},{1,3500}]
Output: UserIds=[1,2], Steps=5500
Explanation: оба прошли все дни; 2000+3500=5500, 1500+4000=5500
```

**Ловушка.** Если user дважды в одном дне — для «присутствия» день считай один раз.

**Краткое решение.**
```go
func getChampions(days [][]Statistics) Result {
    if len(days) == 0 { return Result{} }
    total, present := map[int]int{}, map[int]int{}
    for _, day := range days {
        seen := map[int]bool{}
        for _, s := range day {
            total[s.UserId] += s.Steps
            if !seen[s.UserId] {
                present[s.UserId]++; seen[s.UserId] = true
            }
        }
    }
    best, ids := -1, []int{}
    for uid, sum := range total {
        if present[uid] != len(days) { continue }
        if sum > best { best, ids = sum, []int{uid}
        } else if sum == best { ids = append(ids, uid) }
    }
    sort.Ints(ids)
    if best < 0 { return Result{} }
    return Result{UserIds: ids, Steps: best}
}
```

---

## 2.12. Merge intervals

**Условие.** Дан массив интервалов `intervals`, где `intervals[i] = [start_i, end_i]`. Слей все пересекающиеся. Верни массив непересекающихся интервалов, покрывающих вход.

**Сигнатура.** `func merge(intervals [][]int) [][]int`

**Пример 1**
```
Input:  intervals = [[1,3],[2,6],[8,10],[15,18]]
Output: [[1,6],[8,10],[15,18]]
```

**Пример 2**
```
Input:  intervals = [[1,4],[4,5]]
Output: [[1,5]]
```

**Подход.** Sort by start; расширяй текущий, пока `next.start <= cur.end`.

**Краткое решение.**
```go
func merge(intervals [][]int) [][]int {
    sort.Slice(intervals, func(i, j int) bool {
        return intervals[i][0] < intervals[j][0]
    })
    out := [][]int{intervals[0]}
    for _, cur := range intervals[1:] {
        last := out[len(out)-1]
        if cur[0] <= last[1] {
            if cur[1] > last[1] { last[1] = cur[1] }
        } else {
            out = append(out, cur)
        }
    }
    return out
}
```

---

## 2.13. Longest substring without repeating

**Условие.** Дана строка `s`. Найди длину самой длинной подстроки **без повторяющихся** символов.

**Сигнатура.** `func lengthOfLongestSubstring(s string) int`

**Пример 1**
```
Input:  s = "abcabcbb"
Output: 3
Explanation: "abc"
```

**Пример 2**
```
Input:  s = "bbbbb"
Output: 1
```

**Пример 3**
```
Input:  s = "pwwkew"
Output: 3
Explanation: "wke"
```

**Подход.** Sliding window + last index символа.

**Краткое решение.**
```go
func lengthOfLongestSubstring(s string) int {
    last := map[byte]int{}
    best, left := 0, 0
    for right := 0; right < len(s); right++ {
        c := s[right]
        if i, ok := last[c]; ok && i >= left {
            left = i + 1
        }
        last[c] = right
        if right-left+1 > best {
            best = right - left + 1
        }
    }
    return best
}
```

---

# 3. Concurrency

## 3.1. Semaphore (limit K)

**Условие.** Есть `N` независимых задач (например, скачать N URL). Одновременно может выполняться не больше `K` задач. Реализуй запуск с ограничением через буферизованный канал (`chan struct{}`) + `WaitGroup`.

**Пример**
```
Input:  N=9 jobs, K=3
Output: в любой момент ≤3 горутины в работе;
        все 9 завершаются; программа не зависает
```

**Фраза на собесе.** Это ограничение **параллельности**, не rate (req/s).

**Краткое решение.**
```go
sem := make(chan struct{}, K)
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    sem <- struct{}{} // acquire
    go func(j Job) {
        defer wg.Done()
        defer func() { <-sem }() // release
        do(j)
    }(job)
}
wg.Wait()
```

---

## 3.2. Fan-in (merge channels)

**Условие.** Дан variadic набор каналов `<-chan int`. Верни один канал, в который приходят **все** значения из входов (в любом порядке). Когда все входы закрыты и значения отданы — закрой выходной канал.

**Сигнатура.** `func merge(cs ...<-chan int) <-chan int`

**Пример**
```
Input:
  ch1 yields 1, 3 then close
  ch2 yields 2, 4 then close
Output (порядок может отличаться): 1,2,3,4 затем выходной канал закрыт
```

**Ловушки.** Не `for v := range <-c`. Capture `c` в замыкании. `close(out)` только после WaitGroup.

**Краткое решение.**
```go
func merge(cs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    wg.Add(len(cs))
    for _, c := range cs {
        c := c
        go func() {
            defer wg.Done()
            for v := range c { out <- v }
        }()
    }
    go func() { wg.Wait(); close(out) }()
    return out
}
```

---

## 3.3. Fan-out / worker pool

**Условие.** Есть список jobs. Запусти ровно `W` воркеров, которые читают из общего `jobs` channel. Продюсер отправляет все jobs и закрывает канал. Дождись завершения всех воркеров.

**Пример**
```
Input:  jobs=[1..6], W=3
Output: каждый job обработан ровно один раз;
        после close(jobs) воркеры завершаются; wg.Wait() возвращается
```

**Краткое решение.**
```go
jobs := make(chan int)
var wg sync.WaitGroup
for w := 0; w < workers; w++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs { process(j) }
    }()
}
for _, j := range all { jobs <- j }
close(jobs)
wg.Wait()
```

---

## 3.4. Context timeout

**Условие.** Операция «спит» 200ms, но контекст отменяется через 50ms. Функция должна вернуть ошибку дедлайна, а не успех.

**Пример**
```
Input:  timeout=50ms, work=200ms
Output: error = context deadline exceeded
```

**Связь.** HTTP: `http.NewRequestWithContext(ctx, ...)`.

**Краткое решение.**
```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

select {
case <-time.After(200 * time.Millisecond):
    // успел
case <-ctx.Done():
    return ctx.Err() // deadline exceeded
}
```

---

## 3.5. Retry + backoff + jitter

**Условие.** Напиши `retry(fn, maxRetries, base)`: пока `fn()` возвращает ошибку — повторяй, но не больше `max` раз. Между попытками — exponential backoff `base * 2^i`, желательно с **full jitter** `rand(0, backoff)`. Если успех — верни nil. Если все попытки провалились — верни последнюю ошибку.

**Пример**
```
Input:  fn падает 2 раза, на 3-м успех; max=5, base=40ms
Output: nil, число вызовов fn == 3
        паузы roughly 40ms, 80ms между попытками (с jitter меньше)
```

**Не ретраить:** 400 / validation / non-idempotent payment без ключа.

**Краткое решение.**
```go
func retry(fn func() error, max int, base time.Duration) error {
    var last error
    for i := 0; i < max; i++ {
        if last = fn(); last == nil {
            return nil
        }
        if i == max-1 { break }
        exp := base * time.Duration(1<<i)
        wait := time.Duration(rand.Int63n(int64(exp) + 1)) // full jitter
        time.Sleep(wait)
    }
    return last
}
```

---

## 3.6. HTTP в цикле — фикс

**Условие.** Дан «сломанный» код: горутины в цикле по URL дергают `http.Get` без timeout, без обработки ошибок, с замыканием на loop variable. Исправь.

**Плохой Input (идея)**
```go
for _, url := range urls {
    go func() { http.Get(url) }() // баги
}
```

**Требования к Output/поведению**
- каждая горутина ходит по **своему** URL;
- есть timeout (Client или context);
- ошибки обрабатываются;
- `Body.Close()`;
- `WaitGroup` корректно ждёт всех.

**Краткое решение.**
```go
client := &http.Client{Timeout: 3 * time.Second}
var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        resp, err := client.Get(u)
        if err != nil { return }
        defer resp.Body.Close()
        // ...
    }(url)
}
wg.Wait()
```

---

# 4. Production-паттерны

## 4.1. Token Bucket

**Условие.** Реализуй rate limiter: ведро ёмкости `max` токенов, пополнение со скоростью `rate` токенов/сек. Метод `Allow() bool`: если токен есть — забери один и верни true, иначе false. Потокобезопасно. Допустима lazy refill по времени (без отдельной горутины).

**Пример**
```
Input:  max=3, rate=5/sec; 6 вызовов Allow подряд с паузой ~50ms
Output: первые ~3 true, далее смесь true/false по мере refill
```

Сравни устно Fixed Window / Sliding / Leaky (таблица ниже после кода).

**Краткое решение (lazy refill).**
```go
type TB struct {
    tokens, max, rate float64 // rate = tokens/sec
    last time.Time
    mu   sync.Mutex
}

func (t *TB) Allow() bool {
    t.mu.Lock()
    defer t.mu.Unlock()
    now := time.Now()
    t.tokens += now.Sub(t.last).Seconds() * t.rate
    if t.tokens > t.max { t.tokens = t.max }
    t.last = now
    if t.tokens < 1 { return false }
    t.tokens--
    return true
}
```

| Алгоритм | Плюс | Минус |
|----------|------|--------|
| Fixed Window | Просто | Burst на границе |
| Sliding Log | Точно | Память |
| Token Bucket | Burst после простоя | Нужен refill |
| Leaky | Ровный выход | Меньше накопить |

---

## 4.2. Round Robin LB

**Условие.** Список адресов серверов. `Next()` по кругу отдаёт следующий адрес. Потокобезопасно (atomic или mutex).

**Пример**
```
Input:  servers = ["a:8080", "b:8080", "c:8080"]
        Next() вызван 7 раз
Output: a, b, c, a, b, c, a
```

**Устно.** Когда Weighted / Least Connections / IP Hash вместо RR?

**Краткое решение.**
```go
type RR struct{ servers []string; i uint64 }

func (r *RR) Next() string {
    n := atomic.AddUint64(&r.i, 1) - 1
    return r.servers[n%uint64(len(r.servers))]
}
```
Ещё: Weighted / Least Conn / IP Hash — выбирай по нагрузке и sticky.

---

## 4.3. Circuit Breaker

**Условие.** Реализуй CB с состояниями Closed → Open → Half-Open:
- после `threshold` ошибок подряд — Open;
- в Open сразу возвращай ошибку (в HTTP API это **503**), не вызывая upstream;
- после `coolDown` — Half-Open, пробные вызовы;
- успехи возвращают в Closed, ошибка в Half-Open снова Open.

**Важно.** Mutex **не** держать во время `fn()`.

**Пример сценария**
```
Input:  threshold=2, upstream всегда error
Output: call1 error (Closed), call2 error → Open,
        call3 ErrOpen (fail fast), после coolDown — HalfOpen …
```

**Краткое решение (упрощённо).**
```go
func (cb *CB) Call(fn func() error) error {
    cb.mu.Lock()
    if cb.state == Open {
        if time.Since(cb.openedAt) < cb.coolDown {
            cb.mu.Unlock()
            return ErrOpen // → HTTP 503
        }
        cb.state = HalfOpen
    }
    cb.mu.Unlock()

    err := fn() // ВНЕ mutex!

    cb.mu.Lock()
    defer cb.mu.Unlock()
    if err != nil {
        cb.fails++
        if cb.state == HalfOpen || cb.fails >= cb.threshold {
            cb.state, cb.openedAt = Open, time.Now()
        }
        return err
    }
    cb.fails, cb.state = 0, Closed
    return nil
}
```

---

## 4.4. Health check

**Условие.** Периодически (ticker) проверяй HTTP endpoint с `context.WithTimeout`. По `ctx.Done()` останавливай loop. Объясни разницу **liveness** vs **readiness**.

**Пример поведения**
```
Input:  interval=50ms, endpoint сначала 503, потом 200
Output: healthy=false … затем healthy=true
```

**Краткое решение.**
```go
func (hc *HealthChecker) loop(ctx context.Context) {
    t := time.NewTicker(hc.interval)
    defer t.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.C:
            cctx, cancel := context.WithTimeout(ctx, hc.timeout)
            // GET /ready …
            cancel()
        }
    }
}
```
Liveness ≠ readiness.

---

## 4.5. Cache-Aside

**Условие.** Опиши/напиши Get:
1. смотрим cache;
2. miss → читаем DB;
3. кладём в cache;
4. при записи в DB — invalidate (или update) cache.

**Пример**
```
Input:  Get("user:1") первый раз
Output: cache miss → DB load → cache set → return

Input:  Get("user:1") второй раз
Output: cache hit → return без DB
```

Сравни устно Write-Through / Write-Back.

**Краткое решение.**
```go
func Get(key string) (Val, error) {
    if v, ok := cache.Get(key); ok {
        return v, nil
    }
    v, err := db.Load(key)
    if err != nil { return v, err }
    cache.Set(key, v)
    return v, nil
}
// write: db.Save(v); cache.Delete(key) // invalidate
```

---

## 4.6. Сквозной ответ (30 сек)

**Условие.** Интервьюер: «Как защитить sync-путь сервиса под нагрузкой от входа до зависимости?» Ответь связным рассказом на ~30 секунд, без воды.

**Краткое решение (устно).**

> Rate limit на входе → LB между инстансами → на downstream timeout + retry/jitter и снаружи circuit breaker (open → 503) → health/ready → cache-aside на чтение.

---

# 5. SQL — полный минимум

Схема в голове: `users`, `orders`, `order_items`, `products`.  
Якорь урока 00 (покупки до бана): `00_4th_March_2026/code/05_sql_purchases_before_ban/`.

## 5.1. Топ-N

**Условие.** Топ-10 пользователей по сумме заказов за текущий месяц.

**Input**
```
users:  id | name
         1 | Anna
         2 | Boris
         3 | Clara

orders: id | user_id | amount | created_at
         1 | 1       | 500    | 2026-03-02
         2 | 1       | 700    | 2026-03-10
         3 | 2       | 100    | 2026-03-05
         4 | 3       | 900    | 2025-12-01   -- другой месяц, не считать
```

**Output**
```
id | name  | total
 1 | Anna  | 1200
 2 | Boris |  100
```
(Clara не в топе месяца; LIMIT 10.)

```sql
SELECT u.id, u.name, SUM(o.amount) AS total
FROM users u
JOIN orders o ON o.user_id = u.id
WHERE o.created_at >= DATE_TRUNC('month', NOW())
GROUP BY u.id, u.name
ORDER BY total DESC
LIMIT 10;
```

## 5.2. HAVING

**Условие.** Категории, где средний чек заказа > 1000 (фильтр после агрегации).

**Input**
```
products:     id | category_id
               1 | 10
               2 | 10
               3 | 20

orders:       id | amount
               1 | 1500
               2 | 500
               3 | 2000

order_items:  order_id | product_id
              1        | 1
              2        | 2
              3        | 3
```
(упрощённо: один item ≈ один order)

**Output**
```
category_id | avg_check
         20 | 2000
```
(категория 10: avg (1500+500)/2 = 1000 — не проходит `> 1000`)

```sql
SELECT p.category_id, AVG(o.amount) AS avg_check
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
GROUP BY p.category_id
HAVING AVG(o.amount) > 1000;
```

## 5.3. Клиенты без заказов

**Условие.** Все пользователи, у которых нет ни одной строки в `orders`.

**Input**
```
users:  1 Anna, 2 Boris, 3 Clara
orders: user_id=1, user_id=1
```

**Output**
```
id | name
 2 | Boris
 3 | Clara
```

```sql
SELECT u.*
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE o.id IS NULL;
```

## 5.4. Топ-3 в категории

**Условие.** В каждой `category_id` — три товара с наибольшими `sales` (window `ROW_NUMBER`).

**Input**
```
product_sales:
id | category_id | sales
 1 | A           | 100
 2 | A           | 90
 3 | A           | 80
 4 | A           | 70
 5 | B           | 50
 6 | B           | 40
```

**Output**
```
id 1,2,3 (A) и 5,6 (B) — у B меньше трёх, все что есть; rn<=3
```

```sql
SELECT * FROM (
  SELECT p.*,
         ROW_NUMBER() OVER (PARTITION BY category_id ORDER BY sales DESC) rn
  FROM product_sales p
) t WHERE rn <= 3;
```

## 5.5. «До первого X»

**Условие.** Найти для каждого пользователя момент первого «плохого» события (`MIN(bad_at)`), затем все события с `created_at < first_bad_at`.

**Input**
```
events: user_id | kind | created_at
        1       | ok   | 10:00
        1       | ok   | 10:05
        1       | bad  | 10:10
        1       | ok   | 10:15
        2       | ok   | 11:00
```

**Output**
```
user 1: строки 10:00 и 10:05
user 2: 11:00 (нет bad → все / или пусто — уточни у интервьюера)
```

```sql
WITH first_bad AS (
  SELECT user_id, MIN(created_at) AS first_bad_at
  FROM events
  WHERE kind = 'bad'
  GROUP BY user_id
)
SELECT e.*
FROM events e
JOIN first_bad f ON f.user_id = e.user_id
WHERE e.created_at < f.first_bad_at;
```

## 5.6. DELETE дубликатов email

**Условие.** В `users` несколько строк с одним `email` — оставить одну (с минимальным `id`), остальные удалить.

**Input**
```
id | email
 1 | a@x.com
 2 | a@x.com
 3 | b@x.com
 4 | a@x.com
```

**Output** (после DELETE)
```
id | email
 1 | a@x.com
 3 | b@x.com
```

```sql
DELETE FROM users u
USING users d
WHERE u.email = d.email
  AND u.id > d.id;
-- или: DELETE … WHERE id NOT IN (SELECT MIN(id) … GROUP BY email)
```

## 5.7. Покупки до бана

**Условие.** (Урок 00) Покупки пользователя до даты бана; если бана нет — все покупки. Часто: DISTINCT user+SKU и/или SUM(price) > порог.

**Input**
```
purchases: user_id | sku | price | date
           1       | 1   | 5500  | 2021-02-15
           1       | 1   | 5700  | 2021-01-15
           1       | 2   | 4000  | 2021-02-14
           2       | 3   | 8000  | 2021-03-01
bans:      user_id=1, date_from=2021-03-08
```

**Output** (DISTINCT user+sku до бана)
```
user 1: sku 1, 2
user 2: sku 3
```

```sql
SELECT DISTINCT p.user_id, p.sku
FROM purchases p
LEFT JOIN bans b ON b.user_id = p.user_id
WHERE b.date_from IS NULL
   OR p.date < b.date_from;

-- суммы > 5000 до бана:
SELECT p.user_id, SUM(p.price) AS total
FROM purchases p
LEFT JOIN bans b ON b.user_id = p.user_id
WHERE b.date_from IS NULL OR p.date < b.date_from
GROUP BY p.user_id
HAVING SUM(p.price) > 5000;
```

Код-аналог: `00_4th_March_2026/code/05_sql_purchases_before_ban/main.go`.

## 5.8. Uptime %

**Условие.** Доля успешных проверок (`status = 'up'`) в процентах.

**Input**
```
checks: status
        up, up, down, up
```

**Output**
```
75.0
```

```sql
SELECT COUNT(*) FILTER (WHERE status = 'up') * 100.0 / COUNT(*) AS uptime_pct
FROM checks;
```

## 5.9+. Устно

- Индексы: B-tree; `EXPLAIN ANALYZE`; `LIKE '%x'` без индекса.
- MVCC: UPDATE = новые версии → bloat → VACUUM.
- MV: `REFRESH MATERIALIZED VIEW CONCURRENTLY`.
- N+1 → JOIN / `IN`.
- Шард key `user_id`, риск hot shard.

---

# 6. Устный чеклист

**Сеть:** TCP/UDP · TLS handshake · 401/403/503  
**Go:** GMP · channel buf · Mutex/RWMutex/atomic · nil interface  
**Архитектура:** Outbox · брокер vs sync · Idempotency-Key  

---

# 7. План на 7 дней

| День | Фокус |
|------|--------|
| 1 | §1 язык |
| 2 | §2.1–2.5 |
| 3 | §2.6–2.13 |
| 4 | §3 concurrency |
| 5 | §4 patterns |
| 6 | §5 SQL |
| 7 | Микс + §6 + ответ 4.6 |

---

*Краткие Go-решения — в каждом разделе выше. Разборы уроков: `NN_…/README.md`. Ритуал «собес завтра»: [`MANUSCRIPT.md`](./MANUSCRIPT.md). Smoke: [`scripts/smoke_mains.sh`](./scripts/smoke_mains.sh).*
