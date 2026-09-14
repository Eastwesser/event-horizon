ЗАДАЧИ СЕРГЕЯ:

package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

/*
	https://ya.cc/t/N32jhpbz956niK

	Есть приложение с микросервисной архитектурой.
	Микросервис можно абстрагировать с помощью интерфейса Backend.
	Для доступа к одному экземпляру микросервиса можно использовать
	тип BackendImpl, который уже реализован.

	Для каждого микросервиса есть несколько десятков запущенных
	экземпляров, каждый из которых доступен по своему адресу addr.
	Однако отдельные экземпляры микросервиса ненадежны:
	они могут падать, быть недоступными либо перегруженными.

	Поэтому вам нужно реализовать тип Balancer, который также реализует
	интерфейс Backend и осуществляет client-side балансировку нагрузки
	между экземплярами микросервиса.
*/

type Request interface{}

type Response interface{}

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
}

var _ Backend = &BackendImpl{}

// addr содержит ip:port конкретного экземпляра
func NewBackend(addr string) *BackendImpl {
	return &BackendImpl{}
}

type Balancer struct {
	curr       int64
	backends   []Backend
	idxBackend int
	mu         sync.Mutex
}

var _ Backend = &Balancer{}

// addrs содержат адреса всех балансируемых экземпляров
func NewBalancer(addrs []string) *Balancer {
	backends := make([]Backend, len(addrs))

	for i, addr range addrs {
		    backends[i] = NewBackend(addr)
		}

	return &Balancer{
		    backends: backends,
		    idxBackend: len(backends)
	}
}


func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
	b.mu.Lock()
	b := backends[b.curr]
	curr := b.curr
	b.curr = (b.curr++) % b.idxBackend
	b.mu.Unlock()

	var resp Response
	var err error
	const maxRetry = 2

	// ДЗ - STRATEGY паттерн изать, чтобы в балансировку мы возвращали функцию, linked-list strategy

	// нужен ли тут бесконечный цикл? Если retry больше двух , то выходим,  или по контексту выходим? Через next а не через  current (через связный список нужно)
	for range maxRetry {
		err = nil
		if countRetry == maxRetry {
			return nil, errors.New("ошибка")
		}
		ctxt, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err = b.backends[curr].Invoke(ctxt, req)
		if err == nil {
			break
		}

		curr = (b.curr++) % b.idxBackend
	}

	return resp, err
}


Короче там раунд робин был, и псевдокод. А мне нужны знания, подходы, как можно проще.


https://rutracker.org/forum/viewtopic.php?t=6769958


Напишите функцию, которая проверяет, является ли переданная строка палиндромом

func IsPalindrome(s string) bool {
    runeStr:=[]rune(s)
    right:=len(runeStr)- 1
    left:=0
    for right > left{
        if runeStr[left] != runeStr[right] {
            return false
        } 
        left++
        right--
    }
    return true    
}



func IsPalindromeRecurce(s string) bool {
    if len(s) <= 1 {
        return true
    }
    
    runeStr:=[]rune(s)
    right:=len(runeStr)- 1
    left:=0
    for right > left{
        if runeStr[left] != runeStr[right] {
            return false
        } 
        left++
        right--
    }
    return true    
}
//recursive, non-recursive
 
// Пример
IsPalindrome("level")   // true
IsPalindrome("levvel")   // true
IsPalindrome("Go")      // false
IsPalindrome("あいいあ") // true
// le    G    
// lev   O
if char == 1 true
if char == 0 true
left != rigth false

// lifehack, если на алгосах такая задача, то можем выбрать любой язык

func IsPalindromeRecurce(s string) bool {
    return s === s.spit("").reverse().join("")
    
    rev := ""
    chars := []rune(s)
    
    for i = len(s); i > 0; i-- {
        rev = rev + string(chars[i])
    }
    
    return rev == s
}


---------------------------

/*
1. Монолиты vs микросервисы


*/



/*
Есть приложение с микросервисной архитектурой.
Микросервис можно абстрагировать с помощью интерфейса Backend.
Для доступа к одному экземпляру микросервиса можно использовать
тип BackendImpl, который уже реализован.

Для каждого микросервиса есть несколько десятков запущенных
экземпляров, каждый из которых доступен по своему адресу addr.
Однако отдельные экземпляры микросервиса ненадежны:
они могут падать, быть недоступными либо перегруженными.
Поэтому вам нужно реализовать тип Balancer, который также реализует
интерфейс Backend и осуществляет client-side балансировку нагрузки
между экземплярами микросервиса.
*/

type Request interface{}

type Response interface{}

type Backend interface {
  Invoke(ctx context.Context, req Request) (Response, error)
}

var _ Backend = &BackendImpl{}

// addr содержит ip:port конкретного экземпляра
func NewBackend(addr string) *BackendImpl {
    return &BackendImpl{}
}

type Balancer struct {
  // TODO
  curr int64
  backends []Backend
  idxBackend int
  mu sync.Mutex
  countRetry int64
  
}

var _ Backend = &Balancer{}

// addrs содержат адреса всех балансируемых экземпляров
func NewBalancer(addrs []string) *Balancer {
  backends := make([]Backend, len(addrs))
  
  for i, addr range addrs {
      backends[i] = NewBackend(addr)
  }
  
  return &Balancer{
      backends: backends,
      idxBackend: len(backends)
  }
}

var retry
// linked-list
//strtagy

func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
    b.mu.Lock()
    b := backends[b.curr]
    curr := b.curr
    b.curr = (b.curr++) % b.idxBackend
    b.mu.Unlock()
    // if we chose next: start = (b.next, 1) - 1 % len(backed)
    // for i := 0, i < len(b.backends) { idx := start+i % len(backed) b.backends[idx].Invoke(ctxt, req)}
    
    var resp Response
    var err error
    const maxRetry = 2 // for one backend or for full backends
    
    for range naxRetry {
        err = nil
        if countRetry == maxRetry {
            return nil, errors.New("ошибка")
        }
        ctxt, cancel := context.WithTimeout(ctx, 3 * time.Second)
        defer cancel()
        
        if err := ctx.Err(); err != nil {
           return nil, err
        }

        resp, err = b.backends[curr].Invoke(ctxt, req)
        if err == nil {
            break
        }
        
        curr = (b.curr++) % b.idxBackend
    }
    
    return resp, err
}






 ===

 Блок 1: Постановка задачи и интерфейсы
Текст задачи:

Сергей Дорофеев на главном экране
18 Есть приложение с микросервисной архитектурой.
19 Микросервис можно абстрагировать с помощью интерфейса Backend.
20 Для доступа к одному экземпляру микросервиса можно использовать тип BackendImpl, который уже реализован.
21
22 Для каждого микросервиса есть несколько десятков запущенных экземпляров, каждый из которых доступен по своему адресу addr.
23 Однако отдельные экземпляры микросервиса ненадежны:
24 они могут падать, быть недоступными либо перегруженными.
25 Поэтому вам нужно реализовать тип Balancer, который также реализует интерфейс Backend и осуществляет client-side балансировку нагрузки между экземплярами микросервиса.
26
27 */
28
29 type Request interface{}
30
31 type Response interface{}
32
33 type Backend interface {
34 Invoke(ctx context.Context, req Request) (Response, error)
35 }
36
37 var _ Backend = &BackendImpl{}
38
39 // addr содержит ip:port конкретного экземпляра
40 func NewBackend(addr string) *BackendImpl
41
42 type Balancer struct {
43 // TODO
44 }
45
46 var _ Backend = &Balancer{}
47
48 // addrs содержат адреса всех балансируемых экземпляров
49 func NewBalancer(addrs []string) *Balancer {
50 // TODO
51 }
52
round robin
+ резко поменять стратегию в моменте

Блок 2: Первые варианты реализации (черновики и ошибки)
Фрагмент 1 (попытка реализации Invoke):

go
77 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
78     b.mu.Lock()
79     b := backends[b.curr]
80     curr := b.curr
81     b.curr = (b.curr++) % b.idxBackend
82     b.mu.Unlock()
83     // if we chose next: start = (b.next, 1) - 1 % len(backed)
84     // for i := 0, i < len(b.backends) { idx := start+i % len(backed) b.backe
85
86     var resp Response
87     var err error
88     const maxRetry = 2 // for one backend or for full backends
89
90     for range maxRetry {
91         err = nil
92         if countRetry == maxRetry {
93             return nil, errors.New("ошибка")
94         }
95         ctx, cancel := context.WithTimeout(ctx, 3 * time.Second)
96         defer cancel()
97
98         if err := ctx.Err(); err != nil {
99             return nil, err
100        }
101
102        resp, err = b.backends[curr].Invoke(ctx, req)
103        if err == nil {
104            break
105        }
106
107        curr = (b.curr++) % b.idxBackend
108    }
109
110    return resp, err
111 }
Фрагмент 2 (структура и NewBalancer):

go
59 // addrs содержат адреса всех балансируемых экземпляров
60 func NewBalancer(addrs []string) *Balancer {
61     backends := make([]Backend, len(addrs))
62
63     for i, addr := range addrs {
64         backends[i] = NewBackend(addr)
65     }
66
67     return &Balancer{
68         backends: backends,
69         idxBackend: len(backends)
70     }
71 }
72 // linked-list
73
74
75 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
76     b.mu.Lock()
77     b := backends[b.curr]; b.curr = (b.curr++) % b.idxBackend
78     b.mu.Unlock()
79     // if we chose next: start = (b.next, 1) - 1 % len(backed)
80     // for i := 0, i < len(b.backends) { idx := start+i % len(backed) b.backends[idx].
81     err = nil
82
83     resp, err := b.Invoke(ctx, req)
84
85     b.mu.Lock()
86     if b.curr == b.idxBackend {
87         b.curr = 0
88     } else {
89         b.curr = b.curr + 1
90     }
91     b.mu.Unlock()
92
93     return resp, err
94 }
Блок 3: Эволюция решения (Кольцевой буфер)
Кольцевой буфер (базовая структура):

go
30 */
31
32 type Request interface{}
33
34 type Response interface{}
35
36 type Backend interface {
37     Invoke(ctx context.Context, req Request) (Response, error)
38 }
39
40 var _ Backend = &BackendImpl{}
41
42 // addr содержит ip:port конкретного экземпляра
43 func NewBackend(addr string) *BackendImpl
44
45 type Balancer struct {
46     // TODO
47 }
48
49 var _ Backend = &Balancer{}
50
51 // addrs содержат адреса всех балансируемых экземпляров
52 func NewBalancer(addrs []string) *Balancer {
53     // TODO
54 }
55
56
Реализация с мьютексом (черновик):

go
48 // TODO
49 curr int64
50 backends []Backend
51 mu sync.Mutex
52
53 }
54
55 var _ Backend = &Balancer{}
56
57 // addrs содержат адреса всех балансируемых экземпляров
58 func NewBalancer(addrs []string) *Balancer {
59     backends := make([]Backend, len(addrs))
60
61     for i, addr := range addrs {
62         backends[i] = NewBackend(addr)
63     }
64
65     return &Balancer{
66         backends: addrs,
67     }
68 }
69
70 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
71     b.mu.Lock()
72     defer b.mu.Unlock()
73
74     b := backends[b.curr]
75
76     return b.Invoke(ctx, req)
77 }
Реализация с индексом (Round Robin):

go
58 // addrs содержат адреса всех балансируемых экземпляров
59 func NewBalancer(addrs []string) *Balancer {
60     backends := make([]Backend, len(addrs))
61
62     for i, addr := range addrs {
63         backends[i] = NewBackend(addr)
64     }
65
66     return &Balancer{
67         backends: backends,
68         idxBackend: len(backends) - 1
69     }
70 }
71
72 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
73     b.mu.Lock()
74     b := backends[b.curr]
75
76     if b.curr == b.idxBackend {
77         b.curr = 0
78     } else {
79         b.curr = b.curr + 1
80     }
81     b.mu.Unlock()
82
83     resp, err := b.Invoke(ctx, req)
84
85     return resp, err
86 }
Блок 4: Финальные версии с ретраями и таймаутами
Вариант с циклом range 2:

go
71 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
72     b.mu.Lock()
73     b := backends[b.curr]
74     curr := b.curr
75
76     if b.curr == b.idxBackend {
77         b.curr = 0
78     } else {
79         b.curr = b.curr + 1
80     }
81     b.mu.Unlock()
82
83     var resp Response
84     var err error
85
86     for range 2 {
87         ctx, cancel := context.WithTimeout(ctx, 3 * time.Second)
88         defer cancel()
89
90         if err := ctx.Err(); err != nil {
91             return err
92         }
93
94         resp, err = b.Invoke(ctx, req)
95         if err == nil {
96             break
97         }
98     }
99
100    return resp, err
101 }
Вариант с countRetry и maxRetry:

go
75 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
76     b.mu.Lock()
77     b := backends[b.curr]
78     curr := b.curr
79
80     if b.curr == b.idxBackend {
81         b.curr = 0
82     } else {
83         b.curr = b.curr + 1
84     }
85     b.mu.Unlock()
86
87     var resp Response
88     var err error
89     var countRetry int
90     const maxRetry = 2
91
92     for {
93         if countRetry == maxRetry {
94             return nil, errors.New("ошибка")
95         }
96         ctx, cancel := context.WithTimeout(ctx, 3 * time.Second)
97         defer cancel()
98
99         if err := ctx.Err(); err != nil {
100            return err
101        }
102
103        resp, err = b.Invoke(ctx, req)
104        if err != nil {
105            countRetry++
106        }
107    }
108
109    return resp, err
110 }
Вариант с перебором всех бэкендов (for i := 0; i < len(b.backends)):

go
74 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
75     b.mu.Lock()
76     b := backends[b.curr]
77     curr := b.curr
78     //b.curr = (b.curr + 1) % b.idxBackend + 1
79
80     if b.curr == b.idxBackend {
81         b.curr = 0
82     } else {
83         b.curr = b.curr + 1
84     }
85     b.mu.Unlock()
86
87     var resp Response
88     var err error
89     var countRetry int
90     const maxRetry = 2
91
92     for i := 0; i < len(b.backends) {
93         if countRetry == maxRetry {
94             return nil, errors.New("Ошибка")
95         }
96         ctx, cancel := context.WithTimeout(ctx, 3 * time.Second)
97         defer cancel()
98
99         if err := ctx.Err(); err != nil {
100            return err
101        }
102        idx :=
103
104        resp, err = b.backends[idx].Invoke(ctx, req)
105
106        if err != nil {
107            countRetry++
108        }
109    }
110 }
Последний фрагмент (смесь подходов):

go
53
54 }
55
56 var _ Backend = &Balancer{}
57
58 // addrs содержат адреса всех балансируемых экземпляров
59 func NewBalancer(addrs []string) *Balancer {
60     backends := make([]Backend, len(addrs))
61
62     for i, addr := range addrs {
63         backends[i] = NewBackend(addr)
64     }
65
66     return &Balancer{
67         backends: backends,
68         idxBackend: len(backends) - 1
69     }
70 }
71
72 func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
73     b.mu.Lock()
74     b := backends[b.curr] //b.curr = (b.curr + 1) % b.idxBackend + 1
75     b.mu.Unlock() i := 0; i < len(b.backends)
76         idx :=
77
78     resp, err := b.Invoke(ctx, req)
79
80     b.mu.Lock()
81     if b.curr == b.idxBackend {
82         b.curr = 0
83     } else {
84         b.curr = b.curr + 1
85     }
86     b.mu.Unlock()
87
88     return resp, err
89 }



======


Изображение 1: Основные реализации
go
1 Напишите функцию, которая проверяет, является ли переданная
2
3 func IsPalindrome(s string) bool {
4     runeStr := []rune(s)
5     right := len(runeStr) - 1
6     left := 0
7     for right > left {
8         if runeStr[left] != runeStr[right] {
9             return false
10        }
11        left++
12        right--
13    }
14    return true
15 }
16
17
18
19 func IsPalindromeRecurce(s string) bool {
20    if len(s) <= 1 {
21        return true
22    }
23
24    runeStr := []rune(s)
25    right := len(runeStr) - 1
26    left := 0
27    for right > left {
28        if runeStr[left] != runeStr[right] {
29            return false
30        }
31        left++
32        right--
33    }
34    return true
35 }
36 //recursive, non-recursive
37
38 // Пример
39 IsPalindrome("level")    // true
40 IsPalindrome("level")    // true
41 IsPalindrome("Go")       // false
42 IsPalindrome("аїиаїа")   // true
43 // le   G
44 // lev  0
45 if char == 1 true
46 if char == 0 true
47 left != right false
48
49 func IsPalindromeRecurce(s string) bool {
50     chars := []rune(s)
51
52     recur := func(idx int) {
53         chars[idx]
54     }
55
56     if recur(0) != recur(len(s)-1) {
57         return false
58     }
59
60     return true
61 }
62
63
64 ----------------------------------------
Изображение 2: Альтернативный вариант от пользователя
Артём Булатов:

go
func isPalindrome(str *string) bool {
    rs := *str
    stlen := len(rs)
    for i := 0; i < stlen/2; i++ {
        if rs[i] != rs[stlen-i-1] {
            return false
        }
    }
    return true
}
Изображение 3: Фрагмент с примером и рекурсией
(Этот фрагмент дублирует часть первого изображения, но с небольшими отличиями в форматировании и комментариях)

go
37 // Пример
38 IsPalindrome("level")    // true
39 IsPalindrome("level")    // true
40 IsPalindrome("Go")       // false
41 IsPalindrome("аїиаїа")   // true
42 // le   G
43 // lev  0
44 if char == 1 true
45 if char == 0 true
46 left != right false
47
48 func IsPalindromeRecurce(s string) bool {
49     chars := []rune(s)
50
51     recur := func(idx int) {
52         chars[idx]
53     }
54
55     if recur(0) != recur(len(s)-1) {
56         return false
57     }
58
59     return true
60 }
61
62
63
64 ----------------------------------------


