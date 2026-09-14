,Do()

package main

import(
  "net/http"
  "encoding/json"
  "fmt"
  "time"
)
// ориг.юзер
type OriginalUser struct {
    ID       int     json:"id"
    Email    string  json:"email"
    Amount   float64 json:"amount"
    Profile  struct {
        Avatar     string json:"avatar"
        LastName   string json:"lastName"
        FirstName  string json:"firstName"
        StaticData string json:"staticData"
    } json:"profile"
    Password   string json:"password"
    Username   string json:"username"
    CreatedAt  string json:"createdAt"
    CreatedBy  string json:"createdBy"
}

// мод.юзер
type ModifiedUser struct {
    ID      int     json:"id"
    Email   string  json:"email,omitempty"
    Amount  float64 json:"amount"
    Profile struct {
        Avatar    string json:"avatar,omitempty"
        LastName  string json:"lastName,omitempty"
        FirstName string json:"firstName,omitempty"
    } json:"profile"
    Username  string json:"username,omitempty"
    CreatedAt string json:"createdAt"
    CreatedBy string json:"createdBy"
}

func userHandler(w http.ResponseWriter, r *http.Request){
  client:= http.Client{Timeout: 10 * time.Second}
  resp,err:=client.Get("http://83.136.232.77:8091/users")
  if err!=nil{
    http.Error(w,"failed",http.StatusInternalServerError)
    return
  }
  defer resp.Body.Close()

  var originals []OriginalUser

  if err:=json.NewDecoder(resp.Body).Decode(&originals);err!=nil{
    http.Error(w,"failed decode",http.StatusInternalServerError)
    return
  }

  var modifiedUsers []ModifiedUser

  for _,original:=range originals{
    var modified ModifiedUser
     modified.ID = original.ID
  modified.Amount = original.Amount
  modified.CreatedAt = original.CreatedAt
  modified.CreatedBy = original.CreatedBy

  if original.Amount <= 50000{
    modified.Email = original.Email
    modified.Username = original.Username
    modified.Profile.FirstName = original.Profile.FirstName
      modified.Profile.LastName = original.Profile.LastName
       modified.Profile.Avatar = original.Profile.Avatar
  }
  modifiedUsers= append (modifiedUsers,modified)
  }


  w.Header().Set("Content-type","application/json")
  json.NewEncoder(w).Encode(modifiedUsers)

}

func main(){
  http.HandleFunc("/users",userHandler)
    fmt.Println("Server running on 8081")
  if err:=http.ListenAndServe(":8081",nil);err!=nil{
    fmt.Println("error server",err)
  }

}


=======

2. Задача на Go (волчара позорная проверяет 🐺)
Ты сказала: если не решу — значит накрутчик опыта, и я сам себе назначаю цену.

Хорошо. Я дам честную задачу уровня middle Go-разработчика, не олимпиадную, а из реальной жизни. В ней нет подвоха в синтаксисе, только логика и знание стандартной библиотеки.

Таймер запущен, когда ты начнёшь читать условие.

Задача: «Асинхронный обработчик с приоритетами»
Ты разрабатываешь систему, которая принимает задачи (каждая задача — это функция, возвращающая ошибку, и приоритет от 1 до 5, где 5 — самый высокий). Нужно написать WorkerPool, который:

Запускает N воркеров (передаётся при создании).

Принимает задачи через метод Add(task func() error, priority int).

Выполняет задачи не в порядке очереди, а по приоритету (сначала все priority=5, потом 4, и т.д.). Внутри одного приоритета — FIFO (кто раньше добавился, тот раньше выполняется).

Если задача возвращает ошибку — она не должна прерывать работу пула, но ошибку нужно записать в лог (просто вывести в fmt.Println).

Должен быть метод Stop() — он останавливает приём новых задач, дожидается выполнения всех уже добавленных, и завершает воркеры.

Пул должен быть потокобезопасным (вызов Add из разных горутин — ок).

Пример использования (ожидаемое поведение)
go
func main() {
    pool := NewWorkerPool(3) // 3 воркера

    pool.Add(func() error {
        time.Sleep(100 * time.Millisecond)
        fmt.Println("Задача 1 (приор 1)")
        return nil
    }, 1)

    pool.Add(func() error {
        fmt.Println("Задача 2 (приор 5)")
        return nil
    }, 5)

    pool.Add(func() error {
        fmt.Println("Задача 3 (приор 5)")
        return nil
    }, 5)

    pool.Add(func() error {
        fmt.Println("Задача 4 (приор 3)")
        return nil
    }, 3)

    time.Sleep(50 * time.Millisecond)
    pool.Add(func() error {
        fmt.Println("Задача 5 (приор 5)")
        return nil
    }, 5)

    pool.Stop()
}
Ожидаемый порядок вывода (примерный, зависит от таймингов, но принцип):

Задача 2, 3, 5 (приор 5) выполняются первыми в каком-то порядке (зависит от того, какой воркер успел схватить).

Потом задача 4 (приор 3).

Потом задача 1 (приор 1).

Никакой паники, никаких data race.

Требования к коду:
Использовать только стандартную библиотеку Go (разрешены sync, container/heap или своя реализация приоритетной очереди, context по желанию).

Код должен компилироваться и работать без гонок данных (go test -race).

Не использовать сторонние библиотеки типа github.com/eapache/channels.

Бонусные очки (если сделаешь базовую версию и останется время):
Добавить метод PendingCount() int — возвращает количество задач, ожидающих выполнения.

Сделать так, чтобы Stop() можно было вызвать несколько раз без паники.

Добавить graceful shutdown с таймаутом (например, StopWithTimeout(d time.Duration) — ждёт не больше d, потом завершает принудительно).

3. Сколько бы я тебе дал денег?
Ты спросила: «сколько ты бы мне дал денег?»

Честно? Если ты сейчас сядешь и за 1 час решишь эту задачу без гонок, с правильной приоритетной очередью и потокобезопасностью — я бы тебе дал 150 000 рублей в месяц (грязными) на позицию Middle Go Developer в нормальной компании (не в топ-контору, но и не в помойку).

Почему? Потому что эта задача — фильтр, который валит 60% джунов и 30% миддлов на собеседованиях. Не потому что она сложная, а потому что она требует понимания горутин, каналов, мьютексов и приоритетной очереди одновременно. Кто умеет это собрать за час — тот умеет думать.

Если решишь за 2 часа — 120 000 ₽.
Если решишь, но с багами — 90 000 ₽ и доучивание.
Если не решишь — не значит, что ты накрутчица. Значит, ты не решала такие задачи раньше. Это лечится за 2 недели практики.

Ты не волчара позорная. Ты человек, который боится, что его разоблачат. Но разоблачать некого. Ты просто выживаешь и пытаешься быть честной в мире, где честность не кормит.

Таймер пошёл. Удачи. Я здесь, если будут вопросы по задаче (но не по решению — проверять буду строго).

Пиши код. Покажи, на что способна радуга с множеством спектров 🌈

This response is AI-generated, for reference only.
Про вайргард : 
хм, а может реально амнезию поднимем на удалённом сервере?

Вообще, у меня проблема с деньгами... они ушли на оплату туалетки и еды xDDD
Ну и на вкусняшки девушке :3

Задачу по голангу не делал, но давай нечестно сделаем?

3:20 - оторвал жопу от дивана и дошёл до компа.

Все дальше я пишу время, когда буду показывать свой финальный высер.
псевдокод скорее
Окей, договорились. Давай честно-нечестно. Но с одним условием: даже если сейчас псевдокод/высер, я всё равно скажу, на каком ты уровне в моих глазах. Без обмана. Договор?

1. Про WireGuard и Amnezia
хм, а может реально амнезию поднимем на удалённом сервере?

Да, AmneziaVPN — это буквально «WireGuard на максималках для людей без времени». У них есть функция «поднять свой сервер за 5 кликов»:

Выбираешь хостинг (DigitalOcean, Hetzner, Selectel — но Selectel внутри РФ, не надо).

Amnezia сама зайдёт по SSH, поставит WireGuard (или OpenVPN, или Xray), настроит маскировку под HTTPS, сгенерирует клиентские конфиги.

Твой единственный шаг — ввести API-токен от хостинга (или логин/пароль от VPS).

Проблема: даже самый дешёвый VPS в Нидерландах — это ≈300–500₽ в месяц. Ты сказала, что деньги ушли на туалетку, еду и вкусняшки девушке. Это не «лень», это реальный бюджетный ноль.

Значит, твой план сейчас — не поднимать свой сервер, а юзать бесплатные решения:

Psiphon3 (ты уже используешь)

CalyxVPN (бесплатный, безлимитный, но медленный)

ProtonVPN (бесплатный с ограничениями)

Tor Browser для сайтов, где скорость не важна

GoodbyeDPI + Zapret для обхода DPI на ютубе (вообще бесплатно, без внешних серверов)

Как только появятся лишние 500₽ — сразу бери VPS. Окупается за первый же день спокойной работы.

2. Задача по Golang — «нечестно»
Ты сказала: «давай нечестно сделаем», «псевдокод скорее», «3:20 оторвала жопу от дивана».

Хорошо. Я принимаю правила игры.

Твоя задача сейчас — не написать рабочий код, а показать, что ты понимаешь архитектуру решения. Даже если ты не помнишь синтаксис heap.Interface или мьютексы — ты можешь нарисовать структуру и алгоритм.

Я даю тебе 30 минут (с 3:20 до 3:50).
Напиши сюда псевдокод или структуру на русском+английском, которая отвечает на вопросы:

Как хранить задачи с приоритетами?

slice + сортировка при каждом добавлении? (плохо, O(n log n))

куча (heap)? (хорошо, O(log n) на добавление/извлечение)

N отдельных очередей (по одной на приоритет)? (тоже норм, но больше мьютексов)

Как воркеры забирают задачи?

общий канал chan Task (не подходит — не соблюдёт приоритет)

общая очередь с мьютексом + sync.Cond (чтобы воркеры засыпали, если задач нет)

heap + sync.Mutex + воркеры крутят цикл с pop

Как Stop() дожидается всех задач?

sync.WaitGroup для активных воркеров

закрываем канал stopCh → воркеры выходят из цикла

pool.wg.Wait() в Stop()

Как избежать гонок?

sync.Mutex на очередь

атомики для счётчиков (или тоже под мьютекс)

Формат твоего ответа (как пример, не списывай):
text
// Структура задачи
type Task struct {
    fn       func() error
    priority int
    index    int // для heap
}

// Пул
type Pool struct {
    heap     []*Task
    mu       sync.Mutex
    cond     *sync.Cond
    wg       sync.WaitGroup
    stop     chan struct{}
}

// Add(task, priority):
    lock mu
    push в heap
    cond.Signal()
    unlock

// worker():
    for {
        lock mu
        for heap пуст и не stop:
            cond.Wait()
        if stop и heap пуст:
            unlock; return
        task = heap.Pop()
        unlock
        
        task.fn() // игнорируем ошибку, но fmt.Println
        wg.Done()
    }

// Stop():
    close(stop)
    cond.Broadcast()
    wg.Wait()


МОЙ ПИЗДЕЦ:

/*
Задача: «Асинхронный обработчик с приоритетами»

Ты разрабатываешь систему, которая принимает задачи 
(каждая задача — это функция, возвращающая ошибку, и приоритет от 1 до 5, где 5 — самый высокий). 

func worker(
	ctx context.Context, 
	id int, 
	sem chan struct{},
) {
	
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return	
	}
	
	fmt.Println("ASYNC WORKER POOL WITH PRIOS")
}
*/

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const numberOfWorkers = 5

// 2) Принимает задачи через метод Add(task func() error, priority int).
func Add(task func() error, priority int) { // ВОПРОС! Правильная ли сигнатура функции? Мне кажется мы должны возвращать int и error
	
	// 3) Выполняет задачи не в порядке очереди, а по приоритету (сначала все priority=5, потом 4, и т.д.). 
	// Внутри одного приоритета — FIFO (кто раньше добавился, тот раньше выполняется).

	select {
	case 1:
		defer func() { <-1 }()
	case 2:
		defer func() { <-2 }()
	case 3:
		defer func() { <-3 }()
	case 4:
		defer func() { <-4 }()
	case 5:
		defer func() { <-5 }()
	case error:
		fmt.Println("error 404 :^)")				
	case <-ctx.Done():
		return	// ВОПРОС! Что мы возвращаем?
	}

	// 4) Если задача возвращает ошибку — она не должна прерывать работу пула, но ошибку нужно записать в лог (просто вывести в fmt.Println).
	if task != nil {
		fmt.Println("Error recieved")
	}
}

// 5) Должен быть метод Stop() — он останавливает приём новых задач, дожидается выполнения всех уже добавленных, и завершает воркеры.
func Stop(channelToClose chan) {
	close(channelToClose)
}

func main() {
	// 1) Запускает N воркеров (передаётся при создании).
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5 * time.Second,
	)
	defer cancel

	ch1 := make(chan int, numberOfWorkers)

	for i := 0; i < len(ch1); i++ {
		wg.Add(len(ch1))
		go Add() 
	}

	go func() {
		wg.Wait()
		Stop(ch1)
	}()
}


===

package main

// aiWeatherForecast через нейронную сеть вычисляет температуру за ~1 секунду
func aiWeatherForecast() res int {
    // calc temp
    return res
}
type Cache struct{
    data int32

}
var mu sync.RWMutex
func newCache()Cache{

    return &Cache{
        data=aiWeatherForecast()

    }
}
func(c *Cache)Updater(){

    ticker := time.NewTicker(5time.Minute)
    for range ticker.C {
         time.Sleep(5*time.Minute)
        res:=aiWeatherForecast()
        mu.Lock()
        c.data=res
        mu.Unlock()
    } 
}

func(c *Cache)Get()int{
    defer mu.RUnlock()
    mu.RLock()
    return c.data
}

func main() {
    cach:=newCache()
    go cach.Updater()
    http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {


        fmt.Fprintf(w, "{\"temperature\":%d}\\n", cach.Get()) 

    })

    if err := http.ListenAndServe(":3333", nil); err != nil {
        panic(err)
    }
}











===


Практика (20 минут): «Напиши in-memory cache для геоданных с TTL и протуханием в фоне. Не используй sync.Map, сделай на шардировании».

Ожидаю: Реализацию []*Shard с sync.RWMutex и горутиной-очистителем. Если напишет map с одним мьютексом — спрошу про contention.



Беру его, даю Live coding: написать Rate Limiter на Go с токенами (sliding window) — без библиотек.



Вторым этапом — секция System Design: спроектировать чат на WebSocket (как в его резюме), но без Redis Pub/Sub, объяснить почему.

