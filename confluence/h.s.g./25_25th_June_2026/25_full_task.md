package main

import (
	"fmt"
	"net/http"
	"time"
)

// ТРЕБОВАНИЯ:
// 1. Опросить все URL параллельно.
// 2. Собрать результаты (статус-код ответа или ошибку сетевого запроса).
// 3. Вывести результаты в консоль строго в том порядке, в котором URL идут в исходном слайсе.


type Response struct{
    url string 
    status int
    err error
}

func process (client *http.Client, url string) Response {
    r:=Response{
        url: url,
    }
    resp, err:= client.Get(url)
    if err !=nil{
        r.err=err// не плохо было бы обернуть ошибку в errorf("%w",err)- чтобы знать что за ошибка
        return r
    }
    defer resp.Body.Close()
    
    status:=resp.StatusCode
    r.status=status
    
    return r
    
    
    resp, err:= client.Get(url)
    if err != nil {
        return Response{url: url, err: err}
    }
    
    return Response{url: url, status: resp.StatusCode}
    
  
    
}

func main() {
    var wg sync.WaitGroup
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://httpbin.org/delay/2", // этот сайт отвечает долго
		"https://yandex.ru",
		"https://httpbin.org/status/404",
	}
	

	client:=&http.Client{
	    Timeout: 5*time.Second,
	}
    
    resp:=make([]Response,len(urls))
    
    for i,url:=range urls{
        wg.Add(1)
        go func(){
            defer wg.Done()
            r:=process(client, url)
            resp[i]=r
            resp[i] = process(client, url)
        }()
    }
    wg.Wait()
    fmt.Println(resp)
    
    
    for v := range resCh{
        resp[v.index] = v
    }
    
    for _, res := range resp {
        if res.err != nil{
            
        } else {
            
        }
    }
}

===

i*i в цикле 
 
это корень из логарифма в квадрате сложность

===





