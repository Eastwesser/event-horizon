# MARSHALING TASKS

## Unmarshal в map — когда структура неизвестна

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
)

func main() {
    jsonData := `{"name":"Bob","age":28,"city":"NY"}`

    var result map[string]interface{}
	
    err := json.Unmarshal([]byte(jsonData), &result)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Parsed map:", result)
    fmt.Println("Name:", result["name"])
    fmt.Println("Age:", result["age"])
}
```
------------
## Декодинг из HTTP — потоковый разбор

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type Post struct {
    UserID int    `json:"userId"`
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

func main() {
    resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    var post Post
    err = json.NewDecoder(resp.Body).Decode(&post)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Parsed post: %+v\n", post)
}
```

# Что важно сказать на собеседовании:

- Теги json:"field" — задают имя поля в JSON
- omitempty — не включает поле, если оно пустое (zero value)
- Unmarshal требует указатель на структуру (&user)
- json.Decoder удобен для потокового чтения (например, из http.Response.Body)
- Для неизвестной структуры можно использовать map[string]interface{}, но это менее типобезопасно

-------------------------
## Type assertion — как достать значения из interface{}

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
)

func main() {
    // Пришла какая-то инфа, структура неизвестна
    jsonString := `{
        "name": "Alex",
        "age": 30,
        "skills": ["go", "python"],
        "meta": {"city": "Moscow", "active": true}
    }`

    var data map[string]interface{}
	
    err := json.Unmarshal([]byte(jsonString), &data)
    if err != nil {
        log.Fatal(err)
    }

    // Работаем с распарсенной хуйнёй
    fmt.Println("Parsed:", data)
    
    // Достаем поля с проверкой типа
    if name, ok := data["name"].(string); ok {
        fmt.Println("Name:", name)
    }
    
    if age, ok := data["age"].(float64); ok { // числа в JSON всегда float64
        fmt.Println("Age:", int(age))
    }
    
    if skills, ok := data["skills"].([]interface{}); ok {
        fmt.Println("Skills:", skills)
    }
}
```

# Что важно сказать на собеседовании:

- map[string]interface{} — универсальный способ разобрать JSON неизвестной структуры
- Числа приходят как float64, даже если в JSON они целые
- json.RawMessage позволяет отложить парсинг, когда тип зависит от какого-то поля (например, type)
- json.Decoder удобен для потоков (HTTP, файлы) — не надо читать весь байтовый слайс в память
- Проверка типов через type assertion (.(string), .(float64)) обязательна, чтобы не паниковать
