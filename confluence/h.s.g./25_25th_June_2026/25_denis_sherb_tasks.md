Блок 1: Задача "Один уникальный элемент" (XOR)
Изображение 1:

go
10 Пример 1:
11 - Входные данные: `nums = [2, 2, 1]`
12 - Вывод: `1`
13 - Объяснение: Так как число 2 повторилось дважды, единственным уникальным числом остается 1.
14
15 Пример 2:
16
17 - Входные данные: `nums = [4, 1, 2, 1, 2]`
18 - Вывод: `4` - Объяснение: Числа 1 и 2 встречаются по два раза, а число 4 - только один раз.
19
20
21 Пример 3: - Входные данные: `nums = [1]`
22 - Вывод: `1`
23
24 Ограничения:
25 - Длина массива `nums`: от 1 до 30 000 элементов.
26 - Значение каждого элемента: от -30 000 до 30 000.
27 - Каждый элемент в массиве появляется дважды, кроме одного уникального.
28 */
29
30 package main
31
32 func XOR(in []int) int {
33     x := 0
34
35     for _, v := range in {
36         x ^= v
37     }
38
39     return x
40 }
41
42 func main() {
43     nums := []int{4, 1, 2, 1, 2}
44
45
46
47
48 }
49
Блок 2: Задача "Валидные скобки"
Изображение 2: Условие

go
1 /*
2 Дана строка, состоящая только из символов '(', ')', '{', '}', '[' и ']'.
3 Напишите функцию, которая проверяет, является ли строка валидной.
4
5 Строка валидна, если:
6 1. Открытые скобки закрываются скобками того же типа.
7 2. Открытые скобки закрываются в правильном порядке.
8
9 Примеры:
10 in = "()"
11 out - true
12
13 in = "()[]{}"
14 out - true
15
16 in = "(]"
17 out - false
18
19 in = "([)]"
20 out - false
21 */
22
23 func isValid(s string) bool {
24     // Решение здесь
25 }
Изображение 3: Решение

go
func isValid(s string) bool {
    // Стек для хранения открывающих скобок
    stack := []rune{}

    // Мапа соответствия: закрывающая -> открывающая
    pairs := map[rune]rune{
        ')': '(',
        ']': '[',
        '}': '{',
    }

    for _, ch := range s {
        // Если это закрывающая скобка
        if opening, ok := pairs[ch]; ok {
            // Проверяем, что стек не пуст и последняя открывающая соответствует
            if len(stack) == 0 || stack[len(stack)-1] != opening {
                return false
            }
            // Убираем последнюю открывающую (она закрылась)
            stack = stack[:len(stack)-1]
        } else if ch == '(' || ch == '[' || ch == '{' {
            // Если открывающая - кладём в стек
            stack = append(stack, ch)
        }
        // Любой другой символ игнорируем (по условию задачи)
    }

    // В конце стек должен быть пуст - все скобки закрылись
    return len(stack) == 0
}
Блок 3: Задача "Пересечение массивов"
Изображение 4: Первое решение

go
1 /*
2 Даны два не отсортированных массива целых чисел.
3 Напишите функцию, которая вычислит их пересечение.
4 Каждый элемент в результирующем массиве должен появляться столько раз,
5 сколько он встречается в обоих исходных массивах.
6 Порядок элементов в результате не важен.
7
8 Примеры:
9 nums1 = [1, 2, 2, 1], nums2 = [2, 2]
10 out = [2, 2]
11
12 nums1 = [4, 9, 5], nums2 = [9, 4, 9, 8]
13 out = [4, 9] (или [9, 4])
14 */
15
16 func intersect(nums1 []int, nums2 []int) []int {
17     // Решение здесь
18     out := make([]int, 0, len(nums1))
19     m := make(map[int]int, len(nums2))
20
21     for _, t := range nums1 {
22         m[t]++
23     }
24
25     for _, n := range nums2 {
26         val, ok := m[n]
27         if ok && val > 0 {
28             out = append(out, n)
29             m[n]--
30         }
31     }
32     return out
33 }
Изображение 5: Второе решение (с оптимизацией)

go
1 /*
2 Даны два не отсортированных массива целых чисел.
3 Напишите функцию, которая вычислит их пересечение.
4 Каждый элемент в результирующем массиве должен появляться столько раз,
5 сколько он встречается в обоих исходных массивах.
6 Порядок элементов в результате не важен.
7
8 Примеры:
9 nums1 = [1, 2, 2, 1], nums2 = [2, 2]
10 out = [2, 2]
11
12 nums1 = [4, 9, 5], nums2 = [9, 4, 9, 8]
13 out = [4, 9] (или [9, 4])
14 */
15
16 func intersect(nums1 []int, nums2 []int) []int {
17     // Решение здесь
18     out := make([]int, 0, len(nums1))
19     m := make(map[int]int, len(nums2))
20
21     if len(nums1) > len(nums2) {
22         return intersect(nums2, nums1)
23     }
24
25     for _, t := range nums1 {
26         m[t]++
27     }
28
29     for _, n := range nums2 {
30         if m[n] > 0 {
31             out = append(out, n)
32             m[n]--
33         }
34     }
35
36     return out
37 }
Блок 4: Задача "Удаление дубликатов из отсортированного массива"
Изображение 6:

go
1 /*
2 Дан отсортированный массив целых чисел.
3 Необходимо удалить дубликаты без выделения памяти под новый массив так,
4 чтобы каждый уникальный элемент появлялся только один раз.
5 Функция должна вернуть новую длину логического массива.
6
7 Изменять размер исходного среза в самой функции не обязательно,
8 достаточно переставить элементы и вернуть индекс конца уникальной части.
9
10 Примеры:
11 in = [1, 1, 2]
12 out = 2 (массив изменяется на [1, 2, _])
13
14 in = [0, 0, 1, 1, 1, 2, 2, 3, 3, 4]
15 out = 5 (массив изменяется на [0, 1, 2, 3, 4, _, _, _, _, _])
16 */
17
18 func removeDuplicates(nums []int) int {
19     // Решение здесь
20
21     elem := 1
22
23     for i := 1; i < len(num); i++ {
24         if nums[i] != nums[i - 1] {
25             nums[elem] = nums[i]
26             elem++
27         }
28     }
29
30     return elem
31 }
Блок 5: Задача "Проверка на простое число" (Фрагмент)
Изображение 7:

go
33
34 func isPrime(n int) bool {
35     if n <= 1 {
36         return false
37     }
38
39     // Цикл идет, пока квадрат i меньше или равен n
40     for i := 2; i*i <= n; i++ {
41         if n % i == 0 {
42             return false
43         }
44     }
45
46     return true
47 }
48
Блок 6: Параллельный опрос URL (Health Check) - Эволюция решения
Изображение 8: Условие и начало

go
1 package main
2
3 import (
4     "fmt"
5     "net/http"
6     "time"
7 )
8
9 // ТРЕБОВАНИЯ:
10 // 1. Опросить все URL параллельно.
11 // 2. Собрать результаты (статус-код ответа или ошибку сетевого запроса).
12 // 3. Вывести результаты в консоль строго в том порядке, в котором URL идут в исходном слайсе.
13
14 func main() {
15     urls := []string{
16         "https://google.com",
17         "https://github.com",
18         "https://httpbin.org/delay/2", // этот сайт отвечает долго
19         "https://yandex.ru",
20         "https://httpbin.org/status/404",
21     }
22
23     // TODO
24
25 }
Изображение 9: Структура Response и функция process (начало)

go
1 package main
2
3 import (
4     "fmt"
5     "net/http"
6     "time"
7 )
8
9 // ТРЕБОВАНИЯ:
10 // 1. Опросить все URL параллельно.
11 // 2. Собрать результаты (статус-код ответа или ошибку сетевого запроса).
12 // 3. Вывести результаты в консоль строго в том порядке, в котором URL идут в исходном слайсе.
13
14 type Response struct {
15     url    string
16     status int
17     err    error
18 }
19
20 func process(client *http.Client, url string) Response {
21     r := &Response{
22         url: url,
23     }
24     resp, err := client.Get(url)
25     if err != nil {
26         r.err = err
27         return r
28     }
29     defer resp.Body.Close()
30
31     status := resp.StatusCode
32     r.status = status
33
34     return r
35
36
37
38
39 }
Изображение 10: Реализация main с WaitGroup (первый вариант)

go
41 func main() {
42     var wg sync.WaitGroup
43     urls := []string{
44         "https://google.com",
45         "https://github.com",
46         "https://httpbin.org/delay/2", // этот сайт отвечает долго
47         "https://yandex.ru",
48         "https://httpbin.org/status/404",
49     }
50
51     client := http.Client{
52         Timeout: 5 * time.Second,
53     }
54
55     resp := make([]Response, len(urls))
56     for i, url := range urls {
57         wg.Add(1)
58         go func() {
59             defer wg.Done()
60             r := process(client, url)
61             resp[i] = r
62         }()
63     }
64     wg.Wait()
65     fmt.Println(resp)
66
67 }
Изображение 11: Упрощенная функция process (второй вариант)

go
1 package main
2
3 import (
4     "fmt"
5     "net/http"
6     "time"
7 )
8
9 // ТРЕБОВАНИЯ:
10 // 1. Опросить все URL параллельно.
11 // 2. Собрать результаты (статус-код ответа или ошибку сетевого запроса).
12 // 3. Вывести результаты в консоль строго в том порядке, в котором URL идут в исходном слайсе.
13
14 type Response struct {
15     url    string
16     status int
17     err    error
18 }
19
20 func process(client *http.Client, url string) Response {
21     r := Response{
22         url: url,
23     }
24     resp, err := client.Get(url)
25     if err != nil {
26         r.err = err
27         return r
28     }
29     defer resp.Body.Close()
30
31     status := resp.StatusCode
32     r.status = status
33
34     return r
35
36
37     resp, err := client.Get(url)
38     if err != nil {
39         return Response{url: url, err: err}
40     }
41
42     return Response{url: url, status: resp.StatusCode}
43
44
45
46
47 }
Изображение 12: Финальная реализация main

go
49 func main() {
50     var wg sync.WaitGroup
51     urls := []string{
52         "https://google.com",
53         "https://github.com",
54         "https://httpbin.org/delay/2", // этот сайт отвечает долго
55         "https://yandex.ru",
56         "https://httpbin.org/status/404",
57     }
58
59     client := &http.Client{
60         Timeout: 5 * time.Second,
61     }
62
63     resp := make([]Response, len(urls))
64
65     for i, url := range urls {
66         wg.Add(1)
67         go func() {
68             defer wg.Done()
69             r := process(client, url)
70             resp[i] = r
71             resp[i] = process(client, url)
72         }()
73     }
74     wg.Wait()
75     fmt.Println(resp)
76
77     for _, res := range resp {
78         if res.err != nil {
79
80         } else {
81
82         }
83
84     }
85
86 }