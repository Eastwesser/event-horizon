# 🔥 Golang Insights — Базовые задачи для собеса

**Цель:** Отработать простые алгоритмические блоки, которые часто спрашивают на собесах.

---

## 📋 Список тем

### Simplest (Самые простые)
1. Map to Slice / Slice to Map
2. Max element in array/slice
3. Common occurance (частота элементов)
4. Matrix vice versa (транспонирование матриц)

### Дополнительно: Алгоритмические блоки
5. **Массивы:** Two Sum, Sliding Window, Rotate Array
6. **Строки:** Longest Substring, Anagrams
7. **Связные списки:** Reverse, Cycle Detection
8. **Деревья:** Trie, Ball Tree (геопоиск)
9. **Хеш-таблицы:** Group Anagrams, First Unique Char

---

## 🎯 Детальный разбор

### 1. Map to Slice / Slice to Map

**Map → Slice:**
```go
// Ключи в slice
func keysToSlice(m map[string]int) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

// Значения в slice
func valuesToSlice(m map[string]int) []int {
    values := make([]int, 0, len(m))
    for _, v := range m {
        values = append(values, v)
    }
    return values
}

// Пары key-value в slice
type Pair struct {
    Key   string
    Value int
}

func mapToSlice(m map[string]int) []Pair {
    pairs := make([]Pair, 0, len(m))
    for k, v := range m {
        pairs = append(pairs, Pair{k, v})
    }
    return pairs
}
```

**Slice → Map:**
```go
// Slice → Map (значения как ключи, индексы как значения)
func sliceToMap(s []string) map[string]int {
    m := make(map[string]int, len(s))
    for i, v := range s {
        m[v] = i  // Последний индекс для дубликатов
    }
    return m
}

// Slice → Set (map[string]struct{})
func sliceToSet(s []string) map[string]struct{} {
    set := make(map[string]struct{}, len(s))
    for _, v := range s {
        set[v] = struct{}{}
    }
    return set
}
```

**⚠️ Частая ошибка:**
```go
// ❌ ПЛОХО: не инициализируем capacity
keys := []string{}
for k := range m {
    keys = append(keys, k)  // Много реаллокаций!
}

// ✅ ХОРОШО: инициализируем с capacity
keys := make([]string, 0, len(m))
```

---

### 2. Max Element in Array/Slice

**Базовое решение:**
```go
func maxElement(arr []int) (int, error) {
    if len(arr) == 0 {
        return 0, errors.New("empty array")
    }
    
    max := arr[0]
    for _, v := range arr[1:] {
        if v > max {
            max = v
        }
    }
    return max, nil
}
```

**С индексом:**
```go
func maxWithIndex(arr []int) (value, index int, err error) {
    if len(arr) == 0 {
        return 0, -1, errors.New("empty array")
    }
    
    maxVal, maxIdx := arr[0], 0
    for i, v := range arr[1:] {
        if v > maxVal {
            maxVal, maxIdx = v, i+1
        }
    }
    return maxVal, maxIdx, nil
}
```

**Generic (Go 1.18+):**
```go
func Max[T constraints.Ordered](arr []T) (T, error) {
    var zero T
    if len(arr) == 0 {
        return zero, errors.New("empty array")
    }
    
    max := arr[0]
    for _, v := range arr[1:] {
        if v > max {
            max = v
        }
    }
    return max, nil
}

// Использование
maxInt, _ := Max([]int{1, 5, 3, 9, 2})     // 9
maxStr, _ := Max([]string{"a", "z", "b"})  // "z"
```

---

### 3. Common Occurrence (Частота элементов)

**Подсчёт частоты:**
```go
func frequency(arr []int) map[int]int {
    freq := make(map[int]int)
    for _, v := range arr {
        freq[v]++
    }
    return freq
}
```

**Самый частый элемент:**
```go
func mostFrequent(arr []int) (int, int) {
    freq := make(map[int]int)
    for _, v := range arr {
        freq[v]++
    }
    
    maxCount := 0
    mostFreq := 0
    for val, count := range freq {
        if count > maxCount {
            maxCount = count
            mostFreq = val
        }
    }
    return mostFreq, maxCount
}

// Пример
arr := []int{1, 2, 2, 3, 3, 3, 4}
val, count := mostFrequent(arr)  // val=3, count=3
```

**Top K частых элементов:**
```go
func topKFrequent(nums []int, k int) []int {
    // 1. Подсчёт частоты
    freq := make(map[int]int)
    for _, num := range nums {
        freq[num]++
    }
    
    // 2. Преобразуем в slice для сортировки
    type pair struct {
        num   int
        count int
    }
    
    pairs := make([]pair, 0, len(freq))
    for num, count := range freq {
        pairs = append(pairs, pair{num, count})
    }
    
    // 3. Сортируем по убыванию частоты
    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].count > pairs[j].count
    })
    
    // 4. Берём первые k
    result := make([]int, k)
    for i := 0; i < k; i++ {
        result[i] = pairs[i].num
    }
    return result
}
```

---

### 4. Matrix Vice Versa (Транспонирование)

**Транспонирование матрицы:**
```go
func transpose(matrix [][]int) [][]int {
    rows := len(matrix)
    cols := len(matrix[0])
    
    // Создаём новую матрицу cols x rows
    result := make([][]int, cols)
    for i := range result {
        result[i] = make([]int, rows)
    }
    
    // Копируем с транспонированием
    for i := 0; i < rows; i++ {
        for j := 0; j < cols; j++ {
            result[j][i] = matrix[i][j]
        }
    }
    
    return result
}

// Пример
matrix := [][]int{
    {1, 2, 3},
    {4, 5, 6},
}

transposed := transpose(matrix)
// [[1, 4],
//  [2, 5],
//  [3, 6]]
```

**Поворот матрицы на 90°:**
```go
func rotate90(matrix [][]int) [][]int {
    n := len(matrix)
    result := make([][]int, n)
    for i := range result {
        result[i] = make([]int, n)
    }
    
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            result[j][n-1-i] = matrix[i][j]
        }
    }
    
    return result
}
```

**In-place поворот (экономия памяти):**
```go
func rotate90InPlace(matrix [][]int) {
    n := len(matrix)
    
    // 1. Транспонирование
    for i := 0; i < n; i++ {
        for j := i + 1; j < n; j++ {
            matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j]
        }
    }
    
    // 2. Отражение по вертикали
    for i := 0; i < n; i++ {
        for j := 0; j < n/2; j++ {
            matrix[i][j], matrix[i][n-1-j] = matrix[i][n-1-j], matrix[i][j]
        }
    }
}
```

---

## 🔥 Дополнительные алгоритмические блоки

### 5. Массивы: Two Sum

**Задача:** Найти два индекса, сумма которых равна target.

```go
func twoSum(nums []int, target int) []int {
    seen := make(map[int]int)
    
    for i, num := range nums {
        complement := target - num
        if j, ok := seen[complement]; ok {
            return []int{j, i}
        }
        seen[num] = i
    }
    
    return nil
}
```

**Time:** O(n), **Space:** O(n)

---

### 6. Строки: Longest Substring Without Repeating

**Задача:** Найти длину самой длинной подстроки без повторяющихся символов.

```go
func lengthOfLongestSubstring(s string) int {
    seen := make(map[rune]int)
    maxLen := 0
    start := 0
    
    for i, char := range s {
        if lastIdx, ok := seen[char]; ok && lastIdx >= start {
            start = lastIdx + 1
        }
        
        seen[char] = i
        maxLen = max(maxLen, i-start+1)
    }
    
    return maxLen
}
```

**Sliding Window pattern!**

---

### 7. Связные списки: Reverse

```go
func reverseList(head *ListNode) *ListNode {
    var prev *ListNode
    current := head
    
    for current != nil {
        next := current.Next
        current.Next = prev
        prev = current
        current = next
    }
    
    return prev
}
```

**Cycle Detection (Floyd's Algorithm):**
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

### 8. Деревья: Trie (Префиксное дерево)

```go
type TrieNode struct {
    children map[rune]*TrieNode
    isEnd    bool
}

type Trie struct {
    root *TrieNode
}

func NewTrie() *Trie {
    return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(word string) {
    node := t.root
    for _, char := range word {
        if _, ok := node.children[char]; !ok {
            node.children[char] = &TrieNode{children: make(map[rune]*TrieNode)}
        }
        node = node.children[char]
    }
    node.isEnd = true
}

func (t *Trie) Search(word string) bool {
    node := t.root
    for _, char := range word {
        if _, ok := node.children[char]; !ok {
            return false
        }
        node = node.children[char]
    }
    return node.isEnd
}
```

**Use case:** Автокомплит, словари, поиск префиксов

---

### 9. Хеш-таблицы: Group Anagrams

```go
func groupAnagrams(strs []string) [][]string {
    groups := make(map[string][]string)
    
    for _, str := range strs {
        // Сортируем символы как ключ
        key := sortString(str)
        groups[key] = append(groups[key], str)
    }
    
    result := make([][]string, 0, len(groups))
    for _, group := range groups {
        result = append(result, group)
    }
    
    return result
}

func sortString(s string) string {
    runes := []rune(s)
    sort.Slice(runes, func(i, j int) bool {
        return runes[i] < runes[j]
    })
    return string(runes)
}
```

---

## 🎯 Советы для собеса

### 1. Всегда спрашивай про constraints
- Размер массива/строки?
- Диапазон значений?
- Уникальные элементы или с дубликатами?
- Отсортированный массив?

### 2. Обсуждай trade-offs
- **Time vs Space:** O(n) time, O(n) space vs O(n²) time, O(1) space
- **In-place vs новый массив**
- **Мутировать входные данные?**

### 3. Объясняй подход
1. Brute force (наивное решение)
2. Оптимизация (HashMap, Two Pointers, Sliding Window)
3. Проверка edge cases (пустой массив, один элемент)

### 4. Пиши чистый код
```go
// ❌ ПЛОХО
func f(a []int) int {
    m := a[0]
    for _, v := range a {
        if v > m {
            m = v
        }
    }
    return m
}

// ✅ ХОРОШО
func findMax(numbers []int) (int, error) {
    if len(numbers) == 0 {
        return 0, errors.New("empty slice")
    }
    
    max := numbers[0]
    for _, num := range numbers[1:] {
        if num > max {
            max = num
        }
    }
    return max, nil
}
```

---

## 🔗 См. также

- **CITADEL_COMPLETE_ANSWERS.md** — теория
- **CITADEL_COMPLETE_TASKS.md** — ЧАСТЬ 1: Shuffle, Zip, Merge
- **LeetCode** — практика алгоритмов

---

**Удачи на собесе!** 🚀
