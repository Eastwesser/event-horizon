### Суммарная неудовлетворённость покупателей

На Авито размещено множество товаров, каждый из которых представлен числом. 
У каждого покупателя есть потребность в товаре, также выраженная числом. 
Если точного товара нет, покупатель выбирает ближайший по значению товар, что вызывает неудовлетворённость, 
равную разнице между его потребностью и купленным товаром. Количество каждого товара не ограничено, и один товар могут купить несколько покупателей.
 Рассчитайте суммарную неудовлетворённость всех покупателей.
Нужно написать функцию, которая примет на вход два массива: массив товаров и массив потребностей покупателей, 
вычислит сумму неудовлетворённостей всех покупателей и вернет результат в виде числа.

Пример
ввод:
    goods = [8, 3, 5]
    buyerNeeds = [5, 6]
вывод:
    res = 1  
первый покупатель покупает товар 5 и его неудовлетворённость = 0, второй также покупает товар 5 и его неудовлетворённость = 6-5 = 1


func foo(goods, buyerNeeds []int) {
    res := 0
    slices.Sort(goods)

    centen := len(goods) / 2

    /*
    https://pkg.go.dev/sort#SearchInts
    func SearchStrings(a []string, x string) int
    SearchStrings searches for x in a sorted slice of strings and returns the index as specified by Search. 
    The return value is the index to insert x if x is not present (it could be len(a)). The slice must be sorted in ascending order.
    */
    for _, need := range buyerNeeds {
        diff := need
        for _, g := range goods {
            df := abs(need - g) < diff
            if df < diff {
                diff = df
            }
        }

        res += diff
    }

    return res
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

===
Блок 1: Задача Health Check (Первые два изображения)
Изображение 1: Постановка задачи

go
1 // Список хостов со статусами хранится в глобальной переменной с типом map[string]bool.
2 //
3 // Задание: реализовать систему проверки доступности (health check) HTTP urls (ключи мапы).
4 // http.Get(url)
5
6
7
8
9
10 // sadasdfasdfasdfasdfs dfasdf asdf
Изображение 2: Уточнение задачи и начало функции

go
1 // Список хостов со статусами хранится в глобальной переменной с типом map[string]bool.
2 //
3 // Задание: реализовать систему проверки доступности (health check) HTTP urls (ключи мапы).
4 // 200 - ок, остальное в ошибку. Если ошибка 3 раза, то очередь на retry
5 // client.Get(url)
6
7 var CacheHosts map[string]bool
8
9
10
11
12
13
14
15
16 func healthcheck() {
17     client := http.
18 }
Блок 2: Эволюция решения Health Check (Черновики)
Изображение 3: Первый вариант с циклом range 3

go
16     for {
17         // все таки надо RLock - будет все таки паника с fatal
18         for url, ok := range CacheHosts {
19             ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
20             defer cancel()
21
22             for range 3 {
23                 req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
24                 resp, err := client.Do(req)
25
26             }
27
28             req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
29             resp, err := client.Do(req)
30             if err != nil {
31                 for range 3 {
32                     req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
33                     resp, err := client.Do(req)
34                 }
35             }
36             resp.Body.Close()
37
38             if resp.StatusCode == http.StatusOK {
39                 if !ok {
40                     mu.Lock()
41                     CacheHosts[url] = true
42                     mu.Unlock()
43                 }
44             } else {
Изображение 4: Вариант с комментариями и break

go
     for {
         // все таки надо RLock - будет все таки паника с fatal error: concurrent map iteration and map write
         for url, ok := range CacheHosts {
             ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
             var response *http.Response
             
             for range 3 { // linter: magic number
                 req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
                 resp, err := client.Do(req)
                 cancel()
                 if err == nil {
                     break
                 }
                 resp.Body.Close()
             }
             
             if resp.StatusCode == http.StatusOK {
                 if !ok { // зачем ok? Нам же надо делать healthcheck
                     mu.Lock()
                     CacheHosts[url] = true
                     mu.Unlock()
                 }
             } else {
                 if ok { // зачем ok? Нам же надо делать healthcheck
                     mu.Lock()
                     CacheHosts[url] = false
                     mu.Unlock()
                 }
             }
Изображение 5: Оптимизированный вариант

go
             ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
             var resp *http.Response
             var err error
             
             for range 3 { // linter: magic number
                 req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
                 resp, err = client.Do(req)
                 cancel()
                 
                 if err == nil {
                     resp.Body.Close()
                     break
                 }
             }
             
             if err == nil && resp.StatusCode == http.StatusOK {
                 if !ok { // оптимизация
                     mu.Lock()
                     CacheHosts[url] = true
                     mu.Unlock()
                 }
             } else {
                 if ok { // оптимизация
                     mu.Lock()
                     CacheHosts[url] = false
                     mu.Unlock()
                 }
             }
         }
     }
Изображение 6: Финальная версия с RLock

go
             mu.RUnlock()
             if err == nil && resp.StatusCode == http.StatusOK {
                 if !ok { // оптимизация
                     mu.Lock()
                     if _,ok:= CacheHosts[url]; ok {
                         CacheHosts[url] = true
                         mu.Unlock()
                         mu.RLock()
                     }
                 }
             } else {
                 if ok { // оптимизация
                     mu.Lock()
                     CacheHosts[url] = false
                     mu.Unlock()
                 }
             }
Блок 3: Задача на партиционированный кэш
Изображение 7: Интерфейс кэша

go
1 // написать партиционированный кеш
2
3 type Cache interface {
4     Get(k string) int
5     Set(k string, v int)
6 }
7
8
9
10
11
12
Блок 4: Задача на сервис обработки видео
Изображение 8: Интерфейсы и требования

go
1 // необходимо реализовать сервис на go, где пользователь может скачать видео и выбрать формат видео ( с комментариями)
2 // Реализовать сервис обработки видео
3 // Обработку выполняет Processor
4 // На вход поступает запрос video []byte
5 // интерфейс обработки видео
6
7 type Processor interface {
8     Process([]byte) ([]byte, error) }
9
10 // персистентное хранилище данных
11 type Storage[T any] interface {
12
13 // Save создает/обновляет, при создании сам задает id и возвращает его
14 Save(ctx context.Context, item T) (id int, err error)
15
16 // Get возвращает полный Item по ID.
17 Get(ctx context.Context, id int) (item T, err error)
18
19 // Find ищет и возвращает массив
20 Find(ctx context.Context, q string) (items []T, err error) }
21












