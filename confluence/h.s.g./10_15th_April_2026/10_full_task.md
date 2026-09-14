/*
Нужно построить карту Web-сайта, те для каждой страницы получить список страниц, на которые она ссылается.
 
{
  "https://w.ru/": ["https://w.ru/support", "https://w.ru/news"],
  "https://w.ru/news": ["https://w.ru/support", "https://w.ru/"],
  "https://w.ru/support": [https://w.ru/news],
}

Для скачивания одной HTML страницы можно использовать готовую функцию Get.
Для разбора HTML и извлечения гиперссылок - готовую фукнцию ParseHTML.
Для определения домена - функцию ParseHostname.
 
Если на странице содержатся гиперссылки на другие сайты (домены) - их нужно игнорировать  (не посещать и не добавлять в карту сайта).
Если при скачивании страницы произошла ошибка - ее нужно залогировать и продолжить работу.
Предполагаем что объем сайта небольшой и карта полностью умещается в памяти.
*/

func Get(ctx context.Context, url string) (body string, err error)
 
func ParseHTML(body string) (urls []string)
 
func ParseHostname(url string) (hostname string)

type UrlMap struct {
    main string
    subsidiaries []string
}

func BuildSiteMap(ctx context.Context, startUrl string) (siteMap map[string][]string) {
    visited := make(map[string]bool,1)
    mainhost := ParseHostname(startURl)
    queue:= []string{startUrl}
    out := make(chan UrlMap,3)
    for len(queue)>0 {
        vistited[url]= true
        
        for _,url := range queue{
            go Worker(url,mainhost,out)
        }
        
        for result := range out{
            if !visited[result.subsidiaries]{
                queue = append(queue,result.subsidiaries)
            }
        }
    }
    return siteMap
}
// semaphore
func Worker(url string,main string,out chan){
    var result UrlMap
    result.main = url
    body, err := Get(ctx,url)
        if err != nil {
            println(err)
        }
    urls := ParseHTML(body)
    for _,url := range urls{
        if ParseHostname(url) == main{
            
            result.subsidiaries = append(result.subsidiaries,url)
        }
    }
    out<- result
}

cloud go
k8s, linux




at least once...
excatly once - key UUID
consumer lag
outbox
7 консьюмер и 3 партции, что будет с 4 консьюмерами?

k8s:
logging, container манипулировать, настраивать canary



