package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Order struct {
    Payload string
}

type Result struct {
    Status uint8
}

type usecase interface {
    CreateOrder(order Order) (Result, error)
}

type Request struct {
    UUID string //...
    Payload string `json:"payload"`
}

type Response struct {
    Message string `json:"message"`
    Error   string `json:"error,omitempty"`
}

type Handler struct {
    usecase usecase
}

func NewHandler(usecase usecase) *Handler {
    return &Handler{
        usecase: usecase,
    }
}

//интерфейс юзкейс можно менять если сильно надо
//юзкейс может работать до 10 секунд, надо чтобы ответ отдавался не дольше чем за 1 секунду
//при этом чтобы заказ продолжил создаваться в фоне, если он не успел за 1 секунду
//поменять интерфейс хендлера чтобы он возвращал информацию о создании заказа (например чтобы айдишник заказа возвращался)
//сделать так, чтобы фронт мог дергать другую ручку и спрашивать создался ли заказ уже или нет

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {    
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        _ = json.NewEncoder(w).Encode(Response{
            Error: fmt.Sprintf("method %s not allowed", r.Method),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")

    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(Response{
            Error: fmt.Sprintf("failed to decode request: %v", err),
        })
        return
    }

    if req.Payload == "" {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(Response{
            Error: "Payload is required",
        })
    }

    // тут долгий вызов.
    ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
    res := make(chan Result, 1)
    
    go func(){
        // UUID front request
        // IF UUID уже был ON CONFLICT DO NOTHING
        // 
        // result uint8 =  1,2,3 (1 = success, 2 = failed, 3 = in proccess)
        result, err := h.usecase.CreateOrder(Order{Payload: req.Payload})
        if err != nil {
            res <- Result{Status: false}
            /*
            w.WriteHeader(http.StatusInternalServerError)
            _ = json.NewEncoder(w).Encode(Response{
                Error: fmt.Sprintf("failed to process request: %v", err),
            })
            */
        }
        res <- Result{Status: true}
    }()
    
    select {
        case <-ctx.Done():
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(Response{
            Message: "Order procceed",
        })
        case re := <-res:
        if re.Status {
            w.WriteHeader(http.StatusOK)
            _ = json.NewEncoder(w).Encode(Response{
            Message: "Order created",
        })    
        }
        
     }  

    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(Response{
        Message: "request processed successfully",
    })
}

----------------

/*
Контекст
Необходимо реализовать HTTP-сервис на Go для фиксации и агрегации статистики кликов пользователей по публикациям авторов.
Сервис должен гарантировать корректный подсчёт уникальных взаимодействий в рамках календарных суток.

Регистрация клика:
При получении запроса сервис должен зафиксировать взаимодействие пользователя с автором.
Учитывается только уникальная пара (user_id, author_id) в рамках одних календарных суток.
Повторные клики от того же пользователя тому же автору в тот же день не должны увеличивать итоговый счётчик.

Получение статистики:
Эндпоинт должен возвращать агрегированные данные за предыдущие сутки (интервал 00:00–23:59 по UTC).
На вход подается слайс авторов по которым нужно вернуть метрики.
В ответе ожидается мапа авторов и количество уникальных пользователей, кликнувших по каждому из них за указанный период.
Если данных за дату нет, возвращается пустой объект или 200 OK с {}.
*/

type Result struct {
    user_id uuid
    created_at int // sec
}

Reg := map[author_id][]Result
Get := map[author_id]int

data map[string]map[string]map[string]struct{}

//us 1 2
//a 1 2 3

//us 1 : [1,1,2]; us 2 : [2]

/*
proto:
rpc CreateOrder()
make generatae
// grpc, http

*/

// Impl...
func CreateOrder() {}
// кол-во кешей
// user_id % кол-во кешей = 
//map[int]map[] партиционированный кеш 

-------------------------------------------------------
/*
Условие задачи:
Необходимо реализовать конкурентный поиск документов на серверах. 
Для этого у нас есть сторонняя библиотека с функцией которая осуществляет 
поиск документов на указанном сервере по указанному запросу.

У нас есть 3 идентичных сервера (реплики) и задача состоит в том, 
чтобы конкурентно вызвать эту функцию для всех серверов и вернуть первый успешный 
ответ от любого из серверов не дожидаясь ответов от других серверов.
*/

// q
// wb


// http.Get()

search.Search(server string, query string) ([]string, error) 

type Result struct {
    hashByte []string //
    err error
}

func foo(ctx context.Context, servers []string, query string) (Result, error) {
    ctx, cancel := context.WithCancel(ctx) //
    defer cancel() // отменит все горутины при возврате

    res := make(chan Result, len(servers))
    var wg sync.WaitGroup

    for _, ser := range servers {
        wg.Add(1)
        // 50go
        go func(ser string) {
            defer wg.Done()
            result := search.Search(ser, query)
            select {
            case res <- result:
            case <-ctx.Done():
            }
        }(ser)
    }
    go func() { wg.Wait(); close(res) }()

    for r := range res {
        // 1g
        if r.err == nil {
            return r, nil // defer cancel() сигналит остальным
        }
    }
    return Result{}, errors.New("all failed")
}

// 10, 10






Что происходит по шагам
После return r из главной горутины:

Локальная переменная res в foo исчезает, но сам канал жив — на него держат ссылку:

все 100 рабочих горутин (через замыкание),
горутина-закрывашка (go func() { wg.Wait(); close(res) }()).


Рабочие горутины продолжают крутить search.Search. Когда каждая закончит:

делает case res <- result — отправка не блокируется, в буфере есть место (буфер 100, отправок максимум 100).
вызывает wg.Done() и завершается.


Когда все 100 рабочих завершились, wg.Wait() разблокируется, выполняется close(res), закрывашка тоже завершается.
Все горутины мертвы → на канал больше никто не ссылается → GC собирает его вместе со всеми 99 непрочитанными Result внутри.

То есть буферизованные значения освобождаются автоматически, как только сам канал становится недостижимым.
Где здесь подвох
Утечки нет, но есть временный оверхед:

Пока хоть одна рабочая горутина не завершилась, канал жив и держит до 99 Result в памяти.
Если Result тяжёлый (например, содержит большой []byte с телом ответа), это 99 × размер тела висит в памяти до завершения самого медленного search.Search.
Плюс сами горутины со своими стеками висят до конца.

Так что формально — никакой утечки, программа корректна. Практически — после раннего return ты ещё какое-то время платишь памятью и сетью за работу, результат которой уже никому не нужен. Именно поэтому канонический паттерн — context.WithCancel + defer cancel(): он не убирает буфер (он и так соберётся), но даёт рабочим горутинам шанс выйти раньше через <-ctx.Done(), не дожидаясь окончания search.Search, если тот когда-нибудь станет отменяемым.
Что было бы настоящей утечкой
Если бы буфер был меньше числа горутин — например, make(chan Result, 10) на 100 серверов. Тогда после return 90 горутин навсегда зависнут на case res <- result (буфер забит, никто не читает, ctx.Done() не закрыт). Вот это утечка горутин — и канал тогда тоже никогда не соберётся, потому что эти 90 горутин держат на него ссылку вечно.


