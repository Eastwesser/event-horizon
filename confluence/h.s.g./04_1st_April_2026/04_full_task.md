/*// Задача:
// Реализовать rate limiter по IP-адресам на основе алгоритма token bucket.
//
// Условия:
// 1. Для каждого IP должен храниться отдельный бакет с количеством доступных токенов.
// 2. Метод CanAccept(ip string) bool должен определять, можно ли принять очередной запрос от указанного IP.
// 3. Если токены есть, запрос разрешается и один токен списывается.
// 4. Если токенов нет, запрос должен быть отклонён.
// 5. Токены должны восстанавливаться со временем до заданного максимума.
// 6. Решение должно быть потокобезопасным.
// 7. Для неактивных IP должна быть предусмотрена периодическая очистка устаревших бакетов из памяти
map[ip]Bucket
Bucket{token //}
*/

type backet struct{
    token float64 // кол-во запросов на ip адрес
    lastTime time.time // последнее время обращения
}


type RateLimmiter struct{
    rate float64 //кол-во запросов в секунду
    maxTokens float64 // ограничитель максимального числа
    mu sync.Mutex
    backets map[strin] *backet
}



func NewRateLimmiter(rate, maxTokens float64 ) *RateLimmiter{
    rl:=&RateLimmiter {
        backets : make(map[string]*backet),
        rate :rate,
        maxTokens: maxTokens,
        
    } 
    
    return rl
}

// CanAccept - разрешено ли ip адресу делать запрос
func (rl *RateLimmiter)CanAccept(ip string)bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    now := time.now()
    b, exists := rl.backets[ip]
    if !exists {
        rl.backets[ip] = &backet{
            token: rl.maxTokens - 1
            lastTime : now
        }
        return true
    }
    
    duration = now.Sub(b.lastTime) // промежуток между последним обращением и текущим временем
    addTokens := duration.Seconds() * rl.rate // добавляем токены
    
    b.token = min(b.token + addTokens, maxTokens) // 
    b.lastUpdate = now // если разрешено

    if b.token >= 1 {
        b.token--
        return true
    }        
        
return false        
        
}







// Найти все ошибки
package main

import (
  "bytes"
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "sync"
  _ "github.com/lib/pq"
)

type Record struct {
  Id   int    json:"id" // ``
  Name string json:"name" // ``
}

func main() {
  db, err := sql.Open("postgres", "postgresql://admin:securepass@db-server:5432/appdatabase") // vault, env.prod
  if err != nil {
    log.Println("Database connection error:", err) // errorf
    return
  }
  defer db.Close()

  ctx := context.TODO() //
  var wg sync.WaitGroup

  records := []Record{{1, "John"}, {2, "Jane"}, {3, "Mike"}}

  wg.Add(len(records))
  for _, r := range records {
    go func() {
        defer wg.Done()
        
      r.name = "Updated " + r.name
      statement := fmt.Sprintf("INSERT INTO records (title) VALUES ('%s')", r.name) // sql injection
      _, err := db.Exec(statement)
      if err != nil {
        log.Println("Database insert error:", err) //
      }
    }()
  }

  wg.Wait()

  jsonData, err := json.Marshal(records)
  if err != nil {
    log.Println("JSON encoding error:", err)
    return
  }

  var buffer bytes.Buffer
  buffer.Write(jsonData)

  request, err := http.NewRequestWithContext(ctx, "POST", "api.example.com/endpoint", &buffer)
  if err != nil {
    log.Println("Request creation error:", err)
    return
  }

  response, err := (&http.Client{}).Do(request)
  if err != nil {
    log.Println(err)
    return
  }
  
  response.Body.Close()

  log.Println(io.ReadAll(response.Body))
}






// ЗАДАЧА 50: Что выведет?
func ax() {
    x := make([]int, 5, 10)  // [0,0,0,0,0]0,0,0,0,0 len = 5, cap = 10
    for i := range x {
        x[i] = i + 1 // [1,2,3,4,5]0,0,0,0,0 len = 5, cap = 10
    }
    fmt.Println(x, len(x)) // [1,2,3,4,5], len = 5
    y := x[1:3:4] // [2,3],3,4 cap = 4
    fmt.Println(y) // [2,3]
    fmt.Println(len(y), cap(y)) // 2, 4
    y = append(y, 100) // [2,3,100], 4
    fmt.Println(x) // [1,2,3,100,5]
    fmt.Println(y) // [2,3,100]
}














// 0,0,0,0,0,1,2,3,4,5 len = 10, cap = 10, 










        
    
    
    
    




















