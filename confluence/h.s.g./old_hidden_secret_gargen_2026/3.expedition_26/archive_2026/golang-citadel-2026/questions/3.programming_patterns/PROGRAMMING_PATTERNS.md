# PROGRAMMING PATTERNS INSIGHTS
БЛОК 1: ФУНДАМЕНТ (Как вообще можно думать о коде)
Это про парадигмы и базовые принципы существования кода.

1.1 Парадигмы программирования (Стили мышления)
Представьте, что код — это инструкция для человека:

Императивный — вы говорите КАК именно делать: "Поверни направо, сделай 5 шагов, возьми красную кружку". (C, Go, ассемблер)

Декларативный — вы говорите ЧТО нужно получить: "Принеси мне красную кружку". А как — не ваша забота. (SQL: SELECT * FROM users, HTML: Текст жирный)

Процедурный — разновидность императивного. Вы группируете инструкции в процедуры (функции), чтобы не писать одно и то же. "Сходи в магазин" — это процедура, внутри которой куча шагов.

Функциональный — декларативный подход. Мир состоит из функций, как в математике. Функции не имеют побочных эффектов (не меняют внешний мир, всегда возвращают одно и то же для одних входных данных). Удобно для параллельных вычислений.

1.2 Go-специфика (Идиомы)
Go берет лучшее из разных миров:

ООП без ООП: В Go нет классов, но есть структуры (struct) и интерфейсы.

Композиция вместо наследования: Вместо "Кот наследуется от Животного", мы говорим "Кот включает в себя свойства Животного" (встраивание структур).

Интерфейсы: Описывают поведение. "Что-то, что умеет читать" (io.Reader). "Что-то, что можно превратить в строку" (fmt.Stringer).

Ошибки — это значения: Нет исключений (try-catch). Функция возвращает результат И ошибку. Вы обязаны с этим работать.

Concurrency встроена в язык: Горутины (легкие потоки) и каналы (трубы для общения между ними) — это часть языка, а не библиотека.

БЛОК 2: БАЗОВЫЕ ПРИНЦИПЫ (Правила хорошего тона)
Это аксиомы, которые работают в любом языке.

2.1 KISS (Keep It Simple, Stupid)
Суть: Самое простое решение — самое лучшее. Не надо городить ООП из 15 классов там, где можно обойтись тремя строчками.

Визуализация: Вы пишете код не для компилятора, а для коллеги (и для себя через полгода).

2.2 DRY (Don't Repeat Yourself)
Суть: Если вы скопировали кусок кода и вставили его в другое место — вы что-то делаете не так. Выносите повторяющуюся логику в функции/методы.

Исключение: Иногда лучше два раза скопировать простой код, чем создавать монструозную абстракцию (тут пересекается с KISS).

2.3 YAGNI (You Ain't Gonna Need It)
Суть: Не пиши код "на вырост". Не думай "а вдруг нам потом понадобится универсальная фабрика фасадов". Пишите только то, что нужно сейчас.

2.4 SOLID (5 заповедей ООП)
Важно: В Go нет классов в классическом понимании, но принципы применимы к типам и пакетам.

S (Single Responsibility — Принцип единственной ответственности): У класса/структуры/функции должна быть только одна причина для изменения. Если ваша структура User умеет и логиниться, и сохранять себя в базу, и отправлять email — это беда.

O (Open/Closed — Принцип открытости/закрытости): Код должен быть открыт для расширения (добавления нового функционала), но закрыт для изменения (переписывания старого кода). Достигается через интерфейсы.

Пример: У вас есть функция, которая принимает интерфейс Sender. Вы можете дописать SmsSender или EmailSender, не трогая старый код.

L (Liskov Substitution — Принцип подстановки Барбары Лисков): Если у вас есть функция, которая работает с "Птицей", то передав туда "Утку", всё должно работать. Подкласс/реализация не должен ломать логику базового класса/интерфейса. (Классическая проблема: "Квадрат" не должен наследоваться от "Прямоугольника", если в прямоугольнике можно менять стороны независимо).

I (Interface Segregation — Принцип разделения интерфейсов): Не заставляйте клиента реализовывать методы, которые он не использует. Лучше много маленьких интерфейсов, чем один "толстый".

Плохо: Интерфейс Worker с методами Work() и Eat(). Робот умеет работать, но не умеет есть.

Хорошо: Workable и Eatable.

D (Dependency Inversion — Принцип инверсии зависимостей): Модули верхнего уровня не должны зависеть от модулей нижнего уровня. И те, и другие должны зависеть от абстракций (интерфейсов).

Суть: Ваш сервис заказов (upper) не должен жестко зависеть от конкретной базы данных PostgreSQL (lower). Он должен зависеть от абстракции OrderRepository (интерфейса). А уж PostgreSQL будет просто одной из реализаций этого интерфейса.

БЛОК 3: ПАТТЕРНЫ ПРОЕКТИРОВАНИЯ (Типовые чертежи)
Паттерн — это не код, а чертеж. Это архитектурный прием, проверенный годами.

3.1 Классификация (Три кита)
А. ПОРОЖДАЮЩИЕ (Как создавать объекты, чтобы не стрелять себе в ногу)

Одиночка (Singleton):

Проблема: Нужен один и только один объект на всё приложение.

Решение: Класс сам контролирует свое создание.

В Go: sync.Once + глобальная переменная.

Аналогия: Президент страны. В каждый момент времени может быть только один.

Фабричный метод (Factory Method):

Проблема: Заранее неизвестно, объект какого именно типа нужно создать.

Решение: Создаем "фабрику", у которой есть метод Create. Этот метод на основе входных данных сам решает, какой объект вернуть.

Аналогия: Автомат по выдаче игрушек. Вы кидаете жетон и нажимаете кнопку, а автомат сам решает, дать вам мячик или машинку. Вы не знаете, как он устроен внутри.

Строитель (Builder):

Проблема: Объект очень сложный и создается поэтапно. Есть много вариантов его конфигурации.

Решение: Выносим конструирование в отдельный класс.

Аналогия: Сборка компьютера. Вы не собираете его "в одну строку". Сначала ставите процессор, потом память, потом видеокарту. "Строитель" знает правильную последовательность.

Б. СТРУКТУРНЫЕ (Как организовать классы в большую систему)

Адаптер (Adapter):

Проблема: У вас есть класс с нужным функционалом, но он имеет не тот интерфейс, который ожидает ваш клиент.

Решение: Пишем класс-прокладку (Адаптер), который "оборачивает" неудобный класс и выдает нужный интерфейс.

Аналогия: Розетка (евро) и вилка (американская). Нужен переходник. Переходник — это Адаптер.

Фасад (Facade):

Проблема: Сложная система из кучи классов. Клиенту нужно просто включить телевизор, а не знать про электричество, магнитные поля и видеокодеки.

Решение: Создаем простой интерфейс (Фасад) поверх сложной системы.

Аналогия: Стойка регистрации в отеле. За ней стоит куча служб: уборка, ремонт, бронирование. Но вам не нужно знать об этом. Вы просто подходите к стойке и просите: "Разбудите меня в 7 утра".

Декоратор (Decorator):

Проблема: Нужно динамически добавлять новые обязанности объекту, не меняя его код.

Решение: Создаем "обертку", которая содержит исходный объект и расширяет его функционал.

Аналогия: Заказ кофе. Вы берете черный кофе (базовый объект) и "декорируете" его молоком, потом сиропом, потом взбитыми сливками. Каждый декоратор добавляет что-то свое.

Прокси (Proxy):

Проблема: Нужно контролировать доступ к объекту (ленивая загрузка, проверка прав, логирование).

Решение: Создаем объект-заменитель, который стоит между клиентом и реальным объектом.

Аналогия: Кредитная карта. Это "прокси" вашего счета в банке. Она контролирует доступ к реальным деньгам.

В. ПОВЕДЕНЧЕСКИЕ (Как объекты общаются)

Стратегия (Strategy):

Проблема: Есть семейство алгоритмов, которые должны быть взаимозаменяемы.

Решение: Каждый алгоритм выносится в отдельный класс с общим интерфейсом.

Аналогия: Навигатор. Вы можете выбрать стратегию "пешком", "на машине", "на велосипеде". Алгоритм расчета маршрута меняется, но интерфейс (кнопка "Проложить маршрут") остается тем же.

Цепочка обязанностей (Chain of Responsibility):

Проблема: Запрос должен обрабатываться несколькими объектами, но заранее неизвестно, каким именно.

Решение: Строим цепочку обработчиков. Запрос идет по цепочке, пока кто-то его не обработает.

Аналогия: Техподдержка. Ваш вопрос идет: Бот -> Оператор первой линии -> Старший специалист. Если бот решил — остальным вопрос не идет.

Итератор (Iterator):

Проблема: Нужно пройтись по элементам коллекции, не вдаваясь в подробности, как она устроена внутри (список, массив, дерево).

Решение: Даем клиенту простой интерфейс: Next(), HasNext(), Current().

Аналогия: Лента в супермаркете. Вы просто берете товары по одному, не задумываясь, как они хранились на складе.

БЛОК 4: АРХИТЕКТУРА (Как собрать дом, чтобы не развалился)
4.1 Чистая архитектура (Clean Architecture)
Это про то, как организовать код в проекте, чтобы можно было легко заменить БД или фреймворк.

Главное правило: Зависимости направлены внутрь. Внешний мир зависит от внутреннего, но не наоборот.

Ядро (Domain/Entities): Самая суть. Бизнес-правила, которые не зависят ни от чего. Здесь живут ваши структуры User, Order, интерфейсы репозиториев. Этот слой вообще не знает про базу данных или веб.

Прикладной слой (Application/Use Cases): Сценарии использования. Содержит логику приложения: "Зарегистрировать пользователя", "Создать заказ". Он использует интерфейсы из Ядра, но не знает конкретных реализаций.

Инфраструктура (Infrastructure): Работа с внешним миром. Здесь живут конкретные реализации репозиториев (например, PostgresUserRepository), отправка email, работа с API. Он зависит от Прикладного слоя и Ядра.

Представление (Presentation/Interface): То, через что пользователь или внешняя система общается с приложением. Контроллеры, хендлеры, middleware. Он использует Прикладной слой.

Аналогия с рестораном:

Domain (Ядро): Сама еда (стейк, картошка). Рецепты.

Application (Прикладной слой): Повар. Он знает рецепты и может приготовить блюдо.

Infrastructure (Инфраструктура): Холодильник, плита, поставщики продуктов (БД, API). Повар ими пользуется, но не зависит от конкретной модели плиты.

Presentation (Представление): Официант, который принимает заказ и приносит блюдо.

БЛОК 5: КОНКУРЕНТНОСТЬ (Важно для Go)
Concurrency (Конкурентность): Умение "делать несколько дел одновременно" (в контексте одного процессора — быстро переключаться между ними). Представьте, что вы готовите обед: поставили чайник, пока он греется — режете хлеб.

Parallelism (Параллелизм): Физическое выполнение нескольких дел в один момент времени (на многоядерном процессоре). Вы готовите на двух конфорках одновременно.

В Go: Горутины (легковесные потоки) и каналы (способы коммуникации между ними) позволяют удобно реализовывать конкурентность. Параллелизм — это уже возможность железа, Go умеет ее использовать, если она есть.

Как это учить?
Понять аналогию: Привязывайте каждый сложный термин к бытовой ситуации.

Тезисность: Не читайте "Войну и мир" про паттерны. Используйте конспект как шпаргалку.

Практика: Пытайтесь найти эти паттерны в коде, который вы уже пишете. Адаптер — это же обертка вокруг сторонней библиотеки, верно? А Стратегия — выбор способа оплаты.


Go-паттерны: Код-шпора
1. ОДИНОЧКА (Singleton)
   Гарантирует, что у структуры только один экземпляр.

go
package singleton

import "sync"

type Database struct {
connection string
}

var instance *Database
var once sync.Once

func GetDatabase() *Database {
once.Do(func() {
// Выполнится только один раз, даже из разных горутин
instance = &Database{connection: "postgres://user:pass@localhost/db"}
})
return instance
}

// Использование
db1 := GetDatabase()
db2 := GetDatabase()
fmt.Println(db1 == db2) // true
2. ФАБРИЧНЫЙ МЕТОД (Factory Method)
   Создает объекты без указания конкретного типа.

go
package factory

// Интерфейс продукта
type Animal interface {
Speak() string
}

// Конкретные продукты
type Dog struct{}
func (d Dog) Speak() string { return "Гав!" }

type Cat struct{}
func (c Cat) Speak() string { return "Мяу!" }

// Фабрика
func AnimalFactory(animalType string) Animal {
switch animalType {
case "dog":
return Dog{}
case "cat":
return Cat{}
default:
return nil
}
}

// Использование
animal := AnimalFactory("dog")
fmt.Println(animal.Speak()) // Гав!
3. СТРОИТЕЛЬ (Builder)
   Пошаговое создание сложного объекта.

go
package builder

// Продукт
type Pizza struct {
Dough   string
Sauce   string
Topping string
}

// Строитель
type PizzaBuilder struct {
dough   string
sauce   string
topping string
}

func NewPizzaBuilder() *PizzaBuilder {
return &PizzaBuilder{}
}

func (b *PizzaBuilder) SetDough(dough string) *PizzaBuilder {
b.dough = dough
return b // Возвращаем себя для цепочек вызовов
}

func (b *PizzaBuilder) SetSauce(sauce string) *PizzaBuilder {
b.sauce = sauce
return b
}

func (b *PizzaBuilder) SetTopping(topping string) *PizzaBuilder {
b.topping = topping
return b
}

func (b *PizzaBuilder) Build() Pizza {
return Pizza{
Dough:   b.dough,
Sauce:   b.sauce,
Topping: b.topping,
}
}

// Использование
pizza := NewPizzaBuilder().
SetDough("тонкое").
SetSauce("томатный").
SetTopping("сыр").
Build()
4. АДАПТЕР (Adapter)
   Стыкует несовместимые интерфейсы.

go
package adapter

// Целевой интерфейс (чего хотим)
type USB interface {
ConnectWithUSB() string
}

// Структура, которую нужно адаптировать
type MemoryCard struct{}

func (m MemoryCard) Insert() string {
return "Карта памяти вставлена"
}

// Адаптер
type CardReader struct {
card MemoryCard
}

func (r CardReader) ConnectWithUSB() string {
// Адаптируем Insert() в ConnectWithUSB()
return "Адаптер: " + r.card.Insert()
}

// Использование
card := MemoryCard{}
reader := CardReader{card: card}
// reader теперь можно использовать там, где нужен USB
fmt.Println(reader.ConnectWithUSB())
5. ФАСАД (Facade)
   Простой интерфейс к сложной системе.

go
package facade

// Сложная подсистема
type Amplifier struct{}
func (a Amplifier) On() string    { return "Усилитель включен" }
func (a Amplifier) SetVolume(v int) string { return "Громкость: " + string(rune(v)) }

type DVDPlayer struct{}
func (d DVDPlayer) Play(movie string) string { return "Фильм: " + movie }

type Projector struct{}
func (p Projector) WideScreenMode() string { return "Широкий экран" }

// Фасад
type HomeTheaterFacade struct {
amp     Amplifier
dvd     DVDPlayer
proj    Projector
}

func NewHomeTheaterFacade() *HomeTheaterFacade {
return &HomeTheaterFacade{
amp:  Amplifier{},
dvd:  DVDPlayer{},
proj: Projector{},
}
}

// Один простой метод вместо кучи
func (h *HomeTheaterFacade) WatchMovie(movie string) string {
result := h.amp.On() + "\n"
result += h.proj.WideScreenMode() + "\n"
result += h.dvd.Play(movie)
return result
}

// Использование
theater := NewHomeTheaterFacade()
fmt.Println(theater.WatchMovie("Матрица"))
6. ДЕКОРАТОР (Decorator)
   Динамически добавляет функциональность.

go
package decorator

// Интерфейс
type Coffee interface {
Cost() int
Description() string
}

// Базовый объект
type SimpleCoffee struct{}

func (c SimpleCoffee) Cost() int {
return 100
}

func (c SimpleCoffee) Description() string {
return "Черный кофе"
}

// Декоратор (базовая структура для всех декораторов)
type CoffeeDecorator struct {
coffee Coffee
}

// Конкретный декоратор: Молоко
type MilkDecorator struct {
CoffeeDecorator
}

func NewMilkDecorator(c Coffee) *MilkDecorator {
return &MilkDecorator{CoffeeDecorator{coffee: c}}
}

func (m MilkDecorator) Cost() int {
return m.coffee.Cost() + 30 // молоко +30
}

func (m MilkDecorator) Description() string {
return m.coffee.Description() + ", с молоком"
}

// Использование
coffee := SimpleCoffee{}
fmt.Println(coffee.Description(), coffee.Cost())

withMilk := NewMilkDecorator(coffee)
fmt.Println(withMilk.Description(), withMilk.Cost())
7. СТРАТЕГИЯ (Strategy)
   Взаимозаменяемые алгоритмы.

go
package strategy

// Интерфейс стратегии
type PaymentStrategy interface {
Pay(amount int) string
}

// Конкретные стратегии
type CreditCard struct {
Name   string
Number string
}

func (c CreditCard) Pay(amount int) string {
return fmt.Sprintf("Оплачено %d через карту %s", amount, c.Name)
}

type PayPal struct {
Email string
}

func (p PayPal) Pay(amount int) string {
return fmt.Sprintf("Оплачено %d через PayPal (%s)", amount, p.Email)
}

type Cash struct{}

func (c Cash) Pay(amount int) string {
return fmt.Sprintf("Оплачено %d наличными", amount)
}

// Контекст, использующий стратегию
type Checkout struct {
payment PaymentStrategy
}

func (c *Checkout) SetPaymentStrategy(p PaymentStrategy) {
c.payment = p
}

func (c *Checkout) ProcessOrder(amount int) string {
return c.payment.Pay(amount)
}

// Использование
checkout := &Checkout{}

checkout.SetPaymentStrategy(CreditCard{Name: "Иван", Number: "1234"})
fmt.Println(checkout.ProcessOrder(1000))

checkout.SetPaymentStrategy(PayPal{Email: "ivan@mail.ru"})
fmt.Println(checkout.ProcessOrder(500))

checkout.SetPaymentStrategy(Cash{})
fmt.Println(checkout.ProcessOrder(300))
8. ЦЕПОЧКА ОБЯЗАННОСТЕЙ (Chain of Responsibility)
   Запрос идет по цепочке, пока не обработается.

go
package chain

// Интерфейс обработчика
type Handler interface {
SetNext(handler Handler) Handler
Handle(request string) string
}

// Базовый обработчик
type BaseHandler struct {
next Handler
}

func (h *BaseHandler) SetNext(handler Handler) Handler {
h.next = handler
return handler
}

func (h *BaseHandler) Handle(request string) string {
if h.next != nil {
return h.next.Handle(request)
}
return "Никто не обработал: " + request
}

// Конкретные обработчики
type AuthHandler struct {
BaseHandler
}

func (h *AuthHandler) Handle(request string) string {
if request == "auth" {
return "AuthHandler: запрос авторизован"
}
return h.BaseHandler.Handle(request)
}

type ValidationHandler struct {
BaseHandler
}

func (h *ValidationHandler) Handle(request string) string {
if request == "validate" {
return "ValidationHandler: данные валидны"
}
return h.BaseHandler.Handle(request)
}

type CacheHandler struct {
BaseHandler
}

func (h *CacheHandler) Handle(request string) string {
if request == "cache" {
return "CacheHandler: данные из кэша"
}
return h.BaseHandler.Handle(request)
}

// Использование
auth := &AuthHandler{}
validation := &ValidationHandler{}
cache := &CacheHandler{}

// Строим цепочку: auth -> validation -> cache
auth.SetNext(validation).SetNext(cache)

// Запрос пойдет по цепочке
fmt.Println(auth.Handle("auth"))      // обработает AuthHandler
fmt.Println(auth.Handle("validate"))  // Auth пропустит -> обработает Validation
fmt.Println(auth.Handle("cache"))     // дойдет до Cache
fmt.Println(auth.Handle("unknown"))   // дойдет до конца
9. ИТЕРАТОР (Iterator)
   Обход коллекции без знания внутренней структуры.

go
package iterator

// Интерфейс итератора
type Iterator interface {
HasNext() bool
Next() interface{}
}

// Коллекция
type UserCollection struct {
users []string
}

func (u *UserCollection) Add(user string) {
u.users = append(u.users, user)
}

// Создаем итератор для этой коллекции
func (u *UserCollection) Iterator() Iterator {
return &UserIterator{
users: u.users,
index: 0,
}
}

// Конкретный итератор
type UserIterator struct {
users []string
index int
}

func (it *UserIterator) HasNext() bool {
return it.index < len(it.users)
}

func (it *UserIterator) Next() interface{} {
if it.HasNext() {
user := it.users[it.index]
it.index++
return user
}
return nil
}

// Использование
collection := &UserCollection{}
collection.Add("Анна")
collection.Add("Иван")
collection.Add("Мария")

iter := collection.Iterator()
for iter.HasNext() {
fmt.Println(iter.Next())
}
Что важно запомнить по коду:
Паттерн	Ключевая идея в Go
Singleton	sync.Once
Factory	switch по типу, возвращаем интерфейс
Builder	return b для цепочек вызовов
Adapter	Структура-обертка, которая переводит вызовы
Facade	Структура, собирающая сложные вызовы в один метод
Decorator	Композиция (встраивание интерфейса) + добавление функционала
Strategy	Интерфейс, который можно подменить
Chain	Каждый обработчик хранит ссылку на следующий
Iterator	Состояние (индекс) внутри итератора
Если нужно разобрать какой-то конкретный паттерн глубже — пиши!


Можно ли это назвать фасадом?

type LibraryService interface {
RentBook(bookID, userID int) (*model.Book, error)
ReturnBook(bookID, userID int) (*model.Book, error)
GetAvailableBooks() ([]model.Book, error)
GetBookDetails(bookID int) (*model.Book, error)
GetUserRentals(userID int) ([]model.Book, error)
GetUserDetails(userID int) (*model.User, error)
GetAuthorBooks(authorID int) ([]model.Book, error)
GetTopAuthors(limit int) ([]model.Author, error)
}
Короткий ответ: Нет, это не фасад.

Длинный ответ: это контроллер или API-слой, собранный "в кучу".

Почему это НЕ фасад?
Фасад создают, когда у вас есть сложная подсистема (куча классов/сервисов), и вы хотите дать клиенту простой интерфейс, чтобы не париться с внутренностями.

Ваш интерфейс:

Не скрывает сложность, а просто перечисляет 8 методов.

Не объединяет разные подсистемы — методы явно дергают книги, пользователей, авторов.

Скорее всего, просто проксирует запросы в соответствующие сервисы.

Как выглядел бы настоящий Фасад
Предположим, у библиотеки есть подсистемы:

InventoryService (учет книг)

UserService (пользователи)

PaymentService (штрафы/оплата)

NotificationService (уведомления)

Клиенту (мобильному приложению) нужно просто взять книгу. Но процесс сложный:

Проверить, есть ли книга в наличии (Inventory)

Проверить, нет ли у пользователя долгов (Payment)

Заблокировать книгу в инвентаре (Inventory)

Записать в историю пользователя (User)

Отправить уведомление (Notification)

Если книга последняя — уведомить библиотекаря

Фасад скрывает это всё:
go
// Фасад для мобильного приложения
type MobileLibraryFacade struct {
inventory  *InventoryService
users      *UserService
payments   *PaymentService
notifier   *NotificationService
}

func NewMobileLibraryFacade() *MobileLibraryFacade {
return &MobileLibraryFacade{
inventory:  NewInventoryService(),
users:      NewUserService(),
payments:   NewPaymentService(),
notifier:   NewNotificationService(),
}
}

// ОДИН метод вместо кучи вызовов
func (f *MobileLibraryFacade) BorrowBook(bookID, userID int) error {
// 1. Проверка пользователя
if err := f.users.CheckUser(userID); err != nil {
return err
}

    // 2. Проверка долгов
    if f.payments.HasDebt(userID) {
        return errors.New("у вас долги")
    }
    
    // 3. Проверка наличия
    book, err := f.inventory.GetBook(bookID)
    if err != nil {
        return err
    }
    
    // 4. Выдача
    if err := f.inventory.MarkAsBorrowed(bookID); err != nil {
        return err
    }
    
    // 5. История
    f.users.AddToHistory(userID, bookID)
    
    // 6. Уведомление
    f.notifier.SendUserNotification(userID, "Книга выдана")
    
    // 7. Если последний экземпляр - уведомить библиотекаря
    if f.inventory.IsLastCopy(bookID) {
        f.notifier.NotifyLibrarian("Остался последний экземпляр!")
    }
    
    return nil
}

// Еще один метод фасада
func (f *MobileLibraryFacade) GetUserDashboard(userID int) (*UserDashboard, error) {
// Собирает данные из 3-х сервисов в одну структуру для экрана пользователя
}
Клиент делает:

go
facade.BorrowBook(42, 123)  // ОДИН вызов вместо 6-ти
facade.GetUserDashboard(123) // ОДИН вызов вместо кучи запросов к разным сервисам
Чем является ваш интерфейс?
Ваш код — это типичный толстый интерфейс (Fat Interface) или контроллер, который:

Собирает в кучу методы из разных доменных областей

Скорее всего, будет реализован через вызовы других сервисов

Нарушает ISP (Interface Segregation) — клиенты, которым нужны только книги, вынуждены зависеть от методов про пользователей и авторов

Как исправить?
Разделить на интерфейсы по ответственности:

go
type BookService interface {
RentBook(bookID, userID int) (*model.Book, error)
ReturnBook(bookID, userID int) (*model.Book, error)
GetAvailableBooks() ([]model.Book, error)
GetBookDetails(bookID int) (*model.Book, error)
}

type UserService interface {
GetUserDetails(userID int) (*model.User, error)
GetUserRentals(userID int) ([]model.Book, error)
}

type AuthorService interface {
GetAuthorBooks(authorID int) ([]model.Book, error)
GetTopAuthors(limit int) ([]model.Author, error)
}
А если нужен один объект для внешнего API, то либо:

Контроллер, который дергает эти сервисы по отдельности

Настоящий фасад, который реально скрывает сложность и объединяет операции (как BorrowBook выше)

TL;DR
Характеристика	Ваш код	Фасад
Скрывает сложность	❌ Просто перечисляет методы	✅ Дает простые методы для сложных операций
Объединяет подсистемы	❌ Каждый метод дергает что-то одно	✅ Один метод дергает много подсистем
Упрощает клиенту	❌ Клиент сам должен комбинировать вызовы	✅ Клиент делает 1 вызов вместо N
Пример метода	GetBookDetails	BorrowBook (где внутри проверка, выдача, уведомление)

