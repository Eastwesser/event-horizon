Тебе на код-ревью пришёл код от нового разработчика в команде. 
Проведи код-ревью в обучающем формате, т.е. объясни каждое замечание в коде.
Можешь давать замечания сразу, а можно сначала пролистать весь код.

main.go:

package main

import (
    "database/sql"
    "fmt"
    "io"
    "net/http"
    "os"
    "strconv"
    "strings"

    _ "github.com/lib/pq"
)


// ТЗ: мы хотим собирать информацию по курсам валют в разных банках.
// Требуется написать программу, которая каждую минуту будет отправлять запрос в банк и получать курс нескольких валют и сохранять результат в БД.
// Банков может быть несколько.

func main() {

    if len(os.Args) == 2 {
        cmdName := os.Args[1]
        if cmdName == "help" { // ⭐️ харкод
            fmt.Println("Usage is './currency update'")
                // ⭐️ спросить лучшие практики cli
        // название программы -> ресурс/объект с которым действие (существительное) -> действие (глагол)
        // древовидная структура ключей
        // graceful shutdown и таймауты на контекстах (можно развить про наследования, как дедлайны себя при этом ведут особенно)
        } else if cmdName == "update" {   // ⭐️⭐️⭐️ предложил вынести код для update в отдельную функцию/файл, чтобы main не содержал бизнес логику

            urlsBank := []struct { // ⭐️⭐️⭐️ сделать конструктор
                bankName string
                curFrom  string
                curTo    string
                url      string
            }{
                {
                    bankName: "Bank 1",
                    curFrom: "RUB",
                    curTo:   "USD",
                    url:     "http://bank.example.com/rates/rub-usd",
                },
                {
                    bankName: "Bank 2",
                    curFrom: "RUB",
                    curTo:   "USD",
                    url:     "http://bank2.example.com/rates?currFrom=RUR&currTo=USD",
                },
            }

            clientBank := &http.Client{}

            for _, url := range urlsBank { 
                
                // ⭐️⭐️⭐️ Блок с http запросом и получением курса валюты нужно вынести в отдельную функцию
                req, _ := http.NewRequest(http.MethodGet, url.url, nil) // ⭐️⭐️⭐️

                if url.bankName == "Bank 2" {
                    req.Header.Add("Authorization", "auth_token=\"XXXXXXX\"") // ⭐️⭐️⭐️ токен нужно отдельно хранить
                }

                resp, err := clientBank.Do(req)

                if err != nil {
                    panic(err) // ⭐️⭐️⭐️ скажет про обработку ошибок (как проверять тип ошибки, враппинг и анвраппинг)️
                }
                defer resp.Body.Close()
                // ⭐️⭐️⭐️ дефер закрывать не в фор что (можно потом обратить внимание: есть ли ошибки?)
                body, _ := io.ReadAll(resp.Body) // ⭐️⭐️⭐️ нет проверки статуса ответа и контент тайпа

                strBody := string(body)

                if url.bankName == "Bank 1" {
                    strBody = strings.ReplaceAll(strBody, ",", ".") // Заменяем для Банка 1 запятую на точку 
                } // ⭐️⭐️ бесполезный комментарий, надо писать зачем код что-то делает, а не что делает.

                value, err := strconv.ParseFloat(strBody, 64)
                if err != nil {
                    panic(err)
                }
                // ⭐️⭐️ баг - два раза используется curFrom
                // ⭐️⭐️ вставлять в БД можно батчем
                err = updateCurrency(url.bankName, url.curFrom, url.curFrom, value)
                if err != nil {
                    panic(err)
                }
            }
        }    
    } else { // ⭐️⭐️⭐️ должен сказать про guard clause, т.е. вынести это наверх для быстрого завершения программы
        fmt.Println("Usage is './currency update'")
    }
}

// ⭐️⭐️⭐️ Кандидат должен сказать, что логин/пароль соединения с БД не должны храниться в файле с кодом.
// В идеале - не должны храниться в репозитории приложения. Если назовёт vault (или аналогичное), то ещё ⭐️⭐️
// Если расскажет про приоритет применения переменных (аргументыVSконфиг_файлVSпеременные окружения), например, чтобы локально тестировать, то ещё ⭐️const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = "<password>" 
    dbname   = "<dbname>"
)

// ⭐️ 
func updateCurrency(bank, from, to string, value float64) error { 
    psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

    db, err := sql.Open("postgres", psqlconn) // ⭐️⭐️⭐️ не нужно подключаться к БД на каждый запрос, вынести подключение отдельно от выполнения запроса
    CheckError(err)

    defer db.Close()

    err = db.Ping()
    CheckError(err)

    fmt.Println("Connected!") // ⭐️⭐️⭐️ Это надо писать в логи
    // Ещё ⭐️⭐️, если расскажет про использование контекста и передачу в нём кастомного логгера

    insertStmt := fmt.Sprintf(`insert into currency_rates ("bank", "from", "to", "value") values('%s', '%s', '%s', '%.2f')`, bank, from, to, value) // ⭐️⭐️⭐️ инъекции
    _, err = db.Exec(insertStmt)
    return err
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

3. Замечания можно поделить на три группы и за них выдаются баллы (звёздочки).
Критичное замечание, которое кандидат обязательно должен назвать - ⭐️⭐️⭐️
Замечание, которое не так сильно лежит на поверхности, но тоже важно. Например, если кандидат нашёл баг - ⭐️⭐️
Замечание, которое полезно, но точно не самое ожидаемое от кандидата. Например, "Есть ещё более быстрая http-библиотека" - ⭐️


