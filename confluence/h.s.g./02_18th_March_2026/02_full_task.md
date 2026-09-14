Разработчик дал на ревью код своего нового кэша, нам необходимо провести код-ревью
Кэш будет использоваться под высокой нагрузкой в проде
Частота записи/чтения 20%/80% соответственно

Код:
package main

import (
    "fmt"
    "sync"
)

func main() {
    fmt.Println(GetOrCreate("hello", "world"))
    fmt.Println(Get("hello"))
}


var cache = make(map[string]string)


/*
БЫЛО:
// GetOrCreate проверяет существование ключа key
// Если такого нет, то создает новое значение
func Create(key, value string) string {
    var m sync.Mutex

    m.Lock()
    value, ok = cache[key]
    m.Unlock()

    if ok {
	return value
    }

    m.Lock()
    cache[key] = value
    m.Unlock() 
    
    return value
}


func Get(key string) string {
    var m sync.Mutex
    m.Lock()
    v := cache[key]
    m.Unlock()
    return v
}
*/

var m sync.RwMutex

// GetOrCreate REST: get, update, patch, upsert(insert/update) (get/insert)
// GetOrCreate проверяет существование ключа key
// Если такого нет, то создает новое значение
func Create(key, value string) string {
    m.RLock()
    defer m.RUnlock() 
    cache[key] = value
    return value
}



func Get(key string) (string, bool) {
    m.RLock()
    defer m.RUnlock()
    return cache[key]
}


17 грейд

Заметил что мьютекс — локальная переменная
Дал рекомендацию оформить это в структуру

18 грейд

Посоветовал использовать defer для мьютексов
Предложил переписать GetOrCreate чтобы лочить мьютекс только один раз

19 грейд senior

Предложил использовать RWMutex
Заметил, что два отдельных лока в GetOrCreate могут привести к конкурентной записи

20 грейд архитектор

Предложил использовать sync.Map, рассказал чем подход с RWMutex (а за одно и посомневался почему RWMutex может оказаться медленнее)
Предложил сделать разбить кеш на несколько шардов с отдельным мьютексом










// 2

Это почти реальный пример из одного из наших сервисов. 
Этот код - обертка над кешем, который соотвественно пишет и читает данные из кэша. 
Не смотря на то, что внедрение данного кэша должно было облегчить основное хранилище, однако это не произошло. Почему?

type Storage struct {
    cache *lru.Cache
}

func (s *Storage) Set(wh *warehouse.Warehouse) {
    s.cache.Put(wh.Id, *wh) // тут не нужна звездочка, но почему?
}

func (s *Storage) Get(id types.WarehouseId) *warehouse.Warehouse {
    item, ok := s.cache.Get(id)

    if ok {
        if wh, ok := item.(*warehouse.Warehouse); ok { // тут убрать звездочку
            return wh // &wh
        }
    }

    return nil
}

В кэш кладется `warehouse.Warehouse`, а assert делается с `*warehouse.Warehouse`.














Задача 3:

Что выведет код и почему?
!всё в го передается по значению! a b c 

package main

import "fmt"

func main() {
	lst := []string{"a", "b", "c", "d"} 
	
	for k, v := range lst {
		if k == 0 {
			lst = []string{"aa", "bb", "cc", "dd"}
		}
		fmt.Println(v) // должно было бы вывести "aa", "bb", "cc", "dd", "b", "c", "d", но мы в слайс можем добавлять только append() и поэтому вы увидим ошибку компиляции.
	}
}















Задача 4 (задача из бигтехов ОЗОН/ЯНДЕКС):

Нашей команде необходимо произвести интеграцию с внешним сервисом.
API у данного сервися платное, и поэтому нужно оптимизировать взаимодействие с внешним апи и ограничить количество Идентичных одновременных запросов к бупбличному API

type Dadata interface {
	GetAddressByFias(ctx context.Context, address string) (*AddressByFias, error)
}

type DadataAPIClient struct {
	mu sync.Mutex
	activeKeys map[string]*Result
	externalAPI Dadata
}

type Result struct {
	res *AddressByFias
	err error
	done chan struct{}
}

func (c *DadataAPIClient) GetAddressByFias(ctx context.Context, address string) (*AddressByFias, error) {
	mu.Lock()
	defer mu.Unlock()
	if result, ok := c.keys[address]; ok {
		<-result.done
		return result.res, result.err
	}

	keys[address] = Result{
		done: make(chan struct),
	}

	res, err := d.externalAPI.GetAddressByFias(ctx, address)

	c.activeKeys.res = res
	c.activeKeys.err = err

	close(c.keys[address].done)
	delete(c.keys[address])

	return res, err
}

// это типичная задача на кэш? Кто первый выполнился, тот отправляет остальным ответ, чтобы на 10 одинаковых пришел 1 ответ, и остальным 9 раздал. Мы так не 100 рублей потратим, а 10.
// кстати, зачем нам тут каналы и мьютексы одновременно, если каналы изначально потокобезопасные? Можно ли разобрать Rate limiter через Redis?
// задача на типичный сингл флайт?

И тут можно накидать вопросов:
distribution lock
lock contention
в чем разница между optimistic lock и pessimistic lock
хранение результата в краткосрочной и в долгосрочной перспективе
можно ли хранить пустые ответы
как инвалидировать кэш
как ограничить кол-во запросов к апи до 1 в момент времени (семафор на каналах)
про каналы и мьютексы базово










Задача 5 (НЕ РЕШЕНА)
/*
Дан срез указателей на int. Увеличьте на delta значения только по уникальным адресам 
(если несколько элементов ссылаются на один и тот же int, нужно инкрементировать его ровно один раз).
Верните обновленный слайс и два числа:
updated — сколько уникальных значений было изменено;
duplicates — сколько элементов среза оказались повторными ссылками на уже обновлённый адрес.
*/

func IncrementUniqueBy(nums []*int, delta int) ([]*int, int, int) {
    // ваш код
}

// Пример
a := 1
b := 10
nums := []*int{&a, nil, &b, &a, &a, &b} // &a и &b повторяются

sl, updated, duplicates := IncrementUniqueBy(nums, 3)

// a == 4, b == 13
// sl [4, nil, 13, 4, 4, 13]
// updated == 2
// duplicates == 3












// ЗАДАЧА 6: Once с каналами
//
// Реализуйте структуру once, функцию new и потокобезопасный метод do.
// Реализиция once и new должна использовать каналы, не используйте пакет sync.
// Функция new возвращает указатель на структуру once
// Метод do:
// - получает на вход функцию f
// - исполняет f только в том случае, если do вызывается в первый раз для этого экземпляра once. В противном случае
//   ничего не делает
//
// Функция main должна вывести call в консоль ровно один раз.

package main

import (
    "fmt"
    "sync"
)

const goroutinesNumber = 10

type once struct {
    ch1 chan struct{}
}

func new() *once {
    o := &once{
	ch1: make(chan struct{}, 1),
    }
    o.ch <- struct{}{}

    return o
}

func (o *once) do(f func()) {
    select {
	case <-o.ch:
		f()
	default:
    }
}

func funcToCall() {
    fmt.Printf("call")
}

func main() {
    wg := sync.WaitGroup{}
    so := new()

    wg.Add(goroutinesNumber)
    for i := 0; i < goroutinesNumber; i++ {
        go func(f func()) {
            defer wg.Done()
            so.do(f)
        }(funcToCall)
    }

    wg.Wait()
}




